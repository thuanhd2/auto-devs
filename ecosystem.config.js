module.exports = {
  apps: [
    {
      name: 'auto-devs-api',
      script: 'cmd/npx/dist/server',
      cwd: './',
      env: {
        NODE_ENV: 'production',
        PORT: 8098,
      },
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '1G',
      merge_logs: true,
      error_file: 'logs/api-error.log',
      out_file: 'logs/api-out.log',
      log_file: 'logs/api-combined.log',
      time: true,
    },
    {
      name: 'auto-devs-mcp',
      script: 'mcp-server/dist/index.js',
      cwd: './',
      env: {
        NODE_ENV: 'production',
        AUTO_DEVS_API_URL: 'http://localhost:8098',
        MCP_DEBUG: 'false',
      },
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '512M',
      merge_logs: true,
      error_file: 'logs/mcp-error.log',
      out_file: 'logs/mcp-out.log',
      log_file: 'logs/mcp-combined.log',
      time: true,
    },
  ],

  deploy: {
    production: {
      user: 'node',
      host: 'your-server.com',
      ref: 'origin/main',
      repo: 'git@github.com:your-repo.git',
      path: '/var/www/auto-devs',
      'post-deploy': 'npm install && npm run build && pm2 reload ecosystem.config.js --env production',
    },
  },
};
