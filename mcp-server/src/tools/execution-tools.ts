import { Tool } from '@modelcontextprotocol/sdk/types.js';
import AutoDevsClient from '../client/autodevs-client.js';

const client = new AutoDevsClient();

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
  const taskId = input.taskId as string;
  const page = (input.page as number) || 1;
  const pageSize = (input.pageSize as number) || 10;
  const status = input.status as string | undefined;

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
  const id = input.id as string;
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
  const taskId = input.taskId as string;
  const data: any = {};
  if (input.scheduled) {
    data.scheduled = true;
    if (input.scheduledAt) {
      data.scheduledAt = input.scheduledAt;
    }
  }

  const result = await client.createExecution(taskId, data);
  return JSON.stringify(result, null, 2);
}
