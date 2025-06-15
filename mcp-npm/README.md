# AI Distiller MCP Server

MCP (Model Context Protocol) server for [AI Distiller](https://github.com/janreges/ai-distiller) - a high-performance tool that extracts essential code structure from large codebases, making them digestible for LLMs.

## Quick Start with Claude Code

```bash
# One-line installation
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp

# Or install globally
npm install -g @janreges/ai-distiller-mcp
```

## Features

AI Distiller MCP exposes the following tools for AI agents:

- **distillFile** - Extract code structure from a single file
- **distillDirectory** - Extract structure from entire directories
- **listFiles** - List files with language statistics
- **getFileContent** - Read raw file content with line ranges
- **search** - Search codebase using literal or regex patterns
- **getCapabilities** - Get server information and available tools

## Configuration

### Environment Variables

- `AID_ROOT` - Root directory for analysis (defaults to current directory)
- `AID_CACHE_DIR` - Cache directory (defaults to ~/.cache/aid)

### Claude Desktop Configuration

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "npx",
      "args": ["-y", "@janreges/ai-distiller-mcp"],
      "env": {
        "AID_ROOT": "/path/to/your/project"
      }
    }
  }
}
```

## Tool Examples

### Extract Public API from a File

```javascript
// Claude will call:
distillFile({
  file_path: "src/main.py",
  include_private: false,
  include_implementation: false,
  output_format: "text"
})
```

### Analyze Entire Module Structure

```javascript
// Claude will call:
distillDirectory({
  directory_path: "src/auth",
  recursive: true,
  include_private: false,
  output_format: "json"
})
```

### Search for Patterns

```javascript
// Claude will call:
search({
  query: "TODO|FIXME",
  mode: "regex",
  max_results: 50
})
```

## Requirements

- Node.js >= 14.0.0
- AI Distiller binary (automatically downloaded or available in PATH)
- For search functionality: ripgrep (`rg`) installed

## Security

- Path traversal protection - all paths are sanitized
- File size limits - max 10MB per file
- Operation timeouts - configurable max timeout
- Resource limits - max files per operation

## Development

```bash
# Clone the repository
git clone https://github.com/janreges/ai-distiller.git
cd ai-distiller/mcp-npm

# Install dependencies
npm install

# Build the binary
npm run build

# Test locally
node index.js
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [AI Distiller on GitHub](https://github.com/janreges/ai-distiller)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [Claude Desktop](https://claude.ai)