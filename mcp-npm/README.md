# AI Distiller (aid) MCP Server

Advanced MCP (Model Context Protocol) server for [AI Distiller](https://github.com/janreges/ai-distiller) - Essential code structure extractor that provides LLMs with accurate code signatures, data types, and API contracts from your actual codebase. Reduces AI guesswork and trial-error coding by delivering precise structural information. Accelerates analysis workflows including security audits, performance reviews, git history insights, refactoring suggestions, and comprehensive structural analysis.

## 🚀 Quick Start with Claude Code

### Option 1: Project-specific installation (Recommended)
```bash
# Install for current project only
claude mcp add aid -- npx -y @janreges/ai-distiller-mcp
```

### Option 2: User-wide installation with manual configuration
```bash
# Install globally for all projects
claude mcp add --scope=user aid -- npx -y @janreges/ai-distiller-mcp
```

**⚠️ Important for user-wide installation:**
You must set `AID_ROOT` environment variable in your Claude Code configuration to point to your current project directory. Without this, the aid tool won't know which directory to analyze.

Example configuration:
```json
{
  "mcpServers": {
    "aid": {
      "command": "npx",
      "args": ["-y", "@janreges/ai-distiller-mcp"],
      "env": {
        "AID_ROOT": "/absolute/path/to/your/project"
      }
    }
  }
}
```

## ✨ Features

### 🔑 How AI Distiller Works

AI Distiller (aid) **generates AI prompts with distilled code** - it doesn't analyze code directly. Instead:
1. **aid extracts code structure** (distillation)
2. **Generates specialized AI prompts** for your analysis goal
3. **Outputs to `.aid/` directory** or stdout
4. **AI agents execute the prompts** to perform actual analysis

### 📋 Typical Workflow

1. **User asks**: "Find bugs in my authentication module"
2. **Claude calls**: `aid_hunt_bugs({ target_path: "src/auth/" })`
3. **aid generates**: Bug-hunting prompt + distilled code → `.aid/bug-hunting.md`
4. **Claude reads**: The generated file and executes the analysis
5. **Result**: Actual bug findings and recommendations

### 💾 Output Formats

- **Small analyses**: Return directly via stdout
- **Large analyses**: Save to `.aid/` directory as markdown files
- **File naming**: `.aid/ACTION.TIMESTAMP.FOLDER.md`
- **Content**: AI prompt + distilled code in one file

### 🎯 Specialized AI Analysis Tools (New!)
- **aid_hunt_bugs** - Generates bug-hunting prompts with distilled code
- **aid_suggest_refactoring** - Creates refactoring analysis prompts
- **aid_generate_diagram** - Produces prompts for architectural diagrams
- **aid_analyze_security** - Generates security audit prompts
- **aid_generate_docs** - Creates documentation generation prompts

### Core Analysis Engine
- **aid_analyze** - Direct access to all AI actions for custom workflows
- Supports 10 different AI analysis types
- Outputs to `.aid/` directory for large analyses

### Code Structure Tools
- **distill_file** - Extract code structure from a single file
- **distill_directory** - Extract structure from entire directories
- Control visibility: public, protected, internal, private
- Control detail: with or without implementation

## Configuration

### Environment Variables

- `AID_ROOT` - **REQUIRED for user-scoped installation**: Root directory for analysis. Without this, the tool cannot determine which project to analyze.

### Working Directory Behavior

⚠️ **Critical Information:**
- **Project-scoped installation** (without `--scope=user`): Automatically uses the current project directory
- **User-scoped installation** (with `--scope=user`): MUST set `AID_ROOT` to your project path

### Claude Desktop Configuration

For user-scoped installations, you MUST configure the project path:

```json
{
  "mcpServers": {
    "aid": {
      "command": "npx",
      "args": ["-y", "@janreges/ai-distiller-mcp"],
      "env": {
        "AID_ROOT": "/absolute/path/to/your/project"  // REQUIRED!
      }
    }
  }
}
```

## 🎛️ Controlling Output Size

### Visibility Levels
Control what code elements are included:
- `include_public: true/false` - Public members (default: true)
- `include_protected: true/false` - Protected members (default: false)
- `include_internal: true/false` - Internal/package-private (default: false)
- `include_private: true/false` - Private members (default: false)

### Implementation Control
- `include_implementation: false` - Only signatures (smallest output)
- `include_implementation: true` - Full method/function bodies (largest output)

### Output Size Examples

| Configuration | Output Size | Use Case |
|--------------|-------------|----------|
| Public only, no implementation | Smallest | API documentation |
| All visibility, no implementation | Medium | Architecture overview |
| Public + implementation | Large | Detailed API analysis |
| All visibility + implementation | Largest | Deep code analysis |

### Working with Large Codebases

For large projects, aid's output may exceed AI context limits. Strategies:

1. **Target specific directories**:
   ```javascript
   aid_analyze({
     target_path: "src/auth/",  // Just auth module
     ai_action: "prompt-for-bug-hunting"
   })
   ```

2. **Use file patterns**:
   ```javascript
   aid_analyze({
     target_path: "src/",
     include_patterns: "*.py",  // Python files only
     exclude_patterns: "*test*,*mock*"
   })
   ```

3. **Progressive analysis**:
   - Start with structure only (`include_implementation: false`)
   - Then analyze specific modules with full detail

4. **Output to files**:
   - Large analyses go to `.aid/` directory
   - Can be read by AI agents or copied to tools like Gemini (1M context)

## 📖 Tool Examples

### 🎯 Specialized Tools (Recommended)

#### Generate Bug-Hunting Prompt
```javascript
// Claude will call:
aid_hunt_bugs({
  target_path: "src/",
  focus_area: "memory leaks and race conditions",
  include_private: true,
  include_implementation: true  // Include code bodies
})
// Returns: AI prompt with distilled code for bug analysis
// Output: .aid/bug-hunting-2024-06-20.src.md
```

#### Generate Refactoring Prompt
```javascript
// Claude will call:
aid_suggest_refactoring({
  target_path: "src/auth/",
  refactoring_goal: "improve testability",
  include_implementation: false  // Signatures only for overview
})
// Returns: AI prompt for refactoring suggestions
// Claude can then execute this prompt
```

#### Generate Diagram Creation Prompt
```javascript
// Claude will call:
aid_generate_diagram({
  target_path: "src/",
  diagram_focus: "authentication flow",
  include_private: true,
  include_implementation: false  // Structure only
})
// Returns: AI prompt to create 10 architectural diagrams
```

#### Generate Security Analysis Prompt
```javascript
// Claude will call:
aid_analyze_security({
  target_path: "src/api/",
  security_focus: "input validation and SQL injection",
  include_private: true,
  include_implementation: true  // Need to see actual code
})
// Returns: AI prompt for security vulnerability analysis
```

#### Generate Documentation Prompt
```javascript
// Claude will call:
aid_generate_docs({
  target_path: "src/core/",
  doc_type: "api-reference",
  audience: "external developers",
  include_private: false  // Public API only
})
// Returns: AI prompt to generate comprehensive documentation
```

### Core Analysis Engine

#### Custom AI Analysis Workflows
```javascript
// For advanced users who need specific AI actions:
aid_analyze({
  ai_action: "flow-for-deep-file-to-file-analysis",
  target_path: "src/",
  user_query: "Focus on authentication and authorization patterns",
  include_patterns: "*.py,*.js"
})
```

Available AI actions:
- `flow-for-deep-file-to-file-analysis` - Systematic file-by-file analysis workflow
- `flow-for-multi-file-docs` - Multi-file documentation generation workflow
- `prompt-for-refactoring-suggestion` - Refactoring analysis prompt
- `prompt-for-complex-codebase-analysis` - Enterprise-grade analysis prompt
- `prompt-for-security-analysis` - Security audit prompt
- `prompt-for-performance-analysis` - Performance optimization prompt
- `prompt-for-best-practices-analysis` - Best practices evaluation prompt
- `prompt-for-bug-hunting` - Bug detection prompt
- `prompt-for-single-file-docs` - Single file documentation prompt
- `prompt-for-diagrams` - Diagram generation prompt

## 🤖 AI Agent Workflows

### Quick Bug Hunt
```
User: "Check this codebase for potential bugs"
1. Claude uses aid_hunt_bugs(target_path="src/", include_private=true)
2. Analyzes the structured bug report
3. Provides prioritized fixes
```

### Security Audit
```
User: "Perform a security audit on the API endpoints"
1. Claude uses aid_analyze_security(target_path="src/api/", include_implementation=true)
2. Reviews OWASP-categorized findings
3. Suggests security improvements
```

### Architecture Understanding
```
User: "Help me understand this codebase architecture"
1. Claude uses aid_generate_diagram(target_path="src/")
2. Presents 10 different architectural views
3. Explains key components and relationships
```

### Refactoring Session
```
User: "This module is getting complex, suggest improvements"
1. Claude uses aid_suggest_refactoring(target_path="module/", refactoring_goal="reduce complexity")
2. Provides specific refactoring suggestions
3. Shows before/after code examples
```

## 🚀 Advanced Features

- **Pattern Filtering** - Include/exclude specific file patterns
- **Visibility Control** - Include/exclude private, protected, internal members
- **Implementation Details** - Optionally include function bodies
- **Multi-format Output** - text, markdown, JSON formats
- **Caching** - Intelligent caching for repeated analyses
- **Direct CLI Integration** - Each tool directly calls the aid binary

## 📚 Supported Languages

Python, TypeScript, JavaScript, Go, Java, C#, Rust, Ruby, Swift, Kotlin, PHP, C++, C

## 🔧 Troubleshooting

If you encounter issues:

1. **Check aid binary**: Ensure `aid` is in your PATH or build directory
2. **Verify permissions**: The server needs read access to your project
3. **Check logs**: Run with verbose logging to see detailed errors
4. **File patterns**: Use comma-separated patterns like `*.go,*.py`

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

*AI Distiller (aid) - https://github.com/janreges/ai-distiller*  
*Authored by [Claude Code](https://claude.ai/code) & [Ján Regeš](https://github.com/janreges) from [SiteOne](https://www.siteone.io/) (Czech Republic)*  
*Explore the project on [GitHub](https://github.com/janreges/ai-distiller)*