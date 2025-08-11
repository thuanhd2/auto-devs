import { useState } from 'react'
import type { PullRequest } from '@/types/pull-request'
import {
  GitMerge,
  GitPullRequest,
  GitBranch,
  ExternalLink,
  CheckCircle,
  XCircle,
  Clock,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useWebSocketConnection } from '@/context/websocket-context'
import { usePullRequests } from '@/hooks/use-pull-requests'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'

interface PRActionsProps {
  pr: PullRequest
  loading?: boolean
  onAction?: (
    action: 'sync' | 'merge' | 'close' | 'reopen',
    options?: any
  ) => void
  className?: string
}

type MergeMethod = 'merge' | 'squash' | 'rebase'

interface MergeDialogState {
  open: boolean
  method: MergeMethod
  commitTitle: string
  commitMessage: string
}

interface ConfirmDialogState {
  open: boolean
  action: 'close' | 'reopen' | null
  reason: string
}

const MERGE_METHODS = {
  merge: {
    label: 'Create a merge commit',
    description:
      'All commits from this branch will be added to the base branch via a merge commit.',
  },
  squash: {
    label: 'Squash and merge',
    description:
      'All commits will be combined into a single commit on the base branch.',
  },
  rebase: {
    label: 'Rebase and merge',
    description:
      'All commits will be rebased onto the base branch without a merge commit.',
  },
} as const

