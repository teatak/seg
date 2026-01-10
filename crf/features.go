package crf

// ExtractFeatures generates feature strings for a character at a given index in a sequence.
func ExtractFeatures(runes []rune, idx int) []string {
	// Helper to safely get char
	getChar := func(offset int) string {
		pos := idx + offset
		if pos < 0 || pos >= len(runes) {
			return "_BOS_" // beginning/end of sentence marker placeholder
		}
		return string(runes[pos])
	}

	// Feature templates
	// U00: x[i-2]
	// U01: x[i-1]
	// U02: x[i]
	// U03: x[i+1]
	// U04: x[i+2]
	return []string{
		"U00:" + getChar(-2),
		"U01:" + getChar(-1),
		"U02:" + getChar(0),
		"U03:" + getChar(1),
		"U04:" + getChar(2),
	}
}
