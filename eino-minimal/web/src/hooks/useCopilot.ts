import { useCallback, useRef, useState } from 'react'
import { useSSE } from './useSSE'
import type { ChatMessage, DiagnoseResponse, NodeProgress, Phase, Session, SSEEvent, ToolCall } from '../types'

function genId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

export function useCopilot() {
  const [phase, setPhase] = useState<Phase>('idle')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [diagnosis, setDiagnosis] = useState<DiagnoseResponse | null>(null)
  const [progress, setProgress] = useState<NodeProgress[]>([])
  const [toolCalls, setToolCalls] = useState<ToolCall[]>([])
  const [streamContent, setStreamContent] = useState('')
  const [sessions, setSessions] = useState<Session[]>(() => {
    const id = genId()
    return [{ id, title: 'New Session', createdAt: Date.now() }]
  })
  const [activeSessionId, setActiveSessionId] = useState(() => sessions[0]?.id || genId())
  const contentRef = useRef('')
  const abortRef = useRef<AbortController | null>(null)
  const doneHandledRef = useRef(false)

  // Session message store (in-memory)
  const sessionStoreRef = useRef<Record<string, { messages: ChatMessage[]; diagnosis: DiagnoseResponse | null }>>({})

  const sessionIdRef = useRef(activeSessionId)
  sessionIdRef.current = activeSessionId

  const resetProgress = useCallback(() => {
    setProgress([])
    setToolCalls([])
    setStreamContent('')
    contentRef.current = ''
    doneHandledRef.current = false
  }, [])

  const handleSSEEvent = useCallback((event: SSEEvent) => {
    switch (event.type) {
      case 'progress':
        setProgress(prev => {
          const existing = prev.find(p => p.node === event.node)
          if (existing) {
            return prev.map(p =>
              p.node === event.node
                ? { ...p, status: event.status === 'done' ? 'done' : 'running', endTime: event.status === 'done' ? Date.now() : undefined }
                : p
            )
          }
          return [...prev, { node: event.node!, status: 'running', startTime: Date.now() }]
        })
        break

      case 'tool_call':
        setToolCalls(prev => [...prev, {
          tool: event.tool || '',
          params: event.params || {},
          result: event.result,
          timestamp: Date.now(),
        }])
        break

      case 'content':
        contentRef.current += event.content || ''
        setStreamContent(contentRef.current)
        break

      case 'diagnosis': {
        // Diagnosis complete — save and transition
        const diagResp: DiagnoseResponse = {
          code: 0,
          message: 'success',
          session_id: sessionIdRef.current,
          diagnosis: event.content || '',
          full_analysis: event.result || '',
        }
        setDiagnosis(diagResp)
        // Add the full analysis as an assistant message
        if (contentRef.current) {
          setMessages(prev => [...prev, {
            id: genId(),
            role: 'assistant',
            content: contentRef.current,
            timestamp: Date.now(),
            type: 'diagnosis',
          }])
          setStreamContent('')
          contentRef.current = ''
        }
        setPhase('awaiting_edit')
        doneHandledRef.current = true
        break
      }

      case 'error':
        setPhase('error')
        setMessages(prev => [...prev, {
          id: genId(),
          role: 'assistant',
          content: `Error: ${event.content}`,
          timestamp: Date.now(),
          type: 'text',
        }])
        doneHandledRef.current = true
        break

      case 'done': {
        // Only handle if not already handled by diagnosis/error event
        if (doneHandledRef.current) break
        doneHandledRef.current = true
        const finalContent = contentRef.current
        if (finalContent) {
          setMessages(prev => [...prev, {
            id: genId(),
            role: 'assistant',
            content: finalContent,
            timestamp: Date.now(),
            type: 'plan',
          }])
        }
        setStreamContent('')
        contentRef.current = ''
        setPhase('done')
        break
      }
    }
  }, [])

  const sse = useSSE({
    onEvent: handleSSEEvent,
  })

  // Helper: consume SSE from a fetch POST response
  const consumeFetchSSE = useCallback(async (resp: Response) => {
    const reader = resp.body?.getReader()
    if (!reader) {
      handleSSEEvent({ type: 'error', content: 'No response body' })
      return
    }
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          // Process remaining buffer
          if (buffer.trim()) {
            parseSseLines(buffer, handleSSEEvent)
          }
          // Fallback: ensure done is triggered if not already
          if (!doneHandledRef.current) {
            handleSSEEvent({ type: 'done' })
          }
          break
        }
        buffer += decoder.decode(value, { stream: true })

        // Parse complete SSE events (separated by double newline)
        const parts = buffer.split('\n\n')
        buffer = parts.pop() || ''

        for (const part of parts) {
          parseSseLines(part, handleSSEEvent)
        }
      }
    } catch (err) {
      // Stream read error — save any accumulated content before re-throwing
      if (!doneHandledRef.current && contentRef.current) {
        handleSSEEvent({ type: 'done' })
      }
      throw err
    }
  }, [handleSSEEvent])

  // 全流程 SSE (GET)
  const startFullRun = useCallback((message: string) => {
    resetProgress()
    setPhase('diagnosing')
    setMessages(prev => [...prev, {
      id: genId(),
      role: 'user',
      content: message,
      timestamp: Date.now(),
    }])
    const url = `/api/copilot/stream?session_id=${sessionIdRef.current}&message=${encodeURIComponent(message)}`
    sse.connect(url)
  }, [sse, resetProgress])

  // 分步: 流式诊断 (POST SSE)
  const startDiagnose = useCallback(async (message: string) => {
    resetProgress()
    setPhase('diagnosing')
    setMessages(prev => [...prev, {
      id: genId(),
      role: 'user',
      content: message,
      timestamp: Date.now(),
    }])

    // Update session title
    setSessions(prev => prev.map(s =>
      s.id === sessionIdRef.current
        ? { ...s, title: message.slice(0, 30), lastMessage: message }
        : s
    ))

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const resp = await fetch('/api/copilot/diagnose', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: sessionIdRef.current, message }),
        signal: controller.signal,
      })

      if (!resp.ok) {
        const errText = await resp.text()
        handleSSEEvent({ type: 'error', content: `HTTP ${resp.status}: ${errText}` })
        return
      }

      await consumeFetchSSE(resp)
    } catch (err) {
      if ((err as Error).name === 'AbortError') return
      const lostContent = contentRef.current
      if (lostContent) {
        setMessages(prev => [...prev, {
          id: genId(),
          role: 'assistant',
          content: lostContent,
          timestamp: Date.now(),
          type: 'text',
        }])
        setStreamContent('')
        contentRef.current = ''
      }
      setPhase('error')
      setMessages(prev => [...prev, {
        id: genId(),
        role: 'assistant',
        content: `Network error: ${err}`,
        timestamp: Date.now(),
        type: 'text',
      }])
    }
  }, [resetProgress, handleSSEEvent, consumeFetchSSE])

  // 分步: 从诊断生成方案 (POST SSE)
  const startPlanFrom = useCallback((editedDiagnosis?: string) => {
    if (!diagnosis) return
    setPhase('planning')
    contentRef.current = ''
    setStreamContent('')
    doneHandledRef.current = false

    // Add planner progress nodes
    setProgress(prev => [
      ...prev,
      { node: 'PlannerTemplate', status: 'running', startTime: Date.now() },
    ])

    const body = {
      session_id: sessionIdRef.current,
      diagnosis: editedDiagnosis || diagnosis.diagnosis,
      full_analysis: diagnosis.full_analysis,
    }

    const controller = new AbortController()
    abortRef.current = controller

    fetch('/api/copilot/plan-from', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: controller.signal,
    }).then(async (resp) => {
      if (!resp.ok) {
        const errText = await resp.text()
        handleSSEEvent({ type: 'error', content: `HTTP ${resp.status}: ${errText}` })
        return
      }
      await consumeFetchSSE(resp)
    }).catch((err) => {
      if (err.name === 'AbortError') return
      if (contentRef.current) {
        setMessages(prev => [...prev, {
          id: genId(),
          role: 'assistant',
          content: contentRef.current,
          timestamp: Date.now(),
          type: 'plan',
        }])
        setStreamContent('')
        contentRef.current = ''
      }
      setPhase('error')
    })
  }, [diagnosis, handleSSEEvent, consumeFetchSSE])

  // 取消
  const cancel = useCallback(() => {
    sse.close()
    abortRef.current?.abort()
    doneHandledRef.current = true
    if (contentRef.current) {
      setMessages(prev => [...prev, {
        id: genId(),
        role: 'assistant',
        content: contentRef.current + '\n\n_(Cancelled)_',
        timestamp: Date.now(),
        type: 'plan',
      }])
      setStreamContent('')
      contentRef.current = ''
    }
    setPhase('done')
  }, [sse])

  // 新建 session (保存当前 session)
  const newSession = useCallback(() => {
    // Save current session state
    sessionStoreRef.current[sessionIdRef.current] = {
      messages: [...messages],
      diagnosis,
    }

    const newId = genId()
    setSessions(prev => [{ id: newId, title: 'New Session', createdAt: Date.now() }, ...prev])
    setActiveSessionId(newId)
    setMessages([])
    setDiagnosis(null)
    setProgress([])
    setToolCalls([])
    setStreamContent('')
    contentRef.current = ''
    doneHandledRef.current = false
    setPhase('idle')
  }, [messages, diagnosis])

  // 切换 session
  const switchSession = useCallback((targetId: string) => {
    if (targetId === sessionIdRef.current) return

    // Save current
    sessionStoreRef.current[sessionIdRef.current] = {
      messages: [...messages],
      diagnosis,
    }

    // Restore target
    const stored = sessionStoreRef.current[targetId]
    setActiveSessionId(targetId)
    setMessages(stored?.messages || [])
    setDiagnosis(stored?.diagnosis || null)
    setProgress([])
    setToolCalls([])
    setStreamContent('')
    contentRef.current = ''
    doneHandledRef.current = false
    setPhase(stored?.messages?.length ? 'done' : 'idle')
  }, [messages, diagnosis])

  return {
    phase,
    messages,
    diagnosis,
    progress,
    toolCalls,
    streamContent,
    sessionId: activeSessionId,
    sessions,
    startFullRun,
    startDiagnose,
    startPlanFrom,
    cancel,
    newSession,
    switchSession,
  }
}

// Parse SSE lines from a chunk
function parseSseLines(chunk: string, handler: (event: SSEEvent) => void) {
  const lines = chunk.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (trimmed.startsWith('data:')) {
      const jsonStr = trimmed.slice(5).trim()
      if (!jsonStr) continue
      try {
        const data: SSEEvent = JSON.parse(jsonStr)
        handler(data)
      } catch {
        // ignore malformed JSON
      }
    }
  }
}
