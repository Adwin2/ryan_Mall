package memory

import "unicode/utf8"

func EstimateTokens(text string, provider string) int {
	runeCount := utf8.RuneCountInString(text)
	switch provider {
	case "qwen":
		return int(float64(runeCount) * 1.4)
	case "ark":
		return int(float64(runeCount) * 1.5)
	default:
		return int(float64(runeCount) * 1.5)
	}
}
