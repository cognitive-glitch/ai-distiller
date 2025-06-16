# C++ Language Support

AI Distiller provides support for C++ codebases using the [tree-sitter-cpp](https://github.com/tree-sitter/tree-sitter-cpp) parser.

## Overview

C++ support in AI Distiller is designed to extract the structure of C++ code including classes, functions, templates, and namespaces. The implementation uses tree-sitter for parsing and aims to preserve the essential structure while optimizing for AI consumption.

## Current Status (2025-01)

C++ support is currently **experimental** with several known limitations. Recent improvements include:

- **Fixed**: Constructor and destructor names are now preserved correctly
- **Fixed**: Template parameters are extracted for template classes

However, significant issues remain that affect the accuracy of the distilled output.

## Supported C++ Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ⚠️ Partial | Basic structure extracted, visibility issues |
| **Structs** | ⚠️ Partial | Treated as classes with public default |
| **Functions** | ⚠️ Partial | Standalone functions work, method names sometimes lost |
| **Templates** | ❌ Limited | Template parameters extracted for classes only |
| **Namespaces** | ❌ Not Working | Not extracted at all |
| **Inheritance** | ✅ Full | Base classes detected |
| **Enums** | ✅ Full | Enum values extracted |
| **Unions** | ⚠️ Partial | Basic support |
| **Constructors** | ✅ Full | Now preserves correct names |
| **Destructors** | ✅ Full | Now preserves correct names |
| **Virtual Functions** | ⚠️ Partial | Virtual keyword detected, pure virtual syntax issues |
| **Operator Overloading** | ❌ Not Tested | Unknown support level |
| **Friend Functions** | ⚠️ Partial | Shown as comments |

### Visibility Modifiers

C++ visibility is partially supported but not correctly represented in output:
- **public**: Should be default for structs
- **private**: Should be default for classes
- **protected**: Detected but not shown

Currently, visibility sections are not properly grouped with `public:`, `private:`, `protected:` labels.

## Known Critical Issues

### 1. Type Information Loss
The most severe issue - types are often replaced with `auto` or omitted entirely:
```cpp
// Source
private:
    std::string name_;
    std::vector<int> values_;

// Current output (incorrect)
auto name_;
auto values_;
```

### 2. Template Syntax Not Parsed
Template declarations are missing for functions and methods:
```cpp
// Source
template<typename T>
T max(const T& a, const T& b);

// Current output (incorrect)
T max(T& a, T& b);
```

### 3. Function Names Lost
Methods with templated return types lose their names:
```cpp
// Source
const T& getValue() const { return value_; }

// Current output (incorrect)
T () const
```

### 4. Namespace Support Missing
Namespaces are completely ignored in the output.

### 5. Constructor/Destructor Syntax
While names are now preserved, the syntax representation has issues:
```cpp
// Current output (problematic)
void Point()
void ~Point()

// Should be
Point()
~Point()
```

## Output Example

Given this C++ source:
```cpp
namespace Example {
    template<typename T>
    class Container {
    public:
        Container(const T& value) : value_(value) {}
        const T& getValue() const { return value_; }
        
    private:
        T value_;
    };
}
```

Current output (with issues):
```
template<typename T>
class Container {
    void Container(T& value)
    T () const
    T value_;
};
```

## Best Practices

Given the current limitations, when using AI Distiller with C++:

1. **Use explicit type annotations** - Avoid auto and complex template metaprogramming
2. **Keep templates simple** - Basic template classes work better than complex SFINAE
3. **Avoid namespaces** - They won't be preserved in output
4. **Use clear naming** - Helps when function names might be lost
5. **Consider alternative tools** - For production C++ analysis, consider more mature tools

## Integration Examples

### Basic Usage

```bash
# Extract public API only (default)
aid source.cpp --format text

# Include all members (helps with visibility issues)
aid source.cpp --private=1 --protected=1 --internal=1

# With implementation bodies
aid source.cpp --implementation=1
```

## Future Improvements Needed

1. **Complete type resolution** - Preserve all type information
2. **Full template support** - Handle all template syntax correctly
3. **Namespace extraction** - Support nested namespaces
4. **Visibility sections** - Group members by access level
5. **Modern C++ features** - Support C++11/14/17/20 features
6. **Macro handling** - Process preprocessor directives
7. **Concept support** - Handle C++20 concepts

## Contributing

C++ support needs significant improvements. Key areas:
- Fix type information extraction in tree-sitter processor
- Implement proper template parameter parsing
- Add namespace support
- Fix method name extraction for templated returns
- Improve visibility handling

## Limitations

**Current C++ support should be considered experimental and not suitable for production use.** The parser has fundamental issues that result in incorrect or incomplete output. For reliable C++ code analysis, consider using established tools like:
- Doxygen
- CppDepend
- Understand
- SourceTrail

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>