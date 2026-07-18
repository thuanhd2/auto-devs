import { Tool } from '@modelcontextprotocol/sdk/types.js';
import AutoDevsClient from '../client/autodevs-client.js';

const client = new AutoDevsClient();

export const taskListTool: Tool = {
  name: 'task:list',
  description: 'List tasks by project with optional filters',
  inputSchema: {
    type: 'object',
    properties: {
      projectId: {
        type: 'string',
        description: 'Project ID',
      },
      page: {
        type: 'number',
        description: 'Page number (default: 1)',
      },
      pageSize: {
        type: 'number',
        description: 'Items per page (default: 10)',
      },
      status: {
        type: 'string',
        description: 'Filter by status (e.g., pending, in_progress, completed)',
      },
      priority: {
        type: 'string',
        description: 'Filter by priority (e.g., low, medium, high)',
      },
    },
    required: ['projectId'],
  },
};

export async function executeTaskList(input: Record<string, unknown>): Promise<string> {
  const projectId = input.projectId as string;
  const page = (input.page as number) || 1;
  const pageSize = (input.pageSize as number) || 10;
  const filters = {
    status: input.status,
    priority: input.priority,
  };

  const result = await client.listTasks(projectId, { page, pageSize, ...filters });
  return JSON.stringify(result, null, 2);
}

export const taskCreateTool: Tool = {
  name: 'task:create',
  description: 'Create a new task in a project',
  inputSchema: {
    type: 'object',
    properties: {
      projectId: {
        type: 'string',
        description: 'Project ID',
      },
      title: {
        type: 'string',
        description: 'Task title',
      },
      description: {
        type: 'string',
        description: 'Task description',
      },
      priority: {
        type: 'string',
        description: 'Priority level (low, medium, high)',
      },
    },
    required: ['projectId', 'title'],
  },
};

export async function executeTaskCreate(input: Record<string, unknown>): Promise<string> {
  const projectId = input.projectId as string;
  const taskData = {
    title: input.title as string,
    description: input.description as string,
    priority: input.priority as string,
  };

  const result = await client.createTask(projectId, taskData);
  return JSON.stringify(result, null, 2);
}

export const taskUpdateStatusTool: Tool = {
  name: 'task:update-status',
  description: 'Update task status',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: {
        type: 'string',
        description: 'Task ID',
      },
      status: {
        type: 'string',
        description: 'New status (pending, in_progress, completed, blocked)',
      },
    },
    required: ['taskId', 'status'],
  },
};

export async function executeTaskUpdateStatus(input: Record<string, unknown>): Promise<string> {
  const taskId = input.taskId as string;
  const status = input.status as string;

  const result = await client.updateTaskStatus(taskId, status);
  return JSON.stringify(result, null, 2);
}

export const taskGetTool: Tool = {
  name: 'task:get',
  description: 'Get full task details with history and linked executions',
  inputSchema: {
    type: 'object',
    properties: {
      id: {
        type: 'string',
        description: 'Task ID',
      },
    },
    required: ['id'],
  },
};

export async function executeTaskGet(input: Record<string, unknown>): Promise<string> {
  const id = input.id as string;
  const result = await client.getTask(id);
  return JSON.stringify(result, null, 2);
}

export const taskDeleteTool: Tool = {
  name: 'task:delete',
  description: 'Delete a task',
  inputSchema: {
    type: 'object',
    properties: {
      id: {
        type: 'string',
        description: 'Task ID',
      },
    },
    required: ['id'],
  },
};

export async function executeTaskDelete(input: Record<string, unknown>): Promise<string> {
  const id = input.id as string;
  await client.deleteTask(id);
  return JSON.stringify({ success: true, message: `Task ${id} deleted` });
}
