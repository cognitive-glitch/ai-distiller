# Kotlin Language Support

AI Distiller provides comprehensive support for Kotlin codebases using the [tree-sitter-kotlin](https://github.com/fwcd/tree-sitter-kotlin) parser, with support for modern Kotlin features.

## Overview

Kotlin support in AI Distiller is designed to extract the complete structure of Kotlin code while preserving null safety, immutability, and functional programming constructs. The distilled output maintains Kotlin's expressive syntax and type safety while optimizing for AI consumption.

## Recent Improvements (2025-01)

- **Val/Var Distinction**: Fixed critical issue where all properties were shown as `var` instead of preserving `val`
- **Null Safety**: Fixed preservation of nullable types (`String?`, `T?`)
- **Generic Functions**: Added support for generic type parameters in functions (`<T>`)
- **Object Declarations**: Fixed formatting to show `object` keyword instead of `class`
- **Enum Values**: Fixed to show all enum values (was only showing first)
- **Const Properties**: Added support for `const val` declarations

## Supported Kotlin Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including nested and inner classes |
| **Data Classes** | ✅ Full | Primary constructor properties preserved |
| **Sealed Classes** | ✅ Full | Including all variants |
| **Objects** | ✅ Full | Singleton objects and companion objects |
| **Interfaces** | ✅ Full | Including default methods |
| **Enums** | ✅ Full | With constructor parameters and methods |
| **Functions** | ✅ Full | Including suspend, inline, extension |
| **Properties** | ✅ Full | Val/var distinction preserved |
| **Generic Types** | ✅ Full | Type parameters and constraints |
| **Nullable Types** | ✅ Full | `?` suffix preserved |
| **Extension Functions** | ✅ Full | Receiver type shown in name |
| **Extension Properties** | ⚠️ Partial | Missing receiver type and generics |
| **Type Aliases** | ✅ Full | Type alias declarations |
| **Annotations** | ✅ Full | Custom annotations preserved |
| **Lambda Types** | ⚠️ Partial | Function types simplified |
| **Coroutines** | ✅ Full | `suspend` modifier preserved |
| **Inline Functions** | ✅ Full | `inline` modifier preserved |

### Visibility Modifiers

Kotlin visibility in AI Distiller:
- **public**: Default visibility (not shown explicitly)
- **private**: Member-only access
- **protected**: Class and subclass access
- **internal**: Module-wide access

## Known Issues

### Extension Properties
Extension properties are not fully supported. They appear as regular properties without the receiver type:
```kotlin
// Source
val <T> List<T>.secondOrNull: T?
    get() = if (size >= 2) this[1] else null

// Current output (incorrect)
var secondOrNull: T?

// Expected output
val <T> List<T>.secondOrNull: T?
```

### Companion Object Naming
Companion objects always show as `Companion` even when they have custom names.

### Default Parameter Values
Default parameter values in functions are not extracted.

## Key Features

### 1. **Immutability Preservation**

AI Distiller correctly distinguishes between `val` (immutable) and `var` (mutable):

```kotlin
// Input
data class User(
    val id: Long,
    val name: String,
    var lastLoginTime: LocalDateTime?
)

// Output
data class User(val id: Long, val name: String, var lastLoginTime: LocalDateTime?)
```

### 2. **Null Safety**

Nullable types are preserved with the `?` suffix:

```kotlin
// Input
fun findUser(id: Long): User? {
    return users.find { it.id == id }
}

// Output
fun findUser(id: Long): User?
```

### 3. **Sealed Class Hierarchies**

Sealed classes and all their variants are properly extracted:

```kotlin
// Input
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val message: String) : Result<Nothing>()
    object Loading : Result<Nothing>()
}

// Output
sealed class Result<out T> {
    data class Success<T>(val data: T)
    data class Error(val message: String)
    object Loading
}
```

### 4. **Object Declarations**

Kotlin objects (singletons) are correctly identified:

```kotlin
// Input
object AppConfig {
    const val API_KEY = "xyz"
    val baseUrl = "https://api.example.com"
    
    fun initialize() { }
}

// Output
object AppConfig {
    const val API_KEY
    val baseUrl
    fun initialize()
}
```

## Output Formats

### Text Format (Recommended for AI)

The text format preserves idiomatic Kotlin syntax:

```kotlin
// Input file
package com.example

import kotlinx.coroutines.*

interface Repository<T> {
    suspend fun get(id: Long): T?
    suspend fun save(item: T)
}

class UserRepository : Repository<User> {
    private val cache = mutableMapOf<Long, User>()
    
    override suspend fun get(id: Long): User? {
        return cache[id] ?: fetchFromDatabase(id)
    }
    
    override suspend fun save(item: User) {
        cache[item.id] = item
        saveToDatabase(item)
    }
    
    private suspend fun fetchFromDatabase(id: Long): User? = 
        withContext(Dispatchers.IO) {
            // Database logic
        }
}

// Output (default - public only, no implementation)
<file path="UserRepository.kt">
import kotlinx.coroutines.*
interface Repository<T> {
    suspend fun get(id: Long): T?
    suspend fun save(item: T)
}
class UserRepository : Repository<User> {
    suspend fun get(id: Long): User?
    suspend fun save(item: User)
}
</file>
```

## Best Practices

### 1. **Use Explicit Type Annotations**

While Kotlin has excellent type inference, explicit types help AI understand the code better:

```kotlin
// Good - explicit types
val users: List<User> = repository.getAllUsers()
val count: Int = users.size

// Less optimal - relies on inference
val users = repository.getAllUsers()
val count = users.size
```

### 2. **Leverage Kotlin Idioms**

Use Kotlin's expressive features that are well-preserved:

```kotlin
// Data classes for DTOs
data class UserDto(val id: Long, val name: String)

// Sealed classes for state
sealed class ViewState {
    object Loading : ViewState()
    data class Success(val data: List<Item>) : ViewState()
    data class Error(val message: String) : ViewState()
}

// Extension functions for utilities
fun String.isValidEmail(): Boolean = 
    contains("@") && contains(".")
```

### 3. **Structure for Clarity**

- Use meaningful names for type parameters
- Group related functionality in companion objects
- Use explicit visibility modifiers when needed
- Prefer immutability (`val` over `var`)

## Integration Examples

### Direct CLI Usage

```bash
# Generate optimal context for AI assistants
aid ./src --format text \
    --private=0 --protected=0 --internal=0 --implementation=0 \
    --output context.txt

# Include internal APIs
aid ./src --format text --internal=1 --output full-api.txt
```

### CI/CD Integration

```yaml
# .github/workflows/api-check.yml
- name: Extract Public API
  run: |
    aid ./src --format text --private=0 --protected=0 --internal=0 > api.txt
    
    # Check for breaking changes
    if ! diff api-baseline.txt api.txt; then
      echo "::error::Public API has changed"
      exit 1
    fi
```

## Comparison with Other Tools

| Tool | Purpose | Kotlin Support | AI-Optimized |
|------|---------|---------------|--------------|
| **AI Distiller** | Code structure extraction | Modern Kotlin | ✅ Yes |
| Dokka | Documentation generation | Full | ❌ No |
| detekt | Static analysis | Full | ❌ No |
| ktlint | Code formatting | Full | ❌ No |

## Future Enhancements

- [ ] Full extension property support with receiver types
- [ ] Default parameter value extraction
- [ ] Inline class support
- [ ] Contracts and experimental features
- [ ] Multiplatform annotations

## Contributing

Help improve Kotlin support! Key areas:
- Extension property parsing
- Complex generic scenarios
- Coroutine flow types
- Inline/value classes

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>