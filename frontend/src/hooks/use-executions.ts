import { useEffect, useState, useCallback } from 'react'
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
  ExecutionLog,
} from '@/types/execution'
import { toast } from 'sonner'
import { executionsApi } from '@/lib/api/executions'
import { useWebSocketContext } from '@/context/websocket-context'

const EXECUTIONS_QUERY_KEY = 'executions'
const EXECUTION_QUERY_KEY = 'execution'
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

// Get execution logs for a specific execution
function useExecutionLogsWithWebSocket(executionId: string) {
  const [logs, setLogs] = useState<ExecutionLog[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  useEffect(() => {
    if (!executionId) {
      return
    }
    const fetchLogs = async () => {
      const logs = await executionsApi.getExecutionLogs(executionId)
      setLogs(logs.data)
      setIsLoading(false)
    }
    try {
      fetchLogs()
    } catch (error) {
      setError(error as string)
    }
  }, [executionId])
  const newLogCreated = useCallback(
    (message: CentrifugeMessage) => {
      const { execution_id: incomingExecutionId, logs: incomingLogs } =
        message.data
      if (incomingExecutionId !== executionId) {
        return
      }
      const newLogs = incomingLogs.filter(
        (log) => !logs.some((l) => l.id === log.id)
      )
      setLogs((prev) => [...prev, ...newLogs])
    },
    [executionId, logs, setLogs]
  )
  const { subscribe, unsubscribe } = useWebSocketContext()
  useEffect(() => {
    subscribe('execution_log_created', newLogCreated)
    return () => unsubscribe('execution_log_created')
  }, [executionId, newLogCreated, subscribe, unsubscribe])
  return { logs, isLoading, error }
}

export function useExecution(executionId: string) {
  return useQuery({
    queryKey: [EXECUTION_QUERY_KEY, executionId],
    queryFn: () => executionsApi.getExecution(executionId, true, 1000),
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
