import { useEffect, useRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { ChatMessage } from '../types'

interface Props {
  messages: ChatMessage[]
  streamContent: string
  isStreaming: boolean
}

export function ChatPanel({ messages, streamContent, isStreaming }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamContent])

  return (
    <div className="flex-1 overflow-y-auto p-5 space-y-4" style={{ background: 'var(--bg-deep)' }}>
      {messages.map(msg => (
        <MessageBubble key={msg.id} message={msg} />
      ))}

      {isStreaming && streamContent && (
        <div className="flex justify-start">
          <div className="glass-card max-w-[80%] p-4" style={{
            borderColor: 'rgba(224, 122, 95, 0.15)',
          }}>
            <div className="prose prose-sm max-w-none break-words">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{streamContent}</ReactMarkdown>
            </div>
            <div className="mt-3 flex items-center gap-2">
              <span className="inline-block w-2 h-2 rounded-full neon-pulse" style={{ background: 'var(--primary)' }} />
              <span className="text-xs" style={{ color: 'var(--text-muted)' }}>Generating...</span>
            </div>
          </div>
        </div>
      )}

      {!isStreaming && streamContent && (
        <div className="flex justify-start">
          <div className="glass-card max-w-[80%] p-4" style={{
            borderColor: 'rgba(224, 122, 95, 0.15)',
          }}>
            <div className="prose prose-sm max-w-none break-words">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{streamContent}</ReactMarkdown>
            </div>
          </div>
        </div>
      )}

      {isStreaming && !streamContent && (
        <div className="flex justify-start">
          <div className="glass-card px-5 py-3 flex items-center gap-2" style={{
            borderColor: 'rgba(224, 122, 95, 0.12)'
          }}>
            <span className="inline-block w-2 h-2 rounded-full neon-pulse" style={{ background: 'var(--primary)' }} />
            <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>Thinking...</span>
          </div>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  )
}

function MessageBubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === 'user'

  if (!message.content?.trim()) return null

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[80%] rounded-2xl p-4 ${isUser ? '' : 'glass-card'}`}
        style={isUser ? {
          background: 'var(--primary)',
          color: '#faf8f5',
          boxShadow: 'var(--shadow-glow)'
        } : {}}
      >
        {isUser ? (
          <p className="text-sm whitespace-pre-wrap" style={{ color: '#faf8f5' }}>{message.content}</p>
        ) : (
          <div className="prose prose-sm max-w-none break-words">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{message.content}</ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  )
}
