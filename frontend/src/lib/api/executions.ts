import axios from 'axios'
import { API_CONFIG, API_ENDPOINTS } from '@/config/api'
import type {
  Execution,
  ExecutionWithLogs,
  ExecutionLog,
  ExecutionListResponse,
  ExecutionLogListResponse,
  ExecutionStats,
  LogStats,
  CreateExecutionRequest,
  UpdateExecutionRequest,
  ExecutionFilters,
  ExecutionLogFilters,
} from '@/types/execution'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const executionsApi = {
  // Get all executions for a task
  async getTaskExecutions(
    taskId: string,
    filters?: ExecutionFilters
  ): Promise<ExecutionListResponse> {
    const params = new URLSearchParams()

    if (filters) {
      if (filters.status) {
        params.append('status', filters.status)
      }
      if (filters.statuses && filters.statuses.length > 0) {
        filters.statuses.forEach((status) => params.append('statuses', status))
      }
      if (filters.started_after) {
        params.append('started_after', filters.started_after)
      }
      if (filters.started_before) {
        params.append('started_before', filters.started_before)
      }
      if (filters.with_errors !== undefined) {
        params.append('with_errors', filters.with_errors.toString())
      }
      if (filters.page) {
        params.append('page', filters.page.toString())
      }
      if (filters.page_size) {
        params.append('page_size', filters.page_size.toString())
      }
      if (filters.order_by) {
        params.append('order_by', filters.order_by)
      }
      if (filters.order_dir) {
        params.append('order_dir', filters.order_dir)
      }
    }

    const response = await api.get(
      `${API_ENDPOINTS.TASKS}/${taskId}/executions?${params.toString()}`
    )
    return response.data
  },

  // Get single execution by ID
  async getExecution(
    executionId: string,
    includeLogs: boolean = false,
    logLimit: number = 100
  ): Promise<Execution | ExecutionWithLogs> {
    const params = new URLSearchParams()
    if (includeLogs) {
      params.append('include_logs', 'true')
      params.append('log_limit', logLimit.toString())
    }

    const response = await api.get(
      `${API_ENDPOINTS.EXECUTIONS}/${executionId}?${params.toString()}`
    )
    return response.data
  },

  // Get execution logs with pagination and filtering
  async getExecutionLogs(
    executionId: string,
    filters?: ExecutionLogFilters
  ): Promise<ExecutionLogListResponse> {
    const params = new URLSearchParams()

    if (filters) {
      if (filters.level) {
        params.append('level', filters.level)
      }
      if (filters.levels && filters.levels.length > 0) {
        filters.levels.forEach((level) => params.append('levels', level))
      }
      if (filters.source) {
        params.append('source', filters.source)
      }
      if (filters.sources && filters.sources.length > 0) {
        filters.sources.forEach((source) => params.append('sources', source))
      }
      if (filters.search) {
        params.append('search', filters.search)
      }
      if (filters.time_after) {
        params.append('time_after', filters.time_after)
      }
      if (filters.time_before) {
        params.append('time_before', filters.time_before)
      }
      if (filters.page) {
        params.append('page', filters.page.toString())
      }
      if (filters.page_size) {
        params.append('page_size', filters.page_size.toString())
      }
      if (filters.order_by) {
        params.append('order_by', filters.order_by)
      }
      if (filters.order_dir) {
        params.append('order_dir', filters.order_dir)
      }
    }

    const response = await api.get(
      `${API_ENDPOINTS.EXECUTIONS}/${executionId}/logs?${params.toString()}`
    )
    return response.data
  },

  // Create new execution
  async createExecution(data: CreateExecutionRequest): Promise<Execution> {
    const response = await api.post(API_ENDPOINTS.EXECUTIONS, data)
    return response.data
  },

  // Update execution
  async updateExecution(
    executionId: string,
    data: UpdateExecutionRequest
  ): Promise<Execution> {
    const response = await api.put(
      `${API_ENDPOINTS.EXECUTIONS}/${executionId}`,
      data
    )
    return response.data
  },

  // Delete execution
  async deleteExecution(executionId: string): Promise<void> {
    await api.delete(`${API_ENDPOINTS.EXECUTIONS}/${executionId}`)
  },

  // Get execution statistics
  async getExecutionStats(taskId?: string): Promise<ExecutionStats> {
    const params = new URLSearchParams()
    if (taskId) {
      params.append('task_id', taskId)
    }

    const response = await api.get(
      `${API_ENDPOINTS.EXECUTIONS}/stats?${params.toString()}`
    )
    return response.data
  },

  // Get log statistics
  async getLogStats(executionId: string): Promise<LogStats> {
    const response = await api.get(
      `${API_ENDPOINTS.EXECUTIONS}/${executionId}/logs/stats`
    )
    return response.data
  },

  // Real-time execution updates (for WebSocket integration)
  subscribeToExecutionUpdates(
    _executionId: string,
    _onUpdate: (execution: Execution) => void,
    _onError?: (error: Error) => void
  ): () => void {
    // This would typically use WebSocket connection
    // For now, return a no-op cleanup function
    // Real implementation would establish WebSocket connection to:
    // ws://localhost:8098/ws/executions/{executionId}

    // Mock cleanup function
    return () => {
      // Cleanup WebSocket connection
    }
  },

  // Subscribe to execution log updates
  subscribeToLogUpdates(
    _executionId: string,
    _onLog: (log: ExecutionLog) => void,
    _onError?: (error: Error) => void
  ): () => void {
    // This would typically use WebSocket connection
    // For now, return a no-op cleanup function
    // Real implementation would establish WebSocket connection to:
    // ws://localhost:8098/ws/executions/{executionId}/logs

    // Mock cleanup function
    return () => {
      // Cleanup WebSocket connection
    }
  },
}
