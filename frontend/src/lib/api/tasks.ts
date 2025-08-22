import axios from 'axios'
import { API_CONFIG, API_ENDPOINTS } from '@/config/api'
import type {
  Task,
  TaskStatus,
  CreateTaskRequest,
  UpdateTaskRequest,
  TasksResponse,
  TaskFilters,
  StartPlanningRequest,
  StartPlanningResponse,
  ApprovePlanRequest,
  TaskPlansResponse,
} from '@/types/task'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const tasksApi = {
  async getTasks(
    projectId: string,
    filters?: TaskFilters & { include_done?: boolean }
  ): Promise<TasksResponse> {
    const params = new URLSearchParams()
    if (filters?.include_done) {
      params.append('include_done', 'true')
    }
    // Switch to project-scoped endpoint and exclude DONE by default
    const response = await api.get(
      `/projects/${projectId}/tasks${params.toString() ? `?${params.toString()}` : ''}`
    )
    return response.data
  },

  async getDoneTasks(projectId: string): Promise<TasksResponse> {
    const response = await api.get(`/projects/${projectId}/tasks/done`)
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

  async changeTaskStatus(taskId: string, status: TaskStatus): Promise<Task> {
    const response = await api.put(`${API_ENDPOINTS.TASKS}/${taskId}`, {
      status,
    })
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

  async approvePlan(
    taskId: string,
    request: ApprovePlanRequest
  ): Promise<StartPlanningResponse> {
    const response = await api.post(
      `${API_ENDPOINTS.TASKS}/${taskId}/approve-plan`,
      request
    )
    return response.data
  },

  async openWithCursor(taskId: string): Promise<void> {
    await api.post(`${API_ENDPOINTS.TASKS}/${taskId}/open-with-cursor`)
  },

  async getTaskPlans(taskId: string): Promise<TaskPlansResponse> {
    const response = await api.get(`${API_ENDPOINTS.TASKS}/${taskId}/plans`)
    return response.data
  },

  async updatePlan(
    taskId: string,
    planId: string,
    content: string
  ): Promise<void> {
    await api.put(`${API_ENDPOINTS.TASKS}/${taskId}/plans/${planId}`, {
      content,
    })
  },

  async getTaskDiff(taskId: string): Promise<string> {
    const response = await api.get(`${API_ENDPOINTS.TASKS}/${taskId}/diff`, {
      responseType: 'text',
    })
    return response.data
  },

  async createPullRequestForTask(taskId: string): Promise<any> {
    const response = await api.post(
      `${API_ENDPOINTS.TASKS}/${taskId}/pull-request`
    )
    return response.data
  },
}
