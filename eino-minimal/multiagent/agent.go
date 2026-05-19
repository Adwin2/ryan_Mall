package multiagent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"eino-minimal/multiagent/memory"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func RunOpsAgent(ctx context.Context, infra *AgentInfra, sessionID, message string, eventCh chan<- Event) (*schema.StreamReader[*schema.Message], error) {
	return runAgent(ctx, infra, "ops", sessionID, message, opsSystemPrompt, OpsTools(), eventCh)
}

func RunConsumerAgent(ctx context.Context, infra *AgentInfra, sessionID, message string, eventCh chan<- Event) (*schema.StreamReader[*schema.Message], error) {
	return runAgent(ctx, infra, "consumer", sessionID, message, consumerSystemPrompt, ConsumerTools(), eventCh)
}

func runAgent(
	ctx context.Context,
	infra *AgentInfra,
	agentID, sessionID, message, systemPrompt string,
	tools []tool.BaseTool,
	eventCh chan<- Event,
) (*schema.StreamReader[*schema.Message], error) {
	start := time.Now()

	// Step 1: Route
	eventCh <- Event{Type: "progress", Node: "ModelRouter", Status: "start"}
	decision, chatModel, err := infra.Router.Route(ctx, agentID, message)
	if err != nil {
		return nil, fmt.Errorf("route: %w", err)
	}
	eventCh <- Event{
		Type:    "routing",
		Tier:    string(decision.Tier),
		Profile: decision.ProfileName,
		Model:   decision.Model,
		Reason:  decision.Reason,
	}
	eventCh <- Event{Type: "progress", Node: "ModelRouter", Status: "done"}

	// Step 2: BuildContext
	eventCh <- Event{Type: "progress", Node: "ContextBuilder", Status: "start"}
	msgs, meta, err := infra.Memory.BuildContext(ctx, decision.Profile, sessionID, systemPrompt)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}
	if meta != nil {
		metaJSON, _ := json.Marshal(meta)
		eventCh <- Event{Type: "context", Content: string(metaJSON)}
	}
	msgs = append(msgs, &schema.Message{Role: schema.User, Content: message})
	eventCh <- Event{Type: "progress", Node: "ContextBuilder", Status: "done"}

	// Step 3: Run ReAct Agent with MessageFuture for tool observability
	eventCh <- Event{Type: "progress", Node: "ReActAgent", Status: "start"}
	ag, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: tools},
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	opt, future := react.WithMessageFuture()
	sr, err := ag.Stream(ctx, msgs, opt)
	if err != nil {
		return nil, fmt.Errorf("agent stream: %w", err)
	}

	// Goroutine: observe intermediate messages (tool calls, results, thinking) via MessageFuture
	go func() {
		iter := future.GetMessageStreams()
		for {
			sReader, ok, err := iter.Next()
			if !ok || err != nil {
				return
			}
			msg, err := schema.ConcatMessageStream(sReader)
			if err != nil {
				continue
			}

			// Emit thinking: assistant reasoning before tool calls
			if msg.Role == schema.Assistant && msg.Content != "" {
				eventCh <- Event{
					Type:    "thinking",
					Content: msg.Content,
				}
			}

			// Emit model-native reasoning (Ark/Doubao deep thinking)
			if reasoning, ok := msg.Extra["ark-reasoning-content"].(string); ok && reasoning != "" {
				eventCh <- Event{
					Type:    "thinking",
					Content: reasoning,
				}
			}

			for _, tc := range msg.ToolCalls {
				if tc.Function.Name != "" {
					eventCh <- Event{
						Type:    "tool_call",
						Tool:    tc.Function.Name,
						Content: tc.Function.Arguments,
					}
				}
			}
			if msg.Role == schema.Tool && msg.Content != "" {
				eventCh <- Event{
					Type:   "tool_result",
					Result: truncate(msg.Content, 300),
				}
			}
		}
	}()

	// Copy stream: one for caller, one for memory persistence
	srs := sr.Copy(2)

	go func() {
		var fullContent string
		defer func() {
			srs[1].Close()

			latency := time.Since(start).Milliseconds()
			tokens := int64(memory.EstimateTokens(fullContent, decision.Provider))
			infra.Router.GetMetrics().Record(decision, latency, tokens, true)

			infra.Memory.AppendTurn(ctx, sessionID, memory.Turn{
				ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
				Role: "user", Content: message,
				Model: decision.Model, Tier: string(decision.Tier),
				Ts: start,
			})
			infra.Memory.AppendTurn(ctx, sessionID, memory.Turn{
				ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
				Role: "assistant", Content: fullContent,
				Model: decision.Model, Tier: string(decision.Tier),
				Ts: time.Now(),
			})

			eventCh <- Event{Type: "progress", Node: "ReActAgent", Status: "done"}
			eventCh <- Event{Type: "done"}
		}()

		for {
			chunk, err := srs[1].Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				log.Printf("[%s] stream error: %v", agentID, err)
				return
			}
			fullContent += chunk.Content
		}
	}()

	return srs[0], nil
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}
