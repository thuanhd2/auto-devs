import axios from 'axios'
import { API_CONFIG, API_ENDPOINTS } from '@/config/api'
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectsResponse,
  ProjectFilters,
  ProjectStatistics,
} from '@/types/project'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const projectsApi = {
  async getProjects(filters?: ProjectFilters): Promise<ProjectsResponse> {
    const params = new URLSearchParams()

    if (filters) {
      if (filters.search) {
        params.append('search', filters.search)
      }
      if (filters.sortBy) {
        params.append('sort_by', filters.sortBy)
      }
      if (filters.sortOrder) {
        params.append('sort_order', filters.sortOrder)
      }
      if (filters.archived !== undefined) {
        params.append('archived', filters.archived.toString())
      }
    }

    // Default to show only non-archived projects if not specified
    if (!params.has('archived')) {
      params.append('archived', 'false')
    }

    const response = await api.get(
      `${API_ENDPOINTS.PROJECTS}?${params.toString()}`
    )
    return response.data
  },

  async getProject(projectId: string): Promise<Project> {
    const response = await api.get(`${API_ENDPOINTS.PROJECTS}/${projectId}`)
    return response.data
  },

  async getProjectStatistics(projectId: string): Promise<ProjectStatistics> {
    const response = await api.get(
      `${API_ENDPOINTS.PROJECTS}/${projectId}/statistics`
    )
    return response.data
  },

  async createProject(project: CreateProjectRequest): Promise<Project> {
    const response = await api.post(API_ENDPOINTS.PROJECTS, project)
    return response.data
  },

  async updateProject(
    projectId: string,
    updates: UpdateProjectRequest
  ): Promise<Project> {
    const response = await api.put(
      `${API_ENDPOINTS.PROJECTS}/${projectId}`,
      updates
    )
    return response.data
  },

  async deleteProject(projectId: string): Promise<void> {
    await api.delete(`${API_ENDPOINTS.PROJECTS}/${projectId}`)
  },

  async archiveProject(projectId: string): Promise<void> {
    await api.post(`${API_ENDPOINTS.PROJECTS}/${projectId}/archive`)
  },

  async restoreProject(projectId: string): Promise<void> {
    await api.post(`${API_ENDPOINTS.PROJECTS}/${projectId}/restore`)
  },

  async reinitGitRepository(projectId: string): Promise<void> {
    await api.post(`${API_ENDPOINTS.PROJECTS}/${projectId}/git/reinit`)
  },

  async getProjectBranches(projectId: string): Promise<{
    branches: Array<{
      name: string
      is_current: boolean
      last_commit: string
      last_updated: string
    }>
    total: number
  }> {
    const response = await api.get(
      `${API_ENDPOINTS.PROJECTS}/${projectId}/branches`
    )
    return response.data
  },
}
