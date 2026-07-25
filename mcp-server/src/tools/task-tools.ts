import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import AutoDevsClient from '../client/autodevs-client.js';
import { validateInput } from '../utils/validate-input.js';

const client = new AutoDevsClient();

const taskListSchema = z.object({
  projectId: z.string().min(1, 'Project ID is required'),
  page: z.coerce.number().int().positive().optional().default(1),
  pageSize: z.coerce.number().int().positive().max(100).optional().default(10),
  status: z.string().optional(),
  priority: z.string().optional(),
});

const taskCreateSchema = z.object({
  projectId: z.string().min(1, 'Project ID is required'),
  title: z.string().min(1, 'Title is required'),
  description: z.string().optional(),
  priority: z.string().optional(),
  kanban_task_id: z.string().optional(),
});

const taskUpdateStatusSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  status: z.string().min(1, 'Status is required'),
});

const taskIdSchema = z.object({
  id: z.string().min(1, 'Task ID is required'),
});

const taskStartPlanningSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  branchName: z.string().min(1, 'Branch name is required'),
  aiType: z.string().min(1, 'AI type is required'),
  useRemoteBranch: z.boolean().optional(),
  autoImplement: z.boolean().optional(),
});

const taskApprovePlanSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  aiType: z.string().min(1, 'AI type is required'),
});

const taskStartImplementingDirectSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  branchName: z.string().min(1, 'Branch name is required'),
  aiType: z.string().min(1, 'AI type is required'),
  useRemoteBranch: z.boolean().optional(),
});

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
  const { projectId, page, pageSize, status, priority } = validateInput(taskListSchema, input);

  const result = await client.listTasks(projectId, { page, pageSize, status, priority });
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
  const parsed = validateInput(taskCreateSchema, input);
  const { projectId, ...taskFields } = parsed;

  const taskData: Record<string, unknown> = {
    title: taskFields.title,
  };
  if (taskFields.description !== undefined) {
    taskData.description = taskFields.description;
  }
  if (taskFields.priority !== undefined) {
    taskData.priority = taskFields.priority;
  }
  if (taskFields.kanban_task_id !== undefined) {
    taskData.kanban_task_id = taskFields.kanban_task_id;
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
  const { taskId, status } = validateInput(taskUpdateStatusSchema, input);

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
  const { id } = validateInput(taskIdSchema, input);
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
  const { id } = validateInput(taskIdSchema, input);
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
  const parsed = validateInput(taskStartPlanningSchema, input);
  const result = await client.startPlanning(parsed.taskId, {
    branchName: parsed.branchName,
    aiType: parsed.aiType,
    useRemoteBranch: parsed.useRemoteBranch,
    autoImplement: parsed.autoImplement,
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
  const parsed = validateInput(taskApprovePlanSchema, input);
  const result = await client.approvePlan(parsed.taskId, {
    aiType: parsed.aiType,
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
  const parsed = validateInput(taskStartImplementingDirectSchema, input);
  const result = await client.startImplementingDirect(parsed.taskId, {
    branchName: parsed.branchName,
    aiType: parsed.aiType,
    useRemoteBranch: parsed.useRemoteBranch,
  });
  return JSON.stringify(result, null, 2);
}
