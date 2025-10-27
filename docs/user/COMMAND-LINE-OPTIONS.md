# AI Distiller - Command Line Options Reference

AI Distiller transforms source code into optimized formats for Large Language Models. It compresses codebases by 60-90% while preserving all semantic information needed for AI analysis. The tool generates either compressed code representations or complete AI analysis prompts that can be directly copied to AI tools like Gemini 2.5 Pro (1M context), ChatGPT-o3/4o, or Claude for perfect AI-powered code analysis, refactoring, security audits, and architectural reviews.

## Primary Options

### Input/Output Control

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `<path>` | string | current dir | Relative or absolute path to source directory or file to analyze |
| `-o, --output FILE` | string | .aid/ folder or .aid.*.txt | Write output to specific file instead of auto-generated name |
| `--stdout` | flag | false | Print output to stdout (in addition to file output) |
| `--format FORMAT` | string | text | Output format: `text`, `md`, `jsonl`, `json-structured`, `xml` |

### AI Actions System

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--ai-action ACTION` | string | none | Use predefined AI action configuration (see AI Actions section) |
| `--ai-output FILE` | string | auto-generated | Custom output path for AI action (default: action-specific directory/file, supports template variables) |

#### Available AI Actions

1. **prompt-for-refactoring-suggestion**
   - Generates comprehensive refactoring analysis prompts
   - Includes context awareness, effort sizing, validation steps
   - Output optimized for GPT-4, Claude, Gemini

2. **prompt-for-complex-codebase-analysis**
   - Creates enterprise-grade codebase analysis prompts
   - Architecture diagrams, compliance sections, evidence-based findings
   - Coverage gaps and limitation acknowledgments

3. **prompt-for-security-analysis**
   - Generates security audit prompts with OWASP Top 10 focus
   - Static vs dynamic analysis boundaries, evidence checklists
   - SARIF output integration for CI/CD pipelines

4. **prompt-for-performance-analysis**
   - Creates performance optimization analysis prompts
   - Static analysis constraints, profiling guidance
   - Enterprise scalability considerations

5. **flow-for-deep-file-to-file-analysis**
   - Generates structured task list for comprehensive file-by-file analysis
   - Creates .aid/ directory with analysis infrastructure and templates
   - Ensures consistent analysis methodology across all files

#### Using --raw with AI Actions

Adding `--raw` flag to AI actions includes full source code bodies in the generated prompts for comprehensive analysis.

**Context Size Considerations:**
- **Large codebases**: Analyze specific parts/folders that fit in AI context, or use default filtering (public APIs only, no implementation/comments) which may be insufficient for some analysis types but fits in smaller contexts
- **Small codebases**: Use `--raw` for full source code analysis
- **Recommended**: Gemini 2.5 Pro with 1M context window for largest codebase capacity

**Examples:**
```bash
# Full source analysis for small modules
aid ./small-module --ai-action prompt-for-security-analysis --raw

