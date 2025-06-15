# Example Prompts for Claude with AI Distiller MCP

Here are some example prompts that demonstrate how to effectively use AI Distiller MCP tools:

## Understanding a New Codebase

> "I just cloned this Python project. Can you analyze its architecture and give me an overview of the main components?"

Claude will use:
- `listFiles` to understand project structure
- `distillDirectory` on key directories like `src/`, `lib/`, etc.
- `distillFile` on main entry points

## Finding Implementation Details

> "Show me how the authentication system works in this project. I need to understand the login flow."

Claude will use:
- `search` to find authentication-related files
- `distillDirectory` on auth modules
- `getFileContent` to examine specific implementations

## Code Review Preparation

> "I'm about to review the payment module. Can you summarize what's in `src/payments/` and highlight the main classes and their responsibilities?"

Claude will use:
- `distillDirectory` with `include_private=true` for comprehensive view
- `listFiles` to see all payment-related files

## Searching for Patterns

> "Find all TODO and FIXME comments in the codebase and summarize what needs to be done."

Claude will use:
- `search` with `query="TODO|FIXME"` and `mode="regex"`
- `getFileContent` to read context around findings

## API Documentation

> "Generate a summary of all public APIs in the `api/` directory. I need to know what endpoints are available."

Claude will use:
- `distillDirectory` with default settings (public only)
- Output format can be JSON for structured data

## Dependency Analysis

> "What external libraries does this project use? Check the import statements across all Python files."

Claude will use:
- `distillDirectory` on the entire project
- Focus on import statements in the output

## Test Coverage Understanding

> "Show me the test structure for the user service. I want to know what's being tested."

Claude will use:
- `listFiles` with pattern `test_*.py` or `*_test.py`
- `distillFile` on test files to see test methods

## Refactoring Assistance

> "I need to refactor the database module. First, show me all the classes and their relationships in `src/database/`."

Claude will use:
- `distillDirectory` with `include_private=true`
- May use `search` to find usages elsewhere

## Security Audit

> "Search for potential security issues: hardcoded passwords, API keys, or sensitive data in the code."

Claude will use:
- `search` with patterns like "password=", "api_key=", "secret"
- `getFileContent` to examine suspicious findings

## Performance Analysis

> "Find all database queries in the application and check if they're using proper indexing."

Claude will use:
- `search` for SQL patterns or ORM query methods
- `distillFile` to understand query context

## Tips for Effective Prompts

1. **Be specific about scope**: "in the `src/auth/` directory" vs "in the entire project"
2. **Mention if you need private members**: "including private methods" triggers `include_private=true`
3. **Specify output needs**: "as JSON" or "show me the implementation" affects tool parameters
4. **Combine tools**: Complex tasks often need multiple tools working together

## Advanced Workflow Example

> "I'm implementing a new feature similar to the existing notification system. Can you:
> 1. Show me the current notification architecture
> 2. Find where notifications are triggered
> 3. List all notification types
> 4. Suggest where I should add my new feature"

This complex request will trigger Claude to:
1. Use `distillDirectory` on notification modules
2. Use `search` for notification trigger points
3. Use `listFiles` to find all relevant files
4. Analyze patterns and suggest implementation approach