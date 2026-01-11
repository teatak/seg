package learn

import (
	"math"
	"sort"
	"unicode"
)

// Miner 新词发现器
type Miner struct {
	minFreq    int     // 最小词频
	minMI      float64 // 最小互信息
	minEntropy float64 // 最小左右熵
	maxWordLen int     // 最大词长
}

// Candidate 候选词
type Candidate struct {
	Word         string  `json:"word"`
	Freq         int     `json:"freq"`
	MI           float64 `json:"mi"`            // 互信息
	LeftEntropy  float64 `json:"left_entropy"`  // 左熵
	RightEntropy float64 `json:"right_entropy"` // 右熵
	Score        float64 `json:"score"`         // 综合得分
}

// NewMiner 创建新词发现器
func NewMiner() *Miner {
	return &Miner{
		minFreq:    5,
		minMI:      2.0,
		minEntropy: 0.5,
		maxWordLen: 6,
	}
}

// SetParams 设置参数
func (m *Miner) SetParams(minFreq int, minMI, minEntropy float64, maxWordLen int) {
	m.minFreq = minFreq
	m.minMI = minMI
	m.minEntropy = minEntropy
	m.maxWordLen = maxWordLen
}

// Mine 从语料中挖掘新词
func (m *Miner) Mine(texts []string) []Candidate {
	// 1. 统计 n-gram 词频
	ngramFreq := m.countNgrams(texts)

	// 2. 统计单字频率（用于计算互信息）
	charFreq := m.countChars(texts)
	totalChars := 0
	for _, f := range charFreq {
		totalChars += f
	}

	// 3. 统计左右邻字（用于计算熵）
	leftNeighbors, rightNeighbors := m.countNeighbors(texts)

	// 4. 计算每个候选词的指标
	var candidates []Candidate

	for word, freq := range ngramFreq {
		if freq < m.minFreq {
			continue
		}

		runes := []rune(word)
		if len(runes) < 2 || len(runes) > m.maxWordLen {
			continue
		}

		// 跳过包含非中文字符的词
		if !m.isAllChinese(runes) {
			continue
		}

		// 计算互信息
		mi := m.calcMI(word, freq, charFreq, totalChars, ngramFreq)
		if mi < m.minMI {
			continue
		}

		// 计算左右熵
		leftEntropy := m.calcEntropy(leftNeighbors[word])
		rightEntropy := m.calcEntropy(rightNeighbors[word])

		minEnt := math.Min(leftEntropy, rightEntropy)
		if minEnt < m.minEntropy {
			continue
		}

		// 综合得分
		score := mi * minEnt * math.Log(float64(freq)+1)

		candidates = append(candidates, Candidate{
			Word:         word,
			Freq:         freq,
			MI:           mi,
			LeftEntropy:  leftEntropy,
			RightEntropy: rightEntropy,
			Score:        score,
		})
	}

	// 按得分降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates
}

// countNgrams 统计 n-gram 词频
func (m *Miner) countNgrams(texts []string) map[string]int {
	freq := make(map[string]int)

	for _, text := range texts {
		runes := []rune(text)
		for i := 0; i < len(runes); i++ {
			for length := 2; length <= m.maxWordLen && i+length <= len(runes); length++ {
				word := string(runes[i : i+length])
				freq[word]++
			}
		}
	}

	return freq
}

// countChars 统计单字频率
func (m *Miner) countChars(texts []string) map[rune]int {
	freq := make(map[rune]int)

	for _, text := range texts {
		for _, r := range text {
			freq[r]++
		}
	}

	return freq
}

// countNeighbors 统计左右邻字
func (m *Miner) countNeighbors(texts []string) (left, right map[string]map[rune]int) {
	left = make(map[string]map[rune]int)
	right = make(map[string]map[rune]int)

	for _, text := range texts {
		runes := []rune(text)
		for i := 0; i < len(runes); i++ {
			for length := 2; length <= m.maxWordLen && i+length <= len(runes); length++ {
				word := string(runes[i : i+length])

				// 左邻字
				if i > 0 {
					if left[word] == nil {
						left[word] = make(map[rune]int)
					}
					left[word][runes[i-1]]++
				}

				// 右邻字
				if i+length < len(runes) {
					if right[word] == nil {
						right[word] = make(map[rune]int)
					}
					right[word][runes[i+length]]++
				}
			}
		}
	}

	return left, right
}

// calcMI 计算互信息
// MI(xy) = log2(P(xy) / (P(x) * P(y)))
func (m *Miner) calcMI(word string, freq int, charFreq map[rune]int, totalChars int, ngramFreq map[string]int) float64 {
	runes := []rune(word)
	if len(runes) < 2 {
		return 0
	}

	// P(xy)
	totalNgrams := 0
	for _, f := range ngramFreq {
		totalNgrams += f
	}
	pWord := float64(freq) / float64(totalNgrams)

	// 使用最小切分点的互信息
	minMI := math.MaxFloat64

	for i := 1; i < len(runes); i++ {
		left := string(runes[:i])
		right := string(runes[i:])

		leftFreq := ngramFreq[left]
		rightFreq := ngramFreq[right]

		if leftFreq == 0 || rightFreq == 0 {
			// 使用单字频率
			pLeft := 1.0
			for _, r := range runes[:i] {
				pLeft *= float64(charFreq[r]) / float64(totalChars)
			}
			pRight := 1.0
			for _, r := range runes[i:] {
				pRight *= float64(charFreq[r]) / float64(totalChars)
			}

			mi := math.Log2(pWord / (pLeft * pRight))
			if mi < minMI {
				minMI = mi
			}
		} else {
			pLeft := float64(leftFreq) / float64(totalNgrams)
			pRight := float64(rightFreq) / float64(totalNgrams)

			mi := math.Log2(pWord / (pLeft * pRight))
			if mi < minMI {
				minMI = mi
			}
		}
	}

	if minMI == math.MaxFloat64 {
		return 0
	}
	return minMI
}

// calcEntropy 计算熵
// H = -Σ P(x) * log2(P(x))
func (m *Miner) calcEntropy(neighbors map[rune]int) float64 {
	if len(neighbors) == 0 {
		return 0
	}

	total := 0
	for _, count := range neighbors {
		total += count
	}

	entropy := 0.0
	for _, count := range neighbors {
		p := float64(count) / float64(total)
		entropy -= p * math.Log2(p)
	}

	return entropy
}

// isAllChinese 检查是否全为中文
func (m *Miner) isAllChinese(runes []rune) bool {
	for _, r := range runes {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

// MineWithFilter 挖掘新词并过滤
func (m *Miner) MineWithFilter(texts []string, existingWords map[string]bool) []Candidate {
	candidates := m.Mine(texts)

	// 过滤已存在的词
	var filtered []Candidate
	for _, c := range candidates {
		if !existingWords[c.Word] {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// TopN 获取得分最高的 N 个候选词
func TopN(candidates []Candidate, n int) []Candidate {
	if len(candidates) <= n {
		return candidates
	}
	return candidates[:n]
}
