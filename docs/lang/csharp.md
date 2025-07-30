# C# Language Support

AI Distiller provides comprehensive support for C# codebases using the [tree-sitter-c-sharp](https://github.com/tree-sitter/tree-sitter-c-sharp) parser, with full support for modern C# features up to C# 12.

## Overview

C# support in AI Distiller is designed to extract the complete structure of .NET code while preserving all type information, generic constraints, and modern language features. The distilled output maintains C#'s strong typing and object-oriented design while optimizing for AI consumption.

## Recent Improvements (2025-06-15)

1. **Operator support** (✅ Fixed)
   - **Issue**: User-defined operators were not extracted
   - **Fix**: Added processOperatorDeclaration to handle operator parsing
   - **Impact**: Operators now properly shown in output

2. **Record struct syntax** (✅ Fixed)
   - **Issue**: Record structs were displayed as regular records
   - **Fix**: Added ModifierStruct to distinguish record structs from record classes
   - **Impact**: Correct `record struct` syntax in output

3. **Method Parameter Fix** (Previous fix)
   - Fixed parameter name/type order for methods (was showing `instance T` instead of `T instance`)
   - Fixed method generic constraints to show full type arguments (e.g., `INumber<T>` instead of just `INumber`)
   - Improved formatting of nested record types and brace handling

## Supported C# Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ⚠️ Partial | Basic support, generic parameters on type definitions may be missing |
| **Interfaces** | ⚠️ Partial | Declaration works, but properties may be omitted |
| **Structs** | ⚠️ Partial | Basic support, `readonly` modifier not preserved |
| **Records** | ⚠️ Partial | Parameter order bug in primary constructors, calculated properties misidentified |
| **Properties** | ⚠️ Partial | Basic support, accessor visibility not preserved |
| **Methods** | ✅ Full | Including async, extension methods |
| **Generic Constraints** | ⚠️ Partial | Method constraints work, type definition constraints may be missing |
| **Nullable Reference Types** | ✅ Full | `#nullable enable/disable` directives |
| **Attributes** | ✅ Full | Method, class, property, parameter attributes |
| **Pattern Matching** | ❌ Not tested | Parser support unknown |
| **Global Usings** | ❌ Not supported | Not recognized |
| **File-scoped Namespaces** | ✅ Full | `namespace Foo;` syntax |
| **Enums** | ✅ Full | Including explicit base types |
| **Delegates** | ⚠️ Partial | Type may show as `dynamic` |
| **Events** | ❌ Not supported | Shows as `dynamic` fields |
| **Operators** | ✅ Full | User-defined operators including conversion operators |
| **Indexers** | ❌ Not tested | Support unknown |
| **Tuple Types** | ✅ Full | Named and unnamed tuples |
| **Local Functions** | ❌ Not supported | Not extracted |
| **Init-only Properties** | ✅ Full | `{ get; init; }` |

### Visibility Rules

C# visibility in AI Distiller uses the full keyword representation:
- **public**: Accessible from any code
- **private**: Accessible only within the containing type
- **protected**: Accessible within the type and derived types
- **internal**: Accessible within the same assembly
- **protected internal**: Accessible within assembly or derived types
- **private protected**: Accessible within the containing class or derived types in the same assembly

## Known Issues

### 🔴 Critical Limitations

**Record Parameter Order Bug**
- Record primary constructors show parameters in wrong order (e.g., `Id T` instead of `T Id`)
- This creates syntactically invalid C# code

**Generic Type Parameters Missing**
- Generic parameters are often missing from type definitions (e.g., `EntityBase` instead of `EntityBase<T>`)
- Generic arguments in base class lists are dropped (e.g., `IEntity` instead of `IEntity<T>`)

**Record Property Misidentification**
- Calculated properties are incorrectly included as primary constructor parameters
- Properties with bodies should not be in the parameter list

**Events**
- Events are shown as `dynamic` fields instead of proper event declarations

### 🟡 Major Limitations

**Property Accessor Visibility**
- Different visibility on getters/setters is not preserved (e.g., `{ get; protected set; }`)

**Struct Modifiers**
- `readonly` modifier on structs is not preserved

**Interface Members**
- Interface properties are sometimes omitted from output

**Optional Parameters**
- Default parameter values are not shown in method signatures

