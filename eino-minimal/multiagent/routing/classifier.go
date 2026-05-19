package routing

import (
	"strings"
	"unicode/utf8"
)

type ComplexityTier string

const (
	TierFast   ComplexityTier = "fast"
	TierStrong ComplexityTier = "strong"
)

var strongKeywords = []string{
	"诊断", "分析", "策略", "方案", "规划", "对比", "优化",
	"根据数据", "根据报表", "趋势", "预测", "归因",
	"设计活动", "制定计划", "全链路", "人群画像",
}

var strongPatterns = []string{
	"帮我", "请你", "列出", "分步", "方案一", "方案二",
	"给出建议", "怎么提升", "如何改善",
}

func Classify(message string) ComplexityTier {
	runeCount := utf8.RuneCountInString(message)
	if runeCount > 50 {
		return TierStrong
	}

	lower := strings.ToLower(message)
	for _, kw := range strongKeywords {
		if strings.Contains(lower, kw) {
			return TierStrong
		}
	}
	for _, p := range strongPatterns {
		if strings.Contains(lower, p) {
			return TierStrong
		}
	}

	return TierFast
}
