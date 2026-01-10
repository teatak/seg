package optimizer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
	"github.com/teatak/seg/util"
)

// BatchSegment re-segments the input text file using the provided dictionary to create a training corpus.
func BatchSegment(inputPath, outputPath, dictPath string) error {
	dict := dictionary.NewDictionary()
	if err := dict.Load(dictPath); err != nil {
		log.Printf("Warning: BatchSegment failed to load dictionary (using empty): %v", err)
	}

	seg := segmenter.NewSegmenter(dict)

	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	scanner := bufio.NewScanner(inFile)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := seg.Cut(line, segmenter.ModeDAG)
		var filtered []string
		for _, p := range parts {
			if !util.IsPunctuation(p) {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) > 0 {
			fmt.Fprintln(writer, strings.Join(filtered, " "))
		}
	}
	return writer.Flush()
}

// TrainCRF trains the CRF model using the segmented corpus and optional dictionary words.
func TrainCRF(inputPath, dictPath, outputPath string, iter int) error {
	sentences, err := crf.LoadCorpus(inputPath)
	if err != nil {
		return err
	}

	if dictPath != "" {
		dictSents, err := crf.LoadDictAsCorpus(dictPath)
		if err == nil {
			sentences = append(sentences, dictSents...)
			log.Printf("Added %d words from dictionary to training set.", len(dictSents))
		}
	}

	model := crf.NewModel()

	for it := 1; it <= iter; it++ {
		correctCnt := 0
		totalCnt := 0

		for _, sent := range sentences {
			runes := sent.Runes
			goldTags := sent.Tags

			predTags := model.Decode(runes)

			if len(runes) != len(predTags) {
				continue
			}

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
				continue
			}

			// Update Emissions
			for i := 0; i < len(runes); i++ {
				feats := crf.ExtractFeatures(runes, i)
				gTag := goldTags[i]
				pTag := predTags[i]

				if gTag != pTag {
					for _, f := range feats {
						model.UpdateFeat(f, gTag, 1.0)
						model.UpdateFeat(f, pTag, -1.0)
					}
				}
			}

			// Update Transitions
			for i := 1; i < len(runes); i++ {
				gPrev := goldTags[i-1]
				pPrev := predTags[i-1]
				gCurr := goldTags[i]
				pCurr := predTags[i]

				if gPrev != pPrev || gCurr != pCurr {
					model.Trans[gPrev][gCurr] += 1.0
					model.Trans[pPrev][pCurr] -= 1.0
				}
			}
		}
		// log.Printf("Iteration %d: Accuracy %.2f%%", it, float64(correctCnt)/float64(totalCnt)*100)
	}

	return model.Save(outputPath)
}
