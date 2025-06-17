# Product Requirements Document: AI Distiller

## 1. Executive Summary

AI Distiller is a high-performance command-line tool that solves a critical problem in modern AI-assisted software development: the inability of Large Language Models to comprehend large codebases due to context window limitations. By intelligently extracting and condensing the essential structure, APIs, and relationships from source code, AI Distiller creates compact, semantic "blueprints" that enable LLMs to reason effectively about complex software projects.

Built as a single, dependency-free native binary for all major platforms, AI Distiller uses advanced Abstract Syntax Tree (AST) parsing to support 15+ programming languages with consistent, high-quality results. The tool integrates seamlessly into developer workflows through an intuitive CLI interface and produces output optimized for both human readability and machine consumption.

## 2. Business Case & Market Analysis

### 2.1 Problem Statement

Modern software projects often contain thousands of files and millions of lines of code. When developers seek AI assistance for:
- Understanding unfamiliar codebases
- Architectural reviews
- API documentation generation
- Code refactoring suggestions
- Security analysis

They face a fundamental limitation: even the most advanced LLMs have context windows that cannot accommodate entire codebases. Current solutions involve manual copying of relevant files or using simplistic concatenation tools that produce verbose, unstructured output.

### 2.2 Market Opportunity

**Target Market Size:**
- 26.9 million professional developers worldwide (2025)
- 71% use AI coding assistants regularly
- Growing at 3-5% annually

**Key Market Segments:**
1. **Individual Developers** - Using AI assistants for daily coding tasks
2. **Development Teams** - Onboarding, documentation, code reviews
3. **AI/ML Engineers** - Building tools and training datasets
4. **Technical Writers** - Creating and maintaining documentation
5. **Consultants** - Analyzing client codebases

### 2.3 Competitive Landscape

Current alternatives are inadequate:
- **Manual Selection**: Time-consuming, error-prone, misses context
- **Simple Concatenation Tools**: No semantic understanding, bloated output
- **IDE Plugins**: Platform-specific, limited language support
- **Commercial Solutions**: Expensive, cloud-dependent, privacy concerns

AI Distiller's unique value proposition:
- Single binary, works everywhere
- Semantic understanding via AST parsing
- Privacy-first (fully local processing)
- Extensive language support
- Developer-friendly CLI

## 3. Product Goals & Success Metrics

### 3.1 Primary Goals

1. **Enable Effective AI-Assisted Development**: Allow developers to leverage AI for large codebases
2. **Maximize Information Density**: Fit maximum semantic information in minimum tokens
3. **Ensure Universal Accessibility**: Work on any platform without dependencies
4. **Maintain Developer Trust**: Open source, local processing, no telemetry

### 3.2 Success Metrics

**Adoption Metrics:**
- 10,000+ GitHub stars within first year
- 50,000+ monthly active users by year 2
- Integration into 5+ major AI coding tools

**Quality Metrics:**
- 95%+ accuracy in preserving semantic meaning
- <2 second processing time for 10MB codebases
- Zero runtime dependencies
- <50MB binary size

**Community Metrics:**
- 100+ contributors
- Support for 20+ languages
- 1,000+ GitHub issues resolved

## 4. Functional Requirements

### 4.1 Core Features

**F1: Intelligent Code Distillation**
- Extract code structure using AST parsing
- Preserve semantic meaning while reducing size
- Support for selective extraction (public APIs, signatures, etc.)

**F2: Multi-Language Support**
- Consistent interface across all languages
- Language auto-detection via file extensions
- Extensible architecture for new languages

**F3: Flexible Output Generation**
- Multiple output formats (Markdown, JSONL, XML)
- Configurable output paths
- Streaming to stdout option

**F4: Granular Control**
- Strip specific elements (comments, imports, implementation)
- Include/exclude file patterns
- Recursive directory processing

### 4.2 CLI Specification

