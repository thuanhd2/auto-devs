import { ExternalLink, GitMerge, AlertCircle, CheckCircle2, Clock, XCircle } from 'lucide-react'
import { usePullRequestByTask } from '@/hooks/use-pull-requests'
import { PRStatusBadge } from './pr-status-badge'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface PRIntegrationProps {
  taskId: string
  prUrl?: string
  variant?: 'compact' | 'detailed'
  onPRClick?: () => void
  className?: string
}

export function PRIntegration({ 
  taskId, 
  prUrl, 
  variant = 'compact', 
  onPRClick,
  className 
}: PRIntegrationProps) {
  const { data: pullRequest, isLoading, error } = usePullRequestByTask(taskId)

  // If we have a PR URL but no PR data, show basic link
  if (!pullRequest && prUrl && !isLoading) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <a
          href={prUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          <ExternalLink className="h-3 w-3" />
          <span>View PR</span>
        </a>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <Skeleton className="h-5 w-16" />
      </div>
    )
  }

  if (error || !pullRequest) {
    return null
  }

  if (variant === 'compact') {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <PRStatusBadge
          status={pullRequest.status}
          prNumber={pullRequest.github_pr_number}
          prUrl={pullRequest.github_url}
          variant="compact"
        />
        {pullRequest.is_draft && (
          <Badge variant="outline" className="text-xs px-1.5 py-0.5">
            Draft
          </Badge>
        )}
      </div>
    )
  }

  return (
    <div className={cn('space-y-3', className)}>
      {/* PR Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <PRStatusBadge
            status={pullRequest.status}
            prNumber={pullRequest.github_pr_number}
            prUrl={pullRequest.github_url}
          />
          {pullRequest.is_draft && (
            <Badge variant="outline" className="text-xs">
              Draft
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            onClick={onPRClick}
            className="h-7 px-2 text-xs"
          >
            View Details
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => window.open(pullRequest.github_url, '_blank')}
            className="h-7 px-2 text-xs"
          >
            <ExternalLink className="h-3 w-3" />
          </Button>
        </div>
      </div>

      {/* PR Title */}
      <div>
        <h4 className="text-sm font-medium line-clamp-2">
          {pullRequest.title}
        </h4>
        <div className="flex items-center gap-3 text-xs text-muted-foreground mt-1">
          <span>{pullRequest.repository}</span>
          <span>
            {pullRequest.head_branch} → {pullRequest.base_branch}
          </span>
          {pullRequest.created_by && (
            <span>by {pullRequest.created_by}</span>
          )}
        </div>
      </div>

      {/* PR Stats */}
      {(pullRequest.additions !== undefined || pullRequest.deletions !== undefined || pullRequest.changed_files !== undefined) && (
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          {pullRequest.additions !== undefined && (
            <span className="text-green-600 font-medium">+{pullRequest.additions}</span>
          )}
          {pullRequest.deletions !== undefined && (
            <span className="text-red-600 font-medium">-{pullRequest.deletions}</span>
          )}
          {pullRequest.changed_files !== undefined && (
            <span>{pullRequest.changed_files} files</span>
          )}
        </div>
      )}

      {/* PR Status Indicators */}
      <div className="flex items-center gap-2">
        {/* Merge Status */}
        {pullRequest.mergeable !== undefined && (
          <div className="flex items-center gap-1">
            {pullRequest.mergeable ? (
              <>
                <CheckCircle2 className="h-3 w-3 text-green-600" />
                <span className="text-xs text-green-700">Ready to merge</span>
              </>
            ) : (
              <>
                <AlertCircle className="h-3 w-3 text-red-600" />
                <span className="text-xs text-red-700">Has conflicts</span>
              </>
            )}
          </div>
        )}

        {/* Reviews */}
        {pullRequest.reviews && pullRequest.reviews.length > 0 && (
          <div className="flex items-center gap-1">
            <div className="flex -space-x-1">
              {pullRequest.reviews.slice(0, 3).map((review, index) => {
                const stateIcon = review.state === 'APPROVED' ? CheckCircle2 : 
                                 review.state === 'CHANGES_REQUESTED' ? XCircle : AlertCircle
                const StateIcon = stateIcon
                return (
                  <div
                    key={review.id}
                    className="h-4 w-4 rounded-full border border-background flex items-center justify-center"
                    title={`${review.reviewer}: ${review.state}`}
                  >
                    <StateIcon className="h-2.5 w-2.5" />
                  </div>
                )
              })}
              {pullRequest.reviews.length > 3 && (
                <div className="h-4 w-4 rounded-full bg-muted border border-background flex items-center justify-center text-xs">
                  +{pullRequest.reviews.length - 3}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Checks */}
        {pullRequest.checks && pullRequest.checks.length > 0 && (
          <div className="flex items-center gap-1">
            {pullRequest.checks.some(check => check.status === 'SUCCESS') && (
              <CheckCircle2 className="h-3 w-3 text-green-600" title="All checks passed" />
            )}
            {pullRequest.checks.some(check => check.status === 'FAILURE' || check.status === 'ERROR') && (
              <XCircle className="h-3 w-3 text-red-600" title="Some checks failed" />
            )}
            {pullRequest.checks.some(check => check.status === 'PENDING') && (
              <Clock className="h-3 w-3 text-yellow-600" title="Checks in progress" />
            )}
          </div>
        )}
      </div>
    </div>
  )
}