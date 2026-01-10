package segmenter

import (
	"math"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
)

// Mode defines the segmentation mode.
type Mode int

const (
	ModeDAG    Mode = iota // ModeDAG uses dictionary-based DAG segmentation.
	ModeCRF                // ModeCRF uses pure CRF model-based segmentation.
	ModeHybrid             // ModeHybrid uses a hybrid approach: Dictionary-first, then CRF for OOV.
)

// Segmenter handles the text segmentation.
type Segmenter struct {
	Dict     *dictionary.Dictionary
	CRFModel *crf.Model
}

// NewSegmenter creates a new segmenter with the given dictionary.
func NewSegmenter(dict *dictionary.Dictionary) *Segmenter {
	return &Segmenter{Dict: dict}
}

// Cut segments the text into a slice of strings using the specified mode (defaults to ModeDAG).
func (s *Segmenter) Cut(text string, modes ...Mode) []string {
	mode := ModeDAG
	if len(modes) > 0 {
		mode = modes[0]
	}

	runes := []rune(text)
	blocks := splitTextToBlocks(runes)
	var result []string
	for _, block := range blocks {
		if block.isPureAlphaNum {
			result = append(result, string(block.runes))
		} else {
			switch mode {
			case ModeCRF:
				if s.CRFModel != nil {
					result = append(result, s.cutCRF(string(block.runes))...)
				} else {
					// Fallback if no model
					result = append(result, s.cutDAG(string(block.runes))...)
				}
			case ModeHybrid:
				if s.CRFModel != nil {
					result = append(result, s.cutHybrid(string(block.runes))...)
				} else {
					result = append(result, s.cutDAG(string(block.runes))...)
				}
			default:
				result = append(result, s.cutDAG(string(block.runes))...)
			}
		}
	}
	return result
}

// cutDAG implements the original DAG based segmentation
func (s *Segmenter) cutDAG(text string) []string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return []string{}
	}

	// 1. Build DAG
	// dag[i] contains a list of end indices (inclusive) for words starting at i
	dag := make([][]int, n)
	for i := 0; i < n; i++ {
		dag[i] = []int{}
		// Look ahead to find words
		for j := i; j < n; j++ {
			if j-i+1 > s.Dict.MaxLen {
				break
			}
			word := string(runes[i : j+1])
			if s.Dict.Contains(word) {
				dag[i] = append(dag[i], j)
			}
		}

		// Handle alphanumeric sequences: match the whole sequence as a candidate
		// even if it's not in the dictionary, to avoid splitting numbers like "25" or "PKU".
		if isAlphaNum(runes[i]) {
			j := i
			for j < n && isAlphaNum(runes[j]) {
				j++
			}
			// Check if this end index is already in dag
			found := false
			for _, end := range dag[i] {
				if end == j-1 {
					found = true
					break
				}
			}
			if !found {
				dag[i] = append(dag[i], j-1)
			}
		}

		// If no word found, at least the single character is a candidate
		if len(dag[i]) == 0 {
			dag[i] = append(dag[i], i)
		}
	}

	// 2. Dynamic Programming for Max Probability Path
	type routeNode struct {
		prob float64
		end  int
	}
	route := make([]routeNode, n+1)
	route[n] = routeNode{prob: 0, end: 0}

	for i := n - 1; i >= 0; i-- {
		bestProb := -math.MaxFloat64
		bestEnd := i

		found := false
		for _, end := range dag[i] {
			word := string(runes[i : end+1])
			prob := s.Dict.LogProbability(word) + route[end+1].prob
			if prob > bestProb {
				bestProb = prob
				bestEnd = end
				found = true
			}
		}

		if !found {
			bestProb = -20.0 + route[i+1].prob
			bestEnd = i
		}

		route[i] = routeNode{prob: bestProb, end: bestEnd}
	}

	// 3. Backtrack to generating result
	var result []string
	idx := 0
	for idx < n {
		end := route[idx].end
		word := string(runes[idx : end+1])
		result = append(result, word)
		idx = end + 1
	}

	return result
}

// CutSearch segments the text into a slice of strings, including fine-grained sub-words, using the specified mode (defaults to ModeDAG).
// Typical usage: for search engine indexing.
func (s *Segmenter) CutSearch(text string, modes ...Mode) []string {
	result := []string{}
	defaultSegs := s.Cut(text, modes...)

	for _, word := range defaultSegs {
		s.addSubWords(word, &result)
		result = append(result, word)
	}
	return result
}

func (s *Segmenter) addSubWords(word string, result *[]string) {
	runes := []rune(word)
	if len(runes) <= 2 {
		return
	}

	// 英文或数字单词不进行子词切分 (如 PKU 不要切出 P/K/U)
	isPure := true
	for _, r := range runes {
		if !isAlphaNum(r) {
			isPure = false
			break
		}
	}
	if isPure {
		return
	}

	for i := 0; i < len(runes); i++ {
		for j := i + 1; j <= len(runes); j++ {
			subWord := string(runes[i:j])
			if subWord != word && s.Dict.Contains(subWord) {
				*result = append(*result, subWord)
			}
		}
	}
}

// cutHybrid segments the text using a hybrid approach: Dictionary-first, then CRF for OOV.
func (s *Segmenter) cutHybrid(text string) []string {
	if s.CRFModel == nil {
		return s.cutDAG(text) // Fallback to DAG if no model
	}

	// Hybrid Strategy:
	// 1. Identify words that are definitely in the dictionary (Trust high freq words).
	// 2. Use CRF only for the gaps (Unknown segments).

	dagTokens := s.cutDAG(text)

	var result []string
	var buf []rune

	flushBuf := func() {
		if len(buf) == 0 {
			return
		}
		predictions := s.decodeCRFBlock(buf)
		result = append(result, predictions...)
		buf = []rune{}
	}

	for _, token := range dagTokens {
		r := []rune(token)
		if len(r) > 1 {
			// Trusted long word
			flushBuf()
			result = append(result, token)
		} else {
			// Single char - potential OOV part
			buf = append(buf, r...)
		}
	}
	flushBuf()

	return result
}

// cutCRF segments the text using pure CRF model-based segmentation.
func (s *Segmenter) cutCRF(text string) []string {
	runes := []rune(text)
	return s.decodeCRFBlock(runes)
}

// Low-level CRF decoder for a run of text
func (s *Segmenter) decodeCRFBlock(runes []rune) []string {
	if len(runes) == 0 {
		return nil
	}
	tags := s.CRFModel.Decode(runes)
	var res []string
	var buf []rune
	for i, tag := range tags {
		char := runes[i]
		switch tag {
		case crf.TagB:
			if len(buf) > 0 {
				res = append(res, string(buf))
				buf = []rune{}
			}
			buf = append(buf, char)
		case crf.TagM:
			buf = append(buf, char)
		case crf.TagE:
			buf = append(buf, char)
			res = append(res, string(buf))
			buf = []rune{}
		case crf.TagS:
			if len(buf) > 0 {
				res = append(res, string(buf))
				res = append(res, string(buf))
				buf = []rune{}
			}
			res = append(res, string(char))
		}
	}
	if len(buf) > 0 {
		res = append(res, string(buf))
	}
	return res
}
