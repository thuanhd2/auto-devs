import {
  AlertCircle,
  Bot,
  CheckSquare,
  Settings,
  Terminal,
  User,
} from 'lucide-react'
import { ExecutionLog } from '@/types/execution'

interface StructuredLogItemProps {
  log: ExecutionLog
}

export function StructuredLogItem({ log }: StructuredLogItemProps) {
  // Use structured fields if available, fallback to raw message parsing
  if (log.log_type && log.parsed_content) {
    return <StructuredLogRenderer log={log} />
  }

  // Fallback to legacy JSON parsing for backward compatibility
  return <LegacyLogRenderer log={log} />
}

// Helper function to format result messages
function formatResultMessage(message: string): string {
  // Handle Claude AI usage limit message format
  if (message.match(/Claude AI usage limit reached\|(\d+)/)) {
    const timestamp = parseInt(message.split('|')[1]) * 1000
    return `Claude AI usage limit reached at ${new Date(timestamp).toLocaleString()}`
  }
  return message
}

function StructuredLogRenderer({ log }: { log: ExecutionLog }) {
  const { log_type, tool_name, tool_use_id, parsed_content, is_error, duration_ms, num_turns } = log

  const getIcon = () => {
    switch (log_type) {
      case 'user':
        return <User className='h-4 w-4 text-blue-600' />
      case 'assistant':
        return <Bot className='h-4 w-4 text-green-600' />
      case 'tool_result':
        return <Terminal className='h-4 w-4 text-purple-600' />
      case 'result':
        return <CheckSquare className='h-4 w-4 text-emerald-600' />
      default:
        return <AlertCircle className='h-4 w-4 text-gray-600' />
    }
  }

  const formatContent = () => {
    switch (log_type) {
      case 'user':
        return (
          <div className='text-sm text-blue-700'>
            {parsed_content?.text || 'User message'}
          </div>
        )

      case 'assistant':
        if (parsed_content?.content) {
          return parsed_content.content.map((item: any, index: number) => {
            if (item.type === 'text') {
              return (
                <div
                  key={index}
                  className='text-sm whitespace-pre-wrap text-gray-800'
                >
                  {item.text}
                </div>
              )
            }
            if (item.type === 'tool_use') {
              return (
                <div
                  key={index}
                  className='rounded border-l-2 border-orange-400 bg-orange-50 p-2'
                >
                  <div className='mb-2 flex items-center gap-2'>
                    <Settings className='h-3 w-3 text-orange-600' />
                    <span className='text-xs font-medium text-orange-600'>
                      Tool: {item.name}
                    </span>
                  </div>
                  {item.input && (
                    <div className='text-xs text-gray-600'>
                      {typeof item.input === 'object'
                        ? JSON.stringify(item.input, null, 2)
                        : item.input}
                    </div>
                  )}
                </div>
              )
            }
            return (
              <div key={index} className='text-sm'>
                {JSON.stringify(item)}
              </div>
            )
          })
        }
        return <div className='text-sm text-green-700'>Assistant message</div>

      case 'tool_result':
        return (
          <div className='rounded border-l-2 border-purple-400 bg-purple-50 p-2'>
            <div className='mb-1 text-xs font-medium text-purple-600'>
              Tool Result {tool_use_id && `(${tool_use_id})`}
            </div>
            <div className='text-sm text-gray-800'>
              {parsed_content?.content || 'Tool execution result'}
            </div>
          </div>
        )

      case 'result':
        return (
          <div className={`rounded border-l-2 ${
            is_error 
              ? 'border-red-400 bg-red-50' 
              : 'border-emerald-400 bg-emerald-50'
          } p-2`}>
            <div className={`mb-1 text-xs font-medium ${
              is_error ? 'text-red-600' : 'text-emerald-600'
            }`}>
              Execution Result
            </div>
            <div className='text-sm text-gray-800'>
              {duration_ms && `Duration: ${duration_ms}ms`}
              {num_turns && ` | Turns: ${num_turns}`}
            </div>
            {parsed_content?.result && (
              <div className='mt-2 text-sm text-gray-700'>
                {formatResultMessage(parsed_content.result)}
              </div>
            )}
            {is_error && (
              <div className='mt-2 text-sm text-red-600'>
                Error: {parsed_content?.error || 'Unknown error'}
              </div>
            )}
          </div>
        )

      default:
        return (
          <div className='text-sm text-gray-600'>
            {JSON.stringify(parsed_content, null, 2)}
          </div>
        )
    }
  }

  return (
    <div className='mb-3 border-b border-gray-100 pb-3 last:border-b-0'>
      <div className='flex items-start gap-2'>
        <div className='mt-1 flex-shrink-0'>{getIcon()}</div>
        <div className='min-w-0 flex-1'>
          <div className='mb-1 flex items-center gap-2'>
            <span className='text-xs font-medium text-gray-500'>
              {log_type}
            </span>
            {tool_name && (
              <span className='text-xs text-gray-400'>
                Tool: {tool_name}
              </span>
            )}
            <span className='text-xs text-gray-400'>
              {new Date(log.timestamp).toLocaleTimeString()}
            </span>
          </div>
          <div className='space-y-2'>{formatContent()}</div>
        </div>
      </div>
    </div>
  )
}

