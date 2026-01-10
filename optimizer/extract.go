package optimizer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// ExtractBaseDictFromCorpus counts word frequencies in the corpus and updates the base dictionary.
func ExtractBaseDictFromCorpus(corpusPath, outputPath string, topBrands []string) error {
	file, err := os.Open(corpusPath)
	if err != nil {
		return err
	}
	defer file.Close()

	counts := make(map[string]int)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		for _, w := range words {
			runes := []rune(w)
			// 1. Skip single characters
			if len(runes) <= 1 {
				continue
			}
			// 2. Skip words without ANY Chinese characters (e.g. pure numbers, English, or alphanumeric "IU7")
			if !hasChinese(w) {
				continue
			}

			// 3. New: Skip "Single Char + Brand" fragments
			if len(runes) > 3 && isNoisyBrandExtension(w, topBrands) {
				continue
			}

			counts[w]++
		}
	}

	type kv struct {
		K string
		V int
	}
	var ss []kv
	for k, v := range counts {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].V != ss[j].V {
			return ss[i].V > ss[j].V
		}
		return ss[i].K < ss[j].K
	})

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, item := range ss {
		fmt.Fprintf(writer, "%s %d\n", item.K, item.V)
	}
	return writer.Flush()
}

func hasChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func isNoisyBrandExtension(s string, brands []string) bool {
	runes := []rune(s)
	if len(runes) < 4 {
		return false
	}
	// Check if s is "One Char + Known Brand"
	suffix := string(runes[1:])
	for _, b := range brands {
		if suffix == b {
			return true
		}
	}
	return false
}
