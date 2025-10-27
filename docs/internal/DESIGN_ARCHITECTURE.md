# AI Distiller - System Architecture & Design

## 1. Overview

AI Distiller is a high-performance command-line tool that distills source code into compact, structured formats optimized for LLM context windows. This document defines the complete system architecture based on our PRD and technical decisions.

## 2. Core Design Decisions

### 2.1 Technology Stack
- **Language**: Go 1.21+
- **Parsing**: Tree-sitter via WASM (Wazero runtime)
- **CLI Framework**: Cobra
- **Testing**: Standard Go testing + testify
- **Build**: Make + GitHub Actions

### 2.2 Architectural Principles
1. **Modularity**: Clear separation of concerns
2. **Immutability**: IR nodes are never modified after creation
3. **Performance**: Concurrent pipeline processing
4. **Extensibility**: Easy addition of new languages and formats
5. **Testability**: Comprehensive unit and integration tests

## 3. Directory Structure

```
/ai-distiller/
├── cmd/
│   └── aid/
│       └── main.go                 # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                # Root command setup
│   │   ├── flags.go               # Flag definitions
│   │   └── validation.go          # Input validation
│   ├── distiller/
│   │   ├── distiller.go           # Core orchestration
│   │   ├── pipeline.go            # Concurrent pipeline
│   │   ├── file_discovery.go      # File traversal
│   │   └── distiller_test.go
│   ├── ir/
│   │   ├── types.go               # IR type definitions
│   │   ├── visitor.go             # Visitor interface
│   │   ├── builder.go             # IR construction helpers
│   │   └── types_test.go
│   ├── parser/
│   │   ├── interface.go           # LanguageProcessor interface
│   │   ├── registry.go            # Language registration
│   │   ├── wasm_runtime.go        # WASM runtime management
│   │   └── parser_test.go
│   ├── language/
│   │   ├── python/
│   │   │   ├── processor.go       # Python implementation
│   │   │   ├── processor_test.go
│   │   │   └── python.wasm        # Embedded WASM
│   │   ├── go/
│   │   │   ├── processor.go
│   │   │   └── go.wasm
│   │   └── javascript/
│   │       ├── processor.go
│   │       └── javascript.wasm
│   ├── stripper/
│   │   ├── visitor.go             # Stripping visitor
│   │   ├── rules.go               # Stripping rules
│   │   └── stripper_test.go
│   ├── formatter/
│   │   ├── interface.go           # Formatter interface
│   │   ├── markdown.go            # Markdown formatter
│   │   ├── jsonl.go               # JSONL formatter
│   │   ├── json_structured.go    # Structured JSON
│   │   ├── xml.go                 # XML formatter
│   │   └── formatter_test.go
│   └── config/
│       ├── config.go              # Configuration management
│       └── languages.go           # Language definitions
├── pkg/
│   └── fileutil/
│       ├── glob.go                # Glob pattern matching
│       └── symlink.go             # Symlink handling
├── assets/
│   └── wasm/                      # Pre-compiled WASM modules
├── testdata/
│   ├── projects/                  # Sample projects
│   └── fixtures/                  # Test fixtures
├── docs/
│   ├── README.md
│   ├── ARCHITECTURE.md
│   └── CONTRIBUTING.md
├── scripts/
│   ├── build-wasm.sh             # Build WASM modules
│   └── install.sh                # Installation script
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── Makefile
├── go.mod
└── go.sum
```

## 4. Component Architecture

### 4.1 Data Flow

```
User Input (CLI)
    ↓
CLI Parser & Validation
    ↓
File Discovery
    ↓
Concurrent Pipeline [
    Parser Workers (WASM) → IR Generation
]
    ↓
IR Assembly & Symbol Resolution
    ↓
Stripping Visitor (Single Pass)
    ↓
Output Formatter
    ↓
File/Stdout Output
```

### 4.2 Core Components

#### CLI Layer (`internal/cli/`)
- Parses command-line arguments using Cobra
- Validates input paths and options
- Configures the distiller pipeline

#### Distiller (`internal/distiller/`)
- Orchestrates the entire processing pipeline
- Manages concurrent workers
- Handles file discovery and filtering
- Ensures deterministic output ordering

#### IR System (`internal/ir/`)
- Defines the intermediate representation types
- Implements the visitor pattern
- Provides builder utilities for IR construction
- Maintains immutability guarantees

#### Parser System (`internal/parser/`)
- Manages WASM runtime lifecycle
- Provides language processor registry
- Handles parser dispatch based on file extension
- Abstracts WASM complexity from language processors

#### Language Processors (`internal/language/`)
- Implement language-specific parsing logic
- Convert tree-sitter AST to IR
- Handle language-specific visibility rules
- Embed pre-compiled WASM modules

#### Stripping System (`internal/stripper/`)
- Implements the single-pass visitor
- Applies stripping rules based on options
- Maintains IR immutability (creates new nodes)
- Handles language-specific stripping logic

#### Output Formatters (`internal/formatter/`)
- Convert IR to various output formats
- Handle streaming output
- Manage formatting options
- Ensure consistent output across formats

## 5. Key Interfaces

### 5.1 Language Processor

