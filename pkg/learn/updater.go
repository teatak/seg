package learn

import (
	"sort"

	"github.com/teatak/seg/pkg/dict"
	"github.com/teatak/seg/pkg/seg"
)

// Updater 词典更新器
type Updater struct {
	dictionary    *dict.Dictionary
	feedbackStore *FeedbackStore
	miner         *Miner
	segmenter     *seg.Segmenter

	// 配置
	minFeedbackCount int     // 新词最少需要的反馈次数
	minScore         float64 // 新词发现的最低得分
	decayRate        float64 // 词频衰减率
}

// NewUpdater 创建词典更新器
func NewUpdater(d *dict.Dictionary, fs *FeedbackStore, m *Miner, s *seg.Segmenter) *Updater {
	return &Updater{
		dictionary:       d,
		feedbackStore:    fs,
		miner:            m,
		segmenter:        s,
		minFeedbackCount: 1,
		minScore:         1.0,
		decayRate:        0.99,
	}
}

// SetConfig 设置配置
func (u *Updater) SetConfig(minFeedbackCount int, minScore, decayRate float64) {
	u.minFeedbackCount = minFeedbackCount
	u.minScore = minScore
	u.decayRate = decayRate
}

// UpdateFromFeedback 从用户反馈更新词典
func (u *Updater) UpdateFromFeedback() UpdateResult {
	result := UpdateResult{
		AddedWords:  make([]string, 0),
		UpdatedFreq: make(map[string]int),
	}

	pending := u.feedbackStore.GetPending()
	if len(pending) == 0 {
		return result
	}

	// 统计反馈中的新词出现次数
	wordCount := make(map[string]int)
	for _, fb := range pending {
		for _, word := range fb.NewWords {
			wordCount[word]++
		}
	}

	// 达到阈值的新词加入词典
	var appliedIDs []string
	for word, count := range wordCount {
		if count >= u.minFeedbackCount {
			// 检查是否已存在 (Search using evaluation mode)
			if _, exists := u.dictionary.Search(word, true); !exists {
				u.dictionary.AddWordToStaging(word, count*100) // 初始词频
				result.AddedWords = append(result.AddedWords, word)
			} else {
				u.dictionary.UpdateFreq(word, count*10) // 增加词频
				result.UpdatedFreq[word] = count * 10
			}
		}
	}

	// 标记所有反馈为已应用
	for _, fb := range pending {
		appliedIDs = append(appliedIDs, fb.ID)
	}
	u.feedbackStore.MarkApplied(appliedIDs)

	result.ProcessedFeedbacks = len(pending)
	return result
}

// UpdateFromMiner 从新词发现结果更新词典
func (u *Updater) UpdateFromMiner(candidates []Candidate) UpdateResult {
	result := UpdateResult{
		AddedWords:  make([]string, 0),
		UpdatedFreq: make(map[string]int),
		Candidates:  candidates, // 保存所有候选词
	}

	for _, c := range candidates {
		if c.Score < u.minScore {
			continue
		}

		// 检查是否已存在
		if _, exists := u.dictionary.Search(c.Word, true); !exists {
			freq := int(c.Score * 10)
			if freq < 1 {
				freq = 1
			}
			u.dictionary.AddWordToStaging(c.Word, freq)
			result.AddedWords = append(result.AddedWords, c.Word)
		}
	}

	result.MinedCandidates = len(candidates)
	return result
}

// DecayFrequencies 衰减所有词频（防止过时词语）
func (u *Updater) DecayFrequencies() {
	words := u.dictionary.GetUserWords()
	for _, w := range words {
		newFreq := int(float64(w.Freq) * u.decayRate)
		if newFreq < 1 {
			newFreq = 1
		}
		delta := newFreq - w.Freq
		if delta != 0 {
			u.dictionary.UpdateFreq(w.Word, delta)
		}
	}
}

// SaveDictionary 保存词典
func (u *Updater) SaveDictionary() error {
	return u.dictionary.SaveDicts()
}

// SaveFeedback 保存反馈
func (u *Updater) SaveFeedback() error {
	return u.feedbackStore.Save()
}

// AutoLearn 自动学习（从语料）
func (u *Updater) AutoLearn(texts []string) UpdateResult {
	// 获取所有现有词语（包括基础词典和用户词典）
	existingWords := make(map[string]bool)

	// 获取用户词典中的词
	for _, w := range u.dictionary.GetUserWords() {
		existingWords[w.Word] = true
	}

	// 获取基础词典中的词
	for _, w := range u.dictionary.GetBaseWords() {
		existingWords[w.Word] = true
	}

	// 获取暂存词典中的词
	for _, w := range u.dictionary.GetStagingWords() {
		existingWords[w.Word] = true
	}

	// 挖掘新词（过滤已存在的）
	// 挖掘新词（过滤已存在的）
	candidates := u.miner.MineWithFilter(texts, existingWords)

	// HMM 辅助发现：对文本进行分词，收集会被 HMM 识别出的词
	if u.segmenter != nil {
		hmmCounts := make(map[string]int)
		for _, text := range texts {
			tokens := u.segmenter.Segment(text)
			for _, t := range tokens {
				if t.Type == "hmm" && len([]rune(t.Word)) > 1 {
					hmmCounts[t.Word]++
				}
			}
		}

		// 将 HMM 发现的词也加入候选列表 (如果统计挖掘没发现的话)
		existingCandidates := make(map[string]bool)
		for _, c := range candidates {
			existingCandidates[c.Word] = true
		}

		for word, count := range hmmCounts {
			if !existingCandidates[word] && !existingWords[word] {
				// HMM 发现的词通常可信度较高，给一个基于频率的得分
				score := float64(count) * 2.5 // 给予较高的权重
				candidates = append(candidates, Candidate{
					Word:         word,
					Freq:         count,
					MI:           0, // HMM 结果不依赖 MI
					LeftEntropy:  0, // HMM 结果不依赖熵
					RightEntropy: 0,
					Score:        score,
				})
			}
		}

		// 重新排序
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})
	}

	// 更新词典
	result := u.UpdateFromMiner(candidates)
	result.SourceTexts = len(texts)

	return result
}

// UpdateResult 更新结果
type UpdateResult struct {
	AddedWords         []string       `json:"added_words"`
	UpdatedFreq        map[string]int `json:"updated_freq"`
	ProcessedFeedbacks int            `json:"processed_feedbacks,omitempty"`
	MinedCandidates    int            `json:"mined_candidates,omitempty"`
	SourceTexts        int            `json:"source_texts,omitempty"`
	Candidates         []Candidate    `json:"candidates,omitempty"` // 候选词详情
}

// Summary 获取更新摘要
func (r UpdateResult) Summary() string {
	return ""
}

// AddWordDirectly 直接添加词语
func (u *Updater) AddWordDirectly(word string, freq int) {
	u.dictionary.AddWordToUser(word, freq)
}

// RemoveWord 删除词语
func (u *Updater) RemoveWord(word string) bool {
	return u.dictionary.RemoveWord(word)
}
