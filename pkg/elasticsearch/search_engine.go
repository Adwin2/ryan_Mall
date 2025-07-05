package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// SearchEngine Elasticsearch搜索引擎
type SearchEngine struct {
	client *elasticsearch.Client
	ctx    context.Context
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(addresses []string) (*SearchEngine, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	
	return &SearchEngine{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// ProductDocument 商品文档结构
type ProductDocument struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryID  uint     `json:"category_id"`
	Category    string   `json:"category"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Status      int      `json:"status"`
	Tags        []string `json:"tags"`
	Brand       string   `json:"brand"`
	Images      []string `json:"images"`
	SalesCount  int      `json:"sales_count"`
	ViewCount   int      `json:"view_count"`
	Rating      float64  `json:"rating"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query      string            `json:"query"`
	CategoryID uint              `json:"category_id,omitempty"`
	MinPrice   float64           `json:"min_price,omitempty"`
	MaxPrice   float64           `json:"max_price,omitempty"`
	Brand      string            `json:"brand,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	SortBy     string            `json:"sort_by,omitempty"` // price, sales, rating, created_at
	SortOrder  string            `json:"sort_order,omitempty"` // asc, desc
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Products    []ProductDocument `json:"products"`
	Total       int64             `json:"total"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Suggestions []string          `json:"suggestions,omitempty"`
	TimeTaken   string            `json:"time_taken"`
}

// IndexProduct 索引商品
func (se *SearchEngine) IndexProduct(product *ProductDocument) error {
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	
	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: fmt.Sprintf("%d", product.ID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}
	
	res, err := req.Do(se.ctx, se.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("error indexing product: %s", res.String())
	}
	
	return nil
}

// SearchProducts 智能商品搜索
func (se *SearchEngine) SearchProducts(req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索查询
	query := se.buildSearchQuery(req)
	
	// 执行搜索
	searchReq := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  strings.NewReader(query),
	}
	
	res, err := searchReq.Do(se.ctx, se.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}
	
	// 解析响应
	return se.parseSearchResponse(res, req)
}

// buildSearchQuery 构建搜索查询
func (se *SearchEngine) buildSearchQuery(req *SearchRequest) string {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"status": 1, // 只搜索上架商品
						},
					},
				},
			},
		},
		"sort": se.buildSortClause(req),
		"from": (req.Page - 1) * req.PageSize,
		"size": req.PageSize,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"name":        map[string]interface{}{},
				"description": map[string]interface{}{},
			},
		},
		"aggs": se.buildAggregations(),
	}
	
	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
	
	// 添加文本搜索
	if req.Query != "" {
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"name^3", "description^2", "category", "brand", "tags"},
				"type":   "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}
	
	// 添加过滤条件
	filters := boolQuery["filter"].([]interface{})
	
	if req.CategoryID > 0 {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": req.CategoryID,
			},
		})
	}
	
	if req.MinPrice > 0 || req.MaxPrice > 0 {
		priceRange := map[string]interface{}{}
		if req.MinPrice > 0 {
			priceRange["gte"] = req.MinPrice
		}
		if req.MaxPrice > 0 {
			priceRange["lte"] = req.MaxPrice
		}
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"price": priceRange,
			},
		})
	}
	
	if req.Brand != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"brand.keyword": req.Brand,
			},
		})
	}
	
	if len(req.Tags) > 0 {
		filters = append(filters, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags.keyword": req.Tags,
			},
		})
	}
	
	boolQuery["filter"] = filters
	
	queryBytes, _ := json.Marshal(query)
	return string(queryBytes)
}

// buildSortClause 构建排序子句
func (se *SearchEngine) buildSortClause(req *SearchRequest) []interface{} {
	if req.SortBy == "" {
		// 默认按相关性和销量排序
		return []interface{}{
			"_score",
			map[string]interface{}{
				"sales_count": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	}
	
	order := "desc"
	if req.SortOrder == "asc" {
		order = "asc"
	}
	
	switch req.SortBy {
	case "price":
		return []interface{}{
			map[string]interface{}{
				"price": map[string]interface{}{
					"order": order,
				},
			},
		}
	case "sales":
		return []interface{}{
			map[string]interface{}{
				"sales_count": map[string]interface{}{
					"order": order,
				},
			},
		}
	case "rating":
		return []interface{}{
			map[string]interface{}{
				"rating": map[string]interface{}{
					"order": order,
				},
			},
		}
	case "created_at":
		return []interface{}{
			map[string]interface{}{
				"created_at": map[string]interface{}{
					"order": order,
				},
			},
		}
	default:
		return []interface{}{"_score"}
	}
}

// buildAggregations 构建聚合查询
func (se *SearchEngine) buildAggregations() map[string]interface{} {
	return map[string]interface{}{
		"categories": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "category.keyword",
				"size":  10,
			},
		},
		"brands": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "brand.keyword",
				"size":  10,
			},
		},
		"price_ranges": map[string]interface{}{
			"range": map[string]interface{}{
				"field": "price",
				"ranges": []interface{}{
					map[string]interface{}{"to": 100},
					map[string]interface{}{"from": 100, "to": 500},
					map[string]interface{}{"from": 500, "to": 1000},
					map[string]interface{}{"from": 1000, "to": 5000},
					map[string]interface{}{"from": 5000},
				},
			},
		},
		"avg_price": map[string]interface{}{
			"avg": map[string]interface{}{
				"field": "price",
			},
		},
	}
}

