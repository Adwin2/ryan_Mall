import { useState } from 'react'
import type { ThinkingStep } from '../types'

interface Props {
  steps: ThinkingStep[]
  isStreaming: boolean
}

export function ThinkingBlock({ steps, isStreaming }: Props) {
  const [expanded, setExpanded] = useState(false)

  if (steps.length === 0 && !isStreaming) return null

  return (
    <div className="mx-5 mb-3">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-2 text-xs font-medium px-3 py-1.5 rounded-lg w-full text-left"
        style={{
          background: 'rgba(224, 122, 95, 0.04)',
          border: '1px solid rgba(224, 122, 95, 0.12)',
          color: 'var(--text-muted)',
          cursor: 'pointer',
          transition: 'var(--transition-smooth)',
        }}
      >
        {/* Animated thinking indicator */}
        {isStreaming && steps.length > 0 && (
          <span className="flex gap-0.5">
            <span className="w-1 h-1 rounded-full animate-pulse" style={{ background: 'var(--primary)', animationDelay: '0ms' }} />
            <span className="w-1 h-1 rounded-full animate-pulse" style={{ background: 'var(--primary)', animationDelay: '150ms' }} />
            <span className="w-1 h-1 rounded-full animate-pulse" style={{ background: 'var(--primary)', animationDelay: '300ms' }} />
          </span>
        )}

        {/* Chevron */}
        <svg
          width="12" height="12" viewBox="0 0 12 12" fill="none"
          style={{
            transform: expanded ? 'rotate(90deg)' : 'rotate(0deg)',
            transition: 'transform 0.2s ease',
            opacity: 0.5,
          }}
        >
          <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>

        <span style={{ fontFamily: 'var(--font-display)' }}>
          Thinking{steps.length > 0 ? ` (${steps.length} step${steps.length > 1 ? 's' : ''})` : '...'}
        </span>

        {/* Duration badge */}
        {steps.length >= 2 && (
          <span className="ml-auto text-xs font-mono" style={{ color: 'var(--text-muted)', opacity: 0.6 }}>
            {((steps[steps.length - 1].timestamp - steps[0].timestamp) / 1000).toFixed(1)}s
          </span>
        )}
      </button>

      {expanded && (
        <div
          className="mt-2 rounded-lg overflow-hidden"
          style={{
            background: 'var(--bg-elevated)',
            border: '1px solid var(--glass-border)',
          }}
        >
          {steps.map((step, i) => (
            <div
              key={i}
              className="px-3 py-2 text-xs font-mono leading-relaxed"
              style={{
                color: 'var(--text-muted)',
                opacity: 0.75,
                borderBottom: i < steps.length - 1 ? '1px solid var(--glass-border)' : 'none',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              <span className="inline-block w-4 text-right mr-2 select-none" style={{ opacity: 0.4 }}>
                {i + 1}
              </span>
              {step.content}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
