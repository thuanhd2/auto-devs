import { useState, useEffect } from 'react'
import { format } from 'date-fns'
import { Task } from '@/types/kanban'
import { Clock, ArrowRight, Circle } from 'lucide-react'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

// Mock data for task history - in real app this would come from API
interface TaskHistoryItem {
  id: string
  action: string
  field?: string
  oldValue?: string
  newValue?: string
  timestamp: string
  user?: string
}

interface TaskHistoryProps {
  task: Task
  onClose: () => void
}

export function TaskHistory({ task, onClose }: TaskHistoryProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [history, setHistory] = useState<TaskHistoryItem[]>([])

  useEffect(() => {
    const fetchHistory = async () => {
      setIsLoading(true)
      try {
        // TODO: Implement API call to fetch task history
        const mockHistory: TaskHistoryItem[] = [
          {
            id: '1',
            action: 'created',
            timestamp: new Date(
              Date.now() - 7 * 24 * 60 * 60 * 1000
            ).toISOString(),
            user: 'John Doe',
          },
          {
            id: '2',
            action: 'status_changed',
            field: 'status',
            oldValue: 'TODO',
            newValue: 'PLANNING',
            timestamp: new Date(
              Date.now() - 6 * 24 * 60 * 60 * 1000
            ).toISOString(),
            user: 'Jane Smith',
          },
        ]
        setHistory(mockHistory)
      } catch {
        // Handle error silently
      } finally {
        setIsLoading(false)
      }
    }

    fetchHistory()
  }, [])

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'created':
        return <Circle className='h-4 w-4 text-green-500' />
      case 'status_changed':
        return <ArrowRight className='h-4 w-4 text-blue-500' />
      case 'updated':
        return <Clock className='h-4 w-4 text-gray-500' />
      default:
        return <Circle className='h-4 w-4 text-gray-400' />
    }
  }

  const getActionText = (item: TaskHistoryItem) => {
    switch (item.action) {
      case 'created':
        return 'Task created'
      case 'status_changed':
        return `Status changed from ${getStatusTitle(item.oldValue as any)} to ${getStatusTitle(item.newValue as any)}`
      case 'updated':
        return `${item.field?.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase())} updated`
      default:
        return item.action
    }
  }

  const getStatusBadge = (status: string) => {
    return (
      <Badge className={getStatusColor(status as any)} variant='outline'>
        {getStatusTitle(status as any)}
      </Badge>
    )
  }

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className='max-h-[80vh] overflow-y-auto sm:max-w-[600px]'>
        <DialogHeader>
          <DialogTitle>Task History</DialogTitle>
          <DialogDescription>
            Timeline of changes and updates to this task.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          {history.map((item, index) => (
            <div key={item.id} className='relative'>
              {/* Timeline line */}
              {index < history.length - 1 && (
                <div className='absolute top-8 left-6 h-12 w-0.5 bg-gray-200' />
              )}

              <div className='flex items-start gap-4'>
                {/* Icon */}
                <div className='flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full border-2 border-gray-200'>
                  {getActionIcon(item.action)}
                </div>

                {/* Content */}
                <div className='min-w-0 flex-1'>
                  <div className='mb-1 flex items-center gap-2'>
                    <span className='text-sm font-medium'>
                      {getActionText(item)}
                    </span>
                    {item.user && (
                      <span className='text-xs text-gray-500'>
                        by {item.user}
                      </span>
                    )}
                  </div>

                  {/* Status change details */}
                  {item.action === 'status_changed' &&
                    item.oldValue &&
                    item.newValue && (
                      <div className='mb-2 flex items-center gap-2'>
                        {getStatusBadge(item.oldValue)}
                        <ArrowRight className='h-3 w-3 text-gray-400' />
                        {getStatusBadge(item.newValue)}
                      </div>
                    )}

                  {/* Field update details */}
                  {item.action === 'updated' && item.field && (
                    <div className='mb-2 text-sm text-gray-600'>
                      {item.field === 'branch_name' && item.newValue && (
                        <div className='rounded border bg-gray-50 p-2'>
                          <span className='font-mono text-xs'>
                            {item.newValue}
                          </span>
                        </div>
                      )}
                      {item.field === 'pr_url' && item.newValue && (
                        <div className='rounded border bg-gray-50 p-2'>
                          <a
                            href={item.newValue}
                            target='_blank'
                            rel='noopener noreferrer'
                            className='text-xs text-blue-600 hover:text-blue-700'
                          >
                            {item.newValue}
                          </a>
                        </div>
                      )}
                    </div>
                  )}

                  <div className='flex items-center gap-1 text-xs text-gray-500'>
                    <Clock className='h-3 w-3' />
                    <span>{format(new Date(item.timestamp), 'PPpp')}</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {history.length === 0 && (
          <div className='py-8 text-center text-gray-500'>
            <Clock className='mx-auto mb-4 h-12 w-12 text-gray-300' />
            <p>No history available for this task.</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
