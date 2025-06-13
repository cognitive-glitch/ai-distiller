package java

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TreeSitterProcessor uses tree-sitter for Java parsing
type TreeSitterProcessor struct {
	parser *sitter.Parser
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() *TreeSitterProcessor {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_java.Language()))
	
	return &TreeSitterProcessor{
		parser: parser,
	}
}

// ProcessSource processes Java source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	// Parse the source code
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Java code: %w", err)
	}
	defer tree.Close()

	// Create distilled file
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   int(tree.RootNode().EndPoint().Row) + 1,
			},
		},
		Path:     filename,
		Language: "java",
		Version:  "1.0",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}

	// Process the tree
	p.processNode(tree.RootNode(), source, file, nil)

	return file, nil
}

// processNode recursively processes tree-sitter nodes
func (p *TreeSitterProcessor) processNode(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	nodeType := node.Type()

	switch nodeType {
	case "package_declaration":
		p.processPackageDeclaration(node, source, file)
	case "import_declaration":
		p.processImportDeclaration(node, source, file)
	case "class_declaration":
		p.processClassDeclaration(node, source, file, parent)
	case "interface_declaration":
		p.processInterfaceDeclaration(node, source, file, parent)
	case "enum_declaration":
		p.processEnumDeclaration(node, source, file, parent)
	case "record_declaration":
		p.processRecordDeclaration(node, source, file, parent)
	case "method_declaration":
		p.processMethodDeclaration(node, source, file, parent)
	case "field_declaration":
		p.processFieldDeclaration(node, source, file, parent)
	case "constructor_declaration":
		p.processConstructorDeclaration(node, source, file, parent)
	default:
		// Recursively process children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			p.processNode(child, source, file, parent)
		}
	}
}

// processPackageDeclaration handles package declarations
func (p *TreeSitterProcessor) processPackageDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile) {
	pkg := &ir.DistilledPackage{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "scoped_identifier" || child.Type() == "identifier" {
			pkg.Name = string(source[child.StartByte():child.EndByte()])
			break
		}
	}

	file.Children = append(file.Children, pkg)
}

// processImportDeclaration handles import statements
func (p *TreeSitterProcessor) processImportDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile) {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		ImportType: "import",
		Symbols:    []ir.ImportedSymbol{},
	}

	isStatic := false
	var importPath string
	hasAsterisk := false

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "static":
			isStatic = true
		case "scoped_identifier", "identifier":
			importPath = string(source[child.StartByte():child.EndByte()])
		case "asterisk":
			hasAsterisk = true
		}
	}

	if isStatic {
		imp.ImportType = "static import"
	}

	imp.Module = importPath
	if hasAsterisk {
		imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: "*"})
	} else {
		// Extract the last part as the imported class/member
		parts := strings.Split(importPath, ".")
		if len(parts) > 0 {
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: parts[len(parts)-1]})
		}
	}

	file.Children = append(file.Children, imp)
}

// processClassDeclaration handles class declarations
func (p *TreeSitterProcessor) processClassDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract modifiers, name, extends, implements
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractModifiers(child, source, class)
		case "identifier":
			class.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameters":
			p.extractTypeParameters(child, source, class)
		case "superclass":
			p.extractSuperclass(child, source, class)
		case "super_interfaces":
			p.extractInterfaces(child, source, class)
		case "permits":
			p.extractPermits(child, source, class)
		case "class_body":
			p.processClassBody(child, source, file, class)
		}
	}

	// Set default visibility if not specified
	if class.Visibility == "" {
		class.Visibility = ir.VisibilityPackage
	}

	p.addToParent(file, parent, class)
}

// processInterfaceDeclaration handles interface declarations
func (p *TreeSitterProcessor) processInterfaceDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	iface := &ir.DistilledInterface{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract modifiers, name, extends
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractInterfaceModifiers(child, source, iface)
		case "identifier":
			iface.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameters":
			p.extractInterfaceTypeParameters(child, source, iface)
		case "extends_interfaces":
			p.extractExtendsInterfaces(child, source, iface)
		case "permits":
			p.extractInterfacePermits(child, source, iface)
		case "interface_body":
			p.processInterfaceBody(child, source, file, iface)
		}
	}

	// Set default visibility if not specified
	if iface.Visibility == "" {
		iface.Visibility = ir.VisibilityPublic // Interfaces are public by default
	}

	p.addToParent(file, parent, iface)
}

