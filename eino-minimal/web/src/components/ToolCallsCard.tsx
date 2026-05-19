import type { ToolCall } from '../types'

interface Props {
  toolCalls: ToolCall[]
}

export function ToolCallsCard({ toolCalls }: Props) {
  return (
    <div className="glass-card p-4" style={{ borderRadius: 'var(--radius-card)' }}>
      <h4 className="text-xs font-semibold uppercase mb-3" style={{
        color: 'var(--text-muted)',
        letterSpacing: '0.05em',
        fontFamily: 'var(--font-display)',
      }}>
        Tool Calls
      </h4>

      {toolCalls.length === 0 ? (
        <p className="text-xs" style={{ color: 'var(--text-muted)' }}>No tool calls yet</p>
      ) : (
        <div className="space-y-2">
          {toolCalls.map((tc, i) => (
            <div key={i} className="p-2.5 rounded-lg text-xs" style={{
              background: 'var(--bg-elevated)',
              border: '1px solid var(--glass-border)',
            }}>
              <div className="font-mono font-medium flex items-center gap-1.5" style={{ color: 'var(--primary)' }}>
                <span style={{ fontSize: '10px' }}>&#9654;</span>
                {tc.tool}
              </div>
              {Object.keys(tc.params).length > 0 && (
                <div className="mt-1 truncate font-mono" style={{ color: 'var(--text-muted)', fontSize: '10px' }}>
                  {JSON.stringify(tc.params)}
                </div>
              )}
              {tc.result && (
                <div className="mt-1 truncate" style={{ color: 'var(--success)', fontSize: '10px' }}>
                  {tc.result.slice(0, 80)}{tc.result.length > 80 ? '...' : ''}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
