import { useCallback, useEffect, useRef, useState } from 'react'
import { useSSE } from './useSSE'
import type { AgentType, ChatMessage, ContextInfo, ProfileMetrics, RoutingDecision, SSEEvent, ThinkingStep, ToolCall } from '../types'

interface UseMultiAgentReturn {
  messages: ChatMessage[]
  streamContent: string
  isStreaming: boolean
  routingDecision: RoutingDecision | null
  contextInfo: ContextInfo | null
  metrics: Record<string, ProfileMetrics> | null
  toolCalls: ToolCall[]
  thinkingSteps: ThinkingStep[]
  sendMessage: (message: string) => void
  cancel: () => void
  agent: AgentType
  setAgent: (agent: AgentType) => void
  sessionId: string
}

export function useMultiAgent(): UseMultiAgentReturn {
  const [agent, setAgent] = useState<AgentType>('ops')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [streamContent, setStreamContent] = useState('')
  const [routingDecision, setRoutingDecision] = useState<RoutingDecision | null>(null)
  const [contextInfo, setContextInfo] = useState<ContextInfo | null>(null)
  const [metrics, setMetrics] = useState<Record<string, ProfileMetrics> | null>(null)
  const [toolCalls, setToolCalls] = useState<ToolCall[]>([])
  const [thinkingSteps, setThinkingSteps] = useState<ThinkingStep[]>([])
  const [isStreaming, setIsStreaming] = useState(false)
  const sessionIdRef = useRef(`playground_${Date.now().toString(36)}`)

  const onEvent = useCallback((event: SSEEvent) => {
    switch (event.type) {
      case 'routing':
        setRoutingDecision({
          tier: event.tier || '',
          profile: event.profile || '',
          model: event.model || '',
          reason: event.reason || '',
          timestamp: Date.now(),
        })
        break
      case 'context':
        if (event.content) {
          try {
            setContextInfo(JSON.parse(event.content))
          } catch { /* ignore */ }
        }
        break
      case 'content':
        if (event.content) {
          setStreamContent(prev => prev + event.content)
        }
        break
      case 'thinking':
        if (event.content) {
          setThinkingSteps(prev => [...prev, {
            content: event.content!,
            timestamp: Date.now(),
          }])
        }
        break
      case 'tool_call':
        if (event.tool) {
          let params: Record<string, unknown> = {}
          if (event.content) {
            try { params = JSON.parse(event.content) } catch { /* ignore */ }
          }
          setToolCalls(prev => [...prev, {
            tool: event.tool!,
            params,
            timestamp: Date.now(),
          }])
        }
        break
      case 'tool_result':
        if (event.tool && event.result) {
          setToolCalls(prev => {
            const updated = [...prev]
            const last = updated.findLast(tc => !tc.result)
            if (last) last.result = event.result
            return updated
          })
        }
        break
      case 'error':
        setIsStreaming(false)
        break
    }
  }, [])

  const onDone = useCallback(() => {
    setIsStreaming(false)
    setMessages(prev => {
      const content = streamContentRef.current
      if (!content) return prev
      return [...prev, {
        id: `assistant_${Date.now()}`,
        role: 'assistant' as const,
        content,
        timestamp: Date.now(),
      }]
    })
    setStreamContent('')
  }, [])

  const streamContentRef = useRef('')
  useEffect(() => { streamContentRef.current = streamContent }, [streamContent])

  const { connect, close } = useSSE({ onEvent, onDone })

  const sendMessage = useCallback((message: string) => {
    setMessages(prev => [...prev, {
      id: `user_${Date.now()}`,
      role: 'user',
      content: message,
      timestamp: Date.now(),
    }])
    setStreamContent('')
    setToolCalls([])
    setThinkingSteps([])
    setIsStreaming(true)

    const endpoint = agent === 'ops' ? '/api/ops/stream' : '/api/shop/stream'
    const url = `${endpoint}?session_id=${encodeURIComponent(sessionIdRef.current)}&message=${encodeURIComponent(message)}`
    connect(url)
  }, [agent, connect])

  const cancel = useCallback(() => {
    close()
    setIsStreaming(false)
  }, [close])

  useEffect(() => {
    const fetchMetrics = () => {
      fetch('/api/routing/metrics')
        .then(r => r.json())
        .then(setMetrics)
        .catch(() => {})
    }
    fetchMetrics()
    const interval = setInterval(fetchMetrics, 5000)
    return () => clearInterval(interval)
  }, [])

  return {
    messages,
    streamContent,
    isStreaming,
    routingDecision,
    contextInfo,
    metrics,
    toolCalls,
    thinkingSteps,
    sendMessage,
    cancel,
    agent,
    setAgent,
    sessionId: sessionIdRef.current,
  }
}
