# Go Language Support

## Overview

AI Distiller provides support for Go (Golang) source code analysis and distillation. The Go processor uses the standard `go/ast` parser to build a comprehensive Abstract Syntax Tree, which is then transformed into a simplified, AI-friendly format.

## Main Constructs

Go is a statically typed, compiled language with the following primary constructs:

- **Packages**: Organizational units (`package main`)
- **Imports**: Dependency management with aliasing support
- **Types**: Structs, interfaces, type aliases
- **Functions**: Standalone functions and methods with receivers
- **Variables/Constants**: Package and function-scoped declarations
- **Interfaces**: Behavioral contracts with implicit satisfaction
- **Generics**: Type parameters and constraints (Go 1.18+)

## Key Features

The Go distiller focuses on extracting the essential structure that helps AI understand:

1. **Type System**: All types are preserved including custom types, built-in types, and generic type parameters
2. **Interface Satisfaction**: Implicit interface implementation relationships (currently limited)
3. **Method Receivers**: Distinction between value and pointer receivers
4. **Visibility**: Uppercase = exported (public), lowercase = unexported (package-private)
5. **Import Aliases**: Support for renamed imports (`import m "math"`)

## Text Format Output

The default text format provides a Python-like representation optimized for AI consumption:

```
<file path="example.go">
package main

import (
    "fmt"
    m "math"
)

type Writer interface
    Write(p []byte) (int, error)

type Logger struct
    prefix string
    writer Writer

func NewLogger(prefix string) *Logger
    return &Logger{prefix: prefix}

func (l *Logger) Log(message string)
    fmt.Printf("[%s] %s\n", l.prefix, message)
</file>
```

## Known Issues

### 1. Parameter and Field Grouping
**Status**: Expanded in AST  
**Impact**: Low - Cosmetic difference

Grouped parameters and fields are expanded:
```go
// Original
func Move(x, y int)
type Point struct { X, Y int }

// Distilled
func Move(x int, y int)
type Point struct
    X int
    Y int
```

### 2. Single Import Formatting
**Status**: Always uses parentheses  
**Impact**: Low - Cosmetic issue

Single imports are formatted with parentheses like multi-imports:
```go
// Original
import "fmt"

// Distilled
import (
    "fmt"
)
```

### 3. Implementation Details
**Status**: Some simplification  
**Impact**: Medium - Implementation may differ

Complex implementations may be simplified. For example, compound assignments like `+=` might appear as simple assignments.

### 4. Generic Type Parameters
**Status**: Basic support  
**Impact**: Medium - Complex constraints may not be fully preserved

Basic generic syntax is supported, but complex type constraints might be simplified.

### 5. CGO Support
**Status**: Minimal  
**Impact**: Low - C code blocks not properly preserved

CGO import "C" and associated C code blocks are not handled correctly.

## Examples

<details open><summary>Basic Go Constructs</summary><blockquote>
  <details><summary>construct_1_basic.go - source</summary><blockquote>

```go
// Package basic provides fundamental Go constructs.
// It serves as the baseline for parser testing.
package basic

import (
	"fmt" // Standard library import
	m "math" // Aliased import
)

// Global constant Pi, a fundamental value.
const Pi = 3.14159

// var block for multiple declarations.
var (
	// IsEnabled controls a feature. Doc comments for vars are important.
	IsEnabled = true
	// UserCount is a package-level counter.
	UserCount int64 = 100
)

// Add sums two integers. A trivial function.
// It tests basic function declaration and parameter parsing.
func Add(x int, y int) int { // Line comment on the function signature
	// A comment inside the function body.
	z := x + y // Short variable declaration
	var result = z // Standard variable declaration
	_ = m.Abs(-1) // Using an aliased import
	fmt.Println(result)
	return result
}
```

  </blockquote></details>
  <details open><summary>Default compact AI-friendly version (`default output (public only, no implementation)`)</summary><blockquote>

```
<file path="construct_1_basic.go">
package basic

import (
    "fmt"
    m "math"
)

const Pi = 3.14159

var IsEnabled = true

var UserCount int64 = 100

func Add(x int, y int) int
</file>
```

  </blockquote></details>
  <details><summary>Full version (`--public=1 --protected=1 --internal=1 --private=1 --implementation=1`)</summary><blockquote>

```
<file path="construct_1_basic.go">
//  Package basic provides fundamental Go constructs.
//  It serves as the baseline for parser testing.
package basic

import (
    "fmt"
// Standard library import
    m "math"
// Aliased import
// Global constant Pi, a fundamental value.
)

const Pi = 3.14159

// var block for multiple declarations.
// IsEnabled controls a feature. Doc comments for vars are important.
var IsEnabled = true

// UserCount is a package-level counter.
var UserCount int64 = 100

// Add sums two integers. A trivial function.
// It tests basic function declaration and parameter parsing.
func Add(x int, y int) int
    z := x + y
    var result = z
    _ = m.Abs(-1)
    fmt.Println(result)
    return result
// Line comment on the function signature
// A comment inside the function body.
// Short variable declaration
// Standard variable declaration
// Using an aliased import
</file>
```

  </blockquote></details>
</blockquote></details>

<details><summary>Structs and Methods</summary><blockquote>
  <details><summary>construct_2_simple.go - source</summary><blockquote>

