# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Distiller (`aid`) is a high-performance CLI tool written in Rust that extracts essential code structure from large codebases for LLM consumption. It processes 13 programming languages via tree-sitter, producing ultra-compact output (90-98% reduction) while preserving semantic information.

**Core Purpose**: Enable AI assistants to understand entire codebases by fitting them into context windows, eliminating hallucinations caused by partial code visibility.

## Quick Start Commands

### Building

```bash
# Build debug version
cargo build -p aid-cli

# Build optimized release
cargo build --release -p aid-cli

# Run without installing
cargo run -p aid-cli -- testdata/python/01_basic/source.py -vv
```

### Testing

```bash
# Run all tests (309 tests across 23 crates)
cargo test --all-features

# Run tests for specific crate
cargo test -p lang-python --lib
cargo test -p distiller-core --test integration_tests

# Run tests with output for debugging
cargo test -- --nocapture

# Run specific test
cargo test test_python_class --  --nocapture
```

### Code Quality

```bash
# Check for clippy warnings (must pass in CI)
cargo clippy --all-features -- -D warnings

# Format code
cargo fmt --all

# Check formatting without modifying
cargo fmt --all -- --check
```

### Benchmarking

```bash
# Run benchmarks
cargo bench -p aid-cli
```

## Architecture

### Cargo Workspace Structure

```
crates/
├── aid-cli/              # CLI binary entry point
├── distiller-core/       # Core library (IR, processor, error, stripper)
│   ├── src/ir/          # Intermediate Representation nodes
│   ├── src/parser/      # Tree-sitter parser pooling
│   ├── src/processor/   # File and directory processing
│   └── src/stripper/    # Visitor-based filtering
├── lang-*/              # 13 language processors (Python, TypeScript, Go, etc.)
│   └── src/lib.rs       # Implements LanguageProcessor trait
├── formatter-*/         # 5 output formatters (text, markdown, JSON, JSONL, XML)
│   └── src/lib.rs       # Implements Formatter trait
└── mcp-server/          # Model Context Protocol server (optional)
```

### Data Flow

```
File Input → Language Processor (tree-sitter) → IR Generation →
Stripper (filtering) → Formatter → Output
```

### MCP Server

**Transport**: stdio (standard input/output)
**Status**: ✅ Production-ready (custom implementation) | 🔄 rmcp SDK migration planned

The MCP server provides 4 core operations:
- `distil_directory` - Process entire directory
- `distil_file` - Process single file
- `list_dir` - List directory contents with metadata
- `get_capa` - Get server capabilities

**Integration**: Claude Desktop, Cursor, VS Code via stdio transport
**Future**: Migration to official `rmcp` SDK planned for better standards compliance

### Core Abstractions

**LanguageProcessor Trait** (`distiller-core/src/processor/language.rs`):
```rust
pub trait LanguageProcessor: Send + Sync {
    fn language(&self) -> &'static str;
    fn supported_extensions(&self) -> &'static [&'static str];
    fn can_process(&self, path: &Path) -> bool;
    fn process(&self, source: &str, path: &Path, opts: &ProcessOptions) -> Result<File>;
}
```

**IR Node Types** (`distiller-core/src/ir/`):
- `Node` - Enum: File, Directory, Class, Function, Field, etc.
- `File` - Root node containing classes, functions, imports
- `Class/Struct` - Container for methods and fields
- `Function` - Methods/functions with parameters and return types
- `Field` - Class fields/properties
- `Import` - Import/use statements

**Stripper Pattern** (`distiller-core/src/stripper/`):
Visitor-based filtering system. Apply filtering via `Stripper::new(options)` - never implement custom filtering logic.

## Key Design Decisions

### 1. NO tokio in Core (CRITICAL)

- **Use rayon** for CPU parallelism (NOT tokio/async)
- **Rationale**: AI Distiller is CPU-bound; local filesystem is OS-buffered
- **Benefits**: Simpler code, smaller binaries, cleaner stack traces
- **Exception**: MCP server MAY use minimal tokio for JSON-RPC only

All processing is synchronous. Parallelism is achieved via rayon's thread pool.

### 2. Tree-sitter Integration

- Uses native Rust bindings (`tree-sitter` crate)
- Parser pooling for thread-safe access (`ParserPool`)
- All grammars compiled into binary

### 3. Visitor Pattern for Filtering

Standardized stripper system via the Visitor pattern:

```rust
// ✅ CORRECT: Use standardized stripper
let options = StripperOptions {
    remove_private: true,
    remove_implementations: true,
    remove_comments: true,
};
let stripper = Stripper::new(options);
let filtered_file = stripper.visit_file(&file);

// ❌ WRONG: Custom filtering logic
for node in file.children {
    if node.visibility == Visibility::Private {
        continue;  // DON'T DO THIS
    }
}
```

### 4. Error Handling

Uses `thiserror` for error types and `anyhow` for error context in CLI/applications.

```rust
// In libraries: use thiserror
#[derive(Debug, thiserror::Error)]
pub enum DistilError {
    #[error("Parse error: {0}")]
    ParseError(String),
}

// In binaries: use anyhow
fn main() -> anyhow::Result<()> {
    processor.process_path(path)
        .context("Failed to process path")?
}
```

