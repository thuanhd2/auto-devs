import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
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
import type { TaskGitStatus } from '@/types/task'
import { cn } from '@/lib/utils'

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
        <Icon 
          className={cn(
            'h-3 w-3',
            config.animate && 'animate-spin'
          )} 
        />
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
        <TooltipTrigger asChild>
          {badgeContent}
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-xs">
          <div className="space-y-1">
            <p className="font-medium">{config.label}</p>
            <p className="text-xs text-muted-foreground">
              {config.description}
            </p>
            {branchName && (
              <p className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">
                {branchName}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

export function getGitStatusPriority(status: TaskGitStatus): number {
  const priority: Record<TaskGitStatus, number> = {
    WORKTREE_ERROR: 10,
    PR_MERGED: 9,
    PR_CREATED: 8,
    CHANGES_COMMITTED: 7,
    CHANGES_STAGED: 6,
    CHANGES_PENDING: 5,
    BRANCH_CREATED: 4,
    WORKTREE_CREATED: 3,
    WORKTREE_PENDING: 2,
    NO_GIT: 1,
  }
  return priority[status] || 0
}

export function getGitStatusColor(status: TaskGitStatus): string {
  return GIT_STATUS_CONFIG[status].color
}

export function getGitStatusLabel(status: TaskGitStatus): string {
  return GIT_STATUS_CONFIG[status].label
}