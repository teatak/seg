package crf

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/teatak/seg/util"
)

// Tag constants
const (
	TagB = 0 // Begin
	TagM = 1 // Middle
	TagE = 2 // End
	TagS = 3 // Single
)

// Model represents a Linear Chain CRF model.
type Model struct {
	// Trans[from][to] = weight
	Trans [4][4]float64
	// Feats[feature_string][label_id] = weight
	// feature_string typically "U02:Char" or similar
	Feats map[string]map[int]float64
}

// NewModel creates a new empty model.
func NewModel() *Model {
	return &Model{
		Feats: make(map[string]map[int]float64),
	}
}

// Load loads a simple text-based CRF model.
// Format lines:
// T from_tag to_tag weight
// F feature_string tag weight
func (m *Model) Load(path string) error {
	file, err := os.Open(path)
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
		if len(parts) < 4 {
			continue
		}

		kind := parts[0]
		if kind == "T" {
			from := parseTag(parts[1])
			to := parseTag(parts[2])
			weight, err := strconv.ParseFloat(parts[3], 64)
			if err == nil && from >= 0 && to >= 0 {
				m.Trans[from][to] = weight
			}
		} else if kind == "F" {
			featStr := parts[1]
			tag := parseTag(parts[2])
			weight, err := strconv.ParseFloat(parts[3], 64)
			if err == nil && tag >= 0 {
				if m.Feats[featStr] == nil {
					m.Feats[featStr] = make(map[int]float64)
				}
				m.Feats[featStr][tag] = weight
			}
		}
	}
	return scanner.Err()
}

// Save saves the model to a file.
func (m *Model) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)

	// Save Transitions
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if m.Trans[i][j] != 0 {
				fmt.Fprintf(writer, "T %s %s %f\n", TagStr(i), TagStr(j), m.Trans[i][j])
			}
		}
	}

	// Save Features
	for feat, weights := range m.Feats {
		for tag, w := range weights {
			if w != 0 {
				fmt.Fprintf(writer, "F %s %s %f\n", feat, TagStr(tag), w)
			}
		}
	}
	return writer.Flush()
}

// UpdateFeat updates a feature weight.
func (m *Model) UpdateFeat(feat string, tag int, delta float64) {
	if m.Feats[feat] == nil {
		m.Feats[feat] = make(map[int]float64)
	}
	m.Feats[feat][tag] += delta
	if m.Feats[feat][tag] == 0 {
		delete(m.Feats[feat], tag)
		if len(m.Feats[feat]) == 0 {
			delete(m.Feats, feat)
		}
	}
}

// TagStr returns the string representation of a tag.
func TagStr(t int) string {
	switch t {
	case TagB:
		return "B"
	case TagM:
		return "M"
	case TagE:
		return "E"
	case TagS:
		return "S"
	}
	return "?"
}

func parseTag(s string) int {
	switch s {
	case "B":
		return TagB
	case "M":
		return TagM
	case "E":
		return TagE
	case "S":
		return TagS
	default:
		return -1
	}
}

// Sentence represents a training sentence with its golden tags.
type Sentence struct {
	Runes []rune
	Tags  []int
}

// LoadCorpus loads a segmented corpus file.
func LoadCorpus(path string) ([]Sentence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []Sentence
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		words := strings.Fields(line)
		var runes []rune
		var tags []int
		for _, word := range words {
			if util.IsPunctuation(word) {
				continue
			}
			wRunes := []rune(word)
			if len(wRunes) == 0 {
				continue
			}
			runes = append(runes, wRunes...)
			if len(wRunes) == 1 {
				tags = append(tags, TagS)
			} else {
				tags = append(tags, TagB)
				for k := 0; k < len(wRunes)-2; k++ {
					tags = append(tags, TagM)
				}
				tags = append(tags, TagE)
			}
		}
		if len(runes) > 0 {
			data = append(data, Sentence{runes, tags})
		}
	}
	return data, scanner.Err()
}

// LoadDictAsCorpus loads a dictionary and converts each word into a training sentence.
func LoadDictAsCorpus(path string) ([]Sentence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []Sentence
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		word := parts[0]
		if util.IsPunctuation(word) {
			continue
		}

		runes := []rune(word)
		if len(runes) == 0 {
			continue
		}

		var tags []int
		if len(runes) == 1 {
			tags = []int{TagS}
		} else {
			tags = make([]int, len(runes))
			tags[0] = TagB
			for k := 1; k < len(runes)-1; k++ {
				tags[k] = TagM
			}
			tags[len(runes)-1] = TagE
		}
		data = append(data, Sentence{runes, tags})
	}
	return data, scanner.Err()
}
