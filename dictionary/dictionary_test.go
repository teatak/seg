package dictionary

import (
	"os"
	"testing"
)

func TestDictionary_Load(t *testing.T) {
	content := "南京市 100\n长江大桥 100\n南京 10\n"
	tmpfile, err := os.CreateTemp("", "dict.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	dict := NewDictionary()
	err = dict.Load(tmpfile.Name())
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if dict.Total != 210 {
		t.Errorf("dict.Total = %v, want 210", dict.Total)
	}

	if dict.MaxLen != 4 { // "长江大桥" is 4 characters
		t.Errorf("dict.MaxLen = %v, want 4", dict.MaxLen)
	}

	if !dict.Contains("南京市") {
		t.Errorf("dict should contain '南京市'")
	}
}

func TestDictionary_LogProbability(t *testing.T) {
	dict := NewDictionary()
	dict.Words["A"] = 10
	dict.Words["B"] = 90
	dict.Total = 100

	probA := dict.LogProbability("A")
	// log(10/100) = log(0.1) ≈ -2.302585
	if probA > -2.3 || probA < -2.31 {
		t.Errorf("LogProbability('A') = %v, want ~ -2.3025", probA)
	}

	probUnknown := dict.LogProbability("Unknown")
	if probUnknown != -20.0 {
		t.Errorf("LogProbability('Unknown') = %v, want -20.0", probUnknown)
	}
}
