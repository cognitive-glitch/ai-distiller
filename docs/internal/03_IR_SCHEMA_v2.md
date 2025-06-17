# AI Distiller Intermediate Representation (IR) Schema

## Version 2.0

This document defines the schema for AI Distiller's Intermediate Representation (IR), which serves as the language-agnostic representation of distilled code structure.

## Design Principles

1. **Language Agnostic**: The IR should represent concepts common across all supported languages
2. **Extensible**: Easy to add new node types without breaking existing code
3. **Immutable**: Once created, IR nodes should not be modified
4. **Self-Documenting**: Each node carries sufficient metadata for reconstruction
5. **Source-Mapped**: Every node maintains its relationship to the original source
6. **Declaration-Focused**: Emphasizes API structure over implementation details

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

### Symbol References

```go
// SymbolID uniquely identifies a symbol within a compilation unit
type SymbolID string

// SymbolRef represents a reference to a symbol
type SymbolRef struct {
    ID       SymbolID `json:"id"`
    Name     string   `json:"name"`     // Human-readable name
    Package  string   `json:"package,omitempty"`
    IsBuiltin bool    `json:"is_builtin,omitempty"`
}
```

### Type System

```go
// DistilledType represents type information in a structured way
type DistilledType interface {
    TypeString() string // Human-readable representation
    Accept(visitor TypeVisitor) DistilledType
}

// NamedType represents a named type reference
type NamedType struct {
    Name           string          `json:"name"`
    Package        string          `json:"package,omitempty"`
    TypeArguments  []DistilledType `json:"type_arguments,omitempty"`
    Symbol         *SymbolRef      `json:"symbol,omitempty"`
}

// PointerType represents a pointer/reference type
type PointerType struct {
    Pointee DistilledType `json:"pointee"`
}

// ArrayType represents an array/slice type
type ArrayType struct {
    ElementType DistilledType `json:"element_type"`
    Size        *int         `json:"size,omitempty"` // nil for dynamic arrays
}

// MapType represents a map/dictionary type
type MapType struct {
    KeyType   DistilledType `json:"key_type"`
    ValueType DistilledType `json:"value_type"`
}

// FunctionType represents a function signature as a type
type FunctionType struct {
    Parameters  []DistilledType `json:"parameters"`
    ReturnTypes []DistilledType `json:"return_types"`
    IsVariadic  bool           `json:"is_variadic,omitempty"`
}

// UnionType represents a union/sum type (TypeScript unions, Rust enums)
type UnionType struct {
    Types []DistilledType `json:"types"`
}

// GenericType represents a type parameter
type GenericType struct {
    Name        string        `json:"name"`
    Constraints []DistilledType `json:"constraints,omitempty"`
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
    
    // GetSymbolID returns the symbol ID if this node declares a symbol
    GetSymbolID() *SymbolID
    
    // GetNodeKind returns the kind of node for generic handling
    GetNodeKind() NodeKind
}

// NodeKind represents the type of node
type NodeKind string

const (
    KindFile      NodeKind = "file"
    KindPackage   NodeKind = "package"
    KindImport    NodeKind = "import"
    KindClass     NodeKind = "class"
    KindInterface NodeKind = "interface"
    KindStruct    NodeKind = "struct"
    KindEnum      NodeKind = "enum"
    KindFunction  NodeKind = "function"
    KindField     NodeKind = "field"
    KindTypeAlias NodeKind = "type_alias"
    KindComment   NodeKind = "comment"
    KindError     NodeKind = "error"
)

// BaseNode provides common functionality for all nodes
type BaseNode struct {
    Location   Location        `json:"location"`
    SymbolID   *SymbolID       `json:"symbol_id,omitempty"`
    Extensions *NodeExtensions `json:"extensions,omitempty"`
}

// NodeExtensions provides typed language-specific extensions
type NodeExtensions struct {
    Go         *GoExtensions         `json:"go,omitempty"`
    Python     *PythonExtensions     `json:"python,omitempty"`
    JavaScript *JavaScriptExtensions `json:"javascript,omitempty"`
    TypeScript *TypeScriptExtensions `json:"typescript,omitempty"`
    Java       *JavaExtensions       `json:"java,omitempty"`
    CSharp     *CSharpExtensions     `json:"csharp,omitempty"`
    Rust       *RustExtensions       `json:"rust,omitempty"`
    // Raw attributes for truly custom/rare features
    Attributes map[string]interface{} `json:"attributes,omitempty"`
}
```