```bash
aid [path] [flags]

ARGUMENTS:
  [path]   Path to source directory or file (default: current directory)

FLAGS:
  -o, --output <file>      Output file path (default: .aid.<dir>.[options].txt)
      --stdout             Print to stdout (in addition to file)
      --format <fmt>       Output format: md|jsonl|json-structured|xml (default: md)
      --strip <items>      Remove items: comments,imports,implementation,non-public
                          (comma-separated)
      --include <glob>     Include file patterns (default: all supported)
      --exclude <glob>     Exclude file patterns
  -r, --recursive          Process directories recursively (default: true)
      --absolute-paths     Use absolute paths in output
      --strict             Fail on first syntax error (default: best-effort)
  -v, --verbose           Verbose output (-vv, -vvv for more detail)
  -h, --help              Show help message
      --version           Show version information
```

### 4.3 Output Format Specification

**Markdown Format (Default):**
```markdown
# AI Distiller Output
- Version: 1.0
- Source: /path/to/project
- Generated: 2025-06-15T10:30:00Z

---

## `src/main.go`

```go
package main

type Config struct {
    Port int
    Host string
}

func main()
func loadConfig() (*Config, error)
```
---
```

**JSONL Format (Machine-Readable):**
```json
{"version":"1.0","type":"header","source":"/path/to/project","generated":"2025-06-15T10:30:00Z"}
{"type":"file","path":"src/main.go","language":"go","content":"package main\n\ntype Config struct..."}
```

**JSON-Structured Format (Rich Semantic Data):**
```json
{
  "version": "1.0",
  "source": "/path/to/project",
  "generated": "2025-06-15T10:30:00Z",
  "files": [{
    "path": "src/main.go",
    "language": "go",
    "errors": [],
    "nodes": [{
      "type": "struct",
      "name": "Config",
      "visibility": "public",
      "fields": [
        {"name": "Port", "type": "int", "visibility": "public"},
        {"name": "Host", "type": "string", "visibility": "public"}
      ]
    }, {
      "type": "function",
      "name": "main",
      "signature": "func main()",
      "visibility": "public"
    }, {
      "type": "function",
      "name": "loadConfig",
      "signature": "func loadConfig() (*Config, error)",
      "visibility": "private"
    }]
  }]
}
```

**XML Format (Legacy Support):**
```xml
<aid-distiller-output version="1.0" source="/path/to/project">
  <file path="src/main.go" language="go">
    <struct name="Config" visibility="public">
      <field name="Port" type="int"/>
      <field name="Host" type="string"/>
    </struct>
    <function name="main" signature="func main()" visibility="public"/>
    <function name="loadConfig" signature="func loadConfig() (*Config, error)" visibility="private"/>
  </file>
</aid-distiller-output>
```

## 5. Non-Functional Requirements

### 5.1 Performance Requirements

**NFR-P1: Processing Speed**
- P95 latency: <2 seconds for 10MB codebase (500 files)
- Startup time: <50ms for `aid --help`
- Memory usage: <500MB for typical projects

**NFR-P2: Scalability**
- Support codebases up to 1GB
- Handle directories with 50,000+ files
- Process files up to 10MB individually

### 5.2 Security Requirements

**NFR-S1: Local Processing**
- All processing happens locally
- No network connections
- No telemetry or usage tracking

**NFR-S2: Input Validation**
- Protection against symbolic link loops
- Path traversal prevention
- Resource exhaustion protection

**NFR-S3: Dependency Security**
- Regular vulnerability scanning
- Minimal dependency footprint
- Automated security updates

### 5.3 Usability Requirements

**NFR-U1: Installation**
- Single binary download
- No runtime dependencies
- <1 minute installation process

**NFR-U2: CLI Design**
- POSIX-compliant interface
- Intuitive flag names
- Comprehensive help text
- Clear error messages

### 5.4 Compatibility Requirements

**NFR-C1: Platform Support**
- Windows (10+): x86_64, ARM64
- macOS (11+): x86_64, ARM64
- Linux: x86_64, ARM64
- FreeBSD: x86_64 (stretch goal)

