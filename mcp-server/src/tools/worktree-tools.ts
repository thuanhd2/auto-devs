import { Tool } from '@modelcontextprotocol/sdk/types.js';
import AutoDevsClient from '../client/autodevs-client.js';

const client = new AutoDevsClient();

export const worktreeGetStatusTool: Tool = {
  name: 'worktree:get-status',
  description: 'Get worktree status for a project',
  inputSchema: {
    type: 'object',
    properties: {
      projectId: {
        type: 'string',
        description: 'Project ID',
      },
    },
    required: ['projectId'],
  },
};

export async function executeWorktreeGetStatus(input: Record<string, unknown>): Promise<string> {
  const projectId = input.projectId as string;

  const result = await client.getWorktreeStatus(projectId);
  return JSON.stringify(result, null, 2);
}
