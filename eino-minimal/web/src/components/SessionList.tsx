import type { Session } from '../types'

interface Props {
  sessions: Session[]
  activeSessionId: string
  onNewSession: () => void
  onSwitchSession: (id: string) => void
}

export function SessionList({ sessions, activeSessionId, onNewSession, onSwitchSession }: Props) {
  return (
    <div className="flex flex-col h-full" style={{
      background: 'var(--bg-surface)',
      borderRight: '1px solid var(--glass-border)'
    }}>
      <div className="p-4" style={{ borderBottom: '1px solid var(--glass-border)' }}>
        <button
          onClick={onNewSession}
          className="btn-neon w-full"
        >
          + New Session
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {sessions.map(session => (
          <button
            key={session.id}
            onClick={() => onSwitchSession(session.id)}
            className="w-full text-left px-3 py-2.5 rounded-xl text-sm truncate transition-all duration-200"
            style={{
              background: session.id === activeSessionId ? 'rgba(224, 122, 95, 0.06)' : 'transparent',
              border: session.id === activeSessionId ? '1px solid rgba(224, 122, 95, 0.2)' : '1px solid transparent',
              color: session.id === activeSessionId ? 'var(--primary)' : 'var(--text-secondary)',
              fontWeight: session.id === activeSessionId ? 500 : 400
            }}
          >
            <span
              className="inline-block w-2 h-2 rounded-full mr-2"
              style={{
                background: session.id === activeSessionId ? 'var(--success)' : 'var(--glass-border)',
              }}
            />
            {session.title}
          </button>
        ))}
      </div>

      <div className="p-4 text-xs" style={{ borderTop: '1px solid var(--glass-border)', color: 'var(--text-muted)' }}>
        Eino Copilot v1.0
      </div>
    </div>
  )
}
