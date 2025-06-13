# AI Distiller - Claude Development Instructions

## Project Overview

AI Distiller is a high-performance CLI tool that extracts essential code structure from large codebases, making them digestible for LLMs by removing unnecessary details while preserving semantic information. Think of it as "code compression for AI context windows."

## Key Project Goals

1. **Single native binary** - No runtime dependencies, works everywhere
2. **Blazingly fast** - Process 10MB codebase in <2 seconds
3. **Multi-language support** - 15+ languages via tree-sitter WASM
4. **Flexible output** - Text, Markdown, JSON, JSONL, XML formats
5. **Granular control** - Strip exactly what you don't need

## CLI Interface Specification

The main CLI is `aid` (AI Distiller):

```bash
aid [path] [flags]

# Examples:
aid                                    # Process current directory
aid src/                              # Process src directory
aid main.py                           # Process single file
aid --strip comments,implementation   # Remove comments and implementations
aid --format json --output api.json   # JSON output to file
aid --strip non-public --stdout       # Print only public members to stdout
```

### Important Flags

- `--strip <items>` - **THE MOST IMPORTANT FLAG**
  - Values: `comments`, `imports`, `implementation`, `non-public`
  - Comma-separated: `--strip comments,implementation`
  - Default: Nothing stripped (full output)
  
- `--format <fmt>` - Output format
  - `md` (default) - Human-readable Markdown
  - `text` - Ultra-compact plaintext (best for AI)
  - `jsonl` - One JSON object per file
  - `json-structured` - Rich semantic data
  - `xml` - Structured XML

- `-o, --output <file>` - Output file (default: auto-generated)
  - Default pattern: `.<dirname>.[strip-options].aid.txt`
  - Example: `.MyProject.ncom.nimpl.aid.txt`

## Expected Output Examples

### Text Format (Ultra-Compact for AI)

This is the most compact format, optimized for maximum context efficiency. **Best choice for AI consumption** because:
- Minimal syntax overhead
- Natural code-like appearance
- Maximum information density
- Clear file boundaries with `<file path="...">` tags
- No decorative elements (emojis, tables, etc.)

```
<file path="src/user_service.py">
from typing import List, Optional
from datetime import datetime

class UserService:
    def __init__(self, db_connection):
        self.db = db_connection
        self._cache = {}
    
    def get_user(self, user_id: int) -> Optional[User]:
        # Implementation here if not stripped
        
    def _validate_email(self, email: str) -> bool:
        # Private method
</file>

<file path="src/models.py">
class User:
    def __init__(self, id: int, name: str, email: str):
        self.id = id
        self.name = name
        self.email = email
</file>
```

With `--strip implementation,non-public`:

```
<file path="src/user_service.py">
from typing import List, Optional
from datetime import datetime

class UserService:
    def __init__(self, db_connection)
    def get_user(self, user_id: int) -> Optional[User]
</file>

<file path="src/models.py">
class User:
    def __init__(self, id: int, name: str, email: str)
</file>
```

### Markdown Format (Default)

```markdown
# src/user_service.py

## Structure

📥 **Import** from `typing` import `List`, `Optional` <sub>L1</sub>
📥 **Import** from `datetime` import `datetime` <sub>L2</sub>

🏛️ **Class** `UserService` <sub>L5-45</sub>
  🔧 **Function** `__init__`(`self`, `db_connection`) <sub>L8-10</sub>
    ```python
    self.db = db_connection
    self._cache = {}
    ```
  🔧 **Function** `get_user`(`self`, `user_id`: `int`) → `Optional[User]` <sub>L12-18</sub>
  🔧 **Function** `_validate_email` _private_(`self`, `email`: `str`) → `bool` <sub>L20-25</sub>
```

### With `--strip implementation`:

```markdown
# src/user_service.py

## Structure

📥 **Import** from `typing` import `List`, `Optional` <sub>L1</sub>
📥 **Import** from `datetime` import `datetime` <sub>L2</sub>

🏛️ **Class** `UserService` <sub>L5-45</sub>
  🔧 **Function** `__init__`(`self`, `db_connection`) <sub>L8-10</sub>
  🔧 **Function** `get_user`(`self`, `user_id`: `int`) → `Optional[User]` <sub>L12-18</sub>
  🔧 **Function** `_validate_email` _private_(`self`, `email`: `str`) → `bool` <sub>L20-25</sub>
```

