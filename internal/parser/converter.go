package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// Converter converts tree-sitter AST to our IR format
type Converter struct {
	language string
	source   []byte
	errors   []ir.DistilledError
}

// NewConverter creates a new AST to IR converter
func NewConverter(language string, source []byte) *Converter {
	return &Converter{
		language: language,
		source:   source,
		errors:   make([]ir.DistilledError, 0),
	}
}

// ConvertTree converts a tree-sitter tree to our IR format
func (c *Converter) ConvertTree(ctx context.Context, tree *Tree) (*ir.DistilledFile, error) {
	// Get root node
	root, err := tree.RootNode()
	if err != nil {
		return nil, fmt.Errorf("failed to get root node: %w", err)
	}
	
	// Create file node
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				StartColumn: 1,
			},
		},
		Language: c.language,
		Version:  "2.0.0", // IR version
		Children: make([]ir.DistilledNode, 0),
		Errors:   c.errors,
	}
	
	// Convert root node's children
	childCount, err := root.ChildCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get child count: %w", err)
	}
	
	for i := uint32(0); i < childCount; i++ {
		child, err := root.Child(i)
		if err != nil {
			continue
		}
		
		if node := c.convertNode(ctx, child); node != nil {
			file.Children = append(file.Children, node)
		}
	}
	
	// Update file location based on content
	if len(file.Children) > 0 {
		file.Location.EndLine = c.countLines(c.source)
		file.Location.EndColumn = c.lastLineLength(c.source)
	}
	
	return file, nil
}

// convertNode converts a tree-sitter node to our IR format
func (c *Converter) convertNode(ctx context.Context, node *Node) ir.DistilledNode {
	// Get node type
	nodeType, err := node.Type()
	if err != nil {
		return nil
	}
	
	// Check if it's an error node
	isError, _ := node.IsError()
	if isError {
		return c.createErrorNode(node, nodeType)
	}
	
	// Convert based on language and node type
	switch c.language {
	case "python":
		return c.convertPythonNode(ctx, node, nodeType)
	case "go":
		return c.convertGoNode(ctx, node, nodeType)
	case "javascript", "typescript":
		return c.convertJSNode(ctx, node, nodeType)
	default:
		return c.convertGenericNode(ctx, node, nodeType)
	}
}

// convertPythonNode converts Python-specific nodes
func (c *Converter) convertPythonNode(ctx context.Context, node *Node, nodeType string) ir.DistilledNode {
	switch nodeType {
	case "module":
		// Root module - process children
		return nil // Children are processed at file level
		
	case "class_definition":
		return c.createClassNode(node, nodeType)
		
	case "function_definition":
		return c.createFunctionNode(node, nodeType)
		
	case "import_statement", "import_from_statement":
		return c.createImportNode(node, nodeType)
		
	case "comment":
		return c.createCommentNode(node, nodeType)
		
	default:
		// For other nodes, check if they're named
		if named, _ := node.IsNamed(); named {
			return c.convertGenericNode(ctx, node, nodeType)
		}
		return nil
	}
}

// convertGoNode converts Go-specific nodes
func (c *Converter) convertGoNode(ctx context.Context, node *Node, nodeType string) ir.DistilledNode {
	switch nodeType {
	case "source_file":
		// Root - process children
		return nil
		
	case "package_clause":
		return c.createPackageNode(node, nodeType)
		
	case "import_declaration":
		return c.createImportNode(node, nodeType)
		
	case "type_declaration":
		return c.createTypeNode(node, nodeType)
		
	case "function_declaration", "method_declaration":
		return c.createFunctionNode(node, nodeType)
		
	case "comment":
		return c.createCommentNode(node, nodeType)
		
	default:
		if named, _ := node.IsNamed(); named {
			return c.convertGenericNode(ctx, node, nodeType)
		}
		return nil
	}
}

// convertJSNode converts JavaScript/TypeScript nodes
func (c *Converter) convertJSNode(ctx context.Context, node *Node, nodeType string) ir.DistilledNode {
	switch nodeType {
	case "program":
		// Root - process children
		return nil
		
	case "class_declaration":
		return c.createClassNode(node, nodeType)
		
	case "function_declaration", "arrow_function", "function_expression":
		return c.createFunctionNode(node, nodeType)
		
	case "import_statement":
		return c.createImportNode(node, nodeType)
		
	case "comment":
		return c.createCommentNode(node, nodeType)
		
	default:
		if named, _ := node.IsNamed(); named {
			return c.convertGenericNode(ctx, node, nodeType)
		}
		return nil
	}
}

// convertGenericNode handles generic named nodes
func (c *Converter) convertGenericNode(ctx context.Context, node *Node, nodeType string) ir.DistilledNode {
	// For now, skip generic nodes
	// In a full implementation, we'd handle more node types
	return nil
}

// Node creation helpers

func (c *Converter) createErrorNode(node *Node, nodeType string) ir.DistilledNode {
	loc := c.getNodeLocation(node)
	
	errorNode := &ir.DistilledError{
		BaseNode: ir.BaseNode{Location: loc},
		Message:  fmt.Sprintf("Syntax error: %s", nodeType),
		Severity: "error",
		Code:     "PARSE_ERROR",
	}
	
	// Also add to file errors
	c.errors = append(c.errors, *errorNode)
	
	return errorNode
}

func (c *Converter) createPackageNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract package name from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

func (c *Converter) createImportNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract import details from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

func (c *Converter) createClassNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract class details from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

func (c *Converter) createFunctionNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract function details from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

func (c *Converter) createTypeNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract type details from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

func (c *Converter) createCommentNode(node *Node, nodeType string) ir.DistilledNode {
	// This would extract comment text from the node
	// For now, return nil as we need more tree-sitter integration
	return nil
}

// Helper methods

func (c *Converter) getNodeLocation(node *Node) ir.Location {
	startByte, _ := node.StartByte()
	endByte, _ := node.EndByte()
	
	// Calculate line and column from byte offsets
	startLine, startCol := c.byteToLineCol(startByte)
	endLine, endCol := c.byteToLineCol(endByte)
	
	return ir.Location{
		StartLine:   startLine,
		StartColumn: startCol,
		EndLine:     endLine,
		EndColumn:   endCol,
		StartByte:   int(startByte),
		EndByte:     int(endByte),
	}
}

func (c *Converter) byteToLineCol(offset uint32) (line, col int) {
	line = 1
	col = 1
	
	for i := uint32(0); i < offset && i < uint32(len(c.source)); i++ {
		if c.source[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	
	return line, col
}

func (c *Converter) countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	
	lines := 1
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	
	return lines
}

func (c *Converter) lastLineLength(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	
	lastNewline := strings.LastIndexByte(string(data), '\n')
	if lastNewline == -1 {
		return len(data)
	}
	
	return len(data) - lastNewline - 1
}

func (c *Converter) getNodeText(node *Node) (string, error) {
	startByte, err := node.StartByte()
	if err != nil {
		return "", err
	}
	
	endByte, err := node.EndByte()
	if err != nil {
		return "", err
	}
	
	if startByte >= uint32(len(c.source)) || endByte > uint32(len(c.source)) {
		return "", fmt.Errorf("node bounds out of range")
	}
	
	return string(c.source[startByte:endByte]), nil
}