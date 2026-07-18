# Auto-Devs MCP Server

MCP (Model Context Protocol) server for the Auto-Devs API. Enables AI models and agents to interact with Auto-Devs through standardized tools and resources.

## Features

✅ **11 Production Tools**
- Project management (list, get)
- Task operations (list, create, get, update-status, delete)
- Execution control (list, get, create)
- Worktree management

✅ **Robust Error Handling**
- Structured error codes with meaningful messages
- Automatic retry logic (3x with exponential backoff)
- Field-level validation errors
- Suggestions for resolution

✅ **Performance**
- Request logging and debugging
- Support for response caching
- Efficient API wrapper with connection pooling

## Quick Start

### Setup

```bash
npm install
cp .env.example .env
# Edit .env with your Auto-Devs API URL (default: http://localhost:8098)
```

### Development

```bash
npm run dev
# Runs with hot reload
```

### Build & Run

```bash
npm run build
npm start
```

### With PM2 (Production)

From project root:
```bash
./build-package.sh
pm2 start ecosystem.config.js
pm2 logs auto-devs-mcp
```

## Tools Available

### Phase 1: MVP Tools (7)
1. **project:list** - List all projects with pagination
2. **project:get** - Get project details by ID
3. **task:list** - List tasks by project with filters
4. **task:create** - Create new task
5. **task:update-status** - Update task status
6. **execution:list** - List executions with filtering
7. **worktree:get-status** - Get worktree status

### Phase 2A: Enhanced Tools (4)
8. **task:get** - Get full task details with history
9. **task:delete** - Delete task
10. **execution:get** - Get execution with logs and output
11. **execution:create** - Trigger new execution

All tools include:
- JSON schema validation
- Structured error responses
- Automatic retry on transient failures

## Configuration

Set environment variables in `.env`:

```bash
# Auto-Devs API
AUTO_DEVS_API_URL=http://localhost:8098
AUTO_DEVS_API_KEY=                    # Optional

# MCP Server
MCP_DEBUG=false                       # Enable debug logging
ENABLE_CACHING=true                   # Enable response caching
```

## How to Use MCP Server

### Claude Code Integration

**For Claude Code (AI Developer):**

1. Install MCP server locally or deploy it
2. Configure in Claude Code settings:
   ```json
   {
     "mcpServers": {
       "auto-devs": {
         "command": "node",
         "args": ["/path/to/mcp-server/dist/index.js"]
       }
     }
   }
   ```
3. Claude Code will automatically load all 11 tools
4. Use tools in your AI tasks:
   ```
   Use project:list to see all Auto-Devs projects
   Then use task:create to start a new task
   ```

### Hermes Agents Integration

**For Hermes/Serena Agent Framework:**

1. Setup MCP server in your agent configuration:
   ```yaml
   mcp_servers:
     - name: auto-devs
       command: node
       args:
         - /path/to/mcp-server/dist/index.js
       env:
         AUTO_DEVS_API_URL: "http://localhost:8098"
         MCP_DEBUG: "false"
   ```

2. Define agent tools from MCP:
   ```python
   from mcp import MCPClient
   
   client = MCPClient("auto-devs")
   tools = client.list_tools()  # Returns all 11 tools
   
   # Use in agent
   result = client.call_tool("project:list", {"page": 1, "pageSize": 5})
   ```

3. Example agent usage:
   ```python
   agent = HermesAgent(
       name="task-manager",
       tools=mcp_tools,
       system_prompt="You manage Auto-Devs tasks..."
   )
   
   # Agent will use MCP tools to complete tasks
   response = agent.run("Create a task in project X")
   ```

### Direct Command Line Usage

**For testing and debugging:**

```bash
# Start MCP server
npm start

# In another terminal, send MCP requests:
# (MCP uses JSON-RPC over stdin/stdout)

echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list",
  "params": {}
}' | node dist/index.js

echo '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "project:list",
    "arguments": {"page": 1, "pageSize": 5}
  }
}' | node dist/index.js
```

### Docker Deployment

**Run MCP server in Docker:**

```dockerfile
FROM node:18-alpine

WORKDIR /app
COPY mcp-server .

RUN npm install && npm run build

ENV AUTO_DEVS_API_URL=http://auto-devs-api:8098
ENV MCP_DEBUG=false

CMD ["node", "dist/index.js"]
```

```bash
docker run -e AUTO_DEVS_API_URL=http://host.docker.internal:8098 auto-devs-mcp
```

## Error Handling

When tools fail, responses include:
```json
{
  "code": "NOT_FOUND",
  "message": "Failed to get task xyz: not found",
  "details": {"httpStatus": 404, "endpoint": "/api/v1/tasks/xyz"},
  "suggestion": "Check task ID is correct and task exists"
}
```

Errors are automatically retried for transient failures:
- `RATE_LIMITED` (429)
- `SERVICE_UNAVAILABLE` (503)
- `TIMEOUT` (504)

## Architecture

```
src/
├── client/         - AutoDevsClient API wrapper with retry logic
├── tools/          - 11 MCP tool definitions and handlers
├── errors/         - Error codes and AppError class
├── utils/
│   ├── retry.ts    - Automatic retry with exponential backoff
│   └── logger.ts   - Structured logging
├── types/          - TypeScript interfaces
└── config.ts       - Configuration management
```

## Performance Monitoring

View MCP server metrics:

```bash
# Monitor logs
pm2 logs auto-devs-mcp

# View memory usage
pm2 monit

# Check response times
grep "Tool result" mcp-combined.log | head -20
```

## Roadmap

### Phase 2B (Planned)
- MCP Resources (projects://, tasks://{id}, etc.)
- Response caching layer (LRU with TTL)
- Performance optimization

### Phase 3 (Future)
- WebSocket support for real-time updates
- Streaming logs
- Resource subscriptions
- Metrics/observability

## Development

```bash
# Install dependencies
npm install

# Development with hot reload
npm run dev

# Build
npm run build

# Run tests (when available)
npm test

# Linting
npm run lint
npm run format
```

## Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for:
- PM2 ecosystem configuration
- Production deployment steps
- Health checks and monitoring
- Troubleshooting guide

## License

Same as Auto-Devs project