### With `--strip non-public`:

```markdown
# src/user_service.py

## Structure

📥 **Import** from `typing` import `List`, `Optional` <sub>L1</sub>
📥 **Import** from `datetime` import `datetime` <sub>L2</sub>

🏛️ **Class** `UserService` <sub>L5-45</sub>
  🔧 **Function** `__init__`(`self`, `db_connection`) <sub>L8-10</sub>
  🔧 **Function** `get_user`(`self`, `user_id`: `int`) → `Optional[User]` <sub>L12-18</sub>
```

## Architecture & Implementation Flow

### 1. Component Architecture

```
User Input → CLI Parser → File Discovery → Language Detection →
Parser (tree-sitter WASM) → IR Generation → Stripper Visitor →
Output Formatter → File/Stdout
```

### 2. Directory Structure

```
ai-distiller/
├── cmd/
│   └── aid/              # Main CLI entry point
├── internal/
│   ├── cli/              # CLI logic and flags
│   ├── ir/               # Intermediate Representation
│   ├── parser/           # WASM runtime and tree-sitter
│   ├── language/         # Language-specific processors
│   │   ├── python/       # Python processor
│   │   ├── go/           # Go processor
│   │   └── javascript/   # JavaScript processor
│   ├── stripper/         # Visitor for stripping elements
│   ├── formatter/        # Output formatters
│   └── processor/        # Core processing logic
└── test-data/            # Test files and scenarios
```

### 3. Key Implementation Details

#### Language Processor Interface

```go
type LanguageProcessor interface {
    Language() string
    Version() string
    SupportedExtensions() []string
    CanProcess(filename string) bool
    Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error)
}
```

#### Stripper Options

```go
type Options struct {
    RemovePrivate         bool  // --strip non-public
    RemoveImplementations bool  // --strip implementation
    RemoveComments        bool  // --strip comments
    RemoveImports         bool  // --strip imports
}
```

#### IR Node Types

- `DistilledFile` - Root node for a file
- `DistilledClass` - Class/struct definition
- `DistilledFunction` - Function/method
- `DistilledImport` - Import statement
- `DistilledComment` - Comment/docstring
- `DistilledField` - Class field/property

## Development Workflow for AI Assistants

### CRITICAL: No Mocks or Simulated Functions

**NEVER create mock implementations or simulated functions.** All code must be real, working implementations. This includes:
- No hardcoded test data pretending to be parsed results
- No mock parsers returning fixed data
- No placeholder functions that don't actually work
- Tests must test REAL functionality, not mocked behavior

If something can't be implemented properly, document it as TODO but don't create fake implementations.

### 1. When Adding New Features

1. **Check existing patterns** - Look at similar features
2. **Write REAL implementation** - No mocks or stubs
3. **Update tests first** - TDD approach with real tests
4. **Follow architecture** - Don't break the visitor pattern
5. **Test with real files** - Use test-data/ directory

### 2. When Fixing Bugs

1. **Reproduce in test** - Add failing test case
2. **Fix minimally** - Don't refactor unnecessarily
3. **Run all tests** - `make test` in project root
4. **Check performance** - Must stay fast

### 3. Common Tasks

#### Adding a new language:

1. Create `internal/language/<lang>/processor.go`
2. Implement `LanguageProcessor` interface
3. Add tree-sitter WASM grammar
4. Register in `internal/language/registry.go`
5. Add comprehensive tests

#### Adding a new output format:

1. Create formatter in `internal/formatter/`
2. Implement `Formatter` interface
3. Register in formatter registry
4. Add tests for all node types

