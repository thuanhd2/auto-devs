import axios from 'axios'
import { API_CONFIG, API_ENDPOINTS } from '@/config/api'
import type {
  PullRequest,
  PullRequestsResponse,
  PullRequestFilters,
  CreatePullRequestRequest,
  UpdatePullRequestRequest,
} from '@/types/pull-request'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const pullRequestsApi = {
  async getPullRequests(
    projectId: string,
    filters?: PullRequestFilters
  ): Promise<PullRequestsResponse> {
    const params = new URLSearchParams()
    params.append('project_id', projectId)

    if (filters) {
      if (filters.status && filters.status.length > 0) {
        filters.status.forEach((status) => params.append('status', status))
      }
      if (filters.repository) {
        params.append('repository', filters.repository)
      }
      if (filters.search) {
        params.append('search', filters.search)
      }
      if (filters.created_by) {
        params.append('created_by', filters.created_by)
      }
      if (filters.assignee) {
        params.append('assignee', filters.assignee)
      }
      if (filters.label) {
        params.append('label', filters.label)
      }
      if (filters.sortBy) {
        params.append('sort_by', filters.sortBy)
      }
      if (filters.sortOrder) {
        params.append('sort_order', filters.sortOrder)
      }
    }

    const response = await api.get(`${API_ENDPOINTS.PULL_REQUESTS}?${params.toString()}`)
    return response.data
  },

  async getPullRequest(pullRequestId: string): Promise<PullRequest> {
    const response = await api.get(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}`)
    return response.data
  },

  async getPullRequestByTask(taskId: string): Promise<PullRequest | null> {
    try {
      const response = await api.get(`${API_ENDPOINTS.TASKS}/${taskId}/pull-request`)
      return response.data
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        return null
      }
      throw error
    }
  },

  async createPullRequest(pullRequest: CreatePullRequestRequest): Promise<PullRequest> {
    const response = await api.post(API_ENDPOINTS.PULL_REQUESTS, pullRequest)
    return response.data
  },

  async updatePullRequest(
    pullRequestId: string,
    updates: UpdatePullRequestRequest
  ): Promise<PullRequest> {
    const response = await api.put(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}`, updates)
    return response.data
  },

  async deletePullRequest(pullRequestId: string): Promise<void> {
    await api.delete(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}`)
  },

  async syncPullRequest(pullRequestId: string): Promise<PullRequest> {
    const response = await api.post(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}/sync`)
    return response.data
  },

  async mergePullRequest(pullRequestId: string, mergeMethod?: 'merge' | 'squash' | 'rebase'): Promise<PullRequest> {
    const response = await api.post(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}/merge`, {
      merge_method: mergeMethod || 'merge'
    })
    return response.data
  },

  async closePullRequest(pullRequestId: string): Promise<PullRequest> {
    const response = await api.post(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}/close`)
    return response.data
  },

  async reopenPullRequest(pullRequestId: string): Promise<PullRequest> {
    const response = await api.post(`${API_ENDPOINTS.PULL_REQUESTS}/${pullRequestId}/reopen`)
    return response.data
  },
}