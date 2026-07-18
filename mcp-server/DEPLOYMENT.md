# MCP Server Deployment

## Quick Start with PM2

### Setup

```bash
# Install PM2 globally (if not installed)
npm install -g pm2

# Create .env file
cp .env.example .env
```

### Start with Root Ecosystem Config

From project root:

```bash
# Build everything
./build-package.sh

# Start both API + MCP with pm2
pm2 start ecosystem.config.js

# Check status
pm2 status
```

Expected output:
```
┌────┬──────────────────┬──────────┬─────────┬──────────┐
│ id │ name             │ mode     │ status  │ restart  │
├────┼──────────────────┼──────────┼─────────┼──────────┤
│ 0  │ auto-devs-api    │ fork     │ online  │ 0        │
│ 1  │ auto-devs-mcp    │ fork     │ online  │ 0        │
└────┴──────────────────┴──────────┴─────────┴──────────┘
```

### Verify MCP Server

```bash
# Check if MCP is running
pm2 logs auto-devs-mcp

# Should see:
# [MCP] Starting Auto-Devs MCP Server
# [MCP] API URL: http://localhost:8098
# [MCP] Server connected and ready for requests
```

### Use MCP Server

MCP server listens on stdio. Connect using:

**Claude Claude Code:**
```bash
# In Claude Code, configure MCP server:
# claude_code settings → MCP servers
# Add: node /path/to/dist/index.js
```

**Or programmatically:**
```python
import subprocess

process = subprocess.Popen(
    ["node", "dist/index.js"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True,
)

# Send MCP requests via stdin
```

## Stop & Cleanup

```bash
# Stop MCP server
pm2 stop auto-devs-mcp

# Stop all
pm2 stop all

# Delete from PM2
pm2 delete auto-devs-mcp

# Clear all
pm2 delete all
```

## Logs

```bash
# View MCP logs
pm2 logs auto-devs-mcp

# Real-time monitoring
pm2 monit

# Save startup script
pm2 startup
pm2 save
```

## Development

For development, run separately:

```bash
# Terminal 1: Backend API
make run

# Terminal 2: MCP Server (with hot reload)
npm run dev
```

## Configuration

Environment variables in `.env`:

- `AUTO_DEVS_API_URL` - Backend API URL (default: http://localhost:8098)
- `MCP_DEBUG` - Enable debug logging (default: false)
- `ENABLE_CACHING` - Enable response caching (default: true)

## See Also

- [Root Deployment Guide](../DEPLOYMENT.md)
- [MCP Server README](README.md)
