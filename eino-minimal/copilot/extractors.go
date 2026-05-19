package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	copilottools "eino-minimal/tools/copilot"

	"github.com/cloudwego/eino/schema"
)

// parseInput 将 CopilotInput 转为 map，同时并行预取所有数据
// 优化策略: Parallelization (Anthropic "Building Effective Agents")
// 因为 Orchestrator 的 3 个工具是确定性的（不依赖前一个结果），
// 在 graph 层面并行调用后直接拼装 context，Orchestrator 只需 1 次 LLM 调用
func parseInput(_ context.Context, input *CopilotInput, opts ...any) (map[string]any, error) {
	start := time.Now()

	// 从 goal 中粗略提取类目（简单关键词匹配）
	category := extractCategory(input.Goal)
	campaignType := extractCampaignType(input.Goal)

	// 并行预取 3 个数据源
	var (
		campaignData string
		metricsData  string
		segmentsData string
		wg           sync.WaitGroup
		mu           sync.Mutex
		errs         []error
	)

	ctx := context.Background()
	wg.Add(3)

	go func() {
		defer wg.Done()
		result, err := copilottools.GetCampaignHistory(ctx, &copilottools.CampaignHistoryParams{
			Category:     category,
			CampaignType: campaignType,
			Limit:        5,
		})
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("活动历史: %w", err))
		} else {
			campaignData = result
		}
	}()

	go func() {
		defer wg.Done()
		result, err := copilottools.GetPlatformMetrics(ctx, &copilottools.PlatformMetricsParams{
			Category:  category,
			DateRange: "last_30d",
		})
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("平台指标: %w", err))
		} else {
			metricsData = result
		}
	}()

	go func() {
		defer wg.Done()
		result, err := copilottools.GetUserSegments(ctx, &copilottools.UserSegmentsParams{
			Category: category,
		})
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("用户画像: %w", err))
		} else {
			segmentsData = result
		}
	}()

	wg.Wait()
	log.Printf("[Copilot] 数据预取完成, 耗时: %dms, 错误数: %d", time.Since(start).Milliseconds(), len(errs))

	// 即使部分失败也继续（用空字符串）
	return map[string]any{
		"goal":           input.Goal,
		"history":        input.History,
		"date":           time.Now().Format("2006-01-02"),
		"campaign_data":  campaignData,
		"metrics_data":   metricsData,
		"segments_data":  segmentsData,
		"category":       category,
	}, nil
}

// extractDiagnosis 从 Orchestrator 的输出中提取诊断结果，传递给 Planner
func extractDiagnosis(_ context.Context, msg *schema.Message, opts ...any) (map[string]any, error) {
	content := msg.Content

	// 尝试从 content 中提取 diagnosis_json 代码块
	diagnosisJSON := extractJSONBlock(content, "diagnosis_json")

	return map[string]any{
		"diagnosis":     diagnosisJSON,
		"full_analysis": content,
	}, nil
}

// extractJSONBlock 从文本中提取指定标记的代码块内容
func extractJSONBlock(text, marker string) string {
	markers := []string{"```" + marker, "```json"}

	for _, m := range markers {
		start := indexOf(text, m)
		if start == -1 {
			continue
		}
		contentStart := start + len(m)
		for contentStart < len(text) && text[contentStart] != '\n' {
			contentStart++
		}
		if contentStart < len(text) {
			contentStart++
		}

		end := indexOf(text[contentStart:], "```")
		if end == -1 {
			return text[contentStart:]
		}
		return text[contentStart : contentStart+end]
	}

	return text
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// extractCategory 从自然语言中提取类目关键词
func extractCategory(goal string) string {
	categories := []string{"女装", "数码", "食品", "男装", "美妆", "家居", "运动", "母婴"}
	for _, c := range categories {
		if containsChinese(goal, c) {
			return c
		}
	}
	return "女装" // 默认
}

// extractCampaignType 从自然语言中提取活动类型
func extractCampaignType(goal string) string {
	if containsChinese(goal, "拉新") || containsChinese(goal, "新客") {
		return "new_customer"
	}
	if containsChinese(goal, "清仓") || containsChinese(goal, "清库存") {
		return "clearance"
	}
	return "" // 不过滤，返回全部
}

func containsChinese(s, substr string) bool {
	return indexOf(s, substr) != -1
}

// toJSONPretty 将对象序列化为格式化的 JSON 字符串
func toJSONPretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
