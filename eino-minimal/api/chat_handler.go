package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"eino-minimal/agent"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/sse"
)

// ChatHandler 聊天处理器
func ChatHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		id := c.Query("id")
		message := c.Query("message")
		
		if id == "" || message == "" {
			c.JSON(consts.StatusBadRequest, map[string]string{
				"status": "error",
				"error":  "missing id or message parameter",
			})
			return
		}

		log.Printf("[Chat] Starting chat with ID: %s, Message: %s\n", id, message)

		sr, err := agent.RunAgent(ctx, id, message)
		if err != nil {
			log.Printf("[Chat] Error running agent: %v\n", err)
			c.JSON(consts.StatusInternalServerError, map[string]string{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		s := sse.NewStream(c)
		defer func() {
			sr.Close()
			c.Flush()
			log.Printf("[Chat] Finished chat with ID: %s\n", id)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				log.Printf("[Chat] Context done for chat ID: %s\n", id)
				return
			default:
				msg, err := sr.Recv()
				if errors.Is(err, io.EOF) {
					log.Printf("[Chat] EOF received for chat ID: %s\n", id)
					break outer
				}
				if err != nil {
					log.Printf("[Chat] Error receiving message: %v\n", err)
					break outer
				}

				err = s.Publish(&sse.Event{
					Data: []byte(msg.Content),
				})
				fmt.Print(msg.Content)
				if err != nil {
					log.Printf("[Chat] Error publishing message: %v\n", err)
					break outer
				}
			}
		}
	}
}

// SetupRoutes 设置路由
func SetupRoutes(h *app.RequestContext) {
	// 这个函数可以用于设置路由，但在这个简化版本中我们在main.go中直接设置
}
