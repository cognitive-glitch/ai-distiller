# AI Distiller - Rust Rewrite (In Progress)

> **Status**: Phase 1 - Foundation Complete ✅

This directory contains the Rust rewrite of AI Distiller.

## 🎯 Architecture

**Key Decision: NO tokio in core** - Using rayon for CPU parallelism only.

### Cargo Workspace Structure

```
crates/
├── aid-cli/            # Binary CLI (main entry point)
├── distiller-core/     # Core library (IR, processor, error handling)
├── lang-python/        # Python language processor (TODO)
├── lang-typescript/    # TypeScript processor (TODO)
└── formatter-*/        # Output formatters (TODO)
```

## 🚀 Quick Start

```bash
# Build debug version
cargo build -p aid-cli

# Run with verbosity
cargo run -p aid-cli -- testdata/python/01_basic/source.py -v

# Build optimized release
cargo build --release -p aid-cli

# Run tests
cargo test --all-features

# Check code quality
cargo clippy --all-features -- -D warnings
cargo fmt --all -- --check
```

## 📊 Progress: Phase 1 Complete

- ✅ Cargo workspace setup
- ✅ Error system (thiserror)
- ✅ ProcessOptions types
- ✅ IR type system (Node, Visitor pattern)
- ✅ Basic CLI with clap
- ✅ CI/CD pipeline (GitHub Actions)

### Next: Phase 2 - Core IR & Parser Infrastructure

## 🏗️ Design Principles

1. **Synchronous by default** - No async/await in core
2. **rayon for parallelism** - CPU-bound work, not I/O
3. **Zero unsafe** - Except tree-sitter FFI
4. **Feature-gated languages** - Modular compilation
5. **Native tree-sitter** - No CGO dependency

## 🔧 Performance Targets

- Single file: < 50ms
- Directory: 5000+ files/sec
- Binary size: < 25MB (vs 38MB Go)
- Memory: < 500MB for 10k files

## 📝 Development Notes

See `CLAUDE.md` in repository root for complete refactoring plan.
