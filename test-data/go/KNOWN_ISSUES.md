# Known Issues - Go Language Support

## Comment Handling

**Status**: Not implemented
**Impact**: High - Comments provide critical semantic context for AI understanding

### Current Behavior
The Go AST parser correctly parses comments (using `parser.ParseComments`), but does not extract them into the IR structure. As a result:

1. **Doc comments** are lost (package, function, type documentation)
2. **Inline comments** are lost (end-of-line explanations)  
3. **Implementation comments** are lost (comments within function bodies)
4. **Build constraints** are partially handled but in a non-standard format

### Expected Behavior
All comments should be preserved and associated with their relevant code elements:
- Package doc comments before `package` declaration
- Function/type doc comments before their declarations
- Inline comments at end of lines
- Block comments within implementations

### Technical Details
The issue is in `internal/language/golang/ast_parser.go`. The parser needs to:
1. Use `ast.CommentMap` to associate comments with nodes
2. Extract doc comments from AST nodes (`Doc` field)
3. Handle trailing comments and line comments
4. Preserve comment formatting and position

## Var Block Formatting

**Status**: Partially working
**Impact**: Low - Cosmetic issue

### Current Behavior
```go
var (
    IsEnabled = true
    UserCount int64 = 100
)
```

Is formatted as:
```go
var IsEnabled = true
var UserCount int64 = 100
```

### Expected Behavior
Preserve the original block formatting when multiple variables are declared together.

## Import Grouping

**Status**: Working but could be improved
**Impact**: Low - Cosmetic issue

### Current Behavior
All imports are grouped in a single block, losing the original grouping (stdlib vs third-party).

### Expected Behavior
Preserve import grouping with blank lines between groups.