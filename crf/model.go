package crf

import (
	"bufio"
	"os"
	"strconv"
	"strings"
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
