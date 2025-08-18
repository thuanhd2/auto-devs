import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { pullRequestsApi } from '@/lib/api/pull-requests'
import { tasksApi } from '@/lib/api/tasks'

export function usePullRequestByTask(taskId: string, enabled = true) {
  return useQuery({
    queryKey: ['pull-request-by-task', taskId],
    queryFn: () => pullRequestsApi.getPullRequestByTask(taskId),
    enabled: enabled && !!taskId,
    staleTime: 30000,
    refetchOnWindowFocus: false,
  })
}

export function useCreatePullRequest() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (taskId: string) => tasksApi.createPullRequestForTask(taskId),
    onSuccess: (data, taskId) => {
      // Invalidate the pull request query for this task
      queryClient.invalidateQueries({
        queryKey: ['pull-request-by-task', taskId],
      })
      // Set the new pull request data in the cache
      queryClient.setQueryData(['pull-request-by-task', taskId], data)
    },
  })
}
