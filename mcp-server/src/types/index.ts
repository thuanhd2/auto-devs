export interface Project {
  id: string;
  name: string;
  description: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: string;
  projectId: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  kanban_task_id?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Execution {
  id: string;
  taskId: string;
  status: string;
  startedAt: string;
  completedAt?: string;
  result?: string;
}

export interface Worktree {
  id: string;
  projectId: string;
  path: string;
  branch: string;
  status: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}
