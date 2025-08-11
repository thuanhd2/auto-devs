import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from '@tanstack/react-query'
import type {
  Execution,
  ExecutionListResponse,
  ExecutionLogListResponse,
  UpdateExecutionRequest,
  ExecutionFilters,
  ExecutionLogFilters,
} from '@/types/execution'
import { toast } from 'sonner'
import { executionsApi } from '@/lib/api/executions'

const EXECUTIONS_QUERY_KEY = 'executions'
const EXECUTION_LOGS_QUERY_KEY = 'execution-logs'
const EXECUTION_STATS_QUERY_KEY = 'execution-stats'

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
