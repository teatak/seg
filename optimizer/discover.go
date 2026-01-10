package optimizer

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

// Discover finds common N-grams in a text file and returns them as a dictionary.
func Discover(inputPath, outputPath string, threshold, maxGram int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	counts := make(map[string]int)
	for scanner.Scan() {
		line := scanner.Text()
		blocks := splitToBlocks(line)

		for _, block := range blocks {
			runes := []rune(block)
			n := len(runes)
			if n < 2 {
				continue
			}

			for i := 0; i < n; i++ {
				for k := 2; k <= maxGram; k++ {
					if i+k <= n {
						w := string(runes[i : i+k])
						counts[w]++
					}
				}
			}
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for w, c := range counts {
		if c >= threshold {
			fmt.Fprintf(writer, "%s %d\n", w, c)
		}
	}
	return writer.Flush()
}

func splitToBlocks(s string) []string {
	var blocks []string
	var current []rune

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				blocks = append(blocks, string(current))
				current = nil
			}
		}
	}
	if len(current) > 0 {
		blocks = append(blocks, string(current))
	}
	return blocks
}
