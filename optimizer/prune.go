package optimizer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PruneInterference removes dictionary words that conflict with user feedback.
func PruneInterference(dictPath, newWordsPath, outputPath string) error {
	dictWords, err := loadDictMap(dictPath)
	if err != nil {
		return err
	}

	feedbackLines := loadLines(newWordsPath)
	toRemove := make(map[string]bool)

	for _, line := range feedbackLines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		fullText := strings.Join(parts, "")
		runesFull := []rune(fullText)

		boundaries := make([]int, 0)
		currentLen := 0
		for i := 0; i < len(parts)-1; i++ {
			currentLen += len([]rune(parts[i]))
			boundaries = append(boundaries, currentLen)
		}

		n := len(runesFull)
		for start := 0; start < n; start++ {
			for end := start + 2; end <= n; end++ {
				subStrRunes := runesFull[start:end]
				subStr := string(subStrRunes)

				isStraddle := false
				for _, b := range boundaries {
					if start < b && b < end {
						isStraddle = true
						break
					}
				}

				if isStraddle {
					if _, exists := dictWords[subStr]; exists {
						toRemove[subStr] = true
					}
				}
			}
		}
	}

	if len(toRemove) == 0 {
		return nil // Nothing to change, but maybe we should ensure output exists?
		// If input==output, we are fine.
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for text, freq := range dictWords {
		if toRemove[text] {
			continue
		}
		fmt.Fprintf(w, "%s %d\n", text, freq)
	}
	return w.Flush()
}

func loadDictMap(path string) (map[string]int, error) {
	m := make(map[string]int)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			freq := 0
			if len(parts) > 1 {
				freq, _ = strconv.Atoi(parts[1])
			}
			m[parts[0]] = freq
		}
	}
	return m, nil
}

func loadLines(path string) []string {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		return lines
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t != "" {
			lines = append(lines, t)
		}
	}
	return lines
}
