# CGo Tree-sitter PoC

This proof of concept demonstrates using tree-sitter with CGo for parsing.

## Prerequisites

- Go 1.21+
- C compiler (gcc/clang)
- Zig (for cross-compilation)

## Setup

```bash
# Download dependencies
go mod download

# Clone tree-sitter-python grammar
git submodule add https://github.com/tree-sitter/tree-sitter-python grammars/tree-sitter-python
git submodule update --init
```

## Building

```bash
# Build for current platform
make build

# Run tests
make test

# Run benchmarks
make bench

# Cross-compile for all platforms
make cross
```

## Architecture

- Uses official `github.com/tree-sitter/go-tree-sitter`
- Embeds tree-sitter-python grammar via CGo
- Cross-compiles using Zig CC

## Performance Targets

- Startup time: <50ms
- Parse throughput: Baseline
- Binary size: <50MB