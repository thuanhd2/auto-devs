import type { TaskGitStatus } from '@/types/task'
import {
  GitBranch,
  AlertCircle,
  CheckCircle2,
  Clock,
  GitMerge,
  GitCommit,
  Loader2,
  XCircle,
  GitPullRequest,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface GitStatusBadgeProps {
  status: TaskGitStatus
  branchName?: string
  className?: string
  showIcon?: boolean
  variant?: 'default' | 'compact'
}

const GIT_STATUS_CONFIG = {
  NO_GIT: {
    label: 'No Git',
    description: 'No Git worktree or branch configured',
    icon: XCircle,
    color: 'bg-gray-100 text-gray-600 border-gray-200',
    variant: 'secondary' as const,
  },
  WORKTREE_PENDING: {
    label: 'Creating...',
    description: 'Worktree creation in progress',
    icon: Loader2,
    color: 'bg-blue-100 text-blue-700 border-blue-200',
    variant: 'secondary' as const,
    animate: true,
  },
  WORKTREE_CREATED: {
    label: 'Worktree Ready',
    description: 'Git worktree created and ready for development',
    icon: GitBranch,
    color: 'bg-emerald-100 text-emerald-700 border-emerald-200',
    variant: 'secondary' as const,
  },
  BRANCH_CREATED: {
    label: 'Branch Ready',
    description: 'Branch created in worktree, ready for changes',
    icon: GitBranch,
    color: 'bg-green-100 text-green-700 border-green-200',
    variant: 'secondary' as const,
  },
  CHANGES_PENDING: {
    label: 'Changes',
    description: 'Has uncommitted changes in working directory',
    icon: AlertCircle,
    color: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    variant: 'secondary' as const,
  },
  CHANGES_STAGED: {
    label: 'Staged',
    description: 'Changes staged and ready for commit',
    icon: GitCommit,
    color: 'bg-orange-100 text-orange-700 border-orange-200',
    variant: 'secondary' as const,
  },
  CHANGES_COMMITTED: {
    label: 'Committed',
    description: 'Changes committed to branch',
    icon: CheckCircle2,
    color: 'bg-blue-100 text-blue-700 border-blue-200',
    variant: 'secondary' as const,
  },
  PR_CREATED: {
    label: 'PR Created',
    description: 'Pull request created and ready for review',
    icon: GitPullRequest,
    color: 'bg-purple-100 text-purple-700 border-purple-200',
    variant: 'secondary' as const,
  },
  PR_MERGED: {
    label: 'Merged',
    description: 'Pull request merged successfully',
    icon: GitMerge,
    color: 'bg-green-100 text-green-700 border-green-200',
    variant: 'secondary' as const,
  },
  WORKTREE_ERROR: {
    label: 'Git Error',
    description: 'Error with Git operations',
    icon: XCircle,
    color: 'bg-red-100 text-red-700 border-red-200',
    variant: 'destructive' as const,
  },
} as const

export function GitStatusBadge({
  status,
  branchName,
  className,
  showIcon = true,
  variant = 'default',
}: GitStatusBadgeProps) {
  const config = GIT_STATUS_CONFIG[status]
  const Icon = config.icon

  const badgeContent = (
    <Badge
      variant={config.variant}
      className={cn(
        'flex items-center gap-1 text-xs',
        config.color,
        variant === 'compact' && 'px-1.5 py-0.5 text-xs',
        className
      )}
    >
      {showIcon && (
        <Icon className={cn('h-3 w-3', config.animate && 'animate-spin')} />
      )}
      <span className={variant === 'compact' ? 'hidden sm:inline' : ''}>
        {config.label}
      </span>
    </Badge>
  )

  if (variant === 'compact') {
    return badgeContent
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>{badgeContent}</TooltipTrigger>
        <TooltipContent side='top' className='max-w-xs'>
          <div className='space-y-1'>
            <p className='font-medium'>{config.label}</p>
            <p className='text-muted-foreground text-xs'>
              {config.description}
            </p>
            {branchName && (
              <p className='bg-muted rounded px-1.5 py-0.5 font-mono text-xs'>
                {branchName}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
