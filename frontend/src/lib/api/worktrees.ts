import axios from 'axios'
import { API_CONFIG } from '@/config/api'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
})

export const worktreesApi = {
  async createWorktree(params: {
    taskId: string
    projectId: string
    taskTitle: string
    baseBranchName: string
    useRemoteBranch?: boolean
  }) {
    const response = await api.post('/worktrees', {
      task_id: params.taskId,
      project_id: params.projectId,
      task_title: params.taskTitle,
      base_branch_name: params.baseBranchName,
      use_remote_branch: params.useRemoteBranch ?? false,
    })
    return response.data
  },
}