```go
package simple

import "io"

// Writer defines a simple interface for writing data.
type Writer interface {
	// Write accepts a byte slice and returns bytes written and an error.
	Write(p []byte) (n int, err error)
}

// Data represents a container for a piece of data.
// It demonstrates a basic struct with a single field.
type Data struct {
	value string // unexported field
}

// NewData is a constructor-like function, a common Go idiom.
func NewData(v string) *Data {
	return &Data{value: v}
}

// ReadValue returns the current value. This is a value receiver method.
// It can be called on both Data and *Data types.
func (d Data) ReadValue() string {
	return d.value
}

// UpdateValue modifies the value. This is a pointer receiver method.
// It can only be called on a *Data type.
func (d *Data) UpdateValue(v string) {
	d.value = v
}

// Ensure Data does not satisfy the io.Writer interface, but we use it
// to test how the distiller handles imported interfaces.
var _ io.Writer = (*customWriter)(nil) // compile-time check idiom
type customWriter struct{}
func (cw *customWriter) Write(p []byte) (n int, err error) { return 0, nil }
```

  </blockquote></details>
  <details open><summary>Default compact AI-friendly version</summary><blockquote>

```
<file path="construct_2_simple.go">
package simple

import (
    "io"
)

type Writer interface
    Write(p []byte) (n int, err error)

type Data struct

func (d Data) ReadValue() string

func (d *Data) UpdateValue(v string)

func NewData(v string) *Data
</file>
```

  </blockquote></details>
  <details><summary>Full version</summary><blockquote>

```
<file path="construct_2_simple.go">
package simple

import (
    "io"
)

type Writer interface
    Write(p []byte) (n int, err error)

type Data struct
    value string

func (d Data) ReadValue() string
    return d.value

func (d *Data) UpdateValue(v string)
    d.value = v

var _ io.Writer = (*customWriter)(nil)

type customWriter struct

func (cw *customWriter) Write(p []byte) (n int, err error)
    return 0, nil

func NewData(v string) *Data
    return &Data{value: v}
</file>
```

  </blockquote></details>
</blockquote></details>

<details><summary>Generics and Type Constraints</summary><blockquote>
  <details><summary>construct_4_complex.go - source</summary><blockquote>

```go
package complex

// Number is a constraint that permits any integer or floating-point type.
type Number interface {
	~int | ~int64 | ~float32 | ~float64
}

// Node is a generic struct representing a node in a linked list.
type Node[T any] struct {
	Value T
	Next  *Node[T]
}

// Map applies a function to each element of a slice, returning a new slice.
// This is a generic function with a closure.
func Map[T, V any](input []T, f func(T) V) []V {
	output := make([]V, len(input))
	for i, v := range input {
		output[i] = f(v)
	}
	return output
}

// ProcessNumericChan processes a channel of generic Nodes.
// It uses a type constraint.
func ProcessNumericChan[T Number](ch <-chan *Node[T]) T {
	var total T
	for node := range ch {
		total += node.Value // This operation is only valid because of the Number constraint.
	}
	// This function literal captures 'total' from its surrounding scope.
	defer func() {
		println("Final total:", total)
	}()
	return total
}
```

  </blockquote></details>
  <details open><summary>Default compact AI-friendly version</summary><blockquote>

```
<file path="construct_4_complex.go">
package complex

type Number interface
    ~int | ~int64 | ~float32 | ~float64

type Node[T any] struct
    Value T
    Next *Node[T]

func Map[T any, V any](input []T, f func(T) V) []V

func ProcessNumericChan[T Number](ch <-chan *Node[T]) T
</file>
```

  </blockquote></details>
</blockquote></details>

## Recent Fixes (2025-06-15)

1. **Generic type parameters** (✅ Fixed)
   - **Issue**: Generic type parameters were missing from function signatures
   - **Fix**: Added proper extraction of type parameters from function declarations in AST parser
   - **Impact**: Generics now properly displayed for functions like `Map[T any, V any]`

2. **Method association** (✅ Fixed)
   - **Issue**: Methods were displayed separately from their types
   - **Fix**: Implemented two-pass processing to properly associate methods with types
   - **Impact**: Methods now correctly nested under their receiver types

## Usage Notes

1. **Visibility Rules**: Go's visibility is determined by the first character of identifiers:
   - Uppercase first letter = Exported (public)
   - Lowercase first letter = Unexported (package-private)

2. **Import Organization**: The distiller preserves import grouping and formatting.

3. **Generic Support**: Full support for Go 1.18+ generics including type parameters, constraints, and type inference.

4. **Best Practices**: 
   - Use appropriate visibility flags to focus on public API
   - The text format provides the most compact representation for AI consumption
   - Being aware that methods may appear as standalone functions

## Future Improvements

The following improvements are planned:

1. **Parameter Grouping**: Preserve original parameter grouping syntax
2. **Import Formatting**: Distinguish between single and multi-import blocks
3. **Generic Type Parameters**: Complete support for complex generic constraints
4. **CGO Support**: Handle C code blocks in CGO files
5. **Implementation Accuracy**: Preserve exact implementation details

## Technical Implementation

The Go processor uses:
- **Parser**: Standard `go/parser` with `ParseComments` mode
- **AST Processing**: Native `go/ast` traversal
- **Type Analysis**: Basic visibility detection based on naming conventions
- **Formatter**: Custom text formatter with Go-specific syntax rules

The implementation follows a two-pass approach:
1. First pass: Collect types and functions
2. Second pass: Associate methods with their receiver types
3. Third pass: Analyze interface satisfaction
4. Fourth pass: Extract and position comments

For more details, see the implementation in `internal/language/golang/`.