# How to Release AI Distiller MCP to NPM

This guide documents the complete process for releasing the AI Distiller MCP server to npmjs.org.

## Quick Release Steps

For experienced users, here's the minimal process:

```bash
cd mcp-npm/

# 1. Update version in package.json (other files now read it automatically)
# 2. Run the publish script
./publish.sh

# 3. Tag the release
git add . && git commit -m "chore: release AI Distiller MCP v$(node -p "require('./package.json').version")"
git tag mcp-v$(node -p "require('./package.json').version") && git push && git push --tags
```

## Prerequisites

1. **NPM Account**: Ensure you're logged in to npm:
   ```bash
   npm login
   # Username: janreges
   ```

2. **AI Distiller Binary Release**: The `aid` binary must be released on GitHub first:
   - Release URL format: `https://github.com/cognitive-glitch/ai-distiller-reboot/releases/tag/vX.Y.Z`
   - Required files for each platform:
     - `aid-darwin-amd64-vX.Y.Z.tar.gz`
     - `aid-darwin-arm64-vX.Y.Z.tar.gz`
     - `aid-linux-amd64-vX.Y.Z.tar.gz`
     - `aid-linux-arm64-vX.Y.Z.tar.gz`
     - `aid-windows-amd64-vX.Y.Z.zip`

## Release Process

There are two ways to release: using the automated `publish.sh` script (recommended) or manually.

## Option 1: Automated Release with publish.sh (Recommended)

The `publish.sh` script automates the entire release process and includes safety checks:
- Verifies npm login status
- Ensures TypeScript builds successfully
- Shows package contents preview before publishing
- Requires explicit confirmation before publishing
- Handles all build steps automatically

### 1. Update Version Number

The version is now read dynamically from `package.json`. You only need to update it in one place:

```bash
cd mcp-npm/

# Edit package.json and update the "version" field
vim package.json
```

### 2. Run the Publish Script

The `publish.sh` script handles the entire build and publish process:

```bash
./publish.sh
```

The script will:
1. Check that you're logged in to npm
2. Clean previous builds
3. Install dependencies
4. Build TypeScript
5. Show package contents preview
6. Ask for confirmation before publishing
7. Publish to npm with public access

### 3. Verify Installation

After successful publish, test the package:

```bash
# In a new directory
npx @janreges/ai-distiller-mcp@latest

# Or install globally
npm install -g @janreges/ai-distiller-mcp@latest
aid-mcp --version
```

### 4. Tag and Commit

After successful publish, tag the release in git:

```bash
git add .
git commit -m "chore: release AI Distiller MCP v$(node -p "require('./package.json').version")"
git tag mcp-v$(node -p "require('./package.json').version")
git push && git push --tags
```

## Option 2: Manual Release Process

### 1. Update Version Numbers

Same as Option 1 above.

### 2. Build TypeScript

```bash
cd mcp-npm/
npm install
npm run build
```

### 3. Test Locally

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
npx @janreges/ai-distiller-mcp
```

### 4. Dry Run

Always do a dry run first to see what will be published:

```bash
cd mcp-npm/
npm publish --dry-run
```

Review the file list carefully. Should include:
- `package.json`
- `dist/` (compiled TypeScript files)
- `mcp-server-wrapper.js`
- `scripts/postinstall.js`
- `bin/` (empty directory, will be populated during install)
- `README.md`
- `LICENSE`

### 5. Publish to NPM

When ready, publish with public access (required for scoped packages):

```bash
npm publish --access public
```

### 6. Follow steps 3-4 from Option 1

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