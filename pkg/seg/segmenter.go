package seg

import (
	"sort"
	"unicode"

	"github.com/teatak/seg/pkg/dict"
)

// Token 分词结果
type Token struct {
	Word  string `json:"word"`
	Start int    `json:"start"` // 字符位置
	End   int    `json:"end"`
	Type  string `json:"type"` // word, single, hmm
}

// Segmenter 分词器
type Segmenter struct {
	dict       *dict.Dictionary
	hmm        *HMM
	useStaging bool // 是否使用暂存词典
}

// NewSegmenter 创建分词器
func NewSegmenter(d *dict.Dictionary, hmm *HMM, useStaging bool) *Segmenter {
	return &Segmenter{
		dict:       d,
		hmm:        hmm,
		useStaging: useStaging,
	}
}

// Segment 分词（使用双向最大匹配 + HMM）
func (s *Segmenter) Segment(text string) []Token {
	if text == "" {
		return nil
	}

	// 预处理：按标点符号分割
	blocks := s.splitByPunctuation(text)

	var result []Token
	offset := 0

	for _, block := range blocks {
		if block.isPunct {
			// 标点符号直接输出
			result = append(result, Token{
				Word:  block.text,
				Start: offset,
				End:   offset + len([]rune(block.text)),
				Type:  "punct",
			})
		} else if block.isAlphaNum {
			// 英文/数字直接作为整体输出
			result = append(result, Token{
				Word:  block.text,
				Start: offset,
				End:   offset + len([]rune(block.text)),
				Type:  "alphanum",
			})
		} else {
			// 对中文文本进行分词
			tokens := s.segmentBlock(block.text, offset)
			result = append(result, tokens...)
		}
		offset += len([]rune(block.text))
	}

	return result
}

type textBlock struct {
	text       string
	isPunct    bool
	isAlphaNum bool // 是否为英文/数字块
}

func (s *Segmenter) splitByPunctuation(text string) []textBlock {
	var blocks []textBlock
	runes := []rune(text)
	i := 0

	for i < len(runes) {
		r := runes[i]

		// 1. 标点、符号、空格 -> 直接切分
		if unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r) {
			blocks = append(blocks, textBlock{text: string(r), isPunct: true})
			i++
			continue
		}

		// 2. 收集连续的内容 (直到遇到标点/符号/空格)
		start := i
		hasChinese := false
		for i < len(runes) {
			curr := runes[i]
			if unicode.IsPunct(curr) || unicode.IsSymbol(curr) || unicode.IsSpace(curr) {
				break
			}
			if unicode.Is(unicode.Han, curr) {
				hasChinese = true
			}
			i++
		}

		content := string(runes[start:i])

		// 3. 判断类型
		if !hasChinese {
			// 纯英文/数字 -> 标记为 isAlphaNum (不走分词，作为整体)
			// 注意：这里简单判定非中文且非标点即为 AlphaNum，
			// 实际上已经由外层循环过滤了标点，所以这里就是纯英文/数字/其他非中文文字
			blocks = append(blocks, textBlock{text: content, isPunct: false, isAlphaNum: true})
		} else {
			// 包含中文 (如 "7天", "T恤", "我们") -> 走分词逻辑
			blocks = append(blocks, textBlock{text: content, isPunct: false, isAlphaNum: false})
		}
	}

	return blocks
}

// isEnglishLetter 判断是否为英文字母
func isEnglishLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func (s *Segmenter) segmentBlock(text string, offset int) []Token {
	// 双向最大匹配
	forward := s.forwardMaxMatch(text, offset)
	backward := s.backwardMaxMatch(text, offset)

	// 选择最优结果
	tokens := s.chooseBest(forward, backward)

	// 对未登录词使用 HMM
	if s.hmm != nil {
		tokens = s.processUnknown(text, tokens, offset)
	}

	return tokens
}

// forwardMaxMatch 正向最大匹配
func (s *Segmenter) forwardMaxMatch(text string, offset int) []Token {
	var tokens []Token
	runes := []rune(text)
	pos := 0

	for pos < len(runes) {
		matched := false
		// 从最长开始尝试匹配
		for length := min(len(runes)-pos, 10); length > 0; length-- {
			word := string(runes[pos : pos+length])
			if freq, ok := s.dict.Search(word, s.useStaging); ok {
				tokens = append(tokens, Token{
					Word:  word,
					Start: offset + pos,
					End:   offset + pos + length,
					Type:  "word",
				})
				_ = freq
				pos += length
				matched = true
				break
			}
		}

		if !matched {
			// 单字
			tokens = append(tokens, Token{
				Word:  string(runes[pos]),
				Start: offset + pos,
				End:   offset + pos + 1,
				Type:  "single",
			})
			pos++
		}
	}

	return tokens
}

