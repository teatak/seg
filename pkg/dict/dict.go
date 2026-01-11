package dict

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Dictionary 词典管理器
type Dictionary struct {
	baseTrie    *Trie // 基础词典
	userTrie    *Trie // 用户自定义词典
	stagingTrie *Trie // 暂存词典 (用于新词发现)
	mu          sync.RWMutex
	baseFile    string // 基础词典文件路径
	userFile    string // 用户词典文件路径
	stagingFile string // 暂存词典文件路径
	initialized bool
}

// NewDictionary 创建词典管理器
func NewDictionary(baseFile, userFile, stagingFile string) *Dictionary {
	return &Dictionary{
		baseTrie:    NewTrie(),
		userTrie:    NewTrie(),
		stagingTrie: NewTrie(),
		baseFile:    baseFile,
		userFile:    userFile,
		stagingFile: stagingFile,
	}
}

// Load 加载词典
func (d *Dictionary) Load() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 加载基础词典
	if d.baseFile != "" {
		if _, err := os.Stat(d.baseFile); err == nil {
			if err := d.baseTrie.LoadFromFile(d.baseFile); err != nil {
				return err
			}
		}
	}

	// 加载用户词典
	if d.userFile != "" {
		if _, err := os.Stat(d.userFile); err == nil {
			if err := d.userTrie.LoadFromFile(d.userFile); err != nil {
				return err
			}
		}
	}

	// 加载暂存词典
	if d.stagingFile != "" {
		if _, err := os.Stat(d.stagingFile); err == nil {
			if err := d.stagingTrie.LoadFromFile(d.stagingFile); err != nil {
				return err
			}
		}
	}

	d.initialized = true
	return nil
}

// Search 查找词语（优先用户词典，其次暂存词典）
// useStaging: 是否在暂存词典中查找
func (d *Dictionary) Search(word string, useStaging bool) (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 1. 用户词典
	if freq, ok := d.userTrie.Search(word); ok {
		return freq, true
	}
	// 2. 暂存词典 (如果启用)
	if useStaging {
		if freq, ok := d.stagingTrie.Search(word); ok {
			return freq, true
		}
	}
	// 3. 基础词典
	return d.baseTrie.Search(word)
}

// PrefixMatch 前缀匹配
func (d *Dictionary) PrefixMatch(text string, useStaging bool) []Match {
	d.mu.RLock()
	defer d.mu.RUnlock()

	baseMatches := d.baseTrie.PrefixMatch(text)
	userMatches := d.userTrie.PrefixMatch(text)

	// 后面的覆盖前面的
	matchMap := make(map[string]Match)
	for _, m := range baseMatches {
		matchMap[m.Word] = m
	}
	// 暂存词典
	if useStaging {
		stagingMatches := d.stagingTrie.PrefixMatch(text)
		for _, m := range stagingMatches {
			matchMap[m.Word] = Match{
				Word:  m.Word,
				Start: m.Start,
				End:   m.End,
				Freq:  m.Freq + 5000, // 暂存词典权重提升，但低于用户词典
			}
		}
	}
	// 用户词典优先级最高
	for _, m := range userMatches {
		matchMap[m.Word] = Match{
			Word:  m.Word,
			Start: m.Start,
			End:   m.End,
			Freq:  m.Freq + 10000,
		}
	}

	result := make([]Match, 0, len(matchMap))
	for _, m := range matchMap {
		result = append(result, m)
	}
	return result
}

// MatchAll 在文本中查找所有匹配
func (d *Dictionary) MatchAll(text string, useStaging bool) []Match {
	d.mu.RLock()
	defer d.mu.RUnlock()

	matchMap := make(map[string]Match)
	key := func(m Match) string {
		return m.Word + string(rune(m.Start)) + string(rune(m.End))
	}

	// 基础词典
	for _, m := range d.baseTrie.MatchAll(text) {
		matchMap[key(m)] = m
	}

	// 暂存词典
	if useStaging {
		for _, m := range d.stagingTrie.MatchAll(text) {
			m.Freq += 5000
			matchMap[key(m)] = m
		}
	}

	// 用户词典
	for _, m := range d.userTrie.MatchAll(text) {
		m.Freq += 10000
		matchMap[key(m)] = m
	}

	result := make([]Match, 0, len(matchMap))
	for _, m := range matchMap {
		result = append(result, m)
	}
	return result
}

