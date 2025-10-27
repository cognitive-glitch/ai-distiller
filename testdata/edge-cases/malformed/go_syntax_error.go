/**
 * Edge case: Go file with syntax errors.
 * Parser should handle gracefully without crashing.
 */
package main

// Valid struct
type ValidStruct struct {
    ID   int
    Name string
}

// Syntax error: Missing closing brace
func BrokenFunction(x, y int) int {
    return x + y
// Missing }

// Valid function after error
func ValidFunction() string {
    return "still parsing"
}

// Syntax error: Invalid receiver
func (r *) InvalidReceiver() {
}

// Syntax error: Missing type in struct field
type BrokenStruct struct {
    ID int
    Name  // Missing type
}

// Syntax error: Invalid interface
type BrokenInterface interface {
    Method(x int  // Missing closing paren
    AnotherMethod() string
}
