package memory

import (
	"context"
	"fmt"
	"log"

	"eino-minimal/multiagent/routing"

	"github.com/cloudwego/eino/schema"
)

type BuildMeta struct {
	HistoryTurns    int  `json:"history_turns"`
	MaxTurns        int  `json:"max_turns"`
	SummaryInjected bool `json:"summary_injected"`
	RecallCount     int  `json:"recall_count"`
	TokenEstimate   int  `json:"token_estimate"`
	TokenBudget     int  `json:"token_budget"`
}

type MemoryManager struct {
	storage   MemoryStorage
	longStore LongTermStore
}

func NewMemoryManager(storage MemoryStorage, longStore LongTermStore) *MemoryManager {
	return &MemoryManager{
		storage:   storage,
		longStore: longStore,
	}
}

func (m *MemoryManager) AppendTurn(ctx context.Context, sid string, turn Turn) error {
	return m.storage.AppendTurn(ctx, sid, turn)
}

func (m *MemoryManager) GetSession(ctx context.Context, sid string) (*SessionState, error) {
	return m.storage.LoadSession(ctx, sid)
}

func (m *MemoryManager) StoreFact(ctx context.Context, namespace, key, content string, data map[string]any) error {
	if m.longStore == nil {
		return nil
	}
	return m.longStore.PutFact(ctx, namespace, key, content, data)
}

func (m *MemoryManager) BuildContext(
	ctx context.Context,
	profile *routing.ProfileConfig,
	sessionID string,
	systemPrompt string,
) ([]*schema.Message, *BuildMeta, error) {
	state, err := m.storage.LoadSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("load session: %w", err)
	}

	meta := &BuildMeta{
		MaxTurns: profile.MaxHistoryTurns,
	}

	msgs := make([]*schema.Message, 0, profile.MaxHistoryTurns+2)

	sysContent := systemPrompt
	if profile.UseSummary && state.ShortSummary != "" {
		sysContent += "\n\n## 对话摘要\n" + state.ShortSummary
		meta.SummaryInjected = true
	}

	if profile.UseSemanticRecall && m.longStore != nil {
		lastUserMsg := findLastUserContent(state.Turns)
		if lastUserMsg != "" {
			namespace := "session:" + sessionID
			facts, err := m.longStore.SearchFacts(ctx, namespace, lastUserMsg, profile.MaxRecallFacts)
			if err != nil {
				log.Printf("[MemoryManager] semantic recall failed: %v", err)
			} else if len(facts) > 0 {
				sysContent += "\n\n## 相关记忆\n"
				for _, f := range facts {
					sysContent += fmt.Sprintf("- %s\n", f.Content)
				}
				meta.RecallCount = len(facts)
			}
		}
	}

	msgs = append(msgs, &schema.Message{Role: schema.System, Content: sysContent})

	turns := pickLastNTurns(state.Turns, profile.MaxHistoryTurns)
	tokenBudget := int(float64(profile.MaxTokens) * 0.7)
	sysTokens := EstimateTokens(sysContent, profile.Provider)
	remaining := tokenBudget - sysTokens
	totalTokens := sysTokens

	actualTurns := 0
	for _, t := range turns {
		tTokens := EstimateTokens(t.Content, profile.Provider)
		if remaining-tTokens < 0 {
			break
		}
		remaining -= tTokens
		totalTokens += tTokens
		actualTurns++

		role := schema.User
		switch t.Role {
		case "assistant":
			role = schema.Assistant
		case "tool":
			role = schema.Tool
		}
		msgs = append(msgs, &schema.Message{Role: role, Content: t.Content})
	}

	meta.HistoryTurns = actualTurns
	meta.TokenEstimate = totalTokens
	meta.TokenBudget = tokenBudget

	return msgs, meta, nil
}

func findLastUserContent(turns []Turn) string {
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].Role == "user" {
			return turns[i].Content
		}
	}
	return ""
}

func pickLastNTurns(turns []Turn, n int) []Turn {
	if n <= 0 || len(turns) == 0 {
		return nil
	}
	if len(turns) <= n {
		return turns
	}
	return turns[len(turns)-n:]
}
