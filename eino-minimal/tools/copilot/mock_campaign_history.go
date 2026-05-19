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

// CampaignHistoryParams GetCampaignHistory 入参
type CampaignHistoryParams struct {
	Category     string `json:"category"`
	CampaignType string `json:"campaign_type"`
	Limit        int    `json:"limit"`
}

// CampaignRecord 历史活动记录
type CampaignRecord struct {
	CampaignID   string  `json:"campaign_id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	CampaignType string  `json:"campaign_type"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Budget       float64 `json:"budget"`
	ActualSpend  float64 `json:"actual_spend"`
	GMV          float64 `json:"gmv"`
	NewCustomers int     `json:"new_customers"`
	ROI          float64 `json:"roi"`
	CVR          float64 `json:"cvr"`
	UV           int     `json:"uv"`
	PlayType     string  `json:"play_type"`
	Lessons      string  `json:"lessons"`
}

// GetCampaignHistory 查询历史活动（Mock）
func GetCampaignHistory(_ context.Context, params *CampaignHistoryParams) (string, error) {
	records, err := loadCampaignFixture()
	if err != nil {
		return "", fmt.Errorf("加载活动历史数据失败: %w", err)
	}

	filtered := make([]CampaignRecord, 0)
	for _, r := range records {
		if params.Category != "" && !strings.Contains(r.Category, params.Category) {
			continue
		}
		if params.CampaignType != "" && r.CampaignType != params.CampaignType {
			continue
		}
		filtered = append(filtered, r)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 5
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result, _ := json.MarshalIndent(filtered, "", "  ")
	return string(result), nil
}

// CampaignHistoryTool 返回活动历史工具
func CampaignHistoryTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetCampaignHistory",
		Desc: "查询历史活动数据，包含活动效果、ROI、经验教训等。用于了解过去类似活动的表现。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category": {
				Desc:     "目标类目名称，如：女装、数码、食品",
				Type:     schema.String,
				Required: true,
			},
			"campaign_type": {
				Desc: "活动类型过滤：promotion(促销)/new_customer(拉新)/clearance(清仓)，留空返回全部",
				Type: schema.String,
			},
			"limit": {
				Desc: "返回条数上限，默认5",
				Type: schema.Integer,
			},
		}),
	}, GetCampaignHistory)
}

func loadCampaignFixture() ([]CampaignRecord, error) {
	dir := os.Getenv("COPILOT_FIXTURES_DIR")
	if dir == "" {
		dir = "./data/fixtures"
	}
	data, err := os.ReadFile(dir + "/campaign_history.json")
	if err != nil {
		return nil, err
	}
	var records []CampaignRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}
