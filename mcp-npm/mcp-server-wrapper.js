#!/usr/bin/env node

/**
 * Wrapper script for AI Distiller MCP server
 * Ensures the SDK-based server is available and starts it
 */

const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

// Path to the SDK server
const sdkServerPath = path.join(__dirname, 'dist', 'mcp-server-sdk.js');

// Check if TypeScript build exists
if (!fs.existsSync(sdkServerPath)) {
  console.error('Error: AI Distiller MCP server build not found at:', sdkServerPath);
  console.error('');
  console.error('This is likely a packaging issue. Please report it at:');
  console.error('https://github.com/janreges/ai-distiller/issues');
  process.exit(1);
}

// Start the SDK server
const child = spawn(process.execPath, [sdkServerPath], {
  stdio: 'inherit',
  env: process.env
});

child.on('exit', (code) => {
  process.exit(code || 0);
});

child.on('error', (err) => {
  console.error('Failed to start AI Distiller MCP server:', err);
  process.exit(1);
});