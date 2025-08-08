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

export const EXECUTIONS_QUERY_KEY = 'executions'
export const EXECUTION_LOGS_QUERY_KEY = 'execution-logs'
export const EXECUTION_STATS_QUERY_KEY = 'execution-stats'

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

// Get single execution by ID
export function useExecution(
  executionId: string,
  includeLogs: boolean = false,
  logLimit: number = 100
) {
  return useQuery({
    queryKey: [EXECUTIONS_QUERY_KEY, executionId, { includeLogs, logLimit }],
    queryFn: () =>
      executionsApi.getExecution(executionId, includeLogs, logLimit),
    enabled: !!executionId,
    refetchInterval: (_data) => {
      // Auto-refetch every 3 seconds if execution is active
      // const execution = _data as Execution
      // const isActive =
      //   execution?.status === 'running' || execution?.status === 'pending'
      // return isActive ? 3000 : false
      return false
    },
  })
}

// Get execution logs with infinite query for pagination
export function useExecutionLogs(
  executionId: string,
  filters?: ExecutionLogFilters
) {
  return useInfiniteQuery({
    queryKey: [EXECUTION_LOGS_QUERY_KEY, executionId, filters],
    queryFn: ({ pageParam = 1 }) =>
      executionsApi.getExecutionLogs(executionId, {
        ...filters,
        page: pageParam,
        page_size: filters?.page_size || 50,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const { page, total_pages } = lastPage.meta
      return page < total_pages ? page + 1 : undefined
    },
    enabled: !!executionId,
  })
}

// Get execution logs with standard query (for simpler cases)
export function useExecutionLogsSimple(
  executionId: string,
  filters?: ExecutionLogFilters
) {
  return useQuery({
    queryKey: [EXECUTION_LOGS_QUERY_KEY, executionId, filters],
    queryFn: () => executionsApi.getExecutionLogs(executionId, filters),
    enabled: !!executionId,
    refetchInterval: (_data) => {
      // Auto-refetch every 2 seconds for logs when execution is active
      // This would typically be handled by WebSocket in production
      return false
    },
  })
}

// Get execution statistics
export function useExecutionStats(taskId?: string) {
  return useQuery({
    queryKey: [EXECUTION_STATS_QUERY_KEY, taskId],
    queryFn: () => executionsApi.getExecutionStats(taskId),
    refetchInterval: 30000, // Refetch stats every 30 seconds
  })
}

// Create new execution
export function useCreateExecution() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: executionsApi.createExecution,
    onMutate: async (_newExecution) => {
      // Optimistic update could be added here
      return { _newExecution }
    },
    onSuccess: (_execution, _variables) => {
      // Invalidate and refetch task executions
      queryClient.invalidateQueries({
        queryKey: [EXECUTIONS_QUERY_KEY, 'task', _variables.task_id],
      })

      // Invalidate execution stats
      queryClient.invalidateQueries({
        queryKey: [EXECUTION_STATS_QUERY_KEY],
      })

      toast.success('Execution started successfully')
    },
    onError: (error: unknown) => {
      const errorMessage =
        error instanceof Error
          ? error.message
          : (error as { response?: { data?: { message?: string } } })?.response
              ?.data?.message || 'Failed to start execution'
      toast.error(errorMessage)
    },
  })
}

// Update execution
export function useUpdateExecution() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      executionId,
      data,
    }: {
      executionId: string
      data: UpdateExecutionRequest
    }) => executionsApi.updateExecution(executionId, data),
    onMutate: async ({ executionId, data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({
        queryKey: [EXECUTIONS_QUERY_KEY, executionId],
      })

      // Snapshot previous value
      const previousExecution = queryClient.getQueryData([
        EXECUTIONS_QUERY_KEY,
        executionId,
      ])

      // Optimistically update
      queryClient.setQueryData(
        [EXECUTIONS_QUERY_KEY, executionId],
        (old: Execution) => ({
          ...old,
          ...data,
          updated_at: new Date().toISOString(),
        })
      )

      return { previousExecution, executionId }
    },
    onSuccess: (updatedExecution, { executionId }) => {
      // Update the execution query
      queryClient.setQueryData(
        [EXECUTIONS_QUERY_KEY, executionId],
        updatedExecution
      )

      // Invalidate task executions list
      queryClient.invalidateQueries({
        queryKey: [EXECUTIONS_QUERY_KEY, 'task', updatedExecution.task_id],
      })

      // Invalidate stats
      queryClient.invalidateQueries({
        queryKey: [EXECUTION_STATS_QUERY_KEY],
      })

      toast.success('Execution updated successfully')
    },
    onError: (error: unknown, _variables, context) => {
      // Revert optimistic update on error
      if (context?.previousExecution) {
        queryClient.setQueryData(
          [EXECUTIONS_QUERY_KEY, context.executionId],
          context.previousExecution
        )
      }
      const errorMessage =
        error instanceof Error
          ? error.message
          : (error as { response?: { data?: { message?: string } } })?.response
              ?.data?.message || 'Failed to update execution'
      toast.error(errorMessage)
    },
  })
}