// processEnumDeclaration handles enum declarations
func (p *TreeSitterProcessor) processEnumDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	enum := &ir.DistilledEnum{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract modifiers, name, implements
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractEnumModifiers(child, source, enum)
		case "identifier":
			enum.Name = string(source[child.StartByte():child.EndByte()])
		case "enum_body":
			p.processEnumBody(child, source, file, enum)
		}
	}

	// Set default visibility if not specified
	if enum.Visibility == "" {
		enum.Visibility = ir.VisibilityPackage
	}

	p.addToParent(file, parent, enum)
}

// processRecordDeclaration handles record declarations (Java 14+)
func (p *TreeSitterProcessor) processRecordDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Records are represented as classes with a special modifier
	record := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Modifiers: []ir.Modifier{ir.ModifierData}, // Use data modifier to indicate record
		Children:  []ir.DistilledNode{},
	}

	// Extract modifiers, name, parameters, implements
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractModifiers(child, source, record)
		case "identifier":
			record.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameters":
			p.extractTypeParameters(child, source, record)
		case "formal_parameters":
			// Record components are represented as constructor parameters
			p.extractRecordComponents(child, source, record)
		case "super_interfaces":
			p.extractInterfaces(child, source, record)
		case "class_body":
			p.processClassBody(child, source, file, record)
		}
	}

	// Set default visibility if not specified
	if record.Visibility == "" {
		record.Visibility = ir.VisibilityPackage
	}

	p.addToParent(file, parent, record)
}

// processMethodDeclaration handles method declarations
func (p *TreeSitterProcessor) processMethodDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	method := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Parameters: []ir.Parameter{},
		Modifiers:  []ir.Modifier{},
	}

	// Extract modifiers, return type, name, parameters, body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractMethodModifiers(child, source, method)
		case "type_parameters":
			// TODO: Handle generic methods
		case "void_type":
			method.Returns = &ir.TypeRef{Name: "void"}
		case "type_identifier", "integral_type", "floating_point_type", "boolean_type":
			method.Returns = &ir.TypeRef{Name: string(source[child.StartByte():child.EndByte()])}
		case "array_type", "generic_type":
			method.Returns = p.extractType(child, source)
		case "identifier":
			method.Name = string(source[child.StartByte():child.EndByte()])
		case "formal_parameters":
			p.extractParameters(child, source, method)
		case "throws":
			p.extractThrows(child, source, method)
		case "block":
			method.Implementation = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Set default visibility if not specified
	if method.Visibility == "" {
		method.Visibility = ir.VisibilityPackage
	}

	p.addToParent(file, parent, method)
}

// processConstructorDeclaration handles constructor declarations
func (p *TreeSitterProcessor) processConstructorDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	constructor := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Parameters: []ir.Parameter{},
		Modifiers:  []ir.Modifier{},
		// Constructor is identified by matching name with class name
	}

	// Extract modifiers, name, parameters, body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			p.extractMethodModifiers(child, source, constructor)
		case "identifier":
			constructor.Name = string(source[child.StartByte():child.EndByte()])
		case "formal_parameters":
			p.extractParameters(child, source, constructor)
		case "throws":
			p.extractThrows(child, source, constructor)
		case "constructor_body":
			constructor.Implementation = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Set default visibility if not specified
	if constructor.Visibility == "" {
		constructor.Visibility = ir.VisibilityPackage
	}

	p.addToParent(file, parent, constructor)
}

// processFieldDeclaration handles field declarations
func (p *TreeSitterProcessor) processFieldDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// A field declaration can declare multiple fields
	var fieldType *ir.TypeRef
	var modifiers []ir.Modifier
	var visibility ir.Visibility
	var decorators []string

	// Extract modifiers and type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			modifiers, visibility, decorators = p.extractFieldModifiers(child, source)
		case "type_identifier", "integral_type", "floating_point_type", "boolean_type":
			fieldType = &ir.TypeRef{Name: string(source[child.StartByte():child.EndByte()])}
		case "array_type", "generic_type":
			fieldType = p.extractType(child, source)
		case "variable_declarator":
			// Each variable_declarator is a separate field
			field := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.nodeLocation(child),
				},
				Type:        fieldType,
				Modifiers:   modifiers,
				Visibility:  visibility,
				Decorators:  decorators,
			}

			// Extract field name and value
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				switch grandchild.Type() {
				case "identifier":
					field.Name = string(source[grandchild.StartByte():grandchild.EndByte()])
				case "=":
					// Skip assignment operator
				default:
					// This is the value/initializer
					field.DefaultValue = string(source[grandchild.StartByte():grandchild.EndByte()])
				}
			}

			// Set default visibility if not specified
			if field.Visibility == "" {
				field.Visibility = ir.VisibilityPackage
			}

			p.addToParent(file, parent, field)
		}
	}
}

