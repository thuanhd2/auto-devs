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

MCP server is **spawned by agents** via stdio - agents start it automatically, no separate deployment.

### Hermes/Serena Agents

Configure in your agent config file:

```yaml
mcp_servers:
  - name: auto-devs
    command: node
    args:
      - /path/to/auto-devs/mcp-server/dist/index.js
    env:
      AUTO_DEVS_API_URL: "http://localhost:8098"
      MCP_DEBUG: "false"
```

Agent will:
1. Spawn MCP server process
2. Load all 11 tools
3. Use tools in agent logic
4. Stop MCP when agent stops

### Claude Code

Configure in Claude Code settings:

```json
{
  "mcpServers": {
    "auto-devs": {
      "command": "node",
      "args": ["/path/to/auto-devs/mcp-server/dist/index.js"]
    }
  }
}
```

Claude Code will load all tools and use them in conversations.

### Testing Manually

For debugging and development:

```bash
# Build MCP server
npm run build

# Start MCP directly
npm start

# In another terminal, send requests via stdin:
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list"
}' | node dist/index.js

# Should output JSON with list of 11 tools
```

### Docker

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY mcp-server .
RUN npm install && npm run build
ENV AUTO_DEVS_API_URL=http://auto-devs-api:8098
CMD ["node", "dist/index.js"]
```

Note: Docker image can be used by agents that support mounting custom MCP servers.

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
