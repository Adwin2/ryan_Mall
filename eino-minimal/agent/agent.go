package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"eino-minimal/memory"
	"eino-minimal/models"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

var defaultMemory = memory.GetDefaultMemory()

// UserMessage 用户消息结构
type UserMessage struct {
	Message string            `json:"message"`
	History []*schema.Message `json:"history"`
}

// 节点名称常量
const (
	InputToQuery  = "InputToQuery"
	ChatTemplate  = "ChatTemplate"
	ReactModel    = "ReactModel"
)

// BuildAgent 构建代理
func BuildAgent(ctx context.Context) (compose.Runnable[*UserMessage, *schema.Message], error) {
	g := compose.NewGraph[*UserMessage, *schema.Message]()
	
	// 创建聊天模板
	chatTemplate, err := NewChatTemplate(ctx)
	if err != nil {
		log.Printf("创建聊天模板失败: %v", err)
		return nil, err
	}
	
	// 创建React代理Lambda
	reactModeLambda, err := NewReactAgentLambda(ctx)
	if err != nil {
		log.Printf("创建React代理失败: %v", err)
		return nil, err
	}
	
	// 添加节点
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplate)
	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(inputToQueryLambda), compose.WithNodeName("UserMessageToQuery"))
	_ = g.AddLambdaNode(ReactModel, reactModeLambda, compose.WithNodeName("UserMessageToReactModel"))
	
	// 添加边连接
	_ = g.AddEdge(compose.START, InputToQuery)
	_ = g.AddEdge(InputToQuery, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReactModel)
	_ = g.AddEdge(ReactModel, compose.END)
	
	// 编译图
	r, err := g.Compile(ctx, compose.WithGraphName("EinoMinimal"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		log.Printf("编译代理图失败: %v", err)
		return nil, err
	}
	
	return r, nil
}

// NewReactAgentLambda 创建React代理Lambda
func NewReactAgentLambda(ctx context.Context) (*compose.Lambda, error) {
	r, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: models.GetDefaultChatModel(ctx),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: GetTools(), // 可以为空，表示不使用工具
		},
	})
	if err != nil {
		log.Printf("React代理创建错误: %v", err)
		return nil, err
	}
	
	lba, err := compose.AnyLambda(r.Generate, r.Stream, nil, nil)
	if err != nil {
		log.Printf("React lambda错误: %v", err)
		return nil, err
	}
	
	return lba, nil
}

// inputToQueryLambda 输入转换Lambda
func inputToQueryLambda(ctx context.Context, input *UserMessage, opts ...any) (map[string]any, error) {
	return map[string]any{
		"message": input.Message,
		"history": input.History,
		"date": time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// RunAgent 运行代理
func RunAgent(ctx context.Context, id string, msg string) (*schema.StreamReader[*schema.Message], error) {
	runner, err := BuildAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent graph: %w", err)
	}

	conversation := defaultMemory.GetConversation(id, true)

	userMessage := &UserMessage{
		Message: msg,
		History: conversation.GetMessages(),
	}

	sr, err := runner.Stream(ctx, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	srs := sr.Copy(2)

	go func() {
		// 用于保存到内存
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			// 关闭流
			srs[1].Close()

			// 添加用户输入到历史
			conversation.Append(schema.UserMessage(msg))

			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				log.Printf("连接消息错误: %v", err)
			}
			// 添加代理响应到历史
			conversation.Append(fullMsg)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				log.Printf("上下文完成: %v", ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break outer
					}
				}
				// log.Printf("收到消息: %v", chunk.Content)

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	return srs[0], nil
}
