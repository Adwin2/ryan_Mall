package copilot

import (
	"context"
	"log"

	copilottools "eino-minimal/tools/copilot"
	"eino-minimal/models"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// NewPlannerAgentLambda 创建 Planner 的 react.Agent Lambda
func NewPlannerAgentLambda(ctx context.Context) (*compose.Lambda, error) {
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: models.GetDefaultChatModel(ctx),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: GetPlannerTools(),
		},
	})
	if err != nil {
		log.Printf("Planner Agent 创建失败: %v", err)
		return nil, err
	}

	lambda, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		log.Printf("Planner Lambda 创建失败: %v", err)
		return nil, err
	}

	return lambda, nil
}

// GetPlannerTools 返回 Planner 的工具集（Real API 工具）
func GetPlannerTools() []tool.BaseTool {
	return []tool.BaseTool{
		copilottools.SearchProductsTool(),
		copilottools.CategoryProductsTool(),
	}
}
