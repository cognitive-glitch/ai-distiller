# Swift Language Support

AI Distiller provides support for Swift source code processing, enabling extraction of code structure for AI consumption.

## Overview

Swift is a powerful and intuitive programming language developed by Apple for building apps for iOS, Mac, Apple TV, and Apple Watch. AI Distiller can process Swift code to extract:

- Functions and methods with their signatures
- Classes, structs, enums, and protocols
- Properties and their types
- Access control modifiers
- Protocol conformances and inheritance
- Extensions
- Actors (Swift 5.5+)

## Default Behavior

When running `aid` on Swift files without any `--strip` options, you get the complete code structure including:

- All declarations (public, internal, private, fileprivate)
- Full implementations
- Comments and documentation
- Import statements

## Stripping Options

The `--strip` flag allows you to control what gets removed:

- `--strip non-public`: Removes private, fileprivate, and internal declarations
- `--strip implementation`: Removes function/method bodies, keeping only signatures
- `--strip comments`: Removes all comments including documentation comments
- `--strip imports`: Removes import statements

## Key Features

### Visibility Handling

Swift has a sophisticated access control system:
- `open` and `public`: Accessible from any module
- `internal`: Accessible within the defining module (default)
- `fileprivate`: Accessible within the defining file
- `private`: Accessible within the enclosing declaration

The text formatter uses symbols to indicate visibility:
- `+` for public/open
- `~` for internal
- `-` for private/fileprivate

### Type Inference

Swift's strong type inference means many declarations don't have explicit types. AI Distiller attempts to preserve type information where it's explicitly stated.

### Protocol-Oriented Programming

Swift emphasizes protocol-oriented programming. AI Distiller recognizes:
- Protocol declarations
- Protocol conformances
- Protocol extensions with default implementations

## Known Issues and Limitations

### Current Implementation Status

⚠️ **Important**: Swift support uses a tree-sitter parser with basic functionality implemented. The following features are currently supported:

**Supported Features:**
- Function declarations with parameters and return types
- Class and struct declarations
- Property declarations (with limited type inference)
- Basic visibility modifiers (public, internal, private, fileprivate)
- Comments and documentation
- Basic protocol declarations
- Extensions (with limited type resolution)

**Not Yet Implemented:**
1. **Advanced Type Features**
   - Generic parameters and constraints
   - Associated types in protocols
   - Type aliases with complex types
   - Opaque return types (some keyword)

2. **Swift-Specific Features**
   - Property wrappers (@State, @Published, etc.)
   - Result builders (@ViewBuilder, etc.)
   - Actor isolation annotations
   - Async/await throws annotations
   - Enum associated values

3. **Complex Syntax**
   - Where clauses in extensions and generics
   - Conditional compilation blocks (#if/#endif)
   - Attributes beyond basic visibility

### Fallback Mechanism

If the tree-sitter parser encounters an error, the processor falls back to a simpler line-based parser. This ensures that files are always processed, though with reduced accuracy.

## Examples

<details open><summary>Basic Function</summary><blockquote>
  <details><summary>input.swift</summary><blockquote>
    
```swift
public func calculateSum(a: Int, b: Int) -> Int {
    return a + b
}

private func helperFunction() {
    print("Helper")
}
```
    
  </blockquote></details>
  <details open><summary>Default output</summary><blockquote>
    
```
<file path="input.swift">
+calculateSum(a: Int, b: Int) -> Int {
    return a + b
}
-helperFunction() {
    print("Helper")
}
</file>
```
    
  </blockquote></details>
  <details><summary>With `--strip non-public`</summary><blockquote>
    
```
<file path="input.swift">
+calculateSum(a: Int, b: Int) -> Int {
    return a + b
}
</file>
```
    
  </blockquote></details>
  <details><summary>With `--strip implementation`</summary><blockquote>
    
```
<file path="input.swift">
+calculateSum(a: Int, b: Int) -> Int
-helperFunction()
</file>
```
    
  </blockquote></details>
</blockquote></details>

<details><summary>Class with Properties</summary><blockquote>
  <details><summary>User.swift</summary><blockquote>
    
```swift
public class User {
    public let id: String
    private var name: String
    internal var email: String?
    
    public init(id: String, name: String) {
        self.id = id
        self.name = name
    }
    
    public func updateName(_ newName: String) {
        self.name = newName
    }
    
    private func validate() -> Bool {
        return !name.isEmpty
    }
}
```
    
  </blockquote></details>
  <details open><summary>Default output</summary><blockquote>
    
```
<file path="User.swift">
+class User {
    +let id: String
    -var name: String
    ~var email: String?
    
    +init(id: String, name: String) {
        self.id = id
        self.name = name
    }
    
    +updateName(_ newName: String) {
        self.name = newName
    }
    
    -validate() -> Bool {
        return !name.isEmpty
    }
}
</file>
```
    
  </blockquote></details>
  <details><summary>With `--strip non-public,implementation`</summary><blockquote>
    
```
<file path="User.swift">
+class User {
    +let id: String
    
    +init(id: String, name: String)
    +updateName(_ newName: String)
}
</file>
```
    
  </blockquote></details>
</blockquote></details>

<details><summary>Protocol and Conformance</summary><blockquote>
  <details><summary>Drawable.swift</summary><blockquote>
    
```swift
public protocol Drawable {
    func draw()
    var color: String { get set }
}

public struct Circle: Drawable {
    public var radius: Double
    public var color: String
    
    public func draw() {
        print("Drawing circle with radius \(radius)")
    }
}
```
    
  </blockquote></details>
  <details open><summary>Default output</summary><blockquote>
    
```
<file path="Drawable.swift">
+protocol Drawable {
    func draw()
    var color: String { get set }
}

+struct Circle: Drawable {
    +var radius: Double
    +var color: String
    
    +draw() {
        print("Drawing circle with radius \(radius)")
    }
}
</file>
```
    
  </blockquote></details>
</blockquote></details>

## Best Practices

1. **Type Annotations**: While Swift has excellent type inference, explicit type annotations help AI Distiller provide more accurate output
2. **Access Control**: Use explicit access control modifiers to clearly indicate API boundaries
3. **Documentation**: Use `///` documentation comments for public APIs - these are preserved unless stripped
4. **Organization**: Use `// MARK:` comments to organize code sections

## Integration Tips

When using AI Distiller output with LLMs:

1. **Use `--strip non-public,implementation`** for API overview
2. **Include full output** when debugging implementation details
3. **Consider file size** - Swift files can be large, especially with generics

## Future Roadmap

1. **Tree-sitter Integration**: Full AST-based parsing for accurate code analysis
2. **Swift 6 Support**: Including new concurrency features and strict checking
3. **SwiftUI Support**: Better handling of property wrappers and result builders
4. **Type Resolution**: Inferring types from context where not explicitly stated
5. **Cross-file Analysis**: Understanding extensions and protocol conformances across files