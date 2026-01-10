package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode"
)

// Common words to bootstrap (optional, but good for base coverage)
var builtins = []string{
	"的", "了", "和", "是", "在", "我", "有", "个", "这个", "那个",
	"酒店", "宾馆", "旅馆", "公寓", "中心", "广场", "大厦", "分店",
	"省", "市", "区", "县", "镇", "乡", "街道", "社区", "村", "组",
	"路", "街", "道", "巷", "弄", "里", "号", "楼", "室", "单元",
	"北京", "上海", "广州", "深圳", "成都", "重庆", "天津", "南京", "武汉", "西安", "杭州", "沈阳",
	"人民", "建设", "解放", "交通", "铁路", "高铁", "火车站", "机场", "地铁", "客运站",
	"店", "站", "部", "局", "院", "所", "校", "园",
	"有限公司", "公司", "大学", "学院", "医院", "银行",
}

func main() {
	inputPath := flag.String("input", "data/text.txt", "Path to the raw input text file")
	outputPath := flag.String("output", "data/dictionary.txt", "Path to save the generated dictionary")
	threshold := flag.Int("threshold", 10, "Minimum frequency for a word to be included")
	maxGram := flag.Int("ngram", 4, "Maximum N-gram length (e.g., 4 means discovering up to 4-character words)")
	flag.Parse()

	// 1. Read input
	log.Printf("Reading raw text from %s...", *inputPath)
	file, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("Failed to open input: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size for long lines
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v", err)
	}

	// 2. Count N-grams
	log.Printf("Counting N-grams (2 to %d)...", *maxGram)
	counts := make(map[string]int)

	for _, line := range lines {
		// Clean and split by non-Chinese characters to get cleaner blocks
		// e.g. "CreateTime: 2023, Name: 张三" -> ["", "张三"]
		blocks := splitToChineseBlocks(line)

		for _, block := range blocks {
			runes := []rune(block)
			n := len(runes)
			if n < 2 {
				continue
			}

			// Sliding window for N-grams
			for i := 0; i < n; i++ {
				for k := 2; k <= *maxGram; k++ {
					if i+k <= n {
						w := string(runes[i : i+k])
						counts[w]++
					}
				}
			}
		}
	}

	// 3. Filter and Collect
	log.Printf("Filtering words with frequency >= %d...", *threshold)
	var newWords []string
	seen := make(map[string]bool)

	// Add builtins first (give them a default high count conceptually, or just ensure they exist)
	for _, w := range builtins {
		if !seen[w] {
			newWords = append(newWords, w)
			seen[w] = true
		}
	}

	// Filter discovered words
	discoveredCount := 0
	for w, c := range counts {
		if c >= *threshold {
			if !seen[w] {
				newWords = append(newWords, w)
				seen[w] = true
				discoveredCount++
			} else {
				// If built-in is found, we might want to update frequency,
				// but here we just list effective words.
			}
		}
	}

	// 4. Save to dictionary
	outFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create dict: %v", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, w := range newWords {
		// Calculate a frequency.
		// Ideally: use actual count from 'counts' map, or a default high value for builtins.
		freq := counts[w]
		if freq == 0 {
			freq = 1000 // Default for builtins not found in text
		}

		// Write format: Word Freq
		fmt.Fprintf(writer, "%s %d\n", w, freq)
	}
	writer.Flush()

	log.Printf("Done! Generated dictionary with %d words (%d built-in + %d discovered).", len(newWords), len(builtins), discoveredCount)
	log.Printf("Saved to %s", *outputPath)
}

// splitToChineseBlocks keeps only continuous Chinese characters as blocks.
// Everything else (punctuation, english, numbers) acts as a separator.
// This prevents discovering "word,word" or "abc汉字" as tokens.
func splitToChineseBlocks(s string) []string {
	var blocks []string
	var current []rune

	for _, r := range s {
		if isChinese(r) {
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

func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