**Generic Variance**
- Variance modifiers (`in`, `out`) on generic parameters are not preserved

## Key Features

### 1. **Complete Generic Type Information**

AI Distiller preserves all generic type information including constraints:

```csharp
// Input
public interface IRepository<T> where T : class, IEntity, new()
{
    Task<T?> GetByIdAsync(int id);
    IAsyncEnumerable<T> GetAllAsync();
}

public class UserRepository : IRepository<User>
{
    public async Task<User?> GetByIdAsync(int id) { ... }
    public async IAsyncEnumerable<User> GetAllAsync() { ... }
}
```

```
// Output (with --implementation=0)
public interface IRepository<T> where T : class, IEntity, new() {
    public Task<T?> GetByIdAsync(int id);
    public IAsyncEnumerable<T> GetAllAsync();
}

public class UserRepository : IRepository<User> {
    public async Task<User?> GetByIdAsync(int id);
    public async IAsyncEnumerable<User> GetAllAsync();
}
```

### 2. **Modern C# Features Support**

Records, init-only properties, and nullable reference types:

```csharp
// Input
#nullable enable

public record UserDto(
    [property: Required] string Username,
    [property: EmailAddress] string Email)
{
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
}
```

```
// Output
#nullable enable

public record UserDto(
    [property: Required] string Username,
    [property: EmailAddress] string Email) {
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
}
```

### 3. **Property Syntax Preservation**

Properties are correctly displayed with their accessor syntax:

```csharp
// Input
public class Configuration
{
    public string ConnectionString { get; set; }
    public int Timeout { get; } = 30;
    public bool IsEnabled { get; private set; }
    public required string ApiKey { get; init; }
}
```

```
// Output
public class Configuration {
    public string ConnectionString { get; set; }
    public int Timeout { get; } = 30;
    public bool IsEnabled { get; set; }
    public required string ApiKey { get; init; }
}
```

## Output Formats

### Text Format (Recommended for AI)

The text format preserves idiomatic C# syntax with full visibility keywords:

<details open>
<summary>Example: Repository Pattern with Dependency Injection</summary>
<blockquote>

<details>
<summary>Input: `UserService.cs`</summary>
<blockquote>

