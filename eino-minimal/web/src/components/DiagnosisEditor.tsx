import { useState } from 'react'
import type { DiagnoseResponse } from '../types'

interface Props {
  diagnosis: DiagnoseResponse
  onContinue: (editedDiagnosis?: string) => void
  onRetry: () => void
}

export function DiagnosisEditor({ diagnosis, onContinue, onRetry }: Props) {
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState(diagnosis.diagnosis)

  const handleContinue = () => {
    if (editing && editValue !== diagnosis.diagnosis) {
      onContinue(editValue)
    } else {
      onContinue()
    }
  }

  return (
    <div className="glass-card mx-4 my-2 p-4" style={{
      borderColor: 'rgba(212, 160, 60, 0.25)',
    }}>
      <div className="flex items-center justify-between mb-3">
        <h4 className="text-sm font-semibold" style={{ color: 'var(--warning)', fontFamily: 'var(--font-display)' }}>Diagnosis Complete</h4>
        <button
          onClick={() => setEditing(!editing)}
          className="text-xs underline transition-colors"
          style={{ color: 'var(--text-muted)' }}
          onMouseOver={e => (e.currentTarget.style.color = 'var(--warning)')}
          onMouseOut={e => (e.currentTarget.style.color = 'var(--text-muted)')}
        >
          {editing ? 'Cancel Edit' : 'Edit'}
        </button>
      </div>

      {editing ? (
        <textarea
          value={editValue}
          onChange={e => setEditValue(e.target.value)}
          className="w-full h-48 text-xs font-mono p-3 resize-y outline-none"
          style={{
            background: 'var(--bg-elevated)',
            border: '1.5px solid rgba(212, 160, 60, 0.2)',
            borderRadius: 'var(--radius-sm)',
            color: 'var(--text-primary)'
          }}
        />
      ) : (
        <pre className="text-xs font-mono p-3 max-h-48 overflow-y-auto whitespace-pre-wrap" style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--glass-border)',
          borderRadius: 'var(--radius-sm)',
          color: 'var(--text-secondary)'
        }}>
          {diagnosis.diagnosis || '(No structured diagnosis extracted)'}
        </pre>
      )}

      <div className="flex gap-3 mt-4">
        <button
          onClick={handleContinue}
          className="btn-neon flex-1"
        >
          Continue to Plan
        </button>
        <button
          onClick={onRetry}
          className="btn-ghost"
        >
          Re-diagnose
        </button>
      </div>
    </div>
  )
}
