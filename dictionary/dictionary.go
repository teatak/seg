package dictionary

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
)

// Dictionary holds words and their frequencies/probabilities.
type Dictionary struct {
	Total  float64
	Words  map[string]float64
	MaxLen int
	Loaded bool
}

// NewDictionary creates a new empty dictionary.
func NewDictionary() *Dictionary {
	return &Dictionary{
		Words: make(map[string]float64),
	}
}

// Load loads words from a file.
// File format: word frequency (space separated)
func (d *Dictionary) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		word := parts[0]
		freq := 1.0 // default
		if len(parts) >= 2 {
			f, err := strconv.ParseFloat(parts[1], 64)
			if err == nil {
				freq = f
			}
		} else {
			// If it's a top-level word without frequency, give it a high default
			freq = 20000.0
		}
		d.Words[word] = freq
		d.Total += freq
		if len([]rune(word)) > d.MaxLen {
			d.MaxLen = len([]rune(word))
		}
	}
	d.Loaded = true
	return scanner.Err()
}

// Frequency returns the frequency of a word.
func (d *Dictionary) Frequency(word string) (float64, bool) {
	val, ok := d.Words[word]
	return val, ok
}

// Contains checks if a word exists in the dictionary.
func (d *Dictionary) Contains(word string) bool {
	_, ok := d.Words[word]
	return ok
}

// LogProbability returns the log probability of a word.
// basic smoothing: if total is 0, return extremely small number.
func (d *Dictionary) LogProbability(word string) float64 {
	if d.Total <= 0 {
		return -20.0
	}
	freq, ok := d.Words[word]
	if !ok {
		// Return a very small probability for unknown words (smoothing)
		// Usually handled by HMM or just a penalty in DAG
		return -20.0
	}
	return math.Log(freq / d.Total)
}
