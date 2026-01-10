package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
)

func main() {
	inputPath := flag.String("input", "data/text.txt", "Input file path")
	outputPath := flag.String("output", "data/corpus.txt", "Output corpus file path")
	dictPath := flag.String("dict", "data/dictionary.txt", "Dictionary path")
	flag.Parse()

	// 1. Load Dictionary
	dict := dictionary.NewDictionary()
	if err := dict.Load(*dictPath); err != nil {
		log.Printf("Warning: Failed to load dictionary from %s: %v. Using empty dictionary.", *dictPath, err)
	} else {
		log.Printf("Loaded dictionary from %s (Total tokens: %.0f)", *dictPath, dict.Total)
	}

	seg := segmenter.NewSegmenter(dict)

	// 2. Open Files
	inFile, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	// 3. Process
	scanner := bufio.NewScanner(inFile)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// If the line contains tabs (e.g. "Hotel\tAddress"), we might want to treat them as separators or just segment the whole line.
		// Usually for training, we treat the whole line as text.
		// However, splitting by tab might optimize processing if we want separate sentences.
		// Current simple approach: segment the whole line content.

		parts := seg.Cut(line, segmenter.ModeDAG)

		// Write space-separated tokens
		fmt.Fprintln(writer, strings.Join(parts, " "))
		count++
		if count%1000 == 0 {
			log.Printf("Processed %d lines...", count)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file: %v", err)
	}

	writer.Flush()
	log.Printf("Done. Processed %d lines. Saved to %s", count, *outputPath)
}
