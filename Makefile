.PHONY: all build test bench lint clean install cross-compile

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

# Default target
.DEFAULT_GOAL := build