// parseSearchResponse 解析搜索响应
func (se *SearchEngine) parseSearchResponse(res *esapi.Response, req *SearchRequest) (*SearchResponse, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	hits := result["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))
	timeTaken := result["took"].(float64)
	
	var products []ProductDocument
	for _, hit := range hits["hits"].([]interface{}) {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		
		var product ProductDocument
		sourceBytes, _ := json.Marshal(source)
		json.Unmarshal(sourceBytes, &product)
		
		// 添加高亮信息
		if highlight, ok := hitMap["highlight"]; ok {
			highlightMap := highlight.(map[string]interface{})
			if nameHighlight, ok := highlightMap["name"]; ok {
				product.Name = nameHighlight.([]interface{})[0].(string)
			}
		}
		
		products = append(products, product)
	}
	
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	response := &SearchResponse{
		Products:   products,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		TimeTaken:  fmt.Sprintf("%dms", int(timeTaken)),
	}
	
	// 添加聚合结果
	if aggs, ok := result["aggregations"]; ok {
		response.Aggregations = aggs.(map[string]interface{})
	}
	
	return response, nil
}

// GetSuggestions 获取搜索建议
func (se *SearchEngine) GetSuggestions(query string, size int) ([]string, error) {
	searchQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"product_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "suggest",
					"size":  size,
				},
			},
		},
	}
	
	queryBytes, _ := json.Marshal(searchQuery)
	
	req := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  strings.NewReader(string(queryBytes)),
	}
	
	res, err := req.Do(se.ctx, se.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	var suggestions []string
	if suggest, ok := result["suggest"]; ok {
		suggestMap := suggest.(map[string]interface{})
		if productSuggest, ok := suggestMap["product_suggest"]; ok {
			for _, suggestion := range productSuggest.([]interface{}) {
				suggestionMap := suggestion.(map[string]interface{})
				if options, ok := suggestionMap["options"]; ok {
					for _, option := range options.([]interface{}) {
						optionMap := option.(map[string]interface{})
						suggestions = append(suggestions, optionMap["text"].(string))
					}
				}
			}
		}
	}
	
	return suggestions, nil
}

// RecommendationEngine 推荐引擎
type RecommendationEngine struct {
	searchEngine *SearchEngine
}

// NewRecommendationEngine 创建推荐引擎
func NewRecommendationEngine(searchEngine *SearchEngine) *RecommendationEngine {
	return &RecommendationEngine{
		searchEngine: searchEngine,
	}
}

// GetSimilarProducts 获取相似商品
func (re *RecommendationEngine) GetSimilarProducts(productID uint, size int) ([]ProductDocument, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"more_like_this": map[string]interface{}{
				"fields": []string{"name", "description", "category", "tags"},
				"like": []interface{}{
					map[string]interface{}{
						"_index": "products",
						"_id":    fmt.Sprintf("%d", productID),
					},
				},
				"min_term_freq":      1,
				"max_query_terms":    12,
				"min_doc_freq":       1,
				"minimum_should_match": "30%",
			},
		},
		"size": size,
	}
	
	queryBytes, _ := json.Marshal(query)
	
	req := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  strings.NewReader(string(queryBytes)),
	}
	
	res, err := req.Do(re.searchEngine.ctx, re.searchEngine.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	hits := result["hits"].(map[string]interface{})
	var products []ProductDocument
	
	for _, hit := range hits["hits"].([]interface{}) {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		
		var product ProductDocument
		sourceBytes, _ := json.Marshal(source)
		json.Unmarshal(sourceBytes, &product)
		
		products = append(products, product)
	}
	
	return products, nil
}

// GetHotProducts 获取热门商品
func (re *RecommendationEngine) GetHotProducts(categoryID uint, size int) ([]ProductDocument, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"status": 1,
						},
					},
				},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"sales_count": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"view_count": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": size,
	}
	
	// 如果指定了分类，添加分类过滤
	if categoryID > 0 {
		boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
		filters := boolQuery["filter"].([]interface{})
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": categoryID,
			},
		})
		boolQuery["filter"] = filters
	}
	
	queryBytes, _ := json.Marshal(query)
	
	req := esapi.SearchRequest{
		Index: []string{"products"},
		Body:  strings.NewReader(string(queryBytes)),
	}
	
	res, err := req.Do(re.searchEngine.ctx, re.searchEngine.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	hits := result["hits"].(map[string]interface{})
	var products []ProductDocument
	
	for _, hit := range hits["hits"].([]interface{}) {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		
		var product ProductDocument
		sourceBytes, _ := json.Marshal(source)
		json.Unmarshal(sourceBytes, &product)
		
		products = append(products, product)
	}
	
	return products, nil
}
