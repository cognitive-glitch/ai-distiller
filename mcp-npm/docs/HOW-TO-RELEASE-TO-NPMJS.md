# How to Release AI Distiller MCP to NPM

This guide documents the complete process for releasing the AI Distiller MCP server to npmjs.org.

## Prerequisites

1. **NPM Account**: Ensure you're logged in to npm:
   ```bash
   npm login
   # Username: janreges
   ```

2. **AI Distiller Binary Release**: The `aid` binary must be released on GitHub first:
   - Release URL format: `https://github.com/janreges/ai-distiller/releases/tag/vX.Y.Z`
   - Required files for each platform:
     - `aid-darwin-amd64-vX.Y.Z.tar.gz`
     - `aid-darwin-arm64-vX.Y.Z.tar.gz`
     - `aid-linux-amd64-vX.Y.Z.tar.gz`
     - `aid-linux-arm64-vX.Y.Z.tar.gz`
     - `aid-windows-amd64-vX.Y.Z.zip`

## Release Process

### 1. Update Version Numbers

Update the version in both places to match the `aid` binary version:

```bash
cd mcp-npm/

# Edit package.json - update version field
vim package.json

# Edit postinstall.js - update VERSION constant
vim scripts/postinstall.js
```

### 2. Test Locally

Before publishing, test the package installation locally:

```bash
# Create a test package
npm pack

# This creates: janreges-ai-distiller-mcp-X.Y.Z.tgz

# Test in a clean directory
mkdir /tmp/test-mcp
cd /tmp/test-mcp
npm install /path/to/mcp-npm/janreges-ai-distiller-mcp-X.Y.Z.tgz

# Verify the binary was downloaded and extracted
ls -la node_modules/@janreges/ai-distiller-mcp/bin/

# Test the MCP server
npx aid-mcp --help
```

### 3. Dry Run

Always do a dry run first to see what will be published:

```bash
cd mcp-npm/
npm publish --dry-run
```

Review the file list carefully. Should include:
- `package.json`
- `mcp-server.js`
- `scripts/postinstall.js`
- `bin/` (empty directory, will be populated during install)
- `README.md`
- `LICENSE`

### 4. Publish to NPM

When ready, publish with public access (required for scoped packages):

```bash
npm publish --access public
```

### 5. Verify Installation

Test the published package:

```bash
# In a new directory
npx @janreges/ai-distiller-mcp --version

# Or install globally
npm install -g @janreges/ai-distiller-mcp
aid-mcp --version
```

### 6. Tag and Commit

After successful publish, tag the release in git:

```bash
git add .
git commit -m "chore: release AI Distiller MCP vX.Y.Z"
git tag mcp-vX.Y.Z
git push && git push --tags
```

## Version Synchronization

The MCP package version should generally match the `aid` binary version for clarity:
- aid binary: v1.0.0 → MCP package: 1.0.0
- aid binary: v1.1.0 → MCP package: 1.1.0

However, if you need to fix bugs in the MCP wrapper without a new aid release:
- aid binary: v1.0.0 → MCP package: 1.0.1, 1.0.2, etc.

## Troubleshooting

### Download Fails During Install

1. Check GitHub release exists and files are named correctly
2. Verify VERSION constant in postinstall.js matches the release tag
3. Check network connectivity and proxy settings

### Binary Not Found After Install

1. Check if postinstall script ran: `npm ls @janreges/ai-distiller-mcp`
2. Manually run postinstall: `cd node_modules/@janreges/ai-distiller-mcp && npm run postinstall`
3. Check permissions on extracted binary

### Platform Not Supported

The postinstall script supports:
- macOS: x64 (Intel), arm64 (Apple Silicon)
- Linux: x64, arm64
- Windows: x64

Other platforms will show an error during installation.

## CI/CD Considerations

For users in CI environments:

1. **Caching**: The downloaded binary is cached in `node_modules`. CI systems should cache this directory.

2. **Offline Install**: After first install, the package works offline if node_modules is preserved.

3. **Skip Download**: Users can skip the download with `npm install --ignore-scripts` but the MCP server won't work.

## Security Notes

Currently, the binaries are downloaded without checksum verification. Future improvements:

1. Add SHASUMS256.txt to aid releases
2. Verify checksums in postinstall.js
3. Sign releases with GPG for additional security

## Release Checklist

- [ ] AI Distiller binary released on GitHub
- [ ] Version numbers updated in package.json and postinstall.js
- [ ] Local testing with `npm pack` passed
- [ ] Dry run shows correct files
- [ ] Published to npm with `--access public`
- [ ] Installation test from npm registry passed
- [ ] Git tagged and pushed
- [ ] Release notes updated on GitHub (optional)