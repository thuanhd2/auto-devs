# MCP Server - Agent Integration Guide

## How Agents Use MCP Server

MCP server runs as a **subprocess spawned by agents** (Hermes, etc.) via stdio - no separate deployment needed.

## Agent Configuration

### Hermes/Serena Agent

In your `hermes-config.yaml`:

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

When agent runs:
1. Agent spawns: `node mcp-server/dist/index.js`
2. Agent communicates via stdin/stdout (pipes)
3. MCP loads all 11 tools
4. Agent can use tools to interact with Auto-Devs
5. When agent stops → MCP stops automatically

### Claude Code Integration

In Claude Code settings:

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

## Building MCP Server

MCP server is built as part of main build:

```bash
# From project root
./build-package.sh
```

This builds MCP to `mcp-server/dist/index.js`

Verify:
```bash
ls -la mcp-server/dist/index.js
```

## Development

```bash
# Terminal 1: Backend API
make run

# Terminal 2: MCP Server (with hot reload)
cd mcp-server
npm run dev
```

## Configuration

Environment variables (set in agent config):

- `AUTO_DEVS_API_URL` - Backend API URL (default: http://localhost:8098)
- `MCP_DEBUG` - Enable debug logging (default: false)
- `ENABLE_CACHING` - Enable response caching (default: true)

## Testing MCP Manually

```bash
# Start backend API first
make run

# In another terminal, test MCP directly
cd mcp-server
npm run build
npm start

# In third terminal, send MCP request (stdin)
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list",
  "params": {}
}' | node dist/index.js
```

## Troubleshooting

### Agent can't find MCP

Check:
1. Path to MCP server is correct
2. `MCP_DEBUG=true` to see errors
3. MCP has been built: `ls mcp-server/dist/index.js`

### MCP can't connect to API

```bash
# Check API is running
curl http://localhost:8098/api/v1/projects

# Check MCP can reach API (with MCP_DEBUG=true)
MCP_DEBUG=true node mcp-server/dist/index.js
# Should show: [MCP] API URL: http://localhost:8098
```

### Tools not loading

```bash
# Verify dist folder
ls mcp-server/dist/tools/

# Rebuild
cd mcp-server
npm run build
```

## See Also

- [Main Deployment Guide](../DEPLOYMENT.md)
- [MCP Server README](README.md)
