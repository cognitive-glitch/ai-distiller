package parser

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

// Parser represents a tree-sitter parser instance
type Parser struct {
	module     *WASMModule
	parserPtr  uint32
	memory     api.Memory
}

// Tree represents a parsed syntax tree
type Tree struct {
	parser  *Parser
	treePtr uint32
}

// Node represents a node in the syntax tree
type Node struct {
	tree     *Tree
	nodeData [4]uint32 // tree-sitter TSNode is 16 bytes (4 * uint32)
}

// Point represents a position in the source code
type Point struct {
	Row    uint32
	Column uint32
}

// NewParser creates a new parser for the given WASM module
func NewParser(module *WASMModule) (*Parser, error) {
	if module.ParserNew == nil {
		return nil, fmt.Errorf("parser_new function not found")
	}

	// Call ts_parser_new to create a new parser
	results, err := module.ParserNew.Call(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("parser_new returned no results")
	}

	parserPtr := uint32(results[0])
	if parserPtr == 0 {
		return nil, fmt.Errorf("parser_new returned null pointer")
	}

	parser := &Parser{
		module:    module,
		parserPtr: parserPtr,
		memory:    module.Module.Memory(),
	}

	// Set the language
	if err := parser.SetLanguage(); err != nil {
		parser.Delete()
		return nil, err
	}

	return parser, nil
}

// SetLanguage sets the language for the parser
func (p *Parser) SetLanguage() error {
	if p.module.ParserSetLanguage == nil || p.module.TreeSitterLanguage == nil {
		return fmt.Errorf("required functions not found")
	}

	// Get the language pointer
	langResults, err := p.module.TreeSitterLanguage.Call(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get language: %w", err)
	}

	if len(langResults) == 0 {
		return fmt.Errorf("language function returned no results")
	}

	langPtr := uint32(langResults[0])

	// Set the language
	results, err := p.module.ParserSetLanguage.Call(context.Background(), uint64(p.parserPtr), uint64(langPtr))
	if err != nil {
		return fmt.Errorf("failed to set language: %w", err)
	}

	if len(results) > 0 && results[0] == 0 {
		return fmt.Errorf("failed to set language (returned false)")
	}

	return nil
}

// Parse parses the given source code
func (p *Parser) Parse(ctx context.Context, source []byte) (*Tree, error) {
	if p.module.ParserParse == nil {
		return nil, fmt.Errorf("parser_parse function not found")
	}

	// Allocate memory for source in WASM
	sourcePtr, err := p.allocateAndCopyBytes(source)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate source: %w", err)
	}
	defer p.free(sourcePtr)

	// Call ts_parser_parse
	// ts_parser_parse(parser, old_tree, input)
	// For now, we pass null for old_tree and a simple input structure
	results, err := p.module.ParserParse.Call(ctx,
		uint64(p.parserPtr),
		uint64(0), // old_tree = null
		uint64(sourcePtr), // Simplified - in reality need proper TSInput structure
		uint64(len(source)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("parse returned no results")
	}

	treePtr := uint32(results[0])
	if treePtr == 0 {
		return nil, fmt.Errorf("parse returned null tree")
	}

	return &Tree{
		parser:  p,
		treePtr: treePtr,
	}, nil
}

// Delete frees the parser resources
func (p *Parser) Delete() error {
	if p.module.ParserDelete != nil && p.parserPtr != 0 {
		_, err := p.module.ParserDelete.Call(context.Background(), uint64(p.parserPtr))
		p.parserPtr = 0
		return err
	}
	return nil
}

// RootNode returns the root node of the tree
func (t *Tree) RootNode() (*Node, error) {
	if t.parser.module.TreeRootNode == nil {
		return nil, fmt.Errorf("tree_root_node function not found")
	}

	// ts_tree_root_node returns a TSNode (16 bytes)
	// We need to handle this properly based on the WASM ABI
	results, err := t.parser.module.TreeRootNode.Call(context.Background(), uint64(t.treePtr))
	if err != nil {
		return nil, fmt.Errorf("failed to get root node: %w", err)
	}

	// For now, simulate a node structure
	// In reality, we'd need to properly unmarshal the TSNode structure
	node := &Node{
		tree: t,
		nodeData: [4]uint32{
			uint32(results[0]),
			0, 0, 0, // Simplified for PoC
		},
	}

	return node, nil
}

