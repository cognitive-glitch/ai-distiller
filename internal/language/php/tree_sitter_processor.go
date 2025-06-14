package php

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TreeSitterProcessor processes PHP using tree-sitter
type TreeSitterProcessor struct {
	parser   *sitter.Parser
	source   []byte
	filename string
	
	// State for context-aware parsing
	currentNamespace string
	useAliases      map[string]string
	currentClass    string
	
	// PHPDoc information by line number
	docblocks map[int]*PHPDocInfo
}

// PHPDocInfo stores parsed PHPDoc information
type PHPDocInfo struct {
	propertyType string
	returnType   string
	paramTypes   map[string]string
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() (*TreeSitterProcessor, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_php.LanguagePHP()))
	
	return &TreeSitterProcessor{
		parser:     parser,
		useAliases: make(map[string]string),
		docblocks:  make(map[int]*PHPDocInfo),
	}, nil
}

// ProcessSource processes PHP source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	
	// Reset state
	p.currentNamespace = ""
	p.useAliases = make(map[string]string)
	p.currentClass = ""
	p.docblocks = make(map[int]*PHPDocInfo)
	
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
		Language: "php",
		Version:  "8",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	// First pass: collect all PHPDoc comments
	p.collectPHPDocComments(tree.RootNode())
	
	// Second pass: process nodes with PHPDoc info available
	p.processNode(tree.RootNode(), file, nil)
	
	return file, nil
}

// collectPHPDocComments collects all PHPDoc comments in first pass
func (p *TreeSitterProcessor) collectPHPDocComments(node *sitter.Node) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	
	// Check if this is a PHPDoc comment
	if nodeType == "comment" {
		text := p.getNodeText(node)
		if strings.HasPrefix(text, "/**") {
			// Parse PHPDoc
			info := p.parsePHPDoc(text)
			if info != nil {
				// Store by the line of the NEXT non-comment node
				nextLine := p.findNextCodeLine(node)
				if nextLine > 0 {
					p.docblocks[nextLine] = info
				}
			}
		}
	}
	
	// Process children
	for i := 0; i < int(node.ChildCount()); i++ {
		p.collectPHPDocComments(node.Child(i))
	}
}

// findNextCodeLine finds the line number of the next code element after a comment
func (p *TreeSitterProcessor) findNextCodeLine(commentNode *sitter.Node) int {
	parent := commentNode.Parent()
	if parent == nil {
		return 0
	}
	
	foundComment := false
	for i := 0; i < int(parent.ChildCount()); i++ {
		child := parent.Child(i)
		if child == commentNode {
			foundComment = true
			continue
		}
		
		if foundComment && child.IsNamed() {
			nodeType := child.Type()
			// Skip other comments and PHP tags
			if nodeType != "comment" && nodeType != "php_tag" && nodeType != "text_interpolation" {
				return int(child.StartPoint().Row) + 1
			}
		}
	}
	
	return 0
}

// parsePHPDoc parses PHPDoc comment and extracts type information
func (p *TreeSitterProcessor) parsePHPDoc(text string) *PHPDocInfo {
	info := &PHPDocInfo{
		paramTypes: make(map[string]string),
	}
	
	// Clean up docblock
	text = strings.TrimPrefix(text, "/**")
	text = strings.TrimSuffix(text, "*/")
	
	lines := strings.Split(text, "\n")
	hasTypeInfo := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		
		// Extract @var type for properties
		if strings.HasPrefix(line, "@var ") {
			typeStr := strings.TrimSpace(strings.TrimPrefix(line, "@var"))
			// For @var, take everything as the type (no parameter name)
			if typeStr != "" {
				info.propertyType = p.normalizeArrayType(typeStr)
				hasTypeInfo = true
			}
		}
		
		// Extract @return type for functions
		if strings.HasPrefix(line, "@return ") {
			typeStr := strings.TrimSpace(strings.TrimPrefix(line, "@return"))
			// For @return, take everything as the type (description is usually on next line)
			if typeStr != "" {
				info.returnType = p.normalizeArrayType(typeStr)
				hasTypeInfo = true
			}
		}
		
		// Extract @param types for function parameters
		if strings.HasPrefix(line, "@param ") {
			paramStr := strings.TrimSpace(strings.TrimPrefix(line, "@param"))
			// For @param, we need to find where the parameter name starts
			// Complex types can have spaces, so we look for $paramName pattern
			dollarPos := strings.Index(paramStr, "$")
			if dollarPos > 0 {
				// Extract type (everything before the $)
				typeName := strings.TrimSpace(paramStr[:dollarPos])
				// Extract parameter name
				rest := paramStr[dollarPos+1:]
				spacePos := strings.IndexFunc(rest, unicode.IsSpace)
				paramName := rest
				if spacePos > 0 {
					paramName = rest[:spacePos]
				}
				if typeName != "" && paramName != "" {
					info.paramTypes[paramName] = p.normalizeArrayType(typeName)
					hasTypeInfo = true
				}
			}
		}
	}
	
	if hasTypeInfo {
		return info
	}
	return nil
}

