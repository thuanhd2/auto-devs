import { useState } from 'react'
import type { Task } from '@/types/task'
import { Edit, Trash2, Copy, History, MoreVertical } from 'lucide-react'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { ConfirmDialog } from '../confirm-dialog'
import { TaskActions } from './task-actions'
import { TaskEditForm } from './task-edit-form'
import { TaskHistory } from './task-history'
import { TaskMetadata } from './task-metadata'

interface TaskDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  task: Task | null
  onEdit?: (task: Task) => void
  onDelete?: (taskId: string) => void
  onDuplicate?: (task: Task) => void
  onStatusChange?: (taskId: string, newStatus: Task['status']) => void
}

export function TaskDetailSheet({
  open,
  onOpenChange,
  task,
  onEdit,
  onDelete,
  onDuplicate,
  onStatusChange,
}: TaskDetailSheetProps) {
  const [showEditForm, setShowEditForm] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [showHistory, setShowHistory] = useState(false)

  if (!task) return null

  const statusTitle = getStatusTitle(task.status)
  const statusColor = getStatusColor(task.status)

  const handleEdit = () => {
    setShowEditForm(true)
  }

  const handleDelete = () => {
    setShowDeleteConfirm(true)
  }

  const handleDuplicate = () => {
    onDuplicate?.(task)
  }

  const handleStatusChange = (taskId: string, newStatus: Task['status']) => {
    onStatusChange?.(taskId, newStatus)
  }

  const confirmDelete = () => {
    onDelete?.(task.id)
    setShowDeleteConfirm(false)
    onOpenChange(false)
  }

  const handleEditSave = (updatedTask: Task) => {
    onEdit?.(updatedTask)
    setShowEditForm(false)
  }

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className='w-[600px] overflow-y-auto sm:w-[700px]'>
          <SheetHeader className='pb-4'>
            <div className='flex items-start justify-between gap-4'>
              <div className='flex-1'>
                <SheetTitle className='mb-2 text-xl font-semibold'>
                  {task.title}
                </SheetTitle>
                <div className='flex items-center gap-2'>
                  <Badge className={statusColor}>{statusTitle}</Badge>
                </div>
              </div>

              <div className='flex items-center gap-1'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => setShowHistory(true)}
                >
                  <History className='h-4 w-4' />
                </Button>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant='outline' size='sm'>
                      <MoreVertical className='h-4 w-4' />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align='end'>
                    <DropdownMenuItem onClick={handleEdit}>
                      <Edit className='mr-2 h-4 w-4' />
                      Edit Task
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={handleDuplicate}>
                      <Copy className='mr-2 h-4 w-4' />
                      Duplicate Task
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onClick={handleDelete}
                      className='text-red-600 focus:text-red-600'
                    >
                      <Trash2 className='mr-2 h-4 w-4' />
                      Delete Task
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
          </SheetHeader>

          <div className='space-y-6 px-4 pb-6'>
            {/* Description */}
            {task.description && (
              <div>
                <h4 className='mb-2 text-sm font-medium text-gray-700'>
                  Description
                </h4>
                <p className='rounded border bg-gray-50 p-3 text-sm whitespace-pre-wrap text-gray-600'>
                  {task.description}
                </p>
              </div>
            )}

            {/* Plan */}
            {task.plan && (
              <div>
                <h4 className='mb-2 text-sm font-medium text-gray-700'>Plan</h4>
                <div className='rounded border bg-blue-50 p-3 text-sm whitespace-pre-wrap text-gray-600'>
                  {task.plan}
                </div>
              </div>
            )}

            <Separator />

            {/* Status Actions */}
            <div>
              <h4 className='mb-3 text-sm font-medium text-gray-700'>
                Actions
              </h4>
              <TaskActions
                task={task}
                onEdit={handleEdit}
                onDelete={handleDelete}
                onDuplicate={handleDuplicate}
                onStatusChange={handleStatusChange}
                onViewHistory={() => setShowHistory(true)}
                showStatusActions={true}
                showGitActions={true}
              />
            </div>

            <Separator />

            {/* Metadata */}
            <TaskMetadata
              task={task}
              showGitInfo={true}
              showTimestamps={true}
              showStatusHistory={false}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* Edit Form Modal */}
      {showEditForm && (
        <TaskEditForm
          open={showEditForm}
          onOpenChange={setShowEditForm}
          task={task}
          onSave={handleEditSave}
        />
      )}

      {/* History Modal */}
      {showHistory && (
        <TaskHistory
          open={showHistory}
          onOpenChange={setShowHistory}
          taskId={task.id}
        />
      )}

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
