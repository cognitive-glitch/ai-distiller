package parser

// #cgo CFLAGS: -I../../grammars/tree-sitter-python/src -std=c99
// #cgo LDFLAGS: -L../../grammars/tree-sitter-python/src -ltree-sitter-python
// #include "tree_sitter/api.h"
// #include <stdlib.h>
// 
// // Tree-sitter Python grammar functions
// extern TSLanguage *tree_sitter_python();
import "C"
import (
	"fmt"
	"unsafe"
)

// Re-export tree-sitter types for easier use
type (
	Parser   = C.TSParser
	Tree     = C.TSTree
	Node     = C.TSNode
	Language = C.TSLanguage
)

// PythonParser wraps tree-sitter parser for Python
type PythonParser struct {
	parser *Parser
	lang   *Language
}

// NewPythonParser creates a new Python parser
func NewPythonParser() (*PythonParser, error) {
	parser := C.ts_parser_new()
	if parser == nil {
		return nil, fmt.Errorf("failed to create parser")
	}

	lang := C.tree_sitter_python()
	if lang == nil {
		return nil, fmt.Errorf("failed to load Python grammar")
	}

	if !C.ts_parser_set_language(parser, lang) {
		C.ts_parser_delete(parser)
		return nil, fmt.Errorf("failed to set language")
	}

	return &PythonParser{
		parser: parser,
		lang:   lang,
	}, nil
}

// Parse parses Python source code and returns a tree
func (p *PythonParser) Parse(source []byte) (*Tree, error) {
	if len(source) == 0 {
		return nil, fmt.Errorf("empty source")
	}

	// Create C string from source
	csource := C.CString(string(source))
	defer C.free(unsafe.Pointer(csource))

	// Parse the source
	tree := C.ts_parser_parse_string(
		p.parser,
		nil, // oldTree
		csource,
		C.uint32_t(len(source)),
	)

	if tree == nil {
		return nil, fmt.Errorf("parse failed")
	}

	return tree, nil
}

// Close cleans up parser resources
func (p *PythonParser) Close() {
	if p.parser != nil {
		C.ts_parser_delete(p.parser)
		p.parser = nil
	}
}

// Node wrapper methods for easier use

// String returns S-expression representation of the node
func (n *Node) String() string {
	// This is a simplified version - real implementation would use
	// ts_node_string() but that requires more complex memory management
	nodeType := C.ts_node_type((*C.TSNode)(n))
	return C.GoString(nodeType)
}

// ChildCount returns the number of children
func (n *Node) ChildCount() uint32 {
	return uint32(C.ts_node_child_count((*C.TSNode)(n)))
}

// Child returns the child at the given index
func (n *Node) Child(index int) *Node {
	child := C.ts_node_child((*C.TSNode)(n), C.uint32_t(index))
	return (*Node)(&child)
}

// HasError checks if the node contains any errors
func (n *Node) HasError() bool {
	return bool(C.ts_node_has_error((*C.TSNode)(n)))
}

// Tree wrapper methods

// RootNode returns the root node of the tree
func (t *Tree) RootNode() *Node {
	root := C.ts_tree_root_node((*C.TSTree)(t))
	return (*Node)(&root)
}

// Delete cleans up the tree
func (t *Tree) Delete() {
	C.ts_tree_delete((*C.TSTree)(t))
}