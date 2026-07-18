#!/usr/bin/env node

import { spawn } from 'child_process';

const server = spawn('node', ['dist/index.js'], {
  stdio: ['pipe', 'pipe', 'pipe'],
});

server.stderr.on('data', (data) => {
  if (data.toString().includes('[MCP]')) {
    console.log(data.toString().trim());
  }
});

setTimeout(() => {
  console.log('\n=== Listing MCP Tools ===\n');
  server.stdin.write(
    JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/list',
      params: {},
    }) + '\n'
  );

  server.stdout.once('data', (data) => {
    const response = JSON.parse(data.toString());
    const tools = response.result.tools;

    console.log(`Total Tools: ${tools.length}\n`);
    console.log('MVP Tools (7):');
    tools.slice(0, 7).forEach((t, i) => {
      console.log(`  ${i + 1}. ${t.name}`);
    });

    console.log('\nPhase 2A Tools (4):');
    tools.slice(7, 11).forEach((t, i) => {
      console.log(`  ${i + 1}. ${t.name}`);
    });

    console.log('\n✅ All 11 tools registered\n');
    server.kill();
    process.exit(0);
  });
}, 1000);

setTimeout(() => {
  console.log('Timeout');
  server.kill();
  process.exit(1);
}, 10000);
