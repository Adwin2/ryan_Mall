package copilot

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

var copilotMemory = memory.NewSimpleMemory(memory.SimpleMemoryConfig{
	Dir:           DefaultConfig().MemoryDir,
	MaxWindowSize: DefaultConfig().MemoryWindowSize,
})

// RunCopilotStreamWithProgress 以流式方式运行 Copilot，同时推送 progress 事件
func RunCopilotStreamWithProgress(ctx context.Context, sessionID, message string, progressCh chan<- ProgressEvent) (*schema.StreamReader[*schema.Message], error) {
	// 发送 InputParser start
	progressCh <- ProgressEvent{Type: "progress", Node: NodeInputParser, Status: "start"}

	runner, err := BuildCopilotGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("构建 Copilot Graph 失败: %w", err)
	}

	convID := "copilot_" + sessionID
	conversation := copilotMemory.GetConversation(convID, true)

	input := &CopilotInput{
		Goal:      message,
		SessionID: sessionID,
		History:   conversation.GetMessages(),
	}

	// InputParser done (数据预取已在 BuildCopilotGraph 前完成)
	progressCh <- ProgressEvent{Type: "progress", Node: NodeInputParser, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorTemplate, Status: "start"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorTemplate, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorLLM, Status: "start"}

	sr, err := runner.Stream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("Copilot Stream 执行失败: %w", err)
	}

	// 复制流：一份给 HTTP handler，一份用于持久化 memory
	srs := sr.Copy(2)

	go func() {
		fullMsgs := make([]*schema.Message, 0)
		defer func() {
			srs[1].Close()

			// 保存用户输入
			conversation.Append(schema.UserMessage(message))

			// 合并并保存 Agent 回复
			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				log.Printf("[Copilot] 合并消息失败: %v", err)
				return
			}
			conversation.Append(fullMsg)
		}()

		for {
			chunk, err := srs[1].Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Printf("[Copilot] 流读取错误: %v", err)
				return
			}
			fullMsgs = append(fullMsgs, chunk)
		}
	}()

	return srs[0], nil
}

// RunCopilotStream 以流式方式运行 Copilot (兼容旧接口)
func RunCopilotStream(ctx context.Context, sessionID, message string) (*schema.StreamReader[*schema.Message], error) {
	progressCh := make(chan ProgressEvent, 50)
	go func() {
		for range progressCh {
			// discard
		}
	}()
	sr, err := RunCopilotStreamWithProgress(ctx, sessionID, message, progressCh)
	close(progressCh)
	return sr, err
}

// RunCopilotSync 同步运行 Copilot，返回完整结果
func RunCopilotSync(ctx context.Context, sessionID, message string) (string, error) {
	runner, err := BuildCopilotGraph(ctx)
	if err != nil {
		return "", fmt.Errorf("构建 Copilot Graph 失败: %w", err)
	}

	convID := "copilot_" + sessionID
	conversation := copilotMemory.GetConversation(convID, true)

	input := &CopilotInput{
		Goal:      message,
		SessionID: sessionID,
		History:   conversation.GetMessages(),
	}

	result, err := runner.Invoke(ctx, input)
	if err != nil {
		return "", fmt.Errorf("Copilot Invoke 执行失败: %w", err)
	}

	// 保存到 memory
	conversation.Append(schema.UserMessage(message))
	conversation.Append(result)

	return result.Content, nil
}

// RunDiagnoseOnly 只跑到诊断结束（前 4 个节点），返回诊断结果（非流式，保留兼容）
func RunDiagnoseOnly(ctx context.Context, sessionID, message string) (*DiagnoseResponse, error) {
	start := time.Now()

	input := &CopilotInput{
		Goal:      message,
		SessionID: sessionID,
	}

	templateVars, err := parseInput(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("InputParser 失败: %w", err)
	}

	orchestratorTpl, err := NewOrchestratorTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("OrchestratorTemplate 创建失败: %w", err)
	}

	messages, err := orchestratorTpl.Format(ctx, templateVars)
	if err != nil {
		return nil, fmt.Errorf("OrchestratorTemplate 格式化失败: %w", err)
	}

	chatModel := models.GetDefaultChatModel(ctx)
	if chatModel == nil {
		return nil, fmt.Errorf("无法获取 ChatModel")
	}

	result, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("Orchestrator LLM 生成失败: %w", err)
	}

	diagVars, err := extractDiagnosis(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("DiagnosisExtractor 失败: %w", err)
	}

	log.Printf("[Copilot] DiagnoseOnly 完成, 耗时: %dms", time.Since(start).Milliseconds())

	diagnosis, _ := diagVars["diagnosis"].(string)
	fullAnalysis, _ := diagVars["full_analysis"].(string)

	convID := "copilot_" + sessionID
	conversation := copilotMemory.GetConversation(convID, true)
	conversation.Append(schema.UserMessage(message))
	conversation.Append(result)

	return &DiagnoseResponse{
		Code:      0,
		Message:   "success",
		SessionID: sessionID,
		Diagnosis: diagnosis,
		Analysis:  fullAnalysis,
	}, nil
}

