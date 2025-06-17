# AI Distiller Binary Size Optimization

## Problem

The AI Distiller binary is ~27MB due to tree-sitter parsers, particularly:
- Kotlin parser: 71MB source code (from fwcd/tree-sitter-kotlin)
- C# parser: 33MB
- TypeScript parser: 8.6MB
- Other parsers: 1-16MB each

## Root Cause

We use `github.com/smacker/go-tree-sitter` which bundles ALL parsers in one module. Even if we only import Kotlin and TypeScript from it, the entire module (with all parsers) gets linked.

## Current Solution

### 1. Immediate: Disable Kotlin Support

Kotlin parser alone contributes most to the size. By commenting it out in `registry.go`:

```go
// Register Kotlin processor - TEMPORARILY DISABLED (saves 17MB)
// kotlinProc := kotlin.NewProcessor()
// if err := processor.Register(kotlinProc); err != nil {
//     return err
// }
```

Binary size: 27MB → 23MB

### 2. Build Script with Variants

Created `build.sh` that builds two versions:
- **Full version** (27MB): All 12 languages
- **Lite version** (9.7MB): Without Kotlin, C#, C++, Java, Swift

```bash
./build.sh  # Creates both aid and aid-lite
```

## Implemented Solution: Official TypeScript Parser

Successfully replaced smacker TypeScript parser with official tree-sitter-typescript:

1. Created local Go bindings in `internal/parser/grammars/tree-sitter-typescript/`
2. Added official parser as git submodule
3. Separate compilation for TypeScript and TSX to avoid conflicts
4. Binary size reduced from 31MB → 24MB (with Kotlin disabled)

### Implementation Details

```go
// typescript.go - handles TypeScript
package tree_sitter_typescript

// #cgo CFLAGS: -std=c11 -fPIC -I./source/typescript/src -I./source/common
// #include "source/typescript/src/parser.c"
// #include "source/typescript/src/scanner.c"
import "C"

func Language() unsafe.Pointer {
    return unsafe.Pointer(C.tree_sitter_typescript())
}
```

```go
// tsx.go - handles TSX separately
package tree_sitter_typescript

// #cgo CFLAGS: -std=c11 -fPIC -I./source/tsx/src -I./source/common
// #include "source/tsx/src/parser.c"
// #include "source/tsx/src/scanner.c"
import "C"

func LanguageTSX() unsafe.Pointer {
    return unsafe.Pointer(C.tree_sitter_tsx())
}
```

## Future Improvements

### Option 1: Build Tags (Recommended)

Add build tags to conditionally include parsers:

```go
//go:build with_kotlin
// +build with_kotlin

import _ "github.com/smacker/go-tree-sitter/kotlin"
```

Build command:
```bash
go build -tags "with_kotlin,with_typescript" -ldflags="-s -w" -o aid
```

### Option 2: Replace Smacker Dependencies

For TypeScript, create custom Go bindings for official parser:
1. Use `github.com/tree-sitter/tree-sitter-typescript` (C code)
2. Write minimal Go CGO wrapper
3. Remove dependency on smacker/go-tree-sitter for TypeScript

### Option 3: WASM Parsers

Use WASM versions of parsers (300-700KB each vs 10-70MB native):
- Pure Go, no CGO
- 3-5× slower performance
- Much smaller binary

## Recommendations

1. **For production**: Use `aid-lite` (9.7MB) for most cases
2. **For full support**: Use `aid` (27MB) only when Kotlin/C#/Swift needed
3. **Long term**: Implement build tags for fine-grained control