## File-Level Nodes

```go
// DistilledFile represents a single source file
type DistilledFile struct {
    BaseNode
    Path        string                    `json:"path"`
    Language    string                    `json:"language"`
    Version     string                    `json:"version"` // IR schema version
    Children    []DistilledNode           `json:"nodes"`
    Errors      []DistilledError          `json:"errors,omitempty"`
    SymbolTable map[SymbolID]DistilledNode `json:"-"` // Internal symbol resolution
    Metadata    *FileMetadata             `json:"metadata,omitempty"`
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
    Name    string      `json:"name"`
    Path    string      `json:"path,omitempty"`    // Import path
    Exports []SymbolRef `json:"exports,omitempty"` // Exported symbols
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
    Name  string    `json:"name"`
    Alias string    `json:"alias,omitempty"`
    Ref   SymbolRef `json:"ref,omitempty"`
}
```

### Type Definitions

```go
// DistilledClass represents a class or similar construct
type DistilledClass struct {
    BaseNode
    Name           string            `json:"name"`
    Visibility     Visibility        `json:"visibility"`
    Modifiers      []Modifier        `json:"modifiers,omitempty"`
    TypeParameters []GenericType     `json:"type_parameters,omitempty"`
    Extends        []DistilledType   `json:"extends,omitempty"`
    Implements     []DistilledType   `json:"implements,omitempty"`
    Annotations    []DistilledAnnotation `json:"annotations,omitempty"`
    Members        []DistilledNode   `json:"members"`
}

// DistilledInterface represents an interface or protocol
type DistilledInterface struct {
    BaseNode
    Name           string            `json:"name"`
    Visibility     Visibility        `json:"visibility"`
    TypeParameters []GenericType     `json:"type_parameters,omitempty"`
    Extends        []DistilledType   `json:"extends,omitempty"`
    Annotations    []DistilledAnnotation `json:"annotations,omitempty"`
    Members        []DistilledNode   `json:"members"`
}

// DistilledStruct represents a struct or record type
type DistilledStruct struct {
    BaseNode
    Name           string            `json:"name"`
    Visibility     Visibility        `json:"visibility"`
    Modifiers      []Modifier        `json:"modifiers,omitempty"`
    TypeParameters []GenericType     `json:"type_parameters,omitempty"`
    Annotations    []DistilledAnnotation `json:"annotations,omitempty"`
    Fields         []DistilledField  `json:"fields"`
}

// DistilledEnum represents an enumeration
type DistilledEnum struct {
    BaseNode
    Name        string               `json:"name"`
    Visibility  Visibility           `json:"visibility"`
    Modifiers   []Modifier           `json:"modifiers,omitempty"`
    UnderlyingType DistilledType     `json:"underlying_type,omitempty"`
    Annotations []DistilledAnnotation `json:"annotations,omitempty"`
    Values      []DistilledEnumValue `json:"values"`
}

// DistilledTypeAlias represents a type alias or typedef
type DistilledTypeAlias struct {
    BaseNode
    Name           string          `json:"name"`
    Visibility     Visibility      `json:"visibility"`
    TypeParameters []GenericType   `json:"type_parameters,omitempty"`
    Type           DistilledType   `json:"type"`
}
```

### Member Nodes

