#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
  TextContent,
} from '@modelcontextprotocol/sdk/types.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { TOOLS, getToolByName } from './tools/index.js';
import { config } from './config.js';
import { AppError } from './errors/app-error.js';
import { logger } from './utils/logger.js';

const server = new Server(
  {
    name: 'auto-devs-mcp-server',
    version: '0.1.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Handle list tools request
server.setRequestHandler(ListToolsRequestSchema, () => {
  const tools = TOOLS.map((t) => t.tool);
  if (config.debug) {
    console.error(`[MCP] Listed ${tools.length} tools`);
  }
  return { tools };
});

// Handle call tool request
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const toolName = request.params.name;
  const toolInput = request.params.arguments || {};

  if (config.debug) {
    console.error(`[MCP] Calling tool: ${toolName}`);
    console.error(`[MCP] Input:`, toolInput);
  }

  const handler = getToolByName(toolName);
  if (!handler) {
    return {
      content: [
        {
          type: 'text' as const,
          text: `Error: Tool '${toolName}' not found`,
        },
      ],
      isError: true,
    };
  }

  try {
    const result = await handler.execute(toolInput);
    logger.debug(`Tool result: ${result.substring(0, 200)}`);
    return {
      content: [
        {
          type: 'text' as const,
          text: result,
        },
      ],
    };
  } catch (error) {
    if (error instanceof AppError) {
      logger.error(`Tool error: ${error.code}`, error.toJSON());
      return {
        content: [
          {
            type: 'text' as const,
            text: `Error [${error.code}]: ${error.message}${
              error.suggestion ? `\nSuggestion: ${error.suggestion}` : ''
            }`,
          },
        ],
        isError: true,
      };
    }

    const errorMessage = error instanceof Error ? error.message : String(error);
    logger.error(`Unexpected error: ${errorMessage}`);
    return {
      content: [
        {
          type: 'text' as const,
          text: `Unexpected error: ${errorMessage}`,
        },
      ],
      isError: true,
    };
  }
});

// Start server
async function main() {
  console.error('[MCP] Starting Auto-Devs MCP Server');
  console.error(`[MCP] API URL: ${config.apiUrl}`);
  console.error(`[MCP] Debug mode: ${config.debug}`);

  const transport = new StdioServerTransport();
  await server.connect(transport);

  console.error('[MCP] Server connected and ready for requests');
}

main().catch((error) => {
  console.error('[MCP] Fatal error:', error);
  process.exit(1);
});
