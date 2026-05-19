import type { ProfileMetrics } from '../types'

interface Props {
  metrics: Record<string, ProfileMetrics> | null
}

export function MetricsCard({ metrics }: Props) {
  const entries = metrics ? Object.entries(metrics) : []
  const totalCalls = entries.reduce((sum, [, m]) => sum + m.CallCount, 0)

  return (
    <div className="glass-card p-4" style={{ borderRadius: 'var(--radius-card)' }}>
      <h4 className="text-xs font-semibold uppercase mb-3" style={{
        color: 'var(--text-muted)',
        letterSpacing: '0.05em',
        fontFamily: 'var(--font-display)',
      }}>
        Routing Metrics
      </h4>

      {entries.length === 0 ? (
        <p className="text-xs" style={{ color: 'var(--text-muted)' }}>No data yet</p>
      ) : (
        <div className="space-y-2.5">
          {entries.map(([name, m]) => (
            <div key={name}>
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs font-mono truncate" style={{ color: 'var(--text-secondary)', maxWidth: '60%' }}>
                  {name}
                </span>
                <span className="text-xs font-medium" style={{ color: 'var(--text-primary)' }}>
                  {m.CallCount}
                </span>
              </div>
              <div className="h-1 rounded-full overflow-hidden" style={{ background: 'var(--bg-elevated)' }}>
                <div
                  className="h-full rounded-full"
                  style={{
                    width: `${totalCalls > 0 ? (m.CallCount / totalCalls) * 100 : 0}%`,
                    background: 'var(--primary)',
                    transition: 'width 0.4s ease',
                  }}
                />
              </div>
            </div>
          ))}

          {totalCalls > 0 && (
            <div className="pt-2 mt-2" style={{ borderTop: '1px solid var(--glass-border)' }}>
              <div className="flex justify-between text-xs" style={{ color: 'var(--text-muted)' }}>
                <span>Total calls</span>
                <span style={{ color: 'var(--text-primary)' }}>{totalCalls}</span>
              </div>
              {entries.length > 0 && (() => {
                const avgLatency = entries.reduce((sum, [, m]) => sum + m.TotalLatencyMs, 0) / totalCalls
                return (
                  <div className="flex justify-between text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                    <span>Avg latency</span>
                    <span style={{ color: 'var(--text-primary)' }}>{(avgLatency / 1000).toFixed(1)}s</span>
                  </div>
                )
              })()}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
