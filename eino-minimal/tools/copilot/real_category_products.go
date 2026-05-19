package copilottools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// CategoryProductsParams GetCategoryProducts 入参
type CategoryProductsParams struct {
	CategoryID uint `json:"category_id"`
	Limit      int  `json:"limit"`
}

// GetCategoryProducts 获取类目下的商品（调用真实 monolith API）
func GetCategoryProducts(_ context.Context, params *CategoryProductsParams) (string, error) {
	baseURL := os.Getenv("MONOLITH_API_BASE")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	reqURL := fmt.Sprintf("%s/api/v1/categories/%d/products?page_size=%d",
		baseURL, params.CategoryID, limit)

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("请求类目商品失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf(`{"error": "API返回错误", "status": %d, "body": %q}`, resp.StatusCode, string(body)), nil
	}

	// 解析并精简返回
	var apiResp struct {
		Code int `json:"code"`
		Data []struct {
			ID         uint    `json:"id"`
			Name       string  `json:"name"`
			Price      float64 `json:"price"`
			SalesCount int     `json:"sales_count"`
			Stock      int     `json:"stock"`
			CategoryID uint    `json:"category_id"`
			MainImage  string  `json:"main_image"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return string(body), nil
	}

	// 限制返回数量
	products := apiResp.Data
	if len(products) > limit {
		products = products[:limit]
	}

	result, _ := json.MarshalIndent(map[string]any{
		"category_id": params.CategoryID,
		"count":       len(products),
		"products":    products,
	}, "", "  ")
	return string(result), nil
}

// CategoryProductsTool 返回类目商品工具
func CategoryProductsTool() tool.InvokableTool {
	return utils.NewTool(&schema.ToolInfo{
		Name: "GetCategoryProducts",
		Desc: "获取指定类目下的商品列表。返回商品名称、价格、销量、库存等。用于了解类目商品池和选品。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"category_id": {
				Desc:     "类目ID",
				Type:     schema.Integer,
				Required: true,
			},
			"limit": {
				Desc: "返回条数上限，默认10",
				Type: schema.Integer,
			},
		}),
	}, GetCategoryProducts)
}

