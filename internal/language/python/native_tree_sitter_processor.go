package python

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	"github.com/janreges/ai-distiller/internal/ir"
)

// NativeTreeSitterProcessor uses native Go tree-sitter bindings
type NativeTreeSitterProcessor struct {
	parser   *sitter.Parser
	source   []byte
	filename string
}

// NewNativeTreeSitterProcessor creates a new native tree-sitter processor
func NewNativeTreeSitterProcessor() (*NativeTreeSitterProcessor, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_python.Language()))
	
	return &NativeTreeSitterProcessor{
		parser: parser,
	}, nil
}

// ProcessSource processes Python source code using tree-sitter
func (p *NativeTreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	
	// Parse with tree-sitter
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Close()
	
	// Create IR file
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   int(tree.RootNode().EndPoint().Row) + 1,
			},
		},
		Path:     filename,
		Language: "python",
		Version:  "3",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	// Process root node
	p.processNode(tree.RootNode(), file, nil)
	
	return file, nil
}

// processNode recursively processes tree-sitter nodes
func (p *NativeTreeSitterProcessor) processNode(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	
	// Only process named nodes unless it's a special case
	if !node.IsNamed() && nodeType != "comment" {
		// Process children for anonymous nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		return
	}
	
	switch nodeType {
	case "module":
		// Process all children at module level
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		
	case "import_statement":
		p.processImport(node, file, parent)
		
	case "import_from_statement":
		p.processFromImport(node, file, parent)
		
	case "class_definition":
		p.processClass(node, file, parent)
		
	case "function_definition":
		p.processFunction(node, file, parent, false)
		
	case "async_function_definition":
		p.processFunction(node, file, parent, true)
		
	case "decorated_definition":
		p.processDecoratedDefinition(node, file, parent)
		
	case "expression_statement":
		// Check for docstrings
		if node.ChildCount() > 0 {
			child := node.Child(0)
			if child.Type() == "string" {
				p.processDocstring(child, file, parent)
			}
		}
		
	case "comment":
		p.processComment(node, file, parent)
		
	default:
		// Process children for other node types
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
	}
}

// processImport processes import statements
func (p *NativeTreeSitterProcessor) processImport(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	imp := &ir.DistilledImport{
		BaseNode: p.nodeLocation(node),
		ImportType: "import",
		Symbols: []ir.ImportedSymbol{},
	}
	
	// Find module name and alias
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "dotted_name", "identifier":
			if imp.Module == "" {
				imp.Module = p.getNodeText(child)
			}
		case "aliased_import":
			// Has alias
			if child.ChildCount() >= 3 {
				nameNode := child.Child(0)
				aliasNode := child.Child(2) // Skip "as" keyword
				if nameNode != nil && aliasNode != nil {
					imp.Module = p.getNodeText(nameNode)
					imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
						Name:  imp.Module,
						Alias: p.getNodeText(aliasNode),
					})
				}
			}
		}
	}
	
	p.addNode(file, parent, imp)
}

// processFromImport processes from...import statements
func (p *NativeTreeSitterProcessor) processFromImport(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	imp := &ir.DistilledImport{
		BaseNode: p.nodeLocation(node),
		ImportType: "from",
		Symbols: []ir.ImportedSymbol{},
	}
	
	// Walk through children to find module and imports
	foundFrom := false
	foundImport := false
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		
		if childType == "from" {
			foundFrom = true
			continue
		}
		
		if childType == "import" {
			foundImport = true
			continue
		}
		
		if foundFrom && !foundImport {
			// This is the module name
			if childType == "dotted_name" || childType == "identifier" || childType == "relative_import" {
				imp.Module = p.getNodeText(child)
			}
		} else if foundImport {
			// These are the imported symbols
			switch childType {
			case "aliased_import":
				// import name as alias
				if child.ChildCount() >= 3 {
					nameNode := child.Child(0)
					aliasNode := child.Child(2) // Skip "as"
					if nameNode != nil {
						symbol := ir.ImportedSymbol{
							Name: p.getNodeText(nameNode),
						}
						if aliasNode != nil {
							symbol.Alias = p.getNodeText(aliasNode)
						}
						imp.Symbols = append(imp.Symbols, symbol)
					}
				}
				
			case "identifier":
				// Direct import
				imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
					Name: p.getNodeText(child),
				})
				
			case "*":
				// from module import *
				// Leave symbols empty to indicate star import
				
			case "import_from_as_names":
				// Multiple imports: process children
				for j := 0; j < int(child.ChildCount()); j++ {
					subChild := child.Child(j)
					if subChild.Type() == "aliased_import" {
						if subChild.ChildCount() >= 3 {
							nameNode := subChild.Child(0)
							aliasNode := subChild.Child(2)
							if nameNode != nil {
								symbol := ir.ImportedSymbol{
									Name: p.getNodeText(nameNode),
								}
								if aliasNode != nil {
									symbol.Alias = p.getNodeText(aliasNode)
								}
								imp.Symbols = append(imp.Symbols, symbol)
							}
						}
					} else if subChild.Type() == "identifier" {
						imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
							Name: p.getNodeText(subChild),
						})
					}
				}
			}
		}
	}
	
	p.addNode(file, parent, imp)
}

