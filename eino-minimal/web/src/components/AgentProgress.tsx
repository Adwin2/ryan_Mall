import type { NodeProgress, ToolCall } from '../types'
import { DAG_NODES } from '../types'

interface Props {
  progress: NodeProgress[]
  toolCalls: ToolCall[]
  duration?: number
  onCancel?: () => void
  isRunning: boolean
}

const SVG_WIDTH = 224
const SVG_HEIGHT = 310
const NODE_WIDTH = 88
const NODE_HEIGHT = 28
const NODE_RX = 14

const NODE_POSITIONS = [
  { x: 68,  y: 32 },
  { x: 156, y: 82 },
  { x: 68,  y: 132 },
  { x: 156, y: 182 },
  { x: 68,  y: 244 },
  { x: 156, y: 294 },
]

function getEdgePath(from: { x: number; y: number }, to: { x: number; y: number }): string {
  const startY = from.y + NODE_HEIGHT / 2
  const endY = to.y - NODE_HEIGHT / 2
  const midY = (startY + endY) / 2
  return `M ${from.x} ${startY} C ${from.x} ${midY}, ${to.x} ${midY}, ${to.x} ${endY}`
}

export function AgentProgress({ progress, toolCalls, duration, onCancel, isRunning }: Props) {
  const getNodeStatus = (nodeKey: string) => {
    const p = progress.find(n => n.node === nodeKey)
    return p?.status || 'pending'
  }

  return (
    <div className="flex flex-col h-full" style={{
      background: 'var(--bg-surface)',
      borderLeft: '1px solid var(--glass-border)'
    }}>
      <div className="p-4" style={{ borderBottom: '1px solid var(--glass-border)' }}>
        <h3 className="text-sm font-semibold" style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-display)' }}>Agent Progress</h3>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <svg
          width="100%"
          viewBox={`0 0 ${SVG_WIDTH} ${SVG_HEIGHT}`}
          role="img"
          aria-label="Agent pipeline DAG"
        >
          {DAG_NODES.slice(1).map((node, i) => (
            <DagEdge
              key={`edge-${node.key}`}
              from={NODE_POSITIONS[i]}
              to={NODE_POSITIONS[i + 1]}
              status={getNodeStatus(node.key)}
            />
          ))}

          <line
            x1="24" y1="212" x2="200" y2="212"
            stroke="var(--glass-border)"
            strokeWidth="0.5"
            strokeDasharray="3 3"
          />
          <text
            x={SVG_WIDTH / 2} y="225"
            textAnchor="middle"
            fontSize="9"
            fontFamily="var(--font-body)"
            fill="var(--text-muted)"
          >
            Planning Phase
          </text>

          {DAG_NODES.map((node, i) => (
            <DagNode
              key={node.key}
              position={NODE_POSITIONS[i]}
              label={node.label}
              status={getNodeStatus(node.key)}
              index={i}
            />
          ))}
        </svg>

        {toolCalls.length > 0 && (
          <div className="mt-4 pt-4" style={{ borderTop: '1px solid var(--glass-border)' }}>
            <h4 className="text-xs font-semibold uppercase mb-2" style={{ color: 'var(--text-muted)', letterSpacing: '0.05em' }}>
              Tool Calls
            </h4>
            <div className="space-y-2">
              {toolCalls.map((tc, i) => (
                <div key={i} className="glass-card p-2.5 text-xs" style={{ borderRadius: 'var(--radius-sm)' }}>
                  <div className="font-mono font-medium" style={{ color: 'var(--primary)' }}>{tc.tool}</div>
                  <div className="mt-1 truncate" style={{ color: 'var(--text-muted)' }}>
                    {JSON.stringify(tc.params)}
                  </div>
                  {tc.result && (
                    <div className="mt-1 truncate" style={{ color: 'var(--success)' }}>{tc.result.slice(0, 80)}...</div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="p-4 space-y-3" style={{ borderTop: '1px solid var(--glass-border)' }}>
        {duration !== undefined && (
          <div className="text-xs" style={{ color: 'var(--text-muted)' }}>
            Duration: {(duration / 1000).toFixed(1)}s
          </div>
        )}
        {isRunning && onCancel && (
          <button
            onClick={onCancel}
            className="btn-neon-danger btn-neon w-full"
          >
            Cancel
          </button>
        )}
      </div>
    </div>
  )
}

function DagEdge({ from, to, status }: { from: { x: number; y: number }; to: { x: number; y: number }; status: string }) {
  const d = getEdgePath(from, to)
  const strokeColor = status === 'done'
    ? 'var(--success)'
    : status === 'running'
      ? 'var(--primary)'
      : 'var(--glass-border)'
  const strokeWidth = status === 'done' ? 2 : 1.5

  return (
    <path
      d={d}
      fill="none"
      stroke={strokeColor}
      strokeWidth={strokeWidth}
      strokeLinecap="round"
      className={status === 'running' ? 'dag-edge-running' : ''}
      strokeDasharray={status === 'pending' ? '3 3' : undefined}
      style={{ transition: 'stroke 0.3s ease, stroke-width 0.3s ease' }}
    />
  )
}

function DagNode({ position, label, status, index }: {
  position: { x: number; y: number }
  label: string
  status: string
  index: number
}) {
  const fill = status === 'done'
    ? 'rgba(74, 158, 111, 0.06)'
    : 'var(--bg-surface)'
  const stroke = status === 'done'
    ? 'var(--success)'
    : status === 'running'
      ? 'var(--primary)'
      : 'var(--glass-border)'
  const textColor = status === 'running'
    ? 'var(--primary)'
    : status === 'done'
      ? 'var(--text-primary)'
      : 'var(--text-muted)'

  return (
    <g
      className={status === 'running' ? 'dag-node-running' : ''}
      style={{
        animation: 'dag-node-enter 0.35s ease forwards',
        animationDelay: `${index * 60}ms`,
        opacity: 0,
      }}
    >
      <rect
        x={position.x - NODE_WIDTH / 2}
        y={position.y - NODE_HEIGHT / 2}
        width={NODE_WIDTH}
        height={NODE_HEIGHT}
        rx={NODE_RX}
        fill={fill}
        stroke={stroke}
        strokeWidth={status === 'running' ? 1.5 : 1}
        style={{ transition: 'fill 0.3s ease, stroke 0.3s ease' }}
      />
      {status === 'done' && (
        <text
          x={position.x - NODE_WIDTH / 2 + 12}
          y={position.y + 4}
          fontSize="10"
          fill="var(--success)"
        >
          &#10003;
        </text>
      )}
      <text
        x={position.x + (status === 'done' ? 6 : 0)}
        y={position.y + 4}
        textAnchor="middle"
        fontSize="11"
        fontFamily="var(--font-body)"
        fontWeight={status === 'running' ? 500 : 400}
        fill={textColor}
        style={{ transition: 'fill 0.3s ease' }}
      >
        {label}
      </text>
    </g>
  )
}
