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
    }

    const response = await api.get(`${API_ENDPOINTS.PROJECTS}?${params.toString()}`)
    return response.data
  },

  async getProject(projectId: string): Promise<Project> {
    const response = await api.get(`${API_ENDPOINTS.PROJECTS}/${projectId}`)
    return response.data
  },

  async getProjectStatistics(projectId: string): Promise<ProjectStatistics> {
    const response = await api.get(`${API_ENDPOINTS.PROJECTS}/${projectId}/statistics`)
    return response.data
  },

  async createProject(project: CreateProjectRequest): Promise<Project> {
    const response = await api.post(API_ENDPOINTS.PROJECTS, project)
    return response.data
  },

  async updateProject(projectId: string, updates: UpdateProjectRequest): Promise<Project> {
    const response = await api.put(`${API_ENDPOINTS.PROJECTS}/${projectId}`, updates)
    return response.data
  },

  async deleteProject(projectId: string): Promise<void> {
    await api.delete(`${API_ENDPOINTS.PROJECTS}/${projectId}`)
  },
}