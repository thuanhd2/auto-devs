import { useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  PullRequestFilters,
  CreatePullRequestRequest,
  UpdatePullRequestRequest,
} from '@/types/pull-request'
import { toast } from 'sonner'
import { pullRequestsApi } from '@/lib/api/pull-requests'

interface UsePullRequestsOptions {
  projectId: string
  filters?: PullRequestFilters
  enabled?: boolean
}
interface UsePullRequestOptions {
  pullRequestId: string
  enabled?: boolean
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
