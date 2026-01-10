package crf

import (
	"math"
)

// Decode performs Viterbi decoding to find the best tag sequence.
func (m *Model) Decode(runes []rune) []int {
	n := len(runes)
	if n == 0 {
		return []int{}
	}

	// dp[i][tag] = max score ending at i with tag
	dp := make([][4]float64, n)
	// path[i][tag] = previous tag that gave max score
	path := make([][4]int, n)

	// Initialization (t=0)
	for tag := 0; tag < 4; tag++ {
		// Base score from features
		emission := m.computeEmission(runes, 0, tag)
		// Initial transition prob (simplified, assume transition from Start to tag)
		// Usually we have a specific start state, here we can assume B or S is likely start.
		// Or just 0 if we assume uniform start.
		// Let's rely on emission and let transition handle i>0.
		// Or assume previous is "End" (TagE or TagS) effectively.
		// For simplicity, 0.
		dp[0][tag] = emission
	}

	// Recurrence
	for i := 1; i < n; i++ {
		for curr := 0; curr < 4; curr++ {
			maxScore := -math.MaxFloat64
			bestPrev := -1

			emission := m.computeEmission(runes, i, curr)

			for prev := 0; prev < 4; prev++ {
				score := dp[i-1][prev] + m.Trans[prev][curr] + emission
				if score > maxScore {
					maxScore = score
					bestPrev = prev
				}
			}
			dp[i][curr] = maxScore
			path[i][curr] = bestPrev
		}
	}

	// Termination
	maxScore := -math.MaxFloat64
	bestEnd := -1
	for tag := 0; tag < 4; tag++ {
		// Could add transition to STOP state here if model supports it.
		if dp[n-1][tag] > maxScore {
			maxScore = dp[n-1][tag]
			bestEnd = tag
		}
	}

	// Backtrack
	tags := make([]int, n)
	tags[n-1] = bestEnd
	for i := n - 1; i > 0; i-- {
		tags[i-1] = path[i][tags[i]]
	}

	return tags
}

func (m *Model) computeEmission(runes []rune, idx int, tag int) float64 {
	score := 0.0
	// Feature templates
	features := ExtractFeatures(runes, idx)

	for _, feat := range features {
		if weights, ok := m.Feats[feat]; ok {
			if w, ok := weights[tag]; ok {
				score += w
			}
		}
	}
	return score
}
