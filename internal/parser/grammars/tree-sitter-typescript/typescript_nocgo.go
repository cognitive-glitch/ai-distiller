//go:build !cgo
// +build !cgo

package tree_sitter_typescript

import "unsafe"

// Language returns nil when CGO is not available
func Language() unsafe.Pointer {
	return nil
}

// LanguageTSX returns nil when CGO is not available
func LanguageTSX() unsafe.Pointer {
	return nil
}