// normalizeArrayType normalizes array type notation from PHPDoc
// Converts array<User> to User[], keeps User[] as is
// For associative arrays like array<string, User>, keeps the full notation
func (p *TreeSitterProcessor) normalizeArrayType(typeName string) string {
	// Handle generic array notation: array<Type> or list<Type>
	if (strings.HasPrefix(typeName, "array<") || strings.HasPrefix(typeName, "list<")) && strings.HasSuffix(typeName, ">") {
		start := strings.Index(typeName, "<")
		end := strings.LastIndex(typeName, ">")
		if start != -1 && end > start {
			innerType := typeName[start+1:end]
			
			// Check if it's associative array (has comma)
			if strings.Contains(innerType, ",") {
				// For associative arrays, keep the full notation
				// This preserves array<string, User> and array<string, User[]>
				return typeName
			}
			
			// Simple array, convert to []
			return innerType + "[]"
		}
	}
	
	// Handle collection types like Collection<User>
	if strings.Contains(typeName, "<") && strings.HasSuffix(typeName, ">") {
		// For other generic types, keep the original notation
		// This includes Collection<User>, ArrayObject<User>, etc.
		return typeName
	}
	
	return typeName
}

// processNode recursively processes tree-sitter nodes
func (p *TreeSitterProcessor) processNode(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	// Debug output
	// fmt.Printf("Node: %s, Text: %s\n", nodeType, p.getNodeText(node))
	
	// Skip non-named nodes except comments
	if !node.IsNamed() && nodeType != "comment" && nodeType != "php_tag" {
		// Process children for anonymous nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		return
	}
	
	switch nodeType {
	case "program", "php":
		// Process all children at top level
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		
	case "namespace_definition":
		p.processNamespace(node, file, parent)
		
	case "namespace_use_declaration":
		p.processUseDeclaration(node, file, parent)
		
	case "class_declaration":
		p.processClass(node, file, parent)
		
	case "interface_declaration":
		p.processInterface(node, file, parent)
		
	case "trait_declaration":
		p.processTrait(node, file, parent)
		
	case "enum_declaration":
		p.processEnum(node, file, parent)
		
	case "function_definition":
		p.processFunction(node, file, parent, false)
		
	case "method_declaration":
		p.processMethod(node, file, parent)
		
	case "property_declaration":
		p.processProperty(node, file, parent)
		
	case "const_declaration":
		p.processConstant(node, file, parent)
		
	case "comment":
		p.processComment(node, file, parent)
		
	default:
		// Process children for other node types
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
	}
}

// processNamespace processes namespace declarations
func (p *TreeSitterProcessor) processNamespace(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Get namespace name
	var namespaceName string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "namespace_name" {
			namespaceName = p.getNodeText(child)
			break
		}
	}
	
	// Update current namespace
	p.currentNamespace = namespaceName
	
	// Process namespace body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "compound_statement" || child.Type() == "declaration_list" {
			// Process declarations within namespace
			for j := 0; j < int(child.ChildCount()); j++ {
				p.processNode(child.Child(j), file, parent)
			}
		}
	}
}

