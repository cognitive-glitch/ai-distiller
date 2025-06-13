# PHP Language Support

AI Distiller provides comprehensive support for PHP 7.4+ codebases using the [tree-sitter-php](https://github.com/tree-sitter/tree-sitter-php) parser, with full support for PHP 8.x features including attributes, union types, and constructor property promotion.

## Overview

PHP support in AI Distiller is designed to extract the essential structure of PHP code while preserving type information, visibility modifiers, and relationships between classes, interfaces, and traits. The distilled output maintains PHP's semantic meaning while dramatically reducing token count for AI consumption.

## Core Philosophy

AI Distiller models PHP code as a semantic graph, representing not just the syntax but the relationships between entities. This document explains how PHP constructs are mapped to this representation, focusing on providing AI systems with a clear understanding of your codebase's structure and API surface.

## PHP Version Compatibility

- **Minimum supported**: PHP 7.4
- **Recommended**: PHP 8.0+
- **Fully supported features**: PHP 8.0 (union types, attributes, constructor property promotion), PHP 8.1 (enums, readonly properties, intersection types), PHP 8.2 (readonly classes)

## Supported PHP Constructs

### Foundational Elements

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Namespaces** | ✅ Full | Including grouped use statements |
| **Use statements** | ✅ Full | Classes, functions, constants, aliases |
| **Strict types** | ✅ Full | `declare(strict_types=1)` preserved |
| **File-level code** | ⚠️ Partial | Focus on OOP constructs |

### Object-Oriented Constructs

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including abstract, final, readonly (8.2+) |
| **Interfaces** | ✅ Full | Multiple inheritance supported |
| **Traits** | ⚠️ Partial | Represented as special classes with markers |
| **Enums** | ✅ Full | Pure and backed enums (8.1+) |
| **Properties** | ✅ Full | Typed, readonly, promoted |
| **Methods** | ✅ Full | Including magic methods |
| **Constants** | ✅ Full | Class constants with visibility |
| **Attributes** | ✅ Full | PHP 8.0+ attributes/annotations |

### Type System Features

| Feature | Support Level | Notes |
|---------|--------------|-------|
| **Type declarations** | ✅ Full | Parameters and return types |
| **Union types** | ✅ Full | PHP 8.0+ `string\|int` |
| **Intersection types** | ✅ Full | PHP 8.1+ `Countable&Iterator` |
| **Nullable types** | ✅ Full | `?string` syntax |
| **Mixed type** | ✅ Full | Explicit `mixed` type |
| **Array types** | ✅ Full | Via PHPDoc for specific types |
| **Constructor promotion** | ✅ Full | PHP 8.0+ feature |

### Visibility Rules

PHP visibility in AI Distiller follows standard PHP conventions:
- **Public**: `public` keyword or no modifier (default for methods)
- **Protected**: `protected` keyword
- **Private**: `private` keyword
- **Magic methods**: Always considered public (`__construct`, `__toString`, etc.)

## Key Features

### 1. **Constructor Property Promotion**

AI Distiller correctly extracts properties from PHP 8.0+ constructor promotion:

```php
// Input
class User {
    public function __construct(
        private readonly int $id,
        private string $name,
        protected ?string $email = null
    ) {}
}
```

```
// Output (default stripping)
class User:
    +__construct(id: int, name: string, email: ?string = null)
```

```
// Output (with private members)
class User:
    -readonly id: int
    -name: string
    #email: ?string
    +__construct(id: int, name: string, email: ?string = null)
```

### 2. **Trait Representation**

Traits are represented with special markers and usage comments:

```php
// Input
trait Timestampable {
    private ?DateTime $createdAt = null;
    
    public function touch(): void {
        $this->createdAt = new DateTime();
    }
}

class Post {
    use Timestampable, Sluggable;
}
```

```
// Output
# PHP Trait
trait Timestampable:
    -createdAt: ?DateTime
    +touch() -> void

class Post:
    # Uses traits: Timestampable, Sluggable
```

### 3. **PHPDoc Type Enhancement**

PHPDoc annotations enhance type information, especially for arrays:

```php
// Input
/**
 * @param Product[] $products
 * @return array<string, Product>
 */
public function indexByName(array $products): array
```

```
// Output
+indexByName(products: Product[]) -> array<string, Product>
```

## Output Formats

### Text Format (Recommended for AI)

The text format uses a Python-like syntax optimized for AI comprehension:
- Clear visibility markers (`+` public, `#` protected, `-` private)
- Type information preserved
- Minimal syntax overhead
- Interface implementation shown with `implements`
- Class extension shown with parentheses

## Real-World Examples

<details open>
<summary>Example: Basic Function and Class</summary>
<blockquote>

<details>
<summary>Input: `construct1_basic.php`</summary>
<blockquote>

```php
<?php

declare(strict_types=1);

/**
 * Calculates the final price after applying a discount.
 *
 * @param float $price The original price.
 * @param int $discountPercent The discount percentage.
 * @return float The price after discount.
 */
function calculate_final_price(float $price, int $discountPercent): float
{
    if ($price <= 0) {
        return 0.0;
    }

    $discountAmount = $price * ($discountPercent / 100);

    return $price - $discountAmount;
}

// A simple, empty class definition to test basic OOP parsing.
class Product
{
}

$bookPrice = 20.0;
$finalPrice = calculate_final_price($bookPrice, 15);

echo "Final price: " . $finalPrice;
```

</blockquote>
</details>

<details open>
<summary>Default Output (`--strip 'non-public,comments,implementation'`)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct1_basic.php">
+calculate_final_price(price: float, discountPercent: int) -> float

class Product:
</file>
```

</blockquote>
</details>

<details>
<summary>Full Output (no stripping)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct1_basic.php">
# Calculates the final price after applying a discount.
#
# @param float $price The original price.
# @param int $discountPercent The discount percentage.
# @return float The price after discount.
+calculate_final_price(price: float, discountPercent: int) -> float:
    {
        if ($price <= 0) {
            return 0.0;
        }

        $discountAmount = $price * ($discountPercent / 100);

        return $price - $discountAmount;
    }
# A simple, empty class definition to test basic OOP parsing.

class Product:
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Constructor Property Promotion</summary>
<blockquote>

<details>
<summary>Input: `construct2_property_promotion.php`</summary>
<blockquote>

```php
<?php

declare(strict_types=1);

// Using PHP 8 Constructor Property Promotion
class User
{
    public function __construct(
        private readonly int $id,
        private string $name,
        private string $email
    ) {}

    public function getId(): int
    {
        return $this->id;
    }

    public function getDisplayName(): string
    {
        return "User: " . $this->name;
    }

    public function changeEmail(string $newEmail): void
    {
        // Basic validation logic
        if (!filter_var($newEmail, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid email format provided.");
        }
        $this->email = $newEmail;
    }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`--strip 'non-public,comments,implementation'`)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct2_property_promotion.php">
class User:
    +__construct(id: int, name: string, email: string)
    +getId() -> int
    +getDisplayName() -> string
    +changeEmail(newEmail: string) -> void
</file>
```

</blockquote>
</details>

<details>
<summary>Output with `--strip 'implementation'`</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct2_property_promotion.php">
# Using PHP 8 Constructor Property Promotion

class User:
    -readonly id: int
    -name: string
    -email: string
    +__construct(id: int, name: string, email: string)
    +getId() -> int
    +getDisplayName() -> string
    +changeEmail(newEmail: string) -> void
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Interfaces and Abstract Classes</summary>
<blockquote>

<details>
<summary>Input: `construct3_interfaces_abstract.php`</summary>
<blockquote>

```php
<?php

declare(strict_types=1);

interface Loggable
{
    public function log(string $message): void;
}

abstract class AbstractStorage
{
    protected string $storagePath;

    public function __construct(string $storagePath)
    {
        $this->storagePath = rtrim($storagePath, '/');
    }

    abstract protected function save(string $key, string $data): bool;
    
    final public function getStoragePath(): string
    {
        return $this->storagePath;
    }
}

class FileLogger extends AbstractStorage implements Loggable
{
    public function __construct(string $logDirectory)
    {
        parent::__construct($logDirectory);
    }

    public function log(string $message): void
    {
        $this->save('log_' . date('Y-m-d'), $message . PHP_EOL);
    }

    protected function save(string $key, string $data): bool
    {
        $file = $this->storagePath . '/' . $key . '.log';
        return file_put_contents($file, $data, FILE_APPEND) !== false;
    }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`--strip 'non-public,comments,implementation'`)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct3_interfaces_abstract.php">
interface Loggable:
    +log(message: string) -> void

abstract class AbstractStorage:
    +__construct(storagePath: string)
    +final getStoragePath() -> string

class FileLogger(AbstractStorage) implements Loggable:
    +__construct(logDirectory: string)
    +log(message: string) -> void
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Traits and Union Types</summary>
<blockquote>

<details>
<summary>Input: `construct4_traits_union_types.php`</summary>
<blockquote>

```php
<?php

namespace App\Services\Notification;

use App\Utils\Timestampable;
use Psr\Log\LoggerInterface;

class EmailPayload
{
    use Timestampable;

    public function __construct(
        public readonly string $recipient,
        public readonly string $subject,
        public readonly string $body
    ) {
        $this->touch();
    }
}

class Notifier
{
    public function __construct(private LoggerInterface $logger) {}

    public function send(EmailPayload|string $payload): void
    {
        if (is_string($payload)) {
            $this->logger->info("Raw string notification: {$payload}");
            return;
        }

        $this->logger->info(
            "Sending email to {$payload->recipient} with subject '{$payload->subject}'"
        );
    }
}

namespace App\Utils;

trait Timestampable
{
    private ?\DateTimeImmutable $createdAt = null;

    public function touch(): void
    {
        if ($this->createdAt === null) {
            $this->createdAt = new \DateTimeImmutable();
        }
    }

    public function getCreationDate(): ?\DateTimeImmutable
    {
        return $this->createdAt;
    }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`--strip 'non-public,comments,implementation'`)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct4_traits_union_types.php">
import App\Utils\Timestampable
import Psr\Log\LoggerInterface

class EmailPayload:
    # Uses traits: Timestampable
    +readonly recipient: string
    +readonly subject: string
    +readonly body: string
    +__construct(recipient: string, subject: string, body: string)

class Notifier:
    +__construct(logger: LoggerInterface)
    +send(payload: EmailPayload|string) -> void
# PHP Trait

trait Timestampable:
    +touch() -> void
    +getCreationDate() -> ?\DateTimeImmutable
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Attributes and Complex Interfaces</summary>
<blockquote>

<details>
<summary>Input: `construct5_very_complex.php`</summary>
<blockquote>

```php
<?php

namespace App\Data\Repositories;

use App\Data\Models\Product;
use App\Data\Contracts\{Cacheable, Deletable};
use App\Data\Traits\HasSoftDeletes;
use \Serializable;

#[\Attribute(\Attribute::TARGET_CLASS)]
class RepositoryConfig
{
    public function __construct(public string $model) {}
}

interface FindableById
{
    public function find(int $id);
}

#[RepositoryConfig(model: Product::class)]
class ProductRepository extends BaseRepository implements FindableById, Cacheable, Deletable, Serializable
{
    use HasSoftDeletes;

    private static int $queryCount = 0;
    protected array $searchableFields = ['name', 'sku'];

    public function __construct()
    {
        parent::__construct(Product::class);
    }

    public function find(int $id): ?Product
    {
        self::$queryCount++;
        if ($id === 1) {
            return new Product(1, 'Laptop', 1500.00);
        }
        return null;
    }

    /**
     * @return Product[]
     */
    public function findBy(string $field, mixed $value): array
    {
        self::$queryCount++;
        return [new Product(1, 'Laptop', 1500.00)];
    }

    public function clearCache(): bool { return true; }
    public function serialize(): string { return ''; }
    public function unserialize(string $data): void { }
}
```

</blockquote>
</details>

<details open>
<summary>Default Output (`--strip 'non-public,comments,implementation'`)</summary>
<blockquote>

```
<file path="/home/janreges/ai-distiller/test-data/php/construct5_very_complex.php">
import App\Data\Models\Product
import App\Data\Contracts\Cacheable
import App\Data\Contracts\Deletable
import App\Data\Traits\HasSoftDeletes
import \Serializable

@\Attribute(\Attribute::TARGET_CLASS)
class RepositoryConfig:
    +model: string
    +__construct(model: string)

interface FindableById:
    +find(id: int) -> object|null

@RepositoryConfig(model: Product::class)
class ProductRepository(BaseRepository) implements FindableById, Cacheable, Deletable, Serializable:
    # Uses traits: HasSoftDeletes
    +__construct()
    +find(id: int) -> ?Product
    +findBy(field: string, value: mixed) -> Product[]
    +clearCache() -> bool
    +serialize() -> string
    +unserialize(data: string) -> void
</file>
```

</blockquote>
</details>

</blockquote>
</details>

## Representation Model and Limitations

### Trait Representation

**Model**: Traits are represented as a distinct entity type marked with a special comment. Classes using traits have a comment indicating trait usage.

**Rationale**: This model captures the composition relationship while maintaining clarity about what is a trait versus a class.

**Limitations**:
- Trait conflict resolution (`insteadof`, `as`) is not represented
- Method visibility changes when using traits are not shown
- Trait precedence rules are implicit

### Dynamic Constructs Not Supported

AI Distiller performs static analysis and cannot resolve:
- `eval()` and `create_function()`
- Variable variables (`$$foo`), variable functions (`$func()`)
- Variable class instantiation (`new $class()`)
- Dynamic `include`/`require` paths
- Runtime-resolved method calls

### Framework Magic

AI Distiller does not execute framework bootstrapping:
- **Laravel**: Facades and Service Container resolutions are not traced
- **Symfony**: Dependency Injection container wirings are not resolved
- **Doctrine**: Entity relationships defined via attributes/annotations are preserved but not interpreted

## Best Practices

### 1. **Use Strict Types and Type Declarations**

```php
<?php
declare(strict_types=1);

// Good - Full type information preserved
public function process(array $data, ProcessOptions $options): Result {
    // ...
}

// Less optimal - Generic array type
public function process($data, $options) {
    // ...
}
```

### 2. **Leverage PHPDoc for Array Types**

```php
/**
 * @param Product[] $products Array of products
 * @param array<string, int> $quantities Map of SKU to quantity
 * @return array{total: float, items: Product[]}
 */
public function calculateOrder(array $products, array $quantities): array
```

### 3. **Use PHP 8+ Features**

Constructor property promotion and union types provide cleaner, more maintainable code:

```php
// PHP 8+ - Cleaner and fully supported
public function __construct(
    private readonly int $id,
    private string $name,
    private Status $status = Status::ACTIVE
) {}

// Older style - More verbose
private int $id;
private string $name;
private Status $status;

public function __construct(int $id, string $name, Status $status = Status::ACTIVE) {
    $this->id = $id;
    $this->name = $name;
    $this->status = $status;
}
```

### 4. **Structure for AI Consumption**

- Keep classes focused (under 500 lines)
- Use meaningful names (they're preserved!)
- Group related functionality in namespaces
- Prefer composition over deep inheritance

## Integration Examples

### CLI Usage

```bash
# Generate optimal context for AI assistants
aid ./src --format text \
    --strip "non-public,implementation,comments" \
    --output context.txt

# Extract public API only
aid ./src --strip "non-public" --format text > public-api.txt

# Full structure with implementations
aid ./src --format text > full-structure.txt
```

### MCP Integration with Claude

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

> **You**: Show me the structure of our authentication system and identify where we could add two-factor authentication.

> **Claude**: I'll analyze your authentication system structure.

*Claude automatically calls:*
```
distillDirectory("src/Auth", strip_implementation=true)
search("class.*Auth.*", is_regex=true)
```

> Based on the distilled structure, I can see your authentication system has:
>
> ```
> AuthenticationService:
>   + __construct(UserRepository $users, TokenGenerator $tokens)
>   + authenticate(string $username, string $password): ?User
>   + createSession(User $user): Session
>   + validateSession(string $token): ?Session
>   
> User:
>   + verifyPassword(string $password): bool
>   + updateLastLogin(): void
> ```
>
> For two-factor authentication, I recommend:
> 1. Add a `TwoFactorProvider` interface
> 2. Extend `AuthenticationService::authenticate()` to return an intermediate state
> 3. Add fields to the `User` entity for 2FA configuration

### Composer Integration

```json
{
  "scripts": {
    "analyze:structure": "aid src/ --format json > structure.json",
    "analyze:api": "aid src/ --strip non-public,implementation --format text > api.txt",
    "pre-commit": [
      "@analyze:api",
      "git diff --quiet api.txt || echo 'API surface changed!'"
    ]
  }
}
```

### PHPStan/Psalm Complementary Usage

AI Distiller complements static analyzers:

```bash
# First, ensure code quality
vendor/bin/phpstan analyze
vendor/bin/psalm

# Then extract structure for AI
aid src/ --format text --strip non-public,implementation > structure.txt

# Use both for comprehensive analysis
echo "Code quality checked. Structure extracted for AI analysis."
```

## Framework-Specific Considerations

### Laravel

```php
// Routes, middleware, and service providers need manual documentation
// AI Distiller sees the structure but not the wiring

// Add explicit type hints for better extraction
public function handle(Request $request, Closure $next): Response
{
    // ...
}
```

### Symfony

```php
// Use constructor injection over property injection
public function __construct(
    private LoggerInterface $logger,
    private EntityManagerInterface $em
) {}

// Attributes are preserved
#[Route('/api/products', methods: ['GET'])]
public function list(): JsonResponse
```

### WordPress

```php
// Use modern PHP features where possible
class CustomPostType 
{
    public function __construct(
        private string $postType,
        private array $args = []
    ) {
        add_action('init', [$this, 'register']);
    }
}
```

## Troubleshooting

### "Parser failed" errors

Ensure your PHP syntax is valid:
```bash
php -l yourfile.php
```

### Missing type information

1. Add explicit type declarations
2. Use PHPDoc for array types
3. Upgrade to PHP 7.4+ for property types

### Trait methods not showing

Traits are shown separately. Look for:
- `# PHP Trait` marker
- `# Uses traits: TraitName` in classes

### Large files timing out

Split large files or increase timeout:
```bash
aid large-file.php --timeout 60000
```

## Performance Considerations

- Files are processed in parallel
- Large codebases (10k+ files) process in seconds
- Memory usage is proportional to file count, not size
- Tree-sitter parsing is highly optimized

## Security and Safety

AI Distiller never executes code:
- Safe to run on any codebase
- No risk of side effects
- Credentials in code are preserved (sanitize before sharing!)

## Comparison with Other Tools

| Tool | Purpose | PHP Support | AI-Optimized |
|------|---------|-------------|--------------|
| **AI Distiller** | Structure extraction | Full 7.4-8.3 | ✅ Yes |
| PHPStan | Static analysis | Full | ❌ No |
| Psalm | Static analysis | Full | ❌ No |
| PHP-Parser | AST generation | Full | ❌ No |
| phpDocumentor | Documentation | Full | ❌ No |

## Contributing

Help improve PHP support! Key areas:

1. **Trait conflict resolution** - Implement `insteadof`/`as` support
2. **Anonymous classes** - Add full support
3. **Enum methods** - Enhance enum parsing
4. **Performance optimizations** - Large file handling
5. **Framework patterns** - Recognize common patterns

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Future Enhancements

- [ ] Full trait conflict resolution support
- [ ] Anonymous class extraction
- [ ] Closure and arrow function analysis
- [ ] `@template` PHPDoc support for generics
- [ ] Framework-specific adapters
- [ ] PHP 8.3 typed class constants
- [ ] `.phpstorm.meta.php` integration

---

<sub>Documentation generated for AI Distiller v0.2.0 - PHP Support</sub>