# PHP PSR-19 PHPDoc Tags Support Proposal

## Problem
PHP docblock annotations (PSR-19 tags) are not shown in AI Distiller output, despite being crucial for understanding PHP code semantics, types, and API contracts.

## Solution: Hybrid Approach (Virtual Nodes + Source Provenance)

### 1. Extend DistilledField Structure

```go
type DistilledField struct {
    BaseNode
    Name       string     `json:"name"`
    Visibility Visibility `json:"visibility"`
    Modifiers  []Modifier `json:"modifiers,omitempty"`
    Type       *TypeRef   `json:"type,omitempty"`
    Value      string     `json:"value,omitempty"`
    Decorators []string   `json:"decorators,omitempty"`
    
    // NEW FIELDS:
    Origin      FieldOrigin      `json:"origin,omitempty"`      // 'code' | 'docblock'
    AccessMode  FieldAccessMode  `json:"access_mode,omitempty"` // 'read-write' | 'read-only' | 'write-only'
    Description string           `json:"description,omitempty"`  // Description from docblock
    SourceAnnotation string      `json:"source_annotation,omitempty"` // Original @property line
}

type FieldOrigin string
const (
    FieldOriginCode     FieldOrigin = "code"
    FieldOriginDocblock FieldOrigin = "docblock"
)

type FieldAccessMode string
const (
    FieldAccessReadWrite FieldAccessMode = "read-write"
    FieldAccessReadOnly  FieldAccessMode = "read-only"
    FieldAccessWriteOnly FieldAccessMode = "write-only"
)
```

### 2. Parser Implementation

The PHP tree-sitter processor should:

1. When processing a class, look for the preceding comment node
2. Parse `@property*` annotations from the docblock
3. Create virtual DistilledField nodes with:
   - `Origin: "docblock"`
   - `Visibility: "public"` (always public for magic properties)
   - `AccessMode` based on annotation type
   - Preserve original annotation in `SourceAnnotation`

### 3. Annotation Parsing Rules

Based on PSR-5 (draft) and common usage:

```
@property[-read|-write] [Type] $propertyName [Description]
```

Examples:
- `@property string $name User's full name`
- `@property-read int $id Auto-generated ID`
- `@property-write array<string, mixed> $metadata Write-only metadata`

### 4. Display Format

In text formatter output:
```
class MagicModel {
    property string $name                    // @property
    property-read int $id                    // @property-read
    property-write array<string, mixed> $metadata  // @property-write
    private array<string, mixed> $data       // actual field
    public __get(string $key): mixed
    public __set(string $key, mixed $value): void
}
```

### 5. Benefits

1. **Structured Data**: AI can query magic properties like regular fields
2. **Visibility Filtering**: Can filter by access mode (read-only, etc.)
3. **Source Preservation**: Original annotation preserved for debugging
4. **Future-Proof**: Can extend to `@method` and other annotations
5. **PSR Compliant**: Aligns with PSR-5/PSR-19 drafts

### 6. Implementation Priority

1. Start with `@property*` annotations (highest value)
2. Later extend to `@method` annotations
3. Consider `@var` for class-level properties in future

### 7. PSR-19 Tags Priority

Based on value for AI code understanding, implement in phases:

#### Phase 1: Type/Contract Tags (Highest Priority)
- `@property`, `@property-read`, `@property-write` - Magic properties
- `@method` - Magic methods
- `@param` - Enhanced parameter types (already partially supported)
- `@return` - Enhanced return types (already partially supported)
- `@throws` - Exception information
- `@var` - Variable/property types

#### Phase 2: API/Visibility Tags
- `@api` - Marks public API
- `@internal` - Non-public elements
- `@deprecated` - Obsolete elements
- `@generated` - Auto-generated code

#### Phase 3: Metadata Tags
- `@since` - Version introduced
- `@version` - Current version
- `@author` - Creator information
- `@copyright` - Ownership
- `@see` - Related elements
- `@link` - External references
- `@uses`/`@usedby` - Dependencies

#### Phase 4: Documentation Tags
- `@example` - Code examples
- `@todo` - Pending tasks
- `@package` - Package organization
- `@inheritDoc` - Inherited documentation

### 8. Implementation for Each Tag Type

#### @property* Tags (Already detailed above)
Create virtual DistilledField nodes with origin="docblock"

#### @method Tags
Create virtual DistilledFunction nodes:
```go
type DistilledFunction struct {
    // ... existing fields ...
    Origin MethodOrigin `json:"origin,omitempty"` // 'code' | 'docblock'
    SourceAnnotation string `json:"source_annotation,omitempty"`
}
```

#### @throws Tags
Add to DistilledFunction:
```go
type DistilledFunction struct {
    // ... existing fields ...
    Throws []ThrowsInfo `json:"throws,omitempty"`
}

type ThrowsInfo struct {
    Exception   string `json:"exception"`
    Description string `json:"description,omitempty"`
}
```

#### @deprecated Tags
Add to all relevant nodes:
```go
type DeprecationInfo struct {
    Version     string `json:"version,omitempty"`
    Description string `json:"description,omitempty"`
}

// Add to DistilledClass, DistilledFunction, DistilledField, etc.
Deprecated *DeprecationInfo `json:"deprecated,omitempty"`
```

### 9. Edge Cases

- **Name Conflicts**: Physical properties/methods take precedence
- **Malformed Annotations**: Log warning and skip
- **Complex Types**: Support union types, generics, nullable
- **Inheritance**: Initially skip `@inheritdoc` resolution
- **Multiple Tags**: Handle multiple `@throws`, `@see`, etc.

### 10. Benefits

This comprehensive PSR-19 support provides:
- **Complete API Understanding**: AI sees the full public interface including magic members
- **Type Safety**: Enhanced type information beyond PHP's native types
- **Deprecation Awareness**: AI can suggest modern alternatives
- **Exception Handling**: AI understands error scenarios
- **Semantic Richness**: Full context for better code understanding
- **PSR Compliance**: Aligns with PHP-FIG standards