```go
type LanguageProcessor interface {
    // Language returns the language identifier
    Language() string

    // FileExtensions returns supported extensions
    FileExtensions() []string

    // Parse converts source code to IR
    Parse(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error)

    // WASMModule returns the embedded WASM module
    WASMModule() []byte
}
```

### 5.2 Formatter

```go
type Formatter interface {
    // Format converts IR to output format
    Format(files []*ir.DistilledFile, options FormatterOptions) error

    // FormatStream formats files as they arrive
    FormatStream(ctx context.Context, files <-chan *ir.DistilledFile, options FormatterOptions) error
}
```

### 5.3 Stripping Visitor

```go
type StrippingVisitor struct {
    Options  StrippingOptions
    Language string
    Rules    []StrippingRule
}

func (v *StrippingVisitor) Visit(node ir.DistilledNode) ir.DistilledNode {
    // Single-pass transformation logic
}
```

## 6. Concurrency Model

### 6.1 Pipeline Stages

1. **File Discovery** (1 goroutine)
   - Walks directories
   - Applies include/exclude patterns
   - Sends file paths to channel

2. **Parser Workers** (N goroutines, N = NumCPU)
   - Read files from channel
   - Parse using WASM runtime
   - Generate IR
   - Send to results channel

3. **Stripping** (1 goroutine)
   - Applies visitor pattern
   - Maintains order for deterministic output
   - Sends to formatter channel

4. **Output Formatting** (1 goroutine)
   - Formats IR based on selected format
   - Writes to file or stdout
   - Handles streaming

### 6.2 Channel Design

```go
type Pipeline struct {
    Files    chan FileInfo      // Discovered files
    Parsed   chan ParseResult   // Parsed IR + metadata
    Stripped chan *ir.DistilledFile
    Errors   chan error
}
```

## 7. WASM Integration

### 7.1 Runtime Management
- Single Wazero runtime instance per parser worker
- Lazy initialization of language modules
- Proper cleanup on shutdown

### 7.2 Memory Management
- Pre-allocated WASM memory buffers
- Efficient data transfer between Go and WASM
- Bounded memory usage per file

### 7.3 Error Handling
- Graceful handling of WASM panics
- Timeout protection for parsing
- Fallback to partial results on error

## 8. Configuration

### 8.1 Language Registry

```go
func init() {
    parser.RegisterLanguage(&python.Processor{})
    parser.RegisterLanguage(&golang.Processor{})
    parser.RegisterLanguage(&javascript.Processor{})
    // ... other languages
}
```

### 8.2 Default Options

```go
type Config struct {
    DefaultStrip    []string
    DefaultExclude  []string
    MaxFileSize     int64
    ParseTimeout    time.Duration
    WorkerCount     int
}
```

## 9. Error Handling Strategy

1. **File-Level Errors**: Log and continue
2. **Parse Errors**: Include in IR as error nodes
3. **Fatal Errors**: Clean shutdown with error message
4. **Panics**: Recover in workers, report as errors

## 10. Testing Strategy

### 10.1 Unit Tests
- Each component tested in isolation
- Mock interfaces for dependencies
- Table-driven tests for formatters
- Property-based tests for IR transformations

### 10.2 Integration Tests
- End-to-end CLI tests
- Real project parsing tests
- Cross-language consistency tests
- Performance regression tests

### 10.3 Benchmarks
- Parser performance per language
- Stripping visitor performance
- Memory usage profiling
- Startup time validation

## 11. Build & Release

### 11.1 WASM Module Building
- Automated via `scripts/build-wasm.sh`
- Uses emscripten in Docker
- Validates module size and exports
- Embeds in Go code via `go:embed`

### 11.2 Binary Building
- Single `make build` command
- Cross-compilation via standard Go
- Binary signing for releases
- Automated via GitHub Actions

## 12. Future Extensibility

### 12.1 Adding Languages
1. Create new processor in `internal/language/<lang>/`
2. Implement LanguageProcessor interface
3. Build WASM module from tree-sitter grammar
4. Register in language registry
5. Add tests

### 12.2 Adding Output Formats
1. Create new formatter in `internal/formatter/`
2. Implement Formatter interface
3. Register in formatter registry
4. Add tests

### 12.3 Custom Stripping Rules
1. Implement StrippingRule interface
2. Register rule in rule engine
3. Map to CLI flag if needed

## 13. Performance Optimizations

1. **Worker Pool**: Reuse WASM runtime instances
2. **Memory Pool**: Pre-allocate buffers
3. **Streaming**: Process files as discovered
4. **Lazy IR**: Build IR nodes on demand
5. **Parallel I/O**: Overlap file reading with parsing

## 14. Security Considerations

1. **Path Traversal**: Validate all file paths
2. **Symlink Loops**: Detect and prevent
3. **Resource Limits**: Bounded memory and CPU
4. **WASM Sandbox**: Parser isolation
5. **No Network**: Fully offline operation

## 15. Monitoring & Diagnostics

1. **Verbose Logging**: Multiple levels (-v, -vv, -vvv)
2. **Progress Reporting**: File count and timing
3. **Memory Stats**: Peak usage reporting
4. **Error Summary**: Aggregated error report

---

This architecture provides a solid foundation for implementing AI Distiller with clear separation of concerns, excellent testability, and room for future growth.