**NFR-C2: Backwards Compatibility**
- Semantic versioning
- Stable CLI interface
- Format version in output

## 6. User Stories & Use Cases

### 6.1 Primary User Stories

**As a developer working with an AI assistant:**
1. I want to quickly generate a condensed overview of my entire codebase so that I can ask the AI architectural questions about my project
2. I want to extract just the public APIs from a library so that the AI can help me write integration code
3. I want to remove all implementation details and comments so that the AI can focus on the structure and interfaces

**As an AI/ML engineer building tools:**
1. I want to programmatically extract code structure in a consistent format so that I can build automated documentation or analysis tools
2. I want to process multiple projects in batch so that I can create training datasets or perform large-scale analysis

**As a technical lead onboarding new developers:**
1. I want to generate a high-level overview of our codebase structure so that new team members can quickly understand the architecture
2. I want to extract only the critical interfaces and data models so that developers can see the contract without getting lost in implementation

### 6.2 Detailed Use Cases

**Use Case 1: Architectural Review**
- **Actor**: Senior Developer
- **Goal**: Get AI assistance reviewing system architecture
- **Preconditions**: Large codebase with complex architecture
- **Steps**:
  1. Navigate to project root
  2. Run `aid . --implementation=0,comments`
  3. Copy output to AI assistant
  4. Ask questions about design patterns, potential improvements
- **Success Criteria**: AI provides meaningful architectural insights
- **Alternative Flows**: Use `--format jsonl` for tool integration

**Use Case 2: API Documentation Generation**
- **Actor**: Technical Writer
- **Goal**: Generate API documentation with AI assistance
- **Preconditions**: Codebase with public APIs
- **Steps**:
  1. Run `aid ./src --implementation=0,non-public,comments --include "*.ts"`
  2. Feed output to AI with documentation template
  3. Review and refine generated documentation
- **Success Criteria**: Complete, accurate API documentation
- **Alternative Flows**: Process specific modules separately

**Use Case 3: Security Audit Preparation**
- **Actor**: Security Engineer
- **Goal**: Prepare codebase summary for security analysis
- **Preconditions**: Access to full codebase
- **Steps**:
  1. Run `aid . --comments=0 --include "*.go,*.js" --exclude "*_test.*"`
  2. Provide output to AI security analysis tool
  3. Focus on identified risk areas
- **Success Criteria**: Comprehensive security findings
- **Alternative Flows**: Target specific subsystems

## 7. Technical Architecture

### 7.1 System Design Principles

**Modularity**: Clear separation between core engine, language processors, and output formatters
**Extensibility**: Easy addition of new languages and output formats via well-defined interfaces
**Performance**: Single-pass processing with concurrent pipeline, minimal memory footprint
**Reliability**: Graceful error handling, deterministic output, no silent failures
**Maintainability**: Visitor pattern for transformations, immutable data structures

### 7.2 High-Level Architecture

```
┌─────────────────┐
│   CLI (main)    │
└────────┬────────┘
         │
┌────────▼────────┐
│   Distiller     │◄─── Core orchestration & concurrency
├─────────────────┤
│ - File Discovery│
│ - Worker Pool   │
│ - Pipeline Mgmt │
└────────┬────────┘
         │
┌────────▼────────┐
│  Parser Router  │◄─── Language detection & dispatch
└────────┬────────┘
         │
┌────────▼────────────────────┐
│   Language Processors       │
├─────────────────────────────┤
│ Go │ Python │ JS │ ... │ C# │◄─── AST parsing to IR
└────────┬────────────────────┘
         │
┌────────▼────────┐
│ Intermediate    │◄─── Rich semantic representation
│ Representation  │
└────────┬────────┘
         │
┌────────▼────────┐
│ Stripping       │◄─── Single-pass visitor pattern
│ Visitor         │
└────────┬────────┘
         │
┌────────▼────────┐
│ Output Formatter│◄─── Format generation
├─────────────────┤
│ MD │ JSONL │ XML│ JSON-Structured
└─────────────────┘
```

### 7.3 Core Architecture Components

