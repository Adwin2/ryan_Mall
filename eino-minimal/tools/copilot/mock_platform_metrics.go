package copilottools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// PlatformMetricsParams GetPlatformMetrics 入参
type PlatformMetricsParams struct {
	Category  string `json:"category"`
	DateRange string `json:"date_range"`
}

// PlatformMetrics 平台类目级指标
type PlatformMetrics struct {
	Category         string           `json:"category"`
	Period           string           `json:"period"`
	DailyAvgUV       int              `json:"daily_avg_uv"`
	DailyAvgPV       int              `json:"daily_avg_pv"`
	OverallCVR       float64          `json:"overall_cvr"`
	AvgAOV           float64          `json:"avg_aov"`
	NewCustomerRatio float64          `json:"new_customer_ratio"`
	RepurchaseRate   float64          `json:"repurchase_rate"`
	TopChannels      []ChannelMetric  `json:"top_channels"`
	TrendVsPrev      TrendComparison  `json:"trend_vs_prev"`
}

// ChannelMetric 渠道级指标
type ChannelMetric struct {
	Channel string  `json:"channel"`
	UV      int     `json:"uv"`
	CVR     float64 `json:"cvr"`
	AOV     float64 `json:"aov"`
	Share   float64 `json:"share"`
}

// TrendComparison 环比趋势
type TrendComparison struct {
	UVChange  float64 `json:"uv_change"`
	CVRChange float64 `json:"cvr_change"`
	AOVChange float64 `json:"aov_change"`
}

// GetPlatformMetrics 获取平台指标（Mock）
func GetPlatformMetrics(_ context.Context, params *PlatformMetricsParams) (string, error) {
	metrics, err := loadMetricsFixture()
	if err != nil {
		return "", fmt.Errorf("加载平台指标数据失败: %w", err)
	}

	for _, m := range metrics {
		if strings.Contains(m.Category, params.Category) {
			result, _ := json.MarshalIndent(m, "", "  ")
			return string(result), nil
		}
	}

	return fmt.Sprintf(`{"error": "未找到类目 %s 的指标数据"}`, params.Category), nil
}

// PlatformMetricsTool 返回平台指标工具
func PlatformMetricsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetPlatformMetrics",
		Desc: "获取平台类目级运营指标，包含UV、CVR、AOV、新客占比、渠道分布、环比趋势等。用于诊断当前盘面。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category": {
				Desc:     "目标类目名称，如：女装、数码、食品",
				Type:     schema.String,
				Required: true,
			},
			"date_range": {
				Desc: "统计周期，如 last_30d / last_7d，默认 last_30d",
				Type: schema.String,
			},
		}),
	}, GetPlatformMetrics)
}

func loadMetricsFixture() ([]PlatformMetrics, error) {
	dir := os.Getenv("COPILOT_FIXTURES_DIR")
	if dir == "" {
		dir = "./data/fixtures"
	}
	data, err := os.ReadFile(dir + "/platform_metrics.json")
	if err != nil {
		return nil, err
	}
	var metrics []PlatformMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
