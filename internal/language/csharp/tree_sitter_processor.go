package csharp

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_c_sharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TreeSitterProcessor uses tree-sitter for C# parsing
type TreeSitterProcessor struct {
	parser *sitter.Parser
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() *TreeSitterProcessor {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_c_sharp.Language()))
	
	return &TreeSitterProcessor{
		parser: parser,
	}
}

// ProcessSource processes C# source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	// Parse the source code
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse C# code: %w", err)
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
		Language: "csharp",
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
	case "compilation_unit":
		// Root node - process all children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			p.processNode(child, source, file, parent)
		}
	case "namespace_declaration", "file_scoped_namespace_declaration":
		p.processNamespaceDeclaration(node, source, file, parent)
	case "using_directive", "global_using_directive":
		p.processUsingDirective(node, source, file)
	case "class_declaration":
		p.processClassDeclaration(node, source, file, parent)
	case "interface_declaration":
		p.processInterfaceDeclaration(node, source, file, parent)
	case "enum_declaration":
		p.processEnumDeclaration(node, source, file, parent)
	case "struct_declaration":
		p.processStructDeclaration(node, source, file, parent)
	case "record_declaration", "record_struct_declaration":
		p.processRecordDeclaration(node, source, file, parent)
	case "delegate_declaration":
		p.processDelegateDeclaration(node, source, file, parent)
	case "method_declaration":
		p.processMethodDeclaration(node, source, file, parent)
	case "constructor_declaration":
		p.processConstructorDeclaration(node, source, file, parent)
	case "property_declaration":
		p.processPropertyDeclaration(node, source, file, parent)
	case "field_declaration":
		p.processFieldDeclaration(node, source, file, parent)
	case "event_declaration", "event_field_declaration":
		p.processEventDeclaration(node, source, file, parent)
	default:
		// Recursively process children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			p.processNode(child, source, file, parent)
		}
	}
}

// processNamespaceDeclaration handles namespace declarations
func (p *TreeSitterProcessor) processNamespaceDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	ns := &ir.DistilledPackage{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract namespace name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier", "qualified_name":
			ns.Name = string(source[child.StartByte():child.EndByte()])
		case "declaration_list":
			// Process namespace body
			p.processNamespaceBody(child, source, file, ns)
		}
	}

	// For file-scoped namespaces, continue processing at the same level
	if node.Type() == "file_scoped_namespace_declaration" {
		// Add namespace to file
		file.Children = append(file.Children, ns)
		// Continue processing the rest of the file in this namespace context
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() != "namespace" && child.Type() != "identifier" && child.Type() != "qualified_name" && child.Type() != ";" {
				p.processNode(child, source, file, ns)
			}
		}
	} else {
		p.addToParent(file, parent, ns)
	}
}

// processUsingDirective handles using directives
func (p *TreeSitterProcessor) processUsingDirective(node *sitter.Node, source []byte, file *ir.DistilledFile) {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		ImportType: "using",
		Symbols:    []ir.ImportedSymbol{},
	}

	// Check if it's a global using
	if node.Type() == "global_using_directive" {
		imp.ImportType = "global using"
	}

	// Extract the namespace or type being imported
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier", "qualified_name":
			imp.Module = string(source[child.StartByte():child.EndByte()])
		case "name_equals":
			// Handle using alias (e.g., using Console = System.Console)
			for j := 0; j < int(child.ChildCount()); j++ {
				aliasChild := child.Child(j)
				if aliasChild.Type() == "identifier" {
					alias := string(source[aliasChild.StartByte():aliasChild.EndByte()])
					imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Alias: alias})
					break
				}
			}
		}
	}

	// If no alias, use the last part of the module name
	if len(imp.Symbols) == 0 && imp.Module != "" {
		parts := strings.Split(imp.Module, ".")
		imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: parts[len(parts)-1]})
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

	// Extract modifiers, attributes, name, type parameters, base types
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractClassModifiers(child, source, class)
		case "attribute_list":
			p.extractAttributes(child, source, &class.Decorators)
		case "identifier":
			class.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameter_list":
			p.extractTypeParameters(child, source, class)
		case "base_list":
			p.extractBaseList(child, source, class)
		case "declaration_list":
			p.processClassBody(child, source, file, class)
		}
	}

	// Set default visibility if not specified
	if class.Visibility == "" {
		class.Visibility = ir.VisibilityInternal // C# default for types
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

	// Extract modifiers, attributes, name, type parameters, base interfaces
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractInterfaceModifiers(child, source, iface)
		case "attribute_list":
			// TODO: Add decorator support to interfaces in IR
		case "identifier":
			iface.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameter_list":
			p.extractInterfaceTypeParameters(child, source, iface)
		case "base_list":
			p.extractInterfaceBaseList(child, source, iface)
		case "declaration_list":
			p.processInterfaceBody(child, source, file, iface)
		}
	}

	// Set default visibility if not specified
	if iface.Visibility == "" {
		iface.Visibility = ir.VisibilityInternal // C# default for types
	}

	p.addToParent(file, parent, iface)
}