Example for text formatter:
```go
// internal/formatter/text.go
type TextFormatter struct {
    options Options
}

func (f *TextFormatter) Format(w io.Writer, node ir.DistilledNode) error {
    if file, ok := node.(*ir.DistilledFile); ok {
        fmt.Fprintf(w, "<file path=\"%s\">\n", file.Path)
        
        // Write distilled content as plain text
        for _, child := range file.Children {
            f.formatNode(w, child, 0)
        }
        
        fmt.Fprintln(w, "</file>")
    }
    return nil
}

func (f *TextFormatter) formatNode(w io.Writer, node ir.DistilledNode, indent int) {
    switch n := node.(type) {
    case *ir.DistilledImport:
        if n.ImportType == "from" {
            fmt.Fprintf(w, "from %s import %s\n", n.Module, formatSymbols(n.Symbols))
        } else {
            fmt.Fprintf(w, "import %s\n", n.Module)
        }
    case *ir.DistilledClass:
        fmt.Fprintf(w, "\nclass %s", n.Name)
        if len(n.Extends) > 0 {
            fmt.Fprintf(w, "(%s)", formatTypeRefs(n.Extends))
        }
        fmt.Fprintln(w, ":")
        // Format children with indentation
    case *ir.DistilledFunction:
        fmt.Fprintf(w, "%s%s(%s)", 
            strings.Repeat("    ", indent),
            n.Name,
            formatParams(n.Parameters))
        if n.Returns != nil {
            fmt.Fprintf(w, " -> %s", n.Returns.Name)
        }
        if !f.options.Compact && n.Implementation != "" {
            fmt.Fprintf(w, ":\n%s", n.Implementation)
        }
        fmt.Fprintln(w)
    }
}
```

#### Modifying strip behavior:

1. Update `internal/stripper/stripper.go`
2. Modify the `Visit` method for affected nodes
3. Test with various combinations

## Testing Strategy

### Unit Tests
```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/stripper/...
```

### Integration Tests
```bash
# In test-data directory
make test-all      # Run all test types
make test-unit     # Unit tests only
make test-quick    # Quick smoke tests
```

### Test Scenarios

1. **full_output** - Everything included
2. **no_private** - Strip private members
3. **no_implementation** - Strip function bodies
4. **minimal** - Structure only
5. **complex_imports** - Import handling
6. **edge_cases** - Unicode, async, etc.

## Performance Guidelines

1. **Concurrent but ordered** - Process files in parallel, maintain output order
2. **Stream everything** - Don't load entire codebases in memory
3. **One-pass visiting** - No multiple IR traversals
4. **Bounded channels** - Prevent memory explosions

## Common Pitfalls & Solutions

### Issue: Tests expect `--no-private` flag
**Solution**: Use `--strip non-public` or default behavior (no flag = no private)

### Issue: Parser doesn't find all constructs
**Solution**: Check if line-based parser limitations; full AST via tree-sitter coming

### Issue: Text format not preserving syntax exactly
**Solution**: Text format aims for readability and compactness, not valid source code

### Issue: Output format inconsistency
**Solution**: All formatters must pass `format_validator.go` tests

### Issue: Performance degradation
**Solution**: Profile with `go test -bench`, check for unnecessary allocations

## Communication with User

- Use **Czech** for general communication if user writes in Czech
- Use **English** for all code, comments, and technical documentation
- Be concise but thorough
- Show real examples when explaining

## Communication with AI Assistants (Gemini, o3)

- Always communicate in **English** when using Zen MCP tools
- Use 'pro' model for Gemini for deep analysis
- Use 'o3' model (not 'o3-mini') for o3 conversations
- Request deep thinking modes when appropriate

## Git Commit Guidelines

### CRITICAL: Pre-commit Checklist

**Before EVERY commit, you MUST check for unwanted files:**

1. **Run `git status` to see all changes**
2. **Check for and remove:**
   - Temporary debug files (*.tmp, *.log, debug.*)
   - Built test binaries (aid, test executables)
   - Test output files (.*.aid.txt in root directory)
   - IDE/editor files (.vscode/, .idea/, *.swp)
   - Personal test files not meant for the repo
   - Any one-off debugging scripts

3. **Use `.gitignore` properly:**
   - Check if unwanted files should be in .gitignore
   - Add patterns for commonly generated files
   - Never commit files that should be ignored