#### 7.3.1 Intermediate Representation (IR)

The IR is the heart of the architecture, providing a language-agnostic representation of code structure:

```go
// Base types for all IR nodes
type Location struct {
    StartLine   int
    StartColumn int
    EndLine     int
    EndColumn   int
}

type DistilledNode interface {
    Accept(visitor IRVisitor) DistilledNode
    GetLocation() Location
}

// Concrete IR types
type DistilledFile struct {
    Path     string
    Language string
    Children []DistilledNode
    Errors   []DistilledError
}

type DistilledFunction struct {
    BaseNode
    Name       string
    Signature  string
    Visibility string              // "public", "private", "protected"
    Modifiers  []string            // "static", "async", "virtual"
    Attributes map[string]any      // Language-specific metadata
    Body       []DistilledNode     // Optional, stripped if implementation removed
}

type DistilledClass struct {
    BaseNode
    Name       string
    Visibility string
    Extends    []string
    Implements []string
    Members    []DistilledNode
}

type DistilledError struct {
    BaseNode
    Message  string
    Severity string // "error", "warning"
}
```

#### 7.3.2 Language Processor Interface

```go
type LanguageProcessor interface {
    // Parse produces a rich IR from the AST
    Parse(tree *sitter.Tree, source []byte) (*DistilledFile, error)
    
    // GetLanguage returns the tree-sitter language
    GetLanguage() *sitter.Language
}
```

#### 7.3.3 Visitor Pattern for Stripping

```go
// Visitor interface for single-pass tree traversal
type IRVisitor interface {
    VisitFile(node *DistilledFile) IRVisitor
    VisitFunction(node *DistilledFunction) IRVisitor
    VisitClass(node *DistilledClass) IRVisitor
    VisitComment(node *DistilledComment) IRVisitor
    // ... other node types
}

// Rule-based stripping configuration
type StrippingRule interface {
    Name() string
    ShouldRemove(node DistilledNode, context VisitorContext) bool
    Transform(node DistilledNode, context VisitorContext) DistilledNode
}

type VisitorContext struct {
    Language string
    Parent   DistilledNode
    Depth    int
    Options  StrippingOptions
}
```

### 7.4 Parser Strategy & Technology Choices

#### 7.4.1 Parsing Technology

**Decision Pending**: Dual PoC approach to evaluate:

1. **Option A: Tree-sitter with CGo**
   - Uses official `github.com/tree-sitter/go-tree-sitter`
   - Cross-compilation via Zig CC
   - Native performance

2. **Option B: Tree-sitter with WASM**
   - Uses Wazero runtime for WASM execution
   - Pure Go, no CGo required
   - Sandboxed execution

**Evaluation Criteria**:
- Build complexity
- Binary size (<50MB requirement)
- Startup latency (<50ms requirement)
- Parse throughput
- Grammar compatibility

#### 7.4.2 Supported Languages (Phase 1)

| Language | Extensions | Visibility Detection |
|----------|-----------|---------------------|
| Go | `*.go` | Capital letter convention |
| Python | `*.py` | Underscore prefix convention |
| JavaScript/TypeScript | `*.js`, `*.jsx`, `*.ts`, `*.tsx` | Export keyword |
| Java | `*.java` | Access modifiers |
| C# | `*.cs` | Access modifiers |
| Rust | `*.rs` | `pub` keyword |
| Ruby | `*.rb` | Method visibility methods |
| PHP | `*.php` | Access modifiers |

### 7.5 Concurrency Model

The system uses a concurrent pipeline for maximum performance:

