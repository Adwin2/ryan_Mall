package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"eino-minimal/copilot"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/sse"
)

// CopilotStreamHandler SSE 流式接口 GET /api/copilot/stream
func CopilotStreamHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		sessionID := c.Query("session_id")
		message := c.Query("message")

		if sessionID == "" || message == "" {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "missing session_id or message parameter",
			})
			return
		}

		log.Printf("[Copilot] Stream request - session: %s, message: %s", sessionID, message)

		cfg := copilot.DefaultConfig()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TotalTimeoutSec)*time.Second)
		defer cancel()

		progressCh := make(chan copilot.ProgressEvent, 50)
		sr, err := copilot.RunCopilotStreamWithProgress(timeoutCtx, sessionID, message, progressCh)
		if err != nil {
			log.Printf("[Copilot] Stream error: %v", err)
			c.JSON(consts.StatusInternalServerError, map[string]string{
				"status": "error",
				"error":  err.Error(),
			})
			close(progressCh)
			return
		}

		s := sse.NewStream(c)
		startTime := time.Now()

		// Drain buffered progress events (single goroutine — no concurrent writes to SSE)
		drainProgressChan(s, progressCh)

		// Read stream content
		for {
			select {
			case <-timeoutCtx.Done():
				publishEvent(s, copilot.SSEEvent{Type: "error", Content: "timeout"})
				sr.Close()
				close(progressCh)
				c.Flush()
				return
			default:
				msg, err := sr.Recv()
				if errors.Is(err, io.EOF) {
					sr.Close()
					close(progressCh)
					publishEvent(s, copilot.SSEEvent{Type: "done", Metadata: &copilot.EventMeta{DurationMs: time.Since(startTime).Milliseconds()}})
					c.Flush()
					log.Printf("[Copilot] Stream finished - session: %s, duration: %dms", sessionID, time.Since(startTime).Milliseconds())
					return
				}
				if err != nil {
					log.Printf("[Copilot] Recv error: %v", err)
					publishEvent(s, copilot.SSEEvent{Type: "error", Content: err.Error()})
					sr.Close()
					close(progressCh)
					c.Flush()
					return
				}
				if msg.Content != "" {
					publishEvent(s, copilot.SSEEvent{Type: "content", Content: msg.Content})
				}
			}
		}
	}
}

// CopilotDiagnoseHandler SSE 流式诊断接口 POST /api/copilot/diagnose
func CopilotDiagnoseHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req copilot.DiagnoseRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "invalid request body: " + err.Error(),
			})
			return
		}

		if req.SessionID == "" || req.Message == "" {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "session_id and message are required",
			})
			return
		}

		log.Printf("[Copilot] Diagnose request - session: %s, message: %s", req.SessionID, req.Message)

		cfg := copilot.DefaultConfig()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TotalTimeoutSec)*time.Second)
		defer cancel()

		progressCh := make(chan copilot.ProgressEvent, 100)
		s := sse.NewStream(c)
		startTime := time.Now()

		// Run diagnose stream in background goroutine
		errCh := make(chan error, 1)
		go func() {
			errCh <- copilot.RunDiagnoseStream(timeoutCtx, req.SessionID, req.Message, progressCh)
			close(progressCh)
		}()

		// Consume events from channel and publish SSE (single writer)
		for pe := range progressCh {
			switch pe.Type {
			case "progress":
				publishEvent(s, copilot.SSEEvent{Type: "progress", Node: pe.Node, Status: pe.Status})
			case "content":
				publishEvent(s, copilot.SSEEvent{Type: "content", Content: pe.Data})
			case "diagnosis":
				publishEvent(s, copilot.SSEEvent{Type: "diagnosis", Content: pe.Data, Result: pe.Result})
			}
		}

		// Check for error from RunDiagnoseStream
		if err := <-errCh; err != nil {
			log.Printf("[Copilot] Diagnose error: %v", err)
			publishEvent(s, copilot.SSEEvent{Type: "error", Content: err.Error()})
		}

		publishEvent(s, copilot.SSEEvent{Type: "done", Metadata: &copilot.EventMeta{DurationMs: time.Since(startTime).Milliseconds()}})
		c.Flush()
		log.Printf("[Copilot] Diagnose finished - session: %s, duration: %dms", req.SessionID, time.Since(startTime).Milliseconds())
	}
}

