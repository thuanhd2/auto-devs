import { useState } from 'react'
import type { Task, TaskGitStatus } from '@/types/task'
import {
  GitBranch,
  FolderOpen,
  Trash2,
  Loader2,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfirmDialog } from '../confirm-dialog'
import { GitStatusBadge } from './git-status-badge'

interface GitOperationControlsProps {
  task: Task
  onWorktreeCreate?: (taskId: string) => Promise<void>
  onWorktreeOpen?: (taskId: string, worktreePath: string) => void
  onWorktreeCleanup?: (taskId: string) => Promise<void>
  onRefreshStatus?: (taskId: string) => Promise<void>
  disabled?: boolean
}

export function GitOperationControls({
  task,
  onWorktreeCreate,
  onWorktreeOpen,
  onWorktreeCleanup,
  onRefreshStatus,
  disabled = false,
}: GitOperationControlsProps) {
  const [loading, setLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [showCleanupConfirm, setShowCleanupConfirm] = useState(false)

  const gitStatus = task.git_info?.status || 'NO_GIT'
  const canCreateWorktree = ['NO_GIT', 'WORKTREE_ERROR'].includes(gitStatus)
  const canOpenWorktree =
    task.git_info?.worktree_path &&
    [
      'WORKTREE_CREATED',
      'BRANCH_CREATED',
      'CHANGES_PENDING',
      'CHANGES_STAGED',
      'CHANGES_COMMITTED',
    ].includes(gitStatus)
  const canCleanupWorktree =
    task.git_info?.worktree_path &&
    !['NO_GIT', 'WORKTREE_PENDING'].includes(gitStatus)

  const handleOperation = async (
    operation: string,
    handler?: () => Promise<void>
  ) => {
    if (!handler || disabled) return

    setLoading(operation)
    setError(null)

    try {
      await handler()
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${operation}`)
    } finally {
      setLoading(null)
    }
  }

  const handleCreateWorktree = async () => {
    await handleOperation('create', () => onWorktreeCreate?.(task.id))
  }

  const handleOpenWorktree = () => {
    if (task.git_info?.worktree_path) {
      onWorktreeOpen?.(task.id, task.git_info.worktree_path)
    }
  }

  const handleCleanupWorktree = async () => {
    await handleOperation('cleanup', () => onWorktreeCleanup?.(task.id))
    setShowCleanupConfirm(false)
  }

  const handleRefreshStatus = async () => {
    await handleOperation('refresh', () => onRefreshStatus?.(task.id))
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <GitBranch className='h-5 w-5' />
            Git Operations
          </CardTitle>
          <CardDescription>
            Manage Git worktree and branch operations for this task
          </CardDescription>
        </CardHeader>

        <CardContent className='space-y-4'>
          {/* Current Status */}
          <div className='flex items-center justify-between'>
            <span className='text-sm font-medium'>Current Status:</span>
            <div className='flex items-center gap-2'>
              <GitStatusBadge
                status={gitStatus}
                branchName={task.git_info?.branch_name}
              />
              <Button
                variant='ghost'
                size='sm'
                onClick={handleRefreshStatus}
                disabled={loading === 'refresh' || disabled}
                className='h-6 w-6 p-0'
                title='Refresh Git status'
              >
                {loading === 'refresh' ? (
                  <Loader2 className='h-3 w-3 animate-spin' />
                ) : (
                  <CheckCircle2 className='h-3 w-3' />
                )}
              </Button>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <Alert variant='destructive'>
              <AlertCircle className='h-4 w-4' />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Operation Buttons */}
          <div className='flex flex-wrap gap-2'>
            {/* Create Worktree */}
            {canCreateWorktree && (
              <Button
                onClick={handleCreateWorktree}
                disabled={loading === 'create' || disabled}
                size='sm'
                className='flex items-center gap-2'
              >
                {loading === 'create' ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <GitBranch className='h-4 w-4' />
                )}
                Create Worktree
              </Button>
            )}

            {/* Open Worktree */}
            {canOpenWorktree && (
              <Button
                onClick={handleOpenWorktree}
                disabled={disabled}
                variant='outline'
                size='sm'
                className='flex items-center gap-2'
              >
                <FolderOpen className='h-4 w-4' />
                Open Worktree
              </Button>
            )}

            {/* Cleanup Worktree */}
            {canCleanupWorktree && (
              <Button
                onClick={() => setShowCleanupConfirm(true)}
                disabled={loading === 'cleanup' || disabled}
                variant='outline'
                size='sm'
                className='flex items-center gap-2 text-red-600 hover:text-red-700'
              >
                {loading === 'cleanup' ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Trash2 className='h-4 w-4' />
                )}
                Cleanup Worktree
              </Button>
            )}
          </div>

          {/* Additional Information */}
          {task.git_info && (
            <div className='border-t pt-2'>
              <div className='space-y-2 text-xs text-gray-600'>
                {task.git_info.worktree_path && (
                  <div className='flex items-center gap-1'>
                    <FolderOpen className='h-3 w-3' />
                    <span
                      className='truncate font-mono'
                      title={task.git_info.worktree_path}
                    >
                      {task.git_info.worktree_path}
                    </span>
                  </div>
                )}

                {task.git_info.last_sync && (
                  <div className='text-gray-500'>
                    Last sync:{' '}
                    {new Date(task.git_info.last_sync).toLocaleString()}
                  </div>
                )}

                {(task.git_info.commits_ahead ||
                  task.git_info.commits_behind) && (
                  <div className='flex gap-2'>
                    {task.git_info.commits_ahead && (
                      <Badge variant='secondary' className='text-xs'>
                        +{task.git_info.commits_ahead} ahead
                      </Badge>
                    )}
                    {task.git_info.commits_behind && (
                      <Badge variant='secondary' className='text-xs'>
                        -{task.git_info.commits_behind} behind
                      </Badge>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Cleanup Confirmation Dialog */}
      <ConfirmDialog
        open={showCleanupConfirm}
        onOpenChange={setShowCleanupConfirm}
        title='Cleanup Worktree'
        description='This will remove the Git worktree and all local changes. This action cannot be undone. Are you sure you want to continue?'
        onConfirm={handleCleanupWorktree}
        confirmText='Cleanup'
        variant='destructive'
      />
    </>
  )
}

export function getOperationRecommendations(
  gitStatus: TaskGitStatus
): string[] {
  const recommendations: Record<TaskGitStatus, string[]> = {
    NO_GIT: [
      'Create a Git worktree to start working on this task',
      'This will create an isolated development environment',
    ],
    WORKTREE_PENDING: [
      'Worktree creation is in progress',
      'Please wait for the operation to complete',
    ],
    WORKTREE_CREATED: [
      'Worktree is ready for development',
      'Open the worktree to start coding',
    ],
    BRANCH_CREATED: [
      'Branch is ready in worktree',
      'You can now make changes and commit them',
    ],
    CHANGES_PENDING: [
      'You have uncommitted changes',
      'Review and commit your changes',
    ],
    CHANGES_STAGED: [
      'Changes are staged for commit',
      'Create a commit with your changes',
    ],
    CHANGES_COMMITTED: [
      'Changes have been committed',
      'Consider creating a pull request',
    ],
    PR_CREATED: [
      'Pull request is ready for review',
      'Monitor the PR for feedback and approvals',
    ],
    PR_MERGED: [
      'Pull request has been merged',
      'Task is complete - consider cleanup',
    ],
    WORKTREE_ERROR: [
      'There was an error with Git operations',
      'Try refreshing status or recreating the worktree',
    ],
  }

  return recommendations[gitStatus] || []
}
