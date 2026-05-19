// SSE 事件类型
export type SSEEventType = 'progress' | 'tool_call' | 'tool_result' | 'diagnosis' | 'content' | 'error' | 'done' | 'routing' | 'context' | 'thinking'

export interface SSEEvent {
  type: SSEEventType
  content?: string
  node?: string
  status?: 'start' | 'done'
  tool?: string
  params?: Record<string, unknown>
  result?: string
  metadata?: EventMeta
  // multi-agent routing fields
  tier?: string
  profile?: string
  model?: string
  reason?: string
}

export interface EventMeta {
  duration_ms: number
  tools_called?: string[]
  model?: string
}

// Copilot 状态机
export type Phase = 'idle' | 'diagnosing' | 'awaiting_edit' | 'planning' | 'done' | 'error'

// 消息
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: number
  type?: 'diagnosis' | 'plan' | 'text'
}

// 进度节点
export interface NodeProgress {
  node: string
  status: 'pending' | 'running' | 'done'
  startTime?: number
  endTime?: number
}

// Tool call
export interface ToolCall {
  tool: string
  params: Record<string, unknown>
  result?: string
  timestamp: number
}

// 诊断结果
export interface DiagnoseResponse {
  code: number
  message: string
  session_id: string
  diagnosis: string
  full_analysis: string
}

// Session
export interface Session {
  id: string
  title: string
  createdAt: number
  lastMessage?: string
}

// DAG 节点定义（固定 6 节点）
export const DAG_NODES = [
  { key: 'InputParser', label: '数据预取' },
  { key: 'OrchestratorTemplate', label: '提示注入' },
  { key: 'OrchestratorLLM', label: '诊断分析' },
  { key: 'DiagnosisExtractor', label: '诊断提取' },
  { key: 'PlannerTemplate', label: '方案提示' },
  { key: 'PlannerAgent', label: '方案生成' },
] as const

// Multi-Agent Playground types
export type AgentType = 'ops' | 'consumer'

export interface RoutingDecision {
  tier: string
  profile: string
  model: string
  reason: string
  timestamp: number
}

export interface ContextInfo {
  history_turns: number
  max_turns: number
  summary_injected: boolean
  recall_count: number
  token_estimate: number
  token_budget: number
}

export interface ProfileMetrics {
  CallCount: number
  SuccessCount: number
  FailCount: number
  TotalLatencyMs: number
  TotalTokens: number
  TierDist: Record<string, number>
}

export interface ThinkingStep {
  content: string
  timestamp: number
}
