import { useCallback, useRef, useState } from 'react'
import type { SSEEvent } from '../types'

interface UseSSEOptions {
  onEvent?: (event: SSEEvent) => void
  onError?: (error: string) => void
  onDone?: (event: SSEEvent) => void
}

interface UseSSEReturn {
  status: 'idle' | 'connecting' | 'open' | 'closed' | 'error'
  connect: (url: string) => void
  close: () => void
}

export function useSSE(options: UseSSEOptions = {}): UseSSEReturn {
  const [status, setStatus] = useState<UseSSEReturn['status']>('idle')
  const esRef = useRef<EventSource | null>(null)
  const optionsRef = useRef(options)
  optionsRef.current = options

  const close = useCallback(() => {
    if (esRef.current) {
      esRef.current.close()
      esRef.current = null
    }
    setStatus('closed')
  }, [])

  const connect = useCallback((url: string) => {
    close()
    setStatus('connecting')

    const es = new EventSource(url)
    esRef.current = es

    es.onopen = () => {
      setStatus('open')
    }

    es.onmessage = (event) => {
      try {
        const data: SSEEvent = JSON.parse(event.data)
        optionsRef.current.onEvent?.(data)

        if (data.type === 'error') {
          optionsRef.current.onError?.(data.content || 'Unknown error')
        }
        if (data.type === 'done') {
          optionsRef.current.onDone?.(data)
          close()
        }
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      setStatus('error')
      close()
    }
  }, [close])

  return { status, connect, close }
}
