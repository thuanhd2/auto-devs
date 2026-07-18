# Deployment Guide - Auto-Devs API

## Overview

Auto-Devs consists of:
- **Backend API** - Go server (port 8098)
- **Frontend** - React/TypeScript (Vite)
- **MCP Server** - Node.js (spawned by agents via stdio)

## Local Development

### Build

```bash
./build-package.sh
```

This will:
1. Build frontend (Vite)
2. Build Go server & worker
3. Build MCP server (npm build)

### Run

```bash
# Terminal 1: Start backend API
make run

# Terminal 2: Start frontend dev server (optional)
cd frontend && npm run dev

# Terminal 3: Run Hermes agent
# Agent will spawn MCP server automatically
```

## MCP Server

**Important:** MCP server is spawned by agents (Hermes, etc.) via stdio - no separate deployment needed.

```yaml
# In agent config (hermes-config.yaml)
mcp_servers:
  - name: auto-devs
    command: node
    args: [/path/to/mcp-server/dist/index.js]
    env:
      AUTO_DEVS_API_URL: "http://localhost:8098"
      MCP_DEBUG: "false"
```

When agent runs, it will:
1. Spawn `node mcp-server/dist/index.js` as subprocess
2. Communicate via stdin/stdout (pipes)
3. Automatically stop when agent stops

**Note:** MCP server does NOT run as HTTP service, does NOT occupy a port.

See [mcp-server/README.md](./mcp-server/README.md) for MCP details.

## Environment Configuration

### Backend API

Configure via `cmd/npx/.env`:
- Database connection
- Server port (default: 8098)
- API keys

See `.env.example` for template.

### Frontend

Configure via environment variables or `frontend/.env`:
- API base URL
- WebSocket URL

## Production Deployment

### Prerequisites

```bash
# Build everything
./build-package.sh

# Verify MCP server built
ls -la mcp-server/dist/index.js
```

### API Server

```bash
# Run backend API (option 1: direct)
./cmd/npx/dist/server

# Run backend API (option 2: PM2)
pm2 start ./cmd/npx/dist/server --name auto-devs-api
pm2 save
```

### Frontend

Option 1: Serve static files (dist/) from API server
Option 2: Deploy to CDN/static host separately

### MCP Server

No separate deployment needed - agents spawn it automatically.

## Monitoring

### Health Checks

```bash
# API health
curl http://localhost:8098/swagger/index.html

# Check API is responding
curl http://localhost:8098/api/v1/projects
```

### Logs

Backend logs:
```bash
# Direct run
# Logs printed to stdout/stderr

# PM2 run
pm2 logs auto-devs-api
```

## Troubleshooting

### API not starting

```bash
# Check if port 8098 is in use
lsof -i :8098

# Kill conflicting process
kill -9 <PID>

# Try again
make run
```

### MCP server not building

```bash
# Rebuild MCP server
cd mcp-server
npm install
npm run build
cd ..

# Verify dist folder exists
ls mcp-server/dist/index.js
```

### Agent can't find MCP server

Check in agent config:
1. Path to MCP server is correct
2. `AUTO_DEVS_API_URL` points to running API
3. MCP has been built (`npm run build`)

```bash
# Test MCP manually
node mcp-server/dist/index.js
# (should start without errors, waiting for input)
```

## Performance

### Database

Use connection pooling via GORM configuration in backend.

### API Server

Default runs on single process. For higher load:
```bash
# Run multiple instances behind load balancer
pm2 start ./cmd/npx/dist/server --name auto-devs-api -i 4
```

### Memory Usage

- API Server: ~200-500MB
- MCP Server: ~100-200MB (per agent instance)
- Frontend: Static files only

## Scaling Considerations

- **API Layer**: Use load balancer (nginx, etc.) for multiple instances
- **Database**: Ensure adequate connection pool
- **MCP Server**: No special scaling needed (spawned per agent)
- **Frontend**: Deploy static files to CDN
