import { useState } from 'react'
import {
  GitFork,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
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

interface GitProjectStatusResponse {
  git_enabled: boolean
  worktree_exists: boolean
  repository_valid: boolean
  current_branch?: string
  remote_url?: string
  on_main_branch: boolean
  working_dir_status?: {
    is_clean: boolean
    has_staged_changes: boolean
    has_unstaged_changes: boolean
    has_untracked_files: boolean
  }
  status: string
}

interface GitStatusCardProps {
  projectId: string
  gitEnabled?: boolean
  onStatusUpdate?: (status: any) => void
}

export function GitStatusCard({
  projectId,
  gitEnabled = false,
  onStatusUpdate,
}: GitStatusCardProps) {
  const [status, setStatus] = useState<GitProjectStatusResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchStatus = async () => {
    if (!gitEnabled) return

    setLoading(true)
    setError(null)

    try {
      // TODO: Implement Git status API
      const gitStatus: GitProjectStatusResponse = {
        git_enabled: false,
        worktree_exists: false,
        repository_valid: false,
        on_main_branch: false,
        status: 'Git status not implemented',
      }
      setStatus(gitStatus)
      onStatusUpdate?.(gitStatus)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to fetch Git status'
      )
    } finally {
      setLoading(false)
    }
  }

  const testConnection = async () => {
    setLoading(true)
    setError(null)

    try {
      // TODO: Implement Git connection test
      console.log('Testing Git connection for project:', projectId)
      await fetchStatus() // Refresh status after successful test
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to test Git connection'
      )
    } finally {
      setLoading(false)
    }
  }

  const setupGit = async () => {
    setLoading(true)
    setError(null)

    try {
      // TODO: Implement Git setup
      console.log('Setting up Git for project:', projectId)
      await fetchStatus() // Refresh status after successful setup
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to setup Git project'
      )
    } finally {
      setLoading(false)
    }
  }

  if (!gitEnabled) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <GitFork className='h-5 w-5' />
            Git Integration
          </CardTitle>
          <CardDescription>
            Git integration is not enabled for this project
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className='h-4 w-4' />
            <AlertDescription>
              Enable Git integration in project settings to use advanced Git
              features.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <GitFork className='h-5 w-5' />
          Git Integration Status
        </CardTitle>
        <CardDescription>
          Current status of Git integration for this project
        </CardDescription>
      </CardHeader>

      <CardContent className='space-y-4'>
        {error && (
          <Alert variant='destructive'>
            <XCircle className='h-4 w-4' />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {!status && !loading && (
          <div className='flex items-center justify-between'>
            <p className='text-muted-foreground text-sm'>
              Click "Check Status" to view Git integration status
            </p>
            <Button onClick={fetchStatus} disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Checking...
                </>
              ) : (
                'Check Status'
              )}
            </Button>
          </div>
        )}

        {loading && (
          <div className='flex items-center justify-center py-4'>
            <Loader2 className='h-6 w-6 animate-spin' />
            <span className='ml-2'>Loading Git status...</span>
          </div>
        )}

        {status && (
          <div className='space-y-4'>
            {/* Overall Status */}
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2'>
                {status.repository_valid ? (
                  <CheckCircle className='h-5 w-5 text-green-500' />
                ) : (
                  <XCircle className='h-5 w-5 text-red-500' />
                )}
                <span className='font-medium'>Repository Status</span>
              </div>
              <Badge
                variant={status.repository_valid ? 'default' : 'destructive'}
              >
                {status.repository_valid ? 'Valid' : 'Invalid'}
              </Badge>
            </div>

            {/* Status Details */}
            <div className='grid grid-cols-2 gap-4 text-sm'>
              <div>
                <span className='font-medium'>Worktree:</span>
                <Badge
                  variant={status.worktree_exists ? 'default' : 'secondary'}
                  className='ml-2'
                >
                  {status.worktree_exists ? 'Exists' : 'Missing'}
                </Badge>
              </div>

              <div>
                <span className='font-medium'>Current Branch:</span>
                <span className='text-muted-foreground ml-2'>
                  {status.current_branch || 'Unknown'}
                </span>
              </div>

              <div>
                <span className='font-medium'>On Main Branch:</span>
                <Badge
                  variant={status.on_main_branch ? 'default' : 'secondary'}
                  className='ml-2'
                >
                  {status.on_main_branch ? 'Yes' : 'No'}
                </Badge>
              </div>

              <div>
                <span className='font-medium'>Remote URL:</span>
                <span className='text-muted-foreground ml-2 truncate'>
                  {status.remote_url || 'Not configured'}
                </span>
              </div>
            </div>

            {/* Working Directory Status */}
            {status.working_dir_status && (
              <div className='space-y-2'>
                <h4 className='text-sm font-medium'>Working Directory</h4>
                <div className='grid grid-cols-2 gap-2 text-xs'>
                  <div className='flex items-center gap-1'>
                    <div
                      className={`h-2 w-2 rounded-full ${status.working_dir_status.is_clean ? 'bg-green-500' : 'bg-yellow-500'}`}
                    />
                    <span>
                      Clean: {status.working_dir_status.is_clean ? 'Yes' : 'No'}
                    </span>
                  </div>
                  <div className='flex items-center gap-1'>
                    <div
                      className={`h-2 w-2 rounded-full ${status.working_dir_status.has_staged_changes ? 'bg-blue-500' : 'bg-gray-300'}`}
                    />
                    <span>
                      Staged:{' '}
                      {status.working_dir_status.has_staged_changes
                        ? 'Yes'
                        : 'No'}
                    </span>
                  </div>
                  <div className='flex items-center gap-1'>
                    <div
                      className={`h-2 w-2 rounded-full ${status.working_dir_status.has_unstaged_changes ? 'bg-orange-500' : 'bg-gray-300'}`}
                    />
                    <span>
                      Unstaged:{' '}
                      {status.working_dir_status.has_unstaged_changes
                        ? 'Yes'
                        : 'No'}
                    </span>
                  </div>
                  <div className='flex items-center gap-1'>
                    <div
                      className={`h-2 w-2 rounded-full ${status.working_dir_status.has_untracked_files ? 'bg-purple-500' : 'bg-gray-300'}`}
                    />
                    <span>
                      Untracked:{' '}
                      {status.working_dir_status.has_untracked_files
                        ? 'Yes'
                        : 'No'}
                    </span>
                  </div>
                </div>
              </div>
            )}

            {/* Status Message */}
            <Alert>
              <AlertCircle className='h-4 w-4' />
              <AlertDescription>{status.status}</AlertDescription>
            </Alert>

            {/* Action Buttons */}
            <div className='flex gap-2'>
              <Button
                onClick={testConnection}
                disabled={loading}
                variant='outline'
                size='sm'
              >
                {loading ? (
                  <>
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                    Testing...
                  </>
                ) : (
                  'Test Connection'
                )}
              </Button>

              {!status.worktree_exists && (
                <Button onClick={setupGit} disabled={loading} size='sm'>
                  {loading ? (
                    <>
                      <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      Setting up...
                    </>
                  ) : (
                    'Setup Git'
                  )}
                </Button>
              )}

              <Button
                onClick={fetchStatus}
                disabled={loading}
                variant='ghost'
                size='sm'
              >
                Refresh
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
