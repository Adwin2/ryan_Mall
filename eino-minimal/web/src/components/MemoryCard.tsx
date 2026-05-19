import type { ContextInfo } from '../types'

interface Props {
  info: ContextInfo | null
}

export function MemoryCard({ info }: Props) {
  return (
    <div className="glass-card p-4" style={{ borderRadius: 'var(--radius-card)' }}>
      <h4 className="text-xs font-semibold uppercase mb-3" style={{
        color: 'var(--text-muted)',
        letterSpacing: '0.05em',
        fontFamily: 'var(--font-display)',
      }}>
        Context Builder
      </h4>

      {!info ? (
        <p className="text-xs" style={{ color: 'var(--text-muted)' }}>Awaiting request...</p>
      ) : (
        <div className="space-y-3">
          <MetricRow
            label="History"
            value={`${info.history_turns} / ${info.max_turns} turns`}
            ratio={info.max_turns > 0 ? info.history_turns / info.max_turns : 0}
          />
          <MetricRow
            label="Tokens"
            value={`${(info.token_estimate / 1000).toFixed(1)}k / ${(info.token_budget / 1000).toFixed(1)}k`}
            ratio={info.token_budget > 0 ? info.token_estimate / info.token_budget : 0}
          />
          <div className="flex items-center justify-between">
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>Summary</span>
            <StatusBadge active={info.summary_injected} label={info.summary_injected ? 'injected' : 'none'} />
          </div>
          <div className="flex items-center justify-between">
            <span className="text-xs" style={{ color: 'var(--text-muted)' }}>Semantic Recall</span>
            <StatusBadge
              active={info.recall_count > 0}
              label={info.recall_count > 0 ? `${info.recall_count} facts` : 'disabled'}
            />
          </div>
        </div>
      )}
    </div>
  )
}

function MetricRow({ label, value, ratio }: { label: string; value: string; ratio: number }) {
  return (
    <div>
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs" style={{ color: 'var(--text-muted)' }}>{label}</span>
        <span className="text-xs font-mono" style={{ color: 'var(--text-primary)' }}>{value}</span>
      </div>
      <div className="h-1.5 rounded-full overflow-hidden" style={{ background: 'var(--bg-elevated)' }}>
        <div
          className="h-full rounded-full"
          style={{
            width: `${Math.min(ratio * 100, 100)}%`,
            background: ratio > 0.8 ? 'var(--warning)' : 'var(--primary)',
            transition: 'width 0.4s ease',
          }}
        />
      </div>
    </div>
  )
}

function StatusBadge({ active, label }: { active: boolean; label: string }) {
  return (
    <span className="text-xs px-1.5 py-0.5 rounded" style={{
      background: active ? 'rgba(74, 158, 111, 0.08)' : 'var(--bg-elevated)',
      color: active ? 'var(--success)' : 'var(--text-muted)',
      border: `1px solid ${active ? 'rgba(74, 158, 111, 0.2)' : 'var(--glass-border)'}`,
    }}>
      {active ? '✓' : '✗'} {label}
    </span>
  )
}
