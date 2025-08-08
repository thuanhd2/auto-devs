import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { 
  PullRequest, 
  PullRequestsResponse, 
  PullRequestFilters,
  CreatePullRequestRequest,
  UpdatePullRequestRequest
} from '@/types/pull-request'
import { pullRequestsApi } from '@/lib/api/pull-requests'
import { useWebSocket } from '@/hooks/use-websocket'

interface UsePullRequestsOptions {
  projectId: string
  filters?: PullRequestFilters
  enabled?: boolean
}

export function usePullRequests({ projectId, filters, enabled = true }: UsePullRequestsOptions) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['pull-requests', projectId, filters],
    queryFn: () => pullRequestsApi.getPullRequests(projectId, filters),
    enabled: enabled && !!projectId,
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: false,
  })

  // WebSocket integration for real-time updates
  const { lastMessage } = useWebSocket()

  useEffect(() => {
    if (!lastMessage) return

    try {
      const message = typeof lastMessage === 'string' ? JSON.parse(lastMessage) : lastMessage

      if (message.type === 'pr_updated' && message.data?.project_id === projectId) {
        // Invalidate and refetch PR list
        queryClient.invalidateQueries({ queryKey: ['pull-requests', projectId] })
        
        // Update specific PR if we have detailed data
        if (message.data.pull_request) {
          queryClient.setQueryData(
            ['pull-request', message.data.pull_request.id],
            message.data.pull_request
          )
        }
      }

      if (message.type === 'pr_created' && message.data?.project_id === projectId) {
        // Add new PR to the list
        queryClient.invalidateQueries({ queryKey: ['pull-requests', projectId] })
        
        toast.success('Pull request created', {
          description: `PR #${message.data.pull_request?.github_pr_number} has been created`
        })
      }

      if (message.type === 'pr_merged' && message.data?.project_id === projectId) {
        // Update PR status
        queryClient.invalidateQueries({ queryKey: ['pull-requests', projectId] })
        
        toast.success('Pull request merged', {
          description: `PR #${message.data.pull_request?.github_pr_number} has been merged`
        })
      }

      if (message.type === 'pr_closed' && message.data?.project_id === projectId) {
        // Update PR status
        queryClient.invalidateQueries({ queryKey: ['pull-requests', projectId] })
        
        toast.info('Pull request closed', {
          description: `PR #${message.data.pull_request?.github_pr_number} has been closed`
        })
      }

    } catch (error) {
      console.error('Error processing WebSocket message:', error)
    }
  }, [lastMessage, projectId, queryClient])

  return query
}

interface UsePullRequestOptions {
  pullRequestId: string
  enabled?: boolean
}

export function usePullRequest({ pullRequestId, enabled = true }: UsePullRequestOptions) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['pull-request', pullRequestId],
    queryFn: () => pullRequestsApi.getPullRequest(pullRequestId),
    enabled: enabled && !!pullRequestId,
    staleTime: 30000,
    refetchOnWindowFocus: false,
  })

  // WebSocket integration for real-time updates
  const { lastMessage } = useWebSocket()

  useEffect(() => {
    if (!lastMessage) return

    try {
      const message = typeof lastMessage === 'string' ? JSON.parse(lastMessage) : lastMessage

      if (
        (message.type === 'pr_updated' || message.type === 'pr_comment_added' || message.type === 'pr_review_added') &&
        message.data?.pull_request?.id === pullRequestId
      ) {
        // Update the specific PR data
        queryClient.setQueryData(['pull-request', pullRequestId], message.data.pull_request)
      }
    } catch (error) {
      console.error('Error processing WebSocket message:', error)
    }
  }, [lastMessage, pullRequestId, queryClient])

  return query
}

export function usePullRequestByTask(taskId: string, enabled = true) {
  return useQuery({
    queryKey: ['pull-request-by-task', taskId],
    queryFn: () => pullRequestsApi.getPullRequestByTask(taskId),
    enabled: enabled && !!taskId,
    staleTime: 30000,
    refetchOnWindowFocus: false,
  })
}