```go
// Pipeline stages with bounded channels for backpressure
type Pipeline struct {
    FileDiscovery chan string           // Discovered file paths
    ParseResults  chan *DistilledFile   // Parsed IR objects
    StripResults  chan *DistilledFile   // Stripped IR objects
    Workers       int                   // Number of parser workers
}

// Concurrent processing with deterministic output
func (d *Distiller) Process() error {
    // 1. File discovery (single goroutine)
    go d.discoverFiles()
    
    // 2. Parser workers (N goroutines)
    g, ctx := errgroup.WithContext(context.Background())
    for i := 0; i < d.Workers; i++ {
        g.Go(func() error {
            return d.parseWorker(ctx)
        })
    }
    
    // 3. Stripping (single goroutine for order preservation)
    g.Go(func() error {
        return d.stripWorker(ctx)
    })
    
    // 4. Output formatting (single goroutine)
    g.Go(func() error {
        return d.outputWorker(ctx)
    })
    
    return g.Wait()
}
```

### 7.6 Error Handling Strategy

1. **Syntax Errors**: Best-effort distillation with warnings
   - Check for error nodes in AST
   - Process valid top-level constructs only
   - Include errors in IR for reporting

2. **Processing Errors**: Fail-fast with context
   - File path and line number in errors
   - `--strict` flag to fail on first error
   - Structured error reporting in output

3. **Resource Protection**:
   - Per-file timeouts
   - Memory limits
   - Symlink loop detection

### 7.7 Output Format Enhancements

Beyond basic formats, support for structured output:

**JSON-Structured Format**:
```json
{
  "version": "1.0",
  "source": "/path/to/project",
  "files": [{
    "path": "src/main.go",
    "language": "go",
    "errors": [],
    "nodes": [{
      "type": "function",
      "name": "main",
      "visibility": "public",
      "signature": "func main()",
      "location": {"start": {"line": 10, "col": 1}, "end": {"line": 15, "col": 2}}
    }]
  }]
}
```

## 8. User Configuration

### 8.1 Configuration File Support

Future enhancement to support `.aidrc` configuration file:

```yaml
# .aidrc
strip:
  - implementation
  - comments
exclude:
  - "*_test.go"
  - "vendor/*"
  - "node_modules/*"
format: markdown
recursive: true
```

### 8.2 Environment Variables

- `AID_DEFAULT_FORMAT`: Default output format
- `AID_DEFAULT_STRIP`: Default strip options
- `AID_VERBOSE`: Default verbosity level

## 9. Quality & Testing Requirements

### 9.1 Testing Strategy

**Unit Tests**:
- Each language processor: >90% coverage
- Core distiller logic: >95% coverage
- CLI argument parsing: 100% coverage

**Integration Tests**:
- End-to-end CLI tests for all major use cases
- Cross-platform binary tests
- Performance benchmarks

**Test Data**:
- Comprehensive test files for each language
- Edge cases (empty files, syntax errors)
- Real-world project snapshots

### 9.2 Quality Standards

**Code Quality**:
- Go lint compliance
- Consistent code formatting (gofmt)
- Comprehensive error handling
- No panic in production code

**Documentation Quality**:
- Inline code documentation
- Comprehensive README
- Architecture documentation
- Contributing guidelines

### 9.3 Performance Benchmarks

Automated benchmarks for:
- Various file sizes (1KB to 10MB)
- Different languages
- Strip option combinations
- Memory usage profiling

## 10. Release & Distribution Requirements

### 10.1 Binary Distribution

**Release Artifacts**:
- `aid-windows-amd64.exe`
- `aid-windows-arm64.exe`
- `aid-darwin-amd64`
- `aid-darwin-arm64`
- `aid-linux-amd64`
- `aid-linux-arm64`

**Checksums**: SHA256 for all binaries

### 10.2 Installation Methods

1. **Direct Download**: From GitHub releases
2. **Install Script**: Universal `install.sh`
3. **Package Managers** (future):
   - Homebrew (macOS)
   - Scoop/Chocolatey (Windows)
   - APT/YUM (Linux)

### 10.3 Release Process

1. Automated via GitHub Actions
2. Triggered by version tags (e.g., `v1.0.0`)
3. Runs full test suite
4. Builds all platform binaries
5. Creates GitHub release with artifacts
6. Updates documentation

## 11. Documentation Deliverables

### 11.1 End User Documentation

