package python

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	"github.com/janreges/ai-distiller/internal/debug"
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
	dbg := debug.FromContext(ctx).WithSubsystem("python:tree-sitter")
	defer dbg.Timing(debug.LevelDetailed, "tree-sitter parsing")()
	
	p.source = source
	p.filename = filename
	
	dbg.Logf(debug.LevelDetailed, "Parsing %d bytes with tree-sitter", len(source))
	
	// Parse with tree-sitter
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Close()
	
	// Dump raw AST at trace level
	debug.Lazy(ctx, debug.LevelTrace, func(d debug.Debugger) {
		d.Dump(debug.LevelTrace, "Raw tree-sitter AST", p.dumpTree(tree.RootNode(), 0))
	})
	
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
	
	dbg.Logf(debug.LevelDetailed, "Processed %d top-level nodes", len(file.Children))
	
	// Analyze protocol satisfaction
	p.analyzeProtocolSatisfaction(file)
	
	// Dump final IR structure at trace level
	debug.Lazy(ctx, debug.LevelTrace, func(d debug.Debugger) {
		d.Dump(debug.LevelTrace, "Final Python IR structure", file)
	})
	
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
		
	case "assignment":
		p.processAssignment(node, file, parent)
		
	case "expression_statement":
		// Check for docstrings and assignments
		if node.ChildCount() > 0 {
			child := node.Child(0)
			if child.Type() == "string" {
				// Only process top-level docstrings (module-level), not function docstrings
				if parent == nil { // parent == nil means it's at module level
					p.processDocstring(child, file, parent)
				}
			} else if child.Type() == "assignment" {
				p.processAssignment(child, file, parent)
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
	// Simple approach - just get the text and parse it ourselves
	nodeText := p.getNodeText(node)
	
	imp := &ir.DistilledImport{
		BaseNode: p.nodeLocation(node),
		ImportType: "from",
		Symbols: []ir.ImportedSymbol{},
	}
	
	// Parse "from MODULE import SYMBOL[, SYMBOL2] [as ALIAS]"
	if strings.HasPrefix(nodeText, "from ") && strings.Contains(nodeText, " import ") {
		parts := strings.Split(nodeText, " import ")
		if len(parts) == 2 {
			// Extract module name
			modulePart := strings.TrimPrefix(parts[0], "from ")
			imp.Module = strings.TrimSpace(modulePart)
			
			// Extract symbols
			symbolsPart := strings.TrimSpace(parts[1])
			if symbolsPart == "*" {
				// Star import - leave symbols empty
			} else {
				// Parse individual symbols
				symbols := strings.Split(symbolsPart, ",")
				for _, sym := range symbols {
					sym = strings.TrimSpace(sym)
					if sym == "" {
						continue
					}
					
					// Handle "symbol as alias"
					if strings.Contains(sym, " as ") {
						aliasParts := strings.Split(sym, " as ")
						if len(aliasParts) == 2 {
							imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
								Name:  strings.TrimSpace(aliasParts[0]),
								Alias: strings.TrimSpace(aliasParts[1]),
							})
						}
					} else {
						imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
							Name: sym,
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
	if p.isPrivateName(class.Name) {
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
			// Get implementation
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	// Set visibility
	if p.isPrivateName(fn.Name) {
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
		if p.isPrivateName(name) {
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
	
	// Remove quotes - handle triple quotes first
	if strings.HasPrefix(text, `"""`) && strings.HasSuffix(text, `"""`) && len(text) >= 6 {
		text = text[3 : len(text)-3]
	} else if strings.HasPrefix(text, `'''`) && strings.HasSuffix(text, `'''`) && len(text) >= 6 {
		text = text[3 : len(text)-3]
	} else {
		// Handle single quotes
		text = strings.Trim(text, `"'`)
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

// isPrivateName checks if a name should be considered private in Python
func (p *NativeTreeSitterProcessor) isPrivateName(name string) bool {
	// In Python, names starting with _ are considered private
	// BUT dunder methods (like __init__, __repr__) are public API
	if len(name) == 0 {
		return false
	}
	
	// Dunder methods are public
	if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
		return false
	}
	
	// Single underscore prefix means private
	return name[0] == '_'
}

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

// analyzeProtocolSatisfaction analyzes protocol implementations using duck typing
func (p *NativeTreeSitterProcessor) analyzeProtocolSatisfaction(file *ir.DistilledFile) {
	// Collect all protocols and classes
	protocols := make(map[string]*ir.DistilledClass)
	classes := make(map[string]*ir.DistilledClass)
	
	for _, child := range file.Children {
		if class, ok := child.(*ir.DistilledClass); ok {
			// Check if it's a Protocol by looking for Protocol in inheritance
			isProtocol := false
			for _, ext := range class.Extends {
				if ext.Name == "Protocol" {
					isProtocol = true
					break
				}
			}
			
			if isProtocol {
				protocols[class.Name] = class
			} else {
				classes[class.Name] = class
			}
		}
	}
	
	// For each class, check if it satisfies any protocols (duck typing)
	for _, class := range classes {
		for protocolName, protocol := range protocols {
			if p.classSatisfiesProtocol(class, protocol) {
				// Add implicit protocol satisfaction
				class.Implements = append(class.Implements, ir.TypeRef{
					Name: protocolName + " (duck typing)",
				})
			}
		}
	}
}

// classSatisfiesProtocol checks if a class satisfies a protocol via duck typing
func (p *NativeTreeSitterProcessor) classSatisfiesProtocol(class *ir.DistilledClass, protocol *ir.DistilledClass) bool {
	// Create method map for the class
	classMethods := make(map[string]*ir.DistilledFunction)
	
	for _, child := range class.Children {
		if method, ok := child.(*ir.DistilledFunction); ok {
			// Only consider public methods for protocol satisfaction
			if method.Visibility == ir.VisibilityPublic {
				classMethods[method.Name] = method
			}
		}
	}
	
	// Check if class has all protocol methods
	for _, child := range protocol.Children {
		if protocolMethod, ok := child.(*ir.DistilledFunction); ok {
			classMethod, exists := classMethods[protocolMethod.Name]
			if !exists {
				return false
			}
			
			// Check method signature compatibility (simplified)
			if !p.methodSignaturesCompatible(classMethod, protocolMethod) {
				return false
			}
		}
	}
	
	return true
}

// methodSignaturesCompatible checks if two method signatures are compatible
func (p *NativeTreeSitterProcessor) methodSignaturesCompatible(classMethod, protocolMethod *ir.DistilledFunction) bool {
	// Check parameter count (excluding 'self')
	classParams := p.getNonSelfParameters(classMethod.Parameters)
	protocolParams := p.getNonSelfParameters(protocolMethod.Parameters)
	
	if len(classParams) != len(protocolParams) {
		return false
	}
	
	// Check parameter types (if specified)
	for i, classParam := range classParams {
		protocolParam := protocolParams[i]
		
		// If protocol specifies type, class must match (simplified check)
		if protocolParam.Type.Name != "" && classParam.Type.Name != "" {
			if !p.typesCompatible(classParam.Type, protocolParam.Type) {
				return false
			}
		}
	}
	
	// Check return type compatibility
	if protocolMethod.Returns != nil && classMethod.Returns != nil {
		return p.typesCompatible(*classMethod.Returns, *protocolMethod.Returns)
	}
	
	return true
}

// getNonSelfParameters filters out 'self' parameter
func (p *NativeTreeSitterProcessor) getNonSelfParameters(params []ir.Parameter) []ir.Parameter {
	var result []ir.Parameter
	for _, param := range params {
		if param.Name != "self" && param.Name != "cls" {
			result = append(result, param)
		}
	}
	return result
}

// typesCompatible checks if two types are compatible (simplified)
func (p *NativeTreeSitterProcessor) typesCompatible(type1, type2 ir.TypeRef) bool {
	// Simple string comparison for now
	return strings.TrimSpace(type1.Name) == strings.TrimSpace(type2.Name)
}

// dumpTree creates a structured representation of the AST for debugging
func (p *NativeTreeSitterProcessor) dumpTree(node *sitter.Node, depth int) map[string]interface{} {
	if node == nil {
		return nil
	}
	
	result := map[string]interface{}{
		"type":      node.Type(),
		"named":     node.IsNamed(),
		"startLine": node.StartPoint().Row + 1,
		"endLine":   node.EndPoint().Row + 1,
		"startCol":  node.StartPoint().Column,
		"endCol":    node.EndPoint().Column,
	}
	
	// Add node text for leaf nodes or small nodes
	if node.ChildCount() == 0 || (node.EndByte()-node.StartByte() < 100) {
		text := p.getNodeText(node)
		if text != "" && len(text) < 100 {
			result["text"] = text
		}
	}
	
	// Add children
	if node.ChildCount() > 0 {
		children := make([]map[string]interface{}, 0, node.ChildCount())
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				children = append(children, p.dumpTree(child, depth+1))
			}
		}
		result["children"] = children
	}
	
	return result
}