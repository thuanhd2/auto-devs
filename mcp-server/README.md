# Auto-Devs MCP Server

MCP (Model Context Protocol) server wrapper for the Auto-Devs API. Enables AI models to interact with Auto-Devs through standardized MCP tools.

## Quick Start

### Setup

```bash
npm install
cp .env.example .env
# Edit .env with your Auto-Devs API URL
```

### Development

```bash
npm run dev
```

### Build & Run

```bash
npm run build
npm start
```

## MVP Tools (Phase 1)

1. **project:list** - List all projects with pagination
2. **project:get** - Get project details by ID
3. **task:list** - List tasks by project with filters
4. **task:create** - Create new task
5. **task:update-status** - Update task status
6. **execution:list** - List executions with filtering
7. **worktree:get-status** - Get worktree status

## Configuration

Set environment variables in `.env`:
- `AUTO_DEVS_API_URL` - Auto-Devs API base URL (default: http://localhost:8098)
- `AUTO_DEVS_API_KEY` - Optional API key for authentication
- `MCP_DEBUG` - Enable debug logging (default: false)
- `ENABLE_CACHING` - Enable response caching (default: true)

## Architecture

```
src/
├── client/       - AutoDevsClient for API wrapping
├── tools/        - MCP tool definitions and handlers
├── resources/    - MCP resource definitions
├── types/        - TypeScript interfaces
└── config.ts     - Configuration
```

## Next Steps (Phase 2+)

- Add MCP resources for better context
- Implement caching layer
- Add more tools (task updates, logs, etc.)
- WebSocket support for real-time updates
