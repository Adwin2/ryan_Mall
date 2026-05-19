package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"

	"eino-minimal/multiagent"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/eino/schema"
	"github.com/hertz-contrib/sse"
)

func OpsStreamHandler(infra *multiagent.AgentInfra) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		streamAgent(c, ctx, infra, "ops")
	}
}

func ShopStreamHandler(infra *multiagent.AgentInfra) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		streamAgent(c, ctx, infra, "consumer")
	}
}

func RoutingMetricsHandler(infra *multiagent.AgentInfra) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		snap := infra.Router.GetMetrics().Snapshot()
		ctx.JSON(consts.StatusOK, snap)
	}
}

func RoutingReloadHandler(infra *multiagent.AgentInfra) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		if err := infra.Router.Reload(); err != nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]string{"status": "reloaded"})
	}
}

func streamAgent(c context.Context, ctx *app.RequestContext, infra *multiagent.AgentInfra, agentID string) {
	sessionID := string(ctx.QueryArgs().Peek("session_id"))
	message := string(ctx.QueryArgs().Peek("message"))
	if sessionID == "" || message == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "session_id and message required"})
		return
	}

	s := sse.NewStream(ctx)
	eventCh := make(chan multiagent.Event, 50)

	var sr *schema.StreamReader[*schema.Message]
	var err error

	switch agentID {
	case "ops":
		sr, err = multiagent.RunOpsAgent(c, infra, sessionID, message, eventCh)
	case "consumer":
		sr, err = multiagent.RunConsumerAgent(c, infra, sessionID, message, eventCh)
	}

	if err != nil {
		publishMAEvent(s, multiagent.Event{Type: "error", Content: err.Error()})
		return
	}

	drainMAEvents(s, eventCh)

	for {
		msg, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			publishMAEvent(s, multiagent.Event{Type: "error", Content: err.Error()})
			break
		}
		if msg != nil && msg.Content != "" {
			publishMAEvent(s, multiagent.Event{Type: "content", Content: msg.Content})
		}
		drainMAEvents(s, eventCh)
	}

	drainMAEvents(s, eventCh)
}

func drainMAEvents(s *sse.Stream, ch <-chan multiagent.Event) {
	for {
		select {
		case ev := <-ch:
			publishMAEvent(s, ev)
		default:
			return
		}
	}
}

func publishMAEvent(s *sse.Stream, ev multiagent.Event) {
	data, _ := json.Marshal(ev)
	err := s.Publish(&sse.Event{
		Data: data,
	})
	if err != nil {
		log.Printf("[MultiAgent SSE] publish error: %v", err)
	}
}
