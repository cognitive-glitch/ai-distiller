# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Distiller (`aid`) is a high-performance CLI tool that extracts essential code structure from large codebases for LLM consumption. It processes 12+ programming languages via tree-sitter, producing ultra-compact output (90-98% reduction) while preserving semantic information.

**Core Purpose**: Enable AI assistants to understand entire codebases by fitting them into context windows, eliminating hallucinations caused by partial code visibility.

---

## 🦀 Rust Refactoring (Target Architecture)

### Strategic Goals

AI Distiller is being rewritten in Rust to achieve:

1. **Performance**: 2-3x faster processing (no CGO overhead)
2. **Safety**: Memory safety and fearless concurrency
3. **Simplicity**: Smaller binaries, cleaner architecture
4. **MCP Simplification**: Reduce from 10+ functions to 4 core operations

### Target Architecture

**Cargo Workspace Monorepo**:
```
crates/
├── aid-cli/            # Binary CLI entry point
├── distiller-core/     # Core library (IR, processor, error)
├── lang-python/        # Python language processor
├── lang-typescript/    # TypeScript processor
├── lang-*/             # Other language processors (12+ total)
├── formatter-text/     # Text output formatter
├── formatter-*/        # Other formatters (markdown, JSON, XML)
└── mcp-server/         # Simplified MCP server (4 functions)
```

### Key Design Decisions

#### 1. NO tokio in Core (CRITICAL)
- **Use rayon** for CPU parallelism (NOT tokio/async)
- **Rationale**: AI Distiller has zero network I/O; local filesystem is OS-buffered
- **Benefits**: Simpler code, smaller binaries, cleaner stack traces
- **Exception**: MCP server MAY use minimal tokio for JSON-RPC only

#### 2. Simplified MCP Interface
Reduce from 10+ functions to **4 core operations**:
1. `distil_directory(path, options)` - Process directory
2. `distil_file(path, options)` - Process single file
3. `list_dir(path, filters)` - List directory with metadata
4. `get_capa()` - Get server capabilities

#### 3. Feature-Gated Languages
- Each language is optional cargo crate
- Modular compilation for smaller binaries
- Example: `cargo build --features lang-python,lang-typescript`

#### 4. IR Design
- Enum-based nodes for zero-cost dispatch
- Visitor pattern for extensible operations
- Synchronous trait interfaces (no async)

### Target Performance

| Metric | Go Baseline | Rust Target |
|--------|-------------|-------------|
| Django (970 files) | 231ms | < 150ms |
| Binary size | 38MB | < 25MB |
| Memory (10k files) | ~800MB | < 500MB |
| Startup time | ~80ms | < 50ms |

### Development Status

**See `RUST_PROGRESS.md` for implementation progress.**

**Current Branch**: `clever-river` (Rust rewrite in progress)

---

## Go Implementation (Current)

### Essential Development Commands

#### Building

```bash
# Standard build (CGO required for full language support)
make build                 # Full build with all parsers (~38MB)
make build-verbose         # Shows CGO warnings during build
make aid ARGS="test.py"    # Quick build + run (development)
```

#### Testing

```bash
# Primary test commands
make test                  # Default: enhanced output with gotestsum
make test-pretty          # Package summary with ✓/✖ indicators
make test-dots            # Dot progress (good for large suites)
make test-watch           # Auto-rerun on file changes

# Specialized tests
make test-integration     # Integration tests via testrunner
make test-update          # Update expected test files
make test-regenerate      # Regenerate all expected outputs
make generate-expected-testdata  # Build aid + regenerate all

# Legacy/basic
make test-basic           # Standard Go test output (no gotestsum)
```

#### Development Utilities

```bash
make dev-init             # Initialize dev environment (install tools)
make clean                # Remove build artifacts
make lint                 # Run golangci-lint
make fmt                  # Format code with gofmt
```

---

## Architecture (Go)

### Core Data Flow

```
CLI Input → File Discovery → Language Detection → Parser (tree-sitter)
→ IR Generation → Semantic Analysis → Stripper → Formatter → Output
```

### Module Structure

```
cmd/
├── aid/                   # Main CLI entry point
├── parser-test/          # Parser validation tool
├── performance-test/     # Performance benchmarking
└── semantic-test/        # Semantic analysis testing

internal/
├── cli/                  # Command-line interface & flag handling
├── processor/            # Core file processing orchestration
├── parser/               # Tree-sitter WASM runtime & language parsers
├── language/             # Language-specific processors (Python, Go, TS, etc.)
├── ir/                   # Intermediate Representation (IR) node types
├── semantic/             # Semantic analysis (dependency graphs, call tracking)
├── stripper/             # Visitor pattern for filtering IR nodes
├── formatter/            # Output formatters (text, markdown, JSON, XML)
├── ai/                   # AI integration helpers
├── aiactions/            # AI action handlers (security, refactoring, etc.)
├── ignore/               # .aidignore file handling
├── importfilter/         # Import statement filtering
├── summary/              # Summary generation & token counting
├── performance/          # Performance tracking & metrics
├── project/              # Project root detection
├── debug/                # Debug system (-v, -vv, -vvv)
└── version/              # Version information
```

