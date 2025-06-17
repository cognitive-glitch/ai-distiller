# Example Prompts for Claude with AI Distiller MCP

Here are example prompts that demonstrate how to effectively use the new AI Distiller MCP tools:

## 🎯 Using Specialized Tools (Recommended)

### Bug Hunting
> "I've been getting random crashes in production. Can you check the codebase for potential bugs?"

Claude will use:
- `aid_hunt_bugs` to systematically scan for bugs, race conditions, and quality issues
- Focus on areas like null pointer exceptions, unhandled errors, and concurrency issues

### Refactoring Complex Code
> "The authentication module has grown too complex. Can you suggest how to refactor it?"

Claude will use:
- `aid_suggest_refactoring` with goal "reduce complexity" to get specific suggestions
- Provides before/after code examples and actionable steps

### Understanding Architecture
> "I just joined this project. Can you help me understand the overall architecture?"

Claude will use:
- `aid_generate_diagram` to create 10 different architectural views
- Includes flowcharts, sequence diagrams, class diagrams, and system overviews

### Security Audit
> "We're preparing for a security audit. Can you check our API endpoints for vulnerabilities?"

Claude will use:
- `aid_analyze_security` focusing on OWASP Top 10 vulnerabilities
- Returns categorized findings with risk levels and remediation steps

### Documentation Generation
> "We need to document our core API module for external developers"

Claude will use:
- `aid_generate_docs` with doc_type="api-reference" and audience="developers"
- Creates comprehensive API documentation with examples

## 📚 Advanced Workflows

### Complete Code Review
> "Review the payment processing module for bugs, security issues, and suggest improvements"

Claude will:
1. Use `aid_hunt_bugs` to find potential issues
2. Use `aid_analyze_security` to check for vulnerabilities  
3. Use `aid_suggest_refactoring` to improve code quality
4. Provide a comprehensive review report

### Onboarding New Developer
> "Create an onboarding guide for a new developer joining our team"

Claude will:
1. Use `aid_generate_diagram` to visualize the architecture
2. Use `aid_generate_docs` with audience="maintainers" for key modules
3. Use `list_files` to show project structure
4. Create a structured onboarding document

### Performance Investigation
> "The application has been running slowly. Can you help identify potential bottlenecks?"

Claude will:
1. Use `aid_analyze` with ai_action="prompt-for-performance-analysis"
2. Focus on algorithmic complexity and resource usage patterns
3. Suggest specific optimizations

### Technical Debt Assessment
> "We need to assess technical debt before our next sprint planning"

Claude will:
1. Use `aid_analyze` with ai_action="prompt-for-best-practices-analysis"
2. Use `aid_suggest_refactoring` on problematic modules
3. Create a prioritized technical debt backlog

## 🔧 Using Core Analysis Engine

### Custom Analysis Workflows
> "Analyze how data flows through our microservices, focusing on the order processing system"

Claude will use:
```
aid_analyze({
  ai_action: "flow-for-deep-file-to-file-analysis",
  target_path: "services/",
  user_query: "trace order processing data flow",
  include_patterns: "*.go,*.proto"
})
```

### Multi-File Documentation
> "Generate documentation for all our utility modules"

Claude will use:
```
aid_analyze({
  ai_action: "flow-for-multi-file-docs",
  target_path: "utils/",
  include_patterns: "*.py"
})
```

## 💡 Tips for Effective Prompts

1. **Be specific about your goal**: "find memory leaks" vs "check for bugs"
2. **Mention specific areas**: "in the authentication module" vs "in the codebase"
3. **State the purpose**: "for security audit" vs "for code review"
4. **Include context**: "we're using Python 3.11 with FastAPI"

## 🚀 Power User Tips

### Combining Tools
For comprehensive analysis, Claude will often combine multiple tools:
- Bug hunting + Security analysis for pre-deployment checks
- Diagram generation + Documentation for onboarding materials
- Refactoring suggestions + Best practices for code improvement sprints

### Pattern Filtering
Be specific with file patterns to improve analysis speed:
- `include_patterns: "*.py,*.pyi"` for Python projects
- `exclude_patterns: "*test*,*mock*"` to skip test files
- `include_patterns: "src/**/*.ts"` for TypeScript source files

### Visibility Control
- Use `include_private: true` for bug hunting and security analysis
- Use `include_implementation: true` for refactoring suggestions
- Keep defaults (public only) for API documentation

## 📋 Quick Reference

| Task | Best Tool | Key Parameters |
|------|-----------|----------------|
| Find bugs | `aid_hunt_bugs` | `include_private: true` |
| Improve code | `aid_suggest_refactoring` | `refactoring_goal: "..."` |
| Understand architecture | `aid_generate_diagram` | `diagram_focus: "..."` |
| Security check | `aid_analyze_security` | `security_focus: "..."` |
| Create docs | `aid_generate_docs` | `doc_type: "...", audience: "..."` |
| Custom analysis | `aid_analyze` | `ai_action: "...", user_query: "..."` |

---

*AI Distiller (aid) - https://aid.siteone.io/*  
*Explore more on [GitHub](https://github.com/janreges/ai-distiller)*