//go:build cgo
// +build cgo

package tree_sitter_typescript

// #cgo CFLAGS: -std=c11 -fPIC -I./source/typescript/src -I./source/common
// #include "./source/typescript/src/parser.c"
// #include "./source/typescript/src/scanner.c"
import "C"

import "unsafe"

// Language returns the tree-sitter Language for TypeScript.
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_typescript())
}