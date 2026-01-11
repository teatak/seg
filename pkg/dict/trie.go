package dict

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
)

// TrieNode Trie树节点
type TrieNode struct {
	Children map[rune]*TrieNode `json:"children,omitempty"`
	IsEnd    bool               `json:"is_end,omitempty"`
	Freq     int                `json:"freq,omitempty"` // 词频
	Word     string             `json:"word,omitempty"` // 完整词语
}

// Trie 前缀树
type Trie struct {
	Root *TrieNode
	mu   sync.RWMutex
	Size int // 词语数量
}

// Match 匹配结果
type Match struct {
	Word  string
	Start int
	End   int
	Freq  int
}

// NewTrie 创建新的Trie树
func NewTrie() *Trie {
	return &Trie{
		Root: &TrieNode{
			Children: make(map[rune]*TrieNode),
		},
	}
}

// Insert 插入词语
func (t *Trie) Insert(word string, freq int) {
	if word == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.Root
	for _, r := range word {
		if node.Children == nil {
			node.Children = make(map[rune]*TrieNode)
		}
		if _, ok := node.Children[r]; !ok {
			node.Children[r] = &TrieNode{
				Children: make(map[rune]*TrieNode),
			}
		}
		node = node.Children[r]
	}

	if !node.IsEnd {
		t.Size++
	}
	node.IsEnd = true
	node.Freq = freq
	node.Word = word
}

// Search 查找词语是否存在
func (t *Trie) Search(word string) (int, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.Root
	for _, r := range word {
		if node.Children == nil {
			return 0, false
		}
		if _, ok := node.Children[r]; !ok {
			return 0, false
		}
		node = node.Children[r]
	}
	return node.Freq, node.IsEnd
}

// Delete 删除词语
func (t *Trie) Delete(word string) bool {
	if word == "" {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.deleteHelper(t.Root, []rune(word), 0)
}

func (t *Trie) deleteHelper(node *TrieNode, runes []rune, index int) bool {
	if index == len(runes) {
		if !node.IsEnd {
			return false
		}
		node.IsEnd = false
		node.Word = ""
		node.Freq = 0
		t.Size--
		return len(node.Children) == 0
	}

	r := runes[index]
	child, ok := node.Children[r]
	if !ok {
		return false
	}

	shouldDeleteChild := t.deleteHelper(child, runes, index+1)
	if shouldDeleteChild {
		delete(node.Children, r)
		return len(node.Children) == 0 && !node.IsEnd
	}
	return false
}

// PrefixMatch 前缀匹配，返回所有以给定文本开头的词语
func (t *Trie) PrefixMatch(text string) []Match {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var matches []Match
	runes := []rune(text)
	node := t.Root

	for i, r := range runes {
		if node.Children == nil {
			break
		}
		child, ok := node.Children[r]
		if !ok {
			break
		}
		node = child
		if node.IsEnd {
			matches = append(matches, Match{
				Word:  node.Word,
				Start: 0,
				End:   i + 1,
				Freq:  node.Freq,
			})
		}
	}
	return matches
}

// MatchAll 在文本中查找所有匹配的词语
func (t *Trie) MatchAll(text string) []Match {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var matches []Match
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		node := t.Root
		for j := i; j < len(runes); j++ {
			child, ok := node.Children[runes[j]]
			if !ok {
				break
			}
			node = child
			if node.IsEnd {
				matches = append(matches, Match{
					Word:  node.Word,
					Start: i,
					End:   j + 1,
					Freq:  node.Freq,
				})
			}
		}
	}
	return matches
}

// UpdateFreq 更新词频
func (t *Trie) UpdateFreq(word string, delta int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.Root
	for _, r := range word {
		if node.Children == nil {
			return false
		}
		child, ok := node.Children[r]
		if !ok {
			return false
		}
		node = child
	}
	if node.IsEnd {
		node.Freq += delta
		if node.Freq < 0 {
			node.Freq = 0
		}
		return true
	}
	return false
}

// LoadFromFile 从文件加载词典（格式：词语 词频）
func (t *Trie) LoadFromFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		word := parts[0]
		freq := 1
		if len(parts) > 1 {
			if f, err := strconv.Atoi(parts[1]); err == nil {
				freq = f
			}
		}
		t.Insert(word, freq)
	}
	return scanner.Err()
}

// SaveToFile 保存词典到文件
func (t *Trie) SaveToFile(filepath string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	t.traverse(t.Root, "", func(word string, freq int) {
		writer.WriteString(word + " " + strconv.Itoa(freq) + "\n")
	})
	return writer.Flush()
}

func (t *Trie) traverse(node *TrieNode, prefix string, fn func(word string, freq int)) {
	if node.IsEnd {
		fn(node.Word, node.Freq)
	}
	for r, child := range node.Children {
		t.traverse(child, prefix+string(r), fn)
	}
}

// Serialize 序列化为JSON
func (t *Trie) Serialize() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return json.Marshal(t.Root)
}

// Deserialize 从JSON反序列化
func (t *Trie) Deserialize(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	root := &TrieNode{}
	if err := json.Unmarshal(data, root); err != nil {
		return err
	}
	t.Root = root
	t.Size = t.countWords(root)
	return nil
}

func (t *Trie) countWords(node *TrieNode) int {
	count := 0
	if node.IsEnd {
		count = 1
	}
	for _, child := range node.Children {
		count += t.countWords(child)
	}
	return count
}

// GetAllWords 获取所有词语
func (t *Trie) GetAllWords() []struct {
	Word string
	Freq int
} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var words []struct {
		Word string
		Freq int
	}
	t.traverse(t.Root, "", func(word string, freq int) {
		words = append(words, struct {
			Word string
			Freq int
		}{Word: word, Freq: freq})
	})
	return words
}
