package service

import (
	"context"

	"ryan-mall-microservices/internal/seckill/application/command"
	"ryan-mall-microservices/internal/seckill/application/query"
	"ryan-mall-microservices/internal/seckill/domain/repository"
	"ryan-mall-microservices/internal/seckill/domain/service"
	"ryan-mall-microservices/internal/shared/events"

	"github.com/go-redis/redis/v8"
)

// SeckillApplicationService 秒杀应用服务（CQRS模式）
type SeckillApplicationService struct {
	// 命令处理器（写操作）
	createActivityHandler      *command.CreateSeckillActivityHandler
	startActivityHandler       *command.StartSeckillActivityHandler
	participateInSeckillHandler *command.ParticipateInSeckillHandler
	payOrderHandler            *command.PaySeckillOrderHandler
	cancelOrderHandler         *command.CancelSeckillOrderHandler

	// 查询处理器（读操作）
	getActivityDetailHandler       *query.GetSeckillActivityDetailHandler
	listActiveActivitiesHandler    *query.ListActiveSeckillActivitiesHandler
	listUpcomingActivitiesHandler  *query.ListUpcomingSeckillActivitiesHandler
	getUserOrdersHandler           *query.GetUserSeckillOrdersHandler
	getActivityStatisticsHandler   *query.GetSeckillActivityStatisticsHandler
	searchActivitiesHandler        *query.SearchSeckillActivitiesHandler
}

// NewSeckillApplicationService 创建秒杀应用服务
func NewSeckillApplicationService(
	activityRepo repository.SeckillActivityRepository,
	orderRepo repository.SeckillOrderRepository,
	queryRepo repository.SeckillQueryRepository,
	redisClient *redis.Client,
	eventPublisher *events.EventPublisher,
) *SeckillApplicationService {
	// 创建领域服务
	domainService := service.NewSeckillDomainService(activityRepo, orderRepo, redisClient, nil)

	return &SeckillApplicationService{
		// 初始化命令处理器
		createActivityHandler:       command.NewCreateSeckillActivityHandler(activityRepo, domainService, eventPublisher),
		startActivityHandler:        command.NewStartSeckillActivityHandler(activityRepo, eventPublisher),
		participateInSeckillHandler: command.NewParticipateInSeckillHandler(domainService, eventPublisher),
		payOrderHandler:             command.NewPaySeckillOrderHandler(orderRepo, eventPublisher),
		cancelOrderHandler:          command.NewCancelSeckillOrderHandler(orderRepo, domainService, eventPublisher),

		// 初始化查询处理器
		getActivityDetailHandler:      query.NewGetSeckillActivityDetailHandler(queryRepo, domainService),
		listActiveActivitiesHandler:   query.NewListActiveSeckillActivitiesHandler(queryRepo),
		listUpcomingActivitiesHandler: query.NewListUpcomingSeckillActivitiesHandler(queryRepo),
		getUserOrdersHandler:          query.NewGetUserSeckillOrdersHandler(queryRepo),
		getActivityStatisticsHandler:  query.NewGetSeckillActivityStatisticsHandler(queryRepo, domainService),
		searchActivitiesHandler:       query.NewSearchSeckillActivitiesHandler(queryRepo),
	}
}

// ========== 命令操作（写操作）==========

// CreateSeckillActivity 创建秒杀活动
func (s *SeckillApplicationService) CreateSeckillActivity(ctx context.Context, cmd *command.CreateSeckillActivityCommand) (*command.CreateSeckillActivityResult, error) {
	return s.createActivityHandler.Handle(ctx, cmd)
}

// StartSeckillActivity 启动秒杀活动
func (s *SeckillApplicationService) StartSeckillActivity(ctx context.Context, cmd *command.StartSeckillActivityCommand) error {
	return s.startActivityHandler.Handle(ctx, cmd)
}

// ParticipateInSeckill 参与秒杀
func (s *SeckillApplicationService) ParticipateInSeckill(ctx context.Context, cmd *command.ParticipateInSeckillCommand) (*command.ParticipateInSeckillResult, error) {
	return s.participateInSeckillHandler.Handle(ctx, cmd)
}

// PaySeckillOrder 支付秒杀订单
func (s *SeckillApplicationService) PaySeckillOrder(ctx context.Context, cmd *command.PaySeckillOrderCommand) error {
	return s.payOrderHandler.Handle(ctx, cmd)
}

// CancelSeckillOrder 取消秒杀订单
func (s *SeckillApplicationService) CancelSeckillOrder(ctx context.Context, cmd *command.CancelSeckillOrderCommand) error {
	return s.cancelOrderHandler.Handle(ctx, cmd)
}

