import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import type { Task } from '@/types/task'
import { Calendar, GitBranch, ExternalLink } from 'lucide-react'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { GitStatusBadge } from './git-status-badge'

interface TaskCardProps {
  task: Task
  isDragging?: boolean
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onViewDetails?: (task: Task) => void
}

export function TaskCard({
  task,
  isDragging = false,
  onEdit,
  onDelete,
  onViewDetails,
}: TaskCardProps) {
  const statusColor = getStatusColor(task.status)
  const statusTitle = getStatusTitle(task.status)
  const updatedAgo = formatDistanceToNow(new Date(task.updated_at), {
    addSuffix: true,
  })

  return (
    <Card
      className={`cursor-pointer transition-all duration-200 hover:shadow-md ${isDragging ? 'rotate-2 opacity-50 shadow-lg' : ''} ${statusColor} border-l-4`}
      onClick={onViewDetails}
    >
      <CardHeader className='pb-2'>
        <div className='flex items-start justify-between gap-2'>
          <div className='min-w-0 flex-1'>
            <h3 className='line-clamp-2 text-sm font-medium text-gray-900'>
              {task.title}
            </h3>
            <div className='mt-1 flex items-center gap-1'>
              <Badge variant='secondary' className='text-xs'>
                {statusTitle}
              </Badge>
              {task.git_info && (
                <GitStatusBadge
                  status={task.git_info.status}
                  branchName={task.git_info.branch_name}
                  variant='compact'
                  className='ml-1'
                />
              )}
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className='pt-0'>
        {task.description && (
          <p className='mb-3 line-clamp-2 text-xs text-gray-600'>
            {task.description}
          </p>
        )}

        <div className='flex items-center justify-between text-xs text-gray-500'>
          <div className='flex items-center gap-1'>
            <Calendar className='h-3 w-3' />
            <span>{updatedAgo}</span>
          </div>

          <div className='flex items-center gap-2'>
            {(task.git_info?.branch_name || task.branch_name) && (
              <div className='flex items-center gap-1'>
                <GitBranch className='h-3 w-3' />
                <span className='max-w-16 truncate'>
                  {task.git_info?.branch_name || task.branch_name}
                </span>
              </div>
            )}

            {task.git_info?.worktree_path && (
              <div
                className='flex items-center gap-1 text-xs text-blue-600'
                title={task.git_info.worktree_path}
              >
                <span className='max-w-20 truncate'>
                  {task.git_info.worktree_path.split('/').pop()}
                </span>
              </div>
            )}

            {(task.git_info?.pr_url || task.pr_url) && (
              <a
                href={task.git_info?.pr_url || task.pr_url}
                target='_blank'
                rel='noopener noreferrer'
                className='flex items-center gap-1 transition-colors hover:text-blue-600'
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink className='h-3 w-3' />
                <span>PR</span>
              </a>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