// processStructDeclaration handles struct declarations
func (p *TreeSitterProcessor) processStructDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	strct := &ir.DistilledStruct{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract modifiers, attributes, name, type parameters, interfaces
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractStructModifiers(child, source, strct)
		case "attribute_list":
			// TODO: Add decorator support to structs in IR
		case "identifier":
			strct.Name = string(source[child.StartByte():child.EndByte()])
		case "type_parameter_list":
			p.extractStructTypeParameters(child, source, strct)
		case "base_list":
			// Structs can only implement interfaces
			p.extractStructInterfaces(child, source, strct)
		case "declaration_list":
			p.processStructBody(child, source, file, strct)
		}
	}

	// Set default visibility if not specified
	if strct.Visibility == "" {
		strct.Visibility = ir.VisibilityInternal // C# default for types
	}

	p.addToParent(file, parent, strct)
}

// processEnumDeclaration handles enum declarations
func (p *TreeSitterProcessor) processEnumDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	enum := &ir.DistilledEnum{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Children: []ir.DistilledNode{},
	}

	// Extract modifiers, attributes, name, base type
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractEnumModifiers(child, source, enum)
		case "attribute_list":
			// TODO: Add decorator support to enums in IR
		case "identifier":
			enum.Name = string(source[child.StartByte():child.EndByte()])
		case "enum_base_clause":
			// Extract underlying type (e.g., : byte)
			for j := 0; j < int(child.ChildCount()); j++ {
				baseChild := child.Child(j)
				if baseChild.Type() == "predefined_type" || baseChild.Type() == "identifier" {
					enum.Type = &ir.TypeRef{Name: string(source[baseChild.StartByte():baseChild.EndByte()])}
					break
				}
			}
		case "enum_member_declaration_list":
			p.processEnumBody(child, source, file, enum)
		}
	}

	// Set default visibility if not specified
	if enum.Visibility == "" {
		enum.Visibility = ir.VisibilityInternal // C# default for types
	}

	p.addToParent(file, parent, enum)
}

