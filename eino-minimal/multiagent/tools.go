package multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// --- Shared: ListCategories ---

type ListCategoriesParams struct {
	ParentID *uint `json:"parent_id"`
}

func listCategories(_ context.Context, params *ListCategoriesParams) (string, error) {
	baseURL := os.Getenv("MONOLITH_API_BASE")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	resp, err := http.Get(baseURL + "/api/v1/categories")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf(`{"error":"API error","status":%d}`, resp.StatusCode), nil
	}

	if params.ParentID == nil {
		return string(body), nil
	}

	var result struct {
		Data []struct {
			ID       uint   `json:"id"`
			Name     string `json:"name"`
			ParentID uint   `json:"parent_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return string(body), nil
	}

	filtered := make([]map[string]any, 0)
	for _, c := range result.Data {
		if c.ParentID == *params.ParentID {
			filtered = append(filtered, map[string]any{"id": c.ID, "name": c.Name, "parent_id": c.ParentID})
		}
	}
	data, _ := json.Marshal(map[string]any{"data": filtered})
	return string(data), nil
}

func ListCategoriesTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "ListCategories",
		Desc: "获取商品类目列表。返回所有类目及其层级关系（parent_id=0 为顶级类目）。用户问'有什么品类'、'女装有哪些'时应调用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"parent_id": {Desc: "父类目ID，传0获取顶级类目，传具体ID获取其子类目，不传则返回全部", Type: schema.Integer},
		}),
	}, listCategories)
}

// --- Shared: SearchProducts (复用现有逻辑) ---

type ProductSearchParams struct {
	Keyword    string  `json:"keyword"`
	CategoryID uint    `json:"category_id"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	SortBy     string  `json:"sort_by"`
	Limit      int     `json:"limit"`
}

func searchProducts(_ context.Context, params *ProductSearchParams) (string, error) {
	baseURL := os.Getenv("MONOLITH_API_BASE")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	q := url.Values{}
	if params.Keyword != "" {
		q.Set("keyword", params.Keyword)
	}
	if params.CategoryID > 0 {
		q.Set("category_id", strconv.FormatUint(uint64(params.CategoryID), 10))
	}
	if params.MinPrice > 0 {
		q.Set("min_price", fmt.Sprintf("%.2f", params.MinPrice))
	}
	if params.MaxPrice > 0 {
		q.Set("max_price", fmt.Sprintf("%.2f", params.MaxPrice))
	}
	if params.SortBy != "" {
		q.Set("sort_by", params.SortBy)
	}
	q.Set("page_size", strconv.Itoa(limit))

	resp, err := http.Get(baseURL + "/api/v1/products?" + q.Encode())
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf(`{"error":"API error","status":%d}`, resp.StatusCode), nil
	}
	return string(body), nil
}

func SearchProductsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "SearchProducts",
		Desc: "搜索商品列表。可通过 category_id 浏览某个类目下的全部商品（推荐），也可用关键词模糊搜索。如果用户提到类目名称（如'女装'），应先调用 ListCategories 获取对应 category_id，再用 category_id 搜索。不传任何参数则返回全部商品。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keyword":     {Desc: "搜索关键词（匹配商品名称和描述），可选", Type: schema.String},
			"category_id": {Desc: "类目ID，按类目筛选商品（优先使用此参数而非关键词）", Type: schema.Integer},
			"min_price":   {Desc: "最低价格筛选", Type: schema.Number},
			"max_price":   {Desc: "最高价格筛选", Type: schema.Number},
			"sort_by":     {Desc: "排序方式: sales_count(销量)/price(价格)/created_at(上新)", Type: schema.String},
			"limit":       {Desc: "返回条数上限，默认10", Type: schema.Integer},
		}),
	}, searchProducts)
}

// --- Consumer: GetProductDetail ---

type ProductDetailParams struct {
	ProductID uint `json:"product_id"`
}

