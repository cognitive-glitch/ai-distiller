# IR Depth Analysis and Response to Gemini

## Research Results

### Language Server Protocol Comparison

**DocumentSymbol**: High-level structural navigation
- Focuses on symbols (classes, functions, variables)
- Hierarchical representation
- Used for outline views and navigation

**Semantic Tokens**: Token-level syntax highlighting
- Provides classification for every meaningful element
- Includes type, modifiers, and precise positions
- More granular than DocumentSymbol

### Tree-sitter CST Structure

Tree-sitter produces a **Concrete Syntax Tree** where:
- Each node corresponds directly to grammar symbols
- Contains ALL syntactic information (including whitespace)
- Can be made to contain all CST-relevant data
- Designed to work in reference to a text document

## Critical Decision: IR Depth

Based on our use case (distilling code for LLM context windows), I propose:

### **Declaration-Level IR with Strategic Depth**

We should NOT create a full AST for function bodies. Instead:

1. **Primary Focus**: Declarations, signatures, and structure
2. **Function Bodies**: Optional summary information only
   - Complexity metrics
   - Called functions/methods
   - Used types
   - Control flow summary

### Rationale:

1. **Purpose Alignment**: LLMs need to understand APIs and structure, not implementation details
2. **Performance**: Full AST for large codebases would be prohibitive
3. **Stripping Logic**: `--strip implementation` is a core feature
4. **Practical Value**: Users want "what can I call" not "how is it implemented"

### Proposed Body Representation:

```go
type DistilledFunction struct {
    // ... existing fields ...
    Body         *FunctionBody `json:"body,omitempty"`
}

type FunctionBody struct {
    StartLine    int              `json:"start_line"`
    EndLine      int              `json:"end_line"`
    Complexity   int              `json:"complexity,omitempty"`
    CalledFuncs  []string         `json:"called_functions,omitempty"`
    UsedTypes    []string         `json:"used_types,omitempty"`
    HasErrorHandling bool         `json:"has_error_handling,omitempty"`
    // Language-specific metrics in Attributes
}
```

This provides valuable insights without the full AST overhead.