// processRecordDeclaration handles record declarations (C# 9+)
func (p *TreeSitterProcessor) processRecordDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Records are classes or structs with special compiler-generated members
	isStruct := node.Type() == "record_struct_declaration"
	
	var record ir.DistilledNode
	if isStruct {
		record = &ir.DistilledStruct{
			BaseNode: ir.BaseNode{
				Location: p.nodeLocation(node),
			},
			Children: []ir.DistilledNode{},
		}
	} else {
		record = &ir.DistilledClass{
			BaseNode: ir.BaseNode{
				Location: p.nodeLocation(node),
			},
			Modifiers: []ir.Modifier{ir.ModifierData}, // Mark as record
			Children:  []ir.DistilledNode{},
		}
	}

	// Extract modifiers, attributes, name, parameters, base types
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			if isStruct {
				p.extractStructModifiers(child, source, record.(*ir.DistilledStruct))
			} else {
				p.extractClassModifiers(child, source, record.(*ir.DistilledClass))
			}
		case "attribute_list":
			// Extract attributes
			var decorators *[]string
			if isStruct {
				// TODO: Add decorator support to structs
			} else {
				decorators = &record.(*ir.DistilledClass).Decorators
			}
			if decorators != nil {
				p.extractAttributes(child, source, decorators)
			}
		case "identifier":
			if isStruct {
				record.(*ir.DistilledStruct).Name = string(source[child.StartByte():child.EndByte()])
			} else {
				record.(*ir.DistilledClass).Name = string(source[child.StartByte():child.EndByte()])
			}
		case "parameter_list":
			// Record parameters become properties
			p.extractRecordParameters(child, source, record)
		case "base_list":
			if isStruct {
				p.extractStructInterfaces(child, source, record.(*ir.DistilledStruct))
			} else {
				p.extractBaseList(child, source, record.(*ir.DistilledClass))
			}
		case "declaration_list":
			if isStruct {
				p.processStructBody(child, source, file, record.(*ir.DistilledStruct))
			} else {
				p.processClassBody(child, source, file, record.(*ir.DistilledClass))
			}
		}
	}

	// Set default visibility if not specified
	if isStruct {
		if record.(*ir.DistilledStruct).Visibility == "" {
			record.(*ir.DistilledStruct).Visibility = ir.VisibilityInternal
		}
	} else {
		if record.(*ir.DistilledClass).Visibility == "" {
			record.(*ir.DistilledClass).Visibility = ir.VisibilityInternal
		}
	}

	p.addToParent(file, parent, record)
}

// processDelegateDeclaration handles delegate declarations
func (p *TreeSitterProcessor) processDelegateDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Delegates are type-safe function pointers
	// We'll represent them as a special type alias
	delegate := &ir.DistilledTypeAlias{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
	}

	// Extract modifiers, return type, name, parameters
	var returnType string
	var parameters []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			text := string(source[child.StartByte():child.EndByte()])
			switch text {
			case "public":
				delegate.Visibility = ir.VisibilityPublic
			case "private":
				delegate.Visibility = ir.VisibilityPrivate
			case "protected":
				delegate.Visibility = ir.VisibilityProtected
			case "internal":
				delegate.Visibility = ir.VisibilityInternal
			}
		case "void_keyword":
			returnType = "void"
		case "predefined_type", "identifier", "generic_name", "qualified_name":
			if returnType == "" && delegate.Name == "" {
				returnType = string(source[child.StartByte():child.EndByte()])
			} else if delegate.Name == "" {
				delegate.Name = string(source[child.StartByte():child.EndByte()])
			}
		case "parameter_list":
			// Extract parameter types
			p.extractDelegateParameters(child, source, &parameters)
		}
	}

	// Build delegate type representation
	paramStr := strings.Join(parameters, ", ")
	delegate.Type = ir.TypeRef{
		Name: fmt.Sprintf("delegate %s(%s)", returnType, paramStr),
	}

	// Set default visibility if not specified
	if delegate.Visibility == "" {
		delegate.Visibility = ir.VisibilityInternal
	}

	p.addToParent(file, parent, delegate)
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

	// Extract modifiers, attributes, return type, name, parameters, body
	// We need to be careful about the order - return type comes before method name
	hasSeenType := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractMethodModifiers(child, source, method)
		case "attribute_list":
			p.extractAttributes(child, source, &method.Decorators)
		case "void_keyword":
			method.Returns = &ir.TypeRef{Name: "void"}
			hasSeenType = true
		case "predefined_type", "array_type", "generic_name", "qualified_name":
			if method.Returns == nil {
				method.Returns = p.extractType(child, source)
				hasSeenType = true
			}
		case "identifier":
			// In C#, return type comes before method name
			// So first identifier is return type (unless we already have one)
			// Second identifier is method name
			if !hasSeenType && method.Returns == nil {
				// This is the return type
				method.Returns = p.extractType(child, source)
				hasSeenType = true
			} else if method.Name == "" {
				// This is the method name
				method.Name = string(source[child.StartByte():child.EndByte()])
			}
		case "type_parameter_list":
			// TODO: Handle generic methods
		case "parameter_list":
			p.extractParameters(child, source, method)
		case "block", "arrow_expression_clause":
			method.Implementation = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Check if this is an extension method
	if len(method.Parameters) > 0 && method.Modifiers != nil {
		for _, mod := range method.Modifiers {
			if mod == ir.ModifierStatic {
				// Check first parameter for "this" modifier
				// TODO: Properly detect extension methods
				break
			}
		}
	}

	// Set default visibility if not specified
	if method.Visibility == "" {
		if parent != nil {
			switch parent.(type) {
			case *ir.DistilledInterface:
				method.Visibility = ir.VisibilityPublic // Interface members are public by default
			default:
				method.Visibility = ir.VisibilityPrivate // Class/struct members are private by default
			}
		} else {
			method.Visibility = ir.VisibilityPrivate
		}
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
	}

	// Extract modifiers, attributes, name, parameters, initializer, body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractMethodModifiers(child, source, constructor)
		case "attribute_list":
			p.extractAttributes(child, source, &constructor.Decorators)
		case "identifier":
			constructor.Name = string(source[child.StartByte():child.EndByte()])
		case "parameter_list":
			p.extractParameters(child, source, constructor)
		case "constructor_initializer":
			// TODO: Handle base() and this() calls
		case "block", "arrow_expression_clause":
			constructor.Implementation = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Set default visibility if not specified
	if constructor.Visibility == "" {
		constructor.Visibility = ir.VisibilityPrivate // Default for constructors
	}

	p.addToParent(file, parent, constructor)
}