export function usePullRequestMutations() {
  const queryClient = useQueryClient()

  const createPR = useMutation({
    mutationFn: (data: CreatePullRequestRequest) => pullRequestsApi.createPullRequest(data),
    onSuccess: (newPR) => {
      // Invalidate PR lists
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      // Set the new PR data
      queryClient.setQueryData(['pull-request', newPR.id], newPR)
      queryClient.setQueryData(['pull-request-by-task', newPR.task_id], newPR)
      
      toast.success('Pull request created', {
        description: `PR #${newPR.github_pr_number} has been created successfully`
      })
    },
    onError: (error) => {
      toast.error('Failed to create pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const updatePR = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: UpdatePullRequestRequest }) =>
      pullRequestsApi.updatePullRequest(id, updates),
    onSuccess: (updatedPR) => {
      // Update cached data
      queryClient.setQueryData(['pull-request', updatedPR.id], updatedPR)
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request updated')
    },
    onError: (error) => {
      toast.error('Failed to update pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const deletePR = useMutation({
    mutationFn: (id: string) => pullRequestsApi.deletePullRequest(id),
    onSuccess: (_, deletedId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: ['pull-request', deletedId] })
      
      // Invalidate PR lists
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request deleted')
    },
    onError: (error) => {
      toast.error('Failed to delete pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const syncPR = useMutation({
    mutationFn: (id: string) => pullRequestsApi.syncPullRequest(id),
    onSuccess: (syncedPR) => {
      // Update cached data
      queryClient.setQueryData(['pull-request', syncedPR.id], syncedPR)
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request synchronized', {
        description: 'Updated with latest data from GitHub'
      })
    },
    onError: (error) => {
      toast.error('Failed to sync pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const mergePR = useMutation({
    mutationFn: ({ id, method }: { id: string; method?: 'merge' | 'squash' | 'rebase' }) =>
      pullRequestsApi.mergePullRequest(id, method),
    onSuccess: (mergedPR) => {
      // Update cached data
      queryClient.setQueryData(['pull-request', mergedPR.id], mergedPR)
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request merged', {
        description: `PR #${mergedPR.github_pr_number} has been merged successfully`
      })
    },
    onError: (error) => {
      toast.error('Failed to merge pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const closePR = useMutation({
    mutationFn: (id: string) => pullRequestsApi.closePullRequest(id),
    onSuccess: (closedPR) => {
      // Update cached data
      queryClient.setQueryData(['pull-request', closedPR.id], closedPR)
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request closed', {
        description: `PR #${closedPR.github_pr_number} has been closed`
      })
    },
    onError: (error) => {
      toast.error('Failed to close pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  const reopenPR = useMutation({
    mutationFn: (id: string) => pullRequestsApi.reopenPullRequest(id),
    onSuccess: (reopenedPR) => {
      // Update cached data
      queryClient.setQueryData(['pull-request', reopenedPR.id], reopenedPR)
      
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['pull-requests'] })
      
      toast.success('Pull request reopened', {
        description: `PR #${reopenedPR.github_pr_number} has been reopened`
      })
    },
    onError: (error) => {
      toast.error('Failed to reopen pull request', {
        description: error instanceof Error ? error.message : 'An unknown error occurred'
      })
    },
  })

  return {
    createPR,
    updatePR,
    deletePR,
    syncPR,
    mergePR,
    closePR,
    reopenPR,
  }
}

// Hook for managing PR filters with URL state
export function usePullRequestFilters(initialFilters?: PullRequestFilters) {
  const [filters, setFilters] = useState<PullRequestFilters>(initialFilters || {
    sortBy: 'updated_at',
    sortOrder: 'desc',
  })

  const updateFilters = useCallback((newFilters: Partial<PullRequestFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters }))
  }, [])

  const resetFilters = useCallback(() => {
    setFilters({
      sortBy: 'updated_at',
      sortOrder: 'desc',
    })
  }, [])

  return {
    filters,
    updateFilters,
    resetFilters,
    setFilters,
  }
}