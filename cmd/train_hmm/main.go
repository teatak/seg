package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/teatak/seg/pkg/seg"
)

func main() {
	corpusFile := flag.String("corpus", "", "Path to segmented corpus file (e.g. data/corpus.txt)")
	outputFile := flag.String("out", "data/model/hmm.json", "Output model file")
	flag.Parse()

	if *corpusFile == "" {
		fmt.Println("Usage: go run cmd/train_hmm/main.go -corpus data/corpus.txt")
		return
	}

	fmt.Printf("Reading corpus from %s...\n", *corpusFile)
	corpus, err := readCorpus(*corpusFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Training HMM with %d sentences...\n", len(corpus))
	hmm := seg.NewHMM()
	hmm.Train(corpus)

	fmt.Printf("Saving model to %s...\n", *outputFile)
	if err := os.MkdirAll("data/model", 0755); err != nil {
		log.Fatal(err)
	}
	if err := hmm.SaveToFile(*outputFile); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done! You can now restart the server to use this model.")
}

func readCorpus(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var corpus [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Assuming words are separated by spaces
		rawWords := strings.Fields(line)
		var words []string
		for _, w := range rawWords {
			// Strip POS tags if present (e.g., "鸟/o" -> "鸟", "俭汤/ns" -> "俭汤")
			if idx := strings.LastIndex(w, "/"); idx > 0 {
				w = w[:idx]
			}
			// Remove [ ] grouping symbols often found in people's daily corpus
			w = strings.TrimPrefix(w, "[")
			if idx := strings.Index(w, "]"); idx > 0 { // e.g. 中央/n]
				w = w[:idx]
			}

			if w != "" {
				words = append(words, w)
			}
		}

		if len(words) > 0 {
			corpus = append(corpus, words)
		}
	}
	return corpus, scanner.Err()
}