// processPropertyDeclaration handles property declarations
func (p *TreeSitterProcessor) processPropertyDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Properties are represented as fields with accessor information
	property := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Modifiers: []ir.Modifier{},
	}
	

	// Track accessor information
	hasGetter := false
	hasSetter := false
	hasInit := false

	// Extract modifiers, attributes, type, name, accessors
	hasSeenType := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractFieldModifiers(child, source, property)
		case "attribute_list":
			p.extractAttributes(child, source, &property.Decorators)
		case "predefined_type", "array_type", "generic_name", "qualified_name", "nullable_type":
			if property.Type == nil {
				property.Type = p.extractType(child, source)
				hasSeenType = true
			}
		case "identifier":
			// Property type comes before property name
			if !hasSeenType && property.Type == nil {
				// This is the property type
				property.Type = p.extractType(child, source)
				hasSeenType = true
			} else if property.Name == "" {
				// This is the property name
				property.Name = string(source[child.StartByte():child.EndByte()])
			}
		case "accessor_list":
			// Parse get/set/init accessors
			for j := 0; j < int(child.ChildCount()); j++ {
				accessor := child.Child(j)
				if accessor.Type() == "accessor_declaration" {
					for k := 0; k < int(accessor.ChildCount()); k++ {
						accChild := accessor.Child(k)
						text := string(source[accChild.StartByte():accChild.EndByte()])
						switch text {
						case "get":
							hasGetter = true
						case "set":
							hasSetter = true
						case "init":
							hasInit = true
						}
					}
				}
			}
		case "arrow_expression_clause":
			// Expression-bodied property (implies getter only)
			hasGetter = true
			property.DefaultValue = string(source[child.StartByte():child.EndByte()])
		case "equals_value_clause":
			// Property initializer
			for j := 0; j < int(child.ChildCount()); j++ {
				valueChild := child.Child(j)
				if valueChild.Type() != "=" {
					property.DefaultValue = string(source[valueChild.StartByte():valueChild.EndByte()])
					break
				}
			}
		}
	}

	// Add modifiers to indicate property accessors
	if hasGetter {
		property.Modifiers = append(property.Modifiers, ir.ModifierReadonly)
	}
	if hasInit && !hasSetter {
		// Init-only property - mark as readonly since it can only be set during initialization
		if !hasGetter {
			property.Modifiers = append(property.Modifiers, ir.ModifierReadonly)
		}
	}

	// Set default visibility if not specified
	if property.Visibility == "" {
		property.Visibility = ir.VisibilityPrivate // Default for properties
	}
	

	p.addToParent(file, parent, property)
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
		case "modifier":
			mod, vis := p.extractFieldModifier(child, source)
			if mod != "" {
				modifiers = append(modifiers, mod)
			}
			if vis != "" {
				visibility = vis
			}
		case "attribute_list":
			p.extractAttributes(child, source, &decorators)
		case "predefined_type", "array_type", "generic_name", "qualified_name", "identifier":
			if fieldType == nil {
				fieldType = p.extractType(child, source)
			}
		case "variable_declaration":
			// Process variable declarators
			for j := 0; j < int(child.ChildCount()); j++ {
				varChild := child.Child(j)
				if varChild.Type() == "variable_declarator" {
					field := &ir.DistilledField{
						BaseNode: ir.BaseNode{
							Location: p.nodeLocation(varChild),
						},
						Type:       fieldType,
						Modifiers:  modifiers,
						Visibility: visibility,
						Decorators: decorators,
					}

					// Extract field name and value
					for k := 0; k < int(varChild.ChildCount()); k++ {
						declChild := varChild.Child(k)
						switch declChild.Type() {
						case "identifier":
							field.Name = string(source[declChild.StartByte():declChild.EndByte()])
						case "equals_value_clause":
							// Extract initializer
							for l := 0; l < int(declChild.ChildCount()); l++ {
								valueChild := declChild.Child(l)
								if valueChild.Type() != "=" {
									field.DefaultValue = string(source[valueChild.StartByte():valueChild.EndByte()])
									break
								}
							}
						}
					}

					// Set default visibility if not specified
					if field.Visibility == "" {
						field.Visibility = ir.VisibilityPrivate
					}

					p.addToParent(file, parent, field)
				}
			}
		}
	}
}

