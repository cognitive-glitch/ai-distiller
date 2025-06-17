# AI Distiller MCP Integration Guide

## Overview

AI Distiller integrates with AI assistants through the Model Context Protocol (MCP), enabling Claude Desktop and other AI tools to directly analyze your codebase without manual copy-pasting.

> **🚀 NEW: MCP Server Now Available!** Install with: `claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp`

## Architecture

The MCP server is implemented as a separate binary (`aid-mcp`) that wraps the core `aid` functionality:

- **Standalone MCP server**: Dedicated binary for MCP protocol handling
- **NPM distribution**: Easy installation via `npm install -g @janreges/ai-distiller-mcp`
- **Native binary wrapper**: Calls the main `aid` binary for actual processing
- **Stateless design**: Simple, robust architecture optimized for AI agent workflows

This approach provides:
- Clean separation between CLI and MCP server
- Easy distribution through NPM ecosystem
- Flexibility to evolve independently
- Optimal performance by reusing existing `aid` binary

## Installation

### Quick Start with Claude Code

```bash
# One-line installation for Claude Code users
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp

# Or manual installation
npm install -g @janreges/ai-distiller-mcp
```

### Alternative: Native Binary Installation

```bash
# macOS/Linux
brew install janreges/tap/ai-distiller

# Windows
scoop install ai-distiller

# Or download pre-built binaries from GitHub releases
```

### Configure Claude Desktop

For manual configuration or customization:

1. Open Claude Desktop → Settings → Developer → Edit Config
2. Add AI Distiller configuration:

```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "npx",
      "args": ["-y", "@janreges/ai-distiller-mcp"],
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

### 1. `distillDirectory` 🌟

Extracts structure from an entire directory or namespace - **the killer feature** for understanding code organization.

**Parameters:**
- `directory_path` (string, required): Directory path relative to root
- `recursive` (boolean): Include subdirectories (default: true)
- `include_private` (boolean): Include private/internal members (default: false)
- `include_implementation` (boolean): Include function/method bodies (default: false)
- `include_comments` (boolean): Include comments and docstrings (default: false)
- `include_imports` (boolean): Include import statements (default: true)
- `output_format` (string): "text", "json", or "md" (default: "text")
- `include_pattern` (string): Glob pattern for files to include (e.g., "*.py")
- `exclude_pattern` (string): Glob pattern for files to exclude (e.g., "test_*")

**Example:**
```
Claude: Let me analyze the entire authentication module structure...
[Calling distillDirectory with directory_path="src/auth/", include_implementation=false]
```

**Returns:** Consolidated view of all classes, interfaces, and functions in the directory, perfect for understanding module architecture.

### 2. `distillFile`

Extracts structure from a single source file.

**Parameters:**
- `file_path` (string, required): Relative path from project root
- `include_private` (boolean): Include private/internal members (default: false)
- `include_implementation` (boolean): Include function/method bodies (default: false)
- `include_comments` (boolean): Include comments and docstrings (default: false)
- `include_imports` (boolean): Include import statements (default: true)
- `output_format` (string): "text", "json", or "md" (default: "text")

**Example:**
```
Claude: Let me analyze the user service structure...
[Calling distillFile with file_path="services/user_service.py", include_implementation=false]
```

### 3. `listFiles`

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

### 4. `getFileContent`

Reads raw file content (complements distillFile for full implementation).

**Parameters:**
- `file_path` (string, required): Relative path from project root
- `start_line` (number): Starting line number (1-based)
- `end_line` (number): Ending line number (inclusive)

### 5. `search`

Searches codebase using ripgrep-style patterns.

**Parameters:**
- `query` (string, required): Search pattern
- `mode` (string): "literal" or "regex" (default: "literal")
- `case_sensitive` (boolean): Case-sensitive search (default: false)
- `path` (string): Limit search to this path
- `include_pattern` (string): Include only files matching this pattern
- `exclude_pattern` (string): Exclude files matching this pattern
- `max_results` (number): Maximum number of results (default: 100)

**Example:**
```
Claude: Let me find all SQL queries...
[Calling search with query="SELECT.*FROM", mode="regex"]
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
2. `distillDirectory("src/", include_implementation=false)` - Complete src overview
3. `distillFile("main.py", include_implementation=false)` - Entry point details
4. `distillDirectory("src/models/", include_implementation=false)` - Data layer structure

> **Claude**: Here's the architecture overview based on the complete namespace analysis:
> 
> **Service Layer** (from `src/services/`):
> - `UserService`: 8 public methods for user management
> - `AuthService`: JWT-based authentication with 5 endpoints
> - `EmailService`: Async email handling with template support
> 
> **Data Models** (from `src/models/`):
> - 12 SQLAlchemy models with relationships
> - Base classes: `TimestampedModel`, `SoftDeleteModel`
> - Key entities: `User`, `Organization`, `Project`, `Permission`
> 
> **API Structure** (from `src/routes/`):
> - RESTful endpoints grouped by resource
> - Consistent middleware chain: auth → validation → rate-limit
> - OpenAPI schema generation built-in
> 
> The project follows Domain-Driven Design with clear bounded contexts...

### 2. Implementing a New Feature

> **You**: I need to add a password reset feature. Show me the current auth flow and where to integrate.

> **Claude**: Let me examine your authentication system.

*Claude analyzes existing patterns and provides implementation matching your codebase style*

### 3. Code Review Assistance

> **You**: Can you review the changes I made to the payment module for security issues?

> **Claude**: I'll analyze your payment module for security concerns.

*Claude uses search to find sensitive patterns, distillFile or distillDirectory to understand flow*

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
- ✅ Core tools: distillFile, distillDirectory, listFiles, search, getFileContent
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