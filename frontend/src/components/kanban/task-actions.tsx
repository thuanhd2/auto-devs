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
  Play,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '../confirm-dialog'
import { BranchSelectionDialog } from './branch-selection-dialog'

interface TaskActionsProps {
  task: Task
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onDuplicate?: (task: Task) => void
  onStatusChange?: (taskId: string, newStatus: TaskStatus) => void
  onViewHistory?: () => void
  onStartPlanning?: (taskId: string, branchName: string, aiType: string) => void
  onApprovePlanAndStartImplement?: (taskId: string) => void
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
  onStartPlanning,
  onApprovePlanAndStartImplement,
  showStatusActions = true,
  showGitActions = true,
}: TaskActionsProps) {
  const [showBranchDialog, setShowBranchDialog] = useState(false)

  const handleDelete = () => {
    onDelete?.(task.id)
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

  const handleStartPlanning = () => {
    setShowBranchDialog(true)
  }

  const handleBranchSelected = (branchName: string, aiType: string) => {
    onStartPlanning?.(task.id, branchName, aiType)
  }

  const handleApprovePlanAndStartImplement = () => {
    onApprovePlanAndStartImplement?.(task.id)
  }

  return (
    <>
      <div className='flex flex-wrap items-center gap-2'>
        {/* Start Planning Action - Only show for TODO tasks */}
        {task.status === 'TODO' && onStartPlanning && (
          <Button
            variant='default'
            size='sm'
            onClick={handleStartPlanning}
            title='Start planning for this task'
            className='bg-blue-600 text-white hover:bg-blue-700'
          >
            <Play className='mr-1 h-4 w-4' />
            Start Planning
          </Button>
        )}

        {/* Approve Plan and Start Implement Action - Only show for TODO tasks */}

        {task.status === 'PLAN_REVIEWING' && onApprovePlanAndStartImplement && (
          <Button
            variant='default'
            size='sm'
            onClick={handleApprovePlanAndStartImplement}
            title='Approve plan and start implementing'
            className='bg-green-600 text-white hover:bg-green-700'
          >
            <Play className='mr-1 h-4 w-4' />
            Approve Plan and Start Implement
          </Button>
        )}
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

      {/* Branch Selection Dialog */}
      <BranchSelectionDialog
        open={showBranchDialog}
        onOpenChange={setShowBranchDialog}
        projectId={task.project_id}
        taskTitle={task.title}
        onBranchSelected={handleBranchSelected}
      />
    </>
  )
}
