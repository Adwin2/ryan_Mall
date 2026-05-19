package memory

import "time"

type SessionState struct {
	SessionID    string            `json:"session_id"`
	AgentID      string            `json:"agent_id"`
	Turns        []Turn            `json:"turns"`
	ShortSummary string            `json:"short_summary,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Turn struct {
	ID      string    `json:"id"`
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Model   string    `json:"model,omitempty"`
	Tier    string    `json:"tier,omitempty"`
	Ts      time.Time `json:"ts"`
}

type Fact struct {
	Key       string         `json:"key"`
	Namespace string         `json:"namespace"`
	Content   string         `json:"content"`
	Data      map[string]any `json:"data,omitempty"`
	Score     float64        `json:"score"`
}
