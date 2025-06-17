# AI Distiller MCP Server

Advanced MCP (Model Context Protocol) server for [AI Distiller](https://github.com/janreges/ai-distiller) - a high-performance tool that extracts essential code structure from large codebases, making them digestible for LLMs.

## 🚀 Quick Start with Claude Code

```bash
# One-line installation
claude mcp add ai-distiller -- npx -y @janreges/ai-distiller-mcp

# Or install globally
npm install -g @janreges/ai-distiller-mcp
```

## ✨ Features

### Core Distillation Tools
- **distill_file** - Extract code structure from a single file
- **distill_directory** - Extract structure from entire directories with advanced filtering
- **list_files** - List files with language statistics and metadata

### Enhanced Code Understanding Tools
- **explain_code_structure** - Extract code structure and generate AI prompt for architectural analysis
- **suggest_refactoring** - Analyze code with full implementation for AI refactoring suggestions

### 🤖 AI-Powered Analysis Tools
- **propose_code_analysis_plan** - Generate comprehensive code analysis plan with systematic task list
- **analyze_git_history** - Analyze git commit history with AI insights and patterns
- **analyze_logs** - 🔥 **WOW FEATURE**: Find newest log files, extract recent entries with AI analysis

### 💪 Enhanced Capabilities
- **Pattern Filtering** - Advanced include/exclude patterns (`*.go,*.py` or `--include "*.go" --include "*.py"`)
- **Multi-format Output** - text, markdown, JSON, XML formats
- **Language Detection** - Automatic detection of 12+ programming languages
- **Direct Integration** - Each tool directly calls aid with appropriate flags

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

### Basic Code Analysis

#### Extract Public API from a File
```javascript
// Claude will call:
distill_file({
  file_path: "src/main.py",
  include_private: false,
  include_implementation: false,
  output_format: "text"
})
```

#### Analyze Entire Module Structure
```javascript
// Claude will call:
distill_directory({
  directory_path: "src/auth",
  recursive: true,
  include_private: false,
  output_format: "json"
})
```


### Advanced AI Analysis Workflows

#### Generate Code Analysis Plan
```javascript
// Claude will call:
propose_code_analysis_plan({
  path: "src/",
  include_patterns: "*.go,*.py",
  exclude_patterns: "*test*"
})
// Returns: Comprehensive task list with distilled source code and sophisticated AI analysis prompt
```

#### Git History Analysis with AI Insights
```javascript
// Claude will call:
analyze_git_history({
  git_limit: 100,
  with_analysis_prompt: true
})
// Directly calls: aid .git --git-limit=100 --with-analysis-prompt
// Returns commit patterns, contributor expertise, development phases
```

#### 🔥 Advanced Log Analysis (WOW FEATURE!)
```javascript
// Claude will call:
analyze_logs({
  path: "logs/",
  max_files: 10,
  lines_per_file: 100,
  include_analysis_prompt: true,
  output_format: "ndjson"
})
// Recursively finds 10 newest log files, extracts 100 last lines from each
// Returns structured data with timestamps, log levels, source file markers
// Includes comprehensive AI analysis prompt for error/performance/security insights
```

#### Advanced Directory Analysis with Patterns
```javascript
// Claude will call:
distill_directory({
  directory_path: "src/auth",
  include_patterns: "*.go,*.py",
  exclude_patterns: "*test*,*mock*",
  include_private: false,
  output_format: "json"
})
// Directly calls: aid src/auth --include "*.go,*.py" --exclude "*test*,*mock*" --format json
```

### Code Structure & Refactoring

#### Explain Project Architecture
```javascript
explain_code_structure({
  path: "src/",
  include_patterns: "*.go,*.py",
  exclude_patterns: "*test*"
})
// Returns: Distilled code structure (public only, no implementation) with comprehensive AI analysis prompt
```

#### Suggest Refactorings
```javascript
suggest_refactoring({
  path: "src/user_service.py",
  goal: "improve code quality and maintainability",
  include_patterns: "*.py",
  exclude_patterns: "*test*"
})
// Returns: Full source code with implementation and comprehensive AI refactoring prompt
```


## 🤖 AI Agent Workflows

### Quick Project Understanding
```
1. list_files() → get project overview with language stats
2. distill_directory() → get structural overview 
3. propose_code_analysis_plan() → create systematic analysis task list
4. Agent follows plan using distill_file() for individual files
```

### AI-Driven Analysis Workflow
```
1. propose_code_analysis_plan(path="src/") → get structured task list with source code
2. Agent analyzes the provided distilled code
3. Agent builds comprehensive project understanding
4. Agent provides insights based on complete analysis
```

### Git History Insights  
```
1. analyze_git_history(git_limit=200) → get commit patterns + AI analysis prompt
2. Agent analyzes contributor expertise areas
3. Agent identifies development phases and patterns
4. Agent provides strategic recommendations
```

### Code Structure Analysis
```
1. explain_code_structure() → get architecture with AI analysis prompt
2. Agent analyzes the structure and patterns
3. suggest_refactoring() → get full code with refactoring prompt
4. Agent provides specific improvement recommendations
```

### 🔥 Log Analysis & Troubleshooting (WOW FEATURE!)
```
1. analyze_logs(max_files=10, include_analysis_prompt=true) → get recent log data with AI prompt
2. Agent analyzes errors, performance issues, security threats, operational patterns
3. Agent identifies critical issues requiring immediate attention
4. Agent provides actionable recommendations and risk assessment
5. Follow up with distill_file() for specific files mentioned in logs
```

## ⚡ Performance & Simplicity

- **Direct Integration** - Each tool directly calls aid (no overhead)
- **Synchronous Responses** - Simple request-response pattern
- **Pattern Filtering** - Efficient file selection with include/exclude
- **Multi-format Support** - Choose optimal format for AI consumption
- **Language Detection** - Automatic language identification

## Requirements

- Node.js >= 14.0.0
- AI Distiller binary (automatically downloaded or available in PATH)

## Security

- Path traversal protection - all paths are sanitized
- File size limits - max 10MB per file
- Operation timeouts - configurable max timeout
- Resource limits - max files per operation

## Development

```bash
# Clone the repository
git clone https://github.com/janreges/ai-distiller.git
cd ai-distiller/mcp-npm

# Install dependencies
npm install

# Build the binary
npm run build

# Test locally
node index.js
```

## 🤖 AI Agent Example Usage Patterns

Here are 5 different types of sentences/requests that can be used with each specialized tool:

### `aid_hunt_bugs` - Bug Hunting
1. **Quality Assurance**: "Please scan this authentication module for potential security vulnerabilities and race conditions."
2. **Debugging Session**: "I'm seeing intermittent failures in the payment processing - can you check for any logical errors?"
3. **Code Review**: "Before we ship this feature, let's do a comprehensive bug sweep of the order management system."
4. **Legacy Code**: "This old codebase has been causing issues - hunt for any potential null pointer exceptions or memory leaks."
5. **Pre-production**: "Run a final bug check on the user registration flow to catch any edge cases we might have missed."

### `aid_suggest_refactoring` - Refactoring Suggestions  
1. **Performance**: "This data processing pipeline is too slow - suggest ways to optimize it for better performance."
2. **Maintainability**: "The `UserController` class has become unwieldy - help me break it down into more manageable pieces."
3. **Readability**: "The authentication logic is confusing - suggest refactoring to make it more readable and testable."
4. **Code Smells**: "I suspect there's code duplication in the payment handlers - identify and suggest how to eliminate it."
5. **Architecture**: "This monolithic service is getting complex - suggest how to split it into more focused components."

### `aid_generate_diagram` - Diagram Generation
1. **Architecture Overview**: "Create diagrams showing the overall system architecture and how components interact."
2. **Flow Visualization**: "Generate sequence diagrams showing how the user login process works across different services."
3. **Data Relationships**: "Build entity relationship diagrams from our database models and data access layer."
4. **Process Flow**: "Create flowcharts showing the order fulfillment process from cart to delivery."
5. **System Dependencies**: "Generate dependency diagrams showing how our microservices communicate with each other."

### `aid_analyze_security` - Security Analysis
1. **Vulnerability Assessment**: "Audit the API endpoints for common OWASP Top 10 vulnerabilities like injection flaws."
2. **Input Validation**: "Check all user input handling for potential XSS and SQL injection vulnerabilities."
3. **Authentication Review**: "Analyze the authentication and authorization mechanisms for security weaknesses."
4. **Data Protection**: "Review how sensitive data is handled - encryption, storage, and transmission security."
5. **Compliance Check**: "Perform a security audit to ensure we meet SOC 2 and GDPR compliance requirements."

### `aid_generate_docs` - Documentation Generation
1. **API Documentation**: "Generate comprehensive API documentation for our REST endpoints with usage examples."
2. **Developer Guide**: "Create developer documentation explaining how to set up and contribute to this project."
3. **Architecture Documentation**: "Document the system architecture and design decisions for the engineering team."
4. **Integration Guide**: "Generate documentation showing how to integrate with our webhook system."
5. **Troubleshooting Guide**: "Create documentation helping users troubleshoot common integration issues."

### `aid_analyze` - Core Analysis Engine
1. **Custom Analysis**: "Use the 'prompt-for-performance-analysis' action to deep-dive into CPU-intensive algorithms."
2. **Complex Workflows**: "Run 'flow-for-deep-file-to-file-analysis' to understand cross-file dependencies in this module."
3. **Specialized Tasks**: "Execute 'prompt-for-best-practices-analysis' to evaluate our coding standards compliance."
4. **Multi-file Documentation**: "Use 'flow-for-multi-file-docs' to create structured documentation for the entire project."
5. **Visual Analysis**: "Run 'prompt-for-diagrams' to generate Mermaid diagrams for our microservices architecture."

These examples demonstrate the natural language patterns that AI agents can use to effectively leverage each specialized tool for different analysis scenarios.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- [AI Distiller on GitHub](https://github.com/janreges/ai-distiller)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [Claude Desktop](https://claude.ai)

---

**AI Distiller (aid)** - https://aid.siteone.io/  
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)  
Explore the project on [GitHub](https://github.com/janreges/ai-distiller)