```csharp
using System;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;

namespace MyApp.Services;

public interface IUserService
{
    Task<User?> GetUserAsync(int id);
    Task<bool> UpdateUserAsync(User user);
}

public sealed class UserService : IUserService
{
    private readonly IUserRepository _repository;
    private readonly ILogger<UserService> _logger;
    private readonly ICacheService _cache;

    public UserService(
        IUserRepository repository,
        ILogger<UserService> logger,
        ICacheService cache)
    {
        _repository = repository ?? throw new ArgumentNullException(nameof(repository));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        _cache = cache ?? throw new ArgumentNullException(nameof(cache));
    }

    public async Task<User?> GetUserAsync(int id)
    {
        var cacheKey = $"user:{id}";
        
        if (_cache.TryGet<User>(cacheKey, out var cached))
        {
            _logger.LogDebug("User {UserId} found in cache", id);
            return cached;
        }

        var user = await _repository.GetByIdAsync(id);
        if (user != null)
        {
            await _cache.SetAsync(cacheKey, user, TimeSpan.FromMinutes(5));
        }

        return user;
    }

    public async Task<bool> UpdateUserAsync(User user)
    {
        ArgumentNullException.ThrowIfNull(user);
        
        try
        {
            await _repository.UpdateAsync(user);
            await _cache.RemoveAsync($"user:{user.Id}");
            _logger.LogInformation("User {UserId} updated successfully", user.Id);
            return true;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to update user {UserId}", user.Id);
            return false;
        }
    }

    private async Task<bool> ValidateUserAsync(User user)
    {
        // Validation logic
        return await Task.FromResult(true);
    }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`default output (public only, no implementation)`)</summary>
<blockquote>

```
<file path="UserService.cs">
using System;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
namespace MyApp.Services;
public interface IUserService {
    public Task<User?> GetUserAsync(int id);
    public Task<bool> UpdateUserAsync(User user);
}
public sealed class UserService : IUserService {
    public UserService(IUserRepository repository, ILogger<UserService> logger, ICacheService cache);
    public async Task<User?> GetUserAsync(int id);
    public async Task<bool> UpdateUserAsync(User user);
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Advanced Generic Constraints and Pattern Matching</summary>
<blockquote>

<details>
<summary>Input: `MathProcessor.cs`</summary>
<blockquote>

```csharp
#nullable enable
using System;
using System.Numerics;

namespace AdvancedMath;

public interface IVectorOperations<T> where T : INumber<T>
{
    T DotProduct<TVector>(TVector a, TVector b) 
        where TVector : IVector<T>;
    
    TVector Add<TVector>(TVector a, TVector b) 
        where TVector : IVector<T>, new();
}

public class VectorProcessor<T> : IVectorOperations<T> 
    where T : INumber<T>, IMinMaxValue<T>
{
    private readonly ILogger? _logger;

    public VectorProcessor(ILogger? logger = null)
    {
        _logger = logger;
    }

    public T DotProduct<TVector>(TVector a, TVector b) 
        where TVector : IVector<T>
    {
        if (a.Dimension != b.Dimension)
            throw new ArgumentException("Vectors must have same dimension");

        var result = T.Zero;
        for (int i = 0; i < a.Dimension; i++)
        {
            result += a[i] * b[i];
        }
        
        _logger?.Log($"Dot product calculated: {result}");
        return result;
    }

    public TVector Add<TVector>(TVector a, TVector b) 
        where TVector : IVector<T>, new()
    {
        var result = new TVector { Dimension = a.Dimension };
        for (int i = 0; i < a.Dimension; i++)
        {
            result[i] = a[i] + b[i];
        }
        return result;
    }

    protected virtual T ClampValue(T value)
    {
        return value switch
        {
            _ when value > T.MaxValue => T.MaxValue,
            _ when value < T.MinValue => T.MinValue,
            _ => value
        };
    }

    private static bool IsValidDimension(int dimension) => dimension > 0;
}
```

</blockquote>
</details>

<details open>
<summary>Default Output</summary>
<blockquote>

```
<file path="MathProcessor.cs">
#nullable enable

using System;
using System.Numerics;
namespace AdvancedMath;
public interface IVectorOperations<T> where T : INumber {
    public T DotProduct<TVector>(TVector a, TVector b) where TVector : IVector;
    public TVector Add<TVector>(TVector a, TVector b) where TVector : IVector, new();
}
public class VectorProcessor<T> : IVectorOperations<T> where T : INumber, IMinMaxValue {
    public VectorProcessor(ILogger? logger = null);
    public T DotProduct<TVector>(TVector a, TVector b) where TVector : IVector;
    public TVector Add<TVector>(TVector a, TVector b) where TVector : IVector, new();
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Modern Records and Pattern Matching</summary>
<blockquote>

<details>
<summary>Input: `Commands.cs`</summary>
<blockquote>

```csharp
#nullable enable
using System;
using System.ComponentModel.DataAnnotations;

namespace MyApp.Commands;

public abstract record Command
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public DateTime Timestamp { get; init; } = DateTime.UtcNow;
}

public sealed record CreateUserCommand(
    [property: Required] string Username,
    [property: EmailAddress] string Email,
    [property: StringLength(100)] string? FullName = null) : Command
{
    public bool IsValid => !string.IsNullOrWhiteSpace(Username) && 
                          !string.IsNullOrWhiteSpace(Email);
}

public sealed record UpdateUserCommand(
    [property: Required] Guid UserId,
    [property: StringLength(100)] string? FullName,
    [property: Phone] string? PhoneNumber) : Command;

public sealed record DeleteUserCommand(
    [property: Required] Guid UserId,
    [property: Required] string Reason) : Command;

public static class CommandHandler
{
    public static async Task<CommandResult> HandleAsync(Command command) => command switch
    {
        CreateUserCommand { IsValid: true } cmd => await CreateUserAsync(cmd),
        CreateUserCommand { IsValid: false } => CommandResult.ValidationFailed("Invalid user data"),
        UpdateUserCommand cmd => await UpdateUserAsync(cmd),
        DeleteUserCommand { Reason.Length: > 10 } cmd => await DeleteUserAsync(cmd),
        DeleteUserCommand => CommandResult.ValidationFailed("Deletion reason too short"),
        null => throw new ArgumentNullException(nameof(command)),
        _ => CommandResult.NotSupported($"Command type {command.GetType().Name} not supported")
    };

