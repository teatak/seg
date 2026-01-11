package learn

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Feedback 用户反馈
type Feedback struct {
	ID           string    `json:"id"`
	OriginalText string    `json:"original_text"` // 原始文本
	OriginalSeg  []string  `json:"original_seg"`  // 原始分词结果
	CorrectedSeg []string  `json:"corrected_seg"` // 纠正后的分词
	NewWords     []string  `json:"new_words"`     // 从纠正中提取的新词
	Timestamp    time.Time `json:"timestamp"`
	Applied      bool      `json:"applied"` // 是否已应用
}

// FeedbackStore 反馈存储
type FeedbackStore struct {
	feedbacks []Feedback
	filepath  string
	mu        sync.RWMutex
}

// NewFeedbackStore 创建反馈存储
func NewFeedbackStore(filepath string) *FeedbackStore {
	return &FeedbackStore{
		filepath: filepath,
	}
}

// Load 加载反馈记录
func (fs *FeedbackStore) Load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.filepath == "" {
		return nil
	}

	data, err := os.ReadFile(fs.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &fs.feedbacks)
}

// Save 保存反馈记录
func (fs *FeedbackStore) Save() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if fs.filepath == "" {
		return nil
	}

	data, err := json.MarshalIndent(fs.feedbacks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.filepath, data, 0644)
}

// MaxFeedbackHistory 最大反馈历史记录数
const MaxFeedbackHistory = 100

// Add 添加反馈
func (fs *FeedbackStore) Add(original string, originalSeg, correctedSeg []string) *Feedback {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 提取新词
	newWords := extractNewWords(originalSeg, correctedSeg)

	feedback := Feedback{
		ID:           generateID(),
		OriginalText: original,
		OriginalSeg:  originalSeg,
		CorrectedSeg: correctedSeg,
		NewWords:     newWords,
		Timestamp:    time.Now(),
		Applied:      false,
	}

	fs.feedbacks = append(fs.feedbacks, feedback)

	// 自动清理：只保留最近 MaxFeedbackHistory 条记录
	if len(fs.feedbacks) > MaxFeedbackHistory {
		// 删除最早的记录
		fs.feedbacks = fs.feedbacks[len(fs.feedbacks)-MaxFeedbackHistory:]
	}

	return &feedback
}

// extractNewWords 从纠正中提取新词
func extractNewWords(original, corrected []string) []string {
	originalSet := make(map[string]bool)
	for _, w := range original {
		originalSet[w] = true
	}

	var newWords []string
	for _, w := range corrected {
		// 超过一个字且原分词中不存在的即为新词
		if len([]rune(w)) > 1 && !originalSet[w] {
			newWords = append(newWords, w)
		}
	}

	return newWords
}

func generateID() string {
	return time.Now().Format("20060102150405.000")
}

// GetPending 获取待应用的反馈
func (fs *FeedbackStore) GetPending() []Feedback {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var pending []Feedback
	for _, fb := range fs.feedbacks {
		if !fb.Applied {
			pending = append(pending, fb)
		}
	}
	return pending
}

// MarkApplied 标记反馈已应用
func (fs *FeedbackStore) MarkApplied(ids []string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for i := range fs.feedbacks {
		if idSet[fs.feedbacks[i].ID] {
			fs.feedbacks[i].Applied = true
		}
	}
}

// GetAll 获取所有反馈
func (fs *FeedbackStore) GetAll() []Feedback {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make([]Feedback, len(fs.feedbacks))
	copy(result, fs.feedbacks)
	return result
}

// Stats 统计信息
func (fs *FeedbackStore) Stats() map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	applied := 0
	totalNewWords := 0
	for _, fb := range fs.feedbacks {
		if fb.Applied {
			applied++
		}
		totalNewWords += len(fb.NewWords)
	}

	return map[string]interface{}{
		"total":           len(fs.feedbacks),
		"applied":         applied,
		"pending":         len(fs.feedbacks) - applied,
		"total_new_words": totalNewWords,
	}
}

// Clear 清空反馈记录
func (fs *FeedbackStore) Clear() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.feedbacks = nil
}
