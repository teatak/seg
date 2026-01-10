package optimizer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DictBase     = "data/dict_base.txt"
	DictCore     = "data/dict_core.txt"
	DictUser     = "data/dict_user.txt"
	DictCombined = "data/dictionary_combined.tmp.txt"
	DictCleanTmp = "data/dictionary_auto_clean.tmp"
	TextFile     = "data/text.txt"
	CorpusFile   = "data/corpus.txt"
	ModelFile    = "data/model.crf"
)

// Run executes the optimization pipeline.
func Run(newWordsFile string) error {
	log.SetPrefix("[OPT] ")
	log.Println("=== Starting Optimization Pipeline (Internal) ===")
	log.Printf("Time: %s", time.Now().Format(time.RFC3339))

	if _, err := os.Stat(newWordsFile); os.IsNotExist(err) {
		return fmt.Errorf("new words file not found: %s", newWordsFile)
	}

	ensureFile(DictUser)
	ensureFile(DictBase)

	// 0. Load Top Brands
	topBrands, _ := loadTopBrands(DictBase)

	// 1. Backup User Dict
	log.Println("[1/6] Backing up user dictionary...")
	copyFile(DictUser, DictUser+".bak")

	// 2. Merge New Words
	log.Println("[2/6] Merging new words into user dictionary...")
	if err := mergeNewWords(newWordsFile, DictUser); err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	// 3. Prune Interference (User Only)
	log.Println("[2.5/6] Pruning interference (User Dict only)...")
	if err := PruneInterference(DictUser, newWordsFile, DictUser); err != nil {
		return fmt.Errorf("prune failed: %w", err)
	}

	// 4. Clean User Dict
	log.Println("[3/6] Cleaning user dictionary (preserving all manual entries)...")
	// Use a very high ratio to effectively disable prefix/suffix pruning for manual feedback
	if err := CleanDictionary(DictUser, DictCleanTmp, 1000.0); err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}
	if err := os.Rename(DictCleanTmp, DictUser); err != nil {
		return fmt.Errorf("failed to move clean dict: %w", err)
	}

	// 4.5 Discover New Words (Unsupervised learning from text.txt)
	log.Println("[3.5/6] Discovering new words from raw text...")
	DictDiscovered := "data/dict_discovered.tmp"
	// threshold=5, maxGram=4
	if err := Discover(TextFile, DictDiscovered, 5, 4); err != nil {
		log.Printf("Warning: Discovery failed: %v", err)
	} else {
		// Merge discovered words into Core dictionary (since they are auto-learned)
		if err := combineFiles(DictCore+".tmp", DictCore, DictDiscovered); err == nil {
			os.Rename(DictCore+".tmp", DictCore)
		}
		os.Remove(DictDiscovered)
	}

	// 5. Create Combined Dict (Core -> Base -> User)
	log.Println("[4/6] Creating combined dictionary (Core + Base + User)...")
	if err := combineFiles(DictCombined, DictCore, DictBase, DictUser); err != nil {
		return fmt.Errorf("combine failed: %w", err)
	}

	// 6. Generate Corpus
	log.Println("[5/6] Re-segmenting corpus with best knowledge...")
	if err := BatchSegment(TextFile, CorpusFile, DictCombined); err != nil {
		os.Remove(DictCombined)
		return fmt.Errorf("batch segment failed: %w", err)
	}

	// 6.5 Regenerate Core Dictionary from Corpus (Back-washing)
	log.Println("[5.5/6] Regenerating core dictionary from new corpus...")
	if err := ExtractBaseDictFromCorpus(CorpusFile, DictCore, topBrands); err != nil {
		os.Remove(DictCombined)
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Clean the new core dict to remove noise
	log.Println("[5.6/6] Cleaning regenerated core dictionary...")
	if err := CleanDictionary(DictCore, DictCleanTmp, 0.9); err != nil {
		log.Printf("Warning: core dict clean failed: %v", err)
	} else {
		os.Rename(DictCleanTmp, DictCore)
	}

	// 7. Train CRF
	log.Println("[6/6] Training CRF model (Iter=10)...")
	if err := TrainCRF(CorpusFile, DictCombined, ModelFile, 10); err != nil {
		os.Remove(DictCombined)
		return fmt.Errorf("training failed: %w", err)
	}

	// Cleanup
	os.Remove(DictCombined)
	log.Println("=== Optimization Pipeline Completed ===")
	return nil
}

func ensureFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte(""), 0644)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func mergeNewWords(src, dst string) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)

		isPair := false
		if len(parts) == 2 {
			if _, err := strconv.Atoi(parts[1]); err == nil {
				isPair = true
			}
		}

		if isPair {
			fmt.Fprintln(outFile, line)
		} else {
			for _, w := range parts {
				fmt.Fprintf(outFile, "%s 10000\n", w)
			}
		}
	}
	return scanner.Err()
}

func combineFiles(dst string, srcs ...string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, src := range srcs {
		in, err := os.Open(src)
		if err != nil {
			// If file doesn't exist, just skip it instead of failing early for top/base/auto
			continue
		}

		// For DictTop or any file that might not have frequencies,
		// we should ensure it has a default frequency if merged for segmentation.
		// However, for simplicity, we assume they have frequencies or we handle it in segmenter load.

		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "")
	}
	return nil
}

func loadTopBrands(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var brands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			brands = append(brands, line)
		}
	}
	return brands, scanner.Err()
}
