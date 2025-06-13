#!/bin/bash

# AI Distiller Release Script
# Creates a new release with proper versioning and builds

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Not in a git repository"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    log_error "Working directory is not clean. Please commit your changes."
    exit 1
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
log_info "Current version: $CURRENT_VERSION"

# Parse version components
if [[ $CURRENT_VERSION =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(-.*)?$ ]]; then
    MAJOR=${BASH_REMATCH[1]}
    MINOR=${BASH_REMATCH[2]}
    PATCH=${BASH_REMATCH[3]}
    PRERELEASE=${BASH_REMATCH[4]}
else
    log_warning "Invalid version format, starting from v0.1.0"
    MAJOR=0
    MINOR=1
    PATCH=0
    PRERELEASE=""
fi

# Prompt for version bump type
echo ""
echo "Select version bump type:"
echo "  1) Patch (v$MAJOR.$MINOR.$((PATCH+1)))"
echo "  2) Minor (v$MAJOR.$((MINOR+1)).0)"
echo "  3) Major (v$((MAJOR+1)).0.0)"
echo "  4) Custom version"
echo "  5) Pre-release (v$MAJOR.$MINOR.$PATCH-alpha.$(date +%Y%m%d))"

read -p "Enter choice [1-5]: " choice

case $choice in
    1)
        NEW_VERSION="v$MAJOR.$MINOR.$((PATCH+1))"
        ;;
    2)
        NEW_VERSION="v$MAJOR.$((MINOR+1)).0"
        ;;
    3)
        NEW_VERSION="v$((MAJOR+1)).0.0"
        ;;
    4)
        read -p "Enter custom version (e.g., v1.2.3): " NEW_VERSION
        if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
            log_error "Invalid version format. Use semantic versioning (v1.2.3)"
            exit 1
        fi
        ;;
    5)
        NEW_VERSION="v$MAJOR.$MINOR.$PATCH-alpha.$(date +%Y%m%d)"
        ;;
    *)
        log_error "Invalid choice"
        exit 1
        ;;
esac

log_info "New version will be: $NEW_VERSION"

# Confirm release
read -p "Continue with release $NEW_VERSION? [y/N]: " confirm
if [[ ! $confirm =~ ^[Yy]$ ]]; then
    log_info "Release cancelled"
    exit 0
fi

# Run tests before release
log_info "Running tests..."
if ! go test ./...; then
    log_error "Tests failed. Release cancelled."
    exit 1
fi

log_success "All tests passed"

# Build binaries for verification
log_info "Building binaries for verification..."
BUILD_DIR="build/release-$NEW_VERSION"
mkdir -p "$BUILD_DIR"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    binary_name="aid"
    if [ "$os" = "windows" ]; then
        binary_name="aid.exe"
    fi
    
    output_name="aid-$NEW_VERSION-$os-$arch"
    if [ "$os" = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    log_info "Building for $os/$arch..."
    
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
        -ldflags="-s -w -X 'main.version=$NEW_VERSION' -X 'main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.commit=$(git rev-parse HEAD)'" \
        -o "$BUILD_DIR/$output_name" \
        ./cmd/aid/
done

log_success "All binaries built successfully"

# Generate checksums
log_info "Generating checksums..."
cd "$BUILD_DIR"
sha256sum aid-* > checksums.txt
cd - > /dev/null

# Generate changelog
log_info "Generating changelog..."
CHANGELOG_FILE="$BUILD_DIR/CHANGELOG.md"

if [ "$CURRENT_VERSION" != "v0.0.0" ]; then
    echo "## Changes since $CURRENT_VERSION" > "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
    git log --pretty="- %s (%h)" "$CURRENT_VERSION"..HEAD >> "$CHANGELOG_FILE"
else
    echo "## Initial Release" > "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
    echo "- Initial release of AI Distiller" >> "$CHANGELOG_FILE"
fi

# Add feature summary
cat >> "$CHANGELOG_FILE" << 'EOF'

## Key Features

### 🚀 Core Functionality
- **Multi-language support**: Python, JavaScript, TypeScript, Go, Java, C#, Rust
- **Fast processing**: 10MB codebase in <2 seconds  
- **Multiple output formats**: Markdown, Text, JSON, JSONL, XML
- **Flexible stripping**: Remove comments, implementations, private members

### 🔧 Advanced Features
- **Semantic analysis**: Symbol extraction and call graph generation
- **Tree-sitter parsing**: Accurate AST-based code analysis
- **Performance optimization**: Concurrent processing and intelligent caching
- **Single binary**: No runtime dependencies

### 📊 Usage Examples
```bash
aid                                    # Process current directory
aid src/                              # Process src directory  
aid --strip comments,implementation   # Remove comments and implementations
aid --format json --output api.json   # JSON output to file
aid --strip non-public --stdout       # Print only public members
```

EOF

# Create git tag
log_info "Creating git tag..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"

# Push tag to trigger release workflow
log_info "Pushing tag to origin..."
git push origin "$NEW_VERSION"

log_success "Release $NEW_VERSION initiated!"
log_info "GitHub Actions will now:"
log_info "  1. Run comprehensive tests"
log_info "  2. Build binaries for all platforms"
log_info "  3. Create GitHub release with binaries"
log_info "  4. Build and push Docker image (if configured)"

echo ""
log_info "Local release artifacts created in: $BUILD_DIR"
log_info "View release progress at: https://github.com/janreges/ai-distiller/actions"

echo ""
log_success "Release process completed!"