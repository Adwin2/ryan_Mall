package copilottools

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

// ProductSearchParams SearchProducts 入参
type ProductSearchParams struct {
	Keyword    string  `json:"keyword"`
	CategoryID uint    `json:"category_id"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	SortBy     string  `json:"sort_by"`
	Limit      int     `json:"limit"`
}

// SearchProducts 搜索商品（调用真实 monolith API）
func SearchProducts(_ context.Context, params *ProductSearchParams) (string, error) {
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

	reqURL := baseURL + "/api/v1/products?" + q.Encode()
	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("请求商品列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf(`{"error": "API返回错误", "status": %d, "body": %q}`, resp.StatusCode, string(body)), nil
	}

	// 解析并精简返回（只保留有用字段）
	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			Products []struct {
				ID         uint    `json:"id"`
				Name       string  `json:"name"`
				Price      float64 `json:"price"`
				SalesCount int     `json:"sales_count"`
				Stock      int     `json:"stock"`
				CategoryID uint    `json:"category_id"`
				MainImage  string  `json:"main_image"`
			} `json:"products"`
			Total int64 `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 如果解析失败，直接返回原始内容
		return string(body), nil
	}

	result, _ := json.MarshalIndent(map[string]any{
		"total":    apiResp.Data.Total,
		"products": apiResp.Data.Products,
	}, "", "  ")
	return string(result), nil
}

// SearchProductsTool 返回商品搜索工具
func SearchProductsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "SearchProducts",
		Desc: "搜索商品列表，支持关键词、类目、价格区间、排序等条件。返回商品名称、价格、销量、库存等信息。用于为活动方案选品。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keyword": {
				Desc: "搜索关键词，如：连衣裙、T恤、手机",
				Type: schema.String,
			},
			"category_id": {
				Desc: "类目ID，用于按类目筛选",
				Type: schema.Integer,
			},
			"min_price": {
				Desc: "最低价格",
				Type: schema.Number,
			},
			"max_price": {
				Desc: "最高价格",
				Type: schema.Number,
			},
			"sort_by": {
				Desc: "排序字段：sales_count(按销量) / price(按价格) / created_at(按时间)",
				Type: schema.String,
			},
			"limit": {
				Desc: "返回条数上限，默认10",
				Type: schema.Integer,
			},
		}),
	}, SearchProducts)
}