4. **Clean commit checklist:**
   ```bash
   # Before committing, run:
   git status              # Check what will be committed
   git diff --cached       # Review actual changes
   ls -la                  # Look for unwanted files in root
   find . -name "*.tmp"    # Find temporary files
   find . -name "*.log"    # Find log files
   ```

5. **If you accidentally staged unwanted files:**
   ```bash
   git reset HEAD <file>   # Unstage specific file
   git clean -n            # Preview what would be cleaned
   git clean -f            # Remove untracked files (careful!)
   ```

Remember: A clean repository is a professional repository!

## Parser Development Guide

### Key Principles Learned from Go/TypeScript/Python/JavaScript

This guide documents the successful patterns discovered during systematic parser development:

#### 1. **Architecture Pattern: Tree-Sitter + Stripper Integration**

All parsers follow this proven architecture:

```go
// 1. Tree-sitter AST parsing
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
    // Parse with tree-sitter
    file, err := p.treeparser.ProcessSource(ctx, source, filename)
    if err != nil {
        return nil, err
    }
    
    // 2. Apply standardized stripper
    stripperOpts := stripper.Options{
        RemovePrivate:         !opts.IncludePrivate,
        RemoveImplementations: !opts.IncludeImplementation,
        RemoveComments:        !opts.IncludeComments,
        RemoveImports:         !opts.IncludeImports,
    }
    
    if /* anything to strip */ {
        s := stripper.New(stripperOpts)
        stripped := file.Accept(s)
        return stripped.(*ir.DistilledFile), nil
    }
    
    return file, nil
}
```

#### 2. **Method Association Pattern**

Methods must be properly nested under their classes/structs:

- **Go**: Two-pass processing (collect types → associate methods)
- **TypeScript**: Single-pass with proper AST traversal  
- **Python**: Native tree-sitter handles this automatically
- **JavaScript**: Class methods parsed within class body

#### 3. **Interface/Protocol Satisfaction Detection**

Each language has its own approach:

- **Go**: Structural analysis comparing method sets
- **TypeScript**: Explicit `implements` + disabled structural (too complex)
- **Python**: Duck typing analysis with Protocol detection
- **JavaScript**: Inheritance tracking via `extends`

#### 4. **Visibility Handling**

Language-specific visibility rules:

- **Go**: Uppercase = public, lowercase = package-private
- **TypeScript**: `public`/`private`/`protected` keywords
- **Python**: `_private`, `__dunder__` = public API
- **JavaScript**: `#private`, JSDoc `@private`, `_convention`

#### 5. **Critical Implementation Details**

**Tree-sitter Node Text Safety:**
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

**Stripper Integration (CRITICAL):**
- NEVER use custom `applyOptions` - always use standardized `stripper.New()`
- Ensures consistent behavior across all languages

**AST Traversal Pattern:**
```go
func (p *Parser) processNode(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
    switch node.Type() {
    case "class_declaration", "abstract_class_declaration":
        p.parseClass(node, file, parent)
    case "function_declaration":
        p.parseFunction(node, file, parent)
    // ... other cases
    default:
        // Recurse into children
        for i := 0; i < int(node.ChildCount()); i++ {
            p.processNode(node.Child(i), file, parent)
        }
    }
}
```

### 6. **Testing Validation Pattern**

For each language:
1. Create complex examples with inheritance, generics, privacy
2. Test all stripping modes: full, `--strip implementation`, `--strip non-public`
3. Validate method association and interface detection
4. Compare outputs with AI assistant for accuracy review

### 7. **Common Pitfalls & Solutions**

❌ **Using line-based regex parsing** → ✅ Use tree-sitter AST  
❌ **Custom option filtering** → ✅ Use standardized stripper  
❌ **Missing method association** → ✅ Ensure proper parent/child IR structure  
❌ **Broken node text extraction** → ✅ Add boundary checks  
❌ **Ignoring language-specific features** → ✅ Handle generics, async, etc.

### 8. **Performance Guidelines**

