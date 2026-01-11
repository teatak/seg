package learn

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// RequestLog 请求日志
type RequestLog struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// RequestLogger 请求记录器
type RequestLogger struct {
	logs     []RequestLog
	filepath string
	mu       sync.RWMutex

	// 配置
	maxLogs       int // 最大日志数量
	mineThreshold int // 触发挖掘的阈值
	lastMineCount int // 上次挖掘时的日志数量
}

// NewRequestLogger 创建请求记录器
func NewRequestLogger(filepath string) *RequestLogger {
	return &RequestLogger{
		filepath:      filepath,
		maxLogs:       1000, // 最多保留 1000 条
		mineThreshold: 100,  // 每 100 条触发一次挖掘
	}
}

// SetConfig 设置配置
func (rl *RequestLogger) SetConfig(maxLogs, mineThreshold int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.maxLogs = maxLogs
	rl.mineThreshold = mineThreshold
}

// Load 加载日志
func (rl *RequestLogger) Load() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.filepath == "" {
		return nil
	}

	data, err := os.ReadFile(rl.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &rl.logs)
}

// Save 保存日志
func (rl *RequestLogger) Save() error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if rl.filepath == "" {
		return nil
	}

	data, err := json.MarshalIndent(rl.logs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(rl.filepath, data, 0644)
}

// Add 添加请求记录，返回是否应该触发挖掘
func (rl *RequestLogger) Add(text string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 过滤太短的文本
	if len([]rune(text)) < 5 {
		return false
	}

	// 添加记录
	rl.logs = append(rl.logs, RequestLog{
		Text:      text,
		Timestamp: time.Now(),
	})

	// 超过最大数量时删除旧记录
	if len(rl.logs) > rl.maxLogs {
		rl.logs = rl.logs[len(rl.logs)-rl.maxLogs:]
	}

	// 检查是否应该触发挖掘
	shouldMine := len(rl.logs)-rl.lastMineCount >= rl.mineThreshold
	return shouldMine
}

// MarkMined 标记已挖掘
func (rl *RequestLogger) MarkMined() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastMineCount = len(rl.logs)
}

// GetTexts 获取所有文本
func (rl *RequestLogger) GetTexts() []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	texts := make([]string, len(rl.logs))
	for i, log := range rl.logs {
		texts[i] = log.Text
	}
	return texts
}

// GetRecentTexts 获取最近 N 条文本
func (rl *RequestLogger) GetRecentTexts(n int) []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	start := 0
	if len(rl.logs) > n {
		start = len(rl.logs) - n
	}

	texts := make([]string, len(rl.logs)-start)
	for i, log := range rl.logs[start:] {
		texts[i] = log.Text
	}
	return texts
}

// Stats 统计信息
func (rl *RequestLogger) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"total":          len(rl.logs),
		"last_mine":      rl.lastMineCount,
		"since_last":     len(rl.logs) - rl.lastMineCount,
		"mine_threshold": rl.mineThreshold,
		"max_logs":       rl.maxLogs,
	}
}

// Clear 清空日志
func (rl *RequestLogger) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.logs = nil
	rl.lastMineCount = 0
}
