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