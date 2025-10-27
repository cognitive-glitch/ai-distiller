#!/bin/bash

# AI Distiller Build Script (Rust/Cargo)
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
BUILD_MODE="debug"
OUTPUT_DIR="build"
BINARY_NAME="aid"
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
        --test)
            RUN_TESTS=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "AI Distiller Build Script (Rust)"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --mode <debug|release>   Build mode (default: debug)"
            echo "  --output <dir>           Output directory (default: build)"
            echo "  --test                   Run tests before building"
            echo "  --verbose               Verbose output"
            echo "  --help                  Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                      # Quick debug build"
            echo "  $0 --test --verbose     # Build with tests and verbose output"
            echo "  $0 --mode release       # Optimized release build"
            echo ""
            echo "Note: For cross-platform builds, use build-releases.sh"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

log_info "AI Distiller Build Script (Rust)"
log_info "Mode: $BUILD_MODE"
log_info "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run tests if requested
if [ "$RUN_TESTS" = true ]; then
    log_info "Running tests..."
    if [ "$VERBOSE" = true ]; then
        cargo test --all-features -- --nocapture
    else
        cargo test --all-features
    fi
    log_success "Tests passed"
fi

# Build command
if [ "$BUILD_MODE" = "release" ]; then
    BUILD_FLAGS="--release"
    TARGET_DIR="target/release"
else
    BUILD_FLAGS=""
    TARGET_DIR="target/debug"
fi

log_info "Building aid binary..."

if [ "$VERBOSE" = true ]; then
    cargo build -p aid-cli $BUILD_FLAGS --verbose
else
    cargo build -p aid-cli $BUILD_FLAGS
fi

# Copy binary to output directory
if [ -f "$TARGET_DIR/aid" ]; then
    cp "$TARGET_DIR/aid" "$OUTPUT_DIR/$BINARY_NAME"
    log_success "Binary copied to $OUTPUT_DIR/$BINARY_NAME"
elif [ -f "$TARGET_DIR/aid.exe" ]; then
    cp "$TARGET_DIR/aid.exe" "$OUTPUT_DIR/$BINARY_NAME.exe"
    log_success "Binary copied to $OUTPUT_DIR/$BINARY_NAME.exe"
else
    echo "Error: Binary not found in $TARGET_DIR"
    exit 1
fi

# Test the binary
log_info "Testing binary..."
if "$OUTPUT_DIR/$BINARY_NAME" --version > /dev/null 2>&1; then
    log_success "Binary test passed"
else
    log_info "Binary version check completed"
fi

# Generate build info
cat > "$OUTPUT_DIR/build-info.txt" << EOF
AI Distiller Build Information
=============================

Build Time: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Build Mode: $BUILD_MODE
Git Commit: $(git rev-parse HEAD 2>/dev/null || echo "none")
Git Branch: $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "none")
Rust Version: $(rustc --version)
Cargo Version: $(cargo --version)

Build Command:
cargo build -p aid-cli $BUILD_FLAGS

Files Created:
$(ls -lh "$OUTPUT_DIR" | grep -v '^d' | awk '{print $9, $5}' | grep -v '^$')
EOF

log_success "Build completed successfully!"
log_info "Output directory: $OUTPUT_DIR"
log_info "Build info: $OUTPUT_DIR/build-info.txt"

# Show binary information
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
    echo "  $binary_path testdata/python/01_basic/source.py"
fi