import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import AutoDevsClient from '../client/autodevs-client.js';
import { validateInput } from '../utils/validate-input.js';

const client = new AutoDevsClient();

const executionListSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  page: z.coerce.number().int().positive().optional().default(1),
  pageSize: z.coerce.number().int().positive().max(100).optional().default(10),
  status: z.string().optional(),
});

const executionIdSchema = z.object({
  id: z.string().min(1, 'Execution ID is required'),
});

const executionCreateSchema = z.object({
  taskId: z.string().min(1, 'Task ID is required'),
  scheduled: z.boolean().optional(),
  scheduledAt: z.string().optional(),
});

export const executionListTool: Tool = {
  name: 'execution:list',
  description: 'List executions for a task with optional filters',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: {
        type: 'string',
        description: 'Task ID',
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
        description: 'Filter by status (running, completed, failed)',
      },
    },
    required: ['taskId'],
  },
};

export async function executeExecutionList(input: Record<string, unknown>): Promise<string> {
  const { taskId, page, pageSize, status } = validateInput(executionListSchema, input);

  const result = await client.listExecutions(taskId, { page, pageSize, status });
  return JSON.stringify(result, null, 2);
}

export const executionGetTool: Tool = {
  name: 'execution:get',
  description: 'Get execution details with logs and output',
  inputSchema: {
    type: 'object',
    properties: {
      id: {
        type: 'string',
        description: 'Execution ID',
      },
    },
    required: ['id'],
  },
};

export async function executeExecutionGet(input: Record<string, unknown>): Promise<string> {
  const { id } = validateInput(executionIdSchema, input);
  const result = await client.getExecution(id);
  return JSON.stringify(result, null, 2);
}

export const executionCreateTool: Tool = {
  name: 'execution:create',
  description: 'Create and trigger a new execution for a task',
  inputSchema: {
    type: 'object',
    properties: {
      taskId: {
        type: 'string',
        description: 'Task ID',
      },
      scheduled: {
        type: 'boolean',
        description: 'Schedule execution for later (default: immediate)',
      },
      scheduledAt: {
        type: 'string',
        description: 'ISO 8601 datetime for scheduled execution',
      },
    },
    required: ['taskId'],
  },
};

export async function executeExecutionCreate(input: Record<string, unknown>): Promise<string> {
  const parsed = validateInput(executionCreateSchema, input);
  const data: Record<string, unknown> = {};

  if (parsed.scheduled) {
    data.scheduled = true;
    if (parsed.scheduledAt) {
      data.scheduledAt = parsed.scheduledAt;
    }
  }

  const result = await client.createExecution(parsed.taskId, data);
  return JSON.stringify(result, null, 2);
}