# Comprehensive refactoring analysis with full code
aid ./src --ai-action prompt-for-refactoring-suggestion --raw
```

## Filtering Options

### Visibility Control (Include/Exclude by Access Level)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--public 0\|1` | bool | 1 | Include public members (exported functions, public classes) |
| `--protected 0\|1` | bool | 0 | Include protected members (protected methods, _underscore conventions) |
| `--internal 0\|1` | bool | 0 | Include internal/package-private members (lowercase Go exports, package-private Java) |
| `--private 0\|1` | bool | 0 | Include private members (private fields, __dunder__ methods, #private JS) |

### Content Control (What Parts to Include)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--comments 0\|1` | bool | 0 | Include regular comments (// # /* */) |
| `--docstrings 0\|1` | bool | 1 | Include documentation comments (JSDoc, Python docstrings, Go package docs) |
| `--implementation 0\|1` | bool | 0 | Include function/method bodies and implementation details |
| `--imports 0\|1` | bool | 1 | Include import/require/using statements |
| `--annotations 0\|1` | bool | 1 | Include decorators/annotations (@property, @Override, [Serializable]) |

### Alternative Filtering Syntax

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--include-only CATEGORIES` | string | none | Include ONLY specified categories (comma-separated) |
| `--exclude-items CATEGORIES` | string | none | Exclude specified categories (comma-separated) |

**Valid categories:** `public`, `protected`, `internal`, `private`, `comments`, `docstrings`, `implementation`, `imports`, `annotations`

### File Pattern Filtering

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--include PATTERNS` | string array | none | Include files matching patterns (e.g., "*.py,*.go") |
| `--exclude PATTERNS` | string array | none | Exclude files matching patterns (e.g., "*test*,*.json") |

**Pattern Examples:**
- `*.ext` - Files with specific extension
- `**/pattern` - Recursive directory matching
- `dir/*` - Files in specific directory
- `*test*` - Files containing "test"

## Processing Options

### Language & Parsing Control

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--lang LANGUAGE` | string | auto | Override automatic language detection |
| `--raw` | flag | false | Process all text files without parsing (overrides all content filters, full file content) |
| `--tree-sitter` | flag | false | Use tree-sitter parser (experimental, more accurate) |
| `-r, --recursive 0\|1` | bool | 1 | Process directories recursively |

**Supported Languages:** `auto`, `python`, `typescript`, `javascript`, `go`, `ruby`, `swift`, `rust`, `java`, `csharp`, `kotlin`, `cpp`, `php`

**Note:** `--lang` is particularly useful when sending code via stdin (where only file content is provided without filename context for automatic detection).

### Path Control

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--file-path-type TYPE` | string | relative | Path format in output: `relative`, `absolute` |
| `--relative-path-prefix STR` | string | none | Custom prefix for relative paths in output |
| `-w, --workers NUM` | int | 0 | Number of parallel workers (0=auto/80% CPU cores, 1=serial) |

## Git Mode (Special Mode)

Activated automatically when `<path>` is `.git`

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--git-limit NUM` | int | 200 | Limit number of commits (0=all commits) |
| `--with-analysis-prompt` | flag | false | Prepend comprehensive AI analysis prompt to git output |

**Git Mode Output Format:**
```
[commit_hash] YYYY-MM-DD HH:MM:SS | author_name | commit_subject
    commit_body_line_1
    commit_body_line_2
    (properly indented)
```

## Diagnostics & Debugging

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `-v, --verbose` | flag | false | Verbose output (use -vv, -vvv for more detail) |
| `--strict` | flag | false | Fail on first syntax error instead of continuing |
| `--version` | flag | false | Show version information and exit |

**Verbosity Levels:**
- `-v` (Level 1): Basic info, file counts, phase transitions
- `-vv` (Level 2): Detailed info, individual file processing, timing
- `-vvv` (Level 3): Full trace with data dumps, IR structures

## Help & Documentation

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--help` | flag | false | Show basic help (organized, essential options only) |
| `--help-extended` | flag | false | Show complete documentation (man page style) |
| `--cheat` | flag | false | Show quick reference card |

### Topical Help Commands

| Command | Description |
|---------|-------------|
| `aid help actions` | Detailed AI actions documentation with examples |
| `aid help filtering` | Complete filtering reference with practical combinations |
| `aid help git` | Git mode documentation and workflow examples |

## Output File Naming & Management

AI Distiller uses intelligent file naming strategies to avoid conflicts and enable easy project management.

### Why Automatic File Naming?

**Problem:** Large codebases can generate megabytes of text output that exceed AI tool context limits.
**Solution:** Automatic file generation with recognizable patterns instead of requiring manual file specification.

**Benefits:**
- `.aid` prefix ensures easy recognition in `git status`
- Can be easily added to `.gitignore` if desired
- Avoids requiring users to specify output files for every operation
- Prevents accidental context overflow in AI tools

### Default Output File Patterns

#### Regular Distillation Files
Pattern: `.aid.<basename>.<options>.txt`

**Examples:**
- `.aid.myproject.pub.txt` (public only, default)
- `.aid.myproject.priv.prot.impl.txt` (private + protected + implementation)
- `.aid.myproject.nocomm.txt` (comments excluded)

#### AI Action Output Files
Pattern: `./.aid/<ACTION-NAME>.YYYY-MM-DD.HH-MM-SS.<basename>.md`

**Examples:**
- `./.aid/REFACTORING-SUGGESTION.2025-06-17.14-30-00.myproject.md`
- `./.aid/SECURITY-ANALYSIS.2025-06-17.14-30-00.myproject.md`

#### Flow Action Output (Multiple Files)
Pattern: `./.aid/` directory with subdirectories and multiple markdown files

**Example:** `flow-for-deep-file-to-file-analysis` creates:
- `./.aid/ANALYSIS-TASK-LIST.md` - Sequential task list
- `./.aid/ANALYSIS-SUMMARY.md` - Analysis template
- `./.aid/analysis/` - Directory for individual file analysis reports

### Using --stdout

The `--stdout` flag works with all operations but should be used carefully:

**Safe for small projects:**
```bash
aid ./small-module --stdout | pbcopy  # Copy to clipboard
echo "code" | aid --lang python --stdout  # Stdin processing
```

**Risky for large codebases:**
```bash
aid ./entire-project --stdout  # May generate MBs of text exceeding AI context
```

**Flow actions behavior:**
- Shows summary output to stdout
- Full directory structure still created in `.aid/`

### Git Integration

Add AI Distiller outputs to your `.gitignore` if desired:

```gitignore
# AI Distiller outputs
.aid/
.aid.*.txt
```

**Or commit them** for team collaboration and CI/CD integration:
- Share AI analysis results with team members
- Version control refactoring recommendations
- Track security analysis over time

### Template Variables (for --ai-output)
- `%YYYY-MM-DD%` - Current date
- `%HH-MM-SS%` - Current time
- `%folder-basename%` - Directory name

## Common Usage Patterns

### Basic Distillation
```bash
aid ./src                           # Public APIs only (no implementation/comments)
aid ./src --implementation=1        # Include function bodies
aid ./ --private=1 --protected=1    # Include all visibility levels
```

### AI-Powered Analysis
```bash
aid ./src --ai-action prompt-for-refactoring-suggestion
aid ./ --ai-action prompt-for-security-analysis --private=1
aid ./core --ai-action prompt-for-performance-analysis
```

### Output Control
```bash
aid ./src --format=md -o analysis.md
aid ./src --stdout | pbcopy          # Copy to clipboard (macOS)
aid ./src --format=jsonl > data.jsonl
```

### Filtering Examples
```bash
aid ./ --include "*.go,*.py"          # Only Go and Python files
aid ./ --exclude "*test*,*spec*"      # Exclude test files
aid ./ --include-only public,imports  # Only public APIs and imports
aid ./ --exclude-items comments,implementation
```

### Git Analysis
```bash
aid .git --git-limit=100             # Last 100 commits
aid .git --with-analysis-prompt      # With AI analysis guidance
```

### Advanced Usage
```bash
aid ./mixed-repo --lang=python       # Force Python parsing
aid ./docs --raw                     # Process as plain text (full content, no filters)
aid ./large-project -w 1            # Single-threaded processing
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (file not found, parse error, etc.) |
| 2 | Invalid arguments or conflicting options |

## Legacy Options (Deprecated)

| Option | Type | Status | Replacement |
|--------|------|--------|-------------|
| `--strip <items>` | string | Deprecated | Use individual flags: `--private=0`, `--implementation=0`, etc. |

**Legacy `--strip` values:** `comments`, `imports`, `implementation`, `non-public`, `private`, `protected`

---

*This documentation is automatically loaded by `aid --help-extended` for comprehensive reference.*