import { useState } from 'react'
import { formatDistanceToNow, format } from 'date-fns'
import { 
  ExternalLink, 
  GitBranch, 
  GitMerge, 
  MessageCircle, 
  CheckCircle2, 
  XCircle, 
  Clock, 
  User, 
  Calendar,
  FileText,
  GitCommit,
  AlertCircle,
  Eye,
  ThumbsUp,
  ThumbsDown,
  Loader2,
  CheckCircle,
  X
} from 'lucide-react'
import type { PullRequest, PullRequestComment, PullRequestReview, PullRequestCheck } from '@/types/pull-request'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Avatar } from '@/components/ui/avatar'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface PRDetailProps {
  pr: PullRequest
  loading?: boolean
  onAction?: (action: 'sync' | 'merge' | 'close' | 'reopen') => void
  onAddComment?: (body: string) => void
  className?: string
}

const STATUS_CONFIG = {
  OPEN: {
    label: 'Open',
    color: 'text-green-700 bg-green-100 border-green-200',
    icon: AlertCircle,
  },
  MERGED: {
    label: 'Merged',
    color: 'text-purple-700 bg-purple-100 border-purple-200',
    icon: GitMerge,
  },
  CLOSED: {
    label: 'Closed',
    color: 'text-red-700 bg-red-100 border-red-200',
    icon: XCircle,
  },
}

const REVIEW_STATE_CONFIG = {
  APPROVED: {
    label: 'Approved',
    color: 'text-green-700 bg-green-100',
    icon: ThumbsUp,
  },
  CHANGES_REQUESTED: {
    label: 'Changes Requested',
    color: 'text-red-700 bg-red-100',
    icon: ThumbsDown,
  },
  COMMENTED: {
    label: 'Commented',
    color: 'text-blue-700 bg-blue-100',
    icon: MessageCircle,
  },
}

const CHECK_STATUS_CONFIG = {
  PENDING: { label: 'Pending', color: 'text-yellow-700 bg-yellow-100', icon: Clock },
  SUCCESS: { label: 'Success', color: 'text-green-700 bg-green-100', icon: CheckCircle },
  FAILURE: { label: 'Failed', color: 'text-red-700 bg-red-100', icon: X },
  ERROR: { label: 'Error', color: 'text-red-700 bg-red-100', icon: AlertCircle },
}

