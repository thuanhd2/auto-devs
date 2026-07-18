import { Tool } from '@modelcontextprotocol/sdk/types.js';
import AutoDevsClient from '../client/autodevs-client.js';
import { z } from 'zod';

const client = new AutoDevsClient();

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
  const page = (input.page as number) || 1;
  const pageSize = (input.pageSize as number) || 10;

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
  const id = input.id as string;
  const result = await client.getProject(id);
  return JSON.stringify(result, null, 2);
}