// processEventDeclaration handles event declarations
func (p *TreeSitterProcessor) processEventDeclaration(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Events are special fields with restricted access
	// We'll represent them as fields with a special modifier
	event := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Modifiers: []ir.Modifier{}, // TODO: Add event modifier to IR
	}

	// Extract modifiers, attributes, type, name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "modifier":
			p.extractFieldModifiers(child, source, event)
		case "attribute_list":
			p.extractAttributes(child, source, &event.Decorators)
		case "predefined_type", "generic_name", "qualified_name", "identifier":
			if event.Type == nil && event.Name == "" {
				event.Type = p.extractType(child, source)
			}
		case "variable_declaration":
			// Extract event name(s)
			for j := 0; j < int(child.ChildCount()); j++ {
				varChild := child.Child(j)
				if varChild.Type() == "variable_declarator" {
					for k := 0; k < int(varChild.ChildCount()); k++ {
						declChild := varChild.Child(k)
						if declChild.Type() == "identifier" {
							event.Name = string(source[declChild.StartByte():declChild.EndByte()])
							break
						}
					}
				}
			}
		case "accessor_list":
			// Custom event accessors (add/remove)
			// TODO: Handle custom accessors
		}
	}

	// Set default visibility if not specified
	if event.Visibility == "" {
		event.Visibility = ir.VisibilityPrivate
	}

	p.addToParent(file, parent, event)
}

// Helper methods

// extractClassModifiers extracts modifiers for classes
func (p *TreeSitterProcessor) extractClassModifiers(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		class.Visibility = ir.VisibilityPublic
	case "private":
		class.Visibility = ir.VisibilityPrivate
	case "protected":
		class.Visibility = ir.VisibilityProtected
	case "internal":
		class.Visibility = ir.VisibilityInternal
	case "abstract":
		class.Modifiers = append(class.Modifiers, ir.ModifierAbstract)
	case "sealed":
		class.Modifiers = append(class.Modifiers, ir.ModifierSealed)
	case "static":
		class.Modifiers = append(class.Modifiers, ir.ModifierStatic)
	case "partial":
		class.Modifiers = append(class.Modifiers, ir.ModifierPartial)
	}
}

