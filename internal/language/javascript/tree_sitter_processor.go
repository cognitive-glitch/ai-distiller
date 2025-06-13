package javascript

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TreeSitterProcessor processes JavaScript using tree-sitter
type TreeSitterProcessor struct {
	parser   *sitter.Parser
	source   []byte
	filename string
	
	// State for context-aware parsing
	currentClass    string
	insideClass     bool
	
	// JSDoc information by line number
	jsdocComments map[int]*JSDocInfo
}

// JSDocInfo stores parsed JSDoc information
type JSDocInfo struct {
	description  string
	paramTypes   map[string]string
	returnType   string
	typedefType  string
	isPrivate    bool
	isProtected  bool
	isPublic     bool
	isStatic     bool
	isAsync      bool
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() (*TreeSitterProcessor, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_javascript.Language()))
	
	return &TreeSitterProcessor{
		parser:        parser,
		jsdocComments: make(map[int]*JSDocInfo),
	}, nil
}

// ProcessSource processes JavaScript source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	
	// Reset state
	p.currentClass = ""
	p.insideClass = false
	p.jsdocComments = make(map[int]*JSDocInfo)
	
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
		Language: "javascript",
		Version:  "ES2022",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	// First pass: collect all JSDoc comments
	p.collectJSDocComments(tree.RootNode())
	
	// Second pass: process nodes with JSDoc info available
	p.processNode(tree.RootNode(), file, nil)
	
	return file, nil
}

// collectJSDocComments collects all JSDoc comments in first pass
func (p *TreeSitterProcessor) collectJSDocComments(node *sitter.Node) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	
	// Check if this is a JSDoc comment
	if nodeType == "comment" {
		text := p.getNodeText(node)
		if strings.HasPrefix(text, "/**") && !strings.HasPrefix(text, "/***") {
			// Parse JSDoc
			info := p.parseJSDoc(text)
			if info != nil {
				// Store by the line of the NEXT non-comment node
				nextLine := p.findNextCodeLine(node)
				if nextLine > 0 {
					p.jsdocComments[nextLine] = info
				}
			}
		}
	}
	
	// Process children
	for i := 0; i < int(node.ChildCount()); i++ {
		p.collectJSDocComments(node.Child(i))
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
			// Skip other comments
			if nodeType != "comment" {
				return int(child.StartPoint().Row) + 1
			}
		}
	}
	
	return 0
}

// parseJSDoc parses JSDoc comment and extracts type information
func (p *TreeSitterProcessor) parseJSDoc(text string) *JSDocInfo {
	info := &JSDocInfo{
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
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Parse tags
		if strings.HasPrefix(line, "@") {
			parts := strings.Fields(line)
			if len(parts) < 1 {
				continue
			}
			
			tag := parts[0]
			switch tag {
			case "@param":
				if len(parts) >= 3 {
					// @param {type} name - description
					typeStr := strings.Trim(parts[1], "{}")
					paramName := strings.TrimPrefix(parts[2], "...")
					info.paramTypes[paramName] = typeStr
					hasTypeInfo = true
				}
			case "@returns", "@return":
				if len(parts) >= 2 {
					// @returns {type} description
					info.returnType = strings.Trim(parts[1], "{}")
					hasTypeInfo = true
				}
			case "@type":
				if len(parts) >= 2 {
					// @type {type}
					info.typedefType = strings.Trim(parts[1], "{}")
					hasTypeInfo = true
				}
			case "@typedef":
				if len(parts) >= 2 {
					// @typedef {type} Name
					info.typedefType = strings.Trim(parts[1], "{}")
					hasTypeInfo = true
				}
			case "@private":
				info.isPrivate = true
				hasTypeInfo = true
			case "@protected":
				info.isProtected = true
				hasTypeInfo = true
			case "@public":
				info.isPublic = true
				hasTypeInfo = true
			case "@static":
				info.isStatic = true
				hasTypeInfo = true
			case "@async":
				info.isAsync = true
				hasTypeInfo = true
			}
		} else if info.description == "" {
			// First non-tag line is description
			info.description = line
		}
	}
	
	if hasTypeInfo || info.description != "" {
		return info
	}
	return nil
}

// processNode recursively processes tree-sitter nodes
func (p *TreeSitterProcessor) processNode(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	
	// Skip non-named nodes except specific ones
	if !node.IsNamed() && nodeType != "comment" {
		// Process children for anonymous nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		return
	}
	
	switch nodeType {
	case "program":
		// Process all children at top level
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		
	case "import_statement":
		p.processImport(node, file, parent)
		
	case "export_statement":
		p.processExport(node, file, parent)
		
	case "class_declaration":
		p.processClass(node, file, parent)
		
	case "function_declaration":
		p.processFunction(node, file, parent, false)
		
	case "generator_function_declaration":
		p.processFunction(node, file, parent, false)
		
	case "variable_declaration":
		p.processVariableDeclaration(node, file, parent)
		
	case "lexical_declaration":
		p.processLexicalDeclaration(node, file, parent)
		
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
func (p *TreeSitterProcessor) processImport(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	imp := &ir.DistilledImport{
		BaseNode:   p.nodeLocation(node),
		ImportType: "import",
		Symbols:    []ir.ImportedSymbol{},
	}
	
	// Process import parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "import_clause":
			p.processImportClause(child, imp)
		case "string":
			// Module path
			imp.Module = strings.Trim(p.getNodeText(child), "\"'`")
		}
	}
	
	p.addNode(file, parent, imp)
}

// processImportClause processes the import clause (what's being imported)
func (p *TreeSitterProcessor) processImportClause(node *sitter.Node, imp *ir.DistilledImport) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			// Default import
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name: p.getNodeText(child),
			})
		case "named_imports":
			// { a, b as c }
			p.processNamedImports(child, imp)
		case "namespace_import":
			// * as name
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "identifier" {
					imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
						Name:  "*",
						Alias: p.getNodeText(grandchild),
					})
				}
			}
		}
	}
}

// processNamedImports processes named imports { a, b as c }
func (p *TreeSitterProcessor) processNamedImports(node *sitter.Node, imp *ir.DistilledImport) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "import_specifier" {
			var name, alias string
			
			// Use field names for more reliable parsing
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				name = p.getNodeText(nameNode)
			}
			
			aliasNode := child.ChildByFieldName("alias")
			if aliasNode != nil {
				alias = p.getNodeText(aliasNode)
			}
			
			// Fallback to positional if fields not available
			if name == "" {
				for j := 0; j < int(child.ChildCount()); j++ {
					grandchild := child.Child(j)
					if grandchild.Type() == "identifier" {
						if name == "" {
							name = p.getNodeText(grandchild)
						} else {
							alias = p.getNodeText(grandchild)
						}
					}
				}
			}
			
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  name,
				Alias: alias,
			})
		}
	}
}

// processExport processes export statements
func (p *TreeSitterProcessor) processExport(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Process the exported declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "class_declaration":
			p.processClass(child, file, parent)
		case "function_declaration":
			p.processFunction(child, file, parent, false)
		case "lexical_declaration":
			p.processLexicalDeclaration(child, file, parent)
		case "export_clause":
			// Handle named exports
			p.processExportClause(child, file, parent)
		}
	}
}

// processExportClause processes export { a, b as c }
func (p *TreeSitterProcessor) processExportClause(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// For now, we'll add these as comments
	// TODO: Add proper export tracking in IR
	var exports []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "export_specifier" {
			exports = append(exports, p.getNodeText(child))
		}
	}
	
	if len(exports) > 0 {
		comment := &ir.DistilledComment{
			BaseNode: p.nodeLocation(node),
			Text:     fmt.Sprintf("Exports: %s", strings.Join(exports, ", ")),
			Format:   "export",
		}
		p.addNode(file, parent, comment)
	}
}

// processClass processes class declarations
func (p *TreeSitterProcessor) processClass(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Classes are public by default in JS
		Modifiers:  []ir.Modifier{},
		Extends:    []ir.TypeRef{},
		Implements: []ir.TypeRef{},
		Children:   []ir.DistilledNode{},
		Decorators: []string{},
	}
	
	// Track current class
	prevClass := p.currentClass
	prevInsideClass := p.insideClass
	p.insideClass = true
	
	// Check for JSDoc
	nodeLine := int(node.StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[nodeLine]; exists {
		if jsdoc.isPrivate {
			class.Visibility = ir.VisibilityPrivate
		} else if jsdoc.isProtected {
			class.Visibility = ir.VisibilityProtected
		}
	}
	
	// Process class parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			class.Name = p.getNodeText(child)
			p.currentClass = class.Name
			
		case "class_heritage":
			// extends clause
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "identifier" || grandchild.Type() == "member_expression" {
					class.Extends = append(class.Extends, ir.TypeRef{
						Name: p.getNodeText(grandchild),
					})
				}
			}
			
		case "class_body":
			// Process class members
			p.processClassBody(child, file, class)
		}
	}
	
	p.addNode(file, parent, class)
	
	// Restore previous class context
	p.currentClass = prevClass
	p.insideClass = prevInsideClass
}

// processClassBody processes the body of a class
func (p *TreeSitterProcessor) processClassBody(node *sitter.Node, file *ir.DistilledFile, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "method_definition":
			p.processMethod(child, file, class)
			
		case "field_definition":
			p.processField(child, file, class)
			
		case "static_block":
			// Static initialization block
			comment := &ir.DistilledComment{
				BaseNode: p.nodeLocation(child),
				Text:     "static initialization block",
				Format:   "block",
			}
			p.addNode(file, class, comment)
		}
	}
}

// processMethod processes method definitions
func (p *TreeSitterProcessor) processMethod(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	fn := &ir.DistilledFunction{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Methods are public by default
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
		Decorators: []string{},
	}
	
	var isGetter, isSetter, isStatic, isAsync, isGenerator bool
	var propertyName string
	
	// Check for JSDoc
	nodeLine := int(node.StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[nodeLine]; exists {
		if jsdoc.isPrivate {
			fn.Visibility = ir.VisibilityPrivate
		} else if jsdoc.isProtected {
			fn.Visibility = ir.VisibilityProtected
		}
		if jsdoc.returnType != "" {
			fn.Returns = &ir.TypeRef{Name: jsdoc.returnType}
		}
		// We'll apply param types when processing parameters
	}
	
	// Process method parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := p.getNodeText(child)
		
		switch child.Type() {
		case "get":
			isGetter = true
		case "set":
			isSetter = true
		case "static":
			isStatic = true
		case "async":
			isAsync = true
		case "*":
			isGenerator = true
			
		case "property_identifier":
			propertyName = text
			
		case "private_property_identifier":
			// Private field/method with # prefix
			propertyName = text
			fn.Visibility = ir.VisibilityPrivate
			
		case "computed_property_name":
			// [expression]
			propertyName = fmt.Sprintf("[%s]", p.getNodeText(child))
			
		case "formal_parameters":
			p.processParameters(child, fn)
			
		case "statement_block":
			// Method body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	// Set method name and modifiers
	if isGetter {
		fn.Name = fmt.Sprintf("get %s", propertyName)
	} else if isSetter {
		fn.Name = fmt.Sprintf("set %s", propertyName)
	} else {
		fn.Name = propertyName
	}
	
	if isStatic {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierStatic)
	}
	if isAsync {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
	}
	if isGenerator {
		fn.Name = "*" + fn.Name
	}
	
	// Special case for constructor
	if fn.Name == "constructor" {
		fn.Name = "constructor"
		// No return type for constructors
		fn.Returns = nil
	}
	
	p.addNode(file, parent, fn)
}

// processField processes field definitions
func (p *TreeSitterProcessor) processField(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	field := &ir.DistilledField{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Fields are public by default
		Modifiers:  []ir.Modifier{},
	}
	
	var isStatic bool
	
	// Check for JSDoc
	nodeLine := int(node.StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[nodeLine]; exists {
		if jsdoc.isPrivate {
			field.Visibility = ir.VisibilityPrivate
		} else if jsdoc.isProtected {
			field.Visibility = ir.VisibilityProtected
		}
		if jsdoc.typedefType != "" {
			field.Type = &ir.TypeRef{Name: jsdoc.typedefType}
		}
	}
	
	// Process field parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "static":
			isStatic = true
			
		case "property_identifier":
			field.Name = p.getNodeText(child)
			
		case "private_property_identifier":
			// Private field with # prefix
			field.Name = p.getNodeText(child)
			field.Visibility = ir.VisibilityPrivate
			
		default:
			// Initializer expression
			if child.Type() != "=" && field.DefaultValue == "" {
				field.DefaultValue = p.getNodeText(child)
			}
		}
	}
	
	if isStatic {
		field.Modifiers = append(field.Modifiers, ir.ModifierStatic)
	}
	
	p.addNode(file, parent, field)
}

// processFunction processes function declarations
func (p *TreeSitterProcessor) processFunction(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, isAsync bool) {
	fn := &ir.DistilledFunction{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Functions are public by default
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
		Decorators: []string{},
	}
	
	var isGenerator bool
	
	// Check for JSDoc
	nodeLine := int(node.StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[nodeLine]; exists {
		if jsdoc.isPrivate {
			fn.Visibility = ir.VisibilityPrivate
		} else if jsdoc.isProtected {
			fn.Visibility = ir.VisibilityProtected
		}
		if jsdoc.returnType != "" {
			fn.Returns = &ir.TypeRef{Name: jsdoc.returnType}
		}
		if jsdoc.isAsync {
			isAsync = true
		}
		// We'll apply param types when processing parameters
	}
	
	// Process function parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "async":
			isAsync = true
			
		case "*":
			isGenerator = true
			
		case "identifier":
			fn.Name = p.getNodeText(child)
			
		case "formal_parameters":
			p.processParameters(child, fn)
			
		case "statement_block":
			// Function body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
	
	if isAsync {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
	}
	
	if isGenerator {
		fn.Name = "*" + fn.Name
	}
	
	p.addNode(file, parent, fn)
}

// processParameters processes function/method parameters
func (p *TreeSitterProcessor) processParameters(node *sitter.Node, fn *ir.DistilledFunction) {
	// Get JSDoc for this function
	var jsdocParamTypes map[string]string
	fnLine := int(node.Parent().StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[fnLine]; exists && jsdoc.paramTypes != nil {
		jsdocParamTypes = jsdoc.paramTypes
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			param := ir.Parameter{
				Name: p.getNodeText(child),
			}
			// Check if we have JSDoc type for this param
			if jsdocParamTypes != nil {
				if jsdocType, exists := jsdocParamTypes[param.Name]; exists {
					param.Type = ir.TypeRef{Name: jsdocType}
				}
			}
			fn.Parameters = append(fn.Parameters, param)
			
		case "rest_pattern":
			// ...args
			for j := 0; j < int(child.ChildCount()); j++ {
				if child.Child(j).Type() == "identifier" {
					param := ir.Parameter{
						Name: "..." + p.getNodeText(child.Child(j)),
					}
					if jsdocParamTypes != nil {
						if jsdocType, exists := jsdocParamTypes[p.getNodeText(child.Child(j))]; exists {
							param.Type = ir.TypeRef{Name: jsdocType}
						}
					}
					fn.Parameters = append(fn.Parameters, param)
				}
			}
			
		case "assignment_pattern":
			// param = defaultValue
			var paramName, defaultValue string
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "identifier" {
					paramName = p.getNodeText(grandchild)
				} else if grandchild.Type() != "=" {
					defaultValue = p.getNodeText(grandchild)
				}
			}
			param := ir.Parameter{
				Name:         paramName,
				DefaultValue: defaultValue,
			}
			if jsdocParamTypes != nil {
				if jsdocType, exists := jsdocParamTypes[paramName]; exists {
					param.Type = ir.TypeRef{Name: jsdocType}
				}
			}
			fn.Parameters = append(fn.Parameters, param)
			
		case "object_pattern", "array_pattern":
			// Destructuring parameters
			param := ir.Parameter{
				Name: p.getNodeText(child),
			}
			fn.Parameters = append(fn.Parameters, param)
		}
	}
}

// processVariableDeclaration processes var declarations
func (p *TreeSitterProcessor) processVariableDeclaration(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Process each variable declarator
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_declarator" {
			p.processVariableDeclarator(child, file, parent, "var")
		}
	}
}

// processLexicalDeclaration processes let/const declarations
func (p *TreeSitterProcessor) processLexicalDeclaration(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	var kind string
	
	// Determine if let or const
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "let" || child.Type() == "const" {
			kind = child.Type()
			break
		}
	}
	
	// Process each variable declarator
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_declarator" {
			p.processVariableDeclarator(child, file, parent, kind)
		}
	}
}

// processVariableDeclarator processes individual variable declarations
func (p *TreeSitterProcessor) processVariableDeclarator(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode, kind string) {
	field := &ir.DistilledField{
		BaseNode:   p.nodeLocation(node),
		Visibility: ir.VisibilityPublic, // Variables are public by default
		Modifiers:  []ir.Modifier{},
	}
	
	// Add const modifier if applicable
	if kind == "const" {
		field.Modifiers = append(field.Modifiers, ir.ModifierFinal)
	}
	
	// Check for JSDoc
	nodeLine := int(node.StartPoint().Row) + 1
	if jsdoc, exists := p.jsdocComments[nodeLine]; exists {
		if jsdoc.isPrivate {
			field.Visibility = ir.VisibilityPrivate
		} else if jsdoc.isProtected {
			field.Visibility = ir.VisibilityProtected
		}
		if jsdoc.typedefType != "" {
			field.Type = &ir.TypeRef{Name: jsdoc.typedefType}
		}
	}
	
	// Process declarator parts
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			field.Name = p.getNodeText(child)
			
		case "object_pattern", "array_pattern":
			// Destructuring
			field.Name = p.getNodeText(child)
			
		default:
			// Initializer
			if child.Type() != "=" && field.DefaultValue == "" {
				// Check if it's a function expression, generator function, or arrow function
				if child.Type() == "function_expression" || child.Type() == "arrow_function" || child.Type() == "generator_function" {
					// Store as a function instead
					fn := &ir.DistilledFunction{
						BaseNode:   p.nodeLocation(child),
						Name:       field.Name,
						Visibility: field.Visibility,
						Modifiers:  field.Modifiers,
						Parameters: []ir.Parameter{},
					}
					
					// Process function expression
					if child.Type() == "arrow_function" {
						p.processArrowFunction(child, fn)
					} else {
						p.processFunctionExpression(child, fn)
					}
					
					p.addNode(file, parent, fn)
					return
				} else {
					field.DefaultValue = p.getNodeText(child)
				}
			}
		}
	}
	
	p.addNode(file, parent, field)
}

// processArrowFunction processes arrow function expressions
func (p *TreeSitterProcessor) processArrowFunction(node *sitter.Node, fn *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "async":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
			
		case "identifier":
			// Single parameter without parentheses
			fn.Parameters = append(fn.Parameters, ir.Parameter{
				Name: p.getNodeText(child),
			})
			
		case "formal_parameters":
			p.processParameters(child, fn)
			
		case "statement_block", "expression":
			// Function body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
}

// processFunctionExpression processes function expressions
func (p *TreeSitterProcessor) processFunctionExpression(node *sitter.Node, fn *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "async":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
			
		case "*":
			// Generator
			fn.Name = "*" + fn.Name
			
		case "identifier":
			// Named function expression
			// Keep the original variable name
			
		case "formal_parameters":
			p.processParameters(child, fn)
			
		case "statement_block":
			// Function body
			startByte := child.StartByte()
			endByte := child.EndByte()
			if int(startByte) < len(p.source) && int(endByte) <= len(p.source) {
				fn.Implementation = string(p.source[startByte:endByte])
			}
		}
	}
}

// processComment processes comments
func (p *TreeSitterProcessor) processComment(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	text := p.getNodeText(node)
	
	// Skip JSDoc comments as they're already processed
	if strings.HasPrefix(text, "/**") && !strings.HasPrefix(text, "/***") {
		return
	}
	
	// Determine comment type
	format := "line"
	if strings.HasPrefix(text, "/*") {
		format = "block"
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
	} else if strings.HasPrefix(text, "//") {
		text = strings.TrimPrefix(text, "//")
	}
	
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     strings.TrimSpace(text),
		Format:   format,
	}
	
	p.addNode(file, parent, comment)
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
		case *ir.DistilledFunction:
			// Functions don't typically have children in JavaScript
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