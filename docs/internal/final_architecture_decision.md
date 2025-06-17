# Final Architecture Decision: Visitor Pattern

## Agreement with Gemini's Analysis

I agree that the **single-pass, Visitor-based stripping architecture** is the superior approach. The upfront investment is justified by:

1. **Performance**: Single traversal vs multiple passes
2. **Extensibility**: Rule-based system enables future growth
3. **Maintainability**: Clean separation of concerns
4. **Correctness**: Avoids refactoring debt

## Refined Architecture Proposal

### Core Components

```go
// 1. Rich IR with all semantic information
type DistilledNode interface {
    Accept(visitor IRVisitor) DistilledNode
    GetLocation() Location
}

// 2. Visitor interface for single-pass traversal
type IRVisitor interface {
    VisitFile(node *DistilledFile) IRVisitor
    VisitFunction(node *DistilledFunction) IRVisitor
    VisitClass(node *DistilledClass) IRVisitor
    VisitComment(node *DistilledComment) IRVisitor
    // ... other node types
}

// 3. Rule-based stripping configuration
type StrippingRule interface {
    Name() string
    ShouldRemove(node DistilledNode, context VisitorContext) bool
    Transform(node DistilledNode, context VisitorContext) DistilledNode
}

// 4. Context for language-specific decisions
type VisitorContext struct {
    Language string
    Parent   DistilledNode
    Depth    int
}
```

### Implementation Strategy

#### Phase 1: Foundation (Week 1-2)
- Define complete IR schema with versioning
- Implement Visitor infrastructure
- Create basic StrippingRules (comments, imports)
- Build PoCs for both CGo and WASM

#### Phase 2: Core Languages (Week 3-4)
- Implement processors for Go, Python, JavaScript
- Each produces rich IR from the start
- Visitor handles all stripping logic

#### Phase 3: Extended Languages (Week 5-6)
- Add remaining language processors
- Enhance stripping rules based on usage
- Performance optimization

## Key Architectural Decisions

### 1. Whitespace Handling
As Gemini noted, whitespace/newline control belongs in the **Formatter**, not the Stripper:
- IR is abstract, doesn't store concrete syntax
- Formatter controls output presentation
- `--compact` flag for formatters, not `--strip whitespace`

### 2. Error Handling Strategy
```go
type DistilledError struct {
    BaseNode
    Message string
    Severity string // "error", "warning"
}
```
- Errors are first-class IR nodes
- Visitor can decide to include/exclude them
- Enables rich error reporting in all formats

### 3. Progressive Enhancement
While we define the full IR schema upfront, processors can initially produce partial IR:
- Start with basic structure (functions, classes)
- Add comments, imports progressively
- Visitor gracefully handles missing nodes

## Benefits Over Hybrid Approach

1. **No Refactoring Debt**: Build it right from the start
2. **Plugin Ready**: Rule interface enables community extensions
3. **Performance Guarantee**: O(n) traversal regardless of strip options
4. **Type Safety**: Visitor pattern provides compile-time guarantees

## Conclusion

I fully support the Visitor-based architecture. The additional upfront complexity is minimal compared to the long-term benefits. This design will serve us well as the project grows from 8 to 20+ languages and from basic stripping to advanced transformations.

Let's proceed with:
1. Finalizing the IR schema document
2. Implementing dual PoCs with this architecture
3. Making the CGo vs WASM decision based on benchmarks