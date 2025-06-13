# AI Distiller MCP Integration Guide

## Overview

AI Distiller integrates with AI assistants through the Model Context Protocol (MCP), enabling Claude Desktop and other AI tools to directly analyze your codebase without manual copy-pasting.

## Architecture Decision

Based on extensive analysis with multiple AI perspectives, we've chosen a **phased monorepo approach**:

1. **Phase 1 (v0.3.0)**: MCP server integrated into main `aid` binary with `--mcp-server` flag
2. **Phase 2 (v1.0.0)**: Separate `aid-mcp` binary in `cmd/aid-mcp/` for advanced features

This approach ensures:
- Immediate availability for users
- Shared core logic without duplication
- Clean separation of concerns as features mature

## Installation

### Quick Start (Coming in v0.3.0)

```bash
# macOS/Linux
brew install janreges/tap/ai-distiller

# Windows
scoop install ai-distiller

# Or download pre-built binaries from GitHub releases
```

### Configure Claude Desktop

1. Open Claude Desktop → Settings → Developer → Edit Config
2. Add AI Distiller configuration:

```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "aid",
      "args": ["--mcp-server"],
      "env": {
        "AID_ROOT": "/path/to/your/project",
        "AID_CACHE_DIR": "~/.cache/aid"
      }
    }
  }
}
```

3. Restart Claude Desktop
4. Look for the 🔨 icon in the input field - AI Distiller tools are ready!

## Available MCP Tools

### 1. `distillFile`

Extracts structure from a single source file.

**Parameters:**
- `file_path` (string, required): Relative path from project root
- `strip_comments` (boolean): Remove comments and docstrings
- `strip_implementation` (boolean): Keep only signatures
- `strip_non_public` (boolean): Hide private/internal members
- `output_format` (string): "text", "json", or "md"

**Example:**
```
Claude: Let me analyze the user service structure...
[Calling distillFile with file_path="services/user_service.py", strip_implementation=true]
```

### 2. `listFiles`

Lists files in a directory with language statistics.

**Parameters:**
- `path` (string): Directory path relative to root
- `pattern` (string): File pattern (e.g., "*.py", "test_*.js")
- `recursive` (boolean): Include subdirectories

**Example:**
```
Claude: I'll check what test files you have...
[Calling listFiles with path="tests/", pattern="test_*.py"]
```

### 3. `getFileContent`

Reads raw file content (complements distillFile for full implementation).

**Parameters:**
- `file_path` (string, required): Relative path from project root
- `start_line` (number): Starting line number
- `end_line` (number): Ending line number

### 4. `search`

Searches codebase using ripgrep-style patterns.

**Parameters:**
- `query` (string, required): Search pattern
- `is_regex` (boolean): Enable regex mode
- `is_case_sensitive` (boolean): Case-sensitive search
- `path_includes` (array): Include only these paths
- `path_excludes` (array): Exclude these paths

**Example:**
```
Claude: Let me find all SQL queries...
[Calling search with query="SELECT.*FROM", is_regex=true]
```

## Security Features

### Path Traversal Protection
- All file paths are sanitized and restricted to `AID_ROOT`
- Symlinks are resolved and validated
- Attempts to access outside root return security errors

### Origin Validation
- MCP server validates request origins
- Only whitelisted origins (Claude Desktop) accepted
- CORS headers properly configured

### Resource Limits
- File size limit: 10MB (configurable)
- Search results capped at 200 matches
- Memory-efficient streaming for large operations

## Example Workflows

### 1. Understanding a New Codebase

> **You**: I just cloned this Python project. Can you give me an overview of its architecture?

> **Claude**: I'll analyze the project structure for you.

*Claude's MCP calls:*
1. `listFiles("/", recursive=true)` - Get project layout
2. `distillFile("main.py", strip_implementation=true)` - Entry point
3. `search("class.*Service", is_regex=true)` - Find service layer
4. `distillFile` on key service files

> **Claude**: Here's the architecture overview:
> - **Entry Point**: `main.py` initializes a Flask application
> - **Service Layer**: Found 3 services (UserService, AuthService, EmailService)
> - **Data Layer**: SQLAlchemy models in `models/` directory
> - **API Routes**: RESTful endpoints in `routes/` following a consistent pattern
> 
> The project follows a typical MVC pattern with clear separation of concerns...

### 2. Implementing a New Feature

> **You**: I need to add a password reset feature. Show me the current auth flow and where to integrate.

> **Claude**: Let me examine your authentication system.

*Claude analyzes existing patterns and provides implementation matching your codebase style*

### 3. Code Review Assistance

> **You**: Can you review the changes I made to the payment module for security issues?

> **Claude**: I'll analyze your payment module for security concerns.

*Claude uses search to find sensitive patterns, distillFile to understand flow*

## Performance Optimization

### Caching Strategy

AI Distiller MCP implements a two-tier cache:

1. **In-memory LRU cache**: Hot files cached for instant access
2. **Disk cache** (optional): Persistent cache for large codebases

```bash
# Enable disk cache
aid --mcp-server --cache disk --cache-size 500
```

### Concurrent Operations

- Parallel file processing with controlled concurrency
- Request deduplication for simultaneous identical requests
- Progress reporting for long operations via WebSocket (future)

## Troubleshooting

### MCP Server Won't Start

1. Check if port is already in use:
   ```bash
   aid mcp status
   ```

2. View logs:
   - macOS: `~/Library/Logs/Claude/mcp-ai-distiller.log`
   - Windows: `%APPDATA%\Claude\logs\mcp-ai-distiller.log`

### "No tools available" in Claude

1. Verify configuration in Claude Desktop
2. Restart Claude Desktop completely
3. Run `aid --mcp-server --test` to verify server starts

### Performance Issues

1. Enable caching: `--cache disk`
2. Exclude large directories: Set `AID_EXCLUDE="node_modules,vendor"`
3. Increase memory limit: `--max-memory 2048`

## Future Enhancements

### Phase 1 (v0.3.0) - Current
- ✅ Basic MCP server in main binary
- ✅ Core tools: distillFile, listFiles, search, getFileContent
- ✅ Memory caching
- ✅ Security hardening

### Phase 2 (v0.4.0)
- [ ] Separate `aid-mcp` binary
- [ ] WebSocket support for progress
- [ ] Disk caching
- [ ] Project metadata endpoint

### Phase 3 (v1.0.0)
- [ ] Semantic code search
- [ ] Multi-project support
- [ ] Integration with other MCP tools
- [ ] Custom tool extensions

## Contributing

Help improve MCP integration:

1. Test with your Claude Desktop setup
2. Report issues with detailed logs
3. Suggest new tools based on your workflow
4. Contribute to documentation

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [MCP Quickstart Guide](https://modelcontextprotocol.io/quickstart)
- [Claude Desktop MCP Docs](https://docs.anthropic.com/en/docs/build-with-claude/mcp)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go MCP implementation