// Delete execution
export function useDeleteExecution() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: executionsApi.deleteExecution,
    onMutate: async (executionId) => {
      // Get the execution to know which task to invalidate
      const execution = queryClient.getQueryData([
        EXECUTIONS_QUERY_KEY,
        executionId,
      ]) as Execution

      // Optimistically remove from task executions
      if (execution) {
        queryClient.setQueryData(
          [EXECUTIONS_QUERY_KEY, 'task', execution.task_id],
          (old: ExecutionListResponse) => ({
            ...old,
            data: old.data.filter((e) => e.id !== executionId),
            meta: { ...old.meta, total: old.meta.total - 1 },
          })
        )
      }

      return { execution, executionId }
    },
    onSuccess: (_, executionId, context) => {
      // Remove from individual execution cache
      queryClient.removeQueries({
        queryKey: [EXECUTIONS_QUERY_KEY, executionId],
      })
      queryClient.removeQueries({
        queryKey: [EXECUTION_LOGS_QUERY_KEY, executionId],
      })

      // Invalidate task executions if we have the task ID
      if (context?.execution) {
        queryClient.invalidateQueries({
          queryKey: [EXECUTIONS_QUERY_KEY, 'task', context.execution.task_id],
        })
      }

      // Invalidate stats
      queryClient.invalidateQueries({
        queryKey: [EXECUTION_STATS_QUERY_KEY],
      })

      toast.success('Execution deleted successfully')
    },
    onError: (error: unknown, _executionId, context) => {
      // Revert optimistic deletion
      if (context?.execution) {
        queryClient.invalidateQueries({
          queryKey: [EXECUTIONS_QUERY_KEY, 'task', context.execution.task_id],
        })
      }
      const errorMessage =
        error instanceof Error
          ? error.message
          : (error as { response?: { data?: { message?: string } } })?.response
              ?.data?.message || 'Failed to delete execution'
      toast.error(errorMessage)
    },
  })
}

// Real-time hooks for WebSocket integration
export function useExecutionRealTime(executionId: string) {
  const queryClient = useQueryClient()

  // This would typically set up WebSocket subscriptions
  // For now, we'll use the refetchInterval in the main query
  return {
    subscribeToUpdates: () => {
      // Subscribe to execution updates
      return executionsApi.subscribeToExecutionUpdates(
        executionId,
        (updatedExecution) => {
          queryClient.setQueryData(
            [EXECUTIONS_QUERY_KEY, executionId],
            updatedExecution
          )
        },
        (_error) => {
          // console.error('WebSocket error:', error)
          toast.error('Real-time connection lost')
        }
      )
    },
    subscribeToLogs: () => {
      // Subscribe to new logs
      return executionsApi.subscribeToLogUpdates(
        executionId,
        (newLog) => {
          queryClient.setQueryData(
            [EXECUTION_LOGS_QUERY_KEY, executionId],
            (old: ExecutionLogListResponse) => ({
              ...old,
              data: [newLog, ...old.data],
              meta: { ...old.meta, total: old.meta.total + 1 },
            })
          )
        },
        (_error) => {
          // console.error('Log WebSocket error:', error)
        }
      )
    },
  }
}

// Utility hooks
export function useActiveExecutions(taskId?: string) {
  return useQuery({
    queryKey: [EXECUTIONS_QUERY_KEY, 'active', taskId],
    queryFn: () =>
      executionsApi.getTaskExecutions(taskId || '', {
        statuses: ['running', 'pending'],
        page: 1,
        page_size: 100,
      }),
    enabled: !!taskId,
    refetchInterval: false, // Check active executions every 5 seconds
  })
}

export function useExecutionHistory(taskId: string, limit: number = 10) {
  return useQuery({
    queryKey: [EXECUTIONS_QUERY_KEY, 'history', taskId, limit],
    queryFn: () =>
      executionsApi.getTaskExecutions(taskId, {
        statuses: ['completed', 'failed', 'cancelled'],
        page: 1,
        page_size: limit,
        order_by: 'completed_at',
        order_dir: 'desc',
      }),
    enabled: !!taskId,
  })
}