## Language Processor Development

### Standard Pattern

All language processors follow this structure:

```rust
pub struct PythonProcessor {
    pool: Arc<ParserPool>,
}

impl PythonProcessor {
    pub fn new() -> Result<Self> {
        let pool = Arc::new(ParserPool::new(
            tree_sitter_python::LANGUAGE.into()
        )?);
        Ok(Self { pool })
    }
}

impl LanguageProcessor for PythonProcessor {
    fn language(&self) -> &'static str { "python" }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["py", "pyw"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .is_some_and(|ext| ext == "py" || ext == "pyw")
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let filename = path.to_string_lossy().into_owned();
        // 1. Parse source with tree-sitter via pool
        // 2. Walk AST and build IR nodes
        // 3. Return File node (stripper applied later by processor)
        self.parse_source(source, &filename)
    }
}
```

### Tree-sitter Safety

Always validate node positions:

```rust
fn node_text<'a>(&self, node: tree_sitter::Node, source: &'a [u8]) -> &'a str {
    let start = node.start_byte();
    let end = node.end_byte();

    // Validate boundaries
    if start > end || end > source.len() {
        return "";
    }

    std::str::from_utf8(&source[start..end]).unwrap_or("")
}
```

## Testing

### Test Organization

```
testdata/
├── python/
│   ├── 01_basic/
│   │   ├── source.py              # Input file
│   │   └── expected/
│   │       ├── default.txt        # Public APIs only
│   │       ├── implementation=1.txt
│   │       └── private=1,protected=1,internal=1.txt
│   ├── 02_simple/
│   └── ...
├── typescript/
├── go/
└── ...
```

**Expected File Naming**: Files in `expected/` reflect CLI parameters used:
- `default.txt` - Default output (public, no implementation)
- `implementation=1.txt` - Includes method bodies
- `private=1,protected=1,internal=1.txt` - All visibility levels

### Running Tests

```bash
# All tests
cargo test --all-features

# Specific language
cargo test -p lang-python

# Integration tests
cargo test -p distiller-core --test integration_tests

# Single test with output
cargo test test_python_class -- --nocapture --test-threads=1
```

### Adding Tests

1. Create source file in `testdata/<lang>/<scenario>/source.<ext>`
2. Generate expected output:
   ```bash
   cargo run -p aid-cli -- testdata/python/01_basic/source.py --stdout > testdata/python/01_basic/expected/default.txt
   ```
3. Add test case in language processor's `lib.rs`

## Development Workflow

### Adding a New Language Processor

1. **Create crate structure**:
   ```bash
   cargo new --lib crates/lang-newlang
   ```

2. **Update workspace** in root `Cargo.toml`:
   ```toml
   members = [
       # ... existing members
       "crates/lang-newlang",
   ]
   ```

3. **Add dependencies** in `crates/lang-newlang/Cargo.toml`:
   ```toml
   [dependencies]
   distiller-core = { path = "../distiller-core" }
   tree-sitter = { workspace = true }
   tree-sitter-newlang = "x.y.z"
   # ... other deps
   ```

4. **Implement LanguageProcessor trait**

5. **Register in CLI** (`crates/aid-cli/src/main.rs`):
   ```rust
   processor.register_language(Box::new(
       NewLangProcessor::new()?
   ));
   ```

6. **Add test cases** in `testdata/newlang/`

### Adding a New Formatter

1. Create crate: `cargo new --lib crates/formatter-newformat`
2. Implement formatter trait
3. Register in CLI
4. Add tests

## Common Pitfalls

### ❌ DON'T

- Use tokio/async in core or language processors
- Implement custom filtering logic (use `Stripper`)
- Skip tree-sitter node boundary validation
- Use line-based regex parsing instead of AST traversal
- Create mocks in tests (use real tree-sitter parsing)

### ✅ DO

- Use rayon for parallelism
- Use standardized `Stripper` for filtering
- Validate all tree-sitter node positions
- Use AST-based traversal
- Test against real parser output

## Git Commit Style

```
feat(parser): add support for async/await in TypeScript
fix(go): resolve method association for embedded structs
chore: update expected test files
refactor(core): improve error handling in stripper
perf(python): optimize AST traversal performance
```

## Important Files

- `Cargo.toml` - Workspace configuration
- `RUST_PROGRESS.md` - Implementation progress tracking
- `README.md` - User-facing documentation
- `testdata/` - Integration test fixtures
- `docs/lang/*.md` - Language-specific parser documentation

## Notes for AI Assistants

- All code must use Rust 2024 edition (Rust 1.90.0+)
- Tests must pass before any PR: `cargo test --all-features`
- Clippy must pass: `cargo clippy --all-features -- -D warnings`
- Code must be formatted: `cargo fmt --all`
- No mocks or fake implementations - test against real tree-sitter parsing
- Update testdata expected files when parser behavior changes
- Language processors should be self-contained (minimal dependencies)
- Keep CLI focused - complex logic belongs in libraries
