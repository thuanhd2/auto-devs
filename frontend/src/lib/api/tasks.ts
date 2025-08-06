import axios from 'axios'
import { API_CONFIG, API_ENDPOINTS } from '@/config/api'
import type {
  Task,
  CreateTaskRequest,
  UpdateTaskRequest,
  TasksResponse,
  TaskFilters,
  StartPlanningRequest,
  StartPlanningResponse,
} from '@/types/task'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const tasksApi = {
  async getTasks(
    projectId: string,
    filters?: TaskFilters
  ): Promise<TasksResponse> {
    const params = new URLSearchParams()
    params.append('project_id', projectId)

    if (filters) {
      if (filters.status && filters.status.length > 0) {
        filters.status.forEach((status) => params.append('status', status))
      }
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

    const response = await api.get(
      `${API_ENDPOINTS.TASKS}?${params.toString()}`
    )
    return response.data
  },

  async getTask(taskId: string): Promise<Task> {
    const response = await api.get(`${API_ENDPOINTS.TASKS}/${taskId}`)
    return response.data
  },

  async createTask(task: CreateTaskRequest): Promise<Task> {
    const response = await api.post(API_ENDPOINTS.TASKS, task)
    return response.data
  },

  async updateTask(taskId: string, updates: UpdateTaskRequest): Promise<Task> {
    const response = await api.put(`${API_ENDPOINTS.TASKS}/${taskId}`, updates)
    return response.data
  },

  async deleteTask(taskId: string): Promise<void> {
    await api.delete(`${API_ENDPOINTS.TASKS}/${taskId}`)
  },

  async startPlanning(
    taskId: string,
    request: StartPlanningRequest
  ): Promise<StartPlanningResponse> {
    const response = await api.post(
      `${API_ENDPOINTS.TASKS}/${taskId}/start-planning`,
      request
    )
    return response.data
  },

  async approvePlan(taskId: string): Promise<StartPlanningResponse> {
    const response = await api.post(
      `${API_ENDPOINTS.TASKS}/${taskId}/approve-plan`
    )
    return response.data
  },
}
