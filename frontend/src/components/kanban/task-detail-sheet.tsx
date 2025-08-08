import { useState } from 'react'
import type { Task } from '@/types/task'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { PlanReview } from '../planning'
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
  onStartPlanning?: (taskId: string, branchName: string) => void
  onApprovePlanAndStartImplement?: (taskId: string) => void
}

export function TaskDetailSheet({
  open,
  onOpenChange,
  task,
  onEdit,
  onDelete,
  onDuplicate,
  onStatusChange,
  onStartPlanning,
  onApprovePlanAndStartImplement,
}: TaskDetailSheetProps) {
  const [showEditForm, setShowEditForm] = useState(false)
  const [showHistory, setShowHistory] = useState(false)

  if (!task) return null

  const statusTitle = getStatusTitle(task.status)
  const statusColor = getStatusColor(task.status)

  const handleEdit = () => {
    setShowEditForm(true)
  }

  const handleDuplicate = () => {
    onDuplicate?.(task)
  }

  const handleStatusChange = (taskId: string, newStatus: Task['status']) => {
    onStatusChange?.(taskId, newStatus)
  }

  const handleEditSave = (updatedTask: Task) => {
    onEdit?.(updatedTask)
    setShowEditForm(false)
  }

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className='overflow-y-auto sm:w-[400px] sm:max-w-[400px] lg:w-[800px] lg:max-w-none'>
          <SheetHeader className='pb-4'>
            <div className='flex items-start justify-between gap-4'>
              <div className='flex-1'>
                <SheetTitle className='mb-2 text-xl font-semibold'>
                  {task.title}
                </SheetTitle>
                <div className='flex items-center gap-2'>
                  <Badge className={statusColor} variant='outline'>
                    {statusTitle}
                  </Badge>
                </div>
              </div>
            </div>
          </SheetHeader>

          <div className='space-y-6 px-4 pb-6'>
            {/* Description */}
            {task.description && (
              <div>
                <h4 className='mb-2 text-sm font-medium'>Description</h4>
                <p className='rounded border p-3 text-sm whitespace-pre-wrap'>
                  {task.description}
                </p>
              </div>
            )}

            {/* Status Actions */}
            <div>
              <h4 className='mb-3 text-sm font-medium'>Actions</h4>
              <TaskActions
                task={task}
                onEdit={handleEdit}
                onDelete={onDelete}
                onDuplicate={handleDuplicate}
                onStatusChange={handleStatusChange}
                onStartPlanning={onStartPlanning}
                onApprovePlanAndStartImplement={onApprovePlanAndStartImplement}
                // onViewHistory={() => setShowHistory(true)}
              />
            </div>

            <Separator />

            {/* Tabs for Plan Review and Metadata */}
            <Tabs defaultValue="plan-review" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="plan-review">Plan Review</TabsTrigger>
                <TabsTrigger value="metadata">Metadata</TabsTrigger>
              </TabsList>
              
              <TabsContent value="plan-review" className="mt-4">
                <PlanReview
                  task={task}
                  onPlanUpdate={onEdit}
                  onStatusChange={onStatusChange}
                />
              </TabsContent>
              
              <TabsContent value="metadata" className="mt-4">
                <TaskMetadata
                  task={task}
                  showGitInfo={true}
                  showTimestamps={true}
                  showStatusHistory={true}
                />
              </TabsContent>
            </Tabs>
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
    </>
  )
}