// CopilotPlanFromHandler SSE 流式接口 POST /api/copilot/plan-from
func CopilotPlanFromHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req copilot.PlanFromRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "invalid request body: " + err.Error(),
			})
			return
		}

		if req.SessionID == "" || req.Diagnosis == "" {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "session_id and diagnosis are required",
			})
			return
		}

		log.Printf("[Copilot] PlanFrom request - session: %s", req.SessionID)

		cfg := copilot.DefaultConfig()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TotalTimeoutSec)*time.Second)
		defer cancel()

		progressCh := make(chan copilot.ProgressEvent, 100)
		s := sse.NewStream(c)
		startTime := time.Now()

		// Run plan in background goroutine, all events go through channel
		errCh := make(chan error, 1)
		go func() {
			errCh <- copilot.RunPlanFromDiagnosis(timeoutCtx, &req, progressCh)
			close(progressCh)
		}()

		// Consume events and publish SSE (single writer)
		for pe := range progressCh {
			switch pe.Type {
			case "progress":
				publishEvent(s, copilot.SSEEvent{Type: "progress", Node: pe.Node, Status: pe.Status})
			case "content":
				publishEvent(s, copilot.SSEEvent{Type: "content", Content: pe.Data})
			}
		}

		// Check for error
		if err := <-errCh; err != nil {
			log.Printf("[Copilot] PlanFrom error: %v", err)
			publishEvent(s, copilot.SSEEvent{Type: "error", Content: err.Error()})
		}

		publishEvent(s, copilot.SSEEvent{Type: "done", Metadata: &copilot.EventMeta{DurationMs: time.Since(startTime).Milliseconds()}})
		c.Flush()
		log.Printf("[Copilot] PlanFrom finished - session: %s, duration: %dms", req.SessionID, time.Since(startTime).Milliseconds())
	}
}

// CopilotPlanHandler 同步接口 POST /api/copilot/plan
func CopilotPlanHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var req copilot.CopilotRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "invalid request body: " + err.Error(),
			})
			return
		}

		if req.SessionID == "" || req.Message == "" {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "session_id and message are required",
			})
			return
		}

		log.Printf("[Copilot] Plan request - session: %s, message: %s", req.SessionID, req.Message)

		cfg := copilot.DefaultConfig()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TotalTimeoutSec)*time.Second)
		defer cancel()

		startTime := time.Now()
		plan, err := copilot.RunCopilotSync(timeoutCtx, req.SessionID, req.Message)
		if err != nil {
			log.Printf("[Copilot] Plan error: %v", err)
			c.JSON(consts.StatusInternalServerError, copilot.CopilotResponse{
				Code:    500,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(consts.StatusOK, copilot.CopilotResponse{
			Code:    0,
			Message: "success",
			Data: &copilot.CopilotResult{
				SessionID: req.SessionID,
				Plan:      plan,
				Metadata: copilot.ResultMetadata{
					DurationMs: time.Since(startTime).Milliseconds(),
					GraphNodes: 6,
				},
			},
		})
	}
}

// --- helpers ---

func publishEvent(s *sse.Stream, event copilot.SSEEvent) {
	data, _ := json.Marshal(event)
	_ = s.Publish(&sse.Event{Data: data})
}

func drainProgressChan(s *sse.Stream, ch <-chan copilot.ProgressEvent) {
	for {
		select {
		case pe, ok := <-ch:
			if !ok {
				return
			}
			event := copilot.SSEEvent{
				Type:   pe.Type,
				Node:   pe.Node,
				Status: pe.Status,
				Tool:   pe.Tool,
				Params: pe.Params,
				Result: pe.Result,
			}
			if pe.Type == "diagnosis" || pe.Type == "content" {
				event.Content = pe.Data
			}
			publishEvent(s, event)
		default:
			return
		}
	}
}
