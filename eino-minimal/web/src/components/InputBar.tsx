import { useState } from 'react'
import type { Phase } from '../types'

interface Props {
  onSend: (message: string) => void
  onCancel?: () => void
  phase: Phase
  disabled?: boolean
}

export function InputBar({ onSend, onCancel, phase, disabled }: Props) {
  const [value, setValue] = useState('')

  const isRunning = phase === 'diagnosing' || phase === 'planning'

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!value.trim() || disabled) return
    onSend(value.trim())
    setValue('')
  }

  return (
    <form onSubmit={handleSubmit} className="p-4" style={{
      background: 'var(--bg-surface)',
      borderTop: '1px solid var(--glass-border)'
    }}>
      <div className="flex gap-3">
        <input
          type="text"
          value={value}
          onChange={e => setValue(e.target.value)}
          placeholder={phase === 'idle' ? '输入运营目标，如"帮我在女装类目做618大促，预算50万"' : 'Follow up...'}
          disabled={disabled || isRunning}
          className="flex-1 px-5 py-3 text-sm outline-none"
          style={{
            background: 'var(--bg-elevated)',
            border: '1.5px solid var(--glass-border)',
            borderRadius: 'var(--radius-pill)',
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-body)',
            transition: 'var(--transition-smooth)',
          }}
          onFocus={e => {
            e.currentTarget.style.borderColor = 'var(--primary)'
            e.currentTarget.style.boxShadow = '0 0 0 3px rgba(224, 122, 95, 0.08)'
          }}
          onBlur={e => {
            e.currentTarget.style.borderColor = 'var(--glass-border)'
            e.currentTarget.style.boxShadow = 'none'
          }}
        />
        {isRunning ? (
          <button
            type="button"
            onClick={onCancel}
            className="btn-neon btn-neon-danger"
          >
            Stop
          </button>
        ) : (
          <button
            type="submit"
            disabled={!value.trim() || disabled}
            className="btn-neon"
          >
            Send
          </button>
        )}
      </div>
    </form>
  )
}