// Helper methods

// extractModifiers extracts modifiers for classes
func (p *TreeSitterProcessor) extractModifiers(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := string(source[child.StartByte():child.EndByte()])
		
		switch text {
		case "public":
			class.Visibility = ir.VisibilityPublic
		case "protected":
			class.Visibility = ir.VisibilityProtected
		case "private":
			class.Visibility = ir.VisibilityPrivate
		case "abstract":
			class.Modifiers = append(class.Modifiers, ir.ModifierAbstract)
		case "final":
			class.Modifiers = append(class.Modifiers, ir.ModifierFinal)
		case "static":
			class.Modifiers = append(class.Modifiers, ir.ModifierStatic)
		case "sealed":
			class.Modifiers = append(class.Modifiers, ir.ModifierSealed)
		case "non-sealed":
			// Non-sealed is the absence of sealed, no need for special modifier
		}

		// Handle annotations as decorators
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			p.extractAnnotation(child, source, &class.Decorators)
		}
	}
}

// extractEnumModifiers extracts modifiers for enums
func (p *TreeSitterProcessor) extractEnumModifiers(node *sitter.Node, source []byte, enum *ir.DistilledEnum) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := string(source[child.StartByte():child.EndByte()])
		
		switch text {
		case "public":
			enum.Visibility = ir.VisibilityPublic
		case "protected":
			enum.Visibility = ir.VisibilityProtected
		case "private":
			enum.Visibility = ir.VisibilityPrivate
		}

		// Handle annotations as decorators
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			// TODO: Add decorator support to enums in IR if needed
		}
	}
}

// extractInterfaceModifiers extracts modifiers for interfaces
func (p *TreeSitterProcessor) extractInterfaceModifiers(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := string(source[child.StartByte():child.EndByte()])
		
		switch text {
		case "public":
			iface.Visibility = ir.VisibilityPublic
		case "protected":
			iface.Visibility = ir.VisibilityProtected
		case "private":
			iface.Visibility = ir.VisibilityPrivate
		case "sealed":
			// TODO: Add sealed modifier support to interfaces in IR if needed
		}

		// Handle annotations as decorators
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			// TODO: Add decorator support to interfaces in IR
		}
	}
}

// extractMethodModifiers extracts modifiers for methods
func (p *TreeSitterProcessor) extractMethodModifiers(node *sitter.Node, source []byte, method *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := string(source[child.StartByte():child.EndByte()])
		
		switch text {
		case "public":
			method.Visibility = ir.VisibilityPublic
		case "protected":
			method.Visibility = ir.VisibilityProtected
		case "private":
			method.Visibility = ir.VisibilityPrivate
		case "abstract":
			method.Modifiers = append(method.Modifiers, ir.ModifierAbstract)
		case "final":
			method.Modifiers = append(method.Modifiers, ir.ModifierFinal)
		case "static":
			method.Modifiers = append(method.Modifiers, ir.ModifierStatic)
		case "synchronized":
			// TODO: Add synchronized modifier to IR if needed
		case "native":
			method.Modifiers = append(method.Modifiers, ir.ModifierExtern) // Use extern for native
		case "default":
			// Default methods in interfaces - no special modifier needed
		}

		// Handle annotations as decorators
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			p.extractAnnotation(child, source, &method.Decorators)
		}
	}
}

// extractFieldModifiers extracts modifiers for fields
func (p *TreeSitterProcessor) extractFieldModifiers(node *sitter.Node, source []byte) ([]ir.Modifier, ir.Visibility, []string) {
	var modifiers []ir.Modifier
	var visibility ir.Visibility
	var decorators []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := string(source[child.StartByte():child.EndByte()])
		
		switch text {
		case "public":
			visibility = ir.VisibilityPublic
		case "protected":
			visibility = ir.VisibilityProtected
		case "private":
			visibility = ir.VisibilityPrivate
		case "final":
			modifiers = append(modifiers, ir.ModifierFinal)
		case "static":
			modifiers = append(modifiers, ir.ModifierStatic)
		case "volatile":
			modifiers = append(modifiers, ir.ModifierVolatile)
		case "transient":
			modifiers = append(modifiers, ir.ModifierTransient)
		}

		// Handle annotations as decorators
		if child.Type() == "marker_annotation" || child.Type() == "annotation" {
			p.extractAnnotation(child, source, &decorators)
		}
	}

	return modifiers, visibility, decorators
}

