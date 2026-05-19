import { useMultiAgent } from '../hooks/useMultiAgent'
import { ChatPanel } from '../components/ChatPanel'
import { InputBar } from '../components/InputBar'
import { ThinkingBlock } from '../components/ThinkingBlock'
import { RoutingCard } from '../components/RoutingCard'
import { MemoryCard } from '../components/MemoryCard'
import { MetricsCard } from '../components/MetricsCard'
import { ToolCallsCard } from '../components/ToolCallsCard'

export function AgentPlayground() {
  const {
    messages,
    streamContent,
    isStreaming,
    routingDecision,
    contextInfo,
    metrics,
    toolCalls,
    thinkingSteps,
    sendMessage,
    cancel,
    agent,
    setAgent,
    sessionId,
  } = useMultiAgent()

  return (
    <div className="flex flex-col h-full w-full">
      {/* Agent tab bar */}
      <div className="h-12 flex items-center px-5 gap-4" style={{
        background: 'var(--bg-surface)',
        borderBottom: '1px solid var(--glass-border)',
      }}>
        <AgentTab
          label="Ops Agent"
          active={agent === 'ops'}
          onClick={() => setAgent('ops')}
        />
        <AgentTab
          label="Shop Agent"
          active={agent === 'consumer'}
          onClick={() => setAgent('consumer')}
        />
        <div className="ml-auto flex items-center gap-2">
          <span className="text-xs font-mono px-2 py-0.5 rounded" style={{
            background: 'var(--bg-elevated)',
            color: 'var(--text-muted)',
            border: '1px solid var(--glass-border)',
          }}>
            {sessionId.slice(0, 16)}
          </span>
          {isStreaming && (
            <span className="inline-block w-2 h-2 rounded-full neon-pulse" style={{ background: 'var(--primary)' }} />
          )}
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 min-h-0">
        {/* Chat area */}
        <div className="flex-1 flex flex-col min-w-0">
          <ChatPanel
            messages={messages}
            streamContent={streamContent}
            isStreaming={isStreaming}
          />
          <ThinkingBlock steps={thinkingSteps} isStreaming={isStreaming} />
          <InputBar
            onSend={sendMessage}
            onCancel={cancel}
            phase={isStreaming ? 'planning' : 'idle'}
          />
        </div>

        {/* Right sidebar: infra cards */}
        <div className="w-64 flex-shrink-0 flex flex-col overflow-y-auto p-3 gap-3" style={{
          background: 'var(--bg-surface)',
          borderLeft: '1px solid var(--glass-border)',
        }}>
          <RoutingCard decision={routingDecision} />
          <MemoryCard info={contextInfo} />
          <ToolCallsCard toolCalls={toolCalls} />
          <MetricsCard metrics={metrics} />
        </div>
      </div>
    </div>
  )
}

function AgentTab({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="text-sm font-medium px-3 py-1.5 rounded-lg"
      style={{
        background: active ? 'rgba(224, 122, 95, 0.08)' : 'transparent',
        color: active ? 'var(--primary)' : 'var(--text-secondary)',
        border: active ? '1px solid rgba(224, 122, 95, 0.2)' : '1px solid transparent',
        fontFamily: 'var(--font-display)',
        cursor: 'pointer',
        transition: 'var(--transition-smooth)',
      }}
    >
      {active && <span className="inline-block w-1.5 h-1.5 rounded-full mr-1.5" style={{ background: 'var(--primary)' }} />}
      {label}
    </button>
  )
}