// RunDiagnoseStream 流式运行诊断（前 4 个节点），通过 progressCh 发送 progress 和 content
// 结束时发送 diagnosis 事件，调用者通过 channel 接收所有事件
func RunDiagnoseStream(ctx context.Context, sessionID, message string, progressCh chan<- ProgressEvent) error {
	progressCh <- ProgressEvent{Type: "progress", Node: NodeInputParser, Status: "start"}

	input := &CopilotInput{
		Goal:      message,
		SessionID: sessionID,
	}

	templateVars, err := parseInput(ctx, input)
	if err != nil {
		return fmt.Errorf("InputParser 失败: %w", err)
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodeInputParser, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorTemplate, Status: "start"}

	orchestratorTpl, err := NewOrchestratorTemplate(ctx)
	if err != nil {
		return fmt.Errorf("OrchestratorTemplate 创建失败: %w", err)
	}

	msgs, err := orchestratorTpl.Format(ctx, templateVars)
	if err != nil {
		return fmt.Errorf("OrchestratorTemplate 格式化失败: %w", err)
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorTemplate, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorLLM, Status: "start"}

	chatModel := models.GetDefaultChatModel(ctx)
	if chatModel == nil {
		return fmt.Errorf("无法获取 ChatModel")
	}

	// 流式调用 Orchestrator LLM
	sr, err := chatModel.Stream(ctx, msgs)
	if err != nil {
		return fmt.Errorf("Orchestrator LLM Stream 失败: %w", err)
	}

	var fullContent string
	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("Orchestrator LLM recv 失败: %w", err)
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			progressCh <- ProgressEvent{Type: "content", Data: chunk.Content}
		}
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodeOrchestratorLLM, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodeDiagnosisExtractor, Status: "start"}

	// 提取诊断
	fullMsg := &schema.Message{Role: schema.Assistant, Content: fullContent}
	diagVars, err := extractDiagnosis(ctx, fullMsg)
	if err != nil {
		return fmt.Errorf("DiagnosisExtractor 失败: %w", err)
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodeDiagnosisExtractor, Status: "done"}

	diagnosis, _ := diagVars["diagnosis"].(string)
	fullAnalysis, _ := diagVars["full_analysis"].(string)

	// 发送 diagnosis 事件
	progressCh <- ProgressEvent{Type: "diagnosis", Data: diagnosis, Result: fullAnalysis}

	// 保存到 memory
	convID := "copilot_" + sessionID
	conversation := copilotMemory.GetConversation(convID, true)
	conversation.Append(schema.UserMessage(message))
	conversation.Append(fullMsg)

	return nil
}

// RunPlanFromDiagnosis 从已有诊断出发，只跑 Planner
// 所有事件通过 progressCh 发送（progress + content + done），调用者只需 range channel
func RunPlanFromDiagnosis(ctx context.Context, req *PlanFromRequest, progressCh chan<- ProgressEvent) error {
	progressCh <- ProgressEvent{Type: "progress", Node: NodePlannerTemplate, Status: "start"}

	// 构建 Planner prompt
	plannerTpl, err := NewPlannerTemplate(ctx)
	if err != nil {
		return fmt.Errorf("PlannerTemplate 创建失败: %w", err)
	}

	plannerVars := map[string]any{
		"diagnosis":     req.Diagnosis,
		"full_analysis": req.Analysis,
	}

	messages, err := plannerTpl.Format(ctx, plannerVars)
	if err != nil {
		return fmt.Errorf("PlannerTemplate 格式化失败: %w", err)
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodePlannerTemplate, Status: "done"}
	progressCh <- ProgressEvent{Type: "progress", Node: NodePlannerAgent, Status: "start"}

	// 直接构建 react.Agent 用于流式执行
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: models.GetDefaultChatModel(ctx),
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: GetPlannerTools(),
		},
	})
	if err != nil {
		return fmt.Errorf("PlannerAgent 创建失败: %w", err)
	}

	sr, err := agent.Stream(ctx, messages)
	if err != nil {
		return fmt.Errorf("PlannerAgent Stream 失败: %w", err)
	}

	// 消费流，把 content 发到 channel
	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("PlannerAgent recv 失败: %w", err)
		}
		if chunk.Content != "" {
			progressCh <- ProgressEvent{Type: "content", Data: chunk.Content}
		}
	}

	progressCh <- ProgressEvent{Type: "progress", Node: NodePlannerAgent, Status: "done"}
	return nil
}
