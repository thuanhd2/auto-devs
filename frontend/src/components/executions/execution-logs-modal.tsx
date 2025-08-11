import { useEffect, useState } from 'react'
import { ExecutionLog } from '@/types/execution'
import { AlertTriangle } from 'lucide-react'
import {
  AlertCircle,
  Bot,
  Brain,
  CheckSquare,
  ChevronRight,
  ChevronUp,
  Edit,
  Eye,
  Globe,
  Plus,
  Search,
  Settings,
  Terminal,
  User,
} from 'lucide-react'
import { useWebSocketContext } from '@/context/websocket-context'
import { useExecutionLogs } from '@/hooks/use-executions'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'

interface ExecutionLogsModalProps {
  open: boolean
  executionId: string | null
  onClose: () => void
}

export function ExecutionLogsModal({
  open,
  executionId,
  onClose,
}: ExecutionLogsModalProps) {
  const { data, isLoading, error } = useExecutionLogs(executionId)
  const logs = data?.data || []
  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className='sm:w-[400px] sm:max-w-[400px] lg:h-[80vh] lg:w-[80vw] lg:max-w-none'>
        <DialogHeader>
          <DialogTitle>Execution Logs</DialogTitle>
        </DialogHeader>
        <div className='h-[60vh]'>
          {isLoading && (
            <div className='text-muted-foreground text-sm'>Loading logs...</div>
          )}
          {error && (
            <div className='mb-2 flex items-center gap-2 rounded border border-red-200 bg-red-50 p-2 text-red-700'>
              <AlertTriangle className='h-4 w-4' />
              <span>{error}</span>
            </div>
          )}
          {!isLoading && !error && (
            <ScrollArea
              className='h-full rounded border p-2 text-xs'
              stickToBottom
            >
              {logs ? (
                logs.map((log) => (
                  <div key={log.id}>
                    <ExecutionLogItem log={log} />
                  </div>
                ))
              ) : (
                <span className='text-muted-foreground'>
                  No logs to display.
                </span>
              )}
            </ScrollArea>
          )}
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function ExecutionLogItem({ log }: { log: ExecutionLog }) {
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

      if (logData.type === 'result') {
        return (
          <div className='rounded border-l-2 border-emerald-400 bg-emerald-50 p-2'>
            <div className='mb-1 text-xs font-medium text-emerald-600'>
              Execution Result ({logData.subtype})
            </div>
            <div className='text-sm text-gray-800'>
              Duration: {logData.duration_ms}ms | Turns: {logData.num_turns}
            </div>
            {logData.result && (
              <div className='mt-2 text-sm text-gray-700'>{logData.result}</div>
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
          <AlertTriangle className='mt-1 h-4 w-4 text-red-600' />
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
