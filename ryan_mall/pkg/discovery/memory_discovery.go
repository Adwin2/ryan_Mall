package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ryan-mall-microservices/internal/shared/infrastructure"
)

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	// Register 注册服务
	Register(ctx context.Context, service *ServiceInfo) error

	// Unregister 注销服务
	Unregister(ctx context.Context, serviceName, address string, port int) error

	// Discover 发现服务
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)

	// Watch 监听服务变化
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)

	// Close 关闭服务发现
	Close() error
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID       string            `json:"id"`       // 服务实例ID
	Name     string            `json:"name"`     // 服务名称
	Version  string            `json:"version"`  // 服务版本
	Address  string            `json:"address"`  // 服务地址
	Port     int               `json:"port"`     // 服务端口
	Metadata map[string]string `json:"metadata"` // 元数据
	Health   HealthStatus      `json:"health"`   // 健康状态
	Tags     []string          `json:"tags"`     // 标签
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ServiceHealthy 健康状态常量
const (
	ServiceHealthy   = HealthStatusHealthy
	ServiceUnhealthy = HealthStatusUnhealthy
	ServiceUnknown   = HealthStatusUnknown
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(services []*ServiceInfo) *ServiceInfo
}

// RoundRobinLoadBalancer 轮询负载均衡器
type RoundRobinLoadBalancer struct {
	counter int
	mutex   sync.Mutex
}

// NewRoundRobinLoadBalancer 创建轮询负载均衡器
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

// Select 选择服务实例
func (r *RoundRobinLoadBalancer) Select(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	selected := services[r.counter%len(services)]
	r.counter++
	return selected
}

// ServiceInfoWithHeartbeat 带心跳的服务信息
type ServiceInfoWithHeartbeat struct {
	*ServiceInfo
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// MemoryServiceDiscovery 内存版本的服务发现（用于开发和测试）
type MemoryServiceDiscovery struct {
	services map[string][]*ServiceInfoWithHeartbeat
	mutex    sync.RWMutex
	logger   infrastructure.Logger
}

// NewMemoryServiceDiscovery 创建内存版本的服务发现
func NewMemoryServiceDiscovery() *MemoryServiceDiscovery {
	return &MemoryServiceDiscovery{
		services: make(map[string][]*ServiceInfoWithHeartbeat),
		logger:   infrastructure.GetLogger(),
	}
}

// Register 注册服务
func (m *MemoryServiceDiscovery) Register(ctx context.Context, service *ServiceInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.services[service.Name] == nil {
		m.services[service.Name] = make([]*ServiceInfoWithHeartbeat, 0)
	}

	// 检查是否已存在相同的服务实例
	for _, existing := range m.services[service.Name] {
		if existing.Address == service.Address && existing.Port == service.Port {
			// 更新现有服务信息
			existing.Metadata = service.Metadata
			existing.Health = service.Health
			existing.LastHeartbeat = time.Now()
			m.logger.Info("Service updated",
				infrastructure.String("service", service.Name),
				infrastructure.String("address", service.Address),
				infrastructure.Int("port", service.Port),
			)
			return nil
		}
	}

	// 添加新的服务实例
	serviceWithHeartbeat := &ServiceInfoWithHeartbeat{
		ServiceInfo:   service,
		LastHeartbeat: time.Now(),
	}
	m.services[service.Name] = append(m.services[service.Name], serviceWithHeartbeat)

	m.logger.Info("Service registered",
		infrastructure.String("service", service.Name),
		infrastructure.String("address", service.Address),
		infrastructure.Int("port", service.Port),
	)

	return nil
}

// Unregister 注销服务（实现接口）
func (m *MemoryServiceDiscovery) Unregister(ctx context.Context, serviceName, address string, port int) error {
	return m.Deregister(ctx, serviceName, address, port)
}

// Deregister 注销服务
func (m *MemoryServiceDiscovery) Deregister(ctx context.Context, serviceName, address string, port int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	services, exists := m.services[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	for i, service := range services {
		if service.Address == address && service.Port == port {
			// 移除服务实例
			m.services[serviceName] = append(services[:i], services[i+1:]...)
			m.logger.Info("Service deregistered",
				infrastructure.String("service", serviceName),
				infrastructure.String("address", address),
				infrastructure.Int("port", port),
			)
			return nil
		}
	}

	return fmt.Errorf("service instance %s:%d not found", address, port)
}

// Discover 发现服务
func (m *MemoryServiceDiscovery) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	services, exists := m.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// 过滤健康的服务实例
	healthyServices := make([]*ServiceInfo, 0)
	for _, service := range services {
		if service.Health == ServiceHealthy {
			healthyServices = append(healthyServices, service.ServiceInfo)
		}
	}

	return healthyServices, nil
}

// Watch 监听服务变化
func (m *MemoryServiceDiscovery) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	// 简化实现：返回当前服务列表
	ch := make(chan []*ServiceInfo, 1)
	
	go func() {
		defer close(ch)
		
		// 发送当前服务列表
		services, err := m.Discover(ctx, serviceName)
		if err == nil {
			select {
			case ch <- services:
			case <-ctx.Done():
				return
			}
		}
		
		// 在实际实现中，这里应该监听服务变化
		<-ctx.Done()
	}()
	
	return ch, nil
}

