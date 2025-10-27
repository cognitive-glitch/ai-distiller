#!/bin/bash

# Build script for AI Distiller MCP server
# Creates binaries for all supported platforms

set -e

echo "Building AI Distiller MCP server for all platforms..."

# Get version from package.json
VERSION=$(node -p "require('./package.json').version")
echo "Version: $VERSION"

# Create bin directory
mkdir -p bin

# Build matrix
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r OS ARCH <<< "$PLATFORM"

    OUTPUT="bin/aid-mcp-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    echo "Building for $OS/$ARCH..."

    GOOS=$OS GOARCH=$ARCH go build \
        -ldflags "-s -w -X main.serverVersion=$VERSION" \
        -o "$OUTPUT" \
        ./cmd/aid-mcp

    # Compress for distribution
    if [ "$OS" = "windows" ]; then
        zip -j "bin/aid-mcp_${VERSION}_${OS}_${ARCH}.zip" "$OUTPUT"
    else
        tar -czf "bin/aid-mcp_${VERSION}_${OS}_${ARCH}.tar.gz" -C bin "$(basename $OUTPUT)"
    fi
done

echo "Build complete! Binaries available in bin/"
ls -la bin/