# Java Language Support

AI Distiller provides support for Java 8+ codebases using the [tree-sitter-java](https://github.com/tree-sitter/tree-sitter-java) parser, with basic support for modern Java features including records, sealed classes, and pattern matching.

## Overview

Java support in AI Distiller is designed to extract the essential structure of Java code while preserving type information, visibility modifiers, and relationships between classes, interfaces, and abstract classes. The distilled output maintains Java's semantic meaning while dramatically reducing token count for AI consumption.

## Core Philosophy

AI Distiller models Java code as a semantic graph, representing not just the syntax but the relationships between entities. This document explains how Java constructs are mapped to this representation, focusing on providing AI systems with a clear understanding of your codebase's structure and API surface.

## Java Version Compatibility

- **Minimum supported**: Java 8
- **Recommended**: Java 11+
- **Modern features**: Partial support for Java 14+ (records, sealed classes, pattern matching)

## Supported Java Constructs

### Foundational Elements

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Packages** | ✅ Full | Package declarations and structure |
| **Imports** | ✅ Full | Standard and static imports |
| **Classes** | ✅ Full | Including abstract, final, nested classes |
| **Interfaces** | ✅ Full | Full interface support including sealed interfaces with permits clause |
| **Enums** | ✅ Full | Enum declarations and constants |
| **Records** | ✅ Full | Record syntax, parameters, and implementations |

### Object-Oriented Constructs

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Methods** | ✅ Full | Including constructors, overloading |
| **Fields** | ✅ Full | All modifiers (static, final, volatile, etc.) |
| **Visibility** | ✅ Full | public, private, protected, package-private |
| **Inheritance** | ✅ Full | extends and implements relationships |
| **Method Overriding** | ✅ Full | @Override annotations preserved |
| **Static Members** | ✅ Full | Static methods and fields |

### Type System Features

| Feature | Support Level | Notes |
|---------|--------------|-------|
| **Basic Types** | ✅ Full | Primitives and wrapper classes |
| **Generics** | ⚠️ Partial | Type parameters captured, multiple bounds supported, inheritance generics dropped |
| **Arrays** | ✅ Full | Array type declarations |
| **Annotations** | ✅ Full | Type/method/parameter annotations, @interface definitions with default values |
| **Lambda Expressions** | ❌ Not supported | Not parsed correctly |
| **Method References** | ❌ Not supported | Not parsed correctly |

### Visibility Rules

Java visibility in AI Distiller follows standard Java keywords in text format:
- **Public**: `public` keyword (accessible everywhere)
- **Private**: `private` keyword (class-only access)  
- **Protected**: `protected` keyword (package + subclass access)
- **Package-private**: no keyword (package-only access, Java default)

## Key Features

### 1. **Standard Java Visibility Keywords**

AI Distiller uses standard Java visibility keywords for clear representation:

```java
// Input
public class User {
    private String name;
    protected int id;
    public String getName() { return name; }
}
```

```
// Output (default version - public only)
public class User {
    public String getName();
}
```

### 2. **Method Signature Extraction**

Methods are cleanly extracted with parameter types and return types:

```java
// Input
public Optional<User> findUserById(String id, boolean includeDeleted) {
    // implementation
}
```

```
// Output (no implementation)
public Optional<User> findUserById(String id, boolean includeDeleted);
```

### 3. **Inheritance Relationships**

Class hierarchies and interface implementations are preserved:

```java
// Input
public class UserService extends BaseService implements Cacheable, Auditable {
    // implementation
}
```

```
// Output
public class UserService extends BaseService implements Cacheable, Auditable {
    // members
}
```

## Output Examples

<details open><summary>Basic Class Example</summary><blockquote>
  <details><summary>Basic.java - Source Code</summary><blockquote>
    ```java
    package com.aidi.test.basic;

    /**
     * A basic class to test fundamental parsing
     */
    public class Basic {
        private static final String GREETING_PREFIX = "Hello, ";
        
        public static void main(String[] args) {
            String world = "World";
            System.out.println(createGreeting(world));
        }
        
        private static String createGreeting(String name) {
            return GREETING_PREFIX + name;
        }
    }
    ```
  </blockquote></details>
  <details open><summary>Compact AI-friendly version (`default output (public only, no implementation)`)</summary><blockquote>
    ```
    <file path="Basic.java">
    package com.aidi.test.basic

    public class Basic {
        public static void main(String[] args);
    }
    </file>
    ```
  </blockquote></details>
  <details><summary>Full version (`--public=1 --protected=1 --internal=1 --private=1 --implementation=1`)</summary><blockquote>
    ```
    <file path="Basic.java">
    package com.aidi.test.basic
    import java.util.Objects

    public class Basic {
        private static final String GREETING_PREFIX = "Hello, ";
        public static void main(String[] args) {
            String world = "World";
            int repetitions = 3;
            for (int i = 0; i < repetitions; i++) {
                String message = createGreeting(world, i + 1);
                System.out.println(message);
            }
        }
        private static String createGreeting(String name, int number) {
            return String.format("%s%s #%d", GREETING_PREFIX, name, number);
        }
    }
    </file>
    ```
  </blockquote></details>
</blockquote></details>

<details><summary>Object-Oriented Example</summary><blockquote>
  <details><summary>SimpleOOP.java - Source Code</summary><blockquote>
    ```java
    package com.aidi.test.oop;

    import java.util.Objects;

    public class SimpleOOP {
        public final String id;
        protected String name;
        private int version;
        
        public SimpleOOP(String id, String name) {
            this.id = Objects.requireNonNull(id);
            this.name = name;
            this.version = 1;
        }
        
        public String getName() {
            return name;
        }
        
        @Override
        public String toString() {
            return "SimpleOOP{id='" + id + "', name='" + name + "'}";
        }
    }
    ```
  </blockquote></details>
  <details open><summary>Compact AI-friendly version (`default output (public only, no implementation)`)</summary><blockquote>
    ```
    <file path="SimpleOOP.java">
    package com.aidi.test.oop
    import java.util.Objects

    public class SimpleOOP {
        public final String id;
        public SimpleOOP(String id, String name);
        public SimpleOOP(String id);
        public String getName();
        public void setName(String name);
        @Override
        public String toString();
    }
    </file>
    ```
  </blockquote></details>
  <details><summary>Full version with all visibility (`--strip 'comments,implementation'`)</summary><blockquote>
    ```
    <file path="SimpleOOP.java">
    package com.aidi.test.oop
    import java.util.Objects

    public class SimpleOOP {
        public final String id;
        protected String name;
        private int version;
        private boolean dirty;
        public SimpleOOP(String id, String name);
        public SimpleOOP(String id);
        public String getName();
        public void setName(String name);
        @Override
        public String toString();
    }
    </file>
    ```
  </blockquote></details>
</blockquote></details>

## Recent Improvements (2025-01)

- **Fixed Visibility Detection**: Package-private types no longer incorrectly promoted to public (critical bug fix)
- **Enhanced Annotation Support**: @interface definitions with default values now properly displayed without redundant modifiers
- **Improved Java Modern Features**: Records, sealed interfaces with permits clause, annotation default values
- **Parameter Annotations**: Method parameter annotations are now extracted (e.g., `@NotNull U input`)
- **Generic Bounds**: Shows all bounds in generic type constraints (e.g., `T extends Number & Runnable & Serializable`)

## Known Issues

### 🔴 Critical Limitations

**Dependency Resolution**
- When public classes use package-private types (sealed interfaces, records), those dependencies are excluded from default output
- Example: `ModernJava` uses `Shape`, `Circle`, `Rectangle` but only `ModernJava` appears in public-only output
- Results in incomplete/non-compilable code representations

**Inheritance Resolution**  
- Inherited public methods from parent classes are not displayed in child classes
- Example: `UserStore extends BaseStore` - missing `save()` method from parent
- Public API surface appears incomplete for inherited classes

**Advanced Type Features**
- Wildcard generics `? extends T` and `? super T` are not yet supported
- Type inference with `var` keyword loses type information

### 🟡 Major Limitations

**Throws Clauses**
- Method `throws` declarations are not being extracted despite parser support
- This affects API completeness and exception handling documentation

**Generic Type Arguments in Inheritance**  
- Generic arguments in `extends` clauses are dropped (e.g., `extends BaseStore<User>` becomes `extends BaseStore`)

**Multiple Type Definitions**
- Parser may miss additional classes/interfaces/annotations in the same file
- Nested types and inner classes have limited support

**Lambda Expressions and Method References**
- Lambda syntax `() -> {}` is not parsed
- Method references `String::length` are not recognized
- Functional interfaces lose their context

**Modern Java Pattern Matching**
- Pattern matching syntax is not recognized
- Switch expressions are not properly handled

### 🟢 Minor Issues

**Package Declaration Formatting**
- Package declarations are missing semicolons in output

**Formatting Inconsistencies**
- Method implementations have extra indentation
- Brace placement is inconsistent with Java conventions
- Static initializer blocks format as `static;`

**JavaDoc Parsing**
- JavaDoc tags like `@param` and `@return` are included in text but not structured
- HTML tags in JavaDoc comments are preserved as-is

## Best Practices

### For Optimal Results

1. **Use explicit types**: Avoid `var` keyword where possible for better type extraction
2. **Document with JavaDoc**: Use `/** */` comments for better API documentation in output
3. **Leverage generics**: Generic type information is now fully captured and preserved
4. **Use annotations**: Annotation arguments are extracted and preserved in output

### CLI Usage Examples

```bash
# Extract public API with JavaDoc (recommended for AI)
aid src/ --strip "non-public,implementation" --format text

# Extract all structure with generics and annotations
aid MyClass.java --strip "implementation" --format text

# Complete code preservation including JavaDoc
aid . --strip "" --format text --output java-full.txt
```

## Integration Examples

### With MCP Servers

```typescript
// Query Java codebase structure
const result = await mcp.call("aid", {
  path: "./src/main/java",
  strip: ["non-public", "implementation"],
  format: "text"
});
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Extract Java API
  run: |
    aid src/main/java --strip "non-public,implementation" \
      --format text --output api-surface.txt
```

## Language-Specific Tips

- **Constructor chaining**: `this()` calls are preserved in signatures
- **Method overloading**: All overloaded variants are captured
- **Static imports**: Preserved as regular imports in output
- **Nested classes**: Basic support, complex nesting may be flattened
- **Enum methods**: Enum constants and methods are both captured

## Comparison with Other Tools

| Feature | AI Distiller | JavaDoc | IDEs | AST Tools |
|---------|-------------|---------|------|-----------|
| **AI-optimized output** | ✅ | ❌ | ❌ | ❌ |
| **Visibility notation** | ✅ UML-style | ❌ | ⚠️ Icons | ❌ |
| **Type preservation** | ✅ | ✅ | ✅ | ✅ |
| **Implementation stripping** | ✅ | ❌ | ❌ | ⚠️ |
| **Inheritance tracking** | ✅ | ⚠️ | ✅ | ✅ |
| **Compact output** | ✅ | ❌ | ❌ | ❌ |

## Troubleshooting

### Common Issues

**"Generic constraints not detailed enough"**
- Wildcard generics `? extends T` need specialized handling
- Solution: Most bounded generics `T extends Class` are now captured

**"JavaDoc tags not structured"**
- JavaDoc content is captured but tags like `@param` are in text form
- Solution: JavaDoc is now preserved, structured tag parsing is planned

**"Methods have extra indentation"**
- Formatting issue with implementation blocks
- Impact: Cosmetic only, doesn't affect parsing

**"Annotation arguments simplified"**
- Complex annotation expressions may lose some detail
- Solution: Most annotation arguments are now properly extracted

### Parser Errors

If parsing fails completely:
1. Check for valid Java syntax
2. Ensure file encoding is UTF-8
3. Try with simpler constructs first
4. Report complex cases as issues

## Future Enhancements

### Planned Improvements

- **Wildcard generics**: `? extends T` and `? super T` support
- **Structured JavaDoc**: Parsed `@param`, `@return`, `@throws` tags
- **Modern Java features**: Records, sealed classes, pattern matching
- **Lambda expressions**: Functional programming constructs
- **Annotation definitions**: Proper `@interface` parsing

### Community Contributions

We welcome contributions to improve Java support:
- Wildcard generics implementation
- JavaDoc tag structuring
- Modern Java feature support
- Test case additions

## Contributing

To improve Java language support:

1. **Add test cases**: Create complex Java examples in `test-data/java/`
2. **Report issues**: Document parsing failures with minimal reproductions
3. **Enhance parser**: Improve tree-sitter integration for missing features
4. **Update docs**: Keep this documentation current with capabilities

For questions or contributions, see the main project [README](../../README.md).