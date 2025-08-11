import { useState, useMemo } from 'react'
import { formatDistanceToNow } from 'date-fns'
import type { PullRequest } from '@/types/pull-request'
import {
  GitBranch,
  GitPullRequest,
  ExternalLink,
  RefreshCw,
  Loader2,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Clock,
  User,
  Calendar,
  MessageSquare,
} from 'lucide-react'
import { useWebSocketConnection } from '@/context/websocket-context'
import { usePullRequests } from '@/hooks/use-pull-requests'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface PRListProps {
  pullRequests: PullRequest[]
  loading?: boolean
  onPRSelect?: (pr: PullRequest) => void
  onPRAction?: (
    pr: PullRequest,
    action: 'sync' | 'merge' | 'close' | 'reopen'
  ) => void
  className?: string
}

const PR_STATUS_CONFIG = {
  OPEN: {
    label: 'Open',
    variant: 'default' as const,
    color: 'text-green-700 bg-green-100',
    icon: AlertTriangle,
  },
  MERGED: {
    label: 'Merged',
    variant: 'secondary' as const,
    color: 'text-purple-700 bg-purple-100',
    icon: CheckCircle,
  },
  CLOSED: {
    label: 'Closed',
    variant: 'destructive' as const,
    color: 'text-red-700 bg-red-100',
    icon: XCircle,
  },
} as const