// extractAnnotation extracts annotation information as decorator
func (p *TreeSitterProcessor) extractAnnotation(node *sitter.Node, source []byte, decorators *[]string) {
	var annName string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			annName = "@" + string(source[child.StartByte():child.EndByte()])
		case "annotation_argument_list":
			// For simplicity, just append the whole annotation with arguments
			annName = "@" + string(source[node.StartByte()+1:node.EndByte()])
			break
		}
	}

	if annName != "" {
		*decorators = append(*decorators, annName)
	}
}

// extractTypeParameters extracts generic type parameters
func (p *TreeSitterProcessor) extractTypeParameters(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	// TODO: Extract and store type parameters
}

// extractInterfaceTypeParameters extracts generic type parameters for interfaces
func (p *TreeSitterProcessor) extractInterfaceTypeParameters(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	// TODO: Extract and store type parameters
}

// extractSuperclass extracts the superclass
func (p *TreeSitterProcessor) extractSuperclass(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_identifier" || child.Type() == "scoped_type_identifier" || child.Type() == "generic_type" {
			superType := p.extractType(child, source)
			class.Extends = append(class.Extends, *superType)
			break
		}
	}
}

// extractInterfaces extracts implemented interfaces
func (p *TreeSitterProcessor) extractInterfaces(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_list" {
			p.extractTypeList(child, source, &class.Implements)
		}
	}
}

// extractExtendsInterfaces extracts extended interfaces for interface declarations
func (p *TreeSitterProcessor) extractExtendsInterfaces(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_list" {
			p.extractTypeList(child, source, &iface.Extends)
		}
	}
}

// extractPermits extracts permitted subclasses for sealed classes
func (p *TreeSitterProcessor) extractPermits(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	// Java's permits clause is specific to sealed classes
	// We can store this information in class extensions if needed
	// For now, we'll skip it as the IR doesn't have a Permits field
}

// extractInterfacePermits extracts permitted implementations for sealed interfaces
func (p *TreeSitterProcessor) extractInterfacePermits(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	// Java's permits clause is specific to sealed interfaces
	// We can store this information in interface extensions if needed
	// For now, we'll skip it as the IR doesn't have a Permits field
}

// extractTypeList extracts a list of types
func (p *TreeSitterProcessor) extractTypeList(node *sitter.Node, source []byte, list *[]ir.TypeRef) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_identifier" || child.Type() == "scoped_type_identifier" || child.Type() == "generic_type" {
			typeRef := p.extractType(child, source)
			*list = append(*list, *typeRef)
		}
	}
}

// extractType extracts type information
func (p *TreeSitterProcessor) extractType(node *sitter.Node, source []byte) *ir.TypeRef {
	nodeType := node.Type()
	
	switch nodeType {
	case "type_identifier", "integral_type", "floating_point_type", "boolean_type":
		return &ir.TypeRef{Name: string(source[node.StartByte():node.EndByte()])}
	case "scoped_type_identifier":
		// Handle qualified type names like java.util.List
		return &ir.TypeRef{Name: string(source[node.StartByte():node.EndByte()])}
	case "generic_type":
		// Handle generic types like List<String>
		typeName := ""
		var generics []ir.TypeRef

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			switch child.Type() {
			case "type_identifier", "scoped_type_identifier":
				typeName = string(source[child.StartByte():child.EndByte()])
			case "type_arguments":
				// Extract generic arguments
				for j := 0; j < int(child.ChildCount()); j++ {
					argChild := child.Child(j)
					if argChild.Type() != "<" && argChild.Type() != ">" && argChild.Type() != "," {
						argType := p.extractType(argChild, source)
						generics = append(generics, *argType)
					}
				}
			}
		}

		return &ir.TypeRef{
			Name:     typeName,
			TypeArgs: generics,
		}
	case "array_type":
		// Handle array types
		var elementType *ir.TypeRef
		dimensions := 0

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dimensions" {
				// Count the number of dimensions
				text := string(source[child.StartByte():child.EndByte()])
				dimensions = strings.Count(text, "[")
			} else if elementType == nil {
				elementType = p.extractType(child, source)
			}
		}

		if elementType != nil {
			// Append [] to the type name for each dimension
			arrayType := elementType.Name
			for i := 0; i < dimensions; i++ {
				arrayType += "[]"
			}
			return &ir.TypeRef{Name: arrayType}
		}
	}

	// Default: return the raw text
	return &ir.TypeRef{Name: string(source[node.StartByte():node.EndByte()])}
}

