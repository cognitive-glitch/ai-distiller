.PHONY: all build test bench lint clean install cross-compile test-parser test-performance aid

# Variables
BINARY_NAME = aid
BUILD_DIR = build
INSTALL_DIR = /usr/local/bin
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = gofmt
GOLINT = golangci-lint

# Platforms for cross-compilation
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

all: test build

# Build for current platform
build:
	@echo "==> Building $(BINARY_NAME) $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aid

# Run tests
test:
	@echo "==> Running tests"
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run integration tests with new test runner
test-integration:
	@echo "==> Running integration tests"
	$(GOTEST) -v ./internal/testrunner

# Update expected test files
test-update:
	@echo "==> Updating expected test files"
	UPDATE_EXPECTED=true $(GOTEST) -v ./internal/testrunner

# Regenerate all expected test files
test-regenerate:
	@echo "==> Regenerating all expected test files"
	@./scripts/regenerate-expected.sh

# Generate expected test data for all languages (includes building aid)
generate-expected-testdata: build
	@echo "==> Building aid and regenerating all expected test files"
	@chmod +x regenerate_expected.sh
	@./regenerate_expected.sh

# Audit test structure for consistency issues
test-audit:
	@echo "==> Auditing test structure"
	@go run -ldflags "-X main.mode=audit" ./cmd/test-audit

# Run benchmarks
bench:
	@echo "==> Running benchmarks"
	$(GOTEST) -bench=. -benchmem ./...

# Run linter
lint:
	@echo "==> Running linter"
	$(GOLINT) run ./...

# Format code
fmt:
	@echo "==> Formatting code"
	$(GOFMT) -s -w .

# Install binary
install: build
	@echo "==> Installing $(BINARY_NAME) to $(INSTALL_DIR)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)

# Clean build artifacts
clean:
	@echo "==> Cleaning build artifacts"
	@rm -rf $(BUILD_DIR) coverage.txt

# Download dependencies
deps:
	@echo "==> Downloading dependencies"
	$(GOMOD) download
	$(GOMOD) tidy

# Cross-compile for all platforms
cross-compile: $(PLATFORMS)

$(PLATFORMS):
	@echo "==> Building for $@"
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@))$(if $(findstring windows,$(word 1,$(subst /, ,$@))),.exe) ./cmd/aid

# Build WASM modules
build-wasm:
	@echo "==> Building WASM modules"
	@./scripts/build-wasm.sh

# Development setup
setup:
	@echo "==> Setting up development environment"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOMOD) download

# Initialize development environment with all dependencies
dev-init:
	@echo "==> Initializing development environment..."
	@echo "  - Downloading Go module dependencies"
	@$(GOMOD) download
	@echo "  - Installing testing dependencies"
	@go get -t ./...
	@echo "  - Installing development tools"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/segmentio/golines@latest
	@go install github.com/air-verse/air@latest
	@echo "  - Running go mod tidy"
	@$(GOMOD) tidy
	@echo "==> Development environment initialized successfully!"
	@echo "==> Run 'make test' to verify everything is working"

# Run the application
run: build
	@echo "==> Running $(BINARY_NAME)"
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Quick build and run for development
aid:
	@if $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aid 2>&1 | grep -E "(error|cannot|undefined)" >&2; then \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(BINARY_NAME) $(filter-out $@,$(MAKECMDGOALS))

# Catch-all target to allow passing arguments to aid
%:
	@:

# Test parser functionality
test-parser:
	@echo "==> Running parser functional tests"
	@go run ./cmd/parser-test

# Test performance optimizations
test-performance:
	@echo "==> Running performance tests"
	@go run ./cmd/performance-test $(PERF_ARGS)

# Test semantic analysis
test-semantic:
	@echo "==> Running semantic analysis tests"
	@go run ./cmd/semantic-test

# Test semantic resolver (Pass 2)
test-resolver:
	@echo "==> Running semantic resolver tests"
	@go run ./cmd/semantic-resolver-test

# Run performance comparison
perf-compare:
	@echo "==> Running performance mode comparison"
	@go run ./cmd/performance-test -mode=comparison

