# AI Distiller Intermediate Representation (IR) Schema

## Version 1.0

This document defines the schema for AI Distiller's Intermediate Representation (IR), which serves as the language-agnostic representation of distilled code structure.

## Design Principles

1. **Language Agnostic**: The IR should represent concepts common across all supported languages
2. **Extensible**: Easy to add new node types without breaking existing code
3. **Immutable**: Once created, IR nodes should not be modified
4. **Self-Documenting**: Each node carries sufficient metadata for reconstruction
5. **Source-Mapped**: Every node maintains its relationship to the original source

## Core Types

### Location Information

```go
// Location represents a position in the source file
type Location struct {
    StartLine   int `json:"start_line"`
    StartColumn int `json:"start_column"`
    EndLine     int `json:"end_line"`
    EndColumn   int `json:"end_column"`
    StartByte   int `json:"start_byte,omitempty"`
    EndByte     int `json:"end_byte,omitempty"`
}

// Range represents a continuous span of source code
type Range struct {
    Start Location `json:"start"`
    End   Location `json:"end"`
}
```

### Base Node Types

```go
// DistilledNode is the base interface for all IR nodes
type DistilledNode interface {
    // Accept implements the visitor pattern
    Accept(visitor IRVisitor) DistilledNode
    
    // GetLocation returns the source location of this node
    GetLocation() Location
    
    // GetNodeType returns the type identifier for serialization
    GetNodeType() string
    
    // GetChildren returns child nodes for traversal
    GetChildren() []DistilledNode
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
    Location   Location               `json:"location"`
    Attributes map[string]interface{} `json:"attributes,omitempty"`
}
```

## File-Level Nodes

```go
// DistilledFile represents a single source file
type DistilledFile struct {
    BaseNode
    Path     string           `json:"path"`
    Language string           `json:"language"`
    Version  string           `json:"version"` // IR schema version
    Children []DistilledNode  `json:"nodes"`
    Errors   []DistilledError `json:"errors,omitempty"`
    Metadata FileMetadata     `json:"metadata,omitempty"`
}

// FileMetadata contains additional file information
type FileMetadata struct {
    Size         int64     `json:"size_bytes"`
    Hash         string    `json:"hash,omitempty"`
    LastModified time.Time `json:"last_modified,omitempty"`
    Encoding     string    `json:"encoding,omitempty"`
}

// DistilledError represents a parsing or processing error
type DistilledError struct {
    BaseNode
    Message  string `json:"message"`
    Severity string `json:"severity"` // "error", "warning", "info"
    Code     string `json:"code,omitempty"`
}
```

## Structural Nodes

### Package/Module Level

```go
// DistilledPackage represents a package or module declaration
type DistilledPackage struct {
    BaseNode
    Name    string   `json:"name"`
    Path    string   `json:"path,omitempty"`    // Import path
    Exports []string `json:"exports,omitempty"` // Exported symbols
}

// DistilledImport represents an import statement
type DistilledImport struct {
    BaseNode
    Path      string            `json:"path"`
    Alias     string            `json:"alias,omitempty"`
    Symbols   []ImportedSymbol  `json:"symbols,omitempty"`
    IsDefault bool              `json:"is_default,omitempty"`
}

type ImportedSymbol struct {
    Name  string `json:"name"`
    Alias string `json:"alias,omitempty"`
}
```

### Type Definitions

```go
// DistilledClass represents a class or similar construct
type DistilledClass struct {
    BaseNode
    Name           string           `json:"name"`
    Visibility     string           `json:"visibility"`
    Modifiers      []string         `json:"modifiers,omitempty"`
    TypeParameters []string         `json:"type_parameters,omitempty"`
    Extends        []string         `json:"extends,omitempty"`
    Implements     []string         `json:"implements,omitempty"`
    Members        []DistilledNode  `json:"members"`
}

// DistilledInterface represents an interface or protocol
type DistilledInterface struct {
    BaseNode
    Name           string           `json:"name"`
    Visibility     string           `json:"visibility"`
    TypeParameters []string         `json:"type_parameters,omitempty"`
    Extends        []string         `json:"extends,omitempty"`
    Members        []DistilledNode  `json:"members"`
}

// DistilledStruct represents a struct or record type
type DistilledStruct struct {
    BaseNode
    Name           string           `json:"name"`
    Visibility     string           `json:"visibility"`
    TypeParameters []string         `json:"type_parameters,omitempty"`
    Fields         []DistilledField `json:"fields"`
}

// DistilledEnum represents an enumeration
type DistilledEnum struct {
    BaseNode
    Name       string               `json:"name"`
    Visibility string               `json:"visibility"`
    Type       string               `json:"type,omitempty"`
    Values     []DistilledEnumValue `json:"values"`
}

// DistilledTypeAlias represents a type alias or typedef
type DistilledTypeAlias struct {
    BaseNode
    Name           string   `json:"name"`
    Visibility     string   `json:"visibility"`
    TypeParameters []string `json:"type_parameters,omitempty"`
    Type           string   `json:"type"`
}
```

### Member Nodes

