#!/bin/bash

# AI Distiller Build Script
# Local development build with various options

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Default values
BUILD_MODE="development"
OUTPUT_DIR="build"
BINARY_NAME="aid"
PLATFORMS="local"
RUN_TESTS=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        --test)
            RUN_TESTS=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "AI Distiller Build Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --mode <dev|release>     Build mode (default: development)"
            echo "  --output <dir>           Output directory (default: build)"
            echo "  --platforms <list>       Platforms to build for (default: local)"
            echo "                          Options: local, all, linux, darwin, windows"
            echo "  --test                   Run tests before building"
            echo "  --verbose               Verbose output"
            echo "  --help                  Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                      # Quick local build"
            echo "  $0 --test --verbose     # Build with tests and verbose output"
            echo "  $0 --mode release --platforms all  # Release build for all platforms"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

log_info "AI Distiller Build Script"
log_info "Mode: $BUILD_MODE"
log_info "Output: $OUTPUT_DIR"
log_info "Platforms: $PLATFORMS"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests if requested
if [ "$RUN_TESTS" = true ]; then
    log_info "Running tests..."
    if [ "$VERBOSE" = true ]; then
        go test -v ./...
    else
        go test ./...
    fi
    log_success "Tests passed"
fi

# Set build flags based on mode
if [ "$BUILD_MODE" = "release" ]; then
    LDFLAGS="-s -w -X 'main.version=dev-$(git rev-parse --short HEAD)' -X 'main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.commit=$(git rev-parse HEAD)'"
    CGO_ENABLED=0
else
    LDFLAGS="-X 'main.version=dev' -X 'main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    CGO_ENABLED=1
fi

# Define platform list
case $PLATFORMS in
    local)
        PLATFORM_LIST=("$(go env GOOS)/$(go env GOARCH)")
        ;;
    all)
        PLATFORM_LIST=(
            "linux/amd64"
            "linux/arm64"
            "darwin/amd64"
            "darwin/arm64"
            "windows/amd64"
        )
        ;;
    linux)
        PLATFORM_LIST=(
            "linux/amd64"
            "linux/arm64"
        )
        ;;
    darwin)
        PLATFORM_LIST=(
            "darwin/amd64"
            "darwin/arm64"
        )
        ;;
    windows)
        PLATFORM_LIST=(
            "windows/amd64"
        )
        ;;
    *)
        # Custom platform list
        IFS=',' read -ra PLATFORM_LIST <<< "$PLATFORMS"
        ;;
esac

# Build for each platform
for platform in "${PLATFORM_LIST[@]}"; do
    IFS='/' read -r os arch <<< "$platform"

    output_name="$BINARY_NAME"
    if [ "$os" = "windows" ]; then
        output_name="$output_name.exe"
    fi

    # Add platform suffix for multi-platform builds
    if [ ${#PLATFORM_LIST[@]} -gt 1 ]; then
        if [ "$os" = "windows" ]; then
            output_name="$BINARY_NAME-$os-$arch.exe"
        else
            output_name="$BINARY_NAME-$os-$arch"
        fi
    fi

    output_path="$OUTPUT_DIR/$output_name"

    log_info "Building for $os/$arch -> $output_path"

    build_cmd="GOOS=$os GOARCH=$arch CGO_ENABLED=$CGO_ENABLED go build"

    if [ "$VERBOSE" = true ]; then
        build_cmd="$build_cmd -v"
    fi

    build_cmd="$build_cmd -ldflags=\"$LDFLAGS\" -o \"$output_path\" ./cmd/aid/"

    if [ "$VERBOSE" = true ]; then
        echo "Executing: $build_cmd"
    fi

    eval $build_cmd

    # Test the binary if it's for the local platform
    if [ "$platform" = "$(go env GOOS)/$(go env GOARCH)" ]; then
        log_info "Testing binary..."
        if "$output_path" --version > /dev/null 2>&1; then
            log_success "Binary test passed"
        else
            log_info "Binary version check failed (this is normal for dev builds)"
        fi
    fi
done

# Generate build info
cat > "$OUTPUT_DIR/build-info.txt" << EOF
AI Distiller Build Information
=============================

Build Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Build Mode: $BUILD_MODE
Git Commit: $(git rev-parse HEAD)
Git Branch: $(git rev-parse --abbrev-ref HEAD)
Go Version: $(go version)

Platforms Built:
$(printf '%s\n' "${PLATFORM_LIST[@]}")

Build Command:
LDFLAGS="$LDFLAGS"
CGO_ENABLED=$CGO_ENABLED

Files Created:
$(ls -la "$OUTPUT_DIR" | grep -v '^d' | awk '{print $9, $5}' | grep -v '^$')
EOF

log_success "Build completed successfully!"
log_info "Output directory: $OUTPUT_DIR"
log_info "Build info: $OUTPUT_DIR/build-info.txt"

# Show binary information for local builds
if [ "$PLATFORMS" = "local" ]; then
    binary_path="$OUTPUT_DIR/$BINARY_NAME"
    if [ -f "$binary_path.exe" ]; then
        binary_path="$binary_path.exe"
    fi

    if [ -f "$binary_path" ]; then
        log_info "Binary size: $(du -h "$binary_path" | cut -f1)"
        echo ""
        log_info "Quick test:"
        echo "  $binary_path --help"
        echo "  $binary_path --version"
        echo "  $binary_path . --format text --strip comments"
    fi
fi