# Output Formats

AI Distiller supports multiple output formats optimized for different use cases.

## Text Format (Default)

The text format is the most compact representation, optimized for AI consumption.

### Features
- Minimal syntax overhead
- Natural code-like appearance
- Maximum information density
- Clear file boundaries with `<file path="...">` tags

### Visibility Symbols

The text format uses prefixes to indicate member visibility:

- **public**: no symbol (default)
- **private**: `-` (minus)
- **protected**: `*` (asterisk)
- **internal/package-private**: `~` (tilde)

### Example

```
<file path="src/user_service.py">
from typing import List, Optional

class UserService:
    -_cache: dict           # private field
    *_logger: Logger        # protected field
    ~_config: Config        # internal field
    
    def get_user(id: int)   # public method (no prefix)
    -_validate()            # private method
    *log_access()           # protected method
    ~process_internal()     # internal method
</file>
```

## Markdown Format

Human-readable format with emojis and formatting.

### Features
- Visual structure with emojis
- Line number annotations
- Collapsible sections
- Good for documentation

### Example

```markdown
# src/user_service.py

## Structure

📥 **Import** from `typing` import `List`, `Optional` <sub>L1</sub>

🏛️ **Class** `UserService` <sub>L5-45</sub>
  🔧 **Function** `get_user`(`id`: `int`) → `User` <sub>L12-18</sub>
  🔧 **Function** `_validate` _private_() → `bool` <sub>L20-25</sub>
```

## JSON Lines Format

One JSON object per line, ideal for streaming and processing.

### Example

```jsonl
{"type":"file","path":"src/user.py","language":"python"}
{"type":"class","name":"User","visibility":"public","line":5}
{"type":"function","name":"get_user","visibility":"public","parameters":[{"name":"id","type":"int"}],"returns":"User"}
```

## JSON Structured Format

Complete hierarchical representation with full semantic information.

### Example

```json
{
  "files": [{
    "path": "src/user.py",
    "language": "python",
    "classes": [{
      "name": "User",
      "visibility": "public",
      "methods": [{
        "name": "get_user",
        "visibility": "public",
        "parameters": [{"name": "id", "type": "int"}],
        "returns": "User"
      }]
    }]
  }]
}
```

## XML Format

Structured XML representation for systems requiring XML input.

### Example

```xml
<distilled>
  <file path="src/user.py" language="python">
    <class name="User" visibility="public">
      <method name="get_user" visibility="public">
        <parameter name="id" type="int"/>
        <returns>User</returns>
      </method>
    </class>
  </file>
</distilled>
```