// processUseDeclaration processes use statements
func (p *TreeSitterProcessor) processUseDeclaration(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// PHP use statements can import classes, functions, or constants
	var importType string = "class" // default
	
	// Check for function or const use
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "function" {
			importType = "function"
		} else if child.Type() == "const" {
			importType = "const"
		}
	}
	
	// Check if this is a grouped use statement
	var groupPrefix string
	var hasGroup bool
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "namespace_name" || child.Type() == "qualified_name" {
			groupPrefix = p.getNodeText(child)
		}
		if child.Type() == "namespace_use_group" {
			hasGroup = true
			// Process the group with the prefix
			p.processUseGroup(child, file, parent, importType, groupPrefix)
		}
	}
	
	// If not a grouped statement, process normally
	if !hasGroup {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "namespace_use_clause" {
				p.processUseClause(child, file, parent, importType)
			}
		}
	}
}

// processUseClause processes individual use clauses
func (p *TreeSitterProcessor) processUseClause(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, importType string) {
	var fullName string
	var alias string
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "qualified_name", "namespace_name":
			fullName = p.getNodeText(child)
		case "namespace_aliasing_clause":
			// Get the alias
			for j := 0; j < int(child.ChildCount()); j++ {
				if child.Child(j).Type() == "name" {
					alias = p.getNodeText(child.Child(j))
					break
				}
			}
		}
	}
	
	if fullName != "" {
		// Store alias mapping
		if alias != "" {
			p.useAliases[alias] = fullName
		} else {
			// Use the last part of the name as the alias
			parts := strings.Split(fullName, "\\")
			if len(parts) > 0 {
				p.useAliases[parts[len(parts)-1]] = fullName
			}
		}
		
		// Create import node
		imp := &ir.DistilledImport{
			BaseNode:   p.nodeLocation(node),
			ImportType: "use",
			Module:     fullName,
			Symbols: []ir.ImportedSymbol{
				{
					Name:  fullName,
					Alias: alias,
				},
			},
		}
		
		p.addNode(file, parent, imp)
	}
}

// processUseGroup processes grouped use statements
func (p *TreeSitterProcessor) processUseGroup(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, importType string, prefix string) {
	
	// Process each item in the group
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "namespace_use_clause" {
			// Handle each individual import in the group
			var itemName string
			var alias string
			
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				switch grandchild.Type() {
				case "name", "qualified_name":
					itemName = p.getNodeText(grandchild)
				case "namespace_aliasing_clause":
					// Get the alias
					for k := 0; k < int(grandchild.ChildCount()); k++ {
						if grandchild.Child(k).Type() == "name" {
							alias = p.getNodeText(grandchild.Child(k))
							break
						}
					}
				}
			}
			
			if itemName != "" {
				fullName := prefix + "\\" + itemName
				
				// Store alias mapping
				if alias != "" {
					p.useAliases[alias] = fullName
				} else {
					// Use the last part of the name as the alias
					parts := strings.Split(itemName, "\\")
					if len(parts) > 0 {
						p.useAliases[parts[len(parts)-1]] = fullName
					}
				}
				
				// Create import node
				imp := &ir.DistilledImport{
					BaseNode:   p.nodeLocation(child),
					ImportType: "use",
					Module:     fullName,
					Symbols: []ir.ImportedSymbol{
						{
							Name:  fullName,
							Alias: alias,
						},
					},
				}
				
				p.addNode(file, parent, imp)
			}
		}
	}
}

