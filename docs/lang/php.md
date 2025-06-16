# PHP Language Support

AI Distiller provides comprehensive support for PHP 7.4+ codebases using the [tree-sitter-php](https://github.com/tree-sitter/tree-sitter-php) parser, with full support for PHP 8.x features including enums, attributes, union types, and PSR-19 PHPDoc standards.

## Overview

PHP support in AI Distiller is designed to extract the essential structure of PHP code while preserving type information, visibility modifiers, and API contracts. The distilled output maintains PHP's semantic meaning while dramatically reducing token count for AI consumption.

## Supported PHP Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including abstract, final, readonly (8.2+) |
| **Interfaces** | ✅ Full | Multiple inheritance supported |
| **Traits** | ✅ Full | Shown as special classes with usage markers |
| **Enums** | ✅ Full | Pure and backed enums (8.1+) with values |
| **Functions** | ✅ Full | Global functions with type hints |
| **Methods** | ✅ Full | Including magic methods, final, abstract |
| **Properties** | ✅ Full | Typed, readonly, promoted, magic via @property |
| **Constants** | ✅ Full | Class constants with visibility, proper `const` syntax |
| **Namespaces** | ✅ Full | Including grouped use statements |
| **Attributes** | ✅ Full | PHP 8.0+ attributes/annotations |

### Type System Features

| Feature | Support Level | Notes |
|---------|--------------|-------|
| **Type declarations** | ✅ Full | Parameters, returns, properties |
| **Union types** | ✅ Full | PHP 8.0+ `string\|int` |
| **Intersection types** | ✅ Full | PHP 8.1+ `Countable&Iterator` |
| **Nullable types** | ✅ Full | `?string` syntax |
| **Array types** | ✅ Full | Via PHPDoc: `array<K,V>`, `list<T>`, shapes |
| **Advanced types** | ✅ Full | `class-string<T>`, `key-of<>`, `value-of<>` |
| **Default values** | ✅ Full | Parameter defaults preserved |

### PSR-19 PHPDoc Support

| Annotation | Support Level | Notes |
|------------|--------------|-------|
| **@property** | ✅ Full | Creates virtual public properties |
| **@property-read** | ✅ Full | Creates virtual read-only properties |
| **@property-write** | ✅ Full | Creates virtual write-only properties |
| **@method** | ✅ Full | Creates virtual methods |
| **@deprecated** | ✅ Full | Marks elements as deprecated |
| **@internal** | ✅ Full | API documentation preserved |
| **@param** | ✅ Full | Enhanced parameter types |
| **@return** | ✅ Full | Enhanced return types |
| **@throws** | ✅ Full | Exception documentation |
| **@template** | ⚠️ Partial | Preserved but not in API tags |
| **@psalm-type** | ❌ Excluded | User preference |
| **@phpstan-type** | ❌ Excluded | User preference |

### Visibility Rules

PHP visibility in AI Distiller follows standard PHP conventions:
- **Public**: `public` keyword or default (methods), dunder methods (`__init__`, `__toString__`)
- **Protected**: `protected` keyword
- **Private**: `private` keyword
- **Internal**: Package-private (no PHP equivalent, unused)

## Key Features

### 1. **Magic Property Support (PSR-19)**

AI Distiller transforms PHPDoc `@property*` annotations into virtual properties displayed directly in the class body:

```php
// Input
/**
 * @property-read int $id Auto-generated ID
 * @property string $name User's full name
 * @property-write array<string, mixed> $metadata
 */
class User {
    public function __get(string $key): mixed { /* ... */ }
    public function __set(string $key, mixed $value): void { /* ... */ }
}
```

```
// Output (magic methods hidden when @property exists)
/**
 * @property-read int $id Auto-generated ID
 * @property string $name User's full name
 * @property-write array<string, mixed> $metadata
 */
class User {
}
```

### 2. **Enum Support with Values**

PHP 8.1+ enums are properly displayed with the `enum` keyword and their case values are always shown:

```php
// Input
enum Status: string {
    case DRAFT = 'draft';
    case PUBLISHED = 'published';
    case ARCHIVED = 'archived';
}
```

```
// Output (values always shown, even with --implementation=0)
enum Status: string {
    case DRAFT = 'draft';
    case PUBLISHED = 'published';
    case ARCHIVED = 'archived';
}
```

### 3. **Advanced Type Preservation**

Complex PHPDoc types are fully preserved for AI understanding:

```php
// Input
/**
 * @param array{
 *   id: int,
 *   name: non-empty-string,
 *   tags: non-empty-list<string>,
 *   meta?: array<string, mixed>
 * } $data
 * @return class-string<Model>
 */
public function process(array $data): string
```

```
// Output
public process(array $data): class-string<Model>
```

### 4. **Smart Docblock Handling**

API-defining docblocks are shown even when `--comments=0`:
- Classes with `@property*`, `@method`, `@deprecated` annotations
- Methods with enhanced type information
- Docstrings and comments are properly separated

## Output Formats

### Text Format (Recommended for AI)

The text format is optimized for AI comprehension with minimal syntax:
- No visibility prefixes (unlike other languages)
- Clear type information
- Compact representation
- API docblocks preserved when relevant

<details open>
<summary>Example: PSR-19 Magic Properties</summary>
<blockquote>

<details>
<summary>Input: `magic_properties.php`</summary>
<blockquote>

```php
<?php

namespace App\Models;

/**
 * Active Record model with magic properties
 * 
 * @property-read int $id Primary key
 * @property-read \DateTime $createdAt Creation timestamp
 * @property string $name Full name
 * @property string $email Email address
 * @property-write string $password Hashed password (write-only)
 * @property-read array<string, mixed> $attributes All attributes
 * 
 * @method static self|null find(int $id)
 * @method static self[] findAll()
 * @method bool save()
 */
abstract class ActiveRecord {
    protected array $data = [];
    
    public function __construct(array $attributes = []) {
        $this->fill($attributes);
    }
    
    public function __get(string $name): mixed {
        return $this->data[$name] ?? null;
    }
    
    public function __set(string $name, mixed $value): void {
        $this->data[$name] = $value;
    }
    
    protected function fill(array $attributes): void {
        $this->data = $attributes;
    }
}

class User extends ActiveRecord {
    protected string $table = 'users';
    
    public function getFullName(): string {
        return $this->name;
    }
    
    public function isAdmin(): bool {
        return $this->role === 'admin';
    }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`public only, no implementation`)</summary>
<blockquote>

```
<file path="magic_properties.php">
namespace App\Models;

/**
 * Active Record model with magic properties
 * 
 * @property-read int $id Primary key
 * @property-read \DateTime $createdAt Creation timestamp
 * @property string $name Full name
 * @property string $email Email address
 * @property-write string $password Hashed password (write-only)
 * @property-read array<string, mixed> $attributes All attributes
 * 
 * @method static self|null find(int $id)
 * @method static self[] findAll()
 * @method bool save()
 */
abstract class ActiveRecord {
    public __construct(array $attributes = [])
    public static find(int $id): self|null
    public static findAll(): self[]
    public save(): bool
}

class User extends ActiveRecord {
    public getFullName(): string
    public isAdmin(): bool
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Modern PHP 8 Features</summary>
<blockquote>

<details>
<summary>Input: `modern_features.php`</summary>
<blockquote>

```php
<?php

declare(strict_types=1);

namespace App\Services;

use App\Contracts\{Cacheable, Loggable};
use App\Enums\Permission;

/**
 * User service with modern PHP 8 features
 */
#[\Attribute(\Attribute::TARGET_CLASS)]
class Service {
    public function __construct(
        public readonly string $name,
        public readonly int $version = 1
    ) {}
}

#[Service(name: 'notifications', version: 2)]
class NotificationService implements Cacheable, Loggable {
    public function __construct(
        private readonly DatabaseConnection $db,
        private readonly CacheInterface $cache,
        private LoggerInterface $logger,
        private bool $debug = false
    ) {}
    
    public function send(
        string|Email $message,
        User|string $recipient,
        ?array $options = null
    ): Result|false {
        // Implementation
        return new Result(true);
    }
    
    public function hasPermission(
        User $user,
        Permission $permission
    ): bool {
        return match($permission) {
            Permission::READ => true,
            Permission::WRITE => $user->isAdmin(),
            Permission::DELETE => $user->isSuperAdmin(),
        };
    }
}

enum Permission: string {
    case READ = 'read';
    case WRITE = 'write';
    case DELETE = 'delete';
}

interface CacheInterface {
    public function get(string $key): mixed;
    public function set(string $key, mixed $value, ?int $ttl = null): bool;
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`public only, no implementation`)</summary>
<blockquote>

```
<file path="modern_features.php">
namespace App\Services;
use App\Contracts\Cacheable;
use App\Contracts\Loggable;
use App\Enums\Permission;

/**
 * User service with modern PHP 8 features
 */
#[\Attribute(\Attribute::TARGET_CLASS)]
class Service {
    public readonly name: string
    public readonly version: int
    public __construct(string $name, int $version = 1)
}

#[Service(name: 'notifications', version: 2)]
class NotificationService implements Cacheable, Loggable {
    public __construct(DatabaseConnection $db, CacheInterface $cache, LoggerInterface $logger, bool $debug = false)
    public send(string|Email $message, User|string $recipient, ?array $options = null): Result|false
    public hasPermission(User $user, Permission $permission): bool
}

enum Permission: string {
    case READ = 'read';
    case WRITE = 'write';
    case DELETE = 'delete';
}

interface CacheInterface {
    public get(string $key): mixed
    public set(string $key, mixed $value, ?int $ttl = null): bool
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Complex Type Annotations</summary>
<blockquote>

<details>
<summary>Input: `complex_types.php`</summary>
<blockquote>

```php
<?php

namespace App\Types;

/**
 * Repository with complex type annotations
 */
class Repository {
    /**
     * Find entities by criteria
     * 
     * @param array{
     *   where?: array<string, mixed>,
     *   orderBy?: array<string, 'ASC'|'DESC'>,
     *   limit?: positive-int,
     *   offset?: non-negative-int
     * } $criteria
     * @return list<Entity>
     */
    public function findBy(array $criteria): array {
        // Implementation
    }
    
    /**
     * @param class-string<T> $className
     * @param array<string, mixed> $data
     * @return T
     * @template T of Entity
     */
    public function hydrate(string $className, array $data): object {
        return new $className($data);
    }
    
    /**
     * @param callable(Entity): bool $predicate
     * @return Entity|null
     */
    public function findOneBy(callable $predicate): ?Entity {
        // Implementation
    }
    
    /**
     * @param non-empty-array<int> $ids
     * @return array<int, Entity>
     */
    public function findByIds(array $ids): array {
        // Implementation
    }
    
    /**
     * @param key-of<self::ALLOWED_FIELDS> $field
     * @param value-of<self::ALLOWED_VALUES> $value
     */
    public function validateField(string $field, mixed $value): bool {
        // Implementation
    }
    
    public const ALLOWED_FIELDS = [
        'name' => true,
        'email' => true,
        'status' => true
    ];
    
    public const ALLOWED_VALUES = [
        'active',
        'inactive',
        'pending'
    ];
}
```

</blockquote>
</details>

<details open>
<summary>Default Output</summary>
<blockquote>

```
<file path="complex_types.php">
namespace App\Types;

/**
 * Repository with complex type annotations
 */
class Repository {
    public findBy(array $criteria): list<Entity>
    public hydrate(class-string<T> $className, array<string, mixed> $data): T
    public findOneBy(callable(Entity): bool $predicate): ?Entity
    public findByIds(non-empty-array<int> $ids): array<int, Entity>
    public validateField(key-of<self::ALLOWED_FIELDS> $field, value-of<self::ALLOWED_VALUES> $value): bool
    
    public const ALLOWED_FIELDS = [
        'name' => true,
        'email' => true,
        'status' => true
    ];
    
    public const ALLOWED_VALUES = [
        'active',
        'inactive',
        'pending'
    ];
}
</file>
```

</blockquote>
</details>

</blockquote>
</details>

## Known Issues

### Critical Issues

1. **Empty constructors hidden** (🟢 Intentional)
   - **Behavior**: Constructors without parameters are not shown
   - **Rationale**: Reduces noise in output
   - **Impact**: Minimal - empty constructors have no API significance

### Minor Issues

2. **Template annotations excluded** (🟢 Minor)
   - **Issue**: `@template` tags not in API-defining list
   - **Impact**: Generic type information in docblock only
   - **Workaround**: Templates are rarely used in PHP

3. **Complex array shapes in parameters** (🟡 Minor)
   - **Issue**: Array shape types shown in docblock, not signature
   - **Example**: `@param array{id: int, name: string} $data`
   - **Status**: Preserving PHPDoc is sufficient for AI

## Best Practices

### 1. **Use PSR-19 Annotations for Virtual APIs**

Define public APIs through PHPDoc when using magic methods:

```php
/**
 * @property-read int $id
 * @property string $name  
 * @method static self create(array $data)
 */
class Model {
    // Magic methods implementation
}
```

### 2. **Leverage PHP 8+ Features**

Modern PHP features provide better type safety and cleaner code:

```php
// Constructor property promotion
public function __construct(
    private readonly LoggerInterface $logger,
    private CacheInterface $cache,
    private bool $debug = false
) {}

// Union types instead of mixed
public function process(string|array $data): Result|false

// Enums with backed values
enum Status: string {
    case ACTIVE = 'active';
    case INACTIVE = 'inactive';
}
```

### 3. **Document Array Types Precisely**

Use PHPDoc to specify array structures:

```php
/**
 * @param Product[] $products Simple array of products
 * @param array<string, Product> $indexed Associative array
 * @param list<string> $tags Non-empty list
 * @param array{id: int, name: string} $user Shape definition
 */
```

### 4. **Structure for AI Consumption**

- Keep classes under 500 lines
- Use meaningful names (preserved in output!)
- Group related functionality in namespaces
- Document complex types in PHPDoc
- Hide implementation details with `--implementation=0`

## Integration Examples

### Direct CLI Usage

```bash
# Extract public API for AI context
aid ./src --format text --implementation=0 --output api.txt

# Include all members but no implementation
aid ./src --private=1 --protected=1 --implementation=0

# Full extraction with implementations
aid ./src --private=1 --protected=1 --implementation=1

# Extract specific directory
aid ./src/Models --format text > models.txt
```

### MCP Integration with Claude Desktop

When using AI Distiller through MCP, Claude can directly access your PHP codebase:

**Setup** (in Claude Desktop config.json):
```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "/usr/local/bin/aid",
      "args": ["--mcp-server"],
      "env": {
        "AID_ROOT": "/home/user/projects/my-php-app"
      }
    }
  }
}
```

**Example Conversation:**

> **You**: Analyze our user authentication system and suggest improvements for adding OAuth support.

> **Claude**: I'll analyze your authentication system structure using AI Distiller.

*Claude automatically calls:*
```
distillDirectory("src/Auth", include_private=false, include_implementation=false)
search("OAuth|oauth", is_regex=true)
distillFile("src/Auth/AuthManager.php")
```

*Claude receives the distilled structure and responds:*

> Based on the analysis, your authentication system has:
>
> **Current Structure:**
> ```
> AuthManager:
>   + authenticate(string $username, string $password): ?User
>   + createSession(User $user): Session
>   + validateToken(string $token): bool
>   
> User:
>   @property-read int $id
>   @property string $email
>   @property-write string $password
>   + verifyPassword(string $password): bool
> ```
>
> **OAuth Integration Recommendations:**
>
> 1. Create an `AuthProvider` interface:
> ```php
> interface AuthProvider {
>     public function authenticate(array $credentials): ?User;
>     public function getUserInfo(string $token): array;
> }
> ```
>
> 2. Implement OAuth providers:
> ```php
> class GoogleOAuthProvider implements AuthProvider { }
> class GitHubOAuthProvider implements AuthProvider { }
> ```
>
> 3. Extend `AuthManager` to support multiple providers...

### With Laravel/Symfony

```bash
# Extract Laravel models structure
aid app/Models --format text --implementation=0 > models-api.txt

# Symfony entities with full details  
aid src/Entity --private=1 --protected=1 > entities-full.txt

# Controllers public API only
aid src/Controller --implementation=0 > controllers-api.txt
```

### CI/CD Integration

```yaml
# .github/workflows/api-check.yml
- name: Check API Surface
  run: |
    aid src/ --implementation=0 --format text > api-current.txt
    diff api-baseline.txt api-current.txt || {
      echo "API surface changed! Review the differences:"
      diff -u api-baseline.txt api-current.txt
      exit 1
    }
```

### Composer Scripts

```json
{
  "scripts": {
    "analyze": "aid src/ --format text > structure.txt",
    "analyze:api": "aid src/ --implementation=0 > api.txt",
    "analyze:full": "aid src/ --private=1 --protected=1 --implementation=1 > full.txt"
  }
}
```

## Language-Specific Tips

1. **Magic Properties Best Practices**:
   ```php
   /**
    * Always document magic properties
    * @property-read int $computed This is computed dynamically
    * @property-write array $bulk Write-only for bulk updates
    */
   ```

2. **Enum Usage**:
   ```php
   // Backed enums are fully supported
   enum Status: string {
       case ACTIVE = 'active';
   }
   
   // Use match expressions with enums
   return match($status) {
       Status::ACTIVE => 'Running',
   };
   ```

3. **Type Documentation**:
   ```php
   /**
    * Document complex types in PHPDoc
    * @param array{
    *   id: int,
    *   items: list<Product>,
    *   total: float
    * } $order
    */
   ```

## Comparison with Other Tools

| Tool | Purpose | PHP Support | AI-Optimized |
|------|---------|-------------|--------------|
| **AI Distiller** | Code structure extraction | Full 7.4-8.3 | ✅ Yes |
| PHPStan | Static analysis | Full | ❌ No |
| PHP-Parser | AST generation | Full | ❌ No |
| phpDocumentor | Documentation | Partial | ❌ No |
| PHP CS Fixer | Code formatting | N/A | ❌ No |

## Troubleshooting

### "Magic properties not showing"

Ensure your class has @property annotations in its docblock:
```php
/**
 * @property-read int $id
 */
class Model { }
```

### "Enum showing as class"

Update to latest version. Enums now display with proper `enum` keyword.

### "Missing parameter defaults"

Default values are now preserved. Update if seeing issues.

### "Complex types not preserved"

PHPDoc types like `array<K,V>`, `class-string<T>` are fully supported in latest version.

## Future Enhancements

- [ ] Capture `@param` array shapes in method signatures
- [ ] Anonymous class support
- [ ] Closure and arrow function type extraction
- [ ] `@template` as API-defining tag
- [ ] Trait conflict resolution (`as`, `insteadof`)
- [ ] Property hooks (PHP 8.4+)

## Contributing

Help improve PHP support! See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

Key areas needing help:
- Complex PHPDoc patterns
- Framework-specific patterns
- Performance optimizations
- Real-world test cases

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>