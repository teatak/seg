package segmenter

type textBlock struct {
	runes          []rune
	isPureAlphaNum bool
}

func splitTextToBlocks(runes []rune) []textBlock {
	var blocks []textBlock
	if len(runes) == 0 {
		return blocks
	}

	var current []rune
	inWord := isWordChar(runes[0])

	for _, r := range runes {
		currentIsWord := isWordChar(r)
		if currentIsWord == inWord {
			current = append(current, r)
		} else {
			blocks = append(blocks, createBlock(current))
			current = []rune{r}
			inWord = currentIsWord
		}
	}
	if len(current) > 0 {
		blocks = append(blocks, createBlock(current))
	}
	return blocks
}

func createBlock(runes []rune) textBlock {
	pureAlpha := true
	for _, r := range runes {
		if !isAlphaNum(r) {
			pureAlpha = false
			break
		}
	}
	return textBlock{runes: runes, isPureAlphaNum: pureAlpha && isWordChar(runes[0])}
}

func isWordChar(r rune) bool {
	return isAlphaNum(r) || isCJK(r)
}

func isAlphaNum(r rune) bool {
	if r < 128 {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
	}
	return false
}

func isCJK(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}