// ========== 查询操作（读操作）==========

// GetSeckillActivityDetail 获取秒杀活动详情
func (s *SeckillApplicationService) GetSeckillActivityDetail(ctx context.Context, query *query.GetSeckillActivityDetailQuery) (*repository.ActivityDetailView, error) {
	return s.getActivityDetailHandler.Handle(ctx, query)
}

// ListActiveSeckillActivities 获取激活的秒杀活动列表
func (s *SeckillApplicationService) ListActiveSeckillActivities(ctx context.Context, query *query.ListActiveSeckillActivitiesQuery) (*query.ListActiveSeckillActivitiesResult, error) {
	return s.listActiveActivitiesHandler.Handle(ctx, query)
}

// ListUpcomingSeckillActivities 获取即将开始的秒杀活动列表
func (s *SeckillApplicationService) ListUpcomingSeckillActivities(ctx context.Context, query *query.ListUpcomingSeckillActivitiesQuery) (*query.ListUpcomingSeckillActivitiesResult, error) {
	return s.listUpcomingActivitiesHandler.Handle(ctx, query)
}

// GetUserSeckillOrders 获取用户秒杀订单
func (s *SeckillApplicationService) GetUserSeckillOrders(ctx context.Context, query *query.GetUserSeckillOrdersQuery) (*query.GetUserSeckillOrdersResult, error) {
	return s.getUserOrdersHandler.Handle(ctx, query)
}

// GetSeckillActivityStatistics 获取秒杀活动统计
func (s *SeckillApplicationService) GetSeckillActivityStatistics(ctx context.Context, query *query.GetSeckillActivityStatisticsQuery) (*repository.ActivityStatisticsView, error) {
	return s.getActivityStatisticsHandler.Handle(ctx, query)
}

// SearchSeckillActivities 搜索秒杀活动
func (s *SeckillApplicationService) SearchSeckillActivities(ctx context.Context, query *query.SearchSeckillActivitiesQuery) (*query.SearchSeckillActivitiesResult, error) {
	return s.searchActivitiesHandler.Handle(ctx, query)
}

// ========== 高级业务操作 ==========

// BatchCreateSeckillActivities 批量创建秒杀活动
func (s *SeckillApplicationService) BatchCreateSeckillActivities(ctx context.Context, commands []*command.CreateSeckillActivityCommand) ([]*command.CreateSeckillActivityResult, error) {
	results := make([]*command.CreateSeckillActivityResult, 0, len(commands))
	
	for _, cmd := range commands {
		result, err := s.CreateSeckillActivity(ctx, cmd)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	
	return results, nil
}

// PrewarmSeckillActivity 预热秒杀活动（缓存预热）
func (s *SeckillApplicationService) PrewarmSeckillActivity(ctx context.Context, activityID string) error {
	// 获取活动详情并缓存
	detailQuery := &query.GetSeckillActivityDetailQuery{ActivityID: activityID}
	_, err := s.GetSeckillActivityDetail(ctx, detailQuery)
	if err != nil {
		return err
	}

	// 预热统计信息
	statsQuery := &query.GetSeckillActivityStatisticsQuery{ActivityID: activityID}
	_, err = s.GetSeckillActivityStatistics(ctx, statsQuery)
	if err != nil {
		return err
	}

	return nil
}

// GetSeckillDashboard 获取秒杀仪表板数据
func (s *SeckillApplicationService) GetSeckillDashboard(ctx context.Context) (*SeckillDashboard, error) {
	// 获取激活的活动
	activeQuery := &query.ListActiveSeckillActivitiesQuery{Page: 1, PageSize: 10}
	activeResult, err := s.ListActiveSeckillActivities(ctx, activeQuery)
	if err != nil {
		return nil, err
	}

	// 获取即将开始的活动
	upcomingQuery := &query.ListUpcomingSeckillActivitiesQuery{Page: 1, PageSize: 10}
	upcomingResult, err := s.ListUpcomingSeckillActivities(ctx, upcomingQuery)
	if err != nil {
		return nil, err
	}

	return &SeckillDashboard{
		ActiveActivities:   activeResult.Activities,
		UpcomingActivities: upcomingResult.Activities,
		TotalActive:        activeResult.Total,
		TotalUpcoming:      upcomingResult.Total,
	}, nil
}

// SeckillDashboard 秒杀仪表板
type SeckillDashboard struct {
	ActiveActivities   []*repository.ActivityListView `json:"active_activities"`
	UpcomingActivities []*repository.ActivityListView `json:"upcoming_activities"`
	TotalActive        int64                          `json:"total_active"`
	TotalUpcoming      int64                          `json:"total_upcoming"`
}