// Delete frees the tree resources
func (t *Tree) Delete() error {
	if t.parser.module.TreeDelete != nil && t.treePtr != 0 {
		_, err := t.parser.module.TreeDelete.Call(context.Background(), uint64(t.treePtr))
		t.treePtr = 0
		return err
	}
	return nil
}

// Type returns the type of the node
func (n *Node) Type() (string, error) {
	if n.tree.parser.module.NodeType == nil {
		return "", fmt.Errorf("node_type function not found")
	}

	// Call ts_node_type which returns a string pointer
	results, err := n.tree.parser.module.NodeType.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return "", fmt.Errorf("failed to get node type: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("node_type returned no results")
	}

	// Read string from WASM memory
	strPtr := uint32(results[0])
	return n.tree.parser.readString(strPtr)
}

// ChildCount returns the number of children
func (n *Node) ChildCount() (uint32, error) {
	if n.tree.parser.module.NodeChildCount == nil {
		return 0, fmt.Errorf("node_child_count function not found")
	}

	results, err := n.tree.parser.module.NodeChildCount.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return 0, fmt.Errorf("failed to get child count: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	return uint32(results[0]), nil
}

// Child returns the child at the given index
func (n *Node) Child(index uint32) (*Node, error) {
	if n.tree.parser.module.NodeChild == nil {
		return nil, fmt.Errorf("node_child function not found")
	}

	results, err := n.tree.parser.module.NodeChild.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]),
		uint64(index))
	if err != nil {
		return nil, fmt.Errorf("failed to get child: %w", err)
	}

	// Create child node from results
	child := &Node{
		tree: n.tree,
		nodeData: [4]uint32{
			uint32(results[0]),
			0, 0, 0, // Simplified
		},
	}

	return child, nil
}

// IsNamed returns whether the node is named (not anonymous)
func (n *Node) IsNamed() (bool, error) {
	if n.tree.parser.module.NodeIsNamed == nil {
		return false, fmt.Errorf("node_is_named function not found")
	}

	results, err := n.tree.parser.module.NodeIsNamed.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return false, fmt.Errorf("failed to check if named: %w", err)
	}

	return len(results) > 0 && results[0] != 0, nil
}

// IsError returns whether the node is an error node
func (n *Node) IsError() (bool, error) {
	if n.tree.parser.module.NodeIsError == nil {
		return false, fmt.Errorf("node_is_error function not found")
	}

	results, err := n.tree.parser.module.NodeIsError.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return false, fmt.Errorf("failed to check if error: %w", err)
	}

	return len(results) > 0 && results[0] != 0, nil
}

// StartByte returns the start byte offset of the node
func (n *Node) StartByte() (uint32, error) {
	if n.tree.parser.module.NodeStartByte == nil {
		return 0, fmt.Errorf("node_start_byte function not found")
	}

	results, err := n.tree.parser.module.NodeStartByte.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return 0, fmt.Errorf("failed to get start byte: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	return uint32(results[0]), nil
}

// EndByte returns the end byte offset of the node
func (n *Node) EndByte() (uint32, error) {
	if n.tree.parser.module.NodeEndByte == nil {
		return 0, fmt.Errorf("node_end_byte function not found")
	}

	results, err := n.tree.parser.module.NodeEndByte.Call(context.Background(),
		uint64(n.nodeData[0]), uint64(n.nodeData[1]),
		uint64(n.nodeData[2]), uint64(n.nodeData[3]))
	if err != nil {
		return 0, fmt.Errorf("failed to get end byte: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	return uint32(results[0]), nil
}

// Helper methods for memory management

func (p *Parser) allocateAndCopyBytes(data []byte) (uint32, error) {
	if p.memory == nil {
		return 0, fmt.Errorf("no memory available")
	}

	// Allocate memory (simplified - would use malloc in real implementation)
	ptr := uint32(0x1000) // Hardcoded for simplicity

	// Copy data to WASM memory
	if !p.memory.Write(ptr, data) {
		return 0, fmt.Errorf("failed to write to memory")
	}

	return ptr, nil
}

func (p *Parser) free(ptr uint32) {
	// No-op for now - would call free in real implementation
}

func (p *Parser) readString(ptr uint32) (string, error) {
	if p.memory == nil {
		return "", fmt.Errorf("no memory available")
	}

	// Read null-terminated string from memory
	var result []byte
	for i := uint32(0); ; i++ {
		b, ok := p.memory.ReadByte(ptr + i)
		if !ok || b == 0 {
			break
		}
		result = append(result, b)
	}

	return *(*string)(unsafe.Pointer(&result)), nil
}