// processClass processes class declarations
func (p *TreeSitterProcessor) processClass(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode:   p.nodeLocation(node),
		Modifiers:  []ir.Modifier{},
		Extends:    []ir.TypeRef{},
		Implements: []ir.TypeRef{},
		Children:   []ir.DistilledNode{},
		Decorators: []string{},
	}
	
	// Track current class
	prevClass := p.currentClass
	
	// First pass: get class name for better debugging
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "name" {
			class.Name = p.getNodeText(child)
			p.currentClass = class.Name
			break
		}
	}
	
	// Process class parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		// Debug
		// fmt.Printf("Class %s child %d: %s = %s\n", class.Name, i, child.Type(), p.getNodeText(child))
		switch child.Type() {
		case "visibility_modifier":
			// Handle visibility
			visibility := p.getNodeText(child)
			if visibility == "private" {
				class.Visibility = ir.VisibilityPrivate
			} else if visibility == "protected" {
				class.Visibility = ir.VisibilityProtected
			} else {
				class.Visibility = ir.VisibilityPublic
			}
			
		case "abstract_modifier":
			class.Modifiers = append(class.Modifiers, ir.ModifierAbstract)
			
		case "final_modifier":
			class.Modifiers = append(class.Modifiers, ir.ModifierFinal)
			
		case "readonly_modifier":
			class.Modifiers = append(class.Modifiers, ir.ModifierReadonly)
			
		case "name":
			// Already set in first pass
			
		case "base_clause":
			// Extends clause
			for j := 0; j < int(child.ChildCount()); j++ {
				if child.Child(j).Type() == "qualified_name" || child.Child(j).Type() == "name" {
					// Don't resolve the full name, just use the short name or alias
					// The imports are already tracked separately
					class.Extends = append(class.Extends, ir.TypeRef{
						Name: p.getNodeText(child.Child(j)),
					})
				}
			}
			
		case "class_interface_clause":
			// Implements clause
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "qualified_name" || grandchild.Type() == "name" {
					// Don't resolve the full name, just use the short name or alias
					// The imports are already tracked separately
					class.Implements = append(class.Implements, ir.TypeRef{
						Name: p.getNodeText(grandchild),
					})
				}
			}
			
		case "attribute_list":
			// PHP 8 attributes
			p.processAttributes(child, &class.Decorators)
			
		case "declaration_list":
			// Class body
			p.processClassBody(child, file, class)
		}
	}
	
	// If no explicit visibility, default to public
	if class.Visibility == "" {
		class.Visibility = ir.VisibilityPublic
	}
	
	p.addNode(file, parent, class)
	
	// Restore previous class context
	p.currentClass = prevClass
}

// processClassBody processes the body of a class
func (p *TreeSitterProcessor) processClassBody(node *sitter.Node, file *ir.DistilledFile, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "method_declaration":
			p.processMethod(child, file, class)
			
		case "property_declaration":
			p.processProperty(child, file, class)
			
		case "const_declaration":
			p.processConstant(child, file, class)
			
		case "use_declaration":
			// Trait use
			p.processTraitUse(child, file, class)
		}
	}
}

// processMethod processes method declarations
func (p *TreeSitterProcessor) processMethod(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	fn := &ir.DistilledFunction{
		BaseNode:   p.nodeLocation(node),
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
		Decorators: []string{},
	}
	
	var hasVisibility bool
	
	// Process method parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		// Debug: print all child types
		//fmt.Printf("Class child %d: %s = %s\n", i, child.Type(), p.getNodeText(child))
		switch child.Type() {
		case "visibility_modifier":
			visibility := p.getNodeText(child)
			if visibility == "private" {
				fn.Visibility = ir.VisibilityPrivate
			} else if visibility == "protected" {
				fn.Visibility = ir.VisibilityProtected
			} else {
				fn.Visibility = ir.VisibilityPublic
			}
			hasVisibility = true
			
		case "static_modifier":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierStatic)
			
		case "abstract_modifier":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAbstract)
			
		case "final_modifier":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierFinal)
			
		case "attribute_list":
			p.processAttributes(child, &fn.Decorators)
			
		case "name":
			fn.Name = p.getNodeText(child)
			
		case "formal_parameters":
			p.processParameters(child, fn)
			// For constructor, handle property promotion
			if fn.Name == "__construct" && parent != nil {
				if class, ok := parent.(*ir.DistilledClass); ok {
					p.processPromotedProperties(child, class)
				}
			}
			
		case ":":
			// Return type follows the colon
			if i+1 < int(node.ChildCount()) {
				nextChild := node.Child(i + 1)
				if nextChild.Type() == "primitive_type" || nextChild.Type() == "named_type" ||
				   nextChild.Type() == "union_type" || nextChild.Type() == "intersection_type" ||
				   nextChild.Type() == "optional_type" {
					fn.Returns = &ir.TypeRef{
						Name: p.getNodeText(nextChild),
					}
				}
			}
			
		case "compound_statement":
			// Method body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	// Default visibility for methods is public
	if !hasVisibility {
		fn.Visibility = ir.VisibilityPublic
	}
	
	// Check for PHPDoc type information
	nodeLine := int(node.StartPoint().Row) + 1
	if docInfo, exists := p.docblocks[nodeLine]; exists {
		// Apply return type
		if docInfo.returnType != "" {
			if fn.Returns == nil || fn.Returns.Name == "" {
				fn.Returns = &ir.TypeRef{Name: docInfo.returnType}
			} else if fn.Returns.Name == "array" {
				// PHPDoc has more specific array type
				fn.Returns = &ir.TypeRef{Name: docInfo.returnType}
			}
		}
		
		// Apply param types
		for i := range fn.Parameters {
			if docType, exists := docInfo.paramTypes[fn.Parameters[i].Name]; exists {
				if fn.Parameters[i].Type.Name == "" {
					fn.Parameters[i].Type = ir.TypeRef{Name: docType}
				} else if fn.Parameters[i].Type.Name == "array" {
					// PHPDoc has more specific array type
					fn.Parameters[i].Type = ir.TypeRef{Name: docType}
				}
			}
		}
	}
	
	p.addNode(file, parent, fn)
}

