package query

import (
	"context"

	"ryan-mall-microservices/internal/seckill/domain/repository"
	"ryan-mall-microservices/internal/seckill/domain/service"
	"ryan-mall-microservices/internal/shared/domain"
)

// GetSeckillActivityDetailQuery 获取秒杀活动详情查询
type GetSeckillActivityDetailQuery struct {
	ActivityID string `json:"activity_id" validate:"required"`
}

// GetSeckillActivityDetailHandler 获取秒杀活动详情查询处理器
type GetSeckillActivityDetailHandler struct {
	queryRepo     repository.SeckillQueryRepository
	domainService *service.SeckillDomainService
}

// NewGetSeckillActivityDetailHandler 创建获取秒杀活动详情查询处理器
func NewGetSeckillActivityDetailHandler(
	queryRepo repository.SeckillQueryRepository,
	domainService *service.SeckillDomainService,
) *GetSeckillActivityDetailHandler {
	return &GetSeckillActivityDetailHandler{
		queryRepo:     queryRepo,
		domainService: domainService,
	}
}

// Handle 处理获取秒杀活动详情查询
func (h *GetSeckillActivityDetailHandler) Handle(ctx context.Context, query *GetSeckillActivityDetailQuery) (*repository.ActivityDetailView, error) {
	// 获取活动详情
	detail, err := h.queryRepo.GetActivityDetail(ctx, query.ActivityID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get activity detail", err)
	}
	if detail == nil {
		return nil, domain.NewNotFoundError("seckill activity", query.ActivityID)
	}

	// 获取实时统计信息
	statistics, err := h.domainService.CalculateActivityStatistics(ctx, domain.ID(query.ActivityID))
	if err == nil && statistics != nil {
		// 更新实时库存信息
		detail.RemainingStock = statistics.RemainingStock
		detail.SoldCount = statistics.SoldCount
	}

	return detail, nil
}

