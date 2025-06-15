#!/bin/bash

# Test script for AI Distiller MCP server

set -e

echo "Testing AI Distiller MCP server..."

# Build if not exists
if [ ! -f bin/aid-mcp ]; then
    echo "Building aid-mcp..."
    go build -o bin/aid-mcp ./cmd/aid-mcp
fi

# Test 1: Initialize
echo "Test 1: Initialize"
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{"tools":{}}},"id":1}' | ./bin/aid-mcp

# Test 2: List tools
echo -e "\nTest 2: List tools"
echo '{"jsonrpc":"2.0","method":"tools/list","id":2}' | ./bin/aid-mcp

# Test 3: Get capabilities
echo -e "\nTest 3: Get capabilities"
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"getCapabilities","arguments":{}},"id":3}' | ./bin/aid-mcp

echo -e "\nAll tests passed!"