```go
// DistilledFunction represents a function or method
type DistilledFunction struct {
    BaseNode
    Name           string                `json:"name"`
    Signature      string                `json:"signature"` // Human-readable
    Visibility     Visibility            `json:"visibility"`
    Modifiers      []Modifier            `json:"modifiers,omitempty"`
    TypeParameters []GenericType         `json:"type_parameters,omitempty"`
    Parameters     []DistilledParameter  `json:"parameters"`
    ReturnType     DistilledType         `json:"return_type,omitempty"`
    Annotations    []DistilledAnnotation `json:"annotations,omitempty"`
    Body           *FunctionBody         `json:"body,omitempty"`
}

// FunctionBody represents summarized function implementation
type FunctionBody struct {
    StartLine        int         `json:"start_line"`
    EndLine          int         `json:"end_line"`
    Complexity       int         `json:"complexity,omitempty"`
    CalledFunctions  []SymbolRef `json:"called_functions,omitempty"`
    UsedTypes        []SymbolRef `json:"used_types,omitempty"`
    HasErrorHandling bool        `json:"has_error_handling,omitempty"`
    Metrics          map[string]interface{} `json:"metrics,omitempty"`
}

// DistilledParameter represents a function parameter
type DistilledParameter struct {
    Name         string          `json:"name"`
    Type         DistilledType   `json:"type,omitempty"`
    DefaultValue string          `json:"default_value,omitempty"`
    IsVariadic   bool            `json:"is_variadic,omitempty"`
    IsOptional   bool            `json:"is_optional,omitempty"`
    Annotations  []DistilledAnnotation `json:"annotations,omitempty"`
}

// DistilledField represents a field or property
type DistilledField struct {
    BaseNode
    Name         string          `json:"name"`
    Type         DistilledType   `json:"type,omitempty"`
    Visibility   Visibility      `json:"visibility"`
    Modifiers    []Modifier      `json:"modifiers,omitempty"`
    DefaultValue string          `json:"default_value,omitempty"`
    Getter       *AccessorInfo   `json:"getter,omitempty"`
    Setter       *AccessorInfo   `json:"setter,omitempty"`
    Annotations  []DistilledAnnotation `json:"annotations,omitempty"`
}

// AccessorInfo represents getter/setter information
type AccessorInfo struct {
    Exists     bool       `json:"exists"`
    Visibility Visibility `json:"visibility,omitempty"`
    Body       *FunctionBody `json:"body,omitempty"`
}

// DistilledEnumValue represents an enum value
type DistilledEnumValue struct {
    BaseNode
    Name        string               `json:"name"`
    Value       string               `json:"value,omitempty"`
    Annotations []DistilledAnnotation `json:"annotations,omitempty"`
}
```

### Annotations and Documentation

```go
// DistilledAnnotation represents decorators/annotations/attributes
type DistilledAnnotation struct {
    BaseNode
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// DistilledComment represents a comment block
type DistilledComment struct {
    BaseNode
    Content string     `json:"content"`
    Type    string     `json:"type"` // "line", "block", "doc"
    Target  *SymbolRef `json:"target,omitempty"` // What this documents
}
```

## Visibility and Modifiers

```go
// Visibility represents access control
type Visibility string

const (
    VisibilityPublic      Visibility = "public"
    VisibilityPrivate     Visibility = "private"
    VisibilityProtected   Visibility = "protected"
    VisibilityInternal    Visibility = "internal"
    VisibilityPackage     Visibility = "package"
    VisibilityFilePrivate Visibility = "fileprivate" // Swift
    VisibilityOpen        Visibility = "open"        // Swift
    VisibilityFriend      Visibility = "friend"      // C++
)

// Modifier represents various modifiers
type Modifier string

const (
    ModifierStatic    Modifier = "static"
    ModifierFinal     Modifier = "final"
    ModifierAbstract  Modifier = "abstract"
    ModifierAsync     Modifier = "async"
    ModifierConst     Modifier = "const"
    ModifierReadonly  Modifier = "readonly"
    ModifierOverride  Modifier = "override"
    ModifierVirtual   Modifier = "virtual"
    ModifierInline    Modifier = "inline"
    ModifierExtern    Modifier = "extern"
    ModifierSealed    Modifier = "sealed"    // C#
    ModifierData      Modifier = "data"      // Kotlin/Java
    ModifierReified   Modifier = "reified"   // Kotlin
    ModifierMutable   Modifier = "mut"       // Rust
    ModifierPartial   Modifier = "partial"   // C#
    ModifierVolatile  Modifier = "volatile"
    ModifierTransient Modifier = "transient"
)
```

## Visitor Interface (Simplified)

```go
// IRVisitor defines the visitor pattern interface
type IRVisitor interface {
    // Single visit method with type switching
    Visit(node DistilledNode) IRVisitor
}

// BaseVisitor provides default implementation
type BaseVisitor struct{}

func (v *BaseVisitor) Visit(node DistilledNode) IRVisitor {
    // Default: visit children
    return v
}

// Example specialized visitor
type MyVisitor struct {
    BaseVisitor
}

func (v *MyVisitor) Visit(node DistilledNode) IRVisitor {
    switch n := node.(type) {
    case *DistilledClass:
        // Handle class
    case *DistilledFunction:
        // Handle function
    default:
        // Use base implementation
        return v.BaseVisitor.Visit(node)
    }
    return v
}
```