// processProperty processes property declarations
func (p *TreeSitterProcessor) processProperty(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	var visibility ir.Visibility = ir.VisibilityPublic
	var modifiers []ir.Modifier
	var propertyType *ir.TypeRef
	var attributes []string
	
	// Process property modifiers and type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "visibility_modifier":
			vis := p.getNodeText(child)
			if vis == "private" {
				visibility = ir.VisibilityPrivate
			} else if vis == "protected" {
				visibility = ir.VisibilityProtected
			}
			
		case "static_modifier":
			modifiers = append(modifiers, ir.ModifierStatic)
			
		case "readonly_modifier":
			modifiers = append(modifiers, ir.ModifierReadonly)
			
		case "type", "union_type", "intersection_type", "optional_type", "primitive_type", "named_type":
			propertyType = &ir.TypeRef{
				Name: p.getNodeText(child),
			}
			
		case "attribute_list":
			p.processAttributes(child, &attributes)
			
		case "property_element":
			// Individual property with possible initializer
			p.processPropertyElement(child, file, parent, visibility, modifiers, propertyType, attributes)
		}
	}
}

// processPropertyElement processes individual property elements
func (p *TreeSitterProcessor) processPropertyElement(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, 
	visibility ir.Visibility, modifiers []ir.Modifier, propertyType *ir.TypeRef, attributes []string) {
	
	field := &ir.DistilledField{
		BaseNode:   p.nodeLocation(node),
		Visibility: visibility,
		Modifiers:  modifiers,
		Type:       propertyType,
	}
	
	// Get property name and default value
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "variable_name":
			field.Name = strings.TrimPrefix(p.getNodeText(child), "$")
			
		case "property_initializer":
			// Get default value
			for j := 0; j < int(child.ChildCount()); j++ {
				if child.Child(j).Type() != "=" {
					field.DefaultValue = p.getNodeText(child.Child(j))
					break
				}
			}
		}
	}
	
	// Check for PHPDoc type information
	nodeLine := int(node.StartPoint().Row) + 1
	if docInfo, exists := p.docblocks[nodeLine]; exists && docInfo.propertyType != "" {
		if field.Type == nil || field.Type.Name == "" {
			field.Type = &ir.TypeRef{Name: docInfo.propertyType}
		} else if field.Type.Name == "array" {
			// PHPDoc has more specific array type
			field.Type = &ir.TypeRef{Name: docInfo.propertyType}
		}
	}
	
	p.addNode(file, parent, field)
}