- **README.md**: Project overview, installation, quick start
- **USAGE.md**: Comprehensive usage guide with examples
- **FAQ.md**: Common questions and troubleshooting

### 11.2 Developer Documentation

- **ARCHITECTURE.md**: System design and components
- **CONTRIBUTING.md**: Development setup and guidelines
- **LANGUAGE_SUPPORT.md**: Adding new language processors

### 11.3 Integration Documentation

- **INTEGRATION_GUIDE.md**: Using aid in tools and scripts
- **API.md**: Output format specifications
- **MCP_TOOL.md**: Model Context Protocol integration

## 12. Extensibility & Future Roadmap

### 12.1 Plugin Architecture (Future)

- Dynamic loading of language processors
- Community-contributed processors
- Custom stripping rules

### 12.2 Enhanced Features (Future)

- Dependency graph extraction
- Cross-file relationship mapping
- Semantic diff generation
- IDE integrations

### 12.3 Ecosystem Growth

- Official language processor registry
- Processor development kit
- Community processor validation

## 13. Constraints & Dependencies

### 13.1 Technical Constraints

- Must use tree-sitter for parsing (architectural decision)
- Single binary requirement limits dynamic features
- Memory constraints for very large files

### 13.2 Dependencies

**Build Dependencies**:
- Go 1.21+
- Zig CC (for cross-compilation)
- Tree-sitter grammars (as git submodules)

**Runtime Dependencies**:
- None (static binary)

### 13.3 Assumptions

- Users have basic CLI familiarity
- File system supports standard operations
- Source code may contain syntax errors (handled gracefully)

## 14. Risk Analysis & Mitigation

### 14.1 Technical Risks

**Risk**: Cross-compilation complexity with CGo
- **Impact**: High
- **Probability**: Medium
- **Mitigation**: Dual PoC approach (CGo vs WASM), Zig CC toolchain, Wazero as fallback

**Risk**: Language parser maintenance burden
- **Impact**: Medium
- **Probability**: High
- **Mitigation**: Active tree-sitter community, automated updates

**Risk**: Performance degradation with large files
- **Impact**: Medium
- **Probability**: Medium
- **Mitigation**: Streaming processing, file size limits

### 14.2 Project Risks

**Risk**: Competing solutions emerge
- **Impact**: Medium
- **Probability**: Medium
- **Mitigation**: Fast execution, community building, unique features

**Risk**: AI context windows increase dramatically
- **Impact**: High
- **Probability**: Low (short-term)
- **Mitigation**: Focus on value beyond size (structure, filtering)

## 15. Success Criteria

### 15.1 Launch Success (Month 1)

- [ ] All planned languages supported
- [ ] <100ms startup time achieved
- [ ] Zero critical bugs in production
- [ ] 1,000+ GitHub stars

### 15.2 Growth Success (Month 6)

- [ ] 10,000+ monthly active users
- [ ] 50+ community contributors
- [ ] Integration in 3+ AI tools
- [ ] 15+ languages supported

### 15.3 Maturity Success (Year 1)

- [ ] Industry standard for code distillation
- [ ] 50,000+ GitHub stars
- [ ] Corporate adoption (Fortune 500)
- [ ] Sustainable maintenance model

## Appendices

### A. Glossary

- **AST**: Abstract Syntax Tree - hierarchical representation of source code
- **Context Window**: Maximum input size for an LLM
- **Distillation**: Process of extracting essential information
- **Tree-sitter**: Incremental parsing library for programming languages

### B. References

- Tree-sitter Documentation: https://tree-sitter.github.io/
- Go Modules: https://golang.org/ref/mod
- POSIX CLI Guidelines: https://pubs.opengroup.org/onlinepubs/9699919799/

### C. Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-06-15 | Claude & Gemini | Initial PRD |
| 1.1 | 2025-06-15 | Claude & Gemini | Updated architecture with IR, Visitor pattern, dual PoC strategy |

---

*This document represents the complete product requirements for AI Distiller v1.0. It will be updated as the project evolves based on user feedback and technical discoveries.*