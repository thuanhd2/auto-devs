import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import AutoDevsClient from '../client/autodevs-client.js';
import { validateInput } from '../utils/validate-input.js';

const client = new AutoDevsClient();

const projectListSchema = z.object({
  page: z.coerce.number().int().positive().optional().default(1),
  pageSize: z.coerce.number().int().positive().max(100).optional().default(10),
});

const projectGetSchema = z.object({
  id: z.string().min(1, 'Project ID is required'),
});

export const projectListTool: Tool = {
  name: 'project:list',
  description: 'List all projects with pagination',
  inputSchema: {
    type: 'object',
    properties: {
      page: {
        type: 'number',
        description: 'Page number (default: 1)',
      },
      pageSize: {
        type: 'number',
        description: 'Items per page (default: 10)',
      },
    },
  },
};

export async function executeProjectList(input: Record<string, unknown>): Promise<string> {
  const { page, pageSize } = validateInput(projectListSchema, input);

  const result = await client.listProjects(page, pageSize);
  return JSON.stringify(result, null, 2);
}

export const projectGetTool: Tool = {
  name: 'project:get',
  description: 'Get project details by ID',
  inputSchema: {
    type: 'object',
    properties: {
      id: {
        type: 'string',
        description: 'Project ID',
      },
    },
    required: ['id'],
  },
};

export async function executeProjectGet(input: Record<string, unknown>): Promise<string> {
  const { id } = validateInput(projectGetSchema, input);
  const result = await client.getProject(id);
  return JSON.stringify(result, null, 2);
}