// ListActiveSeckillActivitiesQuery 获取激活的秒杀活动列表查询
type ListActiveSeckillActivitiesQuery struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// ListActiveSeckillActivitiesResult 激活的秒杀活动列表结果
type ListActiveSeckillActivitiesResult struct {
	Activities []*repository.ActivityListView `json:"activities"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// ListActiveSeckillActivitiesHandler 获取激活的秒杀活动列表查询处理器
type ListActiveSeckillActivitiesHandler struct {
	queryRepo repository.SeckillQueryRepository
}

// NewListActiveSeckillActivitiesHandler 创建获取激活的秒杀活动列表查询处理器
func NewListActiveSeckillActivitiesHandler(queryRepo repository.SeckillQueryRepository) *ListActiveSeckillActivitiesHandler {
	return &ListActiveSeckillActivitiesHandler{
		queryRepo: queryRepo,
	}
}

// Handle 处理获取激活的秒杀活动列表查询
func (h *ListActiveSeckillActivitiesHandler) Handle(ctx context.Context, query *ListActiveSeckillActivitiesQuery) (*ListActiveSeckillActivitiesResult, error) {
	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 查询激活的活动列表
	activities, total, err := h.queryRepo.ListActiveActivities(ctx, offset, query.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to list active activities", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &ListActiveSeckillActivitiesResult{
		Activities: activities,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListUpcomingSeckillActivitiesQuery 获取即将开始的秒杀活动列表查询
type ListUpcomingSeckillActivitiesQuery struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// ListUpcomingSeckillActivitiesResult 即将开始的秒杀活动列表结果
type ListUpcomingSeckillActivitiesResult struct {
	Activities []*repository.ActivityListView `json:"activities"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// ListUpcomingSeckillActivitiesHandler 获取即将开始的秒杀活动列表查询处理器
type ListUpcomingSeckillActivitiesHandler struct {
	queryRepo repository.SeckillQueryRepository
}

// NewListUpcomingSeckillActivitiesHandler 创建获取即将开始的秒杀活动列表查询处理器
func NewListUpcomingSeckillActivitiesHandler(queryRepo repository.SeckillQueryRepository) *ListUpcomingSeckillActivitiesHandler {
	return &ListUpcomingSeckillActivitiesHandler{
		queryRepo: queryRepo,
	}
}

// Handle 处理获取即将开始的秒杀活动列表查询
func (h *ListUpcomingSeckillActivitiesHandler) Handle(ctx context.Context, query *ListUpcomingSeckillActivitiesQuery) (*ListUpcomingSeckillActivitiesResult, error) {
	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 查询即将开始的活动列表
	activities, total, err := h.queryRepo.ListUpcomingActivities(ctx, offset, query.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to list upcoming activities", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &ListUpcomingSeckillActivitiesResult{
		Activities: activities,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserSeckillOrdersQuery 获取用户秒杀订单查询
type GetUserSeckillOrdersQuery struct {
	UserID   string `json:"user_id" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// GetUserSeckillOrdersResult 用户秒杀订单结果
type GetUserSeckillOrdersResult struct {
	Orders     []*repository.SeckillOrderView `json:"orders"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// GetUserSeckillOrdersHandler 获取用户秒杀订单查询处理器
type GetUserSeckillOrdersHandler struct {
	queryRepo repository.SeckillQueryRepository
}

// NewGetUserSeckillOrdersHandler 创建获取用户秒杀订单查询处理器
func NewGetUserSeckillOrdersHandler(queryRepo repository.SeckillQueryRepository) *GetUserSeckillOrdersHandler {
	return &GetUserSeckillOrdersHandler{
		queryRepo: queryRepo,
	}
}

// Handle 处理获取用户秒杀订单查询
func (h *GetUserSeckillOrdersHandler) Handle(ctx context.Context, query *GetUserSeckillOrdersQuery) (*GetUserSeckillOrdersResult, error) {
	// 计算偏移量
	offset := (query.Page - 1) * query.PageSize

	// 查询用户秒杀订单
	orders, total, err := h.queryRepo.GetUserSeckillOrders(ctx, query.UserID, offset, query.PageSize)
	if err != nil {
		return nil, domain.NewInternalError("failed to get user seckill orders", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &GetUserSeckillOrdersResult{
		Orders:     orders,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetSeckillActivityStatisticsQuery 获取秒杀活动统计查询
type GetSeckillActivityStatisticsQuery struct {
	ActivityID string `json:"activity_id" validate:"required"`
}

// GetSeckillActivityStatisticsHandler 获取秒杀活动统计查询处理器
type GetSeckillActivityStatisticsHandler struct {
	queryRepo     repository.SeckillQueryRepository
	domainService *service.SeckillDomainService
}

// NewGetSeckillActivityStatisticsHandler 创建获取秒杀活动统计查询处理器
func NewGetSeckillActivityStatisticsHandler(
	queryRepo repository.SeckillQueryRepository,
	domainService *service.SeckillDomainService,
) *GetSeckillActivityStatisticsHandler {
	return &GetSeckillActivityStatisticsHandler{
		queryRepo:     queryRepo,
		domainService: domainService,
	}
}

// Handle 处理获取秒杀活动统计查询
func (h *GetSeckillActivityStatisticsHandler) Handle(ctx context.Context, query *GetSeckillActivityStatisticsQuery) (*repository.ActivityStatisticsView, error) {
	// 获取基础统计信息
	statistics, err := h.queryRepo.GetActivityStatistics(ctx, query.ActivityID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get activity statistics", err)
	}
	if statistics == nil {
		return nil, domain.NewNotFoundError("activity statistics", query.ActivityID)
	}

	// 获取实时统计信息
	realTimeStats, err := h.domainService.CalculateActivityStatistics(ctx, domain.ID(query.ActivityID))
	if err == nil && realTimeStats != nil {
		// 更新实时数据
		statistics.RemainingStock = realTimeStats.RemainingStock
		statistics.SoldCount = realTimeStats.SoldCount
		statistics.SuccessRate = realTimeStats.SuccessRate
	}

	return statistics, nil
}

// SearchSeckillActivitiesQuery 搜索秒杀活动查询
type SearchSeckillActivitiesQuery struct {
	Keyword   string  `json:"keyword"`
	Status    string  `json:"status"`
	StartDate int64   `json:"start_date"`
	EndDate   int64   `json:"end_date"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	Page      int     `json:"page" validate:"min=1"`
	PageSize  int     `json:"page_size" validate:"min=1,max=100"`
}

// SearchSeckillActivitiesResult 搜索秒杀活动结果
type SearchSeckillActivitiesResult struct {
	Activities []*repository.ActivityListView `json:"activities"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

// SearchSeckillActivitiesHandler 搜索秒杀活动查询处理器
type SearchSeckillActivitiesHandler struct {
	queryRepo repository.SeckillQueryRepository
}

// NewSearchSeckillActivitiesHandler 创建搜索秒杀活动查询处理器
func NewSearchSeckillActivitiesHandler(queryRepo repository.SeckillQueryRepository) *SearchSeckillActivitiesHandler {
	return &SearchSeckillActivitiesHandler{
		queryRepo: queryRepo,
	}
}

// Handle 处理搜索秒杀活动查询
func (h *SearchSeckillActivitiesHandler) Handle(ctx context.Context, query *SearchSeckillActivitiesQuery) (*SearchSeckillActivitiesResult, error) {
	// 构建搜索条件
	criteria := &repository.ActivitySearchCriteria{
		Keyword:   query.Keyword,
		Status:    query.Status,
		StartDate: query.StartDate,
		EndDate:   query.EndDate,
		MinPrice:  query.MinPrice,
		MaxPrice:  query.MaxPrice,
		Offset:    (query.Page - 1) * query.PageSize,
		Limit:     query.PageSize,
	}

	// 搜索活动
	activities, total, err := h.queryRepo.SearchActivities(ctx, criteria)
	if err != nil {
		return nil, domain.NewInternalError("failed to search activities", err)
	}

	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	return &SearchSeckillActivitiesResult{
		Activities: activities,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}