// processClass processes class definitions
func (p *NativeTreeSitterProcessor) processClass(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode: p.nodeLocation(node),
		Modifiers: []ir.Modifier{},
		Extends: []ir.TypeRef{},
		Children: []ir.DistilledNode{},
		Decorators: []string{},
	}
	
	// Walk through children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			if class.Name == "" {
				class.Name = p.getNodeText(child)
			}
			
		case "argument_list":
			// Base classes
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)
				if arg.IsNamed() && (arg.Type() == "identifier" || arg.Type() == "attribute") {
					class.Extends = append(class.Extends, ir.TypeRef{
						Name: p.getNodeText(arg),
					})
				}
			}
			
		case "block":
			// Class body
			p.processClassBody(child, file, class)
		}
	}
	
	// Set visibility
	if strings.HasPrefix(class.Name, "_") {
		class.Visibility = ir.VisibilityPrivate
	} else {
		class.Visibility = ir.VisibilityPublic
	}
	
	p.addNode(file, parent, class)
}

// processClassBody processes the body of a class
func (p *NativeTreeSitterProcessor) processClassBody(node *sitter.Node, file *ir.DistilledFile, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "function_definition", "async_function_definition":
			p.processFunction(child, file, class, child.Type() == "async_function_definition")
			
		case "decorated_definition":
			p.processDecoratedDefinition(child, file, class)
			
		case "expression_statement":
			// Check for docstrings or assignments
			if child.ChildCount() > 0 {
				expr := child.Child(0)
				if expr.Type() == "string" {
					p.processDocstring(expr, file, class)
				} else if expr.Type() == "assignment" {
					p.processAssignment(expr, file, class)
				}
			}
			
		case "assignment":
			p.processAssignment(child, file, class)
		}
	}
}

// processFunction processes function definitions
func (p *NativeTreeSitterProcessor) processFunction(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, isAsync bool) {
	fn := &ir.DistilledFunction{
		BaseNode: p.nodeLocation(node),
		Modifiers: []ir.Modifier{},
		Parameters: []ir.Parameter{},
		Decorators: []string{},
	}
	
	if isAsync {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
	}
	
	// Walk through children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			if fn.Name == "" {
				fn.Name = p.getNodeText(child)
			}
			
		case "parameters":
			p.processParameters(child, fn)
			
		case "type":
			// Return type annotation
			fn.Returns = &ir.TypeRef{
				Name: p.getNodeText(child),
			}
			
		case "block":
			// Function body
			// Extract docstring if present  
			if docstring := p.extractDocstring(child); docstring != "" {
				// Create a docstring comment as first child
				docComment := &ir.DistilledComment{
					BaseNode: p.nodeLocation(child.Child(0)),
					Text:     docstring,
					Format:   "docstring",
				}
				p.addNode(file, parent, docComment)
			}
			
			// Get implementation
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	// Set visibility
	if strings.HasPrefix(fn.Name, "_") {
		fn.Visibility = ir.VisibilityPrivate
	} else {
		fn.Visibility = ir.VisibilityPublic
	}
	
	p.addNode(file, parent, fn)
}