// backwardMaxMatch 逆向最大匹配
func (s *Segmenter) backwardMaxMatch(text string, offset int) []Token {
	var tokens []Token
	runes := []rune(text)
	pos := len(runes)

	for pos > 0 {
		matched := false
		// 从最长开始尝试匹配
		for length := min(pos, 10); length > 0; length-- {
			start := pos - length
			word := string(runes[start:pos])
			if freq, ok := s.dict.Search(word, s.useStaging); ok {
				tokens = append(tokens, Token{
					Word:  word,
					Start: offset + start,
					End:   offset + pos,
					Type:  "word",
				})
				_ = freq
				pos = start
				matched = true
				break
			}
		}

		if !matched {
			pos--
			tokens = append(tokens, Token{
				Word:  string(runes[pos]),
				Start: offset + pos,
				End:   offset + pos + 1,
				Type:  "single",
			})
		}
	}

	// 逆序
	for i, j := 0, len(tokens)-1; i < j; i, j = i+1, j-1 {
		tokens[i], tokens[j] = tokens[j], tokens[i]
	}

	return tokens
}

// chooseBest 选择最优分词结果
func (s *Segmenter) chooseBest(forward, backward []Token) []Token {
	// 规则：
	// 1. 词数越少越好
	// 2. 词数相同时，单字词越少越好
	// 3. 都相同时，选择正向结果

	fwdSingles := countSingles(forward)
	bwdSingles := countSingles(backward)

	if len(forward) < len(backward) {
		return forward
	}
	if len(backward) < len(forward) {
		return backward
	}
	if fwdSingles <= bwdSingles {
		return forward
	}
	return backward
}

func countSingles(tokens []Token) int {
	count := 0
	for _, t := range tokens {
		if t.Type == "single" {
			count++
		}
	}
	return count
}

// processUnknown 处理未登录词（单字序列）
func (s *Segmenter) processUnknown(text string, tokens []Token, offset int) []Token {
	if s.hmm == nil {
		return tokens
	}

	var result []Token
	var singleBuffer []Token

	flushSingles := func() {
		if len(singleBuffer) == 0 {
			return
		}

		// 构建单字序列
		var chars []rune
		for _, t := range singleBuffer {
			chars = append(chars, []rune(t.Word)...)
		}

		// 使用 HMM 分词
		hmmTokens := s.hmm.Segment(string(chars), singleBuffer[0].Start)
		result = append(result, hmmTokens...)
		singleBuffer = nil
	}

	for _, t := range tokens {
		if t.Type == "single" {
			singleBuffer = append(singleBuffer, t)
		} else {
			flushSingles()
			result = append(result, t)
		}
	}
	flushSingles()

	return result
}

// SegmentToStrings 分词并返回字符串切片
func (s *Segmenter) SegmentToStrings(text string) []string {
	tokens := s.Segment(text)
	words := make([]string, len(tokens))
	for i, t := range tokens {
		words[i] = t.Word
	}
	return words
}

// SegmentSearch 搜索模式分词（返回所有可能的分词组合用于搜索引擎）
func (s *Segmenter) SegmentSearch(text string) []Token {
	tokens := s.Segment(text)

	// 添加所有匹配的词
	matches := s.dict.MatchAll(text, s.useStaging)

	// 合并去重
	seen := make(map[string]bool)
	var result []Token

	for _, t := range tokens {
		key := t.Word + ":" + string(rune(t.Start))
		if !seen[key] {
			seen[key] = true
			result = append(result, t)
		}
	}

	for _, m := range matches {
		key := m.Word + ":" + string(rune(m.Start))
		if !seen[key] {
			seen[key] = true
			result = append(result, Token{
				Word:  m.Word,
				Start: m.Start,
				End:   m.End,
				Type:  "search",
			})
		}
	}

	// 按位置排序
	sort.Slice(result, func(i, j int) bool {
		if result[i].Start != result[j].Start {
			return result[i].Start < result[j].Start
		}
		return result[i].End < result[j].End
	})

	return result
}
