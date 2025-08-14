import { useState } from 'react'
import {
  GitFork,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { useProject, useReinitGitRepository } from '@/hooks/use-projects'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
  const { data: project, isLoading } = useProject(projectId)
  const { mutate: reinitGit } = useReinitGitRepository()
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
        <Button
          onClick={() => reinitGit(projectId)}
          disabled={loading}
          variant='secondary'
          size='sm'
        >
          {loading ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Reinitializing...
            </>
          ) : (
            'Reload'
          )}
        </Button>
      </CardContent>
    </Card>
  )
}