## Language-Specific Extensions

### Go Extensions
```go
type GoExtensions struct {
    // For types
    IsChannel        bool   `json:"is_channel,omitempty"`
    ChannelDirection string `json:"channel_direction,omitempty"` // "send", "receive", "both"
    
    // For interfaces
    IsEmptyInterface bool `json:"is_empty_interface,omitempty"`
    
    // For functions
    ReceiverType string `json:"receiver_type,omitempty"`
    IsMethod     bool   `json:"is_method,omitempty"`
}
```

### Python Extensions
```go
type PythonExtensions struct {
    // For functions
    IsGenerator     bool   `json:"is_generator,omitempty"`
    IsCoroutine     bool   `json:"is_coroutine,omitempty"`
    IsStaticMethod  bool   `json:"is_static_method,omitempty"`
    IsClassMethod   bool   `json:"is_class_method,omitempty"`
    
    // For classes
    Metaclass       string `json:"metaclass,omitempty"`
    IsDataclass     bool   `json:"is_dataclass,omitempty"`
}
```

### JavaScript/TypeScript Extensions
```go
type JavaScriptExtensions struct {
    // For functions
    IsArrowFunction     bool `json:"is_arrow_function,omitempty"`
    IsGeneratorFunction bool `json:"is_generator_function,omitempty"`
    
    // For classes
    IsAbstractClass bool `json:"is_abstract_class,omitempty"`
}

type TypeScriptExtensions struct {
    JavaScriptExtensions
    
    // TypeScript specific
    IsNamespace bool `json:"is_namespace,omitempty"`
    IsModule    bool `json:"is_module,omitempty"`
}
```

### Other Language Extensions
```go
type JavaExtensions struct {
    IsRecord bool `json:"is_record,omitempty"`
    IsSealed bool `json:"is_sealed,omitempty"`
}

type CSharpExtensions struct {
    IsPartial      bool `json:"is_partial,omitempty"`
    IsRecord       bool `json:"is_record,omitempty"`
    HasInitOnly    bool `json:"has_init_only,omitempty"`
}

type RustExtensions struct {
    IsUnsafe bool   `json:"is_unsafe,omitempty"`
    Lifetime string `json:"lifetime,omitempty"`
}
```

## Serialization Example

```json
{
  "kind": "function",
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
  "symbol_id": "ProcessData_42_1",
  "parameters": [{
    "name": "input",
    "type": {
      "kind": "array",
      "element_type": {
        "kind": "named",
        "name": "byte"
      }
    }
  }],
  "return_type": {
    "kind": "tuple",
    "types": [{
      "kind": "pointer",
      "pointee": {
        "kind": "named",
        "name": "Result"
      }
    }, {
      "kind": "named",
      "name": "error",
      "is_builtin": true
    }]
  },
  "body": {
    "start_line": 43,
    "end_line": 55,
    "complexity": 8,
    "called_functions": [
      {"id": "validateInput_12_3", "name": "validateInput"},
      {"id": "performCalculation_87_5", "name": "performCalculation"}
    ],
    "has_error_handling": true
  }
}
```

## Versioning

The IR schema follows semantic versioning:

- **Major version**: Breaking changes to existing node types
- **Minor version**: New node types or optional fields
- **Patch version**: Documentation or clarification changes

Current version: **2.0.0**

## Implementation Notes

1. **Immutability**: Nodes should be constructed once and never modified. Any transformation should create new nodes.

2. **Symbol Resolution**: The SymbolTable in DistilledFile enables cross-references without string matching.

3. **Type Representation**: The structured type system eliminates ambiguity and enables rich analysis.

4. **Declaration Focus**: Function bodies contain summaries, not full ASTs, aligning with our LLM use case.

5. **Extensibility**: New node types can be added without breaking the visitor pattern.

## Migration from v1.0

Key changes from v1.0:
- Structured type system replacing string types
- Symbol resolution via SymbolID
- Simplified visitor pattern
- Language extensions replacing generic attributes
- Function body summaries instead of full AST
- Enhanced annotation support