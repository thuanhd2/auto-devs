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
      kanban_task_id: {
        type: 'string',
        description: 'Hermes kanban card ID for callback',
      },
    },
    required: ['projectId', 'title'],
  },
};

export async function executeTaskCreate(input: Record<string, unknown>): Promise<string> {
  const projectId = input.projectId as string;
  // Backend expects snake_case field names (see SKILL.md "MCP Bug Fixed")
  const taskData: Record<string, unknown> = {
    title: input.title as string,
    description: input.description as string,
    priority: input.priority as string,
  };
  if (input.kanban_task_id) {
    taskData.kanban_task_id = input.kanban_task_id as string;
  }

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

export const taskStartPlanningTool: Tool = {
  name: 'task:start-planning',
  description: 'Start the planning phase for a task (queues a planning job)',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: { type: 'string', description: 'Task ID' },
      branchName: { type: 'string', description: 'Git branch name for the worktree' },
      aiType: { type: 'string', description: 'AI agent type (e.g., claude-code, gemini, cursor)' },
      useRemoteBranch: {
        type: 'boolean',
        description: 'Check out an existing remote branch (default: false)',
      },
      autoImplement: {
        type: 'boolean',
        description: 'Auto-start implementation after planning completes (default: false)',
      },
    },
    required: ['taskId', 'branchName', 'aiType'],
  },
};

export async function executeTaskStartPlanning(input: Record<string, unknown>): Promise<string> {
  const result = await client.startPlanning(input.taskId as string, {
    branchName: input.branchName as string,
    aiType: input.aiType as string,
    useRemoteBranch: input.useRemoteBranch as boolean | undefined,
    autoImplement: input.autoImplement as boolean | undefined,
  });
  return JSON.stringify(result, null, 2);
}

export const taskApprovePlanTool: Tool = {
  name: 'task:approve-plan',
  description: 'Approve a completed plan and start the implementation phase',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: { type: 'string', description: 'Task ID' },
      aiType: { type: 'string', description: 'AI agent type for implementation' },
    },
    required: ['taskId', 'aiType'],
  },
};

export async function executeTaskApprovePlan(input: Record<string, unknown>): Promise<string> {
  const result = await client.approvePlan(input.taskId as string, {
    aiType: input.aiType as string,
  });
  return JSON.stringify(result, null, 2);
}

export const taskStartImplementingDirectTool: Tool = {
  name: 'task:start-implementing-direct',
  description: 'Skip planning and start implementation directly for a task',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: { type: 'string', description: 'Task ID' },
      branchName: { type: 'string', description: 'Git branch name for the worktree' },
      aiType: { type: 'string', description: 'AI agent type (e.g., claude-code, gemini, cursor)' },
      useRemoteBranch: {
        type: 'boolean',
        description: 'Check out an existing remote branch (default: false)',
      },
    },
    required: ['taskId', 'branchName', 'aiType'],
  },
};

export async function executeTaskStartImplementingDirect(
  input: Record<string, unknown>
): Promise<string> {
  const result = await client.startImplementingDirect(input.taskId as string, {
    branchName: input.branchName as string,
    aiType: input.aiType as string,
    useRemoteBranch: input.useRemoteBranch as boolean | undefined,
  });
  return JSON.stringify(result, null, 2);
}