### Key Interfaces

**LanguageProcessor** (`internal/language/`):
```go
type LanguageProcessor interface {
    Language() string
    SupportedExtensions() []string
    CanProcess(filename string) bool
    ProcessWithOptions(ctx, reader, filename, ProcessOptions) (*ir.DistilledFile, error)
}
```

**IR Node Types** (`internal/ir/`):
- `DistilledFile` - Root node
- `DistilledClass` - Classes/structs
- `DistilledFunction` - Methods/functions
- `DistilledField` - Properties/fields
- `DistilledImport` - Import statements

**Visitor Pattern** (`internal/stripper/`):
Standardized filtering via `stripper.New()` - NEVER implement custom filtering.

---

## Language Parser Development (Go)

### Architecture Pattern (CRITICAL)

All language parsers follow this proven two-stage pattern:

```go
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader,
    filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {

    // Stage 1: Parse with tree-sitter
    file, err := p.treeparser.ProcessSource(ctx, source, filename)
    if err != nil {
        return nil, err
    }

    // Stage 2: Apply standardized stripper (NEVER custom filtering)
    stripperOpts := stripper.Options{
        RemovePrivate:         !opts.IncludePrivate,
        RemoveImplementations: !opts.IncludeImplementation,
        RemoveComments:        !opts.IncludeComments,
        RemoveImports:         !opts.IncludeImports,
    }

    s := stripper.New(stripperOpts)
    stripped := file.Accept(s)
    return stripped.(*ir.DistilledFile), nil
}
```

### Language-Specific Patterns

**Method Association**:
- **Go**: Two-pass (collect types → associate methods)
- **TypeScript/JavaScript**: Single-pass within class body
- **Python**: Native tree-sitter nesting

**Visibility Detection**:
- **Go**: Uppercase = public, lowercase = package-private
- **TypeScript/JavaScript**: Keywords + `#private` fields + JSDoc
- **Python**: `_private`, `__dunder__` conventions

**Tree-sitter Safety** (CRITICAL):
```go
func (p *Parser) nodeText(node *sitter.Node) string {
    if node == nil {
        return ""
    }
    start := node.StartByte()
    end := node.EndByte()
    sourceLen := uint32(len(p.source))
    if start > end || end > sourceLen {
        return ""
    }
    return string(p.source[start:end])
}
```

### Common Pitfalls

❌ **NEVER**: Use custom `applyOptions` filtering
✅ **ALWAYS**: Use standardized `stripper.New()`

❌ **NEVER**: Use line-based regex parsing
✅ **ALWAYS**: Use tree-sitter AST traversal

❌ **NEVER**: Skip boundary checks on node text extraction
✅ **ALWAYS**: Validate start/end byte positions

---

## Testing Strategy (Go)

### Test File Organization

```
testdata/
├── python/
│   ├── 01_basic/
│   │   ├── source.py
│   │   ├── default.txt                    # Public APIs only
│   │   ├── implementation=1.txt           # With implementations
│   │   └── private=1,protected=1,internal=1,implementation=0.txt
│   ├── 02_simple/
│   └── ...
├── typescript/
├── go/
└── ...
```

**Naming Convention**: Expected files reflect non-default CLI parameters:
- `default.txt` - Public APIs, no implementation (default behavior)
- `implementation=1.txt` - Includes method bodies
- `private=1,protected=1,internal=1,implementation=0.txt` - All visibility levels

### Integration Tests

Located in `internal/testrunner/`:
- Uses `testdata/` directory structure
- Validates parser output against expected files
- Run with `make test-integration`

### Updating Expected Files

```bash
# Update all expected files after parser changes
make test-regenerate

# Or update during test run
make test-update

# Generate for specific language after changes
./build/aid testdata/python/01_basic/source.py --stdout > testdata/python/01_basic/default.txt
```

---

## Debugging (Go)

### Debug Levels

```bash
aid src/ -v        # Level 1: Basic info (file counts, phases)
aid src/ -vv       # Level 2: Detailed (timing, AST node counts)
aid src/ -vvv      # Level 3: Full trace (IR dumps, raw structures)
```

