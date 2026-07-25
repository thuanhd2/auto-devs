import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { z } from 'zod';
import AutoDevsClient from '../client/autodevs-client.js';
import { validateInput } from '../utils/validate-input.js';

const client = new AutoDevsClient();

const worktreeGetStatusSchema = z.object({
  projectId: z.string().min(1, 'Project ID is required'),
});

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
  const { projectId } = validateInput(worktreeGetStatusSchema, input);

  const result = await client.getWorktreeStatus(projectId);
  return JSON.stringify(result, null, 2);
}