```go
// DistilledFunction represents a function or method
type DistilledFunction struct {
    BaseNode
    Name           string               `json:"name"`
    Signature      string               `json:"signature"`
    Visibility     string               `json:"visibility"`
    Modifiers      []string             `json:"modifiers,omitempty"`
    TypeParameters []string             `json:"type_parameters,omitempty"`
    Parameters     []DistilledParameter `json:"parameters"`
    ReturnType     string               `json:"return_type,omitempty"`
    Body           []DistilledNode      `json:"body,omitempty"`
    Decorators     []string             `json:"decorators,omitempty"`
}

// DistilledParameter represents a function parameter
type DistilledParameter struct {
    Name         string `json:"name"`
    Type         string `json:"type,omitempty"`
    DefaultValue string `json:"default_value,omitempty"`
    IsVariadic   bool   `json:"is_variadic,omitempty"`
    IsOptional   bool   `json:"is_optional,omitempty"`
}

// DistilledField represents a field or property
type DistilledField struct {
    BaseNode
    Name         string   `json:"name"`
    Type         string   `json:"type,omitempty"`
    Visibility   string   `json:"visibility"`
    Modifiers    []string `json:"modifiers,omitempty"`
    DefaultValue string   `json:"default_value,omitempty"`
    Getter       bool     `json:"getter,omitempty"`
    Setter       bool     `json:"setter,omitempty"`
}

// DistilledEnumValue represents an enum value
type DistilledEnumValue struct {
    BaseNode
    Name  string `json:"name"`
    Value string `json:"value,omitempty"`
}
```

### Documentation Nodes

```go
// DistilledComment represents a comment block
type DistilledComment struct {
    BaseNode
    Content string `json:"content"`
    Type    string `json:"type"` // "line", "block", "doc"
    Target  string `json:"target,omitempty"` // What this documents
}
```

## Visibility Constants

```go
const (
    VisibilityPublic    = "public"
    VisibilityPrivate   = "private"
    VisibilityProtected = "protected"
    VisibilityInternal  = "internal"
    VisibilityPackage   = "package"
)
```

## Common Modifiers

```go
const (
    ModifierStatic    = "static"
    ModifierFinal     = "final"
    ModifierAbstract  = "abstract"
    ModifierAsync     = "async"
    ModifierConst     = "const"
    ModifierReadonly  = "readonly"
    ModifierOverride  = "override"
    ModifierVirtual   = "virtual"
    ModifierInline    = "inline"
    ModifierExtern    = "extern"
)
```

## Visitor Interface

```go
// IRVisitor defines the visitor pattern interface
type IRVisitor interface {
    // File level
    VisitFile(node *DistilledFile) IRVisitor
    VisitPackage(node *DistilledPackage) IRVisitor
    VisitImport(node *DistilledImport) IRVisitor
    
    // Type definitions
    VisitClass(node *DistilledClass) IRVisitor
    VisitInterface(node *DistilledInterface) IRVisitor
    VisitStruct(node *DistilledStruct) IRVisitor
    VisitEnum(node *DistilledEnum) IRVisitor
    VisitTypeAlias(node *DistilledTypeAlias) IRVisitor
    
    // Members
    VisitFunction(node *DistilledFunction) IRVisitor
    VisitField(node *DistilledField) IRVisitor
    
    // Other
    VisitComment(node *DistilledComment) IRVisitor
    VisitError(node *DistilledError) IRVisitor
    
    // Generic fallback
    VisitNode(node DistilledNode) IRVisitor
}
```

## Language-Specific Attributes

The `Attributes` map in `BaseNode` allows for language-specific information without polluting the core schema:

### Go-Specific
```go
// For channels
attributes["is_channel"] = true
attributes["channel_direction"] = "send" // "send", "receive", "bidirectional"

// For interfaces
attributes["is_empty_interface"] = true
```

### Python-Specific
```go
// For functions
attributes["is_generator"] = true
attributes["is_coroutine"] = true

// For classes
attributes["metaclass"] = "ABCMeta"
```

### JavaScript/TypeScript-Specific
```go
// For functions
attributes["is_arrow_function"] = true
attributes["is_generator_function"] = true

// For classes
attributes["is_abstract"] = true
```

### C#-Specific
```go
// For properties
attributes["has_init_accessor"] = true
attributes["is_partial"] = true
```

## Serialization

The IR is designed to be easily serializable to JSON. Each node type should include:

1. A `type` field indicating the node type
2. All structural fields
3. The `location` information
4. Optional `attributes` for language-specific data

Example JSON representation:
```json
{
  "type": "function",
  "name": "ProcessData",
  "visibility": "public",
  "signature": "func ProcessData(input []byte) (*Result, error)",
  "modifiers": ["async"],
  "location": {
    "start_line": 42,
    "start_column": 1,
    "end_line": 56,
    "end_column": 2
  },
  "parameters": [
    {
      "name": "input",
      "type": "[]byte"
    }
  ],
  "return_type": "(*Result, error)",
  "attributes": {
    "complexity": 12,
    "has_error_return": true
  }
}
```

## Versioning

The IR schema follows semantic versioning:

- **Major version**: Breaking changes to existing node types
- **Minor version**: New node types or optional fields
- **Patch version**: Documentation or clarification changes

Current version: **1.0.0**

## Implementation Notes

1. **Immutability**: Nodes should be constructed once and never modified. Any transformation should create new nodes.

2. **Lazy Loading**: For large files, consider lazy loading of function bodies and other nested content.

3. **Memory Efficiency**: Use string interning for repeated values (types, modifiers, etc.).

4. **Validation**: Each node should validate its required fields during construction.

5. **Extensibility**: New node types can be added by implementing the `DistilledNode` interface.

## Future Considerations

1. **Binary Serialization**: For performance, consider Protocol Buffers or MessagePack

2. **Incremental Updates**: Support for representing changes between versions

3. **Cross-References**: Add support for symbol resolution and type references

4. **Semantic Information**: Add flow analysis, dependency graphs, etc.

5. **Metrics**: Add code metrics (complexity, lines of code, etc.) as attributes