export function PRDetail({ pr, loading = false, onAction, onAddComment, className }: PRDetailProps) {
  const [activeTab, setActiveTab] = useState('overview')
  const [newComment, setNewComment] = useState('')
  const [submittingComment, setSubmittingComment] = useState(false)

  if (loading) {
    return (
      <div className={cn('space-y-6', className)}>
        <Skeleton className="h-32" />
        <Skeleton className="h-64" />
        <Skeleton className="h-48" />
      </div>
    )
  }

  const statusConfig = STATUS_CONFIG[pr.status]
  const StatusIcon = statusConfig.icon
  const updatedAgo = formatDistanceToNow(new Date(pr.updated_at), { addSuffix: true })

  const handleSubmitComment = async () => {
    if (!newComment.trim() || submittingComment) return
    
    setSubmittingComment(true)
    try {
      await onAddComment?.(newComment.trim())
      setNewComment('')
    } finally {
      setSubmittingComment(false)
    }
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* PR Header */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-2">
                <Badge className={cn('gap-1 border', statusConfig.color)}>
                  <StatusIcon className="h-3 w-3" />
                  {statusConfig.label}
                </Badge>
                <span className="text-sm text-muted-foreground">
                  #{pr.github_pr_number}
                </span>
                {pr.is_draft && (
                  <Badge variant="outline">Draft</Badge>
                )}
              </div>
              <h1 className="text-2xl font-bold leading-tight mb-2">
                {pr.title}
              </h1>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <div className="flex items-center gap-1">
                  <GitBranch className="h-4 w-4" />
                  <span className="font-mono">
                    {pr.head_branch} → {pr.base_branch}
                  </span>
                </div>
                <div className="flex items-center gap-1">
                  <User className="h-4 w-4" />
                  <span>by {pr.created_by || 'Unknown'}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  <span>Updated {updatedAgo}</span>
                </div>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => window.open(pr.github_url, '_blank')}
                className="gap-2"
              >
                <ExternalLink className="h-4 w-4" />
                View on GitHub
              </Button>
            </div>
          </div>
        </CardHeader>
        
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="md:col-span-2">
              {pr.body ? (
                <div className="prose prose-sm max-w-none">
                  <div className="whitespace-pre-wrap text-sm text-foreground bg-muted/30 p-4 rounded-lg border">
                    {pr.body}
                  </div>
                </div>
              ) : (
                <p className="text-muted-foreground italic">No description provided.</p>
              )}
            </div>
            
            <div className="space-y-4">
              {/* Repository */}
              <div>
                <h4 className="text-sm font-medium mb-2">Repository</h4>
                <p className="text-sm font-mono bg-muted px-2 py-1 rounded">
                  {pr.repository}
                </p>
              </div>
              
              {/* Changes */}
              {(pr.additions !== undefined || pr.deletions !== undefined || pr.changed_files !== undefined) && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Changes</h4>
                  <div className="flex items-center gap-3 text-sm">
                    {pr.additions !== undefined && (
                      <span className="text-green-600 font-medium">+{pr.additions}</span>
                    )}
                    {pr.deletions !== undefined && (
                      <span className="text-red-600 font-medium">-{pr.deletions}</span>
                    )}
                    {pr.changed_files !== undefined && (
                      <span className="text-muted-foreground">{pr.changed_files} files</span>
                    )}
                  </div>
                </div>
              )}
              
              {/* Labels */}
              {pr.labels.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Labels</h4>
                  <div className="flex flex-wrap gap-1">
                    {pr.labels.map((label) => (
                      <Badge key={label} variant="outline" className="text-xs">
                        {label}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
              
              {/* Assignees */}
              {pr.assignees.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Assignees</h4>
                  <div className="space-y-1">
                    {pr.assignees.map((assignee) => (
                      <div key={assignee} className="flex items-center gap-2">
                        <Avatar className="h-5 w-5">
                          <div className="h-full w-full bg-muted flex items-center justify-center text-xs">
                            {assignee.charAt(0).toUpperCase()}
                          </div>
                        </Avatar>
                        <span className="text-sm">{assignee}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              
              {/* Merge Status */}
              {pr.mergeable !== undefined && (
                <div>
                  <h4 className="text-sm font-medium mb-2">Merge Status</h4>
                  <Badge 
                    variant={pr.mergeable ? "default" : "destructive"}
                    className="gap-1"
                  >
                    {pr.mergeable ? (
                      <>
                        <CheckCircle2 className="h-3 w-3" />
                        Ready to merge
                      </>
                    ) : (
                      <>
                        <XCircle className="h-3 w-3" />
                        Conflicts
                      </>
                    )}
                  </Badge>
                  {pr.mergeable_state && (
                    <p className="text-xs text-muted-foreground mt-1">
                      State: {pr.mergeable_state}
                    </p>
                  )}
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="comments" className="gap-2">
            Comments
            {pr.comments && pr.comments.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {pr.comments.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="reviews" className="gap-2">
            Reviews
            {pr.reviews && pr.reviews.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {pr.reviews.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="checks" className="gap-2">
            Checks
            {pr.checks && pr.checks.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {pr.checks.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-6">
          <OverviewTab pr={pr} />
        </TabsContent>

        <TabsContent value="comments" className="mt-6 space-y-6">
          <CommentsTab 
            comments={pr.comments || []} 
            onAddComment={handleSubmitComment}
            newComment={newComment}
            onNewCommentChange={setNewComment}
            submitting={submittingComment}
          />
        </TabsContent>

        <TabsContent value="reviews" className="mt-6">
          <ReviewsTab reviews={pr.reviews || []} />
        </TabsContent>

        <TabsContent value="checks" className="mt-6">
          <ChecksTab checks={pr.checks || []} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function OverviewTab({ pr }: { pr: PullRequest }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-5 w-5" />
          Pull Request Information
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Created:</span>
            <p className="font-medium">
              {format(new Date(pr.created_at), 'PPpp')}
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">Last Updated:</span>
            <p className="font-medium">
              {format(new Date(pr.updated_at), 'PPpp')}
            </p>
          </div>
          {pr.merged_at && (
            <div>
              <span className="text-muted-foreground">Merged:</span>
              <p className="font-medium">
                {format(new Date(pr.merged_at), 'PPpp')}
                {pr.merged_by && <span className="text-muted-foreground"> by {pr.merged_by}</span>}
              </p>
            </div>
          )}
          {pr.closed_at && (
            <div>
              <span className="text-muted-foreground">Closed:</span>
              <p className="font-medium">
                {format(new Date(pr.closed_at), 'PPpp')}
              </p>
            </div>
          )}
          {pr.merge_commit_sha && (
            <div>
              <span className="text-muted-foreground">Merge Commit:</span>
              <p className="font-mono text-sm bg-muted px-2 py-1 rounded">
                {pr.merge_commit_sha.substring(0, 8)}
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function CommentsTab({ 
  comments, 
  onAddComment, 
  newComment, 
  onNewCommentChange, 
  submitting 
}: { 
  comments: PullRequestComment[]
  onAddComment: () => void
  newComment: string
  onNewCommentChange: (value: string) => void
  submitting: boolean
}) {
  return (
    <div className="space-y-6">
      {/* Add Comment */}
      <Card>
        <CardHeader>
          <CardTitle>Add Comment</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Textarea
            placeholder="Write a comment..."
            value={newComment}
            onChange={(e) => onNewCommentChange(e.target.value)}
            className="min-h-24"
          />
          <Button 
            onClick={onAddComment}
            disabled={!newComment.trim() || submitting}
            className="gap-2"
          >
            {submitting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Adding...
              </>
            ) : (
              'Add Comment'
            )}
          </Button>
        </CardContent>
      </Card>

      {/* Comments List */}
      {comments.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <div className="text-center">
              <MessageCircle className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-4 text-lg font-medium">No comments yet</p>
              <p className="text-muted-foreground">
                Be the first to leave a comment on this pull request.
              </p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <CommentCard key={comment.id} comment={comment} />
          ))}
        </div>
      )}
    </div>
  )
}

function CommentCard({ comment }: { comment: PullRequestComment }) {
  const createdAgo = formatDistanceToNow(new Date(comment.created_at), { addSuffix: true })
  
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <div className="h-full w-full bg-muted flex items-center justify-center text-xs">
                {comment.author.charAt(0).toUpperCase()}
              </div>
            </Avatar>
            <span className="font-medium">{comment.author}</span>
            <span className="text-sm text-muted-foreground">commented {createdAgo}</span>
          </div>
          {comment.is_resolved && (
            <Badge variant="outline" className="text-xs">
              Resolved
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {comment.file_path && (
          <div className="mb-2 text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded">
            {comment.file_path}
            {comment.line && `:${comment.line}`}
          </div>
        )}
        <div className="whitespace-pre-wrap text-sm">
          {comment.body}
        </div>
      </CardContent>
    </Card>
  )
}

function ReviewsTab({ reviews }: { reviews: PullRequestReview[] }) {
  if (reviews.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <Eye className="mx-auto h-12 w-12 text-muted-foreground" />
            <p className="mt-4 text-lg font-medium">No reviews yet</p>
            <p className="text-muted-foreground">
              Reviews will appear here once submitted.
            </p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {reviews.map((review) => (
        <ReviewCard key={review.id} review={review} />
      ))}
    </div>
  )
}

function ReviewCard({ review }: { review: PullRequestReview }) {
  const stateConfig = REVIEW_STATE_CONFIG[review.state as keyof typeof REVIEW_STATE_CONFIG]
  const StateIcon = stateConfig?.icon || Eye
  const submittedAgo = review.submitted_at 
    ? formatDistanceToNow(new Date(review.submitted_at), { addSuffix: true })
    : null

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <div className="h-full w-full bg-muted flex items-center justify-center text-xs">
                {review.reviewer.charAt(0).toUpperCase()}
              </div>
            </Avatar>
            <span className="font-medium">{review.reviewer}</span>
            <Badge className={cn('gap-1', stateConfig?.color || 'text-muted-foreground bg-muted')}>
              <StateIcon className="h-3 w-3" />
              {stateConfig?.label || review.state}
            </Badge>
            {submittedAgo && (
              <span className="text-sm text-muted-foreground">{submittedAgo}</span>
            )}
          </div>
        </div>
      </CardHeader>
      {review.body && (
        <CardContent className="pt-0">
          <div className="whitespace-pre-wrap text-sm">
            {review.body}
          </div>
        </CardContent>
      )}
    </Card>
  )
}

function ChecksTab({ checks }: { checks: PullRequestCheck[] }) {
  if (checks.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <CheckCircle2 className="mx-auto h-12 w-12 text-muted-foreground" />
            <p className="mt-4 text-lg font-medium">No checks configured</p>
            <p className="text-muted-foreground">
              CI/CD checks will appear here once configured.
            </p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {checks.map((check) => (
        <CheckCard key={check.id} check={check} />
      ))}
    </div>
  )
}

function CheckCard({ check }: { check: PullRequestCheck }) {
  const statusConfig = CHECK_STATUS_CONFIG[check.status as keyof typeof CHECK_STATUS_CONFIG]
  const StatusIcon = statusConfig?.icon || Clock
  const duration = check.started_at && check.completed_at
    ? formatDistanceToNow(new Date(check.started_at), { addSuffix: false })
    : null

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <StatusIcon 
              className={cn(
                'h-5 w-5',
                check.status === 'PENDING' && 'animate-spin'
              )} 
            />
            <div>
              <h4 className="font-medium">{check.check_name}</h4>
              {check.conclusion && (
                <p className="text-sm text-muted-foreground">
                  {check.conclusion}
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge className={cn('gap-1', statusConfig?.color || '')}>
              {statusConfig?.label || check.status}
            </Badge>
            {check.details_url && (
              <Button variant="outline" size="sm" asChild>
                <a href={check.details_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4" />
                </a>
              </Button>
            )}
          </div>
        </div>
        {duration && (
          <p className="text-xs text-muted-foreground mt-2">
            Took {duration}
          </p>
        )}
      </CardContent>
    </Card>
  )
}