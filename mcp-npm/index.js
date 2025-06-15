#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

// Get the path to the aid-mcp binary
const binaryPath = path.join(__dirname, 'bin', 'aid-mcp');

// Forward all arguments to the binary
const args = process.argv.slice(2);

// Spawn the binary
const child = spawn(binaryPath, args, {
  stdio: 'inherit',
  env: {
    ...process.env,
    // Set default environment variables if not already set
    AID_ROOT: process.env.AID_ROOT || process.cwd(),
    AID_CACHE_DIR: process.env.AID_CACHE_DIR || path.join(require('os').homedir(), '.cache', 'aid')
  }
});

// Handle exit
child.on('exit', (code) => {
  process.exit(code);
});

// Handle errors
child.on('error', (err) => {
  console.error('Failed to start AI Distiller MCP server:', err);
  process.exit(1);
});