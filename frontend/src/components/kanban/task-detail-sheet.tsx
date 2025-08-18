import { useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import type { Task } from '@/types/task'
import { ExternalLink, FolderOpen } from 'lucide-react'
import { tasksApi } from '@/lib/api/tasks'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { useTaskExecutions } from '@/hooks/use-executions'
import { usePullRequestByTask } from '@/hooks/use-pull-requests'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { ExecutionList } from '../executions'
import { PlanReview } from '../planning'
import { TaskActions } from './task-actions'
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
  onStartPlanning?: (taskId: string, branchName: string, aiType: string) => void
  onApprovePlanAndStartImplement?: (taskId: string, aiType: string) => void
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
  const navigate = useNavigate()
  const params = useParams({ strict: false }) as { projectId?: string }
  const [showHistory, setShowHistory] = useState(false)

  // Handle sheet close and URL cleanup
  const handleOpenChange = (isOpen: boolean) => {
    onOpenChange(isOpen)

    if (!isOpen && params.projectId) {
      // Navigate back to project without task ID
      navigate({
        to: '/projects/$projectId',
        params: { projectId: params.projectId },
        replace: true,
      })
    }
  }

  if (!task) return null

  const statusTitle = getStatusTitle(task.status)
  const statusColor = getStatusColor(task.status)

  const handleEdit = () => {
    onEdit?.(task)
  }

  const handleDuplicate = () => {
    onDuplicate?.(task)
  }
  return (
    <>
      <Sheet open={open} onOpenChange={handleOpenChange}>
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
          <SheetDescription className='px-4 text-sm whitespace-pre-wrap'>
            {/* Description */}
            {task.description}
          </SheetDescription>

          <div className='space-y-6 px-4 pb-6'>
            {/* Status Actions */}
            <div>
              <h4 className='mb-3 text-sm font-medium'>Actions</h4>
              <TaskActions
                task={task}
                onEdit={handleEdit}
                onDelete={onDelete}
                onDuplicate={handleDuplicate}
                onStartPlanning={onStartPlanning}
                onApprovePlanAndStartImplement={onApprovePlanAndStartImplement}
                onChangeStatus={
                  onStatusChange
                    ? async (taskId: string, newStatus: Task['status']) => {
                        onStatusChange(taskId, newStatus)
                      }
                    : undefined
                }
              />
            </div>

            <Separator />

            {/* Tabs for Plan Review, Code Changes, Executions, and Metadata */}
            <Tabs defaultValue='executions' className='w-full'>
              <TabsList className='grid w-full grid-cols-4'>
                <TabsTrigger value='executions'>Executions</TabsTrigger>
                <TabsTrigger value='plan-review'>Plan Review</TabsTrigger>
                <TabsTrigger value='code-changes'>Code Changes</TabsTrigger>
                <TabsTrigger value='metadata'>Metadata</TabsTrigger>
              </TabsList>

              <TabsContent value='plan-review' className='mt-4'>
                <PlanReview task={task} />
              </TabsContent>

              <TabsContent value='code-changes' className='mt-4'>
                <CodeChanges taskId={task.id} task={task} />
              </TabsContent>

              <TabsContent value='executions' className='mt-4'>
                <TaskExecutions taskId={task.id} />
              </TabsContent>

              <TabsContent value='metadata' className='mt-4'>
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

// TaskExecutions component for the executions tab
function TaskExecutions({ taskId }: { taskId: string }) {
  const {
    data: executionsData,
    isLoading,
    error,
    refetch,
  } = useTaskExecutions(taskId, {
    page: 1,
    page_size: 20,
    order_by: 'started_at',
    order_dir: 'desc',
  })

  const executions = executionsData?.data || []

  return (
    <>
      <ExecutionList
        executions={executions}
        loading={isLoading}
        error={error?.message}
        onRefresh={refetch}
        compact={true}
        expandable={true}
        showFilters={false}
      />
    </>
  )
}

// CodeChanges component for the code changes tab
function CodeChanges({ taskId, task }: { taskId: string; task?: Task }) {
  const { data: pullRequest, isLoading, error } = usePullRequestByTask(taskId)
  const [isOpeningCursor, setIsOpeningCursor] = useState(false)

  const handleOpenWithCursor = async () => {
    if (!task?.worktree_path) return

    try {
      setIsOpeningCursor(true)
      await tasksApi.openWithCursor(taskId)
      // Success feedback could be added here if needed
    } catch (error) {
      console.error('Failed to open with Cursor:', error)
      // Error handling could be added here
    } finally {
      setIsOpeningCursor(false)
    }
  }

  if (isLoading) {
    return (
      <div className='flex items-center justify-center p-4'>
        <div className='text-muted-foreground text-sm'>
          Loading pull request...
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className='flex items-center justify-center p-4'>
        <div className='text-sm text-red-600'>Error loading pull request</div>
      </div>
    )
  }

  if (!pullRequest) {
    return (
      <div className='space-y-3'>
        <div className='flex items-center gap-2'>
          <h4 className='text-sm font-medium'>Code Changes</h4>
        </div>

        {/* Open with Cursor button */}
        {task?.worktree_path && (
          <Button
            variant='outline'
            size='sm'
            className='w-fit'
            onClick={handleOpenWithCursor}
            disabled={isOpeningCursor}
          >
            <FolderOpen className='mr-2 h-4 w-4' />
            {isOpeningCursor ? 'Opening...' : 'Open With Cursor'}
          </Button>
        )}

        <div className='text-muted-foreground text-sm'>
          No pull request created yet
        </div>
      </div>
    )
  }

  return (
    <div className='space-y-3'>
      <div className='flex items-center gap-2'>
        <h4 className='text-sm font-medium'>Code Changes</h4>
      </div>

      {/* Open with Cursor button */}
      {task?.worktree_path && (
        <Button
          variant='outline'
          size='sm'
          className='w-fit'
          onClick={handleOpenWithCursor}
          disabled={isOpeningCursor}
        >
          <FolderOpen className='mr-2 h-4 w-4' />
          {isOpeningCursor ? 'Opening...' : 'Open With Cursor'}
        </Button>
      )}

      <Button
        variant='outline'
        size='sm'
        className='w-fit'
        onClick={() => window.open(pullRequest.github_url, '_blank')}
      >
        <ExternalLink className='mr-2 h-4 w-4' />
        View Pull Request
      </Button>
      <div className='text-muted-foreground text-xs'>
        #{pullRequest.github_pr_number} - {pullRequest.title}
      </div>
    </div>
  )
}
