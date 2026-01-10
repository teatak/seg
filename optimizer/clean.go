package optimizer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/teatak/seg/util"
)

type Word struct {
	Text string
	Freq int
}

// CleanDictionary filters and cleans the dictionary.
func CleanDictionary(inputPath, outputPath string, ratio float64) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	wordMap := make(map[string]int)
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
				word := parts[0]
				if util.ContainsPunctuation(word) {
					continue
				}
				wordMap[word] += freq
			}
		}
	}

	var words []Word
	for t, f := range wordMap {
		words = append(words, Word{Text: t, Freq: f})
	}

	cleanedWords := prunePrefixes(words, ratio)
	// Handle suffixes via reversing
	cleanedWords2 := pruneSuffixes(cleanedWords, ratio)

	// New: Prune Noisy Extensions (e.g., "城希尔顿" when "希尔顿" is much more frequent)
	finalWords := pruneNoisyExtensions(cleanedWords2)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)

	for _, w := range finalWords {
		fmt.Fprintf(writer, "%s %d\n", w.Text, w.Freq)
	}
	return writer.Flush()
}

func pruneNoisyExtensions(words []Word) []Word {
	dict := make(map[string]int)
	for _, w := range words {
		dict[w.Text] = w.Freq
	}

	keep := make([]bool, len(words))
	for i := range keep {
		keep[i] = true
	}

	for i, w := range words {
		runes := []rune(w.Text)
		if len(runes) <= 2 {
			continue
		}

		// Check if stripping one char from front or back results in a MUCH more frequent word
		// 1. Strip Front: "城希尔顿" -> "希尔顿"
		front := string(runes[1:])
		if f, ok := dict[front]; ok {
			if float64(f)/float64(w.Freq) > 5.0 {
				// Safety: don't prune if the extension itself is a common prefix/marker
				// but here we focus on front stripping which is usually noise (like location prefix)
				keep[i] = false
				continue
			}
		}

		// 2. Strip Tail: "希尔顿店" -> "希尔顿"
		// Safety: Protect legitimate endings like "市", "省", "区", "店", "站"
		lastChar := runes[len(runes)-1]
		if isProtectedSuffix(lastChar) {
			continue
		}

		tail := string(runes[:len(runes)-1])
		if f, ok := dict[tail]; ok {
			if float64(f)/float64(w.Freq) > 5.0 {
				keep[i] = false
				continue
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

func isProtectedSuffix(r rune) bool {
	protected := []rune("市省区县店站路里院校园")
	for _, p := range protected {
		if r == p {
			return true
		}
	}
	return false
}

func prunePrefixes(words []Word, ratio float64) []Word {
	sort.Slice(words, func(i, j int) bool {
		return words[i].Text < words[j].Text
	})

	keep := make([]bool, len(words))
	for i := range keep {
		keep[i] = true
	}

	for i := 0; i < len(words)-1; i++ {
		curr := words[i]
		next := words[i+1]

		if strings.HasPrefix(next.Text, curr.Text) {
			score := float64(next.Freq) / float64(curr.Freq)
			if score >= ratio {
				keep[i] = false
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
	for i := range words {
		words[i].Text = reverse(words[i].Text)
	}
	cleaned := prunePrefixes(words, ratio)
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