// processFunction processes function declarations
func (p *TreeSitterProcessor) processFunction(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, isAsync bool) {
	fn := &ir.DistilledFunction{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Functions are always public
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
		Decorators: []string{},
	}
	
	if isAsync {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
	}
	
	// Process function parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "attribute_list":
			p.processAttributes(child, &fn.Decorators)
			
		case "name":
			fn.Name = p.getNodeText(child)
			
		case "formal_parameters":
			p.processParameters(child, fn)
			
		case ":":
			// Return type follows the colon
			if i+1 < int(node.ChildCount()) {
				nextChild := node.Child(i + 1)
				if nextChild.Type() == "primitive_type" || nextChild.Type() == "named_type" ||
				   nextChild.Type() == "union_type" || nextChild.Type() == "intersection_type" ||
				   nextChild.Type() == "optional_type" {
					fn.Returns = &ir.TypeRef{
						Name: p.getNodeText(nextChild),
					}
				}
			}
			
		case "compound_statement":
			// Function body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	// Check for PHPDoc type information
	nodeLine := int(node.StartPoint().Row) + 1
	if docInfo, exists := p.docblocks[nodeLine]; exists {
		// Apply return type
		if docInfo.returnType != "" {
			if fn.Returns == nil || fn.Returns.Name == "" {
				fn.Returns = &ir.TypeRef{Name: docInfo.returnType}
			} else if fn.Returns.Name == "array" {
				// PHPDoc has more specific array type
				fn.Returns = &ir.TypeRef{Name: docInfo.returnType}
			}
		}
		
		// Apply param types
		for i := range fn.Parameters {
			if docType, exists := docInfo.paramTypes[fn.Parameters[i].Name]; exists {
				if fn.Parameters[i].Type.Name == "" {
					fn.Parameters[i].Type = ir.TypeRef{Name: docType}
				} else if fn.Parameters[i].Type.Name == "array" {
					// PHPDoc has more specific array type
					fn.Parameters[i].Type = ir.TypeRef{Name: docType}
				}
			}
		}
	}
	
	p.addNode(file, parent, fn)
}

// processParameters processes function/method parameters
func (p *TreeSitterProcessor) processParameters(node *sitter.Node, fn *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "simple_parameter" || child.Type() == "variadic_parameter" || 
		   child.Type() == "property_promotion_parameter" {
			param := p.processParameter(child)
			if param.Name != "" {
				fn.Parameters = append(fn.Parameters, param)
			}
		}
	}
}

// processParameter processes a single parameter
func (p *TreeSitterProcessor) processParameter(node *sitter.Node) ir.Parameter {
	param := ir.Parameter{}
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "attribute_list":
			// PHP 8 parameter attributes
			
		case "visibility_modifier":
			// Constructor property promotion
			
		case "readonly_modifier":
			// Constructor property promotion with readonly
			
		case "type", "union_type", "intersection_type", "optional_type", "primitive_type", "named_type":
			param.Type = ir.TypeRef{
				Name: p.getNodeText(child),
			}
			
		case "reference_modifier":
			// &$param
			
		case "variadic_unpacking":
			// ...$params
			
		case "variable_name":
			param.Name = strings.TrimPrefix(p.getNodeText(child), "$")
			
		case "default_value":
			// Get default value
			param.DefaultValue = p.getNodeText(child)
		}
	}
	
	return param
}

// processPromotedProperties processes constructor promoted properties
func (p *TreeSitterProcessor) processPromotedProperties(parametersNode *sitter.Node, class *ir.DistilledClass) {
	for i := 0; i < int(parametersNode.ChildCount()); i++ {
		child := parametersNode.Child(i)
		if child.Type() == "property_promotion_parameter" {
			// This is a promoted property
			field := &ir.DistilledField{
				BaseNode:   p.nodeLocation(child),
				Visibility: ir.VisibilityPublic, // Default
				Modifiers:  []ir.Modifier{},
			}
			
			// Process promotion parameter parts
			for j := 0; j < int(child.ChildCount()); j++ {
				paramChild := child.Child(j)
				switch paramChild.Type() {
				case "visibility_modifier":
					visibility := p.getNodeText(paramChild)
					if visibility == "private" {
						field.Visibility = ir.VisibilityPrivate
					} else if visibility == "protected" {
						field.Visibility = ir.VisibilityProtected
					} else {
						field.Visibility = ir.VisibilityPublic
					}
					
				case "readonly_modifier":
					field.Modifiers = append(field.Modifiers, ir.ModifierReadonly)
					
				case "type", "union_type", "intersection_type", "optional_type", "primitive_type", "named_type":
					field.Type = &ir.TypeRef{
						Name: p.getNodeText(paramChild),
					}
					
				case "variable_name":
					field.Name = strings.TrimPrefix(p.getNodeText(paramChild), "$")
					
				case "default_value":
					field.DefaultValue = p.getNodeText(paramChild)
				}
			}
			
			// Add the field to the class
			class.Children = append(class.Children, field)
		}
	}
}

