# Python Language Support

AI Distiller provides comprehensive support for Python 3.x codebases using the [tree-sitter-python](https://github.com/tree-sitter/tree-sitter-python) parser.

## Overview

Python support in AI Distiller is designed to extract the essential structure of Python code while preserving type information and API contracts. The distilled output maintains Python's semantic meaning while dramatically reducing token count for AI consumption.

## Supported Python Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including inheritance, metaclasses, decorators |
| **Functions** | ✅ Full | Regular, async, generators, decorators |
| **Methods** | ✅ Full | Instance, class, static, properties |
| **Type Hints** | ✅ Full | PEP 484, 526, 544, 585, 604 |
| **Decorators** | ✅ Full | Function and class decorators |
| **Imports** | ✅ Full | Regular, from, aliases |
| **Docstrings** | ✅ Full | Preserved or stripped based on options |
| **Async/Await** | ⚠️ Partial | `async` keyword currently missing in output |
| **Dataclasses** | ✅ Full | Recognized via decorator |
| **Enums** | ✅ Full | Treated as classes |
| **Metaclasses** | ⚠️ Partial | Detected but not shown in output |
| **Nested Classes** | ❌ Not supported | Line parser limitation |

### Visibility Rules

Python visibility in AI Distiller follows these conventions:
- **Public**: No underscore prefix, or dunder methods (`__init__`, `__str__`, etc.)
- **Protected**: Single underscore prefix (`_method`)
- **Private**: Double underscore prefix (`__method`), except dunders

## Key Features

### 1. **Type Information Extraction**

AI Distiller maximizes type information extraction, even from loosely typed Python code:

```python
# Input
def process_data(items: List[Dict[str, Any]], 
                 callback: Callable[[str], None] = None) -> Optional[DataFrame]:
    """Process items and optionally notify via callback."""
    pass
```

```
# Output (with --implementation=0)
+def process_data(items: List[Dict[str, Any]], callback: Callable[[str], None] = None) -> Optional[DataFrame]
```

### 2. **Smart Docstring Handling**

Docstrings are preserved as documentation, not implementation:

```python
# Even with --implementation=0, docstrings remain
def calculate(x: float) -> float:
    """Calculate the square root of x."""
    return math.sqrt(x)
```

### 3. **Import Graph Analysis**

AI Distiller tracks import relationships for dependency understanding:

```python
from typing import List, Optional
from pandas import DataFrame
import numpy as np
from .utils import helper
```

## Output Formats

### Text Format (Recommended for AI)

The text format uses UML-inspired visibility prefixes:
- `+` public members
- `#` protected members  
- `-` private members
- `~` internal/package-private

<details open>
<summary>Example: Basic Class</summary>
<blockquote>

<details>
<summary>Input: `user.py`</summary>
<blockquote>

```python
class User:
    """Represents a user in the system."""
    
    def __init__(self, name: str, email: str):
        self.name = name
        self.email = email
        self._id = self._generate_id()
    
    def get_display_name(self) -> str:
        """Returns the user's display name."""
        return self.name.title()
    
    def _generate_id(self) -> str:
        """Internal method to generate user ID."""
        return f"usr_{hash(self.email)}"
```

</blockquote>
</details>

<details open>
<summary>Default Output (`default output (public only, no implementation)`)</summary>
<blockquote>

```
<file path="user.py">
class User:
    +def __init__(self, name: str, email: str)
    +def get_display_name(self) -> str
</file>
```

</blockquote>
</details>

<details>
<summary>Full Output (no stripping)</summary>
<blockquote>

```
<file path="user.py">
class User:
    """Represents a user in the system."""
    +def __init__(self, name: str, email: str):
        self.name = name
        self.email = email
        self._id = self._generate_id()
    +def get_display_name(self) -> str:
        """Returns the user's display name."""
        return self.name.title()
    #def _generate_id(self) -> str:
        """Internal method to generate user ID."""
        return f"usr_{hash(self.email)}"
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Complex Async Service</summary>
<blockquote>

<details>
<summary>Input: `service.py`</summary>
<blockquote>

```python
from typing import AsyncIterator, Optional
from abc import ABC, abstractmethod
import asyncio

class AsyncService(ABC):
    """Abstract base class for async services."""
    
    def __init__(self, name: str, timeout: float = 30.0):
        self.name = name
        self.timeout = timeout
        self._running = False
    
    @abstractmethod
    async def process(self, data: bytes) -> bytes:
        """Process data asynchronously."""
        pass
    
    async def start(self) -> None:
        """Start the service."""
        self._running = True
        await self._initialize()
    
    async def stop(self) -> None:
        """Stop the service."""
        self._running = False
        await self._cleanup()
    
    async def stream_data(self) -> AsyncIterator[bytes]:
        """Stream processed data."""
        while self._running:
            chunk = await self._get_next_chunk()
            if chunk:
                yield await self.process(chunk)
    
    async def _initialize(self) -> None:
        """Initialize service resources."""
        await asyncio.sleep(0.1)
    
    async def _cleanup(self) -> None:
        """Clean up service resources."""
        await asyncio.sleep(0.1)
    
    async def _get_next_chunk(self) -> Optional[bytes]:
        """Get next data chunk."""
        return b"data"
```

</blockquote>
</details>

<details open>
<summary>Default Output (`default output (public only, no implementation)`)</summary>
<blockquote>

```
<file path="service.py">
from typing import AsyncIterator, Optional
from abc import ABC, abstractmethod
import asyncio

class AsyncService(ABC):
    +def __init__(self, name: str, timeout: float = 30.0)
    @abstractmethod
    +async def process(self, data: bytes) -> bytes
    +async def start(self) -> None
    +async def stop(self) -> None
    +async def stream_data(self) -> AsyncIterator[bytes]
</file>
```

</blockquote>
</details>

</blockquote>
</details>

<details>
<summary>Example: Advanced Type Annotations</summary>
<blockquote>

<details>
<summary>Input: `advanced_types.py`</summary>
<blockquote>

```python
from typing import TypeVar, Generic, Protocol, Union, Literal, TypedDict
from typing import overload

T = TypeVar('T')
K = TypeVar('K')
V = TypeVar('V')

class Comparable(Protocol):
    """Protocol for comparable objects."""
    def __lt__(self, other: 'Comparable') -> bool: ...
    def __le__(self, other: 'Comparable') -> bool: ...

class ConfigDict(TypedDict):
    """Configuration dictionary structure."""
    host: str
    port: int
    debug: bool
    features: list[str]

class Cache(Generic[K, V]):
    """Generic cache implementation."""
    
    def __init__(self, max_size: int = 100):
        self._cache: dict[K, V] = {}
        self.max_size = max_size
    
    @overload
    def get(self, key: K, default: None = None) -> Union[V, None]: ...
    
    @overload
    def get(self, key: K, default: V) -> V: ...
    
    def get(self, key: K, default: Union[V, None] = None) -> Union[V, None]:
        """Get value from cache."""
        return self._cache.get(key, default)
    
    def put(self, key: K, value: V) -> None:
        """Put value in cache."""
        if len(self._cache) >= self.max_size:
            self._evict_oldest()
        self._cache[key] = value
    
    def _evict_oldest(self) -> None:
        """Evict oldest entry from cache."""
        if self._cache:
            oldest = next(iter(self._cache))
            del self._cache[oldest]

def process_config(config: ConfigDict, 
                  mode: Literal['dev', 'prod'] = 'prod') -> bool:
    """Process configuration based on mode."""
    return config['debug'] if mode == 'dev' else False
```

</blockquote>
</details>

<details open>
<summary>Default Output (`default output (public only, no implementation)`)</summary>
<blockquote>

```
<file path="advanced_types.py">
from typing import TypeVar, Generic, Protocol, Union, Literal, TypedDict
from typing import overload

T = TypeVar('T')
K = TypeVar('K')
V = TypeVar('V')

class Comparable(Protocol):
    +def __lt__(self, other: 'Comparable') -> bool
    +def __le__(self, other: 'Comparable') -> bool

class ConfigDict(TypedDict):
    host: str
    port: int
    debug: bool
    features: list[str]

class Cache(Generic[K, V]):
    +def __init__(self, max_size: int = 100)
    @overload
    +def get(self, key: K, default: None = None) -> Union[V, None]
    @overload
    +def get(self, key: K, default: V) -> V
    +def get(self, key: K, default: Union[V, None] = None) -> Union[V, None]
    +def put(self, key: K, value: V) -> None

+def process_config(config: ConfigDict, mode: Literal['dev', 'prod'] = 'prod') -> bool
</file>
```

</blockquote>
</details>

</blockquote>
</details>

## Known Issues

### Critical Issues

1. **Missing `async` keyword** (🔴 Critical)
   - **Issue**: `async def` functions appear as regular `def`
   - **Impact**: Breaks async code semantics completely
   - **Status**: Fix in progress (#101)

2. **Missing metaclass parameters** (🔴 Critical)
   - **Issue**: `class MyClass(Base, metaclass=Meta)` loses metaclass
   - **Impact**: Metaclass behavior lost
   - **Workaround**: Document metaclass usage in docstring

### Major Issues

3. **Nested classes not captured** (🟡 Major)
   - **Issue**: Inner classes are completely omitted
   - **Impact**: Loses important structural information
   - **Workaround**: Refactor to module-level classes

4. **Docstring format changes** (🟡 Major)
   - **Issue**: Multi-line docstrings converted to comments
   - **Impact**: Docstrings are runtime-accessible, comments aren't
   - **Status**: Requires parser enhancement

### Minor Issues

5. **Complex f-string parsing** (🟢 Minor)
   - **Issue**: f-strings with nested braces may fail
   - **Example**: `f"{x:.{precision}f}"` 
   - **Workaround**: Use `.format()` method

6. **Type comment support** (🟢 Minor)
   - **Issue**: PEP 484 type comments not extracted
   - **Example**: `x = []  # type: List[int]`
   - **Workaround**: Use PEP 526 annotations

## Best Practices

### 1. **Write Type-Annotated Code**

AI Distiller works best with fully typed Python:

```python
# Good - Full type information preserved
def process(data: List[Dict[str, Any]], 
           options: Optional[ProcessOptions] = None) -> Result:
    pass

# Less optimal - Types lost
def process(data, options=None):
    pass
```

### 2. **Use Explicit Visibility**

Make your intentions clear:

```python
class Service:
    def public_method(self):  # Clearly public
        pass
    
    def _internal_method(self):  # Clearly internal
        pass
    
    def __private_method(self):  # Clearly private
        pass
```

### 3. **Structure for AI Consumption**

- Keep modules focused and under 1000 lines
- Use descriptive names (they're preserved!)
- Add type hints everywhere possible
- Write clear, concise docstrings

### 4. **Pre-process Your Code**

Before distilling:

```bash
# Format code consistently
black .

# Add type stubs
mypy --install-types

# Remove secrets
detect-secrets scan --baseline .secrets.baseline
```

## Integration Examples

### Direct CLI Usage

```bash
# Generate optimal context for AI assistants
aid ./myproject --format text \
    --strip "non-public,implementation,comments" \
    --output context.txt

# Include in your prompt
cat <<EOF > prompt.txt
Here's my codebase structure:

$(cat context.txt)

Please help me implement a new feature that...
EOF
```

### MCP Integration with Claude Desktop

When using AI Distiller through MCP, Claude can directly access your codebase:

**Setup** (in Claude Desktop config.json):
```json
{
  "mcpServers": {
    "ai-distiller": {
      "command": "/usr/local/bin/aid",
      "args": ["--mcp-server"],
      "env": {
        "AID_ROOT": "/Users/you/projects/myapp"
      }
    }
  }
}
```

**Example Conversation:**

> **You**: I need to add a caching layer to our user authentication. First, show me the current structure of our auth module.

> **Claude**: I'll analyze your authentication module structure using AI Distiller.

*Claude automatically calls:*
```
distillFile("auth/authenticator.py", strip_implementation=true)
listFiles("auth/", pattern="*.py")
```

*Claude receives the distilled structure and responds:*

> Based on the analysis, your authentication module has:
> - `Authenticator` class with methods: `authenticate()`, `validate_token()`, `refresh_token()`
> - `TokenManager` class handling JWT operations
> - `UserSession` dataclass for session management
>
> I can see that adding a caching layer would benefit the `validate_token()` method which currently makes database calls. Here's my implementation plan...

**Advanced Example - Feature Implementation:**

> **You**: We need to implement API rate limiting. Show me all middleware classes and where request handling happens.

> **Claude**: Let me analyze your middleware architecture comprehensively.

*Claude calls multiple MCP tools:*
```
distillDirectory("middleware/", strip_implementation=true)
search("class.*Middleware", is_regex=true)
distillFile("app.py", strip_implementation=true)
```

> I found your complete middleware architecture:
>
> From `distillDirectory` on `middleware/`:
> ```
> BaseMiddleware (abstract):
>   + process_request(request: Request) -> Optional[Response]
>   + process_response(request: Request, response: Response) -> Response
>
> AuthMiddleware(BaseMiddleware):
>   + __init__(jwt_secret: str, excluded_paths: List[str])
>   + process_request(request: Request) -> Optional[Response]
>
> CORSMiddleware(BaseMiddleware):
>   + __init__(allowed_origins: List[str], allowed_methods: List[str])
>   + process_response(request: Request, response: Response) -> Response
>
> LoggingMiddleware(BaseMiddleware):
>   + __init__(logger: Logger, log_body: bool = False)
>   + process_request(request: Request) -> Optional[Response]
>   + process_response(request: Request, response: Response) -> Response
> ```
>
> Your request flow: `app.py:handle_request()` → middleware chain (ordered list)
>
> Here's a rate limiting middleware that fits your established patterns:
>
> ```python
> class RateLimitMiddleware(BaseMiddleware):
>     def __init__(self, requests_per_minute: int = 60, 
>                  redis_client: Optional[Redis] = None):
>         self.limit = requests_per_minute
>         self.redis = redis_client or {}  # Fallback to memory
>     
>     def process_request(self, request: Request) -> Optional[Response]:
>         client_id = self._get_client_id(request)
>         if self._is_rate_limited(client_id):
>             return Response(
>                 status=429, 
>                 body={"error": "Rate limit exceeded"},
>                 headers={"Retry-After": str(self._get_retry_after(client_id))}
>             )
>         return None
> ```

**Power of distillDirectory**: Notice how `distillDirectory` gave us the complete middleware namespace structure in one call, showing all classes with their methods and signatures. This is much more efficient than calling `distillFile` on each file separately!

### With Documentation Tools

```bash
# Extract public API for docs
aid ./src --strip "non-public,implementation" \
    --format json-structured | \
    jq '.files[].symbols[] | select(.visibility == "public")'
```

### CI/CD Integration

```yaml
# .github/workflows/api-check.yml
- name: Check API surface
  run: |
    aid . --strip "non-public,implementation" > api-new.txt
    diff api-baseline.txt api-new.txt || exit 1
```

## Language-Specific Tips

1. **Handle Dynamic Imports**: Static analysis can't detect dynamic imports. Document them:
   ```python
   # AI-DISTILLER: Dynamic imports from plugins/* 
   ```

2. **Type Aliases**: Define at module level for better extraction:
   ```python
   UserID = NewType('UserID', int)
   SessionData = Dict[str, Any]
   ```

3. **Protocol vs ABC**: Prefer `Protocol` for better type information:
   ```python
   # Better for AI Distiller
   class Drawable(Protocol):
       def draw(self) -> None: ...
   ```

4. **Async Context Managers**: Currently shown as regular methods:
   ```python
   async def __aenter__(self): ...  # Shows as regular method
   ```

## Comparison with Other Tools

| Tool | Purpose | Python Support | AI-Optimized |
|------|---------|---------------|--------------|
| **AI Distiller** | Code structure extraction | Full | ✅ Yes |
| `ast.parse()` | AST analysis | Python only | ❌ No |
| `inspect` | Runtime introspection | Python only | ❌ No |
| `cloc` | Line counting | Basic | ❌ No |
| `ctags` | Symbol indexing | Limited | ❌ No |

## Troubleshooting

### "Parser failed with syntax error"

Python code must be syntactically valid. Run:
```bash
python -m py_compile yourfile.py
```

### "Missing async keyword in output"

Known issue. Workaround: Add comment markers:
```python
async def fetch_data():  # async
    pass
```

### "Type information lost"

Ensure you're using Python 3.9+ type hint syntax:
```python
# Old (may not parse correctly)
from typing import List
items: List[str]

# New (better support)
items: list[str]
```

## Future Enhancements

- [ ] Full `async`/`await` support
- [ ] Metaclass parameter preservation  
- [ ] Nested class extraction
- [ ] Type comment (PEP 484) support
- [ ] Protocol implementation detection
- [ ] Import cycle detection
- [ ] Stub file (.pyi) support

## Contributing

Help improve Python support! See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

Key areas needing help:
- Tree-sitter grammar improvements
- Complex type annotation examples
- Real-world test cases
- Performance optimization

---

<sub>Documentation generated for AI Distiller v0.2.0</sub>