// extractInterfaceModifiers extracts modifiers for interfaces
func (p *TreeSitterProcessor) extractInterfaceModifiers(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		iface.Visibility = ir.VisibilityPublic
	case "private":
		iface.Visibility = ir.VisibilityPrivate
	case "protected":
		iface.Visibility = ir.VisibilityProtected
	case "internal":
		iface.Visibility = ir.VisibilityInternal
	}
}

// extractStructModifiers extracts modifiers for structs
func (p *TreeSitterProcessor) extractStructModifiers(node *sitter.Node, source []byte, strct *ir.DistilledStruct) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		strct.Visibility = ir.VisibilityPublic
	case "private":
		strct.Visibility = ir.VisibilityPrivate
	case "protected":
		strct.Visibility = ir.VisibilityProtected
	case "internal":
		strct.Visibility = ir.VisibilityInternal
	case "readonly":
		// TODO: Add readonly struct support
	case "ref":
		// TODO: Add ref struct support
	}
}

// extractEnumModifiers extracts modifiers for enums
func (p *TreeSitterProcessor) extractEnumModifiers(node *sitter.Node, source []byte, enum *ir.DistilledEnum) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		enum.Visibility = ir.VisibilityPublic
	case "private":
		enum.Visibility = ir.VisibilityPrivate
	case "protected":
		enum.Visibility = ir.VisibilityProtected
	case "internal":
		enum.Visibility = ir.VisibilityInternal
	}
}

// extractMethodModifiers extracts modifiers for methods
func (p *TreeSitterProcessor) extractMethodModifiers(node *sitter.Node, source []byte, method *ir.DistilledFunction) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		method.Visibility = ir.VisibilityPublic
	case "private":
		method.Visibility = ir.VisibilityPrivate
	case "protected":
		method.Visibility = ir.VisibilityProtected
	case "internal":
		method.Visibility = ir.VisibilityInternal
	case "abstract":
		method.Modifiers = append(method.Modifiers, ir.ModifierAbstract)
	case "virtual":
		method.Modifiers = append(method.Modifiers, ir.ModifierVirtual)
	case "override":
		method.Modifiers = append(method.Modifiers, ir.ModifierOverride)
	case "sealed":
		method.Modifiers = append(method.Modifiers, ir.ModifierSealed)
	case "static":
		method.Modifiers = append(method.Modifiers, ir.ModifierStatic)
	case "async":
		method.Modifiers = append(method.Modifiers, ir.ModifierAsync)
	case "extern":
		method.Modifiers = append(method.Modifiers, ir.ModifierExtern)
	case "partial":
		method.Modifiers = append(method.Modifiers, ir.ModifierPartial)
	case "new":
		// TODO: Handle new modifier (hides inherited member)
	}
}

// extractFieldModifiers extracts modifiers for fields
func (p *TreeSitterProcessor) extractFieldModifiers(node *sitter.Node, source []byte, field *ir.DistilledField) {
	mod, vis := p.extractFieldModifier(node, source)
	if mod != "" {
		field.Modifiers = append(field.Modifiers, mod)
	}
	if vis != "" {
		field.Visibility = vis
	}
}

// extractFieldModifier extracts a single field modifier
func (p *TreeSitterProcessor) extractFieldModifier(node *sitter.Node, source []byte) (ir.Modifier, ir.Visibility) {
	text := string(source[node.StartByte():node.EndByte()])
	
	switch text {
	case "public":
		return "", ir.VisibilityPublic
	case "private":
		return "", ir.VisibilityPrivate
	case "protected":
		return "", ir.VisibilityProtected
	case "internal":
		return "", ir.VisibilityInternal
	case "static":
		return ir.ModifierStatic, ""
	case "readonly":
		return ir.ModifierReadonly, ""
	case "const":
		return ir.ModifierConst, ""
	case "volatile":
		return ir.ModifierVolatile, ""
	case "new":
		// TODO: Handle new modifier
		return "", ""
	default:
		return "", ""
	}
}

