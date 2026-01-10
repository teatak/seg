package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Word struct {
	Text string
	Freq int
}

func main() {
	inputPath := flag.String("input", "data/dictionary.txt", "Input dictionary path")
	outputPath := flag.String("output", "data/dictionary_clean.txt", "Output dictionary path")
	ratio := flag.Float64("ratio", 0.9, "Frequency ratio threshold (if Freq(Sub)/Freq(Super) >= ratio, prune Sub)")
	flag.Parse()

	// 1. Load Dictionary
	log.Printf("Loading dictionary from %s...", *inputPath)
	file, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("Failed to open input: %v", err)
	}
	defer file.Close()

	var words []Word
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			freq, _ := strconv.Atoi(parts[1])
			if freq > 0 {
				words = append(words, Word{Text: parts[0], Freq: freq})
			}
		}
	}

	totalBefore := len(words)
	log.Printf("Loaded %d words. Sorting...", totalBefore)

	// 2. Sort words by length descending.
	// This helps us process longer words first (superstrings) effectively?
	// Actually, detecting if A is in B requires comparing.
	// A naive O(N^2) approach with N=500k is too slow (250 billion ops).
	// We need a smarter way or just filter obvious ones.

	// Optimization: Only check if ShortWord is substring of LongWord.
	// Map-based approach:
	// Store all words in a map for O(1) lookup?
	// No, we need to know if "丽怡酒" (A) is part of "丽怡酒店" (B).

	// Since N is large (500k), rigorous O(N^2) is impossible in Go in seconds.
	// BUT, most words don't overlap. We can group by containing characters? Still hard.
	//
	// Let's implement a heuristic:
	// "丽怡酒" (len 3) is likely a substring of a word starting with "丽怡" (len 4+) or ending with "怡酒" (unlikely)
	//
	// Let's Sort by Text to bring "丽怡酒" and "丽怡酒店" close together!
	// If words are sorted alphabetically:
	// ...
	// 丽怡
	// 丽怡酒 (Sub)
	// 丽怡酒店 (Super)
	// ...
	// This captures "Prefix" subtitles very efficiently.
	// What about suffixes? "怡酒店"?
	// "怡酒店" won't be next to "丽怡酒店".

	// Strategy:
	// 1. Clean Prefixes: Sort alphabetically. Check if words[i] is prefix of words[i+1].
	// 2. Clean Suffixes: Reverse strings, sort alphabetically, check prefixes (which are original suffixes).
	// This is O(N log N) and very fast. It effectively cleans "Left Substrings" and "Right Substrings".
	// It won't clean "Middle Substrings" (e.g. "怡酒" inside "丽怡酒店"), but those are rarer in N-gram junk.

	cleanedWords := prunePrefixes(words, *ratio)

	// Now handle suffixes by reversing
	// We reconstruct the list
	cleanedWords2 := pruneSuffixes(cleanedWords, *ratio)

	// 3. Save
	log.Printf("Pruning done. %d -> %d words. (Removed %d)", totalBefore, len(cleanedWords2), totalBefore-len(cleanedWords2))

	outFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create output: %v", err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	// Sort by frequency desc for final output (optional)
	// sort.Slice(cleanedWords2, func(i, j int) bool { return cleanedWords2[i].Freq > cleanedWords2[j].Freq })

	for _, w := range cleanedWords2 {
		fmt.Fprintf(writer, "%s %d\n", w.Text, w.Freq)
	}
	writer.Flush()
	log.Printf("Saved to %s", *outputPath)
}

func prunePrefixes(words []Word, ratio float64) []Word {
	// Sort Alphabetically
	sort.Slice(words, func(i, j int) bool {
		return words[i].Text < words[j].Text
	})

	keep := make([]bool, len(words))
	for i := range keep {
		keep[i] = true
	}

	// Check neighbors
	// A, AB, ABC...
	// If A is prefix of B, and Freq(A) approx Freq(B), drop A.
	for i := 0; i < len(words)-1; i++ {
		curr := words[i]
		next := words[i+1]

		if strings.HasPrefix(next.Text, curr.Text) {
			// Check Frequency
			// Usually garbage substring has freq >= superstring
			// e.g. Freq(丽怡酒) >= Freq(丽怡酒店)
			// IF Freq(丽怡酒店) / Freq(丽怡酒) > ratio (meaning most 丽怡酒 is 丽怡酒店)
			// Wait, if Freq(Super) / Freq(Sub) is close to 1, it implies Sub almost always appears as Super.
			// Example:
			// Super=100, Sub=105. Ratio = 100/105 = 0.95 > 0.9. DROP Sub.
			// Super=10, Sub=100. Ratio = 0.1. KEEP Sub (Sub appears independently).

			score := float64(next.Freq) / float64(curr.Freq)
			if score >= ratio {
				keep[i] = false // Drop current (the shorter one)
				// Propagate the freq of the kept superstring to act as the superstring for previous ones?
				// e.g. A, AB, ABC.
				// A is prefix of AB. Drop A.
				// Now we act as if AB is the one to compare?
				// The loop continues to i+1 (AB). AB is prefix of ABC.
				// This works naturally for chain A -> AB -> ABC.
			}
		}
	}

	var res []Word
	for i, w := range words {
		if keep[i] {
			res = append(res, w)
		}
	}
	return res
}

func pruneSuffixes(words []Word, ratio float64) []Word {
	// Reverse strings
	for i := range words {
		words[i].Text = reverse(words[i].Text)
	}

	// Re-use prunePrefixes logic (which is effectively pruneSuffixes now)
	cleaned := prunePrefixes(words, ratio)

	// Reverse back
	for i := range cleaned {
		cleaned[i].Text = reverse(cleaned[i].Text)
	}
	return cleaned
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
