import { useState, useEffect } from 'react'
import { useParams } from '@tanstack/react-router'
import type { PullRequest, PullRequestFilters } from '@/types/pull-request'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  usePullRequests,
  usePullRequest,
  usePullRequestMutations,
  usePullRequestFilters,
} from '@/hooks/use-pull-requests'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { PRActions } from './pr-actions'
import { PRDetail } from './pr-detail'
import { PRList } from './pr-list'

interface PRManagementPageProps {
  projectId: string
  className?: string
}

export function PRManagementPage({
  projectId,
  className,
}: PRManagementPageProps) {
  const [selectedPR, setSelectedPR] = useState<PullRequest | null>(null)
  const [detailSheetOpen, setDetailSheetOpen] = useState(false)
  const { filters, updateFilters, resetFilters } = usePullRequestFilters()

  // Queries
  const {
    data: pullRequestsResponse,
    isLoading: loadingPRs,
    error: prsError,
  } = usePullRequests({
    projectId,
    filters,
  })

  const { data: prDetail, isLoading: loadingDetail } = usePullRequest({
    pullRequestId: selectedPR?.id || '',
    enabled: !!selectedPR?.id,
  })

  // Mutations
  const { syncPR, mergePR, closePR, reopenPR } = usePullRequestMutations()

  const pullRequests = pullRequestsResponse?.pull_requests || []

  const handlePRSelect = (pr: PullRequest) => {
    setSelectedPR(pr)
    setDetailSheetOpen(true)
  }

  const handlePRAction = async (
    pr: PullRequest,
    action: 'sync' | 'merge' | 'close' | 'reopen',
    options?: any
  ) => {
    try {
      switch (action) {
        case 'sync':
          await syncPR.mutateAsync(pr.id)
          break
        case 'merge':
          await mergePR.mutateAsync({
            id: pr.id,
            method: options?.method || 'merge',
          })
          setDetailSheetOpen(false)
          break
        case 'close':
          await closePR.mutateAsync(pr.id)
          setDetailSheetOpen(false)
          break
        case 'reopen':
          await reopenPR.mutateAsync(pr.id)
          break
      }
    } catch {
      // Handle error silently
    }
  }

  const handleDetailAction = async (
    action: 'sync' | 'merge' | 'close' | 'reopen',
    options?: any
  ) => {
    if (!selectedPR) return
    await handlePRAction(selectedPR, action, options)
  }

  const handleAddComment = async (body: string) => {
    // This would be implemented when comment API is available
    console.log('Add comment:', body)
  }

  const handleCloseSheet = () => {
    setDetailSheetOpen(false)
    setSelectedPR(null)
  }

  if (prsError) {
    return (
      <div className={cn('p-6', className)}>
        <Card>
          <CardContent className='flex items-center justify-center py-12'>
            <div className='text-center'>
              <p className='text-destructive text-lg font-medium'>
                Error loading pull requests
              </p>
              <p className='text-muted-foreground'>
                {prsError instanceof Error
                  ? prsError.message
                  : 'An unknown error occurred'}
              </p>
              <Button
                variant='outline'
                className='mt-4'
                onClick={() => window.location.reload()}
              >
                Retry
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='text-2xl font-bold'>Pull Requests</h1>
          <p className='text-muted-foreground'>
            Manage and monitor pull requests for this project
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            variant='outline'
            onClick={resetFilters}
            disabled={loadingPRs}
          >
            Reset Filters
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      {pullRequestsResponse && (
        <div className='grid grid-cols-1 gap-4 md:grid-cols-4'>
          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                Total PRs
              </CardTitle>
            </CardHeader>
            <CardContent className='pt-0'>
              <div className='text-2xl font-bold'>
                {pullRequestsResponse.total}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                Open
              </CardTitle>
            </CardHeader>
            <CardContent className='pt-0'>
              <div className='text-2xl font-bold text-green-600'>
                {pullRequests.filter((pr) => pr.status === 'OPEN').length}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                Merged
              </CardTitle>
            </CardHeader>
            <CardContent className='pt-0'>
              <div className='text-2xl font-bold text-purple-600'>
                {pullRequests.filter((pr) => pr.status === 'MERGED').length}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-3'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                Closed
              </CardTitle>
            </CardHeader>
            <CardContent className='pt-0'>
              <div className='text-2xl font-bold text-red-600'>
                {pullRequests.filter((pr) => pr.status === 'CLOSED').length}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* PR List */}
      <PRList
        pullRequests={pullRequests}
        loading={loadingPRs}
        onPRSelect={handlePRSelect}
        onPRAction={handlePRAction}
      />

      {/* Detail Sheet */}
      <Sheet open={detailSheetOpen} onOpenChange={setDetailSheetOpen}>
        <SheetContent className='w-full overflow-y-auto sm:max-w-2xl'>
          <SheetHeader>
            <div className='flex items-center gap-2'>
              <Button variant='ghost' size='icon' onClick={handleCloseSheet}>
                <ArrowLeft className='h-4 w-4' />
              </Button>
              <SheetTitle>Pull Request Details</SheetTitle>
            </div>
          </SheetHeader>

          <div className='mt-6'>
            {selectedPR && prDetail ? (
              <div className='space-y-6'>
                <div className='grid grid-cols-1 gap-6 lg:grid-cols-3'>
                  <div className='lg:col-span-2'>
                    <PRDetail
                      pr={prDetail}
                      loading={loadingDetail}
                      onAction={handleDetailAction}
                      onAddComment={handleAddComment}
                    />
                  </div>
                  <div className='lg:col-span-1'>
                    <PRActions
                      pr={prDetail}
                      loading={
                        loadingDetail ||
                        syncPR.isPending ||
                        mergePR.isPending ||
                        closePR.isPending ||
                        reopenPR.isPending
                      }
                      onAction={handleDetailAction}
                    />
                  </div>
                </div>
              </div>
            ) : selectedPR ? (
              <div className='space-y-6'>
                <Skeleton className='h-32' />
                <Skeleton className='h-48' />
                <Skeleton className='h-64' />
              </div>
            ) : null}
          </div>
        </SheetContent>
      </Sheet>
    </div>
  )
}

export default PRManagementPage
