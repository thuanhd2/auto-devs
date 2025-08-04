export interface Project {
  id: string
  name: string
  description: string
  repo_url: string
  created_at: string
  updated_at: string
  
  // Git-related fields
  repository_url?: string
  main_branch?: string
  worktree_base_path?: string
  git_auth_method?: string
  git_enabled?: boolean
}

export interface CreateProjectRequest {
  name: string
  description?: string
  worktree_base_path: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
  repo_url?: string
  
  // Git-related fields
  repository_url?: string
  main_branch?: string
  worktree_base_path?: string
  git_auth_method?: string
  git_enabled?: boolean
}

// Git-specific types
export interface GitProjectValidationRequest {
  repository_url: string
  main_branch: string
  worktree_base_path: string
  git_auth_method: string
  git_enabled: boolean
}

export interface GitProjectValidationResponse {
  valid: boolean
  message?: string
  errors?: string[]
}

export interface GitProjectStatusResponse {
  git_enabled: boolean
  worktree_exists: boolean
  repository_valid: boolean
  current_branch?: string
  remote_url?: string
  on_main_branch: boolean
  working_dir_status?: WorkingDirStatus
  status: string
}

export interface WorkingDirStatus {
  is_clean: boolean
  has_staged_changes: boolean
  has_unstaged_changes: boolean
  has_untracked_files: boolean
}

export interface ProjectFilters {
  search?: string
  sortBy?: 'created_at' | 'updated_at' | 'name'
  sortOrder?: 'asc' | 'desc'
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
