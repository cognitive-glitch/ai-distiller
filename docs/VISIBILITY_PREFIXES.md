# Visibility Prefixes in AI Distiller

AI Distiller uses UML notation to indicate the visibility of class members, functions, and fields in the output. This provides a standard, compact way to understand access modifiers at a glance.

## UML Notation Convention

- **`+`** = Public
- **`-`** = Private
- **`#`** = Protected

## Examples

### PHP Example

```php
class Example {
    public string $publicField;
    protected string $protectedField;
    private string $privateField;
    
    public function publicMethod() {}
    protected function protectedMethod() {}
    private function privateMethod() {}
}
```

Output:
```
class Example:
    +publicField: string
    #protectedField: string
    -privateField: string
    +publicMethod()
    #protectedMethod()
    -privateMethod()
```

### Python Example

In Python, visibility is determined by naming convention:
- Names starting with `_` are considered private
- Names starting with `__` are also private (name mangling)

```python
class Example:
    def __init__(self):
        self.public_attr = "public"
        self._private_attr = "private"
    
    def public_method(self):
        pass
    
    def _private_method(self):
        pass
```

Output:
```
class Example:
    +__init__(self)
    +public_method(self)
    -_private_method(self)
```

## Stripping Options

You can control which visibility levels are included in the output:

- `--strip non-public` - Remove both private and protected members
- `--strip private` - Remove only private members (keep protected)
- `--strip protected` - Remove only protected members (keep private)

Examples:
```bash
# Show only public members
aid myproject --strip non-public

# Show public and protected members
aid myproject --strip private

# Show public and private members
aid myproject --strip protected
```

## Language Support

The visibility prefix system works across all supported languages:

- **PHP**: Uses explicit `public`, `protected`, `private` keywords
- **Python**: Uses naming conventions (`_` prefix for private)
- **Java/C#**: Uses explicit access modifiers
- **JavaScript/TypeScript**: Uses `#` for private fields (ES2022+)
- **Go**: Uses capitalization (lowercase = private to package)

The prefixes provide a unified way to understand visibility across different language conventions.