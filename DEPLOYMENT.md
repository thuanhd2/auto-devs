# Deployment Guide - Auto-Devs with MCP Server

## Overview

This guide covers deploying Auto-Devs API and MCP Server using PM2.

- **auto-devs-api** - Go backend server (port 8098)
- **auto-devs-mcp** - Node.js MCP server (stdio transport)

## Local Development

### Build

```bash
./build-package.sh
```

This will:
1. Build frontend (Vite)
2. Build Go server & worker
3. Build MCP server (npm build)

### Run with PM2

```bash
# Install dependencies (if not done)
npm install --prefix mcp-server

# Start both services
pm2 start ecosystem.config.js

# View logs
pm2 logs

# View status
pm2 status
```

## PM2 Commands

```bash
# Start all apps
pm2 start ecosystem.config.js

# Start specific app
pm2 start ecosystem.config.js --only auto-devs-api
pm2 start ecosystem.config.js --only auto-devs-mcp

# Stop
pm2 stop auto-devs-api
pm2 stop auto-devs-mcp

# Restart
pm2 restart auto-devs-api
pm2 restart auto-devs-mcp

# Reload (graceful restart)
pm2 reload ecosystem.config.js

# Remove
pm2 delete auto-devs-api

# View logs
pm2 logs auto-devs-api
pm2 logs auto-devs-mcp
pm2 logs

# Monitor
pm2 monit
```

## Environment Configuration

### MCP Server (.env)

Create `mcp-server/.env`:

```bash
# Auto-Devs API Configuration
AUTO_DEVS_API_URL=http://localhost:8098
AUTO_DEVS_API_KEY=

# MCP Server Configuration
MCP_DEBUG=false
ENABLE_CACHING=true
```

### API Server

Configure via `cmd/npx/.env` (see `.env.example`)

## Production Deployment

For production, update `ecosystem.config.js`:

```javascript
// Change these values
deploy: {
  production: {
    user: 'your-user',
    host: 'your-server.com',
    repo: 'your-repo-url',
    path: '/path/to/auto-devs',
  }
}
```

Then deploy:

```bash
pm2 deploy ecosystem.config.js production setup
pm2 deploy ecosystem.config.js production update
pm2 start ecosystem.config.js --env production
```

## Monitoring

### Health Checks

```bash
# API health
curl http://localhost:8098/swagger/index.html

# MCP server (should be listening on stdin)
ps aux | grep "node.*mcp"
```

### Logs

All logs stored in `logs/` directory:
- `api-error.log` - API errors
- `mcp-error.log` - MCP server errors
- `api-combined.log` - Combined API logs
- `mcp-combined.log` - Combined MCP logs

### Memory Management

- API: 1GB max
- MCP: 512MB max

Adjust in `ecosystem.config.js` if needed.

## Troubleshooting

### MCP Server not connecting to API

```bash
# Check MCP logs
pm2 logs auto-devs-mcp

# Verify API is running
pm2 status auto-devs-api

# Test API connectivity
curl http://localhost:8098/api/v1/projects
```

### PM2 not starting services

```bash
# Check PM2 status
pm2 status

# Check for errors
pm2 logs

# Delete and restart
pm2 delete all
pm2 start ecosystem.config.js
```

### Port conflicts

If port 8098 is in use:

```bash
# Find what's using the port
lsof -i :8098

# Kill the process
kill -9 <PID>

# Or change port in ecosystem.config.js
```

## Development vs Production

**Development:**
```bash
# Use make run for flexibility
make run  # In one terminal
cd mcp-server && npm run dev  # In another terminal
```

**Production:**
```bash
# Use PM2 for reliability
./build-package.sh
pm2 start ecosystem.config.js --env production
```

## Scaling

For multiple API instances, update `ecosystem.config.js`:

```javascript
{
  name: 'auto-devs-api',
  instances: 4,  // Changed from 1
  exec_mode: 'cluster',  // Added
  // ... rest of config
}
```

Note: MCP server should remain single instance (stdio transport).
