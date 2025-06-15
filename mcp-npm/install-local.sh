#!/bin/bash

# Install AI Distiller MCP server locally for Claude Code

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BIN_DIR="$SCRIPT_DIR/bin"

echo "Installing AI Distiller MCP server for Claude Code..."

# Check if binaries exist
if [ ! -f "$BIN_DIR/aid-mcp" ]; then
    echo "Error: aid-mcp binary not found. Please build it first."
    exit 1
fi

if [ ! -f "$BIN_DIR/aid" ]; then
    echo "Error: aid binary not found. Please build it first."
    exit 1
fi

# Create symlinks in ~/.local/bin (user's local bin directory)
USER_BIN="$HOME/.local/bin"
mkdir -p "$USER_BIN"

echo "Creating symlinks in $USER_BIN..."
ln -sf "$BIN_DIR/aid-mcp" "$USER_BIN/aid-mcp"
ln -sf "$BIN_DIR/aid" "$USER_BIN/aid"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$USER_BIN:"* ]]; then
    echo "Adding $USER_BIN to PATH..."
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    echo "Please run: source ~/.bashrc"
fi

echo ""
echo "Installation complete! Now you can add the MCP server to Claude Code:"
echo ""
echo "claude mcp add ai-distiller --scope user $USER_BIN/aid-mcp"
echo ""
echo "The server will analyze the current directory where you run Claude Code."
echo "You can also set AID_ROOT environment variable to specify a different root."