    private static async Task<CommandResult> CreateUserAsync(CreateUserCommand cmd)
    {
        // Implementation
        return await Task.FromResult(CommandResult.Success(cmd.Id));
    }

    private static async Task<CommandResult> UpdateUserAsync(UpdateUserCommand cmd)
    {
        // Implementation
        return await Task.FromResult(CommandResult.Success(cmd.Id));
    }

    private static async Task<CommandResult> DeleteUserAsync(DeleteUserCommand cmd)
    {
        // Implementation  
        return await Task.FromResult(CommandResult.Success(cmd.Id));
    }
}

public record CommandResult(bool IsSuccess, Guid? CommandId, string? Error)
{
    public static CommandResult Success(Guid commandId) => new(true, commandId, null);
    public static CommandResult ValidationFailed(string error) => new(false, null, error);
    public static CommandResult NotSupported(string error) => new(false, null, error);
}
```

</blockquote>
</details>

<details open>
<summary>Default Output</summary>
<blockquote>

```
<file path="Commands.cs">
#nullable enable

using System;
using System.ComponentModel.DataAnnotations;
namespace MyApp.Commands;
public abstract record Command {
    public Guid Id { get; init; } = Guid.NewGuid();
    public DateTime Timestamp { get; init; } = DateTime.UtcNow;
}
public sealed record CreateUserCommand(
    [property: Required] string Username,
    [property: EmailAddress] string Email,
    [property: StringLength(100)] string? FullName = null) : Command {
    public bool IsValid { get; }
}
public sealed record UpdateUserCommand(
    [property: Required] Guid UserId,
    [property: StringLength(100)] string? FullName,
    [property: Phone] string? PhoneNumber) : Command;
public sealed record DeleteUserCommand(
    [property: Required] Guid UserId,
    [property: Required] string Reason) : Command;
