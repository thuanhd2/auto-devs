import { formatDistanceToNow, format } from 'date-fns'
import type { Task } from '@/types/task'
import {
  Calendar,
  Clock,
  GitBranch,
  ExternalLink,
  Activity,
} from 'lucide-react'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

interface TaskMetadataProps {
  task: Task
  showGitInfo?: boolean
  showTimestamps?: boolean
  showStatusHistory?: boolean
}

export function TaskMetadata({
  task,
  showGitInfo = true,
  showTimestamps = true,
  showStatusHistory = false,
}: TaskMetadataProps) {
  const createdAgo = formatDistanceToNow(new Date(task.created_at), {
    addSuffix: true,
  })
  const updatedAgo = formatDistanceToNow(new Date(task.updated_at), {
    addSuffix: true,
  })

  return (
    <div className='space-y-4'>
      {/* Timestamps */}
      {showTimestamps && (
        <div className='grid grid-cols-2 gap-4'>
          <div>
            <h4 className='mb-2 text-sm font-medium text-gray-700'>Created</h4>
            <div className='flex items-center gap-2 text-sm text-gray-600'>
              <Calendar className='h-4 w-4' />
              <span>{createdAgo}</span>
            </div>
            <p className='mt-1 text-xs text-gray-500'>
              {format(new Date(task.created_at), 'PPpp')}
            </p>
          </div>

          <div>
            <h4 className='mb-2 text-sm font-medium text-gray-700'>
              Last Updated
            </h4>
            <div className='flex items-center gap-2 text-sm text-gray-600'>
              <Clock className='h-4 w-4' />
              <span>{updatedAgo}</span>
            </div>
            <p className='mt-1 text-xs text-gray-500'>
              {format(new Date(task.updated_at), 'PPpp')}
            </p>
          </div>
        </div>
      )}

      {/* Completion Date */}
      {task.completed_at && (
        <>
          <Separator />
          <div>
            <h4 className='mb-2 text-sm font-medium text-gray-700'>
              Completed
            </h4>
            <div className='flex items-center gap-2 text-sm text-gray-600'>
              <Calendar className='h-4 w-4' />
              <span>{format(new Date(task.completed_at), 'PPpp')}</span>
            </div>
          </div>
        </>
      )}

      {/* Git Information */}
      {showGitInfo && (task.branch_name || task.pr_url) && (
        <>
          <Separator />
          <div>
            <h4 className='mb-3 text-sm font-medium text-gray-700'>
              Git Information
            </h4>
            <div className='space-y-3'>
              {task.branch_name && (
                <div className='flex items-center gap-2 text-sm'>
                  <GitBranch className='h-4 w-4 text-gray-500' />
                  <span className='rounded bg-gray-100 px-2 py-1 font-mono text-gray-600'>
                    {task.branch_name}
                  </span>
                </div>
              )}

              {task.pr_url && (
                <div className='flex items-center gap-2 text-sm'>
                  <ExternalLink className='h-4 w-4 text-gray-500' />
                  <a
                    href={task.pr_url}
                    target='_blank'
                    rel='noopener noreferrer'
                    className='text-blue-600 hover:text-blue-700 hover:underline'
                  >
                    View Pull Request
                  </a>
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {/* Status History */}
      {showStatusHistory && (
        <>
          <Separator />
          <div>
            <h4 className='mb-3 text-sm font-medium text-gray-700'>
              Status History
            </h4>
            <div className='space-y-2'>
              <div className='flex items-center gap-2 text-sm'>
                <Activity className='h-4 w-4 text-gray-500' />
                <span>Current Status:</span>
                <Badge className={getStatusColor(task.status)}>
                  {getStatusTitle(task.status)}
                </Badge>
              </div>
              {/* In a real app, you would fetch and display status history here */}
              <p className='text-xs text-gray-500'>
                Status history tracking coming soon...
              </p>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
