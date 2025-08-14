import { useQuery } from '@tanstack/react-query'
import { pullRequestsApi } from '@/lib/api/pull-requests'

export function usePullRequestByTask(taskId: string, enabled = true) {
  return useQuery({
    queryKey: ['pull-request-by-task', taskId],
    queryFn: () => pullRequestsApi.getPullRequestByTask(taskId),
    enabled: enabled && !!taskId,
    staleTime: 30000,
    refetchOnWindowFocus: false,
  })
}
