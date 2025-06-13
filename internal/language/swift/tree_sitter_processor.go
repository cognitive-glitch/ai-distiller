package swift

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	swift "tree-sitter-swift"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TreeSitterProcessor processes Swift using tree-sitter
type TreeSitterProcessor struct {
	parser   *sitter.Parser
	source   []byte
	filename string
	
	// State for context-aware parsing
	currentModule string
	importedTypes map[string]string
	
	// Track current context for nested structures
	currentClass    *ir.DistilledClass
	currentProtocol *ir.DistilledInterface
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() (*TreeSitterProcessor, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(swift.Language()))
	
	return &TreeSitterProcessor{
		parser:        parser,
		importedTypes: make(map[string]string),
	}, nil
}

// ProcessSource processes Swift source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	
	// Reset state
	p.currentModule = ""
	p.importedTypes = make(map[string]string)
	p.currentClass = nil
	p.currentProtocol = nil
	
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
		Language: "swift",
		Version:  "5.x",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	// Process the AST
	p.processNode(tree.RootNode(), file, nil)
	
	return file, nil
}

// processNode recursively processes tree-sitter nodes
func (p *TreeSitterProcessor) processNode(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	if node == nil {
		return
	}
	
	nodeType := node.Type()
	
	// Skip non-named nodes except comments
	if !node.IsNamed() && nodeType != "comment" && nodeType != "multiline_comment" {
		// Process children for anonymous nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		return
	}
	
	switch nodeType {
	case "source_file":
		// Process all top-level declarations
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
		
	case "import_declaration":
		p.processImport(node, file, parent)
		
	case "class_declaration":
		// Check the first child to determine if it's a class, struct, or enum
		if node.ChildCount() > 0 {
			firstChild := node.Child(0)
			switch firstChild.Type() {
			case "struct":
				p.processStruct(node, file, parent)
			case "enum":
				p.processEnum(node, file, parent)
			default:
				p.processClass(node, file, parent)
			}
		} else {
			p.processClass(node, file, parent)
		}
		
	case "protocol_declaration":
		p.processProtocol(node, file, parent)
		
	case "extension_declaration":
		p.processExtension(node, file, parent)
		
	case "function_declaration":
		p.processFunction(node, file, parent)
		
	case "property_declaration":
		p.processProperty(node, file, parent)
		
	case "actor_declaration":
		p.processActor(node, file, parent)
		
	case "comment", "multiline_comment":
		p.processComment(node, file, parent)
		
	default:
		// Process children for unhandled node types
		for i := 0; i < int(node.ChildCount()); i++ {
			p.processNode(node.Child(i), file, parent)
		}
	}
}

// Helper functions for processing specific node types

func (p *TreeSitterProcessor) processImport(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	imp := &ir.DistilledImport{
		BaseNode:   p.nodeLocation(node),
		ImportType: "import",
		Module:     p.getImportPath(node),
		Symbols:    []ir.ImportedSymbol{},
	}
	
	p.addNode(file, parent, imp)
}

func (p *TreeSitterProcessor) processClass(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode:    p.nodeLocation(node),
		Name:        p.getNodeName(node),
		Visibility:  p.getVisibility(node),
		Modifiers:   p.getModifiers(node),
		Extends:     p.getSuperclasses(node),
		Implements:  p.getProtocols(node),
		Children:    []ir.DistilledNode{},
		Decorators:  p.getAttributes(node),
	}
	
	// Set current class context
	prevClass := p.currentClass
	p.currentClass = class
	
	// Process class body
	// Process all children of the class declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "class_body" {
			// Process class body contents
			for j := 0; j < int(child.ChildCount()); j++ {
				p.processNode(child.Child(j), file, class)
			}
		}
	}
	
	// Restore previous context
	p.currentClass = prevClass
	
	p.addNode(file, parent, class)
}

func (p *TreeSitterProcessor) processFunction(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	fn := &ir.DistilledFunction{
		BaseNode:    p.nodeLocation(node),
		Name:        p.getFunctionName(node),
		Visibility:  p.getVisibility(node),
		Modifiers:   p.getFunctionModifiers(node),
		Parameters:  p.getParameters(node),
		Returns:     p.getReturnType(node),
		Decorators:  p.getAttributes(node),
		TypeParams:  p.getGenericParameters(node),
	}
	
	// Get implementation if present
	if body := p.findChildByType(node, "function_body"); body != nil {
		fn.Implementation = p.getNodeText(body)
	}
	
	p.addNode(file, parent, fn)
}