**Implementation**:
- Uses `debug.FromContext(ctx)` for propagation
- Subsystem scoping: `dbg.WithSubsystem("parser")`
- Performance guards: `if dbg.IsEnabledFor(debug.LevelDetailed)`
- Output: stderr with format `[HH:MM:SS.mmm] [subsystem] LEVEL: message`

---

## Critical Development Principles

### NO MOCKS OR SIMULATED DATA

**NEVER**:
- Create mock implementations returning fixed data
- Use hardcoded test data pretending to be parser output
- Write placeholder functions that don't work
- Test mocked behavior instead of real functionality

**ALWAYS**:
- Implement real, working parsers using tree-sitter
- Test against actual parser output
- Use `testdata/` files for validation

### Stripper Integration

The `internal/stripper/` package provides standardized filtering via the Visitor pattern:

```go
// ✅ CORRECT: Use standardized stripper
s := stripper.New(stripper.Options{
    RemovePrivate:         true,
    RemoveImplementations: true,
    RemoveComments:        true,
})
stripped := file.Accept(s)

// ❌ WRONG: Custom filtering logic
for _, node := range file.Children {
    if node.Visibility == "private" {
        continue  // DON'T DO THIS
    }
}
```

### Performance Requirements

- **Speed**: Process 10MB codebases in <2 seconds
- **Concurrency**: Default 80% CPU cores (`--workers=0`)
- **Memory**: Stream processing, bounded channels
- **One-pass**: No multiple IR traversals

---

## Git Commit Protocol

**Pre-commit Checklist**:
1. Run `git status` - verify no unwanted files
2. Check for temporary files: `*.tmp`, `*.log`, `.aid.*.txt` in root
3. Review with `git diff --cached`
4. Ensure `.gitignore` is properly configured
5. Run `make test` (Go) or `cargo test` (Rust) - all tests must pass

**Commit Style**:
```
feat(parser): add support for async/await in TypeScript
fix(go): resolve method association for embedded structs
chore: update expected test files for Python parser
feat(rust): Phase 1 Foundation - Cargo workspace, IR, error system
```

---

## Important Files & Documentation

- `BUILD.md` - Cross-compilation setup and build requirements (Go)
- `RUST_PROGRESS.md` - Rust refactoring implementation progress
- `README.rust.md` - Rust implementation quick start
- `docs/TESTING.md` - Comprehensive testing guide with gotestsum formats (Go)
- `docs/CROSS_COMPILATION.md` - Detailed cross-compilation instructions (Go)
- `docs/lang/*.md` - Language-specific parser documentation (Go)
- `.aidignore` - File exclusion patterns (gitignore syntax)

---

## Quick Reference

### Common Tasks (Go)

```bash
# Add new language parser
1. Create internal/language/<lang>/processor.go
2. Implement LanguageProcessor interface
3. Use tree-sitter parser (see internal/parser/grammars/)
4. Register in internal/language/registry.go
5. Add testdata/<lang>/ with test cases
6. Run make generate-expected-testdata

# Fix failing test
1. Reproduce: add failing test case to testdata/
2. Debug: aid testdata/<lang>/file.ext -vvv
3. Fix parser in internal/language/<lang>/
4. Verify: make test-integration
5. Update: make test-regenerate

# Add output format
1. Create internal/formatter/<format>.go
2. Implement Formatter interface
3. Register in formatter registry
4. Test with --format <format>
```

### Common Tasks (Rust)

```bash
# Build and test
cargo build --release -p aid-cli
cargo test --all-features
cargo clippy --all-features -- -D warnings

# Run
cargo run -p aid-cli -- --help
cargo run -p aid-cli -- testdata/python/01_basic/source.py -vvv

# Add language processor
1. Create crates/lang-<lang>/
2. Implement LanguageProcessor trait
3. Add to workspace members in Cargo.toml
4. Write tests with insta snapshots
```

### Performance Debugging (Go)

```bash
# Profile memory usage
go test -memprofile=mem.prof -run=TestProcessor
go tool pprof mem.prof

# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=BenchmarkProcessor
go tool pprof cpu.prof

# Benchmark specific functionality
make bench
```

---

## Notes for AI Assistants

- **Language**: Use English for all communication, code, and documentation
- **CLI Examples**: Already comprehensive in README.md - don't duplicate
- **Focus**: Architecture understanding and development workflow
- **Testing**: Always run tests after changes; update expected files if parser behavior changes
- **Parallelism**: Use `make aid` (Go) or `cargo run -p aid-cli` (Rust) for rapid iteration
- **Progress**: See `RUST_PROGRESS.md` for Rust implementation status
