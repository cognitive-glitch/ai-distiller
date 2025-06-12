.PHONY: all build test bench cross clean

# Variables
BINARY_NAME = poc-wasm
BUILD_DIR = build
GO_FILES = $(shell find . -name '*.go' -not -path "./wasm/*")

# Platform targets - Pure Go makes this trivial!
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64

all: build

build:
	@echo "==> Building $(BINARY_NAME) for current platform"
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

test:
	@echo "==> Running tests"
	go test -v ./...

bench:
	@echo "==> Running benchmarks"
	go test -bench=. -benchmem ./...

# Cross-compilation is trivial with pure Go
cross: $(PLATFORMS)

linux/amd64:
	@echo "==> Building for linux/amd64"
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd

linux/arm64:
	@echo "==> Building for linux/arm64"
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd

windows/amd64:
	@echo "==> Building for windows/amd64"
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd

windows/arm64:
	@echo "==> Building for windows/arm64"
	GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd

darwin/amd64:
	@echo "==> Building for darwin/amd64"
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd

darwin/arm64:
	@echo "==> Building for darwin/arm64"
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd

# Benchmark specific targets
bench-startup:
	@echo "==> Benchmarking startup time (including WASM init)"
	time $(BUILD_DIR)/$(BINARY_NAME) --help

bench-parse:
	@echo "==> Benchmarking parse performance"
	go test -bench=BenchmarkParse -benchmem ./internal/parser

# Size analysis
size:
	@echo "==> Binary size analysis"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)*

clean:
	@echo "==> Cleaning build artifacts"
	rm -rf $(BUILD_DIR)

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  test        - Run tests"
	@echo "  bench       - Run all benchmarks"
	@echo "  cross       - Cross-compile for all platforms (pure Go!)"
	@echo "  clean       - Remove build artifacts"
	@echo "  size        - Show binary sizes"