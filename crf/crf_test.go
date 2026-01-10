package crf

import (
	"reflect"
	"testing"
)

func TestModel_Load(t *testing.T) {
	m := NewModel()
	// Mock model content
	// T from to weight
	// F feat tag weight
	m.Trans[TagB][TagE] = 10.5
	m.Feats["U00:我"] = map[int]float64{TagS: 5.0}

	if m.Trans[TagB][TagE] != 10.5 {
		t.Errorf("expected Trans[TagB][TagE] = 10.5, got %v", m.Trans[TagB][TagE])
	}
}

func TestDecode(t *testing.T) {
	m := NewModel()
	// Setup a simple model that prefers "B E" for 2-char string
	// T B E 10.0
	// F U00:A B 1.0
	// F U00:B E 1.0
	m.Trans[TagB][TagE] = 10.0
	m.Feats["U00:A"] = map[int]float64{TagB: 1.0}
	m.Feats["U00:B"] = map[int]float64{TagE: 1.0}

	runes := []rune("AB")
	got := m.Decode(runes)
	expected := []int{TagB, TagE}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Decode('AB') = %v, want %v", got, expected)
	}
}
