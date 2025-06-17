# AI Distiller Proof of Concepts

This directory contains two proof-of-concept implementations to evaluate the best parsing technology for AI Distiller:

## PoC Implementations

### 1. CGo Approach (`/cgo`)
- Uses official `github.com/tree-sitter/go-tree-sitter` bindings
- Requires CGo for native C library integration
- Cross-compilation via Zig CC toolchain
- Expected: Best performance, complex build process

### 2. WASM Approach (`/wasm`)
- Uses Wazero runtime to execute tree-sitter WASM modules
- Pure Go implementation (no CGo required)
- Sandboxed execution for security
- Expected: Simpler build, potential performance overhead

## Evaluation Criteria

Both PoCs will be evaluated on:

1. **Build Complexity**
   - Time to set up build environment
   - Cross-compilation ease
   - CI/CD integration complexity

2. **Runtime Performance**
   - Startup time (must be <50ms)
   - Parse throughput (files/second)
   - Memory usage

3. **Binary Characteristics**
   - Final binary size (must be <50MB)
   - Dependency requirements
   - Platform compatibility

4. **Grammar Support**
   - Compatibility with all target grammars
   - Ease of adding new languages

## Running the PoCs

Each PoC includes:
- `make build` - Build the binary
- `make test` - Run tests
- `make bench` - Run benchmarks
- `make cross` - Cross-compile for all platforms

See individual README files in each directory for specific instructions.

## Decision Matrix

| Metric | CGo Target | WASM Target | Winner |
|--------|-----------|-------------|---------|
| Build Complexity | - | < 25% of CGo | TBD |
| Binary Size | <50MB | <50MB | TBD |
| Cold Start | <50ms | <50ms | TBD |
| Parse Speed | Baseline | ≥50% of CGo | TBD |
| Security | Native | Sandboxed | TBD |

The final decision will be based on these benchmarks and practical considerations.