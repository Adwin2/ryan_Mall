package copilot

import (
	"context"
	"fmt"
	"log"

	"eino-minimal/models"

	"github.com/cloudwego/eino/compose"
)

// NewOrchestratorAgentLambda 创建 Orchestrator 的 LLM Lambda
// 优化后：数据已在 InputParser 预取，Orchestrator 只做单轮 LLM 合成
// 直接使用 ChatModel，不经过 react.Agent（无需 tool calling）
func NewOrchestratorAgentLambda(ctx context.Context) (*compose.Lambda, error) {
	chatModel := models.GetDefaultChatModel(ctx)
	if chatModel == nil {
		log.Printf("Orchestrator: 无法获取 ChatModel")
		return nil, fmt.Errorf("failed to get chat model")
	}

	lambda, err := compose.AnyLambda(chatModel.Generate, chatModel.Stream, nil, nil)
	if err != nil {
		log.Printf("Orchestrator Lambda 创建失败: %v", err)
		return nil, err
	}

	return lambda, nil
}
