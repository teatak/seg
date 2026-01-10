package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/teatak/seg/crf"
)

func main() {
	inputPath := flag.String("input", "", "Path to the segmented corpus file")
	outputPath := flag.String("output", "crf_model.txt", "Path to save the model")
	iter := flag.Int("iter", 5, "Number of training iterations")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Please provide input corpus using -input")
		os.Exit(1)
	}

	sentences, err := loadCorpus(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading corpus: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d sentences.\n", len(sentences))

	model := crf.NewModel()

	for it := 1; it <= *iter; it++ {
		correctCnt := 0
		totalCnt := 0

		for _, sent := range sentences {
			runes := sent.runes
			goldTags := sent.tags

			// 1. Predict
			predTags := model.Decode(runes)

			// 2. Compare and Update
			if len(runes) != len(predTags) {
				// Should not happen
				continue
			}

			// Check if correct
			correct := true
			for i := range goldTags {
				if goldTags[i] != predTags[i] {
					correct = false
					break
				}
			}

			totalCnt++
			if correct {
				correctCnt++
				continue // No update needed if perfectly correct
			}

			// 3. Update Weights (Structured Perceptron)
			// w = w + phi(gold) - phi(pred)

			// Update Emissions
			for i := 0; i < len(runes); i++ {
				feats := crf.ExtractFeatures(runes, i)
				gTag := goldTags[i]
				pTag := predTags[i]

				if gTag != pTag {
					for _, f := range features(feats) {
						updateFeat(model, f, gTag, 1.0)
						updateFeat(model, f, pTag, -1.0)
					}
				}
			}

			// Update Transitions
			for i := 0; i < len(runes); i++ {
				var gPrev, pPrev int
				if i == 0 {
					// Usually we might have a Start tag, but simplified model assumes implicit start.
					// Or we just don't update transition for first char (effectively start transition fixed or ignored).
					// Our Decode impl used dp[0][tag] = emission, implying 0 transition score from start.
					// So we only update transitions for i > 0.
					continue
				} else {
					gPrev = goldTags[i-1]
					pPrev = predTags[i-1]
				}

				gCurr := goldTags[i]
				pCurr := predTags[i]

				if gPrev != pPrev || gCurr != pCurr {
					model.Trans[gPrev][gCurr] += 1.0
					model.Trans[pPrev][pCurr] -= 1.0
				}
			}
		}
		fmt.Printf("Iteration %d: Accuracy %.2f%%\n", it, float64(correctCnt)/float64(totalCnt)*100)
	}

	err = saveModel(model, *outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving model: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Model saved to %s\n", *outputPath)
}

type sentence struct {
	runes []rune
	tags  []int
}

func loadCorpus(path string) ([]sentence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []sentence
	scanner := bufio.NewScanner(file)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		words := strings.Fields(line)
		var runes []rune
		var tags []int
		for _, word := range words {
			wRunes := []rune(word)
			if len(wRunes) == 0 {
				continue
			}
			runes = append(runes, wRunes...)
			if len(wRunes) == 1 {
				tags = append(tags, crf.TagS)
			} else {
				tags = append(tags, crf.TagB)
				for k := 0; k < len(wRunes)-2; k++ {
					tags = append(tags, crf.TagM)
				}
				tags = append(tags, crf.TagE)
			}
		}
		if len(runes) > 0 {
			data = append(data, sentence{runes, tags})
		}
	}
	return data, scanner.Err()
}

func feature(fs []string) []string {
	return fs
}

// simple wrapper
func features(fs []string) []string { return fs }

func updateFeat(m *crf.Model, feat string, tag int, delta float64) {
	if m.Feats[feat] == nil {
		m.Feats[feat] = make(map[int]float64)
	}
	m.Feats[feat][tag] += delta
	if m.Feats[feat][tag] == 0 {
		delete(m.Feats[feat], tag)
		if len(m.Feats[feat]) == 0 {
			delete(m.Feats, feat)
		}
	}
}

func saveModel(m *crf.Model, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)

	// Save Transitions
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if m.Trans[i][j] != 0 {
				fmt.Fprintf(writer, "T %s %s %f\n", tagStr(i), tagStr(j), m.Trans[i][j])
			}
		}
	}

	// Save Features
	for feat, weights := range m.Feats {
		for tag, w := range weights {
			if w != 0 {
				fmt.Fprintf(writer, "F %s %s %f\n", feat, tagStr(tag), w)
			}
		}
	}
	return writer.Flush()
}

func tagStr(t int) string {
	switch t {
	case crf.TagB:
		return "B"
	case crf.TagM:
		return "M"
	case crf.TagE:
		return "E"
	case crf.TagS:
		return "S"
	}
	return "?"
}