# Find optimal configuration
perf-optimize:
	@echo "==> Finding optimal performance configuration"
	@go run ./cmd/performance-test -mode=config

# NPM package version (read from npm/package.json)
NPM_VERSION := $(shell node -p "require('./npm/package.json').version" 2>/dev/null || echo "0.0.0")

# NPM release preparation
npm-prepare:
	@echo "==> Preparing NPM package (version $(NPM_VERSION))"
	@mkdir -p npm
	@cp LICENSE npm/ 2>/dev/null || echo "No LICENSE file found"
	@echo "==> NPM package prepared in ./npm"

# Build release binaries for NPM
npm-build-binaries:
	@echo "==> Building release binaries for all platforms"
	@mkdir -p dist
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/aid-linux-amd64 ./cmd/aid
	cd dist && tar -czf aid-linux-amd64.tar.gz aid-linux-amd64
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/aid-linux-arm64 ./cmd/aid
	cd dist && tar -czf aid-linux-arm64.tar.gz aid-linux-arm64
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/aid-darwin-amd64 ./cmd/aid
	cd dist && tar -czf aid-darwin-amd64.tar.gz aid-darwin-amd64
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/aid-darwin-arm64 ./cmd/aid
	cd dist && tar -czf aid-darwin-arm64.tar.gz aid-darwin-arm64
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/aid-windows-amd64.exe ./cmd/aid
	cd dist && tar -czf aid-windows-amd64.tar.gz aid-windows-amd64.exe
	# Windows ARM64
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/aid-windows-arm64.exe ./cmd/aid
	cd dist && tar -czf aid-windows-arm64.tar.gz aid-windows-arm64.exe
	@echo "==> Release binaries built in ./dist"

# Test NPM package locally
npm-test-local:
	@echo "==> Testing NPM package locally"
	cd npm && npm pack
	@echo "==> Install the package locally with: npm install -g ./npm/*.tgz"

# Publish to NPM (requires npm login)
npm-publish:
	@echo "==> Publishing to NPM"
	@echo "==> Checking NPM login status..."
	@npm whoami || (echo "Error: Not logged in to NPM. Run 'npm login' first." && exit 1)
	cd npm && npm publish --access public
	@echo "==> Published @janreges/ai-distiller-mcp@$(NPM_VERSION) to NPM"

# Update NPM version
npm-version:
	@echo "==> Current NPM version: $(NPM_VERSION)"
	@echo "==> To update version, run: cd npm && npm version <patch|minor|major>"

# Complete NPM release process
npm-release: npm-prepare npm-build-binaries
	@echo "==> Creating GitHub release v$(NPM_VERSION)"
	@echo "==> Upload binaries from ./dist to GitHub release"
	@echo "==> Then run 'make npm-publish' to publish to NPM"

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  test            - Run tests"
	@echo "  bench           - Run benchmarks"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  install         - Install binary to system"
	@echo "  clean           - Remove build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  cross-compile   - Build for all platforms"
	@echo "  build-wasm      - Build WASM modules"
	@echo "  setup           - Set up development environment"
	@echo "  dev-init        - Initialize dev environment with all dependencies"
	@echo "  run ARGS=...    - Run the application with arguments"
	@echo "  test-parser     - Run parser functional tests"
	@echo "  test-performance- Run performance tests"
	@echo "  test-semantic   - Run semantic analysis tests"
	@echo "  test-resolver   - Run semantic resolver tests"
	@echo "  generate-expected-testdata - Build aid and regenerate all expected test files"
	@echo "  perf-compare    - Compare performance modes"
	@echo "  perf-optimize   - Find optimal configuration"
	@echo ""
	@echo "NPM Distribution targets:"
	@echo "  npm-prepare     - Prepare NPM package structure"
	@echo "  npm-build-binaries - Build binaries for all platforms"
	@echo "  npm-test-local  - Create local NPM package for testing"
	@echo "  npm-publish     - Publish package to NPM registry"
	@echo "  npm-version     - Show current NPM package version"
	@echo "  npm-release     - Full release process (prepare + build)"

# Default target
.DEFAULT_GOAL := build