// Close 关闭服务发现
func (m *MemoryServiceDiscovery) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.services = make(map[string][]*ServiceInfoWithHeartbeat)
	m.logger.Info("Memory service discovery closed")
	return nil
}

// GetAllServices 获取所有服务
func (m *MemoryServiceDiscovery) GetAllServices() map[string][]*ServiceInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string][]*ServiceInfo)
	for name, services := range m.services {
		result[name] = make([]*ServiceInfo, len(services))
		for i, service := range services {
			result[name][i] = service.ServiceInfo
		}
	}

	return result
}

// HealthCheck 健康检查
func (m *MemoryServiceDiscovery) HealthCheck(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for serviceName, services := range m.services {
		for _, service := range services {
			// 检查心跳超时（30秒）
			if now.Sub(service.LastHeartbeat) > 30*time.Second {
				service.Health = ServiceUnhealthy
				m.logger.Warn("Service marked as unhealthy due to timeout",
					infrastructure.String("service", serviceName),
					infrastructure.String("address", service.Address),
					infrastructure.Int("port", service.Port),
				)
			}
		}
	}

	return nil
}

// Heartbeat 发送心跳
func (m *MemoryServiceDiscovery) Heartbeat(ctx context.Context, serviceName, address string, port int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	services, exists := m.services[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	for _, service := range services {
		if service.Address == address && service.Port == port {
			service.LastHeartbeat = time.Now()
			service.Health = ServiceHealthy
			return nil
		}
	}

	return fmt.Errorf("service instance %s:%d not found", address, port)
}

// StartHealthChecker 启动健康检查器
func (m *MemoryServiceDiscovery) StartHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.HealthCheck(ctx); err != nil {
				m.logger.Error("Health check failed", infrastructure.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

// PreregisterServices 预注册一些服务（用于开发测试）
func (m *MemoryServiceDiscovery) PreregisterServices() {
	services := []*ServiceInfo{
		{
			Name:    "user-service",
			Address: "localhost",
			Port:    8081,
			Health:  ServiceHealthy,
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		},
		{
			Name:    "product-service",
			Address: "localhost",
			Port:    8082,
			Health:  ServiceHealthy,
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		},
		{
			Name:    "order-service",
			Address: "localhost",
			Port:    8083,
			Health:  ServiceHealthy,
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		},
		{
			Name:    "seckill-service",
			Address: "localhost",
			Port:    8084,
			Health:  ServiceHealthy,
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		},
		{
			Name:    "payment-service",
			Address: "localhost",
			Port:    8085,
			Health:  ServiceHealthy,
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		},
	}

	for _, service := range services {
		m.Register(context.Background(), service)
	}
}
