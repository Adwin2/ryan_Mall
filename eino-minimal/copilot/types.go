package copilot

import "github.com/cloudwego/eino/schema"

// CopilotInput 是 compose.Graph 的入口类型
type CopilotInput struct {
	Goal      string            `json:"goal"`
	SessionID string            `json:"session_id"`
	History   []*schema.Message `json:"history"`
}

// CampaignGoal 从用户自然语言中解析出的结构化目标
type CampaignGoal struct {
	Category     string  `json:"category"`
	CampaignType string  `json:"campaign_type"`
	Budget       float64 `json:"budget"`
	KPI          KPITarget `json:"kpi"`
	TimeRange    string  `json:"time_range"`
	Constraints  []string `json:"constraints"`
}

// KPITarget 核心 KPI 目标
type KPITarget struct {
	Metric      string  `json:"metric"`
	TargetValue float64 `json:"target_value"`
	Unit        string  `json:"unit"`
}

// DiagnosisReport Orchestrator 生成的诊断报告
type DiagnosisReport struct {
	CurrentState   CurrentState   `json:"current_state"`
	KeyBottlenecks []string       `json:"key_bottlenecks"`
	HistoricalRef  *HistoricalRef `json:"historical_reference"`
	ParsedGoal     CampaignGoal   `json:"parsed_goal"`
	Opportunities  []string       `json:"opportunities"`
}

// CurrentState 当前盘面数据摘要
type CurrentState struct {
	Category         string   `json:"category"`
	NewCustomerRatio float64  `json:"new_customer_ratio"`
	AvgUVDaily       int      `json:"avg_uv_daily"`
	AvgCVR           float64  `json:"avg_cvr"`
	AvgAOV           float64  `json:"avg_aov"`
	MainChannels     []string `json:"main_channels"`
	LastCampaignROI  float64  `json:"last_campaign_roi"`
}

// HistoricalRef 历史活动参考
type HistoricalRef struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	Result       string `json:"result"`
	Lesson       string `json:"lesson"`
}

// CampaignPlan 单套活动方案（结构化参考，V1 Planner 直接输出 Markdown）
type CampaignPlan struct {
	Name               string             `json:"name"`
	Strategy           string             `json:"strategy"`
	PlayType           string             `json:"play_type"`
	ObjectiveBreakdown ObjectiveBreakdown `json:"objective_breakdown"`
	ProductPool        []ProductSummary   `json:"product_pool"`
	DiscountStrategy   string             `json:"discount_strategy"`
	ChannelPlan        []ChannelPlan      `json:"channel_plan"`
	BudgetBreakdown    []BudgetItem       `json:"budget_breakdown"`
	ExpectedOutcome    OutcomeRange       `json:"expected_outcome"`
	Rationale          string             `json:"rationale"`
}

// ObjectiveBreakdown 目标拆解
type ObjectiveBreakdown struct {
	TargetGMV   float64 `json:"target_gmv"`
	TargetUV    int     `json:"target_uv"`
	TargetCVR   float64 `json:"target_cvr"`
	TargetAOV   float64 `json:"target_aov"`
	Explanation string  `json:"explanation"`
}

// ProductSummary 商品摘要
type ProductSummary struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	SalesCount int     `json:"sales_count"`
	Stock      int     `json:"stock"`
	CategoryID uint    `json:"category_id"`
	MainImage  string  `json:"main_image,omitempty"`
}

// ChannelPlan 渠道分配
type ChannelPlan struct {
	Channel    string  `json:"channel"`
	Percentage float64 `json:"percentage"`
	Strategy   string  `json:"strategy"`
}

// BudgetItem 预算拆解项
type BudgetItem struct {
	Item   string  `json:"item"`
	Amount float64 `json:"amount"`
}

// OutcomeRange 预期效果区间
type OutcomeRange struct {
	MetricName string  `json:"metric_name"`
	LowBound   float64 `json:"low_bound"`
	HighBound  float64 `json:"high_bound"`
	ROIRange   string  `json:"roi_range"`
}

// SSEEvent SSE 推送事件（增强版）
type SSEEvent struct {
	Type     string      `json:"type"`
	Content  string      `json:"content,omitempty"`
	Node     string      `json:"node,omitempty"`
	Status   string      `json:"status,omitempty"`
	Tool     string      `json:"tool,omitempty"`
	Params   any         `json:"params,omitempty"`
	Result   string      `json:"result,omitempty"`
	Metadata *EventMeta  `json:"metadata,omitempty"`
}

// EventMeta 完成事件的元数据
type EventMeta struct {
	DurationMs  int64    `json:"duration_ms"`
	ToolsCalled []string `json:"tools_called"`
	Model       string   `json:"model"`
}

// ProgressEvent 进度事件（用于 channel 通信）
type ProgressEvent struct {
	Type   string `json:"type"`   // "progress", "tool_call", "diagnosis"
	Node   string `json:"node,omitempty"`
	Status string `json:"status,omitempty"` // "start", "done"
	Tool   string `json:"tool,omitempty"`
	Params any    `json:"params,omitempty"`
	Result string `json:"result,omitempty"`
	Data   string `json:"data,omitempty"` // diagnosis JSON
}

// DiagnoseRequest 分步诊断请求
type DiagnoseRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// DiagnoseResponse 分步诊断响应
type DiagnoseResponse struct {
	Code      int              `json:"code"`
	Message   string           `json:"message"`
	SessionID string           `json:"session_id"`
	Diagnosis string           `json:"diagnosis"`      // raw JSON string
	Analysis  string           `json:"full_analysis"`  // full markdown analysis
}

// PlanFromRequest 从诊断继续生成方案
type PlanFromRequest struct {
	SessionID string `json:"session_id"`
	Diagnosis string `json:"diagnosis"`      // edited diagnosis JSON
	Analysis  string `json:"full_analysis"`  // original full analysis
}

// CopilotRequest 同步接口请求体
type CopilotRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// CopilotResponse 同步接口响应体
type CopilotResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *CopilotResult `json:"data"`
}

// CopilotResult 完整结果
type CopilotResult struct {
	SessionID string          `json:"session_id"`
	Diagnosis *DiagnosisReport `json:"diagnosis"`
	Plan      string          `json:"plan"`
	Metadata  ResultMetadata  `json:"metadata"`
}

// ResultMetadata 结果元数据
type ResultMetadata struct {
	DurationMs  int64    `json:"duration_ms"`
	ToolsCalled []string `json:"tools_called"`
	Model       string   `json:"model"`
	GraphNodes  int      `json:"graph_nodes"`
}
