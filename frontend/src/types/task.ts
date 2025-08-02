export type TaskStatus = 
  | 'TODO'
  | 'PLANNING'
  | 'PLAN_REVIEWING'
  | 'IMPLEMENTING'
  | 'CODE_REVIEWING'
  | 'DONE'
  | 'CANCELLED'

export interface Task {
  id: string
  project_id: string
  title: string
  description: string
  status: TaskStatus
  plan: string
  branch_name: string
  pr_url: string
  created_at: string
  updated_at: string
  completed_at?: string
}

export interface CreateTaskRequest {
  project_id: string
  title: string
  description?: string
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  status?: TaskStatus
  plan?: string
  branch_name?: string
  pr_url?: string
}

export interface TaskFilters {
  status?: TaskStatus[]
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'title'
  sortOrder?: 'asc' | 'desc'
}

export interface TasksResponse {
  tasks: Task[]
  total: number
  page: number
  limit: number
}