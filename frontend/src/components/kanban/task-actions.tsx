import { useState } from 'react'
import type { Task, TaskStatus } from '@/types/task'
import { Edit, Trash2, Copy, Play, ArrowUpDown, FolderOpen, Zap } from 'lucide-react'
import { tasksApi } from '@/lib/api/tasks'
import { Button } from '@/components/ui/button'
import { BranchSelectionDialog } from './branch-selection-dialog'
import { ChangeStatusDialog } from './change-status-dialog'
import { ImplementationConfirmationDialog } from './implementation-confirmation-dialog'

interface TaskActionsProps {
  task: Task
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onDuplicate?: (task: Task) => void
  onStartPlanning?: (taskId: string, branchName: string, aiType: string, autoImplement: boolean) => void
  onApprovePlanAndStartImplement?: (taskId: string, aiType: string) => void
  onChangeStatus?: (taskId: string, newStatus: TaskStatus) => Promise<void>
  onImplementDirect?: (taskId: string, branchName: string, aiType: string) => void
}

export function TaskActions({
  task,
  onEdit,
  onDelete,
  onDuplicate,
  onStartPlanning,
  onApprovePlanAndStartImplement,
  onChangeStatus,
  onImplementDirect,
}: TaskActionsProps) {
  const [showBranchDialog, setShowBranchDialog] = useState(false)
  const [showDirectImplementDialog, setShowDirectImplementDialog] = useState(false)
  const [showImplementationDialog, setShowImplementationDialog] =
    useState(false)
  const [showChangeStatusDialog, setShowChangeStatusDialog] = useState(false)
  const [isOpeningCursor, setIsOpeningCursor] = useState(false)

  const handleDelete = () => {
    onDelete?.(task.id)
  }

  const handleStartPlanning = () => {
    setShowBranchDialog(true)
  }

  const handleBranchSelected = (branchName: string, aiType: string, autoImplement: boolean) => {
    onStartPlanning?.(task.id, branchName, aiType, autoImplement)
  }

  const handleDirectImplementBranchSelected = (branchName: string, aiType: string, _autoImplement: boolean) => {
    onImplementDirect?.(task.id, branchName, aiType)
  }

  const handleApprovePlanAndStartImplement = () => {
    setShowImplementationDialog(true)
  }

  const handleImplementationConfirm = (aiType: string) => {
    onApprovePlanAndStartImplement?.(task.id, aiType)
  }

  const handleChangeStatus = async (newStatus: TaskStatus) => {
    if (onChangeStatus) {
      await onChangeStatus(task.id, newStatus)
    }
  }

  const handleOpenWithCursor = async () => {
    if (!task?.worktree_path) return

    try {
      setIsOpeningCursor(true)
      await tasksApi.openWithCursor(task.id)
      // Success feedback could be added here if needed
    } catch (error) {
      console.error('Failed to open with Cursor:', error)
      // Error handling could be added here
    } finally {
      setIsOpeningCursor(false)
    }
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

        {/* Implement Directly - Only show for TODO tasks */}
        {task.status === 'TODO' && onImplementDirect && (
          <Button
            variant='default'
            size='sm'
            onClick={() => setShowDirectImplementDialog(true)}
            title='Skip planning and implement directly'
            className='bg-orange-600 text-white hover:bg-orange-700'
          >
            <Zap className='mr-1 h-4 w-4' />
            Implement Directly
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

        {/* Open With Cursor button */}
        {task?.worktree_path && (
          <Button
            variant='outline'
            size='sm'
            onClick={handleOpenWithCursor}
            disabled={isOpeningCursor}
            title='Open task workspace with Cursor'
          >
            <FolderOpen className='mr-1 h-4 w-4' />
            {isOpeningCursor ? 'Opening...' : 'Open With Cursor'}
          </Button>
        )}

        {onEdit && (
          <Button variant='outline' size='sm' onClick={() => onEdit(task)}>
            <Edit className='h-4 w-4' /> Edit
          </Button>
        )}

        {onChangeStatus && (
          <Button
            variant='outline'
            size='sm'
            onClick={() => setShowChangeStatusDialog(true)}
            title='Change task status'
          >
            <ArrowUpDown className='h-4 w-4' /> Change Status
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

      {/* Direct Implement Branch Selection Dialog */}
      <BranchSelectionDialog
        open={showDirectImplementDialog}
        onOpenChange={setShowDirectImplementDialog}
        projectId={task.project_id}
        taskTitle={task.title}
        onBranchSelected={handleDirectImplementBranchSelected}
        mode='implementing'
      />

      {/* Implementation Confirmation Dialog */}
      <ImplementationConfirmationDialog
        open={showImplementationDialog}
        onOpenChange={setShowImplementationDialog}
        taskTitle={task.title}
        onConfirm={handleImplementationConfirm}
      />

      {/* Change Status Dialog */}
      <ChangeStatusDialog
        open={showChangeStatusDialog}
        onOpenChange={setShowChangeStatusDialog}
        task={task}
        onStatusChange={handleChangeStatus}
      />
    </>
  )
}
