# AI Distiller MCP Server

Advanced MCP (Model Context Protocol) server for [AI Distiller](https://github.com/janreges/ai-distiller) - a high-performance tool that extracts essential code structure from large codebases, making them digestible for LLMs.

## 🚀 Quick Start with Claude Desktop

```bash
# One-line installation
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp

# Or install globally
npm install -g @janreges/ai-distiller-mcp
```

## ✨ Features

### 🎯 Specialized AI Analysis Tools (New!)
- **aid_hunt_bugs** - Systematically scan for bugs, logic errors, and quality issues
- **aid_suggest_refactoring** - Get specific refactoring suggestions with examples
- **aid_generate_diagram** - Generate 10 architectural Mermaid diagrams
- **aid_analyze_security** - OWASP-focused security vulnerability analysis
- **aid_generate_docs** - Create comprehensive documentation workflows

### Core Analysis Engine
- **aid_analyze** - Direct access to all AI actions for advanced workflows
- Supports all 10 AI actions from the CLI tool

### Legacy Tools (Backwards Compatibility)
- **distill_file** - Extract code structure from a single file
- **distill_directory** - Extract structure from entire directories
- **list_files** - List files with language statistics
- **get_capabilities** - Get server capabilities and supported features

## Configuration

### Environment Variables

- `AID_ROOT` - Root directory for analysis (defaults to current directory)
- `AID_CACHE_DIR` - Cache directory (defaults to ~/.cache/aid)

### Claude Desktop Configuration

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "npx",
      "args": ["-y", "@janreges/ai-distiller-mcp"],
      "env": {
        "AID_ROOT": "/path/to/your/project"
      }
    }
  }
}
```

## 📖 Tool Examples

### 🎯 Specialized Tools (Recommended)

#### Hunt for Bugs
```javascript
// Claude will call:
aid_hunt_bugs({
  target_path: "src/",
  focus_area: "race conditions",
  include_private: true,
  include_patterns: "*.go,*.py"
})
// Returns: Detailed bug analysis with risk levels and fixes
```

#### Get Refactoring Suggestions
```javascript
// Claude will call:
aid_suggest_refactoring({
  target_path: "src/auth/",
  refactoring_goal: "improve testability",
  include_implementation: true
})
// Returns: Specific refactoring suggestions with before/after examples
```

#### Generate Architecture Diagrams
```javascript
// Claude will call:
aid_generate_diagram({
  target_path: "src/",
  diagram_focus: "authentication flow"
})
// Returns: 10 Mermaid diagrams covering different architectural aspects
```

#### Security Analysis
```javascript
// Claude will call:
aid_analyze_security({
  target_path: "src/api/",
  security_focus: "input validation",
  include_private: true,
  include_implementation: true
})
// Returns: OWASP-based security findings with remediation steps
```

#### Generate Documentation
```javascript
// Claude will call:
aid_generate_docs({
  target_path: "src/core/",
  doc_type: "api-reference",
  audience: "developers"
})
// Returns: Structured documentation workflow or single file docs
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

*AI Distiller (aid) - https://aid.siteone.io/*  
*Authored by [Claude Code](https://claude.ai/code) & [Ján Regeš](https://github.com/janreges) from [SiteOne](https://www.siteone.io/) (Czech Republic)*  
*Explore the project on [GitHub](https://github.com/janreges/ai-distiller)*