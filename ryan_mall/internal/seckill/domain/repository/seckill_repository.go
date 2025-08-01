package repository

import (
	"context"
	"time"

	"ryan-mall-microservices/internal/seckill/domain/entity"
	"ryan-mall-microservices/internal/shared/domain"
)

// SeckillActivityRepository 秒杀活动仓储接口
type SeckillActivityRepository interface {
	// Save 保存秒杀活动
	Save(ctx context.Context, activity *entity.SeckillActivity) error
	
	// FindByID 根据ID查找秒杀活动
	FindByID(ctx context.Context, id domain.ID) (*entity.SeckillActivity, error)
	
	// FindByProductID 根据商品ID查找秒杀活动
	FindByProductID(ctx context.Context, productID domain.ProductID) ([]*entity.SeckillActivity, error)
	
	// FindActiveActivities 查找激活的秒杀活动
	FindActiveActivities(ctx context.Context, now time.Time) ([]*entity.SeckillActivity, error)
	
	// FindUpcomingActivities 查找即将开始的秒杀活动
	FindUpcomingActivities(ctx context.Context, now time.Time, limit int) ([]*entity.SeckillActivity, error)
	
	// Update 更新秒杀活动
	Update(ctx context.Context, activity *entity.SeckillActivity) error
	
	// Delete 删除秒杀活动
	Delete(ctx context.Context, id domain.ID) error
	
	// List 分页查询秒杀活动列表
	List(ctx context.Context, offset, limit int) ([]*entity.SeckillActivity, int64, error)
	
	// FindByTimeRange 根据时间范围查找活动
	FindByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*entity.SeckillActivity, error)
}

// SeckillOrderRepository 秒杀订单仓储接口
type SeckillOrderRepository interface {
	// Save 保存秒杀订单
	Save(ctx context.Context, order *entity.SeckillOrder) error
	
	// FindByID 根据ID查找秒杀订单
	FindByID(ctx context.Context, id domain.OrderID) (*entity.SeckillOrder, error)
	
	// FindByUserID 根据用户ID查找秒杀订单
	FindByUserID(ctx context.Context, userID domain.UserID, offset, limit int) ([]*entity.SeckillOrder, int64, error)
	
	// FindByActivityID 根据活动ID查找秒杀订单
	FindByActivityID(ctx context.Context, activityID domain.ID, offset, limit int) ([]*entity.SeckillOrder, int64, error)
	
	// FindByUserAndActivity 根据用户ID和活动ID查找订单
	FindByUserAndActivity(ctx context.Context, userID domain.UserID, activityID domain.ID) ([]*entity.SeckillOrder, error)
	
	// FindPendingOrders 查找待支付订单
	FindPendingOrders(ctx context.Context, createdBefore time.Time) ([]*entity.SeckillOrder, error)
	
	// Update 更新秒杀订单
	Update(ctx context.Context, order *entity.SeckillOrder) error
	
	// Delete 删除秒杀订单
	Delete(ctx context.Context, id domain.OrderID) error
	
	// CountByActivity 统计活动订单数量
	CountByActivity(ctx context.Context, activityID domain.ID) (int64, error)
	
	// CountByActivityAndStatus 根据活动和状态统计订单数量
	CountByActivityAndStatus(ctx context.Context, activityID domain.ID, status entity.SeckillOrderStatus) (int64, error)
}

// SeckillQueryRepository 秒杀查询仓储接口（读模型）
type SeckillQueryRepository interface {
	// GetActivityDetail 获取活动详情
	GetActivityDetail(ctx context.Context, activityID string) (*ActivityDetailView, error)
	
	// ListActiveActivities 获取激活的活动列表
	ListActiveActivities(ctx context.Context, offset, limit int) ([]*ActivityListView, int64, error)
	
	// ListUpcomingActivities 获取即将开始的活动列表
	ListUpcomingActivities(ctx context.Context, offset, limit int) ([]*ActivityListView, int64, error)
	
	// GetUserSeckillOrders 获取用户秒杀订单
	GetUserSeckillOrders(ctx context.Context, userID string, offset, limit int) ([]*SeckillOrderView, int64, error)
	
	// GetActivityStatistics 获取活动统计信息
	GetActivityStatistics(ctx context.Context, activityID string) (*ActivityStatisticsView, error)
	
	// SearchActivities 搜索活动
	SearchActivities(ctx context.Context, criteria *ActivitySearchCriteria) ([]*ActivityListView, int64, error)
}

// ActivityDetailView 活动详情视图（读模型）
type ActivityDetailView struct {
	ActivityID     string  `json:"activity_id"`
	Name           string  `json:"name"`
	ProductID      string  `json:"product_id"`
	ProductName    string  `json:"product_name"`
	OriginalPrice  float64 `json:"original_price"`
	SeckillPrice   float64 `json:"seckill_price"`
	DiscountRate   float64 `json:"discount_rate"`
	TotalStock     int     `json:"total_stock"`
	RemainingStock int     `json:"remaining_stock"`
	SoldCount      int     `json:"sold_count"`
	Status         string  `json:"status"`
	StartTime      int64   `json:"start_time"`
	EndTime        int64   `json:"end_time"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

// ActivityListView 活动列表视图（读模型）
type ActivityListView struct {
	ActivityID    string  `json:"activity_id"`
	Name          string  `json:"name"`
	ProductName   string  `json:"product_name"`
	SeckillPrice  float64 `json:"seckill_price"`
	OriginalPrice float64 `json:"original_price"`
	DiscountRate  float64 `json:"discount_rate"`
	TotalStock    int     `json:"total_stock"`
	SoldCount     int     `json:"sold_count"`
	Status        string  `json:"status"`
	StartTime     int64   `json:"start_time"`
	EndTime       int64   `json:"end_time"`
}

// SeckillOrderView 秒杀订单视图（读模型）
type SeckillOrderView struct {
	OrderID      string  `json:"order_id"`
	ActivityID   string  `json:"activity_id"`
	ActivityName string  `json:"activity_name"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	TotalAmount  float64 `json:"total_amount"`
	Status       string  `json:"status"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}

// ActivityStatisticsView 活动统计视图
type ActivityStatisticsView struct {
	ActivityID       string  `json:"activity_id"`
	TotalStock       int     `json:"total_stock"`
	RemainingStock   int     `json:"remaining_stock"`
	SoldCount        int     `json:"sold_count"`
	SuccessRate      float64 `json:"success_rate"`
	TotalOrders      int64   `json:"total_orders"`
	PaidOrders       int64   `json:"paid_orders"`
	CancelledOrders  int64   `json:"cancelled_orders"`
	ExpiredOrders    int64   `json:"expired_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	Status           string  `json:"status"`
}

// ActivitySearchCriteria 活动搜索条件
type ActivitySearchCriteria struct {
	Keyword   string `json:"keyword"`
	Status    string `json:"status"`
	StartDate int64  `json:"start_date"`
	EndDate   int64  `json:"end_date"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}