function LegacyLogRenderer({ log }: { log: ExecutionLog }) {
  const message = log.message
  if (!message) {
    return null
  }

  try {
    const logData = JSON.parse(message)

    // Get icon based on message type
    const getIcon = () => {
      switch (logData.type) {
        case 'user':
          return <User className='h-4 w-4 text-blue-600' />
        case 'assistant':
          // Check if this is a tool_use message
          if (
            logData.message?.content?.some(
              (item: any) => item.type === 'tool_use'
            )
          ) {
            return <Settings className='h-4 w-4 text-orange-600' />
          }
          return <Bot className='h-4 w-4 text-green-600' />
        case 'tool_result':
          return <Terminal className='h-4 w-4 text-purple-600' />
        case 'result':
          return <CheckSquare className='h-4 w-4 text-emerald-600' />
        default:
          return <AlertCircle className='h-4 w-4 text-gray-600' />
      }
    }

    // Format message content
    const formatContent = () => {
      if (logData.type === 'user') {
        // Handle user messages (including tool results)
        const content = logData.message?.content
        if (Array.isArray(content)) {
          return content.map((item: any, index: number) => {
            if (item.type === 'tool_result') {
              return (
                <div
                  key={index}
                  className='rounded border-l-2 border-purple-400 bg-gray-50 p-2'
                >
                  <div className='mb-1 text-xs font-medium text-purple-600'>
                    Tool Result ({item.tool_use_id})
                  </div>
                  <div className='text-sm'>
                    {Array.isArray(item.content)
                      ? item.content.map((c: any) => c.text).join(' ')
                      : item.content}
                  </div>
                </div>
              )
            }
            return (
              <div key={index} className='text-sm'>
                {item.text || JSON.stringify(item)}
              </div>
            )
          })
        }
        return <div className='text-sm text-blue-700'>User message</div>
      }

      if (logData.type === 'assistant') {
        const content = logData.message?.content
        if (Array.isArray(content)) {
          return content.map((item: any, index: number) => {
            if (item.type === 'text') {
              return (
                <div
                  key={index}
                  className='text-sm whitespace-pre-wrap text-gray-800'
                >
                  {item.text}
                </div>
              )
            }
            if (item.type === 'tool_use') {
              return (
                <div
                  key={index}
                  className='rounded border-l-2 border-orange-400 bg-orange-50 p-2'
                >
                  <div className='mb-2 flex items-center gap-2'>
                    <Settings className='h-3 w-3 text-orange-600' />
                    <span className='text-xs font-medium text-orange-600'>
                      Tool: {item.name}
                    </span>
                  </div>
                  {item.name === 'TodoWrite' && item.input?.todos && (
                    <div className='space-y-1'>
                      <div className='flex items-center gap-2'>
                        <CheckSquare className='h-3 w-3 text-emerald-600' />
                        <span className='text-xs font-medium text-emerald-600'>
                          TODO List Update
                        </span>
                      </div>
                      {item.input.todos.map((todo: any, todoIndex: number) => (
                        <div key={todoIndex} className='py-1 pl-5 text-xs'>
                          <span
                            className={`mr-2 rounded px-1.5 py-0.5 text-xs font-medium ${
                              todo.status === 'completed'
                                ? 'bg-green-100 text-green-700'
                                : todo.status === 'in_progress'
                                  ? 'bg-yellow-100 text-yellow-700'
                                  : 'bg-gray-100 text-gray-700'
                            }`}
                          >
                            {todo.status}
                          </span>
                          {todo.content}
                        </div>
                      ))}
                    </div>
                  )}
                  {item.name !== 'TodoWrite' && (
                    <div className='text-xs text-gray-600'>
                      {typeof item.input === 'object'
                        ? JSON.stringify(item.input, null, 2)
                        : item.input}
                    </div>
                  )}
                </div>
              )
            }
            return (
              <div key={index} className='text-sm'>
                {JSON.stringify(item)}
              </div>
            )
          })
        }
        return <div className='text-sm text-green-700'>Assistant message</div>
      }

      if (logData.type === 'tool_result') {
        return (
          <div className='rounded border-l-2 border-purple-400 bg-purple-50 p-2'>
            <div className='mb-1 text-xs font-medium text-purple-600'>
              Tool Result ({logData.tool_use_id})
            </div>
            <div className='text-sm text-gray-800'>
              {logData.content || 'Tool execution result'}
            </div>
          </div>
        )
      }

      if (logData.type === 'result') {
        const isError = logData.is_error
        return (
          <div className={`rounded border-l-2 ${
            isError 
              ? 'border-red-400 bg-red-50' 
              : 'border-emerald-400 bg-emerald-50'
          } p-2`}>
            <div className={`mb-1 text-xs font-medium ${
              isError ? 'text-red-600' : 'text-emerald-600'
            }`}>
              Execution Result ({logData.subtype})
            </div>
            <div className='text-sm text-gray-800'>
              Duration: {logData.duration_ms}ms | Turns: {logData.num_turns}
            </div>
            {logData.result && (
              <div className='mt-2 text-sm text-gray-700'>
                {formatResultMessage(logData.result)}
              </div>
            )}
          </div>
        )
      }

      // Default fallback
      return (
        <div className='text-sm text-gray-600'>
          {JSON.stringify(logData, null, 2)}
        </div>
      )
    }

    return (
      <div className='mb-3 border-b border-gray-100 pb-3 last:border-b-0'>
        <div className='flex items-start gap-2'>
          <div className='mt-1 flex-shrink-0'>{getIcon()}</div>
          <div className='min-w-0 flex-1'>
            <div className='mb-1 flex items-center gap-2'>
              <span className='text-xs font-medium text-gray-500'>
                {logData.type}
              </span>
              <span className='text-xs text-gray-400'>
                {new Date(log.timestamp).toLocaleTimeString()}
              </span>
            </div>
            <div className='space-y-2'>{formatContent()}</div>
          </div>
        </div>
      </div>
    )
  } catch (error) {
    // Fallback for invalid JSON
    return (
      <div className='mb-3 border-b border-gray-100 pb-3 last:border-b-0'>
        <div className='flex items-start gap-2'>
          <AlertCircle className='mt-1 h-4 w-4 text-red-600' />
          <div className='min-w-0 flex-1'>
            <div className='mb-1 text-xs font-medium text-red-600'>
              Invalid log format
            </div>
            <div className='rounded bg-gray-50 p-2 font-mono text-sm text-gray-600'>
              {message}
            </div>
          </div>
        </div>
      </div>
    )
  }
}
