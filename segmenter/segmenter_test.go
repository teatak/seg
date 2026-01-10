package segmenter

import (
	"reflect"
	"testing"

	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
)

func TestCut(t *testing.T) {
	dict := dictionary.NewDictionary()
	// Manually add words for testing (skipping file load for reliability in unit test)
	dict.Words["南京市"] = 100
	dict.Words["长江大桥"] = 100 // High freq to prefer this over 长江 + 大桥
	dict.Words["南京"] = 10
	dict.Words["市长"] = 10
	dict.Words["长江"] = 10
	dict.Words["大桥"] = 10
	dict.Words["江"] = 5
	dict.Words["大"] = 5
	dict.Words["桥"] = 5
	dict.Total = 1000
	dict.MaxLen = 4
	dict.Loaded = true

	seg := NewSegmenter(dict)

	tests := []struct {
		text     string
		expected []string
	}{
		{"南京市长江大桥", []string{"南京市", "长江大桥"}},
		{"我是程序员", []string{"我", "是", "程", "序", "员"}}, // OOV example
	}

	for _, tt := range tests {
		got := seg.Cut(tt.text, ModeDAG)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("Cut(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestCutForSearch(t *testing.T) {
	dict := dictionary.NewDictionary()
	dict.Words["南京市"] = 100
	dict.Words["长江大桥"] = 100
	dict.Words["南京"] = 10
	dict.Words["市"] = 5
	dict.Words["长江"] = 10
	dict.Words["大桥"] = 10
	dict.MaxLen = 4 // Important to set max len
	dict.Total = 1000
	dict.Loaded = true

	seg := NewSegmenter(dict)

	tests := []struct {
		text     string
		expected []string
	}{
		{"南京市长江大桥", []string{"南京", "市", "南京市", "长江", "大桥", "长江大桥"}},
	}

	for _, tt := range tests {
		got := seg.CutSearch(tt.text, ModeDAG)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("CutSearch(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}
func TestCutCRFForSearch(t *testing.T) {
	dict := dictionary.NewDictionary()
	dict.Words["南京"] = 10
	dict.Words["长江"] = 10
	dict.Words["大桥"] = 10
	dict.MaxLen = 2
	dict.Loaded = true

	seg := NewSegmenter(dict)

	// Mock CRF model that would segment "长江大桥" as one word
	// B M M E
	m := crf.NewModel()
	m.Trans[crf.TagB][crf.TagM] = 10.0
	m.Trans[crf.TagM][crf.TagM] = 10.0
	m.Trans[crf.TagM][crf.TagE] = 10.0
	// Character features to trigger B M M E
	m.Feats["U02:长"] = map[int]float64{crf.TagB: 10.0}
	m.Feats["U02:江"] = map[int]float64{crf.TagM: 10.0}
	m.Feats["U02:大"] = map[int]float64{crf.TagM: 10.0}
	m.Feats["U02:桥"] = map[int]float64{crf.TagE: 10.0}

	seg.CRFModel = m

	text := "长江大桥"
	got := seg.CutSearch(text, ModeCRF)
	// Sub-words of "长江大桥" (len 4) in dict: "长江", "大桥"
	expected := []string{"长江", "大桥", "长江大桥"}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("CutSearch(%q, ModeCRF) = %v, want %v", text, got, expected)
	}
}
