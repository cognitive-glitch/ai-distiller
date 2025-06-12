# WASM Tree-sitter PoC

This proof of concept demonstrates using tree-sitter with WebAssembly via Wazero.

## Prerequisites

- Go 1.21+
- emscripten (for compiling tree-sitter to WASM)

## Architecture

- Uses Wazero for WASM runtime (pure Go, no CGo)
- Loads tree-sitter parsers compiled to WASM
- No cross-compilation issues (pure Go)

## Setup

```bash
# Download dependencies
go mod download

# Build tree-sitter-python WASM module
./setup.sh
```

## Building

```bash
# Build for current platform
make build

# Run tests
make test

# Run benchmarks
make bench

# Cross-compile for all platforms (trivial with pure Go)
make cross
```

## Performance Targets

- Startup time: <50ms (including WASM initialization)
- Parse throughput: e50% of CGo version
- Binary size: <50MB (including embedded WASM)