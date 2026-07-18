#!/usr/bin/env node

import { spawn } from 'child_process';

// Start MCP server
const server = spawn('node', ['dist/index.js'], {
  stdio: ['pipe', 'pipe', 'pipe'],
});

let serverOutput = '';
let serverErr = '';

server.stderr.on('data', (data) => {
  serverErr += data.toString();
  console.log('[SERVER]', data.toString().trim());
});

server.stdout.on('data', (data) => {
  serverOutput += data.toString();
  console.log('[SERVER STDOUT]', data.toString().trim());
});

// Wait for server to start
setTimeout(() => {
  console.log('\n=== Testing MCP Server ===\n');

  // Test 1: List tools
  console.log('→ Requesting tool list...');
  server.stdin.write(
    JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/list',
      params: {},
    }) + '\n'
  );

  // Test 2: Call project:list tool
  setTimeout(() => {
    console.log('\n→ Calling project:list tool...');
    server.stdin.write(
      JSON.stringify({
        jsonrpc: '2.0',
        id: 2,
        method: 'tools/call',
        params: {
          name: 'project:list',
          arguments: { page: 1, pageSize: 5 },
        },
      }) + '\n'
    );
  }, 1000);

  // Exit after tests
  setTimeout(() => {
    console.log('\n=== Test Complete ===\n');
    server.kill();
    process.exit(0);
  }, 3000);
}, 1000);

server.on('error', (err) => {
  console.error('Server error:', err);
  process.exit(1);
});

server.on('close', (code) => {
  console.log(`\nServer exited with code ${code}`);
});