- Single-pass AST traversal where possible
- Bounded context (avoid full-program analysis)
- Cache tree-sitter parsers
- Stream processing for large files

## Next Steps & TODOs

1. **More languages** - Java, C#, Rust following these patterns
2. **Semantic features** - Call graphs, dependency analysis  
3. **Performance optimization** - Sub-50ms for small files
4. **Release automation** - GitHub Actions for multi-platform builds

Remember: The goal is to make code understandable for AI, not humans. Optimize for context efficiency!

## Comprehensive Language Testing Protocol

### Original Instructions from User

The user requested systematic comprehensive testing of AI Distiller across all 12+ supported programming languages with the following specific workflow:

### Testing Workflow for Each Language

1. **Gemini Construct Design Phase**
   - Collaborate with Gemini to design 5 constructs per language (basic → very complex)
   - Each construct should represent real-world code patterns
   - Focus on language-specific features and idiomatic patterns
   - Constructs should progressively increase in complexity:
     - Construct 1: Basic (simple functions, basic types)
     - Construct 2: Simple (classes/structs, basic OOP)
     - Construct 3: Medium (interfaces, inheritance, advanced features)
     - Construct 4: Complex (advanced language features, frameworks)
     - Construct 5: Very Complex (edge cases, meta-programming, extreme patterns)

2. **Implementation Phase**
   - Create source files (.py, .ts, .go, etc.) for each construct
   - Generate expected output files for 3 distillation options:
     - `expected_full.txt` - Full distillation (all content)
     - `expected_no_private.txt` - `--strip non-public` (hide private/internal members)
     - `expected_no_impl.txt` - `--strip implementation` (preserve signatures only)

3. **Testing Phase** 
   - Implement comprehensive unit tests (5 constructs × 3 options = 15 test cases per language)
   - Tests must validate actual parsing results, not mock data
   - Include construct-specific validations for language features

4. **CRITICAL: Gemini Output Review Phase**
   - **Generate actual distillation outputs** using AI Distiller
   - **Send outputs to Gemini for accuracy verification**
   - Gemini reviews against 4 criteria:
     a. **Correctness Against Specification** - Accuracy, completeness, signature integrity
     b. **Idiomatic Representation** - Natural language feel, clean formatting
     c. **Edge Case Robustness** - Complex syntax handling, feature interpretation
     d. **Contextual Integrity** - Import handling, dependency preservation
   - Fix any issues identified by Gemini
   - Iterate until Gemini confirms outputs are correct

5. **Completion and Progression**
   - Mark language as completed only after Gemini approval
   - Commit comprehensive implementation
   - Move to next language systematically

### Language Priority Order

Based on project support and complexity:
1. ✅ **Python** - Completed (5 constructs, 15 tests, Gemini review pending)
2. ✅ **TypeScript** - Completed (6 constructs, 18 tests, Gemini review pending)
3. 🔄 **Go** - In progress (constructs designed by Gemini)
4. **Rust** - Planned
5. **Swift** - Planned
6. **Ruby** - Planned
7. **Java** - Planned
8. **C#** - Planned
9. **Kotlin** - Planned
10. **C++** - Planned
11. **JavaScript** - Planned
12. **PHP** - Planned

### Current Status & Next Actions

**IMMEDIATE PRIORITY:** Before continuing with Go, must complete Gemini review of Python and TypeScript actual outputs as specified in original instructions.

### Key Testing Principles

- **No mocks or simulated data** - All tests use real AI Distiller output
- **Gemini collaboration is mandatory** - Both for design and output verification
- **Systematic progression** - Complete each language fully before moving to next
- **Real-world relevance** - Constructs represent actual development patterns
- **Language-specific focus** - Test unique features of each language

### Communication Protocol

- **User communication:** Czech language for general discussion
- **Technical documentation:** English for all code, comments, technical specs  
- **Gemini communication:** English only, use 'pro' model for deep analysis
- **o3 communication:** English only, use 'o3' model (not 'o3-mini') when needed

This comprehensive testing ensures AI Distiller can accurately handle real-world codebases across all supported languages with verified accuracy.