// extractParameters extracts method parameters
func (p *TreeSitterProcessor) extractParameters(node *sitter.Node, source []byte, method *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "formal_parameter" || child.Type() == "spread_parameter" {
			param := p.extractParameter(child, source)
			method.Parameters = append(method.Parameters, *param)
		}
	}
}

// extractParameter extracts a single parameter
func (p *TreeSitterProcessor) extractParameter(node *sitter.Node, source []byte) *ir.Parameter {
	param := &ir.Parameter{}
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifiers":
			// Handle parameter modifiers like final
			for j := 0; j < int(child.ChildCount()); j++ {
				modChild := child.Child(j)
				if string(source[modChild.StartByte():modChild.EndByte()]) == "final" {
					// TODO: Add IsFinal support to Parameter in IR if needed
				}
			}
		case "type_identifier", "integral_type", "floating_point_type", "boolean_type":
			param.Type = *p.extractType(child, source)
		case "array_type", "generic_type":
			param.Type = *p.extractType(child, source)
		case "variable_declarator":
			// Extract parameter name
			for j := 0; j < int(child.ChildCount()); j++ {
				varChild := child.Child(j)
				if varChild.Type() == "identifier" {
					param.Name = string(source[varChild.StartByte():varChild.EndByte()])
					break
				}
			}
		case "identifier":
			param.Name = string(source[child.StartByte():child.EndByte()])
		case "...":
			param.IsVariadic = true
		}
	}

	// Handle spread parameters (varargs)
	if node.Type() == "spread_parameter" {
		param.IsVariadic = true
	}

	return param
}

// extractRecordComponents extracts record component parameters
func (p *TreeSitterProcessor) extractRecordComponents(node *sitter.Node, source []byte, record *ir.DistilledClass) {
	// Record components are like constructor parameters but also define fields
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "formal_parameter" {
			param := p.extractParameter(child, source)
			
			// Create a field for each record component
			field := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.nodeLocation(child),
				},
				Name:       param.Name,
				Type:       &param.Type,
				Visibility: ir.VisibilityPublic, // Record components are always public
				Modifiers:  []ir.Modifier{ir.ModifierFinal}, // Record fields are always final
			}
			
			record.Children = append(record.Children, field)
		}
	}
}

// extractThrows extracts thrown exceptions
func (p *TreeSitterProcessor) extractThrows(node *sitter.Node, source []byte, method *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_identifier" || child.Type() == "scoped_type_identifier" {
			exception := p.extractType(child, source)
			method.Throws = append(method.Throws, *exception)
		}
	}
}

// processClassBody processes the body of a class
func (p *TreeSitterProcessor) processClassBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
	}
}

// processInterfaceBody processes the body of an interface
func (p *TreeSitterProcessor) processInterfaceBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
	}
}

// processEnumBody processes the body of an enum
func (p *TreeSitterProcessor) processEnumBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "enum_constant" {
			p.processEnumConstant(child, source, file, parent)
		} else if child.Type() == "enum_body_declarations" {
			// Process methods and fields in enum
			for j := 0; j < int(child.ChildCount()); j++ {
				decl := child.Child(j)
				p.processNode(decl, source, file, parent)
			}
		}
	}
}

// processEnumConstant processes enum constants
func (p *TreeSitterProcessor) processEnumConstant(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierFinal},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			field.Name = string(source[child.StartByte():child.EndByte()])
		case "argument_list":
			// Capture the arguments, including parentheses
			field.DefaultValue = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Enum constants have the enum type
	if enum, ok := parent.(*ir.DistilledEnum); ok {
		field.Type = &ir.TypeRef{Name: enum.Name}
	}

	p.addToParent(file, parent, field)
}

// nodeLocation converts tree-sitter node position to IR location
func (p *TreeSitterProcessor) nodeLocation(node *sitter.Node) ir.Location {
	return ir.Location{
		StartLine:   int(node.StartPoint().Row) + 1,
		StartColumn: int(node.StartPoint().Column) + 1,
		EndLine:     int(node.EndPoint().Row) + 1,
		EndColumn:   int(node.EndPoint().Column) + 1,
	}
}

// addToParent adds a node to its parent or to the file
func (p *TreeSitterProcessor) addToParent(file *ir.DistilledFile, parent ir.DistilledNode, child ir.DistilledNode) {
	if parent != nil {
		switch p := parent.(type) {
		case *ir.DistilledClass:
			p.Children = append(p.Children, child)
		case *ir.DistilledInterface:
			p.Children = append(p.Children, child)
		case *ir.DistilledEnum:
			p.Children = append(p.Children, child)
		default:
			file.Children = append(file.Children, child)
		}
	} else {
		file.Children = append(file.Children, child)
	}
}