export function PRList({
  pullRequests,
  loading = false,
  onPRSelect,
  onPRAction,
  className,
}: PRListProps) {
  const [filters, setFilters] = useState({
    sortBy: 'updated_at',
    sortOrder: 'desc',
  })
  const [searchQuery, setSearchQuery] = useState('')

  // Filter and sort PRs
  const filteredPRs = useMemo(() => {
    let filtered = pullRequests

    // Apply search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (pr) =>
          pr.title.toLowerCase().includes(query) ||
          pr.body.toLowerCase().includes(query) ||
          pr.repository.toLowerCase().includes(query) ||
          pr.head_branch.toLowerCase().includes(query) ||
          pr.created_by?.toLowerCase().includes(query)
      )
    }

    // Apply sorting
    filtered.sort((a, b) => {
      const aValue = a[filters.sortBy as keyof PullRequest]
      const bValue = b[filters.sortBy as keyof PullRequest]

      if (filters.sortOrder === 'asc') {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
      }
    })

    return filtered
  }, [pullRequests, searchQuery, filters])

  // Get unique repositories for filter dropdown
  const repositories = useMemo(
    () => Array.from(new Set(pullRequests.map((pr) => pr.repository))).sort(),
    [pullRequests]
  )

  const handleStatusFilter = (statuses: string[]) => {
    setFilters((prev) => ({ ...prev, status: statuses }))
  }

  const handleSortChange = (sortBy: string) => {
    setFilters((prev) => ({ ...prev, sortBy }))
  }

  const handleRepositoryFilter = (repository: string) => {
    setFilters((prev) => ({ ...prev, repository }))
  }

  if (loading) {
    return (
      <div className='space-y-4'>
        <div className='flex gap-4'>
          <Loader2 className='text-primary h-10 w-10 animate-spin' />
          <Skeleton className='h-10 flex-1' />
          <Skeleton className='h-10 w-32' />
          <Skeleton className='h-10 w-32' />
        </div>
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className='h-32' />
        ))}
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      {/* Filters and Search */}
      <div className='flex flex-col gap-4 sm:flex-row'>
        <div className='relative flex-1'>
          <Search className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
          <Input
            placeholder='Search PRs by title, description, repository, branch...'
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className='pl-10'
          />
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='outline' size='sm' className='gap-2'>
              <Filter className='h-4 w-4' />
              Status
              {filters.status && filters.status.length > 0 && (
                <Badge variant='secondary' className='ml-1'>
                  {filters.status.length}
                </Badge>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            {Object.entries(PR_STATUS_CONFIG).map(([status, config]) => (
              <DropdownMenuCheckboxItem
                key={status}
                checked={filters.status?.includes(status as string) || false}
                onCheckedChange={(checked) => {
                  const newStatuses = filters.status || []
                  if (checked) {
                    handleStatusFilter([...newStatuses, status as string])
                  } else {
                    handleStatusFilter(newStatuses.filter((s) => s !== status))
                  }
                }}
              >
                {config.label}
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        <Select
          value={filters.repository || 'all'}
          onValueChange={handleRepositoryFilter}
        >
          <SelectTrigger className='w-48'>
            <SelectValue placeholder='Repository' />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value='all'>All Repositories</SelectItem>
            {repositories.map((repo) => (
              <SelectItem key={repo} value={repo}>
                {repo}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='outline' size='sm' className='gap-2'>
              {filters.sortOrder === 'asc' ? (
                <SortAsc className='h-4 w-4' />
              ) : (
                <SortDesc className='h-4 w-4' />
              )}
              Sort
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            <DropdownMenuItem onClick={() => handleSortChange('updated_at')}>
              Last Updated
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleSortChange('created_at')}>
              Created Date
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleSortChange('title')}>
              Title
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => handleSortChange('github_pr_number')}
            >
              PR Number
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleSortChange('status')}>
              Status
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* PR List */}
      <div className='space-y-4'>
        {filteredPRs.length === 0 ? (
          <Card>
            <CardContent className='flex items-center justify-center py-12'>
              <div className='text-center'>
                <GitMerge className='text-muted-foreground mx-auto h-12 w-12' />
                <p className='mt-4 text-lg font-medium'>
                  No pull requests found
                </p>
                <p className='text-muted-foreground'>
                  {searchQuery || filters.status || filters.repository
                    ? 'Try adjusting your filters or search query.'
                    : 'Pull requests will appear here once created.'}
                </p>
              </div>
            </CardContent>
          </Card>
        ) : (
          filteredPRs.map((pr) => (
            <PRCard
              key={pr.id}
              pr={pr}
              onClick={onPRSelect}
              onAction={onPRAction}
            />
          ))
        )}
      </div>

      {/* Results count */}
      {filteredPRs.length > 0 && (
        <p className='text-muted-foreground text-center text-sm'>
          Showing {filteredPRs.length} of {pullRequests.length} pull requests
        </p>
      )}
    </div>
  )
}

interface PRCardProps {
  pr: PullRequest
  onClick?: (pr: PullRequest) => void
  onAction?: (
    pr: PullRequest,
    action: 'sync' | 'merge' | 'close' | 'reopen'
  ) => void
}

function PRCard({ pr, onClick, onAction }: PRCardProps) {
  const statusConfig = PR_STATUS_CONFIG[pr.status]
  const StatusIcon = statusConfig.icon

  const updatedAgo = formatDistanceToNow(new Date(pr.updated_at), {
    addSuffix: true,
  })
  const createdAgo = formatDistanceToNow(new Date(pr.created_at), {
    addSuffix: true,
  })

  const handleCardClick = () => {
    onClick?.(pr)
  }

  const handleActionClick = (
    e: React.MouseEvent,
    action: 'sync' | 'merge' | 'close' | 'reopen'
  ) => {
    e.stopPropagation()
    onAction?.(pr, action)
  }

  return (
    <Card
      className='cursor-pointer transition-all duration-200 hover:shadow-md'
      onClick={handleCardClick}
    >
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between gap-3'>
          <div className='min-w-0 flex-1'>
            <div className='mb-2 flex items-center gap-2'>
              <Badge className={cn('gap-1', statusConfig.color)}>
                <StatusIcon className='h-3 w-3' />
                {statusConfig.label}
              </Badge>
              <span className='text-muted-foreground text-sm'>
                #{pr.github_pr_number}
              </span>
              {pr.is_draft && (
                <Badge variant='outline' className='text-xs'>
                  Draft
                </Badge>
              )}
            </div>
            <h3 className='mb-1 line-clamp-2 text-lg leading-tight font-semibold'>
              {pr.title}
            </h3>
            <div className='text-muted-foreground flex items-center gap-4 text-sm'>
              <span>{pr.repository}</span>
              <span>
                {pr.head_branch} → {pr.base_branch}
              </span>
              {pr.created_by && <span>by {pr.created_by}</span>}
            </div>
          </div>

          <div className='flex items-center gap-2'>
            <a
              href={pr.github_url}
              target='_blank'
              rel='noopener noreferrer'
              className='text-muted-foreground hover:text-foreground transition-colors'
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className='h-4 w-4' />
            </a>
          </div>
        </div>
      </CardHeader>

      <CardContent className='pt-0'>
        {pr.body && (
          <p className='text-muted-foreground mb-3 line-clamp-2 text-sm'>
            {pr.body}
          </p>
        )}

        <div className='flex items-center justify-between'>
          <div className='text-muted-foreground flex items-center gap-4 text-xs'>
            <span>Updated {updatedAgo}</span>
            <span>Created {createdAgo}</span>
            {pr.additions !== undefined && pr.deletions !== undefined && (
              <div className='flex items-center gap-1'>
                <span className='text-green-600'>+{pr.additions}</span>
                <span className='text-red-600'>-{pr.deletions}</span>
              </div>
            )}
            {pr.changed_files !== undefined && (
              <span>{pr.changed_files} files</span>
            )}
          </div>

          <div className='flex items-center gap-2'>
            {pr.labels.length > 0 && (
              <div className='flex gap-1'>
                {pr.labels.slice(0, 3).map((label) => (
                  <Badge
                    key={label}
                    variant='outline'
                    className='px-1.5 py-0.5 text-xs'
                  >
                    {label}
                  </Badge>
                ))}
                {pr.labels.length > 3 && (
                  <Badge variant='outline' className='px-1.5 py-0.5 text-xs'>
                    +{pr.labels.length - 3}
                  </Badge>
                )}
              </div>
            )}

            {pr.reviewers.length > 0 && (
              <div className='flex -space-x-1'>
                {pr.reviewers.slice(0, 3).map((reviewer, index) => (
                  <div
                    key={reviewer}
                    className='bg-muted border-background flex h-6 w-6 items-center justify-center rounded-full border-2 text-xs font-medium'
                    title={reviewer}
                  >
                    {reviewer.charAt(0).toUpperCase()}
                  </div>
                ))}
                {pr.reviewers.length > 3 && (
                  <div className='bg-muted border-background flex h-6 w-6 items-center justify-center rounded-full border-2 text-xs font-medium'>
                    +{pr.reviewers.length - 3}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
