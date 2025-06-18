# Building AI Distiller

## Requirements

AI Distiller requires CGO for tree-sitter parsers to support all programming languages. Without CGO, the tool would only support Go, which makes it essentially useless for its intended purpose.

## Building for Current Platform

```bash
# Standard build with full language support
make build

# Or directly:
CGO_ENABLED=1 go build -o build/aid ./cmd/aid
```

## Cross-Compilation

Cross-compilation with CGO requires appropriate toolchains for each target platform. For detailed instructions, see [docs/CROSS_COMPILATION.md](docs/CROSS_COMPILATION.md).

### Quick Setup

**Ubuntu/Debian:**
```bash
# For Linux ARM64 and Windows
sudo apt-get install gcc-aarch64-linux-gnu gcc-mingw-w64-x86-64

# For macOS (requires osxcross)
# See: tools/osxcross/INSTALLATION_STATUS.md
```

### Build All Platforms

```bash
# Run the build script
./scripts/build-releases.sh
```

This will create release archives for:
- ✅ Linux AMD64 (native) → `aid-linux-amd64.tar.gz`
- ✅ Linux ARM64 (if toolchain installed) → `aid-linux-arm64.tar.gz`
- ✅ Windows AMD64 (if mingw installed) → `aid-windows-amd64.zip`
- ❌ macOS AMD64/ARM64 (requires osxcross) → `aid-darwin-*.tar.gz`

Each archive contains the `aid` binary (or `aid.exe` for Windows) ready for distribution.

### Manual Cross-Compilation Examples

```bash
# Linux ARM64
CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o aid-linux-arm64 ./cmd/aid

# Windows AMD64
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o aid-windows-amd64.exe ./cmd/aid

# macOS (requires macOS host or osxcross)
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o aid-darwin-amd64 ./cmd/aid
```

## Binary Sizes

With full language support (CGO enabled):
- Linux: ~38 MB
- Windows: ~50 MB
- macOS: ~38 MB

## Testing Your Build

```bash
# Verify version
./aid --version

# Test multiple language support
echo 'print("Hello")' | ./aid --lang python --stdout
echo 'console.log("Hello")' | ./aid --lang javascript --stdout
echo 'package main' | ./aid --lang go --stdout
```

If any language returns "parser requires CGO to be enabled", the build is incorrect.

## Docker Build

For consistent builds across platforms:

```dockerfile
FROM golang:1.23
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build -o aid ./cmd/aid
```

## GitHub Actions

For automated builds, you'll need to set up appropriate build matrices with platform-specific toolchains. CGO cross-compilation in CI is complex and may require custom Docker images or self-hosted runners.