import { useState } from 'react'
import { useCopilot } from './hooks/useCopilot'
import { SessionList } from './components/SessionList'
import { ChatPanel } from './components/ChatPanel'
import { AgentProgress } from './components/AgentProgress'
import { InputBar } from './components/InputBar'
import { DiagnosisEditor } from './components/DiagnosisEditor'
import { AgentPlayground } from './pages/AgentPlayground'

type View = 'copilot' | 'playground'

function App() {
  const [view, setView] = useState<View>('playground')

  return (
    <div className="flex flex-col h-screen" style={{ background: 'var(--bg-deep)' }}>
      {/* Top nav */}
      <nav className="h-11 flex items-center px-5 gap-1 flex-shrink-0" style={{
        background: 'var(--bg-surface)',
        borderBottom: '1px solid var(--glass-border)',
      }}>
        <span className="text-sm font-semibold mr-4" style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-display)' }}>
          Eino
        </span>
        <NavTab label="Copilot" active={view === 'copilot'} onClick={() => setView('copilot')} />
        <NavTab label="Agent Playground" active={view === 'playground'} onClick={() => setView('playground')} />
      </nav>

      {/* View content — both mounted, toggle visibility to preserve state */}
      <div className="flex-1 min-h-0" style={{ display: view === 'copilot' ? 'flex' : 'none' }}>
        <CopilotView />
      </div>
      <div className="flex-1 min-h-0" style={{ display: view === 'playground' ? 'flex' : 'none' }}>
        <AgentPlayground />
      </div>
    </div>
  )
}

function NavTab({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="text-xs font-medium px-3 py-1.5 rounded-lg"
      style={{
        background: active ? 'rgba(224, 122, 95, 0.06)' : 'transparent',
        color: active ? 'var(--primary)' : 'var(--text-muted)',
        border: active ? '1px solid rgba(224, 122, 95, 0.15)' : '1px solid transparent',
        cursor: 'pointer',
        transition: 'var(--transition-smooth)',
        fontFamily: 'var(--font-display)',
      }}
    >
      {label}
    </button>
  )
}

function CopilotView() {
  const {
    phase,
    messages,
    diagnosis,
    progress,
    toolCalls,
    streamContent,
    sessionId,
    sessions,
    startDiagnose,
    startPlanFrom,
    cancel,
    newSession,
    switchSession,
  } = useCopilot()

  const isRunning = phase === 'diagnosing' || phase === 'planning'

  const handleSend = (message: string) => {
    startDiagnose(message)
  }

  return (
    <div className="flex h-full w-full">
      {/* Left: Sessions */}
      <div className="w-56 flex-shrink-0">
        <SessionList
          sessions={sessions}
          activeSessionId={sessionId}
          onNewSession={newSession}
          onSwitchSession={switchSession}
        />
      </div>

      {/* Center: Chat */}
      <div className="flex-1 flex flex-col min-w-0">
        <div className="h-12 flex items-center px-5" style={{
          background: 'var(--bg-surface)',
          borderBottom: '1px solid var(--glass-border)'
        }}>
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-display)' }}>Eino Copilot</h2>
          <span className="ml-2 text-xs px-2 py-0.5 rounded-full" style={{
            background: isRunning ? 'rgba(224, 122, 95, 0.08)' : 'var(--bg-elevated)',
            border: `1px solid ${isRunning ? 'rgba(224, 122, 95, 0.25)' : 'var(--glass-border)'}`,
            color: isRunning ? 'var(--primary)' : 'var(--text-muted)',
            fontWeight: 500
          }}>
            {phase}
          </span>
        </div>

        <ChatPanel
          messages={messages}
          streamContent={streamContent}
          isStreaming={phase === 'planning' || phase === 'diagnosing'}
        />

        {phase === 'awaiting_edit' && diagnosis && (
          <DiagnosisEditor
            diagnosis={diagnosis}
            onContinue={(edited) => startPlanFrom(edited)}
            onRetry={() => {
              const lastUserMsg = messages.filter(m => m.role === 'user').pop()
              if (lastUserMsg) startDiagnose(lastUserMsg.content)
            }}
          />
        )}

        <InputBar
          onSend={handleSend}
          onCancel={cancel}
          phase={phase}
          disabled={phase === 'awaiting_edit'}
        />
      </div>

      {/* Right: Progress */}
      <div className="w-64 flex-shrink-0">
        <AgentProgress
          progress={progress}
          toolCalls={toolCalls}
          isRunning={isRunning}
          onCancel={cancel}
        />
      </div>
    </div>
  )
}

export default App