public static class CommandHandler {
    public static async Task<CommandResult> HandleAsync(Command command);
}
public record CommandResult(bool IsSuccess, Guid? CommandId, string? Error) {
    public static CommandResult Success(Guid commandId);
    public static CommandResult ValidationFailed(string error);
    public static CommandResult NotSupported(string error);
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

## Best Practices

### 1. **Use Full Type Annotations**

AI Distiller preserves all type information:

```csharp
// Good - Complete type information
public async Task<Result<User>> GetUserAsync(
    int id, 
    CancellationToken cancellationToken = default)
{
    // ...
}

// Less optimal - Generic type info may be incomplete  
public async Task GetUserAsync(int id)
{
    // ...
}
```

### 2. **Leverage Modern C# Features**

Modern syntax is fully supported and recommended:

```csharp
// File-scoped namespaces (C# 10+)
namespace MyApp.Services;

// Global usings (C# 10+)
global using System.Text.Json;

// Records with primary constructors (C# 9+)
public record UserDto(string Name, string Email);

// Pattern matching (C# 8+)
var message = user switch
{
    { IsActive: true } => "Active user",
    { IsActive: false } => "Inactive user",
    null => "User not found"
};
```

### 3. **Structure for AI Understanding**

- Use clear, descriptive names
- Add XML documentation comments
- Group related functionality
- Use interfaces for contracts
- Apply consistent visibility modifiers

### 4. **Optimize for Distillation**

```csharp
// Good - Clear API surface
public interface IUserService
{
    Task<User?> GetUserAsync(int id);
    Task<IEnumerable<User>> GetActiveUsersAsync();
    Task<bool> DeactivateUserAsync(int id);
}

// Implementation details are stripped by default
public class UserService : IUserService
{
    private readonly IDbContext _context;
    private readonly ICache _cache;
    
    // Public methods define the API
    public async Task<User?> GetUserAsync(int id) { }
    
    // Private methods are stripped by default
    private async Task<bool> ValidateUserAsync(User user) { }
}
```

## Integration Examples

### Direct CLI Usage

```bash
# Generate optimal context for AI assistants
aid ./src --format text \
    --private=0 --protected=0 --internal=0 --implementation=0 \
    --output context.txt

# Include in your prompt
cat <<EOF > prompt.txt
Here's my C# codebase structure:

$(cat context.txt)

Please help me implement a new feature that...
EOF
```

### MCP Integration with Claude Desktop

When using AI Distiller through MCP, Claude can directly access your C# codebase:

**Setup** (in Claude Desktop config.json):
```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "/usr/local/bin/aid",
      "args": ["--mcp-server"],
      "env": {
        "AID_ROOT": "/Users/you/projects/MyDotNetApp"
      }
    }
  }
}
```

**Example Conversation:**

> **You**: Show me the structure of our payment processing system and identify potential improvements.

> **Claude**: I'll analyze your payment processing system using AI Distiller.

*Claude automatically calls:*
```
distillDirectory("Services/Payment/", include_implementation=false)
listFiles("Services/Payment/", pattern="*.cs")
```

*Claude receives the distilled structure and responds:*

> Based on the analysis, your payment processing system has:
> 
> **Core Components:**
> - `IPaymentProcessor` interface with methods: `ProcessPaymentAsync()`, `RefundAsync()`, `GetTransactionStatusAsync()`
> - `StripePaymentProcessor : IPaymentProcessor` - Stripe implementation
> - `PayPalPaymentProcessor : IPaymentProcessor` - PayPal implementation
> - `PaymentService` - Orchestrator using strategy pattern
> 
> **Identified Improvements:**
> 1. **Missing Circuit Breaker**: Add resilience for external API calls
> 2. **No Idempotency**: Payment operations should be idempotent
> 3. **Logging Enhancement**: Current logging lacks structured data
> 
> Here's how to implement these improvements...

### With Source Generators

```bash
# Extract metadata for source generators
aid ./Models --format json-structured --private=0 --protected=0 --internal=0 | \
  jq '.files[].symbols[] | select(.type == "class" and .attributes[] | contains("GenerateDto"))'
```

### CI/CD Integration

```yaml
# .github/workflows/api-surface-check.yml
- name: Check Public API Surface
  run: |
    aid ./src --format text --private=0 --protected=0 --internal=0 --implementation=0 > api-current.txt
    
    # Compare with baseline
    if ! diff api-baseline.txt api-current.txt > api-changes.diff; then
      echo "::warning::Public API surface has changed"
      cat api-changes.diff
    fi
```

## Language-Specific Tips

1. **Nullable Reference Types**: Always include `#nullable enable` directives - they're preserved as raw content

2. **Generic Constraints**: Use meaningful constraints for better AI understanding:
   ```csharp
   // Good - Clear intent
   where T : class, IEntity, new()
   
   // Better - Even more specific
   where T : BaseEntity, IValidatable, new()
   ```

3. **Records vs Classes**: Use records for DTOs and immutable data:
   ```csharp
   // DTO - use record
   public record UserDto(string Name, string Email);
   
   // Service - use class
   public class UserService : IUserService { }
   ```

4. **Extension Methods**: Mark extension classes clearly:
   ```csharp
   public static class StringExtensions
   {
       public static bool IsNullOrEmpty(this string? value) => 
           string.IsNullOrEmpty(value);
   }
   ```

## Comparison with Other Tools

| Tool | Purpose | C# Support | AI-Optimized |
|------|---------|-----------|--------------|
| **AI Distiller** | Code structure extraction | Full C# 12 | ✅ Yes |
| Roslyn | Compiler/Analysis | Native | ❌ No |
| dotnet-format | Code formatting | Full | ❌ No |
| StyleCop | Style analysis | Full | ❌ No |
| ILSpy | Decompilation | Full | ❌ No |

## Troubleshooting

### "Parser failed with syntax error"

Ensure your C# code compiles:
```bash
dotnet build
```

### "Generic constraints not showing"

Check that constraints are on the same line (for single constraints) or properly formatted:
```csharp
// Good
public class Service<T> where T : class, new()

// Also good  
public class ComplexService<T, U> 
    where T : class, IEntity
    where U : struct, IComparable<U>
```

### "Properties showing as fields"

This has been fixed in the latest version. Ensure you're using the latest AI Distiller build.

## Future Enhancements

- [ ] Source generator support
- [ ] Partial class merging
- [ ] XML documentation extraction
- [ ] Analyzer/diagnostic integration
- [ ] .NET Aspire integration
- [ ] Primary constructor body support

## Contributing

Help improve C# support! See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

Key areas needing help:
- Complex generic scenarios
- Source generator examples
- Blazor component patterns
- MAUI/WPF specific constructs

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>