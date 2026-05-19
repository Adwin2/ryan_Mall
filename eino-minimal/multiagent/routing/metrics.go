package routing

import (
	"sync"
	"sync/atomic"
	"time"
)

type ProfileMetrics struct {
	CallCount      int64            `json:"call_count"`
	SuccessCount   int64            `json:"success_count"`
	FailCount      int64            `json:"fail_count"`
	TotalLatencyMs int64            `json:"total_latency_ms"`
	TotalTokens    int64            `json:"total_tokens"`
	TierDist       map[string]int64 `json:"tier_distribution"`
	LastCalledAt   time.Time        `json:"last_called_at"`
}

type Metrics struct {
	mu       sync.RWMutex
	counters map[string]*ProfileMetrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		counters: make(map[string]*ProfileMetrics),
	}
}

func (m *Metrics) Record(decision *RouteDecision, latencyMs int64, tokens int64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pm, ok := m.counters[decision.ProfileName]
	if !ok {
		pm = &ProfileMetrics{TierDist: make(map[string]int64)}
		m.counters[decision.ProfileName] = pm
	}

	atomic.AddInt64(&pm.CallCount, 1)
	if success {
		atomic.AddInt64(&pm.SuccessCount, 1)
	} else {
		atomic.AddInt64(&pm.FailCount, 1)
	}
	atomic.AddInt64(&pm.TotalLatencyMs, latencyMs)
	atomic.AddInt64(&pm.TotalTokens, tokens)
	pm.TierDist[string(decision.Tier)]++
	pm.LastCalledAt = time.Now()
}

func (m *Metrics) Snapshot() map[string]*ProfileMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap := make(map[string]*ProfileMetrics, len(m.counters))
	for k, v := range m.counters {
		cp := &ProfileMetrics{
			CallCount:      atomic.LoadInt64(&v.CallCount),
			SuccessCount:   atomic.LoadInt64(&v.SuccessCount),
			FailCount:      atomic.LoadInt64(&v.FailCount),
			TotalLatencyMs: atomic.LoadInt64(&v.TotalLatencyMs),
			TotalTokens:    atomic.LoadInt64(&v.TotalTokens),
			TierDist:       make(map[string]int64),
			LastCalledAt:   v.LastCalledAt,
		}
		for tier, count := range v.TierDist {
			cp.TierDist[tier] = count
		}
		snap[k] = cp
	}
	return snap
}
