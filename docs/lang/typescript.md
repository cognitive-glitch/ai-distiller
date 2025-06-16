# TypeScript Language Support

AI Distiller provides comprehensive support for TypeScript codebases using the [tree-sitter-typescript](https://github.com/tree-sitter/tree-sitter-typescript) parser, with full support for modern TypeScript features including generics, decorators, and advanced type system constructs.

## Overview

TypeScript support in AI Distiller is designed to capture the rich type information and structural patterns that make TypeScript powerful. The distilled output preserves type safety contracts, generic constraints, and architectural relationships while minimizing token count for efficient AI processing.

## Supported TypeScript Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including abstract classes, inheritance, generics |
| **Interfaces** | ✅ Full | With extends, generics, index signatures |
| **Functions** | ✅ Full | Regular, async, arrow functions, overloads |
| **Methods** | ✅ Full | Public, private, protected, abstract, static |
| **Type System** | ✅ Full | Union, intersection, conditional, mapped types |
| **Generics** | ✅ Full | With constraints, defaults, inference |
| **Type Aliases** | ✅ Full | Including complex utility types |
| **Imports/Exports** | ✅ Full | ES6 modules, type-only imports |
| **Enums** | ✅ Full | Const and regular enums |
| **Decorators** | ⚠️ Partially supported | Applied to classes/methods but values not captured |
| **Namespaces** | ❌ Not supported | Modern modules preferred |
| **Parameter Properties** | ✅ Full | Constructor shorthand syntax |

### Visibility Rules

TypeScript visibility in AI Distiller follows the language's access modifiers:
- **Public**: Default visibility or explicit `public` keyword, exported declarations
- **Private**: `private` keyword, `#` private fields, or underscore-prefixed non-exported members
- **Protected**: `protected` keyword
- **Internal**: Non-exported top-level declarations (module-private)

## Key Features

### 1. **Advanced Generic Type Preservation**

AI Distiller captures complex generic constraints and relationships:

```typescript
// Input
class EventEmitter<TEventMap extends object> {
  emit<K extends keyof TEventMap>(
    event: K, 
    payload: TEventMap[K]
  ): void { /* ... */ }
}

type Predicate<T> = T extends (...args: any[]) => infer R ? R : never;
```

```
# Output
class EventEmitter<TEventMap extends object>:
    function emit<K extends keyof TEventMap>(event: K, payload: TEventMap[K]) -> void

type Predicate<T> = T extends (...args: any[]) => infer R ? R : never
```

### 2. **Interface and Type Relationships**

Inheritance and implementation contracts are preserved:

```typescript
// Input
interface INotifiable<P> {
  handleNotification(payload: P): void;
}

abstract class BaseService {
  abstract process<T extends {id: string}>(item: T): Promise<boolean>;
}

class EmailService extends BaseService implements INotifiable<EmailPayload> {
  // Implementation
}
```

```
# Output
interface INotifiable<P>:
    method handleNotification(payload: P): void

abstract class BaseService:
    abstract function process<T extends {id: string}>(item: T) -> Promise<boolean>

class EmailService extends BaseService implements INotifiable<EmailPayload>:
    # Members...
```

### 3. **Parameter Properties**

TypeScript's constructor shorthand is intelligently processed:

```typescript
// Input
class UserModel {
  constructor(
    public readonly id: number,
    private email: string,
    protected lastLogin?: Date
  ) {}
}
```

```
# Output
class UserModel:
    field public readonly id: number
    field private email: string
    field protected lastLogin: Date
    function constructor(id: number, email: string, lastLogin: Date)
```

## Output Formats

### Text Format (Recommended for AI)

AI Distiller uses a TypeScript-specific formatter that preserves language-specific syntax while maintaining readability. The text format is optimized for AI comprehension:

<details open>
<summary>Example: Complex TypeScript Class</summary>
<blockquote>

<details>
<summary>Input: `plugin-manager.ts`</summary>
<blockquote>

```typescript
export interface IPlugin {
  name: string;
  execute(data: any): void;
}

export class PluginManager {
  private plugins: Map<string, IPlugin> = new Map();

  constructor(private settings: PluginSettings = {}) {}

  @LogExecution
  public registerPlugin(plugin: IPlugin): void {
    if (this.plugins.has(plugin.name)) {
      throw new Error(`Plugin "${plugin.name}" already registered`);
    }
    this.plugins.set(plugin.name, plugin);
  }

  public async runAll(data: any): Promise<void> {
    for (const plugin of this.plugins.values()) {
      await plugin.execute(data);
    }
  }
}
```

</blockquote>
</details>

<details>
<summary>Output: Default (full)</summary>
<blockquote>

```
interface IPlugin:
    property name: string
    method execute(data: any): void

class PluginManager:
    field private plugins: Map<string, IPlugin> = new Map()
    field private settings: PluginSettings
    function constructor(settings: PluginSettings)
        // implementation
    function registerPlugin(plugin: IPlugin) -> void
        // implementation
    async function runAll(data: any) -> Promise<void>
        // implementation
```

</blockquote>
</details>

<details>
<summary>Output: `--private=0 --protected=0 --internal=0`</summary>
<blockquote>

```
interface IPlugin:
    property name: string
    method execute(data: any): void

class PluginManager:
    function constructor(settings: PluginSettings)
        // implementation
    function registerPlugin(plugin: IPlugin) -> void
        // implementation
    async function runAll(data: any) -> Promise<void>
        // implementation
```

</blockquote>
</details>

<details>
<summary>Output: `--implementation=0`</summary>
<blockquote>

```
interface IPlugin:
    property name: string
    method execute(data: any): void

class PluginManager:
    field private plugins: Map<string, IPlugin> = new Map()
    field private settings: PluginSettings
    function constructor(settings: PluginSettings)
    function registerPlugin(plugin: IPlugin) -> void
    async function runAll(data: any) -> Promise<void>
```

</blockquote>
</details>

</blockquote>
</details>

### Markdown Format

The markdown format provides a hierarchical view with emojis for visual parsing:

```markdown
# plugin-manager.ts

## Structure

📦 **Interface** `IPlugin`
  📝 **Property** `name`: `string`
  🔧 **Method** `execute`(`data`: `any`) → `void`

🏛️ **Class** `PluginManager`
  🔒 **Field** `plugins`: `Map<string, IPlugin>` = `new Map()`
  🔒 **Field** `settings`: `PluginSettings`
  🔧 **Constructor** (`settings`: `PluginSettings`)
  🔧 **Method** `registerPlugin`(`plugin`: `IPlugin`) → `void`
  ⚡ **Async Method** `runAll`(`data`: `any`) → `Promise<void>`
```

## Advanced TypeScript Features

### Conditional Types and Type Inference

```typescript
type ChangeEvent<T extends string> = `${T}Changed`;
type Payload<T> = T extends (payload: infer P) => void ? P : never;

type ListenerMap<TEventMap extends object> = {
  [K in keyof TEventMap as ChangeEvent<K & string>]: (payload: TEventMap[K]) => void;
};
```

AI Distiller preserves these complex type manipulations verbatim, maintaining the full expressiveness of TypeScript's type system.

### Mapped Types and Index Signatures

```typescript
type PluginSettings = {
  [K in 'timeout' | 'retries']?: K extends 'timeout' ? number : number;
};

interface StringIndex {
  [key: string]: any;
  length: number;  // Named property
}
```

## Integration with AI Workflows

### Optimizing for Context Windows

TypeScript code often includes extensive type definitions. Use stripping options strategically:

```bash
# For implementation analysis - keep types, remove docs
aid src/ --comments=0 --format text

# For API understanding - keep signatures, remove implementation
aid src/ --implementation=0 --format text

# For architecture overview - public API only
aid src/ --private=0 --protected=0 --internal=0,implementation --format text
```

### Framework-Specific Patterns

AI Distiller recognizes common TypeScript framework patterns:

- **Angular**: Components, services, decorators (decorator values not captured)
- **React**: Function components, hooks, props interfaces
- **Node.js**: Express middleware, async handlers
- **NestJS**: Controllers, providers, modules

## Known Limitations

1. **Decorators**: Decorator applications (e.g., `@Component()`) are detected but decorator arguments/values are not captured
2. **Constructor Parameter Defaults**: Default values in constructor parameters are not shown
3. **Function Overloads**: Multiple signatures are merged into a single signature
4. **Namespaces**: Not supported (use ES6 modules instead)
5. **Triple-slash Directives**: Reference comments are ignored
6. **Const Arrow Functions**: Displayed as `const Name` instead of `const Name = () => {}`

## Best Practices

### 1. Type-First Development

AI Distiller works best with well-typed TypeScript:

```typescript
// ✅ Good - Explicit types preserved
function process(data: UserData[]): ProcessedResult {
  return data.map(transform);
}

// ❌ Avoid - Type information lost
function process(data) {
  return data.map(transform);
}
```

### 2. Interface Segregation

Smaller, focused interfaces distill better:

```typescript
// ✅ Good - Clear, focused interfaces
interface Readable {
  read(): Buffer;
}

interface Writable {
  write(data: Buffer): void;
}

// ❌ Avoid - Large, monolithic interfaces
interface Stream {
  read(): Buffer;
  write(data: Buffer): void;
  // ... 20 more methods
}
```

### 3. Consistent Visibility

Use TypeScript's visibility modifiers consistently for better distillation:

```typescript
class Service {
  public api(): void {}      // Explicitly public
  protected hook(): void {}  // For extension points
  private helper(): void {}  // Internal only
}
```

## Known Issues & Recent Fixes

### Recently Fixed (December 2024)

1. **Non-exported classes visibility** (✅ Fixed)
   - **Issue**: Non-exported abstract classes were filtered out in default view
   - **Fix**: Improved visibility handling for module-internal declarations
   - **Impact**: Abstract base classes now properly appear in output

2. **Underscore-prefixed members** (✅ Fixed)
   - **Issue**: Functions like `_normalizeString` appeared in default output
   - **Fix**: Added convention-based private detection for underscore prefix
   - **Impact**: Better alignment with JavaScript/TypeScript conventions

3. **Top-level function visibility** (✅ Fixed)
   - **Issue**: Top-level functions showed "public" prefix incorrectly
   - **Fix**: Updated formatter to omit visibility for top-level declarations
   - **Impact**: Cleaner, more idiomatic output

### Current Limitations

1. **Decorator values not captured** (🟡 Minor)
   - Decorators are shown but their arguments/values aren't preserved
   - Workaround: Decorator presence is sufficient for most AI use cases

2. **Namespace support missing** (🟢 By design)
   - Modern ES modules are preferred over legacy namespaces
   - Workaround: Use ES6 modules instead

## Performance Characteristics

TypeScript parsing performance with tree-sitter:

| Codebase Size | Parse Time | Memory Usage |
|--------------|------------|--------------|
| Small (< 1MB) | < 100ms | ~10MB |
| Medium (1-10MB) | < 500ms | ~50MB |
| Large (10-50MB) | < 2s | ~200MB |
| Huge (> 50MB) | < 5s | ~500MB |

## Contributing

TypeScript support is actively maintained. Key areas for contribution:

1. **Decorator Support**: Implement decorator parsing in `ast_parser.go`
2. **Parameter Defaults**: Capture default values in function parameters
3. **JSDoc Integration**: Parse and include JSDoc type annotations
4. **Performance**: Optimize tree-sitter grammar loading

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development setup.