// AddWordToStaging 添加新词到暂存词典
func (d *Dictionary) AddWordToStaging(word string, freq int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stagingTrie.Insert(word, freq)
}

// AddWordToUser 添加新词到用户词典
func (d *Dictionary) AddWordToUser(word string, freq int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.userTrie.Insert(word, freq)
}

// RemoveWord 删除词语 (同时尝试从用户词典和暂存词典删除)
func (d *Dictionary) RemoveWord(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	deletedUser := d.userTrie.Delete(word)
	deletedStaging := d.stagingTrie.Delete(word)
	return deletedUser || deletedStaging
}

// MergeStagingToUser 合并暂存词典到用户词典
func (d *Dictionary) MergeStagingToUser() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	stagingWords := d.stagingTrie.GetAllWords()
	for _, w := range stagingWords {
		d.userTrie.Insert(w.Word, w.Freq)
	}

	// 清空暂存词典
	d.stagingTrie = NewTrie()

	// 保存用户词典
	if err := d.saveTrie(d.userTrie, d.userFile); err != nil {
		return err
	}
	// 保存（清空）暂存词典
	return d.saveTrie(d.stagingTrie, d.stagingFile)
}

// UpdateFreq 更新词频 (在用户或暂存词典中更新)
func (d *Dictionary) UpdateFreq(word string, delta int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 优先用户词典
	if d.userTrie.UpdateFreq(word, delta) {
		return true
	}
	// 2. 暂存词典
	if d.stagingTrie.UpdateFreq(word, delta) {
		return true
	}
	// 3. 基础词典有词 -> 复制到暂存词典并更新
	if freq, ok := d.baseTrie.Search(word); ok {
		d.stagingTrie.Insert(word, freq+delta)
		return true
	}
	return false
}

// SaveDicts 保存变动词典 (用户 + 暂存)
func (d *Dictionary) SaveDicts() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if err := d.saveTrie(d.userTrie, d.userFile); err != nil {
		return err
	}
	return d.saveTrie(d.stagingTrie, d.stagingFile)
}

func (d *Dictionary) saveTrie(t *Trie, path string) error {
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return t.SaveToFile(path)
}

// Stats 返回词典统计信息
func (d *Dictionary) Stats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"base_words":    d.baseTrie.Size,
		"user_words":    d.userTrie.Size,
		"staging_words": d.stagingTrie.Size,
		"total":         d.baseTrie.Size + d.userTrie.Size + d.stagingTrie.Size,
		"base_file":     d.baseFile,
		"user_file":     d.userFile,
		"staging_file":  d.stagingFile,
	}
}

// GetUserWords 获取所有用户词语
func (d *Dictionary) GetUserWords() []struct {
	Word string
	Freq int
} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.userTrie.GetAllWords()
}

// GetStagingWords 获取所有暂存词语
func (d *Dictionary) GetStagingWords() []struct {
	Word string
	Freq int
} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stagingTrie.GetAllWords()
}

// GetBaseWords 获取所有基础词语
func (d *Dictionary) GetBaseWords() []struct {
	Word string
	Freq int
} {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.baseTrie.GetAllWords()
}

// Export 导出词典为JSON
func (d *Dictionary) Export() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	data := struct {
		BaseWords []struct {
			Word string `json:"word"`
			Freq int    `json:"freq"`
		} `json:"base_words"`
		UserWords []struct {
			Word string `json:"word"`
			Freq int    `json:"freq"`
		} `json:"user_words"`
		StagingWords []struct {
			Word string `json:"word"`
			Freq int    `json:"freq"`
		} `json:"staging_words"`
	}{}

	toStruct := func(w string, f int) struct {
		Word string `json:"word"`
		Freq int    `json:"freq"`
	} {
		return struct {
			Word string `json:"word"`
			Freq int    `json:"freq"`
		}{Word: w, Freq: f}
	}

	for _, w := range d.baseTrie.GetAllWords() {
		data.BaseWords = append(data.BaseWords, toStruct(w.Word, w.Freq))
	}
	for _, w := range d.userTrie.GetAllWords() {
		data.UserWords = append(data.UserWords, toStruct(w.Word, w.Freq))
	}
	for _, w := range d.stagingTrie.GetAllWords() {
		data.StagingWords = append(data.StagingWords, toStruct(w.Word, w.Freq))
	}

	return json.MarshalIndent(data, "", "  ")
}
