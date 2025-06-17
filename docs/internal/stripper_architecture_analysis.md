# Stripper Architecture Analysis

## The Question: Embedded vs Separate Stripper Component

Gemini raises a critical architectural decision about the `--strip` logic placement. Let me analyze both approaches:

### Option A: Embedded in LanguageProcessor

```go
type LanguageProcessor interface {
    Distill(tree *sitter.Tree, source []byte, options StrippingOptions) (*DistilledFile, error)
}
```

**Pros:**
- Simpler initial implementation
- Language-specific stripping logic co-located with parsing
- Direct access to AST for complex decisions

**Cons:**
- Violates Single Responsibility Principle
- Harder to test stripping logic in isolation
- Duplicated stripping patterns across languages

### Option B: Separate Stripper Component

```go
type LanguageProcessor interface {
    Parse(tree *sitter.Tree, source []byte) (*DistilledFile, error)
}

type Stripper interface {
    Strip(file *DistilledFile, options StrippingOptions) (*DistilledFile, error)
}
```

**Pros:**
- Clean separation of concerns
- Testable in isolation
- Enables future plugin architecture
- Allows for cross-language stripping patterns

**Cons:**
- Requires rich enough IR to make stripping decisions
- Potential performance overhead (walking IR twice)
- More complex initial architecture

## My Recommendation: Hybrid Approach

I propose a **hybrid architecture** that gets the best of both worlds:

```go
// Core parsing produces a rich, unstripped IR
type LanguageProcessor interface {
    Parse(tree *sitter.Tree, source []byte) (*DistilledFile, error)
}

// Generic stripper handles common patterns
type GenericStripper struct{}

func (s *GenericStripper) Strip(file *DistilledFile, options StrippingOptions) (*DistilledFile, error) {
    // Handle universal stripping: comments, whitespace, etc.
}

// Language-specific strippers extend generic behavior
type GoStripper struct {
    GenericStripper
}

func (s *GoStripper) Strip(file *DistilledFile, options StrippingOptions) (*DistilledFile, error) {
    // First apply generic stripping
    file, err := s.GenericStripper.Strip(file, options)
    if err != nil {
        return nil, err
    }
    
    // Then apply Go-specific rules (e.g., non-exported identifiers)
    if options.NonPublic {
        file = s.stripNonExported(file)
    }
    
    return file, nil
}

// Factory to get the right stripper
func GetStripper(language string) Stripper {
    switch language {
    case "go":
        return &GoStripper{}
    case "python":
        return &PythonStripper{}
    default:
        return &GenericStripper{}
    }
}
```

## Benefits of This Approach

1. **Modularity**: Parsing and stripping are separate concerns
2. **Reusability**: Common stripping logic is shared
3. **Extensibility**: Easy to add language-specific rules
4. **Testability**: Each component can be tested independently
5. **Performance**: Single IR traversal with language-specific optimizations

## Implementation Strategy

1. **Phase 1**: Start with embedded stripping in LanguageProcessor (simpler)
2. **Phase 2**: Extract common patterns to GenericStripper
3. **Phase 3**: Refactor to full separation with language-specific strippers

This allows us to ship faster while maintaining a clear path to the ideal architecture.

## Critical Consideration: IR Richness

For this to work, our IR must capture enough semantic information:

```go
type DistilledFunction struct {
    BaseNode
    Name       string
    Signature  string
    Visibility string // "public", "private", "protected"
    Modifiers  []string // "static", "async", "virtual", etc.
    Attributes map[string]any // Language-specific metadata
}
```

The `Visibility` and `Modifiers` fields enable cross-language stripping decisions, while `Attributes` handles language-specific nuances.