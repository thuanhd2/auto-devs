export type TaskStatus =
  | 'TODO'
  | 'PLANNING'
  | 'PLAN_REVIEWING'
  | 'IMPLEMENTING'
  | 'CODE_REVIEWING'
  | 'DONE'
  | 'CANCELLED'
export type TaskGitStatus =
  | 'NO_GIT' // No Git worktree/branch
  | 'WORKTREE_PENDING' // Worktree creation requested but not created
  | 'WORKTREE_CREATED' // Worktree created successfully
  | 'BRANCH_CREATED' // Branch created in worktree
  | 'CHANGES_PENDING' // Has uncommitted changes
  | 'CHANGES_STAGED' // Has staged changes ready for commit
  | 'CHANGES_COMMITTED' // Changes committed to branch
  | 'PR_CREATED' // Pull request created
  | 'PR_MERGED' // Pull request merged
  | 'WORKTREE_ERROR' // Error with worktree operations

interface TaskGitInfo {
  status: TaskGitStatus
  branch_name?: string
  worktree_path?: string
  pr_url?: string
  has_uncommitted_changes?: boolean
  has_staged_changes?: boolean
  commits_ahead?: number
  commits_behind?: number
  last_sync?: string
}

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
  worktree_path?: string
  // Git information
  git_info?: TaskGitInfo
  plans?: TaskPlan[]
}

export interface TaskPlan {
  id: string
  task_id: string
  content: string
  created_at: string
  updated_at: string
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
  git_status?: TaskGitStatus[]
  search?: string
  branch_search?: string
  sortBy?: 'created_at' | 'updated_at' | 'title' | 'git_status'
  sortOrder?: 'asc' | 'desc'
}

export interface TasksResponse {
  tasks: Task[]
  total: number
  page: number
  limit: number
}

// Start Planning types
export interface StartPlanningRequest {
  branch_name: string
  ai_type: string
}

export interface StartPlanningResponse {
  message: string
  job_id: string
}

export interface ApprovePlanRequest {
  ai_type: string
}
