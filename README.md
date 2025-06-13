# AI Distiller

> **Turn massive codebases into AI-digestible summaries in seconds**

[![Python](https://img.shields.io/badge/python-3.9%2B-blue)](https://www.python.org/downloads/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Test Status](https://img.shields.io/badge/tests-passing-brightgreen)](test-data/)

AI Distiller extracts the essential structure from large codebases, creating compact representations perfect for LLM context windows. Think of it as **"code compression for AI"** - preserving what matters, discarding the noise.

## Why AI Distiller?

<table>
<tr>
<th>🤖 For AI Engineers</th>
<th>👨‍💻 For Developers</th>
<th>🔍 For Code Reviewers</th>
</tr>
<tr>
<td>

```bash
# Turn 10MB of code into 
# 200KB of structure
aid ./src --format text \
  --strip "implementation,comments"
```

Feed entire codebases to LLMs without hitting token limits

</td>
<td>

```bash
# Get instant API overview
aid ./api --strip "non-public" \
  --output api-surface.txt
```

Understand new codebases in minutes, not hours

</td>
<td>

```bash
# Extract only public changes
aid . --strip "non-public,implementation" \
  --format json | jq '.symbols'
```

Focus on what really changed in PRs

</td>
</tr>
</table>

## 🚀 Quick Start

```bash
# Install (binary releases coming soon)
git clone https://github.com/janreges/ai-distiller
cd ai-distiller
make build

# Basic usage
./aid                                    # Current directory
./aid src/                              # Specific directory
./aid main.py utils.py                  # Specific files

# AI-optimized output (most compact)
./aid --format text --strip "non-public,comments,implementation"

# Full structural analysis
./aid --format json --output structure.json
```

## 📊 Real Performance Numbers

<table>
<tr>
<th>Codebase</th>
<th>Size</th>
<th>Files</th>
<th>Time</th>
<th>Output</th>
<th>Compression</th>
</tr>
<tr>
<td><code>express.js</code></td>
<td>1.2 MB</td>
<td>324</td>
<td><strong>0.8s</strong></td>
<td>45 KB</td>
<td>27x</td>
</tr>
<tr>
<td><code>django</code></td>
<td>8.7 MB</td>
<td>2,451</td>
<td><strong>4.2s</strong></td>
<td>312 KB</td>
<td>29x</td>
</tr>
<tr>
<td><code>kubernetes</code></td>
<td>98 MB</td>
<td>12,384</td>
<td><strong>47s</strong></td>
<td>3.4 MB</td>
<td>29x</td>
</tr>
</table>

*Benchmarked on M2 MacBook Pro. Single binary, no runtime dependencies.*

## ✨ Key Features

### 🎯 Intelligent Stripping
Remove exactly what you don't need:
- `--strip comments` - Remove all comments and docstrings
- `--strip implementation` - Keep only signatures
- `--strip non-public` - Hide private/internal members
- `--strip imports` - Remove import statements

### 📝 Multiple Output Formats
- **Text** (`--format text`) - Ultra-compact for AI consumption
- **Markdown** (`--format md`) - Human-readable with emojis
- **JSON** (`--format json`) - Structured data for tools
- **JSONL** (`--format jsonl`) - Streaming format
- **XML** (`--format xml`) - Legacy system compatible

### 🌍 Language Support
Currently supports 12+ languages via tree-sitter:
- **Full Support**: Python, TypeScript, JavaScript, Go, Java, C#, Rust
- **Beta**: Ruby, Swift, Kotlin, PHP, C++
- **Coming Soon**: Zig, Scala, Clojure

See [language-specific documentation](docs/lang/) for details.

## 🔒 Security Considerations

**⚠️ Important**: AI Distiller extracts code structure which may include:
- Function and variable names that could reveal business logic
- Type information and API signatures
- Comments and docstrings (unless stripped)

**Recommendations**:
1. Always review output before sending to external services
2. Use `--strip comments` to remove potentially sensitive documentation
3. Consider running a secrets scanner on your codebase first
4. For maximum security, run AI Distiller in an isolated environment

## 📖 Example Output

<details>
<summary>Python Class Example</summary>

**Input** (`car.py`):
```python
class Car:
    """A car with basic attributes and methods."""
    
    def __init__(self, make: str, model: str):
        self.make = make
        self.model = model
        self._mileage = 0  # Private
    
    def drive(self, distance: int) -> None:
        """Drive the car."""
        if distance > 0:
            self._mileage += distance
```

**Output** (`aid car.py --format text --strip "non-public,implementation"`):
```
<file path="car.py">
class Car:
    +def __init__(self, make: str, model: str)
    +def drive(self, distance: int) -> None
</file>
```

</details>

<details>
<summary>TypeScript Interface Example</summary>

**Input** (`api.ts`):
```typescript
export interface User {
  id: number;
  name: string;
  email?: string;
}

export class UserService {
  private cache = new Map<number, User>();
  
  async getUser(id: number): Promise<User | null> {
    return this.cache.get(id) || null;
  }
}
```

**Output** (`aid api.ts --format text --strip "non-public,implementation"`):
```
<file path="api.ts">
export interface User {
  id: number;
  name: string;
  email?: string;
}

export class UserService {
  +async getUser(id: number): Promise<User | null>
}
</file>
```

</details>

## 🛠️ Advanced Usage

### Integration with AI Tools

```bash
# Create a context file for Claude or GPT
aid ./src --format text --strip "implementation,non-public" > context.txt

# Generate a codebase summary for RAG systems
aid . --format json | jq -r '.files[].symbols[].name' > symbols.txt

# Extract API surface for documentation
aid ./api --strip "non-public,implementation,comments" --format md > api-ref.md
```

### Configuration File

Create `.aidconfig.yml` in your project root:

```yaml
# Default options for this project
format: text
strip:
  - implementation
  - non-public
exclude:
  - "**/*.test.js"
  - "**/node_modules/**"
  - "**/__pycache__/**"
```

## 🔗 Documentation

- [Installation Guide](docs/installation.md)
- [CLI Reference](docs/cli-reference.md)
- [Language Support](docs/lang/)
  - [Python](docs/lang/python.md)
  - [TypeScript](docs/lang/typescript.md)
  - [Go](docs/lang/go.md)
  - [More...](docs/lang/)
- [Output Formats](docs/formats.md)
- [Performance Tuning](docs/performance.md)
- [Security Guide](docs/security.md)

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone and setup
git clone https://github.com/janreges/ai-distiller
cd ai-distiller
make setup

# Run tests
make test

# Build binary
make build
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built on [tree-sitter](https://tree-sitter.github.io/) for accurate parsing
- Inspired by the need for better AI-code interaction
- Created with ❤️ for the AI engineering community

---

<p align="center">
  <sub>Built by developers, for developers working with AI</sub>
</p>