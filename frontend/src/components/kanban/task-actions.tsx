import { useState } from 'react'
import type { Task, TaskStatus } from '@/types/task'
import {
  Edit,
  Trash2,
  Copy,
  MoreVertical,
  History,
  ExternalLink,
  GitBranch,
} from 'lucide-react'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ConfirmDialog } from '../confirm-dialog'

interface TaskActionsProps {
  task: Task
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onDuplicate?: (task: Task) => void
  onStatusChange?: (taskId: string, newStatus: TaskStatus) => void
  onViewHistory?: () => void
  showStatusActions?: boolean
  showGitActions?: boolean
}

export function TaskActions({
  task,
  onEdit,
  onDelete,
  onDuplicate,
  onStatusChange,
  onViewHistory,
  showStatusActions = true,
  showGitActions = true,
}: TaskActionsProps) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const handleDelete = () => {
    setShowDeleteConfirm(true)
  }

  const confirmDelete = () => {
    onDelete?.(task.id)
    setShowDeleteConfirm(false)
  }

  const handleStatusChange = (newStatus: TaskStatus) => {
    onStatusChange?.(task.id, newStatus)
  }

  const handleGitAction = (action: 'branch' | 'pr') => {
    if (action === 'branch' && task.branch_name) {
      // Copy branch name to clipboard
      navigator.clipboard.writeText(task.branch_name)
    } else if (action === 'pr' && task.pr_url) {
      // Open PR in new tab
      window.open(task.pr_url, '_blank')
    }
  }

  return (
    <>
      <div className='flex flex-wrap items-center gap-2'>
        {/* Git Actions */}
        {(task.branch_name || task.pr_url) && (
          <div className='flex items-center gap-1'>
            {task.branch_name && (
              <Button
                variant='outline'
                size='sm'
                onClick={() => handleGitAction('branch')}
                title='Copy branch name'
              >
                <GitBranch className='h-4 w-4' />
              </Button>
            )}

            {task.pr_url && (
              <Button
                variant='outline'
                size='sm'
                onClick={() => handleGitAction('pr')}
                title='Open Pull Request'
              >
                <ExternalLink className='h-4 w-4' />
              </Button>
            )}
          </div>
        )}

        {/* History Button */}
        {onViewHistory && (
          <Button
            variant='outline'
            size='sm'
            onClick={onViewHistory}
            title='View History'
          >
            <History className='h-4 w-4' /> View History
          </Button>
        )}
        {onEdit && (
          <Button variant='outline' size='sm' onClick={() => onEdit(task)}>
            <Edit className='h-4 w-4' /> Edit
          </Button>
        )}

        {onDuplicate && (
          <Button variant='outline' size='sm' onClick={() => onDuplicate(task)}>
            <Copy className='h-4 w-4' /> Duplicate
          </Button>
        )}

        {onDelete && (
          <Button variant='destructive' size='sm' onClick={handleDelete}>
            <Trash2 className='h-4 w-4' /> Delete
          </Button>
        )}
      </div>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        title='Delete Task'
        description='Are you sure you want to delete this task? This action cannot be undone.'
        onConfirm={confirmDelete}
        confirmText='Delete'
        variant='destructive'
      />
    </>
  )
}
