#!/bin/bash

# Test script for local development

set -e

echo "Testing AI Distiller MCP server locally..."

# Build the binary
echo "Building aid-mcp..."
go build -o bin/aid-mcp ./cmd/aid-mcp

# Make sure aid binary exists
if ! command -v aid &> /dev/null; then
    echo "Error: 'aid' binary not found in PATH"
    echo "Please build and install AI Distiller first"
    exit 1
fi

# Test with sample commands
echo "Testing getCapabilities..."
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"getCapabilities","arguments":{}},"id":1}' | ./bin/aid-mcp

echo ""
echo "MCP server is ready for testing!"
echo "You can now run: ./bin/aid-mcp"