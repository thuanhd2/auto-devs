export type PullRequestStatus = 'OPEN' | 'MERGED' | 'CLOSED'

export interface PullRequestComment {
  id: string
  pull_request_id: string
  github_id?: number
  author: string
  body: string
  file_path?: string
  line?: number
  is_resolved: boolean
  created_at: string
  updated_at: string
}

export interface PullRequestReview {
  id: string
  pull_request_id: string
  github_id?: number
  reviewer: string
  state: 'APPROVED' | 'CHANGES_REQUESTED' | 'COMMENTED'
  body?: string
  submitted_at?: string
  created_at: string
  updated_at: string
}

export interface PullRequestCheck {
  id: string
  pull_request_id: string
  check_name: string
  status: 'PENDING' | 'SUCCESS' | 'FAILURE' | 'ERROR'
  conclusion?: string
  details_url?: string
  started_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface PullRequest {
  id: string
  task_id: string
  github_pr_number: number
  repository: string
  title: string
  body: string
  status: PullRequestStatus
  head_branch: string
  base_branch: string
  github_url: string
  merge_commit_sha?: string
  merged_at?: string
  closed_at?: string
  created_by?: string
  merged_by?: string
  reviewers: string[]
  labels: string[]
  assignees: string[]
  is_draft: boolean
  mergeable?: boolean
  mergeable_state?: string
  additions?: number
  deletions?: number
  changed_files?: number
  created_at: string
  updated_at: string
  // Relationships
  task?: {
    id: string
    title: string
    status: string
  }
  comments?: PullRequestComment[]
  reviews?: PullRequestReview[]
  checks?: PullRequestCheck[]
}

export interface PullRequestFilters {
  status?: PullRequestStatus[]
  repository?: string
  search?: string
  created_by?: string
  assignee?: string
  label?: string
  sortBy?: 'created_at' | 'updated_at' | 'title' | 'status' | 'github_pr_number'
  sortOrder?: 'asc' | 'desc'
}

export interface PullRequestsResponse {
  pull_requests: PullRequest[]
  total: number
  page: number
  limit: number
}

export interface CreatePullRequestRequest {
  task_id: string
  github_pr_number: number
  repository: string
  title: string
  body?: string
  head_branch: string
  base_branch?: string
  github_url: string
}

export interface UpdatePullRequestRequest {
  title?: string
  body?: string
  status?: PullRequestStatus
  merged_at?: string
  closed_at?: string
  merged_by?: string
  reviewers?: string[]
  labels?: string[]
  assignees?: string[]
  is_draft?: boolean
  mergeable?: boolean
  mergeable_state?: string
  additions?: number
  deletions?: number
  changed_files?: number
}