# Distributing AI Distiller via NPM

This guide explains how to distribute the `aid` binary through NPM using the [go-npm](https://github.com/sanathkr/go-npm) package, enabling trivial installation like `npm install -g @janreges/ai-distiller-mcp`.

## Overview

The go-npm package allows us to:
- Distribute cross-platform Go binaries via NPM
- Automatically download the correct binary for user's platform
- Add binary to PATH automatically
- Support simple `npx` execution

## Architecture

We'll create a separate NPM package (`@janreges/ai-distiller-mcp`) that:
1. Downloads the appropriate `aid` binary on installation
2. Exposes it via NPM's bin mechanism
3. Supports both global install and `npx` usage

## Step-by-Step Implementation

### 1. Create NPM Package Structure

Create a new directory `npm-package/` in the repository:

```bash
mkdir -p npm-package
cd npm-package
```

### 2. Initialize NPM Package

Create `npm-package/package.json`:

```json
{
  "name": "@janreges/ai-distiller-mcp",
  "version": "0.3.0",
  "description": "AI Distiller MCP server - Extract code structure for AI assistants",
  "keywords": ["mcp", "ai", "code-analysis", "ast", "distiller"],
  "homepage": "https://github.com/janreges/ai-distiller",
  "bugs": {
    "url": "https://github.com/janreges/ai-distiller/issues"
  },
  "repository": {
    "type": "git",
    "url": "https://github.com/janreges/ai-distiller.git"
  },
  "license": "MIT",
  "author": "Jan Reges",
  "main": "index.js",
  "bin": {
    "aid": "./bin/aid",
    "ai-distiller": "./bin/aid"
  },
  "scripts": {
    "postinstall": "go-npm install",
    "preuninstall": "go-npm uninstall"
  },
  "dependencies": {
    "go-npm": "^0.1.9"
  },
  "goBinary": {
    "name": "aid",
    "path": "./bin",
    "url": "https://github.com/janreges/ai-distiller/releases/download/v{{version}}/ai-distiller_{{version}}_{{platform}}_{{arch}}.tar.gz"
  },
  "engines": {
    "node": ">=14.0.0"
  }
}
```

### 3. Create Wrapper Script

Create `npm-package/index.js`:

```javascript
#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

// Handle MCP server mode for Claude
const args = process.argv.slice(2);
if (args.length === 0 || args[0] === '--mcp-server') {
  args.unshift('--mcp-server');
}

const binary = path.join(__dirname, 'bin', 'aid');
const child = spawn(binary, args, {
  stdio: 'inherit',
  env: process.env
});

child.on('exit', (code) => {
  process.exit(code);
});
```

### 4. Create README for NPM

Create `npm-package/README.md`:

```markdown
# AI Distiller MCP Server

AI Distiller extracts essential code structure from large codebases for AI assistants.

## Quick Start with Claude Desktop

```bash
# One-line installation
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp

# Or install globally
npm install -g @janreges/ai-distiller-mcp
```

## Manual Configuration

Add to Claude Desktop's config.json:

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

## Available MCP Tools

- `distillDirectory` - Extract structure from entire directories
- `distillFile` - Extract structure from a single file
- `listFiles` - List files with language stats
- `search` - Search codebase with regex

## Documentation

Full documentation: https://github.com/janreges/ai-distiller
```

### 5. Update Main Makefile

Add these targets to the main `Makefile`:

```makefile
# NPM package version should match main version
NPM_VERSION = $(VERSION)
NPM_DIR = npm-package

# Build binaries for all platforms (required for NPM release)
npm-build: cross-compile
	@echo "==> Preparing binaries for NPM distribution"
	@mkdir -p $(BUILD_DIR)/npm-release
	# Create tar.gz files in the format expected by go-npm
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary=$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then binary="$$binary.exe"; fi; \
		tar -czf $(BUILD_DIR)/npm-release/ai-distiller_$(NPM_VERSION)_$${os}_$${arch}.tar.gz \
			-C $(BUILD_DIR) $$binary; \
	done

# Update NPM package version
npm-version:
	@echo "==> Updating NPM package to version $(NPM_VERSION)"
	@cd $(NPM_DIR) && npm version $(NPM_VERSION) --no-git-tag-version

# Test NPM package locally
npm-test: npm-build
	@echo "==> Testing NPM package locally"
	@cd $(NPM_DIR) && npm install
	@cd $(NPM_DIR) && node index.js --version

# Publish to NPM
npm-publish: npm-build npm-version
	@echo "==> Publishing to NPM"
	@cd $(NPM_DIR) && npm publish --access public

# Create GitHub release with binaries
github-release: npm-build
	@echo "==> Creating GitHub release"
	@gh release create v$(VERSION) \
		--title "AI Distiller v$(VERSION)" \
		--notes-file CHANGELOG.md \
		$(BUILD_DIR)/npm-release/*.tar.gz

# Full release process
release: test lint npm-build github-release npm-publish
	@echo "==> Release v$(VERSION) completed!"
	@echo "==> Users can now install with: npm install -g @janreges/ai-distiller-mcp"
```

### 6. GitHub Release Naming Convention

For go-npm to work, our GitHub releases must follow this pattern:
- Tag: `v0.3.0`
- Assets: `ai-distiller_0.3.0_darwin_amd64.tar.gz`, etc.

Each tar.gz should contain just the binary (named `aid` or `aid.exe` for Windows).

### 7. Platform Mapping

go-npm uses these platform mappings:

| Node.js Platform | Go Platform | Architecture |
|-----------------|-------------|--------------|
| darwin | darwin | x64 → amd64, arm64 → arm64 |
| linux | linux | x64 → amd64, arm64 → arm64 |
| win32 | windows | x64 → amd64, arm64 → arm64 |

### 8. Testing Locally

Before publishing:

```bash
# Build binaries
make npm-build

# Test the NPM package
make npm-test

# Test npx execution
cd npm-package
npx . --version
```

### 9. Publishing Process

```bash
# 1. Update version in main project
git tag v0.3.0

# 2. Run full release
make release

# This will:
# - Run tests
# - Build cross-platform binaries
# - Create GitHub release with binaries
# - Publish NPM package
```

## Maintenance

### Updating Versions

The NPM package version should match the main binary version:

```bash
# Update both versions
VERSION=0.3.1 make release
```

### Platform Support

Currently supporting:
- macOS (Intel & Apple Silicon)
- Linux (x64 & ARM64)
- Windows (x64 & ARM64)

### Troubleshooting

1. **Binary not found after install**
   - Check `node_modules/.bin/` directory
   - Verify PATH includes npm bin directory

2. **Wrong architecture downloaded**
   - go-npm should auto-detect, but check `process.arch`
   - Manual override: `npm_config_arch=arm64 npm install`

3. **Permission errors**
   - Use `npx` instead of global install
   - Or fix npm permissions: `npm config set prefix ~/.npm`

## CI/CD Integration

Add to `.github/workflows/release.yml`:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          registry-url: 'https://registry.npmjs.org'

      - name: Build and Release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          make release
```

## Benefits

1. **Simple Installation**: `npm install -g @janreges/ai-distiller-mcp`
2. **Cross-Platform**: Automatic platform detection
3. **npx Support**: `npx @janreges/ai-distiller-mcp --help`
4. **Auto-Updates**: `npm update -g @janreges/ai-distiller-mcp`
5. **PATH Management**: NPM handles PATH automatically
6. **MCP Integration**: Works seamlessly with Claude Desktop

## References

- [go-npm Documentation](https://github.com/sanathkr/go-npm)
- [NPM Publishing Guide](https://docs.npmjs.com/cli/v8/commands/npm-publish)
- [GitHub Releases API](https://docs.github.com/en/rest/releases)
- [Model Context Protocol](https://modelcontextprotocol.io)