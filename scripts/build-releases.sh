#!/bin/bash

# Build script for creating releases using Rust cross-compilation
# Requires: cargo, cross (cargo install cross)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Read version
VERSION=$(cat VERSION)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build directory
BUILD_DIR="build/releases"
mkdir -p $BUILD_DIR

echo -e "${GREEN}Building AI Distiller v$VERSION (Rust)${NC}"
echo ""

# Check if cross is installed
if ! command -v cross &> /dev/null; then
    echo -e "${YELLOW}Warning: 'cross' not found. Install with: cargo install cross${NC}"
    echo -e "${YELLOW}Falling back to native cargo build for current platform only${NC}"
    USE_CROSS=false
else
    USE_CROSS=true
fi

# Function to build for a platform
build_platform() {
    local TARGET=$1
    local PLATFORM_NAME=$2
    local EXT=$3
    local BINARY_NAME="aid$EXT"
    local TEMP_DIR="$BUILD_DIR/temp-$PLATFORM_NAME"

    echo -e "${YELLOW}Building for $PLATFORM_NAME ($TARGET)...${NC}"

    # Create temp directory for this platform
    mkdir -p "$TEMP_DIR"

    # Build command
    if [ "$USE_CROSS" = true ]; then
        cross build --release --target $TARGET -p aid-cli
    else
        cargo build --release --target $TARGET -p aid-cli
    fi

    if [ $? -eq 0 ]; then
        # Copy binary to temp directory
        cp "target/$TARGET/release/aid$EXT" "$TEMP_DIR/$BINARY_NAME"

        echo -e "${GREEN}✓ Built: $BINARY_NAME for $PLATFORM_NAME${NC}"

        # Create archive
        cd "$TEMP_DIR"
        if [[ "$PLATFORM_NAME" == *"windows"* ]]; then
            # Create ZIP for Windows
            zip -q "../aid-$PLATFORM_NAME-v$VERSION.zip" "$BINARY_NAME"
            echo -e "${GREEN}✓ Created: aid-$PLATFORM_NAME-v$VERSION.zip${NC}"
        else
            # Create tar.gz for Unix systems
            tar -czf "../aid-$PLATFORM_NAME-v$VERSION.tar.gz" "$BINARY_NAME"
            echo -e "${GREEN}✓ Created: aid-$PLATFORM_NAME-v$VERSION.tar.gz${NC}"
        fi
        cd - > /dev/null

        # Clean up temp directory
        rm -rf "$TEMP_DIR"
    else
        echo -e "${RED}✗ Failed to build $PLATFORM_NAME${NC}"
    fi
    echo ""
}

# Define target platforms
# Format: "rust-target" "platform-name" "extension"
PLATFORMS=(
    "x86_64-unknown-linux-gnu:linux-amd64:"
    "aarch64-unknown-linux-gnu:linux-arm64:"
    "x86_64-apple-darwin:darwin-amd64:"
    "aarch64-apple-darwin:darwin-arm64:"
    "x86_64-pc-windows-gnu:windows-amd64:.exe"
)

# Build for each platform
for platform_spec in "${PLATFORMS[@]}"; do
    IFS=':' read -r target name ext <<< "$platform_spec"

    # Check if we can build this target
    if [ "$USE_CROSS" = true ] || rustup target list --installed | grep -q "$target"; then
        build_platform "$target" "$name" "$ext"
    else
        echo -e "${YELLOW}Skipping $name - target $target not installed${NC}"
        echo -e "${YELLOW}Install with: rustup target add $target${NC}"
        echo ""
    fi
done

echo -e "${GREEN}Build complete!${NC}"
echo ""
echo "Release archives created in: $BUILD_DIR"
echo ""

# List created archives
if ls $BUILD_DIR/*.tar.gz $BUILD_DIR/*.zip 2>/dev/null >/dev/null; then
    echo "Created archives:"
    ls -lh $BUILD_DIR/*.tar.gz $BUILD_DIR/*.zip 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
    echo ""

    # Generate checksums
    echo "Generating checksums..."
    cd $BUILD_DIR
    sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt
    echo -e "${GREEN}✓ Created checksums.txt${NC}"
    cd - > /dev/null
else
    echo -e "${YELLOW}No archives were created. Check build errors above.${NC}"
fi

echo ""
echo "To enable cross-compilation:"
echo "  cargo install cross"
echo ""
echo "Or install specific targets:"
echo "  rustup target add x86_64-unknown-linux-gnu"
echo "  rustup target add aarch64-unknown-linux-gnu"
echo "  rustup target add x86_64-apple-darwin"
echo "  rustup target add aarch64-apple-darwin"
echo "  rustup target add x86_64-pc-windows-gnu"