// extractAttributes extracts attributes (decorators in IR terms)
func (p *TreeSitterProcessor) extractAttributes(node *sitter.Node, source []byte, decorators *[]string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "attribute" {
			// Extract the full attribute text
			attrText := string(source[child.StartByte():child.EndByte()])
			*decorators = append(*decorators, "[" + attrText + "]")
		}
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

// extractStructTypeParameters extracts generic type parameters for structs
func (p *TreeSitterProcessor) extractStructTypeParameters(node *sitter.Node, source []byte, strct *ir.DistilledStruct) {
	// TODO: Extract and store type parameters
}

// extractBaseList extracts base class and interfaces
func (p *TreeSitterProcessor) extractBaseList(node *sitter.Node, source []byte, class *ir.DistilledClass) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == ":" {
			continue
		}
		
		// First base type is usually the base class (unless it's an interface)
		baseType := p.extractType(child, source)
		if len(class.Extends) == 0 {
			// Assume first is base class (C# only allows single inheritance)
			class.Extends = append(class.Extends, *baseType)
		} else {
			// Rest are interfaces
			class.Implements = append(class.Implements, *baseType)
		}
		
		// TODO: Properly distinguish between base class and interfaces
	}
}

// extractInterfaceBaseList extracts base interfaces
func (p *TreeSitterProcessor) extractInterfaceBaseList(node *sitter.Node, source []byte, iface *ir.DistilledInterface) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == ":" || child.Type() == "," {
			continue
		}
		
		baseType := p.extractType(child, source)
		iface.Extends = append(iface.Extends, *baseType)
	}
}

// extractStructInterfaces extracts implemented interfaces
func (p *TreeSitterProcessor) extractStructInterfaces(node *sitter.Node, source []byte, strct *ir.DistilledStruct) {
	// Structs can only implement interfaces, not inherit from base types
	// TODO: Add interface implementation support to structs in IR
}

// extractRecordParameters extracts record parameters as properties
func (p *TreeSitterProcessor) extractRecordParameters(node *sitter.Node, source []byte, record ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter" {
			// Create a property for each record parameter
			property := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.nodeLocation(child),
				},
				Visibility: ir.VisibilityPublic, // Record properties are public by default
				Modifiers:  []ir.Modifier{ir.ModifierReadonly}, // Record properties are readonly
			}

			// Extract parameter type and name
			for j := 0; j < int(child.ChildCount()); j++ {
				paramChild := child.Child(j)
				switch paramChild.Type() {
				case "predefined_type", "array_type", "generic_name", "qualified_name":
					if property.Type == nil {
						property.Type = p.extractType(paramChild, source)
					}
				case "identifier":
					if property.Name == "" {
						property.Name = string(source[paramChild.StartByte():paramChild.EndByte()])
					} else if property.Type == nil {
						// This might be the type
						property.Type = p.extractType(paramChild, source)
					}
				}
			}

			// Add property to record
			switch r := record.(type) {
			case *ir.DistilledClass:
				r.Children = append(r.Children, property)
			case *ir.DistilledStruct:
				r.Children = append(r.Children, property)
			}
		}
	}
}

// extractType extracts type information
func (p *TreeSitterProcessor) extractType(node *sitter.Node, source []byte) *ir.TypeRef {
	nodeType := node.Type()
	
	switch nodeType {
	case "predefined_type", "identifier", "qualified_name":
		return &ir.TypeRef{Name: string(source[node.StartByte():node.EndByte()])}
	case "generic_name":
		// Handle generic types like List<string>
		typeName := ""
		var generics []ir.TypeRef

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			switch child.Type() {
			case "identifier":
				typeName = string(source[child.StartByte():child.EndByte()])
			case "type_argument_list":
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
		var ranks []string

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "array_rank_specifier" {
				// Count dimensions
				ranks = append(ranks, string(source[child.StartByte():child.EndByte()]))
			} else if elementType == nil {
				elementType = p.extractType(child, source)
			}
		}

		if elementType != nil {
			// Append array notation to type name
			arrayType := elementType.Name
			for _, rank := range ranks {
				arrayType += rank
			}
			return &ir.TypeRef{Name: arrayType}
		}
	case "nullable_type":
		// Handle nullable types (e.g., int?, string?)
		// The nullable_type node contains the base type as first child and "?" as second
		if node.ChildCount() > 0 {
			// Get the base type (first child)
			baseType := p.extractType(node.Child(0), source)
			baseType.IsNullable = true
			return baseType
		}
		// Fallback: extract type name without the ?
		typeName := string(source[node.StartByte():node.EndByte()])
		typeName = strings.TrimSuffix(typeName, "?")
		return &ir.TypeRef{Name: typeName, IsNullable: true}
	}

	// Default: return the raw text
	return &ir.TypeRef{Name: string(source[node.StartByte():node.EndByte()])}
}

