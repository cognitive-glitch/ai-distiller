//go:build cgo
// +build cgo

package tree_sitter_typescript

// #cgo CFLAGS: -std=c11 -fPIC -I./source/tsx/src -I./source/common
// #include "./source/tsx/src/parser.c"
// #include "./source/tsx/src/scanner.c"
import "C"

import "unsafe"

// LanguageTSX returns the tree-sitter Language for TSX.
func LanguageTSX() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_tsx())
}