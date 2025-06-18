# AI Distiller

> **Turn a million-line codebase into a 100K-token AI prompt in 30 seconds**

[![Go](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-488%20passing-brightgreen)](test-data/)
[![Tree-sitter](https://img.shields.io/badge/powered%20by-tree--sitter-green)](https://tree-sitter.github.io/)
[![MCP](https://img.shields.io/badge/MCP-compatible-blue)](https://modelcontextprotocol.io/)

AI Distiller extracts the essential structure from large codebases, creating compact representations perfect for LLM context windows. Think of it as **"code compression for AI"** - preserving what matters, discarding the noise.

```bash
# Example: Django's 10M tokens → 256K tokens in 0.23s
$ aid django/
Processing 970 files at 4,199 files/s...
✓ Reduced 10M tokens to 256K (-97%)
✓ Entire framework now fits in Claude's context!
```

## Why AI Distiller?

<table>
<tr>
<th>🤖 For AI Engineers</th>
<th>👨‍💻 For Developers</th>
<th>🔍 For Code Reviewers</th>
</tr>
<tr>
<td>

```bash
# Turn 10MB of code into 
# 200KB of structure
aid ./src --format text \
  --strip "implementation,comments"
```

Feed entire codebases to LLMs without hitting token limits

</td>
<td>

```bash
# Get instant API overview
aid ./api --strip "non-public" \
  --output api-surface.txt
```

Understand new codebases in minutes, not hours

</td>
<td>

```bash
# Extract only public changes
aid . --strip "non-public,implementation" \
  --format json | jq '.symbols'
```

Focus on what really changed in PRs

</td>
</tr>
</table>

## 🎯 How It Works

1. **Scans** your codebase recursively for supported file types
2. **Parses** each file using language-specific tree-sitter parsers (all bundled, no dependencies)
3. **Extracts** only what you need: public APIs, type signatures, class hierarchies
4. **Outputs** in your preferred format: ultra-compact text, markdown, or structured JSON

All tree-sitter grammars are compiled into the `aid` binary - zero external dependencies!

## 🚀 Quick Start

```bash
# Install via Go
go install github.com/janreges/ai-distiller/cmd/aid@latest

# Or build from source
git clone https://github.com/janreges/ai-distiller
cd ai-distiller
make build

# Basic usage
aid                                      # Current directory (parallel by default)
aid src/                                 # Specific directory
aid main.py utils.py                     # Specific files
aid -w 1 src/                           # Force serial processing
aid -w 16 src/                          # Use 16 parallel workers

# AI-optimized output (most compact)
aid --format text --strip "non-public,comments,implementation"

# Full structural analysis
aid --format json --output structure.json

# Generate AI analysis workflow (NEW!)
aid --ai-analysis-task-list                    # Complete project analysis
aid src/ --include "*.go,*.py" --ai-analysis-task-list  # Focus on specific languages
```

### 🤖 Use with Claude Desktop (MCP)

AI Distiller now integrates seamlessly with Claude Desktop through the Model Context Protocol (MCP), enabling AI agents to analyze and understand codebases directly within conversations.

```bash
# One-line installation
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp
```

Available MCP tools for AI agents:
- `distillDirectory` - Extract structure from entire codebase or specific modules/directories
- `distillFile` - Get detailed structure info for specific files
- `search` - Pattern search across entire codebase
- `listFiles` - Browse directories with file statistics
- `getFileContent` - Access raw file content when needed

**Smart Context Management**: AI agents can analyze your entire project for understanding the big picture, then zoom into specific modules (auth, API, database) for detailed work. No more overwhelming AI with irrelevant code!

See [MCP Integration Guide](docs/mcp-integration.md) for detailed setup instructions and advanced usage.

## 🚀 Transform Massive Codebases Into AI-Friendly Context

> **The Problem**: Modern codebases contain thousands of files with millions of lines. But for AI to understand your code architecture, suggest improvements, or help with development, it doesn't need to see every implementation detail - it needs the **structure and public interfaces**.

> **The Solution**: AI Distiller extracts only what matters - public APIs, types, and signatures - reducing codebase size by 90-98% while preserving all essential information for AI comprehension.

<table>
<tr>
<th align="left">Project</th>
<th align="center">Files</th>
<th align="right">Original Tokens</th>
<th align="right">Distilled Tokens</th>
<th align="center">Fits in Context<sup>1</sup></th>
<th align="center">Speed<sup>2</sup></th>
<th align="center">Language</th>
</tr>
<tr>
<td>⚛️ <code>react</code></td>
<td align="center">1,781</td>
<td align="right">~5.5M</td>
<td align="right"><strong>250K</strong> (-95%)</td>
<td align="center">✅ Claude/ChatGPT-4o</td>
<td align="center">2,875 files/s</td>
<td align="center">📜</td>
</tr>
<tr>
<td>🎨 <code>vscode</code></td>
<td align="center">4,768</td>
<td align="right">~22.5M</td>
<td align="right"><strong>2M</strong> (-91%)</td>
<td align="center">⚠️ Needs chunking</td>
<td align="center">5,072 files/s</td>
<td align="center">📘</td>
</tr>
<tr>
<td>🐍 <code>django</code></td>
<td align="center">970</td>
<td align="right">~10M</td>
<td align="right"><strong>256K</strong> (-97%)</td>
<td align="center">✅ Claude/ChatGPT-4o</td>
<td align="center">4,199 files/s</td>
<td align="center">🐍</td>
</tr>
<tr>
<td>📦 <code>prometheus</code></td>
<td align="center">685</td>
<td align="right">~8.5M</td>
<td align="right"><strong>154K</strong> (-98%)</td>
<td align="center">✅ Claude/Gemini</td>
<td align="center">3,071 files/s</td>
<td align="center">🐹</td>
</tr>
<tr>
<td>🦀 <code>rust-analyzer</code></td>
<td align="center">1,275</td>
<td align="right">~5.5M</td>
<td align="right"><strong>172K</strong> (-97%)</td>
<td align="center">✅ Claude/Gemini</td>
<td align="center">10,451 files/s</td>
<td align="center">🦀</td>
</tr>
<tr>
<td>🚀 <code>astro</code></td>
<td align="center">1,058</td>
<td align="right">~10.5M</td>
<td align="right"><strong>149K</strong> (-99%)</td>
<td align="center">✅ Claude/Gemini</td>
<td align="center">5,212 files/s</td>
<td align="center">📘</td>
</tr>
<tr>
<td>💎 <code>rails</code></td>
<td align="center">394</td>
<td align="right">~1M</td>
<td align="right"><strong>104K</strong> (-90%)</td>
<td align="center">✅ ChatGPT-4o</td>
<td align="center">4,864 files/s</td>
<td align="center">💎</td>
</tr>
<tr>
<td>🐘 <code>laravel</code></td>
<td align="center">1,443</td>
<td align="right">~3M</td>
<td align="right"><strong>238K</strong> (-92%)</td>
<td align="center">✅ Claude/ChatGPT-4o</td>
<td align="center">4,613 files/s</td>
<td align="center">🐘</td>
</tr>
<tr>
<td>⚡ <code>nestjs</code></td>
<td align="center">802</td>
<td align="right">~1.5M</td>
<td align="right"><strong>107K</strong> (-93%)</td>
<td align="center">✅ ChatGPT-4o</td>
<td align="center">8,813 files/s</td>
<td align="center">📘</td>
</tr>
<tr>
<td>👻 <code>ghost</code></td>
<td align="center">2,184</td>
<td align="right">~8M</td>
<td align="right"><strong>235K</strong> (-97%)</td>
<td align="center">✅ Claude/ChatGPT-4o</td>
<td align="center">4,719 files/s</td>
<td align="center">📜</td>
</tr>
</table>

<sub><sup>1</sup> Context windows: ChatGPT-4o (128K), Claude (200K), Gemini (2M). ✅ = fits completely, ⚠️ = needs splitting</sub><br>
<sub><sup>2</sup> Processing speed with 12 parallel workers on AMD Ryzen 7945HX. Use `-w 1` for serial mode or `-w N` for custom workers.</sub><br>
<sub><sup>3</sup> Token counts estimated using OpenAI's cl100k_base tokenizer (1 token ≈ 4 characters). Actual counts may vary by model.</sub>

### 🎯 Why This Matters for AI-Assisted Development

**Large codebases are overwhelming for AI models.** A typical web framework like Django has ~10 million tokens of source code. Even with Claude's 200K context window, you'd need to split it into 50+ chunks, losing coherence and relationships between components.

**But here's the good news**: Most real-world projects that teams have invested hundreds to thousands of hours developing are much smaller. Thanks to AI Distiller, the vast majority of typical business applications, SaaS products, and internal tools can fit entirely within AI context windows, enabling unprecedented AI assistance quality.

### ⚠️ The Hidden Problem with AI Coding Tools

**Most AI agents and IDEs are "context misers"** - they try to save tokens at the expense of actual codebase knowledge. They rely on:
- 🔍 **Grep/search** to find relevant code snippets
- 📄 **Limited context** showing only 10-50 lines around matches  
- 🎲 **Guessing interfaces** based on partial information

**This is why AI-generated code often fails on first attempts** - the AI is literally guessing method signatures, parameter types, and return values because it can't see the full picture.

**AI Distiller changes the game** by giving AI complete knowledge of:
- ✅ **Exact interfaces** of all classes, methods, and functions
- ✅ **All parameter types** and their expected values
- ✅ **Return types** and data structures
- ✅ **Full inheritance hierarchies** and relationships

Instead of playing "code roulette", AI can now write correct code from the start.

**Result**: Django's 10M tokens compress to just 256K tokens - suddenly the **entire framework fits in a single AI conversation**, leading to:
- 🎯 **More accurate suggestions** - AI sees all available APIs at once
- 🚀 **Less hallucination** - No more inventing methods that don't exist
- 💡 **Better architecture advice** - AI understands the full system design
- ⚡ **Faster development** - Especially for "vibe coding" with AI assistance
- 💰 **40x cost reduction** - Pay for 256K tokens instead of 10M on API calls

### 🔧 Flexible for Different Use Cases

```bash
# Process entire codebase (default: public APIs only)
aid ./my-project

# Process specific directories or modules
aid ./my-project/src/auth ./my-project/src/api
aid ./my-project/core/*.py ./my-project/utils/

# Process individual files
aid src/main.py src/config.py src/models/*.py

# Include protected/private for deeper analysis
aid ./my-project --private=1 --protected=1

# Include implementations for small projects
aid ./my-small-lib --implementation=1

# Everything for complete understanding
aid ./micro-service --private=1 --protected=1 --implementation=1
```

**Granular Control**: Process your entire codebase, specific modules, directories, or even individual files. Perfect for focusing AI on the exact context it needs - whether that's understanding the whole system architecture or diving deep into a specific authentication module.

📈 **[Full benchmark details](benchmark/BENCHMARK_RESULTS.md)** | 🧪 **[Reproduce these results](benchmark/test_benchmark.sh)**

## 🤔 How is AI Distiller Different?

<table>
<tr>
<th>Tool</th>
<th>What it does</th>
<th>AI Distiller advantage</th>
</tr>
<tr>
<td><strong>ctags/etags</strong></td>
<td>Generates index of names/locations</td>
<td>AI Distiller provides full signatures with types, parameters, and return values - exactly what LLMs need</td>
</tr>
<tr>
<td><strong>LSP servers</strong></td>
<td>Real-time IDE protocol for code intelligence</td>
<td>AI Distiller is a CLI tool for bulk, offline analysis creating a single context file optimized for LLMs</td>
</tr>
<tr>
<td><strong>GitHub Copilot</strong></td>
<td>Limited context window, guesses from nearby code</td>
<td>AI Distiller gives complete codebase overview, preventing hallucinated APIs</td>
</tr>
<tr>
<td><strong>Sourcegraph</strong></td>
<td>Code search and navigation platform</td>
<td>AI Distiller focuses on aggressive compression for AI context windows, not human browsing</td>
</tr>
</table>

## ✨ Key Features

### 🎯 Intelligent Filtering

Control exactly what to include with our new granular flag system:

**Visibility Control**:
- `--public=1` (default) - Include public members
- `--protected=0` (default) - Exclude protected members
- `--internal=0` (default) - Exclude internal/package-private
- `--private=0` (default) - Exclude private members

**Content Control**:
- `--comments=0` (default) - Exclude comments
- `--docstrings=1` (default) - Include documentation
- `--implementation=0` (default) - Exclude function bodies
- `--imports=1` (default) - Include import statements

**Default behavior**: Shows only public API signatures with documentation - perfect for AI understanding while maintaining maximum compression.

### 🤖 AI-Powered Analysis Workflows

Generate comprehensive code analysis task lists for AI assistants:

- **`--ai-analysis-task-list`** - Create structured workflows for security, performance, and maintainability audits
- **Multi-pattern filtering** - `--include "*.go,*.py"` or `--include "*.go" --include "*.py"`  
- **Smart scope control** - Focus on specific directories, file types, or exclude test files
- **Pre-generated infrastructure** - Ready-to-use directories, task lists, and summary tables
- **Color-coded results** - Visual HTML formatting for critical issues and excellent scores
- **4-dimensional scoring** - Security, Performance, Maintainability, Readability (0-100%)

Perfect for systematic code reviews, security audits, and onboarding new team members.

### 📝 Multiple Output Formats
- **Text** (`--format text`) - Ultra-compact for AI consumption (default)
- **Markdown** (`--format md`) - Human-readable with emojis
- **JSON** (`--format json`) - Structured data for tools
- **JSONL** (`--format jsonl`) - Streaming format
- **XML** (`--format xml`) - Legacy system compatible

### 📁 Smart Project Root Detection

AI Distiller automatically detects your project root and centralizes all outputs in a `.aid/` directory:

- **Automatic detection**: Searches upward for `.aidrc`, `go.mod`, `package.json`, `.git`, etc.
- **Consistent location**: All outputs go to `<project-root>/.aid/` regardless of where you run `aid`
- **Cache management**: MCP cache stored in `.aid/cache/` for better organization
- **Easy cleanup**: Add `.aid/` to `.gitignore` to keep outputs out of version control

**Detection priority:**
1. **`.aidrc` file** - Create this empty file to explicitly mark your project root
2. **Language markers** - `go.mod`, `package.json`, `pyproject.toml`, etc.
3. **Version control** - `.git` directory
4. **Environment variable** - `AID_PROJECT_ROOT` (fallback if no markers found)
5. **Current directory** - Final fallback with warning

```bash
# Mark a specific directory as project root (recommended)
touch /my/project/.aidrc

# Run from anywhere in your project - outputs always go to project root
cd deep/nested/directory
aid ../../../src  # Output: <project-root>/.aid/aid.src.txt

# Use environment variable as fallback (useful for CI/CD)
AID_PROJECT_ROOT=/build/workspace aid src/
```

### 🌍 Language Support
Currently supports 12+ languages via tree-sitter:
- **Full Support**: Python, TypeScript, JavaScript, Go, Java, C#, Rust
- **Beta**: Ruby, Swift, Kotlin, PHP, C++
- **Coming Soon**: Zig, Scala, Clojure

#### Language-Specific Documentation:
- [Python](docs/lang/python.md) - Full Python 3.x support with type hints, async/await, decorators
- [C#](docs/lang/csharp.md) - Complete C# 12 support with records, nullable reference types, pattern matching
- [PHP](docs/lang/php.md) - PHP 7.4+ with PHP 8.x features (attributes, union types, enums)
- [Swift](docs/lang/swift.md) - Swift 5.x support with tree-sitter parser (basic features implemented)
- More language guides coming soon!

See [all language documentation](docs/lang/) for complete details.

## ⚠️ Limitations

- **Syntax Errors**: Files with syntax errors may be skipped or partially processed
- **Dynamic Features**: Runtime-determined types/interfaces in dynamic languages are not resolved
- **Macro Expansion**: Complex macros (Rust, C++) show pre-expansion source
- **Generated Code**: Consider using `.aidignore` to skip generated files
- **Large Files**: Files over 10MB are processed in chunks for memory efficiency

## 🔒 Security Considerations

**⚠️ Important**: AI Distiller extracts code structure which may include:
- Function and variable names that could reveal business logic (e.g., `processPayment`, `calculateTaxEvasion`)
- API endpoints and internal routes (e.g., `/api/v1/internal/user-data`)
- Type information and data structures
- Comments and docstrings (unless stripped)
- File paths revealing project structure or codenames

**Recommendations**:
1. Always review output before sending to external services
2. Use `--comments=0` to remove potentially sensitive documentation
3. Consider running a secrets scanner on your codebase first
4. For maximum security, run AI Distiller in an isolated environment
5. Future: We're exploring an `--obfuscate` flag to anonymize sensitive identifiers

## 📖 Example Output

<details>
<summary>Python Class Example</summary>

**Input** (`car.py`):
```python
class Car:
    """A car with basic attributes and methods."""
    
    def __init__(self, make: str, model: str):
        self.make = make
        self.model = model
        self._mileage = 0  # Private
    
    def drive(self, distance: int) -> None:
        """Drive the car."""
        if distance > 0:
            self._mileage += distance
```

**Output** (`aid car.py --format text --implementation=0`):
```
<file path="car.py">
class Car:
    +def __init__(self, make: str, model: str)
    +def drive(self, distance: int) -> None
</file>
```

</details>

<details>
<summary>TypeScript Interface Example</summary>

**Input** (`api.ts`):
```typescript
export interface User {
  id: number;
  name: string;
  email?: string;
}

export class UserService {
  private cache = new Map<number, User>();
  
  async getUser(id: number): Promise<User | null> {
    return this.cache.get(id) || null;
  }
}
```

**Output** (`aid api.ts --format text --implementation=0`):
```
<file path="api.ts">
export interface User {
  id: number;
  name: string;
  email?: string;
}

export class UserService {
  +async getUser(id: number): Promise<User | null>
}
</file>
```

</details>

## 🤖 AI-Driven Code Analysis Workflow

AI Distiller now includes a powerful workflow for comprehensive codebase analysis using AI assistants like Claude or ChatGPT. Generate structured task lists and detailed security/performance audits automatically.

### 🎯 Generate Analysis Task Lists

Create comprehensive analysis workflows for any codebase:

```bash
# Generate complete project analysis task list
aid --ai-analysis-task-list

# Analyze specific directories only  
aid internal/api --ai-analysis-task-list

# Focus on specific file types
aid --include "*.go,*.py,*.ts" --ai-analysis-task-list

# Exclude test files and configs
aid --exclude "*test*,*.json,*.yml" --ai-analysis-task-list
```

**What gets generated**:
- 📋 **Task list** with checkboxes for each file (`.aid/ANALYSIS-TASK-LIST.PROJECT.DATE.md`)
- 📊 **Summary table** for collecting results (`.aid/ANALYSIS-SUMMARY.PROJECT.DATE.md`)
- 📁 **Pre-created directories** for individual file reports
- 🎨 **Color-coded output** with HTML styling for critical issues

### 🔍 Comprehensive Analysis Coverage

Each file gets analyzed across **4 dimensions**:

| Dimension | Focus Areas |
|-----------|-------------|
| **🛡️ Security** | Vulnerabilities, exposed secrets, dangerous patterns |
| **⚡ Performance** | Efficiency, scalability concerns, resource usage |
| **🔧 Maintainability** | Code complexity, structure, documentation quality |
| **📖 Readability** | Code clarity, naming, organization, comments |

**Scoring system**: 0-100% with specific point deductions:
- Critical issues: -30 points
- High issues: -15 points  
- Medium issues: -5 points
- Low issues: -2 points

### 🎨 Visual Results with Color Coding

The analysis generates beautiful reports with automatic color coding:

```markdown
| File | Security % | Performance % | Critical | High |
|------|:----------:|:-------------:|:--------:|:----:|
| auth.go | 45 | 78 | 2 | 1 |
| cache.go | 92 | 65 | 0 | 0 |
```

Becomes:
- **Critical scores** (<50%): <span style="color:#ff0000; font-weight: bold">45</span>
- **High issues**: <span style="color:#ff6600; font-weight: bold">2</span>
- **Excellent scores** (90%+): <span style="color:#00aa00; font-weight: bold">92</span>

### 🔧 Flexible Scope Control

Control exactly what gets analyzed using multiple pattern syntaxes:

```bash
# Comma-separated patterns
aid --include "*.go,*.py,*.ts" --exclude "*test*,*spec*" --ai-analysis-task-list

# Multiple flags (same result)
aid --include "*.go" --include "*.py" --exclude "*test*" --ai-analysis-task-list

# Language-specific analysis
aid --include "*.vue,*.svelte" --ai-analysis-task-list  # Frontend components
aid --include "*.twig,*.latte,*.j2" --ai-analysis-task-list  # Templates
aid --exclude "*.json,*.yaml,*.env" --ai-analysis-task-list  # Skip configs
```

### 📈 Example Workflow

1. **Generate task list**:
   ```bash
   aid src/ --exclude "*test*" --ai-analysis-task-list
   ```

2. **Follow the generated instructions** in `.aid/ANALYSIS-TASK-LIST.PROJECT.DATE.md`

3. **Get structured results** with:
   - Individual detailed reports for each file
   - Centralized summary table with scores  
   - Color-coded visualization of critical issues
   - Final project-level conclusions and recommendations

### 🎯 Perfect for AI Assistants

The generated workflow is designed to work seamlessly with:
- **Claude Code** - Direct file analysis and report generation
- **ChatGPT/Claude Web** - Copy-paste friendly format  
- **Custom AI tools** - Structured JSON/markdown output
- **Code review processes** - Comprehensive audit trails

**Pro tip**: Use `aid internal/ --include "*.go" --ai-analysis-task-list` to focus analysis on your core business logic and skip test files for faster results.

## 🛠️ Advanced Usage

### ⚡ Parallel Processing

AI Distiller now supports parallel processing for significantly faster analysis of large codebases:

```bash
# Use default parallel processing (80% of CPU cores)
aid ./src

# Force serial processing (original behavior)
aid ./src -w 1

# Use specific number of workers
aid ./src -w 16

# Check performance with verbose output
aid ./src -v  # Shows: "Using 25 parallel workers (32 CPU cores available)"
```

**Performance Benefits**:
- React packages: 3.5s → 0.5s (7x faster)
- Large codebases: Near-linear speedup with CPU cores
- Maintains identical output order as serial processing

### Processing from stdin

AI Distiller can process code directly from stdin, perfect for:
- Quick code snippet analysis
- Pipeline integration
- Testing without creating files
- Dynamic code generation workflows

```bash
# Auto-detect language from stdin
echo 'class User { getName() { return this.name; } }' | aid --format text

# Explicit language specification
cat mycode.php | aid --lang php --private=0 --protected=0

# Use "-" to explicitly read from stdin
aid - --lang python < snippet.py

# Pipeline example: extract structure from generated code
generate-code.sh | aid --lang typescript --format json
```

**Language Detection**: When using stdin without `--lang`, AI Distiller automatically detects the language based on syntax patterns. Supported languages for auto-detection: python, typescript, javascript, go, ruby, swift, rust, java, c#, kotlin, c++, php.

### Integration with AI Tools

```bash
# Create a context file for Claude or GPT
aid ./src --format text --implementation=0 > context.txt

# Generate a codebase summary for RAG systems
aid . --format json | jq -r '.files[].symbols[].name' > symbols.txt

# Extract API surface for documentation
aid ./api --comments=0 --implementation=0 --format md > api-ref.md
```

### MCP Server Mode

AI Distiller can run as an MCP server, providing codebase analysis capabilities to AI agents:

```bash
# Start as MCP server
aid --mcp-server

# With specific root directory
aid --mcp-server --root /path/to/project
```

See [MCP Integration Guide](docs/mcp-integration.md) for detailed configuration and usage.

### Configuration File

Create `.aidconfig.yml` in your project root:

```yaml
# Default options for this project
format: text
strip:
  - implementation
  - non-public
exclude:
  - "**/*.test.js"
  - "**/node_modules/**"
  - "**/__pycache__/**"
```

## 🎯 Git History Analysis Mode

AI Distiller includes a special mode for analyzing git repositories. When you pass a `.git` directory, it switches to git log mode:

```bash
# View formatted git history
aid .git

# Limit to recent commits (default is 200)
aid .git --git-limit=100

# Include AI analysis prompt for comprehensive insights
aid .git --with-analysis-prompt
```

The `--with-analysis-prompt` flag adds a sophisticated prompt that guides AI to generate:
- **Contributor statistics** with expertise areas and collaboration patterns
- **Timeline analysis** with development phases and activity visualization
- **Functional categorization** of commits (features, fixes, refactoring)
- **Codebase evolution insights** including technology shifts
- **Actionable recommendations** based on discovered patterns

Perfect for understanding project history, identifying knowledge silos, or generating impressive development reports.

## 🤖 AI-Driven Code Analysis Workflow

AI Distiller includes a revolutionary feature for comprehensive codebase analysis. Generate structured task lists that guide AI assistants through systematic file-by-file analysis:

```bash
# Generate comprehensive analysis task list
aid --ai-analysis-task-list ./my-project
```

This creates:
- **📋 Task List**: Structured checklist with AI analysis instructions
- **📊 Summary Table**: Centralized results with security, performance, and maintainability scores
- **📁 Analysis Reports**: Individual detailed reports for each file
- **🎯 Project Conclusion**: Synthesized findings and recommendations

**Perfect for**:
- Security audits and vulnerability assessments
- Code quality reviews and technical debt analysis  
- Onboarding new team members to complex codebases
- Pre-deployment health checks
- AI-assisted refactoring planning

**Workflow**: AI assistants like Claude Code follow the generated task list, analyzing each file systematically, scoring security/performance/maintainability, and building a comprehensive project health dashboard. The result? Professional-grade analysis reports that would typically require senior engineers weeks to produce.

## 🔗 Documentation

- [Installation Guide](docs/installation.md)
- [CLI Reference](docs/cli-reference.md)
- [Project Root Detection](docs/user/project-root-detection.md) - How `.aid/` directory location is determined 🆕
- [MCP Integration Guide](docs/mcp-integration.md)
- [Language Support](docs/lang/)
  - [Python](docs/lang/python.md)
  - [C#](docs/lang/csharp.md)
  - [TypeScript](docs/lang/typescript.md)
  - [Go](docs/lang/go.md)
  - [JavaScript](docs/lang/javascript.md)
  - [PHP](docs/lang/php.md)
  - [More...](docs/lang/)
- [Output Formats](docs/formats.md)
- [Performance Tuning](docs/performance.md)
- [Security Guide](docs/security.md)

## 🐛 Debugging

AI Distiller includes a comprehensive 3-level debugging system for troubleshooting and understanding the parsing process:

### Debug Levels

```bash
# Level 1 (-v): Basic information
aid main.go -v
# Shows: file counts, phase transitions, configuration

# Level 2 (-vv): Detailed processing info  
aid main.go -vv
# Shows: individual file processing, parser selection, timing

# Level 3 (-vvv): Full data structure dumps
aid main.go -vvv
# Shows: complete AST/IR structures, before/after stripping
```

### Debug Output Features

- **Subsystem prefixes**: `[processor]`, `[python:tree-sitter]`, `[golang:ast]`
- **Automatic timing**: Operation durations for performance analysis
- **Data structure dumps**: Full AST and IR representations (like PHP's `print_r`)
- **Before/after comparisons**: See how stripping transforms the IR

### Example Debug Session

```bash
# Debug Python parsing with full traces
echo 'class Example: pass' | aid -vvv --lang python

# Output includes:
# - Raw tree-sitter AST structure
# - Initial IR before stripping  
# - Final IR after stripping
# - Formatting phase details
```

### Performance Profiling

Debug output includes timing for each phase:
- Parser initialization
- AST parsing
- IR generation
- Stripping application
- Output formatting

This helps identify bottlenecks in processing large codebases.

## ❓ FAQ

<details>
<summary><strong>How accurate are the token counts?</strong></summary>

Token counts are estimated using OpenAI's cl100k_base tokenizer (1 token ≈ 4 characters). Actual token usage varies by model - Claude and GPT-4 use similar tokenizers, while others may differ by ±10%.
</details>

<details>
<summary><strong>Can AI Distiller handle very large repositories?</strong></summary>

Yes! We've tested on repositories with 50,000+ files. The parallel processing mode (`-w` flag) scales linearly with CPU cores. Memory usage is bounded - large files are processed in streaming chunks.
</details>

<details>
<summary><strong>What about generated code and vendor directories?</strong></summary>

Create a `.aidignore` file (same syntax as `.gitignore`) to exclude generated files, vendor directories, or any paths you don't want processed.
</details>

<details>
<summary><strong>What happens with unsupported file types?</strong></summary>

Files with unknown or unsupported extensions are automatically skipped - no errors, no interruption. AI Distiller only processes files it has parsers for, ensuring clean and relevant output. This means you can safely run it on mixed repositories containing documentation, images, configs, etc.
</details>

<details>
<summary><strong>Is my code sent anywhere?</strong></summary>

No! AI Distiller runs 100% locally. It only extracts and formats your code structure - you decide what to do with the output. The tool itself makes no network connections.
</details>

<details>
<summary><strong>Which programming languages are supported?</strong></summary>

Currently 12+ languages via tree-sitter: Python, TypeScript, JavaScript, Go, Java, C#, Rust, Ruby, Swift, Kotlin, PHP, C++. All parsers are bundled in the binary - no external dependencies needed.
</details>

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone and setup
git clone https://github.com/janreges/ai-distiller
cd ai-distiller
make dev-init    # Initialize development environment

# Run tests
make test         # Unit tests
make test-integration  # Integration tests

# Build binary
make build        # Build for current platform
make cross-compile  # Build for all platforms
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built on [tree-sitter](https://tree-sitter.github.io/) for accurate parsing
- Inspired by the need for better AI-code interaction
- Created with ❤️ for the AI engineering community

---

<p align="center">
  <sub>Built by developers, for developers working with AI</sub>
</p>