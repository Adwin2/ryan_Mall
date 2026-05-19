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

// UserSegmentsParams GetUserSegments 入参
type UserSegmentsParams struct {
	Category string `json:"category"`
}

// UserSegment 用户分群画像
type UserSegment struct {
	SegmentID   string   `json:"segment_id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Size        int      `json:"size"`
	Percentage  float64  `json:"percentage"`
	AvgSpend    float64  `json:"avg_spend"`
	Frequency   float64  `json:"frequency"`
	PrefChannel string   `json:"pref_channel"`
	PriceRange  string   `json:"price_range"`
	Features    []string `json:"features"`
}

// GetUserSegments 获取用户画像（Mock）
func GetUserSegments(_ context.Context, params *UserSegmentsParams) (string, error) {
	segments, err := loadSegmentsFixture()
	if err != nil {
		return "", fmt.Errorf("加载用户画像数据失败: %w", err)
	}

	filtered := make([]UserSegment, 0)
	for _, s := range segments {
		if strings.Contains(s.Category, params.Category) {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return fmt.Sprintf(`{"error": "未找到类目 %s 的用户画像数据"}`, params.Category), nil
	}

	result, _ := json.MarshalIndent(filtered, "", "  ")
	return string(result), nil
}

// UserSegmentsTool 返回用户画像工具
func UserSegmentsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetUserSegments",
		Desc: "获取指定类目的用户分群画像，包含人群规模、消费特征、偏好渠道、价格敏感度等。用于制定精准营销策略。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category": {
				Desc:     "目标类目名称，如：女装、数码、食品",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, GetUserSegments)
}

func loadSegmentsFixture() ([]UserSegment, error) {
	dir := os.Getenv("COPILOT_FIXTURES_DIR")
	if dir == "" {
		dir = "./data/fixtures"
	}
	data, err := os.ReadFile(dir + "/user_segments.json")
	if err != nil {
		return nil, err
	}
	var segments []UserSegment
	if err := json.Unmarshal(data, &segments); err != nil {
		return nil, err
	}
	return segments, nil
}
