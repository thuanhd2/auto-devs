import { useQuery } from '@tanstack/react-query'
import type { ExecutionFilters } from '@/types/execution'
import { executionsApi } from '@/lib/api/executions'

const EXECUTIONS_QUERY_KEY = 'executions'
const EXECUTION_QUERY_KEY = 'execution'

// Get executions for a specific task
export function useTaskExecutions(taskId: string, filters?: ExecutionFilters) {
  return useQuery({
    queryKey: [EXECUTIONS_QUERY_KEY, 'task', taskId, filters],
    queryFn: () => executionsApi.getTaskExecutions(taskId, filters),
    enabled: !!taskId,
    refetchInterval: (_data) => {
      // Auto-refetch every 5 seconds if there are any active executions
      // const hasActiveExecution = data?.data.some(execution =>
      //   execution.status === 'running' || execution.status === 'pending'
      // )
      // return hasActiveExecution ? 5000 : false
      return false
    },
  })
}

export function useExecution(executionId: string | null) {
  return useQuery({
    queryKey: [EXECUTION_QUERY_KEY, executionId],
    queryFn: () => executionsApi.getExecutionWithLogs(executionId || ''),
    enabled: !!executionId,
    refetchInterval: (data) => {
      const executionStatus = data.state?.data?.status
      if (executionStatus === 'RUNNING' || executionStatus === 'PENDING') {
        return 1000
      }
      return false
    },
  })
}
