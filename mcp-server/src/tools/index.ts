import { Tool } from '@modelcontextprotocol/sdk/types.js';
import {
  projectListTool,
  executeProjectList,
  projectGetTool,
  executeProjectGet,
} from './project-tools.js';
import {
  taskListTool,
  executeTaskList,
  taskCreateTool,
  executeTaskCreate,
  taskUpdateStatusTool,
  executeTaskUpdateStatus,
  taskGetTool,
  executeTaskGet,
  taskDeleteTool,
  executeTaskDelete,
} from './task-tools.js';
import {
  executionListTool,
  executeExecutionList,
  executionGetTool,
  executeExecutionGet,
  executionCreateTool,
  executeExecutionCreate,
} from './execution-tools.js';
import {
  worktreeGetStatusTool,
  executeWorktreeGetStatus,
} from './worktree-tools.js';

export interface ToolHandler {
  tool: Tool;
  execute: (input: Record<string, unknown>) => Promise<string>;
}

export const TOOLS: ToolHandler[] = [
  {
    tool: projectListTool,
    execute: executeProjectList,
  },
  {
    tool: projectGetTool,
    execute: executeProjectGet,
  },
  {
    tool: taskListTool,
    execute: executeTaskList,
  },
  {
    tool: taskCreateTool,
    execute: executeTaskCreate,
  },
  {
    tool: taskUpdateStatusTool,
    execute: executeTaskUpdateStatus,
  },
  {
    tool: taskGetTool,
    execute: executeTaskGet,
  },
  {
    tool: taskDeleteTool,
    execute: executeTaskDelete,
  },
  {
    tool: executionListTool,
    execute: executeExecutionList,
  },
  {
    tool: executionGetTool,
    execute: executeExecutionGet,
  },
  {
    tool: executionCreateTool,
    execute: executeExecutionCreate,
  },
  {
    tool: worktreeGetStatusTool,
    execute: executeWorktreeGetStatus,
  },
];

export function getToolByName(name: string): ToolHandler | undefined {
  return TOOLS.find((t) => t.tool.name === name);
}

export function getAllTools(): Tool[] {
  return TOOLS.map((t) => t.tool);
}