// processDecoratedDefinition processes decorated functions/classes
func (p *NativeTreeSitterProcessor) processDecoratedDefinition(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	decorators := []string{}
	var definition *sitter.Node
	
	// Collect decorators and find the definition
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "decorator" {
			// Extract decorator name (skip @ symbol)
			for j := 0; j < int(child.ChildCount()); j++ {
				decChild := child.Child(j)
				if decChild.Type() == "identifier" || decChild.Type() == "attribute" || decChild.Type() == "call" {
					decorators = append(decorators, p.getNodeText(decChild))
				}
			}
		} else if child.Type() == "function_definition" || child.Type() == "async_function_definition" || child.Type() == "class_definition" {
			definition = child
		}
	}
	
	if definition == nil {
		return
	}
	
	// Track current child count to find the newly added node
	var childCount int
	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			childCount = len(class.Children)
		}
	} else {
		childCount = len(file.Children)
	}
	
	// Process the definition
	switch definition.Type() {
	case "function_definition", "async_function_definition":
		p.processFunction(definition, file, parent, definition.Type() == "async_function_definition")
		
	case "class_definition":
		p.processClass(definition, file, parent)
	}
	
	// Add decorators to the newly created node
	var nodes []ir.DistilledNode
	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			nodes = class.Children
		}
	} else {
		nodes = file.Children
	}
	
	if childCount < len(nodes) {
		newNode := nodes[childCount]
		switch n := newNode.(type) {
		case *ir.DistilledFunction:
			n.Decorators = decorators
		case *ir.DistilledClass:
			n.Decorators = decorators
		}
	}
}

// processParameters processes function parameters
func (p *NativeTreeSitterProcessor) processParameters(node *sitter.Node, fn *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		// Skip punctuation
		if !child.IsNamed() {
			continue
		}
		
		param := &ir.Parameter{}
		
		switch child.Type() {
		case "identifier":
			param.Name = p.getNodeText(child)
			
		case "typed_parameter":
			// parameter: type
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" && param.Name == "" {
					param.Name = p.getNodeText(subChild)
				} else if subChild.Type() == "type" {
					param.Type = ir.TypeRef{
						Name: p.getNodeText(subChild),
					}
				}
			}
			
		case "default_parameter":
			// parameter = default
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" && param.Name == "" {
					param.Name = p.getNodeText(subChild)
				} else if subChild.IsNamed() && j > 0 { // Skip identifier and =
					param.DefaultValue = p.getNodeText(subChild)
				}
			}
			
		case "typed_default_parameter":
			// parameter: type = default
			foundType := false
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" && param.Name == "" {
					param.Name = p.getNodeText(subChild)
				} else if subChild.Type() == "type" && !foundType {
					param.Type = ir.TypeRef{
						Name: p.getNodeText(subChild),
					}
					foundType = true
				} else if subChild.IsNamed() && foundType {
					param.DefaultValue = p.getNodeText(subChild)
				}
			}
			
		case "list_splat_parameter":
			// *args
			if child.ChildCount() > 0 {
				nameNode := child.Child(0)
				if nameNode != nil && nameNode.Type() == "identifier" {
					param.Name = "*" + p.getNodeText(nameNode)
				}
			}
			
		case "dictionary_splat_parameter":
			// **kwargs
			if child.ChildCount() > 0 {
				nameNode := child.Child(0)
				if nameNode != nil && nameNode.Type() == "identifier" {
					param.Name = "**" + p.getNodeText(nameNode)
				}
			}
		}
		
		if param.Name != "" {
			fn.Parameters = append(fn.Parameters, *param)
		}
	}
}