// processInterface processes interface declarations
func (p *TreeSitterProcessor) processInterface(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Use proper DistilledInterface type
	intf := &ir.DistilledInterface{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic,
		Extends:    []ir.TypeRef{},
		Children:   []ir.DistilledNode{},
	}
	
	// Process interface parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "name":
			intf.Name = p.getNodeText(child)
			
		case "interface_extends_clause", "base_clause":
			// Interfaces can extend multiple interfaces
			// Debug: print child types
			//fmt.Printf("DEBUG: interface_extends_clause has %d children\n", child.ChildCount())
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				//fmt.Printf("DEBUG: child[%d] type=%s text=%s\n", j, grandchild.Type(), p.getNodeText(grandchild))
				if grandchild.Type() == "qualified_name" || grandchild.Type() == "name" {
					// Don't resolve the full name, just use the short name or alias
					// The imports are already tracked separately
					intf.Extends = append(intf.Extends, ir.TypeRef{
						Name: p.getNodeText(grandchild),
					})
				}
			}
			
		case "declaration_list":
			// Interface body - process methods
			for j := 0; j < int(child.ChildCount()); j++ {
				methodNode := child.Child(j)
				if methodNode.Type() == "method_declaration" {
					p.processMethod(methodNode, file, intf)
				}
			}
		}
	}
	
	p.addNode(file, parent, intf)
}

// processTrait processes trait declarations
func (p *TreeSitterProcessor) processTrait(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Create a special comment to indicate this is a trait
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     "PHP Trait",
		Format:   "line",
	}
	p.addNode(file, parent, comment)
	
	// Process trait as a class but with a special indicator
	class := &ir.DistilledClass{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{}, 
		Children:   []ir.DistilledNode{},
		Decorators: []string{},
	}
	
	// Process trait parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "name":
			// Prefix with "trait " to distinguish from classes
			class.Name = "trait " + p.getNodeText(child)
			
		case "declaration_list":
			// Trait body
			p.processClassBody(child, file, class)
		}
	}
	
	p.addNode(file, parent, class)
}

// processEnum processes enum declarations (PHP 8.1+)
func (p *TreeSitterProcessor) processEnum(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// For now, treat enums as classes
	class := &ir.DistilledClass{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{ir.ModifierFinal}, // Enums are implicitly final
		Children:   []ir.DistilledNode{},
		Decorators: []string{},
	}
	
	// Process enum parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "name":
			class.Name = p.getNodeText(child)
			
		case "enum_backing_type":
			// : string or : int
			
		case "enum_declaration_list":
			// Enum cases
			p.processEnumBody(child, file, class)
		}
	}
	
	p.addNode(file, parent, class)
}

// processEnumBody processes enum cases
func (p *TreeSitterProcessor) processEnumBody(node *sitter.Node, file *ir.DistilledFile, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "enum_case" {
			// Process enum case as a constant
			p.processEnumCase(child, file, class)
		}
	}
}

// processEnumCase processes individual enum cases
func (p *TreeSitterProcessor) processEnumCase(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	field := &ir.DistilledField{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierFinal},
	}
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "name":
			field.Name = p.getNodeText(child)
			
		case "enum_case_expression":
			// Case value
			for j := 0; j < int(child.ChildCount()); j++ {
				if child.Child(j).Type() != "=" {
					field.DefaultValue = p.getNodeText(child.Child(j))
					break
				}
			}
		}
	}
	
	p.addNode(file, parent, field)
}

// processConstant processes constant declarations
func (p *TreeSitterProcessor) processConstant(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Process each constant in the declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "const_element" {
			field := &ir.DistilledField{
				BaseNode:   p.nodeLocation(child),
				Visibility: ir.VisibilityPublic,
				Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierFinal},
			}
			
			// Get constant name and value
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				switch grandchild.Type() {
				case "name":
					field.Name = p.getNodeText(grandchild)
				default:
					if grandchild.Type() != "=" {
						field.DefaultValue = p.getNodeText(grandchild)
					}
				}
			}
			
			p.addNode(file, parent, field)
		}
	}
}

