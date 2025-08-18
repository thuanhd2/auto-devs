import { useNavigate, useParams } from '@tanstack/react-router'
import { useState } from 'react'
import type { Task } from '@/types/task'
import { ExternalLink, FolderOpen } from 'lucide-react'
import { tasksApi } from '@/lib/api/tasks'
import { getStatusColor, getStatusTitle } from '@/lib/kanban'
import { useTaskExecutions } from '@/hooks/use-executions'
import { usePullRequestByTask, useCreatePullRequest } from '@/hooks/use-pull-requests'
import { useTaskDiff } from '@/hooks/use-tasks'
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
import { Diff, parseDiff, Hunk } from 'react-diff-view'
import 'react-diff-view/style/index.css'
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

// Helper function to parse git diff string
function parseDiffString(diffString: string) {
  try {
    // React-diff-view expects unified diff format
    // If the diff is already in the correct format, parse it directly
    const files = parseDiff(diffString)
    return files
  } catch (error) {
    console.error('Error parsing diff:', error)
    return []
  }
}

// CodeChanges component for the code changes tab
function CodeChanges({ taskId, task }: { taskId: string; task?: Task }) {
  const { data: pullRequest, isLoading: isPRLoading } =
    usePullRequestByTask(taskId)
  const {
    data: diff,
    isLoading: isDiffLoading,
    error: diffError,
  } = useTaskDiff(taskId)
  const [isOpeningCursor, setIsOpeningCursor] = useState(false)
  const createPRMutation = useCreatePullRequest()

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

  const handleCreatePR = async () => {
    try {
      await createPRMutation.mutateAsync(taskId)
      // Success feedback could be added here if needed
    } catch (error) {
      console.error('Failed to create pull request:', error)
      // Error handling could be added here
    }
  }

  const isLoading = isPRLoading || isDiffLoading

  if (isLoading) {
    return (
      <div className='flex items-center justify-center p-4'>
        <div className='text-muted-foreground text-sm'>
          Loading code changes...
        </div>
      </div>
    )
  }

  return (
    <div className='space-y-4'>
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

      {/* Pull Request Section */}
      <div className='space-y-2'>
        {pullRequest ? (
          <>
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
          </>
        ) : (
          <Button
            variant='outline'
            size='sm'
            className='w-fit'
            onClick={handleCreatePR}
            disabled={createPRMutation.isPending}
          >
            <ExternalLink className='mr-2 h-4 w-4' />
            {createPRMutation.isPending ? 'Creating...' : 'Create Pull Request'}
          </Button>
        )}
      </div>

      {/* Diff Display */}
      <div className='space-y-2'>
        <h5 className='text-muted-foreground text-xs font-medium'>
          Diff (Base Branch vs Task Branch)
        </h5>
        {diffError ? (
          <div className='rounded bg-red-50 p-2 text-sm text-red-600'>
            Error loading diff: {diffError.message}
          </div>
        ) : diff === undefined ? (
          <div className='text-muted-foreground bg-muted/50 rounded p-2 text-sm'>
            Loading diff...
          </div>
        ) : diff === '' || diff === 'No code changes' ? (
          <div className='text-muted-foreground bg-muted/50 rounded p-2 text-sm'>
            No code changes
          </div>
) : (
          <DiffViewer diffString={diff} />
        )}
      </div>
    </div>
  )
}

// DiffViewer component to display formatted git diff
function DiffViewer({ diffString }: { diffString: string }) {
  const files = parseDiffString(diffString)
  
  if (files.length === 0) {
    return (
      <div className='max-h-96 overflow-auto rounded-md border'>
        <pre className='bg-slate-50 p-4 font-mono text-xs whitespace-pre-wrap'>
          {diffString}
        </pre>
      </div>
    )
  }

  return (
    <div className='max-h-96 overflow-auto rounded-md border bg-white'>
      {files.map((file, index) => (
        <div key={index} className='border-b last:border-b-0'>
          <div className='bg-gray-50 px-4 py-2 text-sm font-medium text-gray-700 border-b'>
            {file.oldPath === file.newPath ? file.newPath : `${file.oldPath} → ${file.newPath}`}
          </div>
          <Diff
            key={file.oldPath + file.newPath}
            viewType='unified'
            diffType={file.type}
            hunks={file.hunks}
            className='text-xs'
          >
            {(hunks) =>
              hunks.map((hunk) => (
                <Hunk key={hunk.content} hunk={hunk} />
              ))
            }
          </Diff>
        </div>
      ))}
    </div>
  )
}
