package routing

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cloudwego/eino/components/model"
)

type RouteDecision struct {
	ProfileName string         `json:"profile"`
	Profile     *ProfileConfig `json:"-"`
	Provider    string         `json:"provider"`
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens"`
	Tier        ComplexityTier `json:"tier"`
	Reason      string         `json:"reason"`
	Variant     string         `json:"variant"`
	Timestamp   time.Time      `json:"timestamp"`
}

type ModelRouter struct {
	configMgr *ConfigManager
	metrics   *Metrics
	factory   *ModelFactory
}

func NewModelRouter(configPath string) (*ModelRouter, error) {
	cm, err := NewConfigManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("init config: %w", err)
	}

	return &ModelRouter{
		configMgr: cm,
		metrics:   NewMetrics(),
		factory:   NewModelFactory(),
	}, nil
}

func (r *ModelRouter) Route(ctx context.Context, agentID, message string) (*RouteDecision, model.ToolCallingChatModel, error) {
	tier := Classify(message)
	cfg := r.configMgr.Get()

	rule := r.findRule(cfg, agentID, string(tier))
	if rule == nil {
		return nil, nil, fmt.Errorf("no routing rule for agent=%s tier=%s", agentID, tier)
	}

	variant := r.selectVariant(rule.Variants)
	profile, ok := cfg.Profiles[variant.Profile]
	if !ok {
		return nil, nil, fmt.Errorf("profile %q not found", variant.Profile)
	}

	reason := fmt.Sprintf("tier=%s(分类器判定)", tier)
	if tier == TierStrong {
		reason = classifyReason(message)
	}

	decision := &RouteDecision{
		ProfileName: variant.Profile,
		Profile:     profile,
		Provider:    profile.Provider,
		Model:       profile.Model,
		MaxTokens:   profile.MaxTokens,
		Tier:        tier,
		Reason:      reason,
		Variant:     variant.Profile,
		Timestamp:   time.Now(),
	}

	LogDecision(decision, agentID)

	chatModel, err := r.factory.GetModel(ctx, profile)
	if err != nil {
		r.metrics.Record(decision, 0, 0, false)
		return nil, nil, fmt.Errorf("create model: %w", err)
	}

	return decision, chatModel, nil
}

func (r *ModelRouter) Reload() error {
	return r.configMgr.Reload()
}

func (r *ModelRouter) GetMetrics() *Metrics {
	return r.metrics
}

func (r *ModelRouter) Close() {
	r.configMgr.Close()
}

func (r *ModelRouter) findRule(cfg *RoutingConfig, agentID, tier string) *RoutingRule {
	for i := range cfg.Rules {
		if cfg.Rules[i].Agent == agentID && cfg.Rules[i].Tier == tier {
			return &cfg.Rules[i]
		}
	}
	for i := range cfg.Rules {
		if cfg.Rules[i].Agent == "*" && cfg.Rules[i].Tier == tier {
			return &cfg.Rules[i]
		}
	}
	return nil
}

func (r *ModelRouter) selectVariant(variants []RouteVariant) RouteVariant {
	if len(variants) == 1 {
		return variants[0]
	}

	totalWeight := 0
	for _, v := range variants {
		totalWeight += v.Weight
	}

	pick := rand.Intn(totalWeight)
	cumulative := 0
	for _, v := range variants {
		cumulative += v.Weight
		if pick < cumulative {
			return v
		}
	}
	return variants[0]
}

func classifyReason(message string) string {
	for _, kw := range strongKeywords {
		if contains(message, kw) {
			return fmt.Sprintf("含关键词'%s',判定复杂任务", kw)
		}
	}
	if len([]rune(message)) > 50 {
		return "query长度>50字,判定复杂任务"
	}
	return "规则引擎判定复杂任务"
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRune(s, sub))
}

func containsRune(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func LogDecision(d *RouteDecision, agentID string) {
	log.Printf("[ModelRouter] agent=%s tier=%s profile=%s model=%s reason=%q variant=%s",
		agentID, d.Tier, d.ProfileName, d.Model, d.Reason, d.Variant)
}
