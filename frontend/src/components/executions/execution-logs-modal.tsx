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
      <DialogContent className='w-full max-w-2xl'>
        <DialogHeader>
          <DialogTitle>Execution Logs</DialogTitle>
        </DialogHeader>
        <div className='max-h-[400px] min-h-[200px]'>
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
            <ScrollArea className='h-64 rounded border p-2 text-xs'>
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
  const object = JSON.parse(message)
  // TODO need show the content as humman-readable
  // dummy data can find in fake-cli/fake-output.log
  return <div>{object.type}</div>
}
