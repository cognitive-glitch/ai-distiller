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

- `--private=0 --protected=0 --internal=0`: Removes private, fileprivate, and internal declarations
- `--implementation=0`: Removes function/method bodies, keeping only signatures
- `--comments=0`: Removes all comments including documentation comments
- `--imports=0`: Removes import statements

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

⚠️ **Important**: Swift support currently uses a line-based parser due to tree-sitter stability issues. The following features are currently supported:

**Supported Features:**
- Enum declarations with raw values (String, Int)
- Basic enum cases (associated values not fully supported)
- Class, struct, and enum type detection
- Property declarations with type annotations
- Computed properties with get/set accessors
- Basic visibility modifiers (open, public, internal, fileprivate, private)
- Comments and documentation
- Protocol declarations (requirements extraction limited)
- Extensions with naming
- Import statements
- Actor declarations

## Recent Fixes (2025-06-15)

1. **Tree-sitter segfault** (✅ Fixed)
   - **Issue**: Tree-sitter Swift parser was causing segmentation faults
   - **Fix**: Temporarily disabled tree-sitter, using improved line-based parser
   - **Impact**: Stable parsing, though with some limitations

2. **Missing `func` keyword** (✅ Fixed)
   - **Issue**: Functions were displayed without the `func` keyword
   - **Fix**: Updated Swift formatter to include proper function syntax
   - **Impact**: More recognizable Swift syntax in output

3. **Protocol inheritance syntax** (✅ Fixed)
   - **Issue**: Extra `{` appeared in protocol inheritance
   - **Fix**: Updated regex to exclude opening braces from inheritance capture
   - **Impact**: Clean protocol declarations

**Known Issues:**
1. **Line-based parser limitations**
   - Multi-line function signatures not fully captured
   - Complex generic constraints may be missed
   - Associated types in protocols not parsed

2. **Not Yet Implemented**
   - Property wrappers (@State, @Published, etc.)
   - Result builders (@ViewBuilder, etc.)
   - Async/await/throws modifiers (partially supported)
   - Where clauses in extensions and generics
   - Conditional compilation blocks (#if/#endif)
   - Subscripts
   - Type aliases

3. **Minor Issues**
   - Some computed property syntax variations
   - Complex enum associated values

### Parser Implementation

Due to stability issues with the tree-sitter Swift parser, AI Distiller currently uses an improved line-based parser that provides reliable parsing with good coverage of common Swift constructs.

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
public calculateSum() -> Int
</file>
```
    
  </blockquote></details>
  <details><summary>With `--private=0 --protected=0 --internal=0`</summary><blockquote>
    
```
<file path="input.swift">
public calculateSum() -> Int
</file>
```
    
  </blockquote></details>
  <details><summary>With `--implementation=0`</summary><blockquote>
    
```
<file path="input.swift">
public calculateSum() -> Int
private helperFunction()
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
  <details><summary>With `--private=0 --protected=0 --internal=0,implementation`</summary><blockquote>
    
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

1. **Use `--private=0 --protected=0 --internal=0,implementation`** for API overview
2. **Include full output** when debugging implementation details
3. **Consider file size** - Swift files can be large, especially with generics

## Future Roadmap

1. **Fix Critical Issues**: Function parameters, protocol requirements, generic constraints
2. **Swift 6 Support**: Including new concurrency features and strict checking
3. **SwiftUI Support**: Property wrappers, result builders, and view modifiers
4. **Type Resolution**: Better handling of type inference and associated types
5. **Cross-file Analysis**: Understanding extensions and protocol conformances across files
6. **Complete Async Support**: Full async/await/throws modifier support