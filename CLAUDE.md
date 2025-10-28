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

# Run with stdin for quick testing
echo 'class User: pass' | cargo run -p aid-cli -- -
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
cargo test test_python_class -- --nocapture

# Run single test with single thread (for debugging)
cargo test test_python_class -- --nocapture --test-threads=1
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

### Quick Testing with stdin

For rapid development testing, `aid` supports stdin input with automatic language detection:

```bash
# Auto-detect language from code patterns
echo 'class User { getName() { return this.name; } }' | cargo run -p aid-cli -- - --format text

# Explicit language specification
cat snippet.ts | cargo run -p aid-cli -- - --lang typescript --implementation=1

# Test parser with full debug trace
echo 'def foo(): pass' | cargo run -p aid-cli -- - -vvv

# Test from file
cargo run -p aid-cli -- - --lang python < test.py
```

Automatic language detection works for: python, typescript, javascript, go, ruby, swift, rust, java, c#, kotlin, c++, php.

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

The MCP server provides 4 core operations:
- `distill_directory` - Process entire directory
- `distill_file` - Process single file
- `list_dir` - List directory contents with metadata
- `get_capabilities` - Get server capabilities

Plus 10 specialized AI analysis tools that generate prompts:
- `aid_hunt_bugs` - Bug hunting prompts
- `aid_suggest_refactoring` - Refactoring analysis prompts
- `aid_generate_diagram` - Mermaid diagram prompts
- `aid_analyze_security` - OWASP Top 10 security audit prompts
- And 6 more specialized analysis prompts

**Integration**: Claude Desktop, Cursor, VS Code, Codex via stdio transport

**NPM Package**: `@cognitive/ai-distiller-mcp` - Installable via `npx`

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

**CRITICAL**: This trait is SYNCHRONOUS (no async/await). Use rayon for parallelism at the processor level.

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
- All grammars compiled into binary - zero external dependencies

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

## Important CLI Features

### AI Actions and Prompt Generation

AI Distiller generates specialized analysis prompts combined with distilled code. These are NOT executed by `aid` - they create prompts for AI agents to execute.

**Available Actions** (via `--ai-action` flag):
- `flow-for-deep-file-to-file-analysis` - Systematic file-by-file analysis task list
- `flow-for-multi-file-docs` - Multi-file documentation workflow
- `prompt-for-security-analysis` - OWASP Top 10 security audit prompt
- `prompt-for-refactoring-suggestion` - Refactoring suggestions prompt
- `prompt-for-complex-codebase-analysis` - Enterprise-grade analysis with diagrams
- `prompt-for-performance-analysis` - Performance optimization prompt
- `prompt-for-best-practices-analysis` - Code quality assessment prompt
- `prompt-for-bug-hunting` - Bug detection and pattern analysis prompt
- `prompt-for-single-file-docs` - Single file documentation prompt
- `prompt-for-diagrams` - Generate 10+ Mermaid architecture diagrams

**Usage**:
```bash
cargo run -p aid-cli -- src/ --ai-action=prompt-for-security-analysis --private=1
```

Output is saved to `.aid/` directory with pattern: `.aid/<action>.<timestamp>.<dirname>.md`

### .aidignore System

Uses `.gitignore` syntax to exclude files from processing.

**Automatically ignored directories**:
- `node_modules/`, `vendor/`, `target/`, `build/`, `dist/`
- `__pycache__/`, `.venv/`, `.pytest_cache/`, `venv/`, `env/`
- `.gradle/`, `Pods/`, `.bundle/`
- `.vs/`, `.idea/`, `.vscode/`
- `.git/`, `.svn/`, `.hg/`

**Special feature**: Use `!` prefix to include normally-ignored content:
```bash
# .aidignore
!vendor/my-local-package/  # Include specific vendor package
!*.md                       # Include markdown files
```

Place `.aidignore` files in any directory for nested control.

### Git History Analysis Mode

Special mode activated when path is `.git`:

```bash
cargo run -p aid-cli -- .git --git-limit=500 --with-analysis-prompt
```

Generates formatted commit history with optional AI analysis prompt for:
- Contributor statistics and expertise areas
- Timeline analysis and development phases
- Functional categorization (features, fixes, refactoring)
- Codebase evolution insights
- Actionable recommendations

Output includes both the prompt and formatted git history.

### Dependency-Aware Distillation

**Status**: Experimental feature for call graph analysis

Traces function calls across files to include only used code:

```bash
cargo run -p aid-cli -- main.py --dependency-aware --max-depth=2 --implementation=1
```

**Language Support Quality**:
- 🟢 **Very Good** (production-ready): Python, JavaScript, Go, Rust, Java, Swift, PHP, Ruby
- 🟡 **Limited** (basic functionality): TypeScript, C#, C++ (language processor limitations)

### Summary Output Types

After processing, `aid` displays summary with compression statistics.

**Available formats** (via `--summary-type`):
- `visual-progress-bar` (default) - Progress bar with green/red dots
- `stock-ticker` - Compact stock market style
- `speedometer-dashboard` - Multi-line dashboard
- `minimalist-sparkline` - Single line with sparkline
- `ci-friendly` - Clean format for CI/CD
- `json` - Machine-readable JSON
- `off` - Disable summary

Use `--no-emoji` to remove emojis from any format.

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
- Use stdin (`-`) for quick testing during development
- Use `-vvv` for full trace including IR dumps

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
- `.aidignore` - File exclusion patterns (user-created, gitignore syntax)
- `.aid/` - Default output directory (auto-generated, add to .gitignore)
- `benchmark/` - Performance benchmark scripts and results

## Notes for AI Assistants

- All code must use Rust 2024 edition (Rust 1.90.0+)
- Tests must pass before any PR: `cargo test --all-features`
- Clippy must pass: `cargo clippy --all-features -- -D warnings`
- Code must be formatted: `cargo fmt --all`
- No mocks or fake implementations - test against real tree-sitter parsing
- Update testdata expected files when parser behavior changes
- Language processors should be self-contained (minimal dependencies)
- Keep CLI focused - complex logic belongs in libraries
- Use stdin for quick iterations: `echo 'code' | cargo run -p aid-cli -- -`
- AI actions generate prompts, they don't perform analysis
- Check `.aidignore` when files aren't being processed as expected
- Git mode: path `.git` activates special commit history analysis
- The LanguageProcessor trait is SYNCHRONOUS - never use async/await in processors
