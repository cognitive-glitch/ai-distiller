//go:build !cgo
// +build !cgo

package tree_sitter_swift

import "unsafe"

// Language returns nil when CGO is not available
func Language() unsafe.Pointer {
	return nil
}