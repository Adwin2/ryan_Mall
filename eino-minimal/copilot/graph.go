package copilot

import (
	"context"
	"log"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// Graph 节点名称
const (
	NodeInputParser          = "InputParser"
	NodeOrchestratorTemplate = "OrchestratorTemplate"
	NodeOrchestratorLLM      = "OrchestratorLLM"
	NodeDiagnosisExtractor   = "DiagnosisExtractor"
	NodePlannerTemplate      = "PlannerTemplate"
	NodePlannerAgent         = "PlannerAgent"
)

// BuildCopilotGraph 构建 Copilot 的 6 节点线性 DAG
//
// 优化后的拓扑（Parallelization 模式）:
//
//	InputParser: 并行预取 3 个数据源（campaign_history, metrics, segments）+ 解析类目
//	OrchestratorTemplate: 将预取数据注入 prompt
//	OrchestratorLLM: 单轮 LLM 合成诊断报告（不再 tool call）
//	DiagnosisExtractor: 提取 JSON 诊断
//	PlannerTemplate: 注入诊断到 Planner prompt
//	PlannerAgent: ReAct Agent 调真实 API 选品，流式输出方案
//
// 对比优化前: Orchestrator 从 4+ 轮 LLM (ReAct tool calling) → 1 轮 LLM
func BuildCopilotGraph(ctx context.Context) (compose.Runnable[*CopilotInput, *schema.Message], error) {
	g := compose.NewGraph[*CopilotInput, *schema.Message]()

	// Node 1: InputParser — 解析目标 + 并行预取数据
	_ = g.AddLambdaNode(NodeInputParser,
		compose.InvokableLambdaWithOption(parseInput),
		compose.WithNodeName("ParseAndPrefetch"))

	// Node 2: OrchestratorTemplate — 将预取数据 + 目标注入 prompt
	orchestratorTpl, err := NewOrchestratorTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(NodeOrchestratorTemplate, orchestratorTpl)

	// Node 3: OrchestratorLLM — 单轮 LLM 合成诊断（无 tool calling）
	orchestratorLambda, err := NewOrchestratorAgentLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(NodeOrchestratorLLM, orchestratorLambda,
		compose.WithNodeName("OrchestratorSynthesis"))

	// Node 4: DiagnosisExtractor — 提取 JSON 诊断报告
	_ = g.AddLambdaNode(NodeDiagnosisExtractor,
		compose.InvokableLambdaWithOption(extractDiagnosis),
		compose.WithNodeName("ExtractDiagnosis"))

	// Node 5: PlannerTemplate — 注入诊断结果
	plannerTpl, err := NewPlannerTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(NodePlannerTemplate, plannerTpl)

	// Node 6: PlannerAgent — ReAct Agent (Real API tools, 流式输出)
	plannerLambda, err := NewPlannerAgentLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(NodePlannerAgent, plannerLambda,
		compose.WithNodeName("PlannerReActAgent"))

	// Edge wiring: 线性 DAG
	_ = g.AddEdge(compose.START, NodeInputParser)
	_ = g.AddEdge(NodeInputParser, NodeOrchestratorTemplate)
	_ = g.AddEdge(NodeOrchestratorTemplate, NodeOrchestratorLLM)
	_ = g.AddEdge(NodeOrchestratorLLM, NodeDiagnosisExtractor)
	_ = g.AddEdge(NodeDiagnosisExtractor, NodePlannerTemplate)
	_ = g.AddEdge(NodePlannerTemplate, NodePlannerAgent)
	_ = g.AddEdge(NodePlannerAgent, compose.END)

	// 编译
	r, err := g.Compile(ctx,
		compose.WithGraphName("CopilotDAG"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		log.Printf("Copilot Graph 编译失败: %v", err)
		return nil, err
	}

	return r, nil
}