func getProductDetail(_ context.Context, params *ProductDetailParams) (string, error) {
	baseURL := os.Getenv("MONOLITH_API_BASE")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/products/%d", baseURL, params.ProductID))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func GetProductDetailTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetProductDetail",
		Desc: "获取商品详细信息，包括规格、库存、描述等。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"product_id": {Desc: "商品ID", Type: schema.Integer, Required: true},
		}),
	}, getProductDetail)
}

// --- Consumer: AddToCart ---

type AddToCartParams struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

func addToCart(_ context.Context, params *AddToCartParams) (string, error) {
	baseURL := os.Getenv("MONOLITH_API_BASE")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	qty := params.Quantity
	if qty <= 0 {
		qty = 1
	}

	result := map[string]any{
		"status":     "success",
		"message":    fmt.Sprintf("商品 #%d 已加入购物车，数量: %d", params.ProductID, qty),
		"product_id": params.ProductID,
		"quantity":   qty,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func AddToCartTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "AddToCart",
		Desc: "将商品加入用户购物车。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"product_id": {Desc: "商品ID", Type: schema.Integer, Required: true},
			"quantity":   {Desc: "数量，默认1", Type: schema.Integer},
		}),
	}, addToCart)
}

// --- Ops: GetPlatformMetrics ---

type MetricsParams struct {
	Category string `json:"category"`
}

func getPlatformMetrics(_ context.Context, params *MetricsParams) (string, error) {
	data, err := os.ReadFile("./data/fixtures/platform_metrics.json")
	if err != nil {
		return `{"error":"metrics data not found"}`, nil
	}
	return string(data), nil
}

func GetPlatformMetricsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetPlatformMetrics",
		Desc: "获取平台运营指标数据（UV、转化率、客单价、渠道表现等）。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category": {Desc: "类目名称，可选", Type: schema.String},
		}),
	}, getPlatformMetrics)
}

// --- Ops: GetCampaignHistory ---

type CampaignHistoryParams struct {
	Category     string `json:"category"`
	CampaignType string `json:"campaign_type"`
}

func getCampaignHistory(_ context.Context, params *CampaignHistoryParams) (string, error) {
	data, err := os.ReadFile("./data/fixtures/campaign_history.json")
	if err != nil {
		return `{"error":"campaign history not found"}`, nil
	}
	return string(data), nil
}

func GetCampaignHistoryTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetCampaignHistory",
		Desc: "查询历史营销活动记录，了解过往活动效果和经验教训。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category":      {Desc: "类目名称", Type: schema.String},
			"campaign_type": {Desc: "活动类型", Type: schema.String},
		}),
	}, getCampaignHistory)
}

// --- Ops: UpdateCarousel ---

type UpdateCarouselParams struct {
	ProductIDs []uint `json:"product_ids"`
	Reason     string `json:"reason"`
}

func updateCarousel(_ context.Context, params *UpdateCarouselParams) (string, error) {
	result := map[string]any{
		"status":      "success",
		"message":     "首页轮播已更新",
		"product_ids": params.ProductIDs,
		"reason":      params.Reason,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	os.MkdirAll("./data/fixtures", 0755)
	os.WriteFile("./data/fixtures/carousel.json", data, 0644)

	return string(data), nil
}

func UpdateCarouselTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "UpdateCarousel",
		Desc: "更新首页轮播商品。需要提供商品ID列表和更换理由。这是一个执行动作，请在调用前向用户说明理由。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"product_ids": {Desc: "商品ID列表", Type: schema.Array, Required: true},
			"reason":      {Desc: "更换理由", Type: schema.String, Required: true},
		}),
	}, updateCarousel)
}

// --- Tool Sets ---

func OpsTools() []tool.BaseTool {
	return []tool.BaseTool{
		ListCategoriesTool(),
		SearchProductsTool(),
		GetPlatformMetricsTool(),
		GetCampaignHistoryTool(),
		UpdateCarouselTool(),
	}
}

func ConsumerTools() []tool.BaseTool {
	return []tool.BaseTool{
		ListCategoriesTool(),
		SearchProductsTool(),
		GetProductDetailTool(),
		AddToCartTool(),
	}
}
