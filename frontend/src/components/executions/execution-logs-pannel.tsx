import { useEffect, useState, useRef } from 'react'
import { AlertTriangle } from 'lucide-react'
import { useExecution } from '@/hooks/use-executions'
import { ScrollArea } from '@/components/ui/scroll-area'
import { StructuredLogItem } from './structured-log-item'

interface ExecutionLogsPannelProps {
  executionId: string | null
}

export function ExecutionLogsPannel({ executionId }: ExecutionLogsPannelProps) {
  const { data: execution, isLoading, error } = useExecution(executionId)
  const logs = (execution?.logs || []).sort((a, b) => a.line - b.line)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const [prevLogsLength, setPrevLogsLength] = useState(0)

  // Auto scroll to bottom when new logs arrive
  useEffect(() => {
    if (logs.length > prevLogsLength && scrollAreaRef.current) {
      const scrollContainer = scrollAreaRef.current.querySelector(
        '[data-radix-scroll-area-viewport]'
      )
      if (scrollContainer) {
        scrollContainer.scrollTo({
          top: scrollContainer.scrollHeight,
          behavior: 'smooth',
        })
      }
    }
    setPrevLogsLength(logs.length)
  }, [logs.length, prevLogsLength])
  return (
    <div className='h-[400px]'>
      {isLoading && (
        <div className='text-muted-foreground text-sm'>Loading logs...</div>
      )}
      {error && (
        <div className='mb-2 flex items-center gap-2 rounded border border-red-200 bg-red-50 p-2 text-red-700'>
          <AlertTriangle className='h-4 w-4' />
          <span>{error.message}</span>
        </div>
      )}
      {!isLoading && !error && (
        <ScrollArea
          ref={scrollAreaRef}
          className='h-full rounded border p-2 text-xs'
          stickToBottom
        >
          {logs && logs.length > 0 ? (
            logs.map((log, index) => (
              <div
                key={log.id}
                className={`animate-in fade-in-0 slide-in-from-bottom-2 duration-300 ${
                  index >= prevLogsLength ? 'animate-in' : ''
                }`}
                style={{
                  animationDelay:
                    index >= prevLogsLength
                      ? `${(index - prevLogsLength) * 50}ms`
                      : '10ms',
                }}
              >
                <StructuredLogItem log={log} />
              </div>
            ))
          ) : (
            <span className='text-muted-foreground'>No logs to display.</span>
          )}
        </ScrollArea>
      )}
    </div>
  )
}
