import { GitMerge, GitPullRequest, XCircle, AlertCircle, CheckCircle2 } from 'lucide-react'
import type { PullRequestStatus } from '@/types/pull-request'
import { Badge } from '@/components/ui/badge'
import { 
  Tooltip, 
  TooltipContent, 
  TooltipProvider, 
  TooltipTrigger 
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

interface PRStatusBadgeProps {
  status: PullRequestStatus
  prNumber?: number
  prUrl?: string
  className?: string
  showIcon?: boolean
  variant?: 'default' | 'compact'
}

const PR_STATUS_CONFIG = {
  OPEN: {
    label: 'Open',
    description: 'Pull request is open and ready for review',
    icon: GitPullRequest,
    color: 'bg-green-100 text-green-700 border-green-200',
    variant: 'secondary' as const,
  },
  MERGED: {
    label: 'Merged',
    description: 'Pull request has been merged successfully',
    icon: GitMerge,
    color: 'bg-purple-100 text-purple-700 border-purple-200',
    variant: 'secondary' as const,
  },
  CLOSED: {
    label: 'Closed',
    description: 'Pull request was closed without merging',
    icon: XCircle,
    color: 'bg-red-100 text-red-700 border-red-200',
    variant: 'destructive' as const,
  },
} as const

export function PRStatusBadge({
  status,
  prNumber,
  prUrl,
  className,
  showIcon = true,
  variant = 'default',
}: PRStatusBadgeProps) {
  const config = PR_STATUS_CONFIG[status]
  const Icon = config.icon

  const badgeContent = (
    <Badge
      variant={config.variant}
      className={cn(
        'flex items-center gap-1 text-xs border',
        config.color,
        variant === 'compact' && 'px-1.5 py-0.5 text-xs',
        className
      )}
    >
      {showIcon && <Icon className="h-3 w-3" />}
      <span className={variant === 'compact' ? 'hidden sm:inline' : ''}>
        {config.label}
        {prNumber && ` #${prNumber}`}
      </span>
    </Badge>
  )

  if (variant === 'compact') {
    return prUrl ? (
      <a
        href={prUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-block hover:scale-105 transition-transform"
        onClick={(e) => e.stopPropagation()}
      >
        {badgeContent}
      </a>
    ) : (
      badgeContent
    )
  }

  const tooltipContent = (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {prUrl ? (
            <a
              href={prUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-block hover:scale-105 transition-transform"
              onClick={(e) => e.stopPropagation()}
            >
              {badgeContent}
            </a>
          ) : (
            badgeContent
          )}
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-xs">
          <div className="space-y-1">
            <p className="font-medium">{config.label}</p>
            <p className="text-xs text-muted-foreground">
              {config.description}
            </p>
            {prNumber && (
              <p className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">
                PR #{prNumber}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )

  return tooltipContent
}

export function getPRStatusPriority(status: PullRequestStatus): number {
  const priority: Record<PullRequestStatus, number> = {
    OPEN: 3,
    MERGED: 2,
    CLOSED: 1,
  }
  return priority[status] || 0
}

export function getPRStatusColor(status: PullRequestStatus): string {
  return PR_STATUS_CONFIG[status].color
}

export function getPRStatusLabel(status: PullRequestStatus): string {
  return PR_STATUS_CONFIG[status].label
}