// extractParameters extracts method parameters
func (p *TreeSitterProcessor) extractParameters(node *sitter.Node, source []byte, method *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter" {
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
		case "attribute_list":
			// Handle parameter attributes
		case "modifier":
			// Handle parameter modifiers (ref, out, in, this, params)
			text := string(source[child.StartByte():child.EndByte()])
			switch text {
			case "ref", "out", "in":
				// TODO: Add parameter modifier support to IR
			case "this":
				// This indicates an extension method
			case "params":
				param.IsVariadic = true
			}
		case "predefined_type", "array_type", "generic_name", "qualified_name":
			if param.Type.Name == "" {
				param.Type = *p.extractType(child, source)
			}
		case "identifier":
			if param.Name == "" {
				param.Name = string(source[child.StartByte():child.EndByte()])
			} else if param.Type.Name == "" {
				// This might be the parameter type
				param.Type = *p.extractType(child, source)
			}
		case "equals_value_clause":
			// Extract default value
			for j := 0; j < int(child.ChildCount()); j++ {
				valueChild := child.Child(j)
				if valueChild.Type() != "=" {
					param.DefaultValue = string(source[valueChild.StartByte():valueChild.EndByte()])
					param.IsOptional = true
					break
				}
			}
		}
	}

	return param
}

// extractDelegateParameters extracts delegate parameter types
func (p *TreeSitterProcessor) extractDelegateParameters(node *sitter.Node, source []byte, parameters *[]string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter" {
			// Extract parameter type
			for j := 0; j < int(child.ChildCount()); j++ {
				paramChild := child.Child(j)
				switch paramChild.Type() {
				case "predefined_type", "array_type", "generic_name", "qualified_name", "identifier":
					typeRef := p.extractType(paramChild, source)
					*parameters = append(*parameters, typeRef.Name)
					break
				}
			}
		}
	}
}

// processNamespaceBody processes the body of a namespace
func (p *TreeSitterProcessor) processNamespaceBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
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

// processStructBody processes the body of a struct
func (p *TreeSitterProcessor) processStructBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
	}
}

// processEnumBody processes the body of an enum
func (p *TreeSitterProcessor) processEnumBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "enum_member_declaration" {
			p.processEnumMember(child, source, file, parent)
		}
	}
}

// processEnumMember processes enum members
func (p *TreeSitterProcessor) processEnumMember(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierConst},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "attribute_list":
			p.extractAttributes(child, source, &field.Decorators)
		case "identifier":
			field.Name = string(source[child.StartByte():child.EndByte()])
		case "equals_value_clause":
			// Extract enum value
			for j := 0; j < int(child.ChildCount()); j++ {
				valueChild := child.Child(j)
				if valueChild.Type() != "=" {
					field.DefaultValue = string(source[valueChild.StartByte():valueChild.EndByte()])
					break
				}
			}
		}
	}

	// Enum members have the enum type
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
		case *ir.DistilledStruct:
			p.Children = append(p.Children, child)
		case *ir.DistilledEnum:
			p.Children = append(p.Children, child)
		case *ir.DistilledPackage:
			p.Children = append(p.Children, child)
		default:
			file.Children = append(file.Children, child)
		}
	} else {
		file.Children = append(file.Children, child)
	}
}