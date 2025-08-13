export type ExecutorType = 'claude-code' | 'fake-code'

export interface Project {
  id: string
  name: string
  description: string
  repository_url?: string
  worktree_base_path?: string
  init_workspace_script?: string
  executor_type: ExecutorType
  created_at: string
  updated_at: string
}

export interface CreateProjectRequest {
  name: string
  description?: string
  worktree_base_path: string
  init_workspace_script?: string
  executor_type?: ExecutorType
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
  repository_url?: string
  worktree_base_path?: string
  init_workspace_script?: string
  executor_type?: ExecutorType
}

export interface ProjectFilters {
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'name'
  sortOrder?: 'asc' | 'desc'
  archived?: boolean
}

export interface ProjectsResponse {
  projects: Project[]
  total: number
  page: number
  limit: number
}

export interface ProjectStatistics {
  total_tasks: number
  tasks_by_status: {
    TODO: number
    PLANNING: number
    PLAN_REVIEWING: number
    IMPLEMENTING: number
    CODE_REVIEWING: number
    DONE: number
    CANCELLED: number
  }
  recent_activity: number
}