export function PRActions({
  pr,
  loading = false,
  onAction,
  className,
}: PRActionsProps) {
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [mergeDialog, setMergeDialog] = useState<MergeDialogState>({
    open: false,
    method: 'merge',
    commitTitle: pr.title,
    commitMessage: '',
  })
  const [confirmDialog, setConfirmDialog] = useState<ConfirmDialogState>({
    open: false,
    action: null,
    reason: '',
  })

  const canMerge = pr.status === 'OPEN' && pr.mergeable !== false
  const canClose = pr.status === 'OPEN'
  const canReopen = pr.status === 'CLOSED'
  const isMerged = pr.status === 'MERGED'

  const handleAction = async (action: string, options?: any) => {
    setActionLoading(action)
    try {
      await onAction?.(action as any, options)
    } finally {
      setActionLoading(null)
    }
  }

  const handleSync = () => {
    handleAction('sync')
  }

  const handleMerge = (
    method: MergeMethod,
    commitTitle: string,
    commitMessage: string
  ) => {
    handleAction('merge', { method, commitTitle, commitMessage })
    setMergeDialog((prev) => ({ ...prev, open: false }))
  }

  const handleCloseReopen = (action: 'close' | 'reopen', reason?: string) => {
    handleAction(action, { reason })
    setConfirmDialog({ open: false, action: null, reason: '' })
  }

  const openMergeDialog = () => {
    setMergeDialog({
      open: true,
      method: 'merge',
      commitTitle: pr.title,
      commitMessage: pr.body || '',
    })
  }

  const openConfirmDialog = (action: 'close' | 'reopen') => {
    setConfirmDialog({
      open: true,
      action,
      reason: '',
    })
  }

  return (
    <>
      <Card className={className}>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <GitPullRequest className='h-5 w-5' />
            Actions
          </CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          {/* Merge Status */}
          <div className='space-y-2'>
            <h4 className='text-sm font-medium'>Merge Status</h4>
            {isMerged ? (
              <div className='flex items-center gap-2'>
                <Badge
                  variant='secondary'
                  className='gap-1 bg-purple-100 text-purple-700'
                >
                  <GitMerge className='h-3 w-3' />
                  Merged
                </Badge>
                {pr.merged_at && (
                  <span className='text-muted-foreground text-sm'>
                    {new Date(pr.merged_at).toLocaleDateString()}
                  </span>
                )}
              </div>
            ) : pr.status === 'CLOSED' ? (
              <div className='flex items-center gap-2'>
                <Badge variant='destructive' className='gap-1'>
                  <XCircle className='h-3 w-3' />
                  Closed
                </Badge>
                {pr.closed_at && (
                  <span className='text-muted-foreground text-sm'>
                    {new Date(pr.closed_at).toLocaleDateString()}
                  </span>
                )}
              </div>
            ) : pr.mergeable === false ? (
              <div className='flex items-center gap-2'>
                <Badge variant='destructive' className='gap-1'>
                  <CheckCircle className='h-3 w-3' />
                  Conflicts
                </Badge>
                <span className='text-muted-foreground text-sm'>
                  Resolve conflicts before merging
                </span>
              </div>
            ) : pr.mergeable === true ? (
              <div className='flex items-center gap-2'>
                <Badge
                  variant='default'
                  className='gap-1 bg-green-100 text-green-700'
                >
                  <CheckCircle className='h-3 w-3' />
                  Ready to merge
                </Badge>
              </div>
            ) : (
              <div className='flex items-center gap-2'>
                <Badge variant='outline' className='gap-1'>
                  <Clock className='h-3 w-3 animate-spin' />
                  Checking
                </Badge>
                <span className='text-muted-foreground text-sm'>
                  Checking merge status...
                </span>
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className='flex flex-col gap-2'>
            {/* Sync Button */}
            <Button
              variant='outline'
              onClick={handleSync}
              disabled={loading || actionLoading === 'sync'}
              className='justify-start gap-2'
            >
              {actionLoading === 'sync' ? (
                <Clock className='h-4 w-4 animate-spin' />
              ) : (
                <GitPullRequest className='h-4 w-4' />
              )}
              Sync with GitHub
            </Button>

            {/* Merge Button */}
            {canMerge && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant='default'
                    disabled={loading || !!actionLoading || !canMerge}
                    className='justify-start gap-2'
                  >
                    {actionLoading === 'merge' ? (
                      <Clock className='h-4 w-4 animate-spin' />
                    ) : (
                      <GitMerge className='h-4 w-4' />
                    )}
                    Merge pull request
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align='start'>
                  <DropdownMenuItem onClick={openMergeDialog}>
                    <GitMerge className='mr-2 h-4 w-4' />
                    Merge with options
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() =>
                      handleMerge('merge', pr.title, pr.body || '')
                    }
                  >
                    Create a merge commit
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() =>
                      handleMerge('squash', pr.title, pr.body || '')
                    }
                  >
                    Squash and merge
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() =>
                      handleMerge('rebase', pr.title, pr.body || '')
                    }
                  >
                    Rebase and merge
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}

            {/* Close/Reopen Button */}
            {canClose && (
              <Button
                variant='destructive'
                onClick={() => openConfirmDialog('close')}
                disabled={loading || actionLoading === 'close'}
                className='justify-start gap-2'
              >
                {actionLoading === 'close' ? (
                  <Clock className='h-4 w-4 animate-spin' />
                ) : (
                  <XCircle className='h-4 w-4' />
                )}
                Close pull request
              </Button>
            )}

            {canReopen && (
              <Button
                variant='outline'
                onClick={() => openConfirmDialog('reopen')}
                disabled={loading || actionLoading === 'reopen'}
                className='justify-start gap-2'
              >
                {actionLoading === 'reopen' ? (
                  <Clock className='h-4 w-4 animate-spin' />
                ) : (
                  <GitBranch className='h-4 w-4' />
                )}
                Reopen pull request
              </Button>
            )}

            {/* External Links */}
            <Button
              variant='outline'
              onClick={() => window.open(pr.github_url, '_blank')}
              className='justify-start gap-2'
            >
              <ExternalLink className='h-4 w-4' />
              Open on GitHub
            </Button>
          </div>

          {/* Additional Information */}
          {pr.is_draft && (
            <div className='rounded-lg border border-yellow-200 bg-yellow-50 p-3'>
              <div className='flex items-start gap-2'>
                <GitPullRequest className='mt-0.5 h-4 w-4 text-yellow-600' />
                <div>
                  <p className='text-sm font-medium text-yellow-800'>
                    Draft Pull Request
                  </p>
                  <p className='text-sm text-yellow-700'>
                    This PR is in draft mode. Mark it as ready for review on
                    GitHub to enable merging.
                  </p>
                </div>
              </div>
            </div>
          )}

          {pr.mergeable === false && (
            <div className='rounded-lg border border-red-200 bg-red-50 p-3'>
              <div className='flex items-start gap-2'>
                <CheckCircle className='mt-0.5 h-4 w-4 text-red-600' />
                <div>
                  <p className='text-sm font-medium text-red-800'>
                    Merge conflicts
                  </p>
                  <p className='text-sm text-red-700'>
                    This PR has conflicts that must be resolved before it can be
                    merged.
                  </p>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Merge Dialog */}
      <Dialog
        open={mergeDialog.open}
        onOpenChange={(open) => setMergeDialog((prev) => ({ ...prev, open }))}
      >
        <DialogContent className='sm:max-w-[500px]'>
          <DialogHeader>
            <DialogTitle>Merge pull request</DialogTitle>
            <DialogDescription>
              Choose how you want to merge this pull request into{' '}
              {pr.base_branch}.
            </DialogDescription>
          </DialogHeader>

          <div className='space-y-4'>
            <div>
              <Label htmlFor='merge-method'>Merge method</Label>
              <Select
                value={mergeDialog.method}
                onValueChange={(value: MergeMethod) =>
                  setMergeDialog((prev) => ({ ...prev, method: value }))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(MERGE_METHODS).map(([method, config]) => (
                    <SelectItem key={method} value={method}>
                      <div>
                        <div className='font-medium'>{config.label}</div>
                        <div className='text-muted-foreground text-xs'>
                          {config.description}
                        </div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor='commit-title'>Commit title</Label>
              <input
                id='commit-title'
                type='text'
                value={mergeDialog.commitTitle}
                onChange={(e) =>
                  setMergeDialog((prev) => ({
                    ...prev,
                    commitTitle: e.target.value,
                  }))
                }
                className='border-input w-full rounded-md border px-3 py-2 text-sm'
              />
            </div>

            <div>
              <Label htmlFor='commit-message'>Commit message (optional)</Label>
              <Textarea
                id='commit-message'
                value={mergeDialog.commitMessage}
                onChange={(e) =>
                  setMergeDialog((prev) => ({
                    ...prev,
                    commitMessage: e.target.value,
                  }))
                }
                rows={4}
                placeholder='Add additional details about this merge...'
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant='outline'
              onClick={() =>
                setMergeDialog((prev) => ({ ...prev, open: false }))
              }
            >
              Cancel
            </Button>
            <Button
              onClick={() =>
                handleMerge(
                  mergeDialog.method,
                  mergeDialog.commitTitle,
                  mergeDialog.commitMessage
                )
              }
              disabled={
                !mergeDialog.commitTitle.trim() || actionLoading === 'merge'
              }
            >
              {actionLoading === 'merge' ? (
                <>
                  <Clock className='mr-2 h-4 w-4 animate-spin' />
                  Merging...
                </>
              ) : (
                <>
                  <GitMerge className='mr-2 h-4 w-4' />
                  Confirm merge
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Confirm Dialog */}
      <Dialog
        open={confirmDialog.open}
        onOpenChange={(open) => setConfirmDialog((prev) => ({ ...prev, open }))}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {confirmDialog.action === 'close'
                ? 'Close pull request'
                : 'Reopen pull request'}
            </DialogTitle>
            <DialogDescription>
              {confirmDialog.action === 'close'
                ? 'This will close the pull request without merging. You can reopen it later if needed.'
                : 'This will reopen the pull request and allow further changes and reviews.'}
            </DialogDescription>
          </DialogHeader>

          <div>
            <Label htmlFor='reason'>Reason (optional)</Label>
            <Textarea
              id='reason'
              value={confirmDialog.reason}
              onChange={(e) =>
                setConfirmDialog((prev) => ({
                  ...prev,
                  reason: e.target.value,
                }))
              }
              rows={3}
              placeholder={`Why are you ${confirmDialog.action === 'close' ? 'closing' : 'reopening'} this PR?`}
            />
          </div>

          <DialogFooter>
            <Button
              variant='outline'
              onClick={() =>
                setConfirmDialog({ open: false, action: null, reason: '' })
              }
            >
              Cancel
            </Button>
            <Button
              variant={
                confirmDialog.action === 'close' ? 'destructive' : 'default'
              }
              onClick={() =>
                confirmDialog.action &&
                handleCloseReopen(confirmDialog.action, confirmDialog.reason)
              }
              disabled={actionLoading === confirmDialog.action}
            >
              {actionLoading === confirmDialog.action ? (
                <>
                  <Clock className='mr-2 h-4 w-4 animate-spin' />
                  {confirmDialog.action === 'close'
                    ? 'Closing...'
                    : 'Reopening...'}
                </>
              ) : (
                <>
                  {confirmDialog.action === 'close' ? (
                    <XCircle className='mr-2 h-4 w-4' />
                  ) : (
                    <GitBranch className='mr-2 h-4 w-4' />
                  )}
                  Confirm {confirmDialog.action}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
