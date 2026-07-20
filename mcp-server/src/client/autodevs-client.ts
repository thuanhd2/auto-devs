import axios, { AxiosInstance, AxiosError } from 'axios';
import { config } from '../config.js';
import { AppError, ErrorCode } from '../errors/app-error.js';
import { withRetry } from '../utils/retry.js';
import { logger } from '../utils/logger.js';
import {
  Project,
  Task,
  Execution,
  Worktree,
  ListResponse,
  ApiResponse,
} from '../types/index.js';

export class AutoDevsClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: config.apiUrl,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        ...(config.apiKey && { Authorization: `Bearer ${config.apiKey}` }),
      },
    });

    this.client.interceptors.response.use(
      (response) => {
        logger.debug(`API call: ${response.config.method?.toUpperCase()} ${response.config.url} - ${response.status}`);
        return response;
      },
      (error) => {
        const msg = `API error: ${error.config?.method?.toUpperCase()} ${error.config?.url}`;
        logger.error(msg, { status: error.response?.status });
        return Promise.reject(error);
      }
    );
  }

  async listProjects(page = 1, pageSize = 10): Promise<ListResponse<Project>> {
    return withRetry(
      async () => {
        try {
          const response = await this.client.get('/api/v1/projects', {
            params: { page, pageSize },
          });
          return response.data;
        } catch (error) {
          throw this.handleError(error, 'Failed to list projects');
        }
      },
      { maxRetries: 3 }
    );
  }

  async getProject(id: string): Promise<Project> {
    return withRetry(async () => {
      try {
        const response = await this.client.get(`/api/v1/projects/${id}`);
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to get project ${id}`);
      }
    });
  }

  async listTasks(
    projectId: string,
    { page = 1, pageSize = 10, status, priority }: any = {}
  ): Promise<ListResponse<Task>> {
    return withRetry(async () => {
      try {
        const params: any = { page, pageSize };
        if (status) params.status = status;
        if (priority) params.priority = priority;

        const response = await this.client.get(`/api/v1/projects/${projectId}/tasks`, { params });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to list tasks for project ${projectId}`);
      }
    });
  }

  async createTask(projectId: string, data: Partial<Task>): Promise<Task> {
    return withRetry(async () => {
      try {
        const response = await this.client.post('/api/v1/tasks', {
          project_id: projectId,
          ...data,
        });
        return response.data;
      } catch (error) {
        throw this.handleError(error, 'Failed to create task');
      }
    });
  }

  async updateTaskStatus(taskId: string, status: string): Promise<Task> {
    return withRetry(async () => {
      try {
        const response = await this.client.put(`/api/v1/tasks/${taskId}`, { status });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to update task ${taskId} status`);
      }
    });
  }

  async listExecutions(
    taskId: string,
    { page = 1, pageSize = 10, status }: any = {}
  ): Promise<ListResponse<Execution>> {
    return withRetry(async () => {
      try {
        const params: any = { page, pageSize };
        if (status) params.status = status;

        const response = await this.client.get(`/api/v1/tasks/${taskId}/executions`, { params });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to list executions for task ${taskId}`);
      }
    });
  }

  async getWorktreeStatus(projectId: string): Promise<any> {
    return withRetry(async () => {
      try {
        const response = await this.client.get(`/api/v1/worktrees/project/${projectId}`);
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to get worktree status for project ${projectId}`);
      }
    });
  }

  async getTask(id: string): Promise<Task> {
    return withRetry(async () => {
      try {
        const response = await this.client.get(`/api/v1/tasks/${id}`);
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to get task ${id}`);
      }
    });
  }

  async deleteTask(id: string): Promise<void> {
    return withRetry(async () => {
      try {
        await this.client.delete(`/api/v1/tasks/${id}`);
      } catch (error) {
        throw this.handleError(error, `Failed to delete task ${id}`);
      }
    });
  }

  async getExecution(id: string): Promise<Execution> {
    return withRetry(async () => {
      try {
        const response = await this.client.get(`/api/v1/executions/${id}`);
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to get execution ${id}`);
      }
    });
  }

  async createExecution(taskId: string, data?: any): Promise<Execution> {
    return withRetry(async () => {
      try {
        const response = await this.client.post('/api/v1/executions', {
          task_id: taskId,
          ...data,
        });
        return response.data;
      } catch (error) {
        throw this.handleError(error, 'Failed to create execution');
      }
    });
  }

  async startPlanning(
    taskId: string,
    data: { branchName: string; aiType: string; autoImplement?: boolean; useRemoteBranch?: boolean }
  ): Promise<{ message: string; job_id: string }> {
    return withRetry(async () => {
      try {
        const response = await this.client.post(`/api/v1/tasks/${taskId}/start-planning`, {
          branch_name: data.branchName,
          ai_type: data.aiType,
          auto_implement: data.autoImplement ?? false,
          use_remote_branch: data.useRemoteBranch ?? false,
        });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to start planning for task ${taskId}`);
      }
    });
  }

  async approvePlan(
    taskId: string,
    data: { aiType: string }
  ): Promise<{ message: string; job_id: string }> {
    return withRetry(async () => {
      try {
        const response = await this.client.post(`/api/v1/tasks/${taskId}/approve-plan`, {
          ai_type: data.aiType,
        });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to approve plan for task ${taskId}`);
      }
    });
  }

  async startImplementingDirect(
    taskId: string,
    data: { branchName: string; aiType: string; useRemoteBranch?: boolean }
  ): Promise<{ message: string; job_id: string }> {
    return withRetry(async () => {
      try {
        const response = await this.client.post(`/api/v1/tasks/${taskId}/start-implementing-direct`, {
          branch_name: data.branchName,
          ai_type: data.aiType,
          use_remote_branch: data.useRemoteBranch ?? false,
        });
        return response.data;
      } catch (error) {
        throw this.handleError(error, `Failed to start implementing task ${taskId}`);
      }
    });
  }

  private handleError(error: unknown, defaultMessage: string): AppError {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status || 0;
      const apiError = error.response?.data?.error;
      const message = apiError || error.message || defaultMessage;

      let code: ErrorCode;
      if (status === 404) {
        code = ErrorCode.NOT_FOUND;
      } else if (status === 429) {
        code = ErrorCode.RATE_LIMITED;
      } else if (status === 503) {
        code = ErrorCode.SERVICE_UNAVAILABLE;
      } else if (status >= 400 && status < 500) {
        code = ErrorCode.VALIDATION_ERROR;
      } else {
        code = ErrorCode.INTERNAL_ERROR;
      }

      return new AppError(code, `${defaultMessage}: ${message}`, {
        details: { httpStatus: status, endpoint: error.config?.url },
        retryAfter: error.response?.headers['retry-after']
          ? parseInt(error.response.headers['retry-after'])
          : undefined,
      });
    }

    return new AppError(ErrorCode.INTERNAL_ERROR, defaultMessage);
  }
}

export default AutoDevsClient;
