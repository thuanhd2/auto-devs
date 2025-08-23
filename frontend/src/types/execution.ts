export type ExecutionStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'PAUSED'
  | 'COMPLETED'
  | 'FAILED'
  | 'CANCELLED'

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface ExecutionResult {
  output: string
  files: string[]
  metrics: Record<string, unknown>
  duration: number // in nanoseconds
}

export interface Execution {
  id: string
  task_id: string
  status: ExecutionStatus
  started_at: string
  completed_at?: string
  error?: string
  progress: number // 0.0 to 1.0
  result?: ExecutionResult
  duration?: number // in nanoseconds
  created_at: string
  updated_at: string
}

export interface ExecutionLog {
  id: string
  execution_id: string
  process_id?: string
  level: LogLevel
  message: string
  timestamp: string
  source: string
  metadata?: unknown
  created_at: string
  line: number
  // New structured fields
  log_type?: string
  tool_name?: string
  tool_use_id?: string
  parsed_content?: any
  is_error?: boolean
  duration_ms?: number
  num_turns?: number
}

export interface ExecutionWithLogs extends Execution {
  logs: ExecutionLog[]
}

export interface ExecutionStats {
  total_executions: number
  completed_executions: number
  failed_executions: number
  average_progress: number
  average_duration: number
  status_distribution: Record<ExecutionStatus, number>
  recent_activity: Execution[]
}

export interface LogStats {
  total_logs: number
  logs_by_level: Record<LogLevel, number>
  logs_by_source: Record<string, number>
  error_count: number
  warning_count: number
  first_log_time?: string
  last_log_time?: string
  recent_error_logs: ExecutionLog[]
  log_size_bytes: number
}

// Request types
export interface CreateExecutionRequest {
  task_id: string
}

export interface UpdateExecutionRequest {
  status?: ExecutionStatus
  progress?: number
  error?: string
}

// Filter types
export interface ExecutionFilters {
  status?: ExecutionStatus
  statuses?: ExecutionStatus[]
  started_after?: string
  started_before?: string
  with_errors?: boolean
  page?: number
  page_size?: number
  order_by?: 'started_at' | 'completed_at' | 'progress' | 'status'
  order_dir?: 'asc' | 'desc'
}

export interface ExecutionLogFilters {
  level?: LogLevel
  levels?: LogLevel[]
  source?: string
  sources?: string[]
  search?: string
  time_after?: string
  time_before?: string
  page?: number
  page_size?: number
  order_by?: 'timestamp' | 'level' | 'source'
  order_dir?: 'asc' | 'desc'
}

// Response types
export interface ExecutionListResponse {
  data: Execution[]
  meta: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface ExecutionLogListResponse {
  data: ExecutionLog[]
  meta: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

// Status colors for UI
export const EXECUTION_STATUS_COLORS: Record<ExecutionStatus, string> = {
  PENDING: 'bg-gray-100 text-gray-800',
  RUNNING: 'bg-blue-100 text-blue-800',
  PAUSED: 'bg-yellow-100 text-yellow-800',
  COMPLETED: 'bg-green-100 text-green-800',
  FAILED: 'bg-red-100 text-red-800',
  CANCELLED: 'bg-gray-100 text-gray-600',
}
