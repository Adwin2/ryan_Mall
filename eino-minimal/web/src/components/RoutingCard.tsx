import type { RoutingDecision } from '../types'

interface Props {
  decision: RoutingDecision | null
}

export function RoutingCard({ decision }: Props) {
  return (
    <div className="glass-card p-4" style={{ borderRadius: 'var(--radius-card)' }}>
      <h4 className="text-xs font-semibold uppercase mb-3" style={{
        color: 'var(--text-muted)',
        letterSpacing: '0.05em',
        fontFamily: 'var(--font-display)',
      }}>
        Model Routing
      </h4>

      {!decision ? (
        <p className="text-xs" style={{ color: 'var(--text-muted)' }}>Awaiting request...</p>
      ) : (
        <div className="space-y-2.5">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>
              {decision.model}
            </span>
            <span className="text-xs px-2 py-0.5 rounded-full font-medium" style={{
              background: decision.tier === 'strong'
                ? 'rgba(74, 158, 111, 0.08)'
                : 'rgba(59, 130, 246, 0.08)',
              color: decision.tier === 'strong' ? 'var(--success)' : '#3b82f6',
              border: `1px solid ${decision.tier === 'strong' ? 'rgba(74, 158, 111, 0.2)' : 'rgba(59, 130, 246, 0.2)'}`,
            }}>
              {decision.tier}
            </span>
          </div>

          <div>
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>Profile: </span>
            <span className="text-xs font-mono" style={{ color: 'var(--primary)' }}>{decision.profile}</span>
          </div>

          <div className="text-xs" style={{ color: 'var(--text-secondary)' }}>
            {decision.reason}
          </div>
        </div>
      )}
    </div>
  )
}
