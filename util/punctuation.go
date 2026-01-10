package util

import (
	"unicode"
)

// IsPunctuation checks if a string consists entirely of punctuation or special CJK symbols.
func IsPunctuation(s string) bool {
	for _, r := range s {
		if !isPunct(r) {
			return false
		}
	}
	return true
}

func isPunct(r rune) bool {
	if unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return true
	}
	// CJK Symbols and Punctuation
	if r >= 0x3000 && r <= 0x303F {
		return true
	}
	// Full-width forms
	if r >= 0xFF00 && r <= 0xFFEF {
		return true
	}
	return false
}

// ContainsPunctuation checks if any part of the string contains punctuation or special symbols.
func ContainsPunctuation(s string) bool {
	for _, r := range s {
		if isPunct(r) {
			return true
		}
	}
	return false
}
