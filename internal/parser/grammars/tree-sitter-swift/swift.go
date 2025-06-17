//go:build cgo
// +build cgo

package tree_sitter_swift

// Re-export from bindings/go
// This allows the module to work with the replace directive

// #cgo CFLAGS: -std=c11 -fPIC -I./src
// #include "src/scanner.c"
// #undef TOKEN_COUNT
// #include "src/parser.c"
// extern const TSLanguage *tree_sitter_swift();
import "C"

import "unsafe"

// Language returns the tree-sitter language for Swift
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_swift())
}