// processTraitUse processes trait usage in a class
func (p *TreeSitterProcessor) processTraitUse(node *sitter.Node, file *ir.DistilledFile, class *ir.DistilledClass) {
	// For now, we'll add a comment indicating trait usage
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Format:   "trait_use",
	}
	
	var traits []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "qualified_name" || child.Type() == "name" {
			// Don't resolve the full name, just use the short name or alias
			traits = append(traits, p.getNodeText(child))
		}
	}
	
	if len(traits) > 0 {
		comment.Text = fmt.Sprintf("Uses traits: %s", strings.Join(traits, ", "))
		p.addNode(file, class, comment)
	}
}

// processComment processes comments
func (p *TreeSitterProcessor) processComment(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	text := p.getNodeText(node)
	
	// Determine comment type
	format := "line"
	if strings.HasPrefix(text, "/**") {
		format = "docblock"
		// Clean up docblock
		text = strings.TrimPrefix(text, "/**")
		text = strings.TrimSuffix(text, "*/")
		// Remove leading asterisks from each line
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		}
		text = strings.Join(lines, "\n")
		
		// Extract type information from PHPDoc if this is attached to a node
		if parent != nil {
			p.extractPHPDocTypes(text, parent)
		}
	} else if strings.HasPrefix(text, "/*") {
		format = "block"
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
	} else if strings.HasPrefix(text, "//") {
		text = strings.TrimPrefix(text, "//")
	} else if strings.HasPrefix(text, "#") {
		text = strings.TrimPrefix(text, "#")
	}
	
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     strings.TrimSpace(text),
		Format:   format,
	}
	
	p.addNode(file, parent, comment)
}

// extractPHPDocTypes extracts type information from PHPDoc comments
func (p *TreeSitterProcessor) extractPHPDocTypes(docText string, node ir.DistilledNode) {
	// This method is not used anymore as PHPDoc parsing is handled
	// in the collectPHPDocComments and parsePHPDoc methods
}

// processAttributes processes PHP 8 attributes
func (p *TreeSitterProcessor) processAttributes(node *sitter.Node, decorators *[]string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "attribute_group" {
			// Process each attribute in the group
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "attribute" {
					// Get attribute text
					text := p.getNodeText(grandchild)
					*decorators = append(*decorators, text)
				}
			}
		}
	}
}

// resolveTypeName resolves a type name using current namespace and use statements
func (p *TreeSitterProcessor) resolveTypeName(name string) string {
	// If already fully qualified, return as is
	if strings.HasPrefix(name, "\\") {
		return name
	}
	
	// Check if it's an aliased type
	parts := strings.SplitN(name, "\\", 2)
	if len(parts) > 0 {
		if fullName, ok := p.useAliases[parts[0]]; ok {
			if len(parts) > 1 {
				return fullName + "\\" + parts[1]
			}
			return fullName
		}
	}
	
	// If in a namespace, prepend it
	if p.currentNamespace != "" {
		return p.currentNamespace + "\\" + name
	}
	
	return name
}

// Helper methods

func (p *TreeSitterProcessor) nodeLocation(node *sitter.Node) ir.BaseNode {
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

func (p *TreeSitterProcessor) getNodeText(node *sitter.Node) string {
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

func (p *TreeSitterProcessor) addNode(file *ir.DistilledFile, parent ir.DistilledNode, node ir.DistilledNode) {
	if parent != nil {
		switch p := parent.(type) {
		case *ir.DistilledClass:
			p.Children = append(p.Children, node)
		case *ir.DistilledInterface:
			p.Children = append(p.Children, node)
		case *ir.DistilledFunction:
			// Functions don't typically have children in PHP
		}
	} else {
		file.Children = append(file.Children, node)
	}
}

// Close cleans up resources
func (p *TreeSitterProcessor) Close() error {
	if p.parser != nil {
		p.parser.Close()
	}
	return nil
}