// Helper methods

func (p *TreeSitterProcessor) nodeLocation(node *sitter.Node) ir.BaseNode {
	return ir.BaseNode{
		Location: ir.Location{
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
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

func (p *TreeSitterProcessor) findChildByType(node *sitter.Node, nodeType string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return child
		}
	}
	return nil
}

func (p *TreeSitterProcessor) getNodeName(node *sitter.Node) string {
	// For class/struct/enum declarations, the name is in type_identifier
	if nameNode := p.findChildByType(node, "type_identifier"); nameNode != nil {
		return p.getNodeText(nameNode)
	}
	// For functions and other declarations, it's in simple_identifier
	if nameNode := p.findChildByType(node, "simple_identifier"); nameNode != nil {
		return p.getNodeText(nameNode)
	}
	if nameNode := p.findChildByType(node, "identifier"); nameNode != nil {
		return p.getNodeText(nameNode)
	}
	return ""
}

func (p *TreeSitterProcessor) getVisibility(node *sitter.Node) ir.Visibility {
	// Look for modifiers child node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "modifiers" {
			// Look for visibility_modifier within modifiers
			for j := 0; j < int(child.ChildCount()); j++ {
				modChild := child.Child(j)
				if modChild.Type() == "visibility_modifier" {
					// Get the actual visibility keyword
					if modChild.ChildCount() > 0 {
						visKeyword := modChild.Child(0)
						switch visKeyword.Type() {
						case "private":
							return ir.VisibilityPrivate
						case "fileprivate":
							return ir.VisibilityFilePrivate
						case "internal":
							return ir.VisibilityInternal
						case "public":
							return ir.VisibilityPublic
						case "open":
							return ir.VisibilityOpen
						}
					}
				}
			}
		}
	}
	// Default visibility in Swift is internal
	return ir.VisibilityInternal
}

func (p *TreeSitterProcessor) getModifiers(node *sitter.Node) []ir.Modifier {
	var modifiers []ir.Modifier
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := p.getNodeText(child)
		
		switch text {
		case "final":
			modifiers = append(modifiers, ir.ModifierFinal)
		case "static":
			modifiers = append(modifiers, ir.ModifierStatic)
		case "class":
			if child.Type() == "ownership_modifier" {
				modifiers = append(modifiers, ir.ModifierStatic)
			}
		}
	}
	
	return modifiers
}

func (p *TreeSitterProcessor) getFunctionModifiers(node *sitter.Node) []ir.Modifier {
	modifiers := p.getModifiers(node)
	
	// Check for async
	if p.findChildByType(node, "async_keyword") != nil {
		modifiers = append(modifiers, ir.ModifierAsync)
	}
	
	// Check for mutating
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "mutation_modifier" && p.getNodeText(child) == "mutating" {
			modifiers = append(modifiers, ir.ModifierMutating)
		}
	}
	
	return modifiers
}

// Stub implementations for remaining helper methods
func (p *TreeSitterProcessor) getImportPath(node *sitter.Node) string {
	// Look for identifier after import keyword
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "simple_identifier" {
			return p.getNodeText(child)
		}
	}
	return ""
}

func (p *TreeSitterProcessor) getSuperclasses(node *sitter.Node) []ir.TypeRef {
	var superclasses []ir.TypeRef
	
	// Look for type_inheritance_clause
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_inheritance_clause" {
			// Process each inherited type
			for j := 0; j < int(child.ChildCount()); j++ {
				inheritChild := child.Child(j)
				if inheritChild.Type() == "user_type" {
					superclasses = append(superclasses, ir.TypeRef{
						Name: p.getNodeText(inheritChild),
					})
				}
			}
		}
	}
	
	return superclasses
}

func (p *TreeSitterProcessor) getProtocols(node *sitter.Node) []ir.TypeRef {
	// In Swift, both superclasses and protocols are in the type_inheritance_clause
	// For now, we'll treat them the same as getSuperclasses
	// A more sophisticated implementation would distinguish between them
	return p.getSuperclasses(node)
}

func (p *TreeSitterProcessor) getAttributes(node *sitter.Node) []string {
	// TODO: Implement attribute extraction
	return nil
}

func (p *TreeSitterProcessor) getFunctionName(node *sitter.Node) string {
	return p.getNodeName(node)
}

func (p *TreeSitterProcessor) getParameters(node *sitter.Node) []ir.Parameter {
	var params []ir.Parameter
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter" {
			param := p.extractParameter(child)
			if param.Name != "" {
				params = append(params, param)
			}
		}
	}
	
	return params
}

func (p *TreeSitterProcessor) extractParameter(node *sitter.Node) ir.Parameter {
	param := ir.Parameter{}
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		switch child.Type() {
		case "simple_identifier":
			if param.Name == "" {
				// First identifier is external name or parameter name
				param.Name = p.getNodeText(child)
			}
		case "user_type", "optional_type", "array_type":
			param.Type = ir.TypeRef{
				Name: p.getNodeText(child),
			}
		}
	}
	
	return param
}

func (p *TreeSitterProcessor) getReturnType(node *sitter.Node) *ir.TypeRef {
	// Look for -> followed by a type
	foundArrow := false
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		
		if child.Type() == "->" {
			foundArrow = true
			continue
		}
		
		if foundArrow && (child.Type() == "user_type" || child.Type() == "optional_type" || 
			child.Type() == "array_type" || child.Type() == "tuple_type") {
			return &ir.TypeRef{
				Name: p.getNodeText(child),
			}
		}
	}
	
	return nil
}

func (p *TreeSitterProcessor) getGenericParameters(node *sitter.Node) []ir.TypeParam {
	// TODO: Implement generic parameter extraction
	return nil
}

func (p *TreeSitterProcessor) processStruct(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Structs are similar to classes but value types
	struct_ := &ir.DistilledClass{
		BaseNode:    p.nodeLocation(node),
		Name:        p.getNodeName(node),
		Visibility:  p.getVisibility(node),
		Modifiers:   append(p.getModifiers(node), ir.ModifierStruct),
		Extends:     p.getSuperclasses(node),
		Implements:  p.getProtocols(node),
		Children:    []ir.DistilledNode{},
		Decorators:  p.getAttributes(node),
	}
	
	// Set current class context
	prevClass := p.currentClass
	p.currentClass = struct_
	
	// Process all children of the struct declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.IsNamed() {
			p.processNode(child, file, struct_)
		}
	}
	
	// Restore previous context
	p.currentClass = prevClass
	
	p.addNode(file, parent, struct_)
}

func (p *TreeSitterProcessor) processEnum(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	enum := &ir.DistilledEnum{
		BaseNode:   p.nodeLocation(node),
		Name:       p.getNodeName(node),
		Visibility: p.getVisibility(node),
		Children:   []ir.DistilledNode{},
	}
	
	// Process enum body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "enum_class_body" {
			// Process enum cases
			for j := 0; j < int(child.ChildCount()); j++ {
				bodyChild := child.Child(j)
				if bodyChild.Type() == "enum_entry" {
					// Add enum case as a field
					caseName := ""
					for k := 0; k < int(bodyChild.ChildCount()); k++ {
						caseChild := bodyChild.Child(k)
						if caseChild.Type() == "simple_identifier" {
							caseName = p.getNodeText(caseChild)
							break
						}
					}
					if caseName != "" {
						enumCase := &ir.DistilledField{
							BaseNode: p.nodeLocation(bodyChild),
							Name:     caseName,
							Visibility: ir.VisibilityPublic, // Enum cases are always public
						}
						enum.Children = append(enum.Children, enumCase)
					}
				}
			}
		}
	}
	
	p.addNode(file, parent, enum)
}

func (p *TreeSitterProcessor) processProtocol(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	protocol := &ir.DistilledInterface{
		BaseNode:   p.nodeLocation(node),
		Name:       p.getNodeName(node),
		Visibility: p.getVisibility(node),
		Extends:    p.getProtocols(node), // Protocols can inherit from other protocols
		Children:   []ir.DistilledNode{},
	}
	
	// Set current protocol context
	prevProtocol := p.currentProtocol
	p.currentProtocol = protocol
	
	// Process protocol body
	if body := p.findChildByType(node, "protocol_body"); body != nil {
		for i := 0; i < int(body.ChildCount()); i++ {
			p.processNode(body.Child(i), file, protocol)
		}
	}
	
	// Restore previous context
	p.currentProtocol = prevProtocol
	
	p.addNode(file, parent, protocol)
}

func (p *TreeSitterProcessor) processExtension(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Extensions in Swift add functionality to existing types
	// We'll treat them as a special kind of class for now
	class := &ir.DistilledClass{
		BaseNode:    p.nodeLocation(node),
		Name:        "extension " + p.getExtendedTypeName(node),
		Visibility:  ir.VisibilityPublic,
		Implements:  p.getProtocols(node),
		Children:    []ir.DistilledNode{},
	}
	
	// Process extension body
	if body := p.findChildByType(node, "extension_body"); body != nil {
		for i := 0; i < int(body.ChildCount()); i++ {
			p.processNode(body.Child(i), file, class)
		}
	}
	
	p.addNode(file, parent, class)
}

func (p *TreeSitterProcessor) processActor(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Actors are similar to classes but with concurrency guarantees
	class := &ir.DistilledClass{
		BaseNode:    p.nodeLocation(node),
		Name:        p.getNodeName(node),
		Visibility:  p.getVisibility(node),
		Modifiers:   append(p.getModifiers(node), ir.ModifierActor),
		Implements:  p.getProtocols(node),
		Children:    []ir.DistilledNode{},
		Decorators:  p.getAttributes(node),
	}
	
	// Process actor body
	if body := p.findChildByType(node, "actor_body"); body != nil {
		for i := 0; i < int(body.ChildCount()); i++ {
			p.processNode(body.Child(i), file, class)
		}
	}
	
	p.addNode(file, parent, class)
}

func (p *TreeSitterProcessor) processProperty(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	field := &ir.DistilledField{
		BaseNode:   p.nodeLocation(node),
		Name:       p.getPropertyName(node),
		Visibility: p.getVisibility(node),
		Modifiers:  p.getModifiers(node),
		Type:       p.getPropertyType(node),
	}
	
	// Check for default value
	if init := p.findChildByType(node, "property_initializer"); init != nil {
		field.DefaultValue = p.getNodeText(init)
	}
	
	p.addNode(file, parent, field)
}

func (p *TreeSitterProcessor) processComment(node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	comment := &ir.DistilledComment{
		BaseNode: p.nodeLocation(node),
		Text:     p.getNodeText(node),
		Format:   p.getCommentFormat(node),
	}
	
	p.addNode(file, parent, comment)
}

func (p *TreeSitterProcessor) getExtendedTypeName(node *sitter.Node) string {
	// The type being extended is usually a `user_type` child of the extension_declaration node
	if typeNode := p.findChildByType(node, "user_type"); typeNode != nil {
		return p.getNodeText(typeNode)
	}
	// Also check for simple_identifier (for basic types)
	if idNode := p.findChildByType(node, "simple_identifier"); idNode != nil {
		return p.getNodeText(idNode)
	}
	return "Unknown" // Keep as fallback
}

func (p *TreeSitterProcessor) getPropertyName(node *sitter.Node) string {
	// Look for pattern -> simple_identifier
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "pattern" {
			// Look for simple_identifier within pattern
			for j := 0; j < int(child.ChildCount()); j++ {
				patternChild := child.Child(j)
				if patternChild.Type() == "simple_identifier" {
					return p.getNodeText(patternChild)
				}
			}
		}
	}
	return ""
}

func (p *TreeSitterProcessor) getPropertyType(node *sitter.Node) *ir.TypeRef {
	// Look for type_annotation -> type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_annotation" {
			// Look for the type within type_annotation
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild.IsNamed() && typeChild.Type() != ":" {
					return &ir.TypeRef{
						Name: p.getNodeText(typeChild),
					}
				}
			}
		}
	}
	return nil
}

func (p *TreeSitterProcessor) getCommentFormat(node *sitter.Node) string {
	text := p.getNodeText(node)
	if strings.HasPrefix(text, "///") || strings.HasPrefix(text, "/**") {
		return "doc"
	}
	if strings.Contains(text, "\n") {
		return "block"
	}
	return "line"
}

func (p *TreeSitterProcessor) addNode(file *ir.DistilledFile, parent ir.DistilledNode, node ir.DistilledNode) {
	if parent == nil {
		file.Children = append(file.Children, node)
		return
	}
	
	switch p := parent.(type) {
	case *ir.DistilledClass:
		p.Children = append(p.Children, node)
	case *ir.DistilledInterface:
		p.Children = append(p.Children, node)
	case *ir.DistilledEnum:
		p.Children = append(p.Children, node)
	default:
		file.Children = append(file.Children, node)
	}
}

// Close releases resources
func (p *TreeSitterProcessor) Close() {
	if p.parser != nil {
		p.parser.Close()
	}
}