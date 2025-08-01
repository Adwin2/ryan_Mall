package proxy

import (
	"sync"
	"time"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed   CircuitBreakerState = iota // 关闭状态（正常）
	StateOpen                                // 打开状态（熔断）
	StateHalfOpen                            // 半开状态（尝试恢复）
)

// String 返回状态字符串
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           // 失败阈值
	RecoveryTimeout  time.Duration // 恢复超时时间
	SuccessThreshold int           // 成功阈值（半开状态下）
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastSuccessTime time.Time
	nextAttempt     time.Time
	mutex           sync.RWMutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// IsOpen 检查熔断器是否打开
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return false
	case StateOpen:
		// 检查是否可以尝试恢复
		if now.After(cb.nextAttempt) {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			// 双重检查
			if cb.state == StateOpen && now.After(cb.nextAttempt) {
				cb.state = StateHalfOpen
				cb.successCount = 0
			}
			cb.mutex.Unlock()
			cb.mutex.RLock()
		}
		return cb.state == StateOpen
	case StateHalfOpen:
		return false
	default:
		return false
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case StateClosed:
		// 在关闭状态下，重置失败计数
		cb.failureCount = 0
	case StateHalfOpen:
		// 在半开状态下，增加成功计数
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			// 达到成功阈值，转为关闭状态
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.lastFailureTime = time.Now()
	cb.failureCount++

	switch cb.state {
	case StateClosed:
		// 在关闭状态下，检查是否达到失败阈值
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.nextAttempt = time.Now().Add(cb.config.RecoveryTimeout)
		}
	case StateHalfOpen:
		// 在半开状态下，任何失败都会导致重新打开
		cb.state = StateOpen
		cb.nextAttempt = time.Now().Add(cb.config.RecoveryTimeout)
		cb.successCount = 0
	}
}

// State 获取当前状态
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// FailureCount 获取失败计数
func (cb *CircuitBreaker) FailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}

// SuccessCount 获取成功计数
func (cb *CircuitBreaker) SuccessCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.successCount
}

// LastFailureTime 获取最后失败时间
func (cb *CircuitBreaker) LastFailureTime() time.Time {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.lastFailureTime
}

// LastSuccessTime 获取最后成功时间
func (cb *CircuitBreaker) LastSuccessTime() time.Time {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.lastSuccessTime
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.lastSuccessTime = time.Time{}
	cb.nextAttempt = time.Time{}
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_failure_time": cb.lastFailureTime,
		"last_success_time": cb.lastSuccessTime,
		"next_attempt":      cb.nextAttempt,
		"config": map[string]interface{}{
			"failure_threshold": cb.config.FailureThreshold,
			"recovery_timeout":  cb.config.RecoveryTimeout,
			"success_threshold": cb.config.SuccessThreshold,
		},
	}
}

// CircuitBreakerManager 熔断器管理器
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager 创建熔断器管理器
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetCircuitBreaker 获取熔断器
func (cbm *CircuitBreakerManager) GetCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.RLock()
	cb, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()

	if exists {
		return cb
	}

	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// 双重检查
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	cb = NewCircuitBreaker(config)
	cbm.breakers[name] = cb
	return cb
}

// GetAllStats 获取所有熔断器统计信息
func (cbm *CircuitBreakerManager) GetAllStats() map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, cb := range cbm.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// ResetAll 重置所有熔断器
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	for _, cb := range cbm.breakers {
		cb.Reset()
	}
}

// Reset 重置指定熔断器
func (cbm *CircuitBreakerManager) Reset(name string) bool {
	cbm.mutex.RLock()
	cb, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()

	if !exists {
		return false
	}

	cb.Reset()
	return true
}