// processAssignment processes variable assignments
func (p *NativeTreeSitterProcessor) processAssignment(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	var name string
	var typeRef *ir.TypeRef
	var value string
	
	// Find left and right sides
	var leftNode, rightNode *sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "=" {
			if i > 0 {
				leftNode = node.Child(i - 1)
			}
			if i < int(node.ChildCount())-1 {
				rightNode = node.Child(i + 1)
			}
			break
		}
	}
	
	if leftNode == nil {
		// Try pattern: first child is left, last is right
		if node.ChildCount() >= 3 {
			leftNode = node.Child(0)
			rightNode = node.Child(int(node.ChildCount()) - 1)
		}
	}
	
	// Extract name from left side
	if leftNode != nil {
		if leftNode.Type() == "identifier" {
			name = p.getNodeText(leftNode)
		} else if leftNode.Type() == "typed_assignment" {
			// variable: Type = value
			for i := 0; i < int(leftNode.ChildCount()); i++ {
				child := leftNode.Child(i)
				if child.Type() == "identifier" && name == "" {
					name = p.getNodeText(child)
				} else if child.Type() == "type" {
					typeRef = &ir.TypeRef{
						Name: p.getNodeText(child),
					}
				}
			}
		}
	}
	
	// Extract value from right side
	if rightNode != nil {
		value = p.getNodeText(rightNode)
	}
	
	if name != "" {
		field := &ir.DistilledField{
			BaseNode: p.nodeLocation(node),
			Name:     name,
			Type:     typeRef,
			DefaultValue: value,
		}
		
		// Set visibility
		if strings.HasPrefix(name, "_") {
			field.Visibility = ir.VisibilityPrivate
		} else {
			field.Visibility = ir.VisibilityPublic
		}
		
		p.addNode(file, parent, field)
	}
}

// processDocstring processes docstring comments
func (p *NativeTreeSitterProcessor) processDocstring(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	text := p.getNodeText(node)
	
	// Remove quotes
	text = strings.Trim(text, `"'`)
	if strings.HasPrefix(text, `"""`) || strings.HasPrefix(text, `'''`) {
		text = text[3 : len(text)-3]
	}
	
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     strings.TrimSpace(text),
		Format:   "docstring",
	}
	
	p.addNode(file, parent, comment)
}

// processComment processes regular comments
func (p *NativeTreeSitterProcessor) processComment(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	text := p.getNodeText(node)
	
	// Remove # and trim
	text = strings.TrimPrefix(text, "#")
	text = strings.TrimSpace(text)
	
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     text,
		Format:   "line",
	}
	
	p.addNode(file, parent, comment)
}

// Helper methods

func (p *NativeTreeSitterProcessor) nodeLocation(node *sitter.Node) ir.BaseNode {
	startPoint := node.StartPoint()
	endPoint := node.EndPoint()
	
	return ir.BaseNode{
		Location: ir.Location{
			StartLine:   int(startPoint.Row) + 1, // tree-sitter uses 0-based lines
			StartColumn: int(startPoint.Column),
			EndLine:     int(endPoint.Row) + 1,
			EndColumn:   int(endPoint.Column),
			StartByte:   int(node.StartByte()),
			EndByte:     int(node.EndByte()),
		},
	}
}

func (p *NativeTreeSitterProcessor) getNodeText(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if int(start) < len(p.source) && int(end) <= len(p.source) {
		return string(p.source[start:end])
	}
	return ""
}

func (p *NativeTreeSitterProcessor) addNode(file *ir.DistilledFile, parent ir.DistilledNode, node ir.DistilledNode) {
	if parent != nil {
		switch p := parent.(type) {
		case *ir.DistilledClass:
			p.Children = append(p.Children, node)
		case *ir.DistilledFunction:
			// Functions don't typically have children in Python
		}
	} else {
		file.Children = append(file.Children, node)
	}
}

func (p *NativeTreeSitterProcessor) extractDocstring(bodyNode *sitter.Node) string {
	// Check if first statement is a string (docstring)
	if bodyNode.ChildCount() > 0 {
		firstStmt := bodyNode.Child(0)
		if firstStmt != nil && firstStmt.Type() == "expression_statement" {
			if firstStmt.ChildCount() > 0 {
				expr := firstStmt.Child(0)
				if expr != nil && expr.Type() == "string" {
					docstring := p.getNodeText(expr)
					// Clean up docstring
					docstring = strings.Trim(docstring, `"'`)
					if strings.HasPrefix(docstring, `"""`) || strings.HasPrefix(docstring, `'''`) {
						docstring = docstring[3 : len(docstring)-3]
					}
					return strings.TrimSpace(docstring)
				}
			}
		}
	}
	return ""
}