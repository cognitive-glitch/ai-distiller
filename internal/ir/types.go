package ir

import (
	"encoding/json"
	"time"
)

// Location represents a position in the source file
type Location struct {
	StartLine   int `json:"start_line"`
	StartColumn int `json:"start_column"`
	EndLine     int `json:"end_line"`
	EndColumn   int `json:"end_column"`
	StartByte   int `json:"start_byte,omitempty"`
	EndByte     int `json:"end_byte,omitempty"`
}

// SymbolID uniquely identifies a symbol within a compilation unit
type SymbolID string

// SymbolRef represents a reference to a symbol
type SymbolRef struct {
	ID        SymbolID `json:"id"`
	Name      string   `json:"name"`
	Package   string   `json:"package,omitempty"`
	IsBuiltin bool     `json:"is_builtin,omitempty"`
}

// NodeKind represents the type of node
type NodeKind string

const (
	KindFile      NodeKind = "file"
	KindDirectory NodeKind = "directory"
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

// Visibility represents access control
type Visibility string

const (
	VisibilityPublic           Visibility = "public"
	VisibilityPrivate          Visibility = "private"
	VisibilityProtected        Visibility = "protected"
	VisibilityInternal         Visibility = "internal"
	VisibilityPackage          Visibility = "package"
	VisibilityFilePrivate      Visibility = "fileprivate"
	VisibilityOpen             Visibility = "open"
	VisibilityFriend           Visibility = "friend"
	VisibilityProtectedInternal Visibility = "protected internal"  // C# protected internal
	VisibilityPrivateProtected  Visibility = "private protected"   // C# private protected
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
	ModifierSealed    Modifier = "sealed"
	ModifierData      Modifier = "data"
	ModifierReified   Modifier = "reified"
	ModifierMutable   Modifier = "mut"
	ModifierPartial   Modifier = "partial"
	ModifierVolatile  Modifier = "volatile"
	ModifierTransient Modifier = "transient"
	ModifierEmbedded  Modifier = "embedded"
	ModifierActor     Modifier = "actor"
	ModifierMutating  Modifier = "mutating"
	ModifierStruct    Modifier = "struct"
)

// DistilledNode is the base interface for all IR nodes
type DistilledNode interface {
	// Accept implements the visitor pattern
	Accept(visitor Visitor) DistilledNode

	// GetLocation returns the source location of this node
	GetLocation() Location

	// GetSymbolID returns the symbol ID if this node declares a symbol
	GetSymbolID() *SymbolID

	// GetNodeKind returns the kind of node for generic handling
	GetNodeKind() NodeKind

	// GetChildren returns child nodes for traversal
	GetChildren() []DistilledNode
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	Location   Location        `json:"location"`
	SymbolID   *SymbolID       `json:"symbol_id,omitempty"`
	Extensions *NodeExtensions `json:"extensions,omitempty"`
}

// GetLocation implements DistilledNode
func (n *BaseNode) GetLocation() Location {
	return n.Location
}

// GetSymbolID implements DistilledNode
func (n *BaseNode) GetSymbolID() *SymbolID {
	return n.SymbolID
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
	Attributes map[string]any        `json:"attributes,omitempty"`
}

// Language-specific extensions
type GoExtensions struct {
	IsChannel        bool   `json:"is_channel,omitempty"`
	ChannelDirection string `json:"channel_direction,omitempty"`
	IsEmptyInterface bool   `json:"is_empty_interface,omitempty"`
	ReceiverType     string `json:"receiver_type,omitempty"`
	IsMethod         bool   `json:"is_method,omitempty"`
}

type PythonExtensions struct {
	IsGenerator     bool   `json:"is_generator,omitempty"`
	IsCoroutine     bool   `json:"is_coroutine,omitempty"`
	IsStaticMethod  bool   `json:"is_static_method,omitempty"`
	IsClassMethod   bool   `json:"is_class_method,omitempty"`
	Metaclass       string `json:"metaclass,omitempty"`
	IsDataclass     bool   `json:"is_dataclass,omitempty"`
}

type JavaScriptExtensions struct {
	IsArrowFunction     bool `json:"is_arrow_function,omitempty"`
	IsGeneratorFunction bool `json:"is_generator_function,omitempty"`
	IsAbstractClass     bool `json:"is_abstract_class,omitempty"`
}

type TypeScriptExtensions struct {
	JavaScriptExtensions
	IsNamespace bool `json:"is_namespace,omitempty"`
	IsModule    bool `json:"is_module,omitempty"`
}

type JavaExtensions struct {
	IsRecord bool `json:"is_record,omitempty"`
	IsSealed bool `json:"is_sealed,omitempty"`
}

type CSharpExtensions struct {
	IsPartial   bool `json:"is_partial,omitempty"`
	IsRecord    bool `json:"is_record,omitempty"`
	HasInitOnly bool `json:"has_init_only,omitempty"`
}

type RustExtensions struct {
	IsUnsafe bool   `json:"is_unsafe,omitempty"`
	Lifetime string `json:"lifetime,omitempty"`
}

// File-level nodes

// DistilledFile represents a single source file
type DistilledFile struct {
	BaseNode
	Path        string                       `json:"path"`
	Language    string                       `json:"language"`
	Version     string                       `json:"version"`
	Children    []DistilledNode              `json:"nodes"`
	Errors      []DistilledError             `json:"errors,omitempty"`
	SymbolTable map[SymbolID]DistilledNode  `json:"-"`
	Metadata    *FileMetadata                `json:"metadata,omitempty"`
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
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
}

// GetNodeKind implements DistilledNode for DistilledFile
func (n *DistilledFile) GetNodeKind() NodeKind {
	return KindFile
}

// GetChildren implements DistilledNode for DistilledFile
func (n *DistilledFile) GetChildren() []DistilledNode {
	return n.Children
}

// Accept implements DistilledNode for DistilledFile
func (n *DistilledFile) Accept(visitor Visitor) DistilledNode {
	return visitor.Visit(n)
}

// GetNodeKind implements DistilledNode for DistilledError
func (n *DistilledError) GetNodeKind() NodeKind {
	return KindError
}

// GetChildren implements DistilledNode for DistilledError
func (n *DistilledError) GetChildren() []DistilledNode {
	return nil
}

// Accept implements DistilledNode for DistilledError
func (n *DistilledError) Accept(visitor Visitor) DistilledNode {
	return visitor.Visit(n)
}

// MarshalJSON implements json.Marshaler for DistilledFile
func (n *DistilledFile) MarshalJSON() ([]byte, error) {
	type Alias DistilledFile
	return json.Marshal(&struct {
		Kind string `json:"kind"`
		*Alias
	}{
		Kind:  string(n.GetNodeKind()),
		Alias: (*Alias)(n),
	})
}

// MarshalJSON implements json.Marshaler for DistilledError
func (n *DistilledError) MarshalJSON() ([]byte, error) {
	type Alias DistilledError
	return json.Marshal(&struct {
		Kind string `json:"kind"`
		*Alias
	}{
		Kind:  string(n.GetNodeKind()),
		Alias: (*Alias)(n),
	})
}