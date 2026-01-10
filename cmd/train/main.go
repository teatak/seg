package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	inputPath := flag.String("input", "", "Path to the segmented corpus file (space separated)")
	outputPath := flag.String("output", "dictionary.txt", "Path to save the generated dictionary")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Please provide an input file using -input flag")
		os.Exit(1)
	}

	fmt.Printf("Reading corpus from %s...\n", *inputPath)
	file, err := os.Open(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	counts := make(map[string]int)
	scanner := bufio.NewScanner(file)

	// Set buffer size to handle potentially long lines
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		words := strings.Fields(line)
		for _, word := range words {
			word = strings.TrimSpace(word)
			if word != "" {
				counts[word]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d unique words. Generating dictionary...\n", len(counts))

	// Sort by frequency descending for better readability (optional)
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range counts {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	outFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, kv := range ss {
		_, err := fmt.Fprintf(writer, "%s %d\n", kv.Key, kv.Value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
	}
	writer.Flush()

	fmt.Printf("Dictionary saved to %s\n", *outputPath)
}
