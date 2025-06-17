package typescript

import (
	"context"
	"fmt"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
	sitter "github.com/smacker/go-tree-sitter"
	typescript "tree-sitter-typescript"
)

// ASTParser provides tree-sitter based TypeScript parsing
type ASTParser struct {
	parser   *sitter.Parser
	source   []byte
	filename string
	isTSX    bool
}

// NewASTParser creates a new tree-sitter TypeScript parser
func NewASTParser() *ASTParser {
	parser := sitter.NewParser()
	// Language will be set in ProcessSource based on file type

	return &ASTParser{
		parser: parser,
	}
}

// ProcessSource processes TypeScript source using tree-sitter
func (p *ASTParser) ProcessSource(ctx context.Context, source []byte, filename string, isTSX bool) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	p.isTSX = isTSX

	// Set the appropriate language based on file type
	if isTSX {
		p.parser.SetLanguage(sitter.NewLanguage(typescript.LanguageTSX()))
	} else {
		p.parser.SetLanguage(sitter.NewLanguage(typescript.Language()))
	}

	// Parse the source code
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TypeScript: %w", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()

	// Create the distilled file
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   int(rootNode.EndPoint().Row) + 1,
			},
		},
		Path:     filename,
		Language: "typescript",
		Version:  "5.0",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}

	// Parse the program
	if rootNode.Type() == "program" {
		p.parseProgram(rootNode, file)
	}

	// FIXME: The current implementation is flawed and produces incorrect results.
	// Disable until it can be made more robust.
	// p.analyzeInterfaceSatisfaction(file)

	return file, nil
}

// parseProgram parses the top-level program node
func (p *ASTParser) parseProgram(node *sitter.Node, file *ir.DistilledFile) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "import_statement":
			if imp := p.parseImport(child); imp != nil {
				file.Children = append(file.Children, imp)
			}
		case "export_statement":
			if nodes := p.parseExport(child); nodes != nil {
				file.Children = append(file.Children, nodes...)
			}
		case "class_declaration", "abstract_class_declaration":
			if class := p.parseClass(child, false); class != nil {
				file.Children = append(file.Children, class)
			}
		case "interface_declaration":
			if intf := p.parseInterface(child, false); intf != nil {
				file.Children = append(file.Children, intf)
			}
		case "type_alias_declaration":
			if alias := p.parseTypeAlias(child, false); alias != nil {
				file.Children = append(file.Children, alias)
			}
		case "enum_declaration":
			if enum := p.parseEnum(child, false); enum != nil {
				file.Children = append(file.Children, enum)
			}
		case "function_declaration":
			if fn := p.parseFunction(child, false); fn != nil {
				file.Children = append(file.Children, fn)
			}
		case "lexical_declaration", "variable_declaration":
			if vars := p.parseVariableDeclaration(child, false); vars != nil {
				file.Children = append(file.Children, vars...)
			}
		}
	}
}

// parseImport parses import statements
func (p *ASTParser) parseImport(node *sitter.Node) *ir.DistilledImport {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		ImportType: "import",
		Symbols:    []ir.ImportedSymbol{},
	}

	// Find the source module
	sourceNode := p.findChild(node, "string")
	if sourceNode != nil {
		moduleText := p.nodeText(sourceNode)
		// Remove quotes
		if len(moduleText) >= 2 {
			imp.Module = moduleText[1 : len(moduleText)-1]
		}
	}

	// Check for type-only import
	if p.findChild(node, "type") != nil {
		imp.IsType = true
	}

	// Parse import clause
	importClause := p.findChild(node, "import_clause")
	if importClause != nil {
		p.parseImportClause(importClause, imp)
	}

	return imp
}

// parseImportClause parses the import clause
func (p *ASTParser) parseImportClause(node *sitter.Node, imp *ir.DistilledImport) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "identifier":
			// Default import
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name: p.nodeText(child),
			})
		case "namespace_import":
			// import * as name
			asClause := p.findChild(child, "identifier")
			if asClause != nil {
				imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
					Name:  "*",
					Alias: p.nodeText(asClause),
				})
			}
		case "named_imports":
			p.parseNamedImports(child, imp)
		}
	}
}

// parseNamedImports parses named imports
func (p *ASTParser) parseNamedImports(node *sitter.Node, imp *ir.DistilledImport) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil || child.Type() != "import_specifier" {
			continue
		}

		name := ""
		alias := ""

		// Find name and alias
		for j := 0; j < int(child.ChildCount()); j++ {
			subchild := child.Child(j)
			if subchild == nil {
				continue
			}

			if subchild.Type() == "identifier" {
				if name == "" {
					name = p.nodeText(subchild)
				} else {
					alias = p.nodeText(subchild)
				}
			}
		}

		if name != "" {
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  name,
				Alias: alias,
			})
		}
	}
}

// parseExport parses export statements
func (p *ASTParser) parseExport(node *sitter.Node) []ir.DistilledNode {
	var nodes []ir.DistilledNode

	// Debug: print export node structure
	// fmt.Printf("Export node has %d children\n", node.ChildCount())
	// for i := 0; i < int(node.ChildCount()); i++ {
	// 	child := node.Child(i)
	// 	if child != nil {
	// 		fmt.Printf("  Child %d: type=%s, text=%q\n", i, child.Type(), p.nodeText(child))
	// 	}
	// }

	// Find the exported declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "class_declaration", "abstract_class_declaration":
			if class := p.parseClass(child, true); class != nil {
				nodes = append(nodes, class)
			}
		case "interface_declaration":
			if intf := p.parseInterface(child, true); intf != nil {
				nodes = append(nodes, intf)
			}
		case "type_alias_declaration":
			if alias := p.parseTypeAlias(child, true); alias != nil {
				nodes = append(nodes, alias)
			}
		case "enum_declaration":
			if enum := p.parseEnum(child, true); enum != nil {
				nodes = append(nodes, enum)
			}
		case "function_declaration":
			if fn := p.parseFunction(child, true); fn != nil {
				nodes = append(nodes, fn)
			}
		case "lexical_declaration", "variable_declaration":
			if vars := p.parseVariableDeclaration(child, true); vars != nil {
				nodes = append(nodes, vars...)
			}
		}
	}

	return nodes
}

// parseClass parses class declarations
func (p *ASTParser) parseClass(node *sitter.Node, isExported bool) *ir.DistilledClass {
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}

	// Parse class name
	nameNode := p.findChild(node, "type_identifier")
	if nameNode != nil {
		class.Name = p.nodeText(nameNode)
	}

	// Parse modifiers
	if p.hasModifier(node, "abstract") {
		class.Modifiers = append(class.Modifiers, ir.ModifierAbstract)
	}
	if isExported {
		class.Modifiers = append(class.Modifiers, ir.ModifierExport)
	}

	// Parse type parameters
	typeParams := p.findChild(node, "type_parameters")
	if typeParams != nil {
		class.TypeParams = p.parseTypeParameters(typeParams)
	}

	// Parse extends clause
	extendsClause := p.findChild(node, "class_heritage")
	if extendsClause != nil {
		for i := 0; i < int(extendsClause.ChildCount()); i++ {
			child := extendsClause.Child(i)
			if child == nil {
				continue
			}

			if child.Type() == "extends_clause" {
				// Look for type_identifier or identifier
				typeNode := p.findChild(child, "type_identifier")
				if typeNode == nil {
					typeNode = p.findChild(child, "identifier")
				}
				if typeNode != nil {
					typeName := p.nodeText(typeNode)

					// Check for type arguments
					typeArgs := p.findChild(child, "type_arguments")
					if typeArgs != nil {
						typeName += p.nodeText(typeArgs)
					}

					class.Extends = append(class.Extends, ir.TypeRef{
						Name: typeName,
					})
				}
			} else if child.Type() == "implements_clause" {
				// Parse implements
				for j := 0; j < int(child.ChildCount()); j++ {
					implChild := child.Child(j)
					if implChild == nil {
						continue
					}

					if implChild.Type() == "type_identifier" {
						class.Implements = append(class.Implements, ir.TypeRef{
							Name: p.nodeText(implChild),
						})
					} else if implChild.Type() == "generic_type" {
						// Handle generic implements like Cacheable<string, User>
						baseType := p.findChild(implChild, "type_identifier")
						if baseType != nil {
							typeName := p.nodeText(baseType)
							typeArgs := p.findChild(implChild, "type_arguments")
							if typeArgs != nil {
								typeName += p.nodeText(typeArgs)
							}
							class.Implements = append(class.Implements, ir.TypeRef{
								Name: typeName,
							})
						}
					}
				}
			}
		}
	}

	// Parse class body
	classBody := p.findChild(node, "class_body")
	if classBody != nil {
		p.parseClassBody(classBody, class)
	}

	return class
}

// parseClassBody parses class body members
func (p *ASTParser) parseClassBody(node *sitter.Node, class *ir.DistilledClass) {
	// First pass: collect all fields including parameter properties
	var fields []ir.DistilledNode
	var methods []ir.DistilledNode
	var constructorMethod *ir.DistilledFunction
	var pendingDecorators []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "decorated_definition":
			// Collect decorators and locate the real definition
			var localDecs []string
			var def *sitter.Node
			for j := 0; j < int(child.ChildCount()); j++ {
				sub := child.Child(j)
				if sub == nil {
					continue
				}
				if sub.Type() == "decorator" {
					if d := p.parseDecorator(sub); d != "" {
						localDecs = append(localDecs, d)
					}
				} else {
					def = sub // first non-decorator child = wrapped definition
				}
			}
			if def != nil {
				switch def.Type() {
				case "method_definition":
					if m := p.parseMethod(def); m != nil {
						m.Decorators = append(m.Decorators, localDecs...)
						if m.Name == "constructor" {
							constructorMethod = m
							for k := 0; k < int(def.ChildCount()); k++ {
								if fc := def.Child(k); fc != nil && fc.Type() == "formal_parameters" {
									for _, f := range p.parseParameterProperties(fc) {
										fields = append(fields, f)
									}
									break
								}
							}
						} else {
							methods = append(methods, m)
						}
					}
				case "public_field_definition", "private_field_definition", "protected_field_definition":
					if f := p.parseField(def); f != nil {
						f.Decorators = append(f.Decorators, localDecs...)
						fields = append(fields, f)
					}
				}
			}
		case "decorator":
			// Collect decorators for the next method/field
			if decoratorText := p.parseDecorator(child); decoratorText != "" {
				pendingDecorators = append(pendingDecorators, decoratorText)
			}
		case "method_definition":
			if method := p.parseMethod(child); method != nil {
				// Attach any pending decorators
				if len(pendingDecorators) > 0 {
					method.Decorators = append(method.Decorators, pendingDecorators...)
					pendingDecorators = nil // Reset for next item
				}

				if method.Name == "constructor" {
					constructorMethod = method
					// Find parameter properties from constructor
					for j := 0; j < int(child.ChildCount()); j++ {
						methodChild := child.Child(j)
						if methodChild != nil && methodChild.Type() == "formal_parameters" {
							paramProps := p.parseParameterProperties(methodChild)
							// Add parameter properties to fields
							for _, field := range paramProps {
								fields = append(fields, field)
							}
							break
						}
					}
				} else {
					methods = append(methods, method)
				}
			}
		case "public_field_definition", "private_field_definition", "protected_field_definition":
			if field := p.parseField(child); field != nil {
				// Attach any pending decorators
				if len(pendingDecorators) > 0 {
					field.Decorators = append(field.Decorators, pendingDecorators...)
					pendingDecorators = nil // Reset for next item
				}
				fields = append(fields, field)
			}
		case "method_signature", "abstract_method_signature":
			if method := p.parseMethodSignature(child); method != nil {
				// Attach any pending decorators
				if len(pendingDecorators) > 0 {
					method.Decorators = append(method.Decorators, pendingDecorators...)
					pendingDecorators = nil // Reset for next item
				}

				// Mark as abstract if it's an abstract_method_signature
				// (and parseMethodSignature didn't already add it)
				if child.Type() == "abstract_method_signature" {
					hasAbstract := false
					for _, mod := range method.Modifiers {
						if mod == ir.ModifierAbstract {
							hasAbstract = true
							break
						}
					}
					if !hasAbstract {
						method.Modifiers = append(method.Modifiers, ir.ModifierAbstract)
					}
				}
				methods = append(methods, method)
			}
		}
	}

	// Add members in the correct order: fields first, then constructor, then methods
	class.Children = append(class.Children, fields...)
	if constructorMethod != nil {
		class.Children = append(class.Children, constructorMethod)
	}
	class.Children = append(class.Children, methods...)
}

// parseMethod parses method definitions
func (p *ASTParser) parseMethod(node *sitter.Node) *ir.DistilledFunction {
	method := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
	}

	// Parse accessibility modifiers and decorators
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "decorator":
			// Process decorators that are children of the method_definition
			if decoratorText := p.parseDecorator(child); decoratorText != "" {
				method.Decorators = append(method.Decorators, decoratorText)
			}
		case "accessibility_modifier":
			switch p.nodeText(child) {
			case "private":
				method.Visibility = ir.VisibilityPrivate
			case "protected":
				method.Visibility = ir.VisibilityProtected
			case "public":
				method.Visibility = ir.VisibilityPublic
			}
		case "static":
			method.Modifiers = append(method.Modifiers, ir.ModifierStatic)
		case "async":
			method.Modifiers = append(method.Modifiers, ir.ModifierAsync)
		case "abstract":
			method.Modifiers = append(method.Modifiers, ir.ModifierAbstract)
		case "readonly":
			method.Modifiers = append(method.Modifiers, ir.ModifierReadonly)
		case "property_name":
			if nameChild := child.Child(0); nameChild != nil {
				if nameChild.Type() == "identifier" || nameChild.Type() == "property_identifier" {
					method.Name = p.nodeText(nameChild)
				}
			}
		case "property_identifier":
			method.Name = p.nodeText(child)
		case "type_parameters":
			method.TypeParams = p.parseTypeParameters(child)
		case "formal_parameters":
			method.Parameters = p.parseParameters(child)
		case "type_annotation":
			// Extract the actual type from the type annotation (skip the colon)
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild != nil && typeChild.Type() != ":" {
					method.Returns = &ir.TypeRef{Name: p.nodeText(typeChild)}
					break
				}
			}
		case "statement_block":
			method.Implementation = p.nodeText(child)
		}
	}

	return method
}

// parseMethodSignature parses method signatures (in interfaces)
func (p *ASTParser) parseMethodSignature(node *sitter.Node) *ir.DistilledFunction {
	method := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Parameters: []ir.Parameter{},
		Modifiers:  []ir.Modifier{},
	}

	// Parse method signature components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "accessibility_modifier":
			switch p.nodeText(child) {
			case "private":
				method.Visibility = ir.VisibilityPrivate
			case "protected":
				method.Visibility = ir.VisibilityProtected
			case "public":
				method.Visibility = ir.VisibilityPublic
			}
		case "abstract":
			method.Modifiers = append(method.Modifiers, ir.ModifierAbstract)
		case "property_name":
			if nameChild := child.Child(0); nameChild != nil {
				if nameChild.Type() == "identifier" || nameChild.Type() == "property_identifier" {
					method.Name = p.nodeText(nameChild)
				}
			}
		case "property_identifier":
			method.Name = p.nodeText(child)
		case "type_parameters":
			method.TypeParams = p.parseTypeParameters(child)
		case "formal_parameters":
			method.Parameters = p.parseParameters(child)
		case "type_annotation":
			// Extract the actual type from the type annotation (skip the colon)
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild != nil && typeChild.Type() != ":" {
					method.Returns = &ir.TypeRef{Name: p.nodeText(typeChild)}
					break
				}
			}
		}
	}

	return method
}

// parseField parses field definitions
func (p *ASTParser) parseField(node *sitter.Node) *ir.DistilledField {
	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
	}

	// Debug
	// fmt.Printf("Parsing field, node type: %s, children: %d\n", node.Type(), node.ChildCount())
	// for i := 0; i < int(node.ChildCount()); i++ {
	// 	child := node.Child(i)
	// 	if child != nil {
	// 		fmt.Printf("  Child %d: type=%s, text=%q\n", i, child.Type(), p.nodeText(child))
	// 	}
	// }

	// Parse field components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "accessibility_modifier":
			switch p.nodeText(child) {
			case "private":
				field.Visibility = ir.VisibilityPrivate
			case "protected":
				field.Visibility = ir.VisibilityProtected
			case "public":
				field.Visibility = ir.VisibilityPublic
			}
		case "static":
			field.Modifiers = append(field.Modifiers, ir.ModifierStatic)
		case "readonly":
			field.Modifiers = append(field.Modifiers, ir.ModifierReadonly)
		case "property_name":
			if nameChild := child.Child(0); nameChild != nil {
				if nameChild.Type() == "identifier" || nameChild.Type() == "property_identifier" {
					field.Name = p.nodeText(nameChild)
				}
			}
		case "property_identifier":
			field.Name = p.nodeText(child)
		case "type_annotation":
			// The type annotation includes the colon, so we need to extract just the type
			// Find the actual type node (skip the colon)
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild != nil && typeChild.Type() != ":" {
					field.Type = &ir.TypeRef{Name: p.nodeText(typeChild)}
					break
				}
			}
		case "identifier": // For simple field declarations
			field.Name = p.nodeText(child)
		}
	}

	// Look for field initializer (e.g., = null)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "=" {
			// Next child should be the initializer value
			if i+1 < int(node.ChildCount()) {
				valueNode := node.Child(i + 1)
				if valueNode != nil {
					field.DefaultValue = p.nodeText(valueNode)
				}
			}
			break
		}
	}

	return field
}

// parseInterface parses interface declarations
func (p *ASTParser) parseInterface(node *sitter.Node, isExported bool) *ir.DistilledInterface {
	intf := &ir.DistilledInterface{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}

	if isExported {
		intf.Modifiers = append(intf.Modifiers, ir.ModifierExport)
	}

	// Debug: print interface AST structure
	// fmt.Printf("Interface AST for node at line %d:\n", node.StartPoint().Row+1)
	// for i := 0; i < int(node.ChildCount()); i++ {
	// 	child := node.Child(i)
	// 	if child != nil {
	// 		fmt.Printf("  Child %d: type=%s, text=%q\n", i, child.Type(), p.nodeText(child))
	// 	}
	// }

	// Parse interface name
	nameNode := p.findChild(node, "type_identifier")
	if nameNode != nil {
		intf.Name = p.nodeText(nameNode)
	}

	// Parse type parameters
	typeParams := p.findChild(node, "type_parameters")
	if typeParams != nil {
		intf.TypeParams = p.parseTypeParameters(typeParams)
	}

	// Parse extends clause
	extendsClause := p.findChild(node, "extends_type_clause")
	if extendsClause != nil {
		// Debug: print what we find in extends clause
		// fmt.Printf("Found extends clause with %d children\n", extendsClause.ChildCount())
		// for i := 0; i < int(extendsClause.ChildCount()); i++ {
		// 	child := extendsClause.Child(i)
		// 	if child != nil {
		// 		fmt.Printf("  Child %d: type=%s, text=%q\n", i, child.Type(), p.nodeText(child))
		// 	}
		// }

		for i := 0; i < int(extendsClause.ChildCount()); i++ {
			child := extendsClause.Child(i)
			if child != nil && (child.Type() == "type_identifier" || child.Type() == "identifier") {
				intf.Extends = append(intf.Extends, ir.TypeRef{
					Name: p.nodeText(child),
				})
			}
		}
	}

	// Parse interface body
	interfaceBody := p.findChild(node, "interface_body")
	if interfaceBody != nil {
		p.parseInterfaceBody(interfaceBody, intf)
	}

	return intf
}

// parseInterfaceBody parses interface body members
func (p *ASTParser) parseInterfaceBody(node *sitter.Node, intf *ir.DistilledInterface) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "property_signature":
			if field := p.parsePropertySignature(child); field != nil {
				intf.Children = append(intf.Children, field)
			}
		case "method_signature":
			if method := p.parseMethodSignature(child); method != nil {
				intf.Children = append(intf.Children, method)
			}
		}
	}
}

// parsePropertySignature parses property signatures
func (p *ASTParser) parsePropertySignature(node *sitter.Node) *ir.DistilledField {
	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
	}

	// Parse property signature components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "readonly":
			field.Modifiers = append(field.Modifiers, ir.ModifierReadonly)
		case "property_name":
			if nameChild := child.Child(0); nameChild != nil {
				if nameChild.Type() == "identifier" || nameChild.Type() == "property_identifier" {
					field.Name = p.nodeText(nameChild)
				}
			}
		case "property_identifier":
			field.Name = p.nodeText(child)
		case "type_annotation":
			// Extract the actual type from the type annotation (skip the colon)
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild != nil && typeChild.Type() != ":" {
					field.Type = &ir.TypeRef{Name: p.nodeText(typeChild)}
					break
				}
			}
		}
	}

	return field
}

// parseTypeAlias parses type alias declarations
func (p *ASTParser) parseTypeAlias(node *sitter.Node, isExported bool) *ir.DistilledTypeAlias {
	alias := &ir.DistilledTypeAlias{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Modifiers:  []ir.Modifier{},
	}

	if isExported {
		alias.Modifiers = append(alias.Modifiers, ir.ModifierExport)
	}

	// Parse type alias name
	nameNode := p.findChild(node, "type_identifier")
	if nameNode != nil {
		alias.Name = p.nodeText(nameNode)
	}

	// Parse type parameters
	typeParams := p.findChild(node, "type_parameters")
	if typeParams != nil {
		alias.TypeParams = p.parseTypeParameters(typeParams)
	}

	// Parse the aliased type by looking for any type-like node after the =
	// The type could be a union_type, intersection_type, or other type node
	foundEquals := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// Skip until we find the equals sign
		if !foundEquals && p.nodeText(child) == "=" {
			foundEquals = true
			continue
		}

		// After equals, the next non-punctuation node is our type
		if foundEquals && child.Type() != "=" {
			alias.Type = ir.TypeRef{Name: p.nodeText(child)}
			break
		}
	}

	return alias
}

// parseEnum parses enum declarations
func (p *ASTParser) parseEnum(node *sitter.Node, isExported bool) *ir.DistilledEnum {
	enum := &ir.DistilledEnum{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Children:   []ir.DistilledNode{},
	}

	// Parse enum name
	nameNode := p.findChild(node, "identifier")
	if nameNode != nil {
		enum.Name = p.nodeText(nameNode)
	}

	// Parse enum body
	enumBody := p.findChild(node, "enum_body")
	if enumBody != nil {
		p.parseEnumBody(enumBody, enum)
	}

	return enum
}

// parseEnumBody parses enum body members
func (p *ASTParser) parseEnumBody(node *sitter.Node, enum *ir.DistilledEnum) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "enum_assignment" {
			// Parse enum assignment like ADMIN = "admin"
			nameNode := p.findChild(child, "property_identifier")
			if nameNode != nil {
				field := &ir.DistilledField{
					BaseNode: ir.BaseNode{
						Location: p.nodeToLocation(child),
					},
					Name:       p.nodeText(nameNode),
					Visibility: ir.VisibilityPublic,
					Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierReadonly},
				}

				// Get the assigned value
				for j := 0; j < int(child.ChildCount()); j++ {
					valueChild := child.Child(j)
					if valueChild != nil && (valueChild.Type() == "string" || valueChild.Type() == "number") {
						field.DefaultValue = p.nodeText(valueChild)
						break
					}
				}

				enum.Children = append(enum.Children, field)
			}
		} else if child.Type() == "property_identifier" {
			// Simple enum member without assignment
			field := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.nodeToLocation(child),
				},
				Name:       p.nodeText(child),
				Visibility: ir.VisibilityPublic,
				Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierReadonly},
			}

			enum.Children = append(enum.Children, field)
		}
	}
}

// parseFunction parses function declarations
func (p *ASTParser) parseFunction(node *sitter.Node, isExported bool) *ir.DistilledFunction {
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
	}

	if isExported {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierExport)
	}

	// Parse function components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "async":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
		case "identifier":
			fn.Name = p.nodeText(child)
			// Check for underscore prefix indicating private
			if strings.HasPrefix(fn.Name, "_") && !isExported {
				fn.Visibility = ir.VisibilityPrivate
			}
		case "type_parameters":
			fn.TypeParams = p.parseTypeParameters(child)
		case "formal_parameters":
			fn.Parameters = p.parseParameters(child)
		case "type_annotation":
			if typeNode := p.findChild(child, "type"); typeNode != nil {
				fn.Returns = &ir.TypeRef{Name: p.nodeText(typeNode)}
			}
		case "statement_block":
			fn.Implementation = p.nodeText(child)
		}
	}

	return fn
}

// parseVariableDeclaration parses variable declarations
func (p *ASTParser) parseVariableDeclaration(node *sitter.Node, isExported bool) []ir.DistilledNode {
	var nodes []ir.DistilledNode

	// Find variable declarators
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil || child.Type() != "variable_declarator" {
			continue
		}

		// Check if this variable declarator contains an arrow function
		arrowFunc := p.findChild(child, "arrow_function")
		if arrowFunc != nil {
			// This is a function, not a field
			fn := p.parseArrowFunction(arrowFunc, isExported)
			if fn != nil {
				// Get the function name from the variable declarator
				nameNode := p.findChild(child, "identifier")
				if nameNode != nil {
					fn.Name = p.nodeText(nameNode)
					// Check for underscore prefix indicating private
					if strings.HasPrefix(fn.Name, "_") && !isExported {
						fn.Visibility = ir.VisibilityPrivate
					}
				}
				// fmt.Printf("Parsed arrow function %s with %d params\n", fn.Name, len(fn.Parameters))

				// Check if it's const (makes it immutable)
				if p.hasToken(node, "const") {
					fn.Modifiers = append(fn.Modifiers, ir.ModifierFinal)
				}

				// Add export modifier if needed
				if isExported {
					fn.Modifiers = append(fn.Modifiers, ir.ModifierExport)
				}

				nodes = append(nodes, fn)
			}
		} else {
			// Regular variable/field
			field := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.nodeToLocation(child),
				},
				Visibility: p.getVisibility(isExported),
				Modifiers:  []ir.Modifier{},
			}

			// Check if it's const
			if p.hasToken(node, "const") {
				field.Modifiers = append(field.Modifiers, ir.ModifierFinal)
			}

			// Add export modifier if needed
			if isExported {
				field.Modifiers = append(field.Modifiers, ir.ModifierExport)
			}

			// Parse variable declarator
			nameNode := p.findChild(child, "identifier")
			if nameNode != nil {
				field.Name = p.nodeText(nameNode)
				// Check for underscore prefix indicating private
				if strings.HasPrefix(field.Name, "_") && !isExported {
					field.Visibility = ir.VisibilityPrivate
				}
			}

			typeAnnotation := p.findChild(child, "type_annotation")
			if typeAnnotation != nil {
				if typeNode := p.findChild(typeAnnotation, "type"); typeNode != nil {
					field.Type = &ir.TypeRef{Name: p.nodeText(typeNode)}
				}
			}

			nodes = append(nodes, field)
		}
	}

	return nodes
}

// parseArrowFunction parses arrow function expressions
func (p *ASTParser) parseArrowFunction(node *sitter.Node, isExported bool) *ir.DistilledFunction {
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeToLocation(node),
		},
		Visibility: p.getVisibility(isExported),
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
	}

	// Debug: print the AST structure
	// fmt.Printf("\nArrow function AST for node at line %d:\n", node.StartPoint().Row+1)
	// for i := 0; i < int(node.ChildCount()); i++ {
	// 	child := node.Child(i)
	// 	if child != nil {
	// 		fmt.Printf("  Child %d: type=%s, text=%q\n", i, child.Type(), p.nodeText(child))
	// 	}
	// }

	// Parse arrow function components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "async":
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
		case "formal_parameters":
			fn.Parameters = p.parseParameters(child)
		case "identifier":
			// Single parameter without parentheses (e.g., x => x * 2)
			if len(fn.Parameters) == 0 && i == 0 {
				fn.Parameters = append(fn.Parameters, ir.Parameter{
					Name: p.nodeText(child),
				})
			}
		case "type_annotation":
			// Return type annotation
			for j := 0; j < int(child.ChildCount()); j++ {
				typeChild := child.Child(j)
				if typeChild != nil && typeChild.Type() != ":" {
					fn.Returns = &ir.TypeRef{Name: p.nodeText(typeChild)}
					break
				}
			}
		case "statement_block":
			fn.Implementation = p.nodeText(child)
		case "=>":
			// Arrow token, skip
			continue
		default:
			// If it's not one of the above and we're past the arrow, it's likely the body
			if fn.Implementation == "" && child.Type() != "=>" {
				// Check if we've seen the arrow already
				hasArrow := false
				for j := 0; j < i; j++ {
					if prev := node.Child(j); prev != nil && prev.Type() == "=>" {
						hasArrow = true
						break
					}
				}
				if hasArrow {
					fn.Implementation = p.nodeText(child)
				}
			}
		}
	}

	return fn
}

// parseParameterProperties parses parameter properties from constructor parameters
func (p *ASTParser) parseParameterProperties(node *sitter.Node) []*ir.DistilledField {
	var fields []*ir.DistilledField

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "required_parameter", "optional_parameter":
			// Check if this parameter has visibility modifiers
			var visibility ir.Visibility
			var hasVisibility bool
			var paramName string
			var paramType *ir.TypeRef

			for j := 0; j < int(child.ChildCount()); j++ {
				subchild := child.Child(j)
				if subchild == nil {
					continue
				}

				switch subchild.Type() {
				case "accessibility_modifier":
					hasVisibility = true
					switch p.nodeText(subchild) {
					case "private":
						visibility = ir.VisibilityPrivate
					case "protected":
						visibility = ir.VisibilityProtected
					case "public":
						visibility = ir.VisibilityPublic
					}
				case "identifier":
					paramName = p.nodeText(subchild)
				case "type_annotation":
					// Get the actual type node
					for k := 0; k < int(subchild.ChildCount()); k++ {
						typeChild := subchild.Child(k)
						if typeChild != nil && typeChild.Type() != ":" {
							paramType = &ir.TypeRef{Name: p.nodeText(typeChild)}
							break
						}
					}
				}
			}

			// If this parameter has a visibility modifier, it's a parameter property
			if hasVisibility && paramName != "" {
				field := &ir.DistilledField{
					BaseNode: ir.BaseNode{
						Location: p.nodeToLocation(child),
					},
					Name:       paramName,
					Visibility: visibility,
					Type:       paramType,
					Modifiers:  []ir.Modifier{},
				}
				fields = append(fields, field)
			}
		}
	}

	return fields
}

// parseParameters parses function parameters
func (p *ASTParser) parseParameters(node *sitter.Node) []ir.Parameter {
	var params []ir.Parameter

	// Debug
	// fmt.Printf("Parsing parameters, node has %d children\n", node.ChildCount())

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		// fmt.Printf("  Param child %d: type=%s\n", i, child.Type())

		switch child.Type() {
		case "required_parameter", "optional_parameter":
			param := ir.Parameter{
				IsOptional: child.Type() == "optional_parameter",
			}

			// Parse parameter components
			for j := 0; j < int(child.ChildCount()); j++ {
				subchild := child.Child(j)
				if subchild == nil {
					continue
				}

				switch subchild.Type() {
				case "identifier":
					param.Name = p.nodeText(subchild)
				case "type_annotation":
					// Get the actual type node, which is the first child after the colon
					for k := 0; k < int(subchild.ChildCount()); k++ {
						typeChild := subchild.Child(k)
						if typeChild != nil && typeChild.Type() != ":" {
							param.Type = ir.TypeRef{Name: p.nodeText(typeChild)}
							break
						}
					}
				}
			}

			params = append(params, param)
		}
	}

	return params
}

// parseTypeParameters parses type parameters
func (p *ASTParser) parseTypeParameters(node *sitter.Node) []ir.TypeParam {
	var typeParams []ir.TypeParam

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil || child.Type() != "type_parameter" {
			continue
		}

		param := ir.TypeParam{}

		// Parse type parameter components
		nameNode := p.findChild(child, "type_identifier")
		if nameNode != nil {
			param.Name = p.nodeText(nameNode)
		}

		constraint := p.findChild(child, "constraint")
		if constraint != nil {
			// The constraint contains "extends" followed by the type
			// We need to extract the type part after "extends"
			constraintText := p.nodeText(constraint)
			if strings.HasPrefix(constraintText, "extends ") {
				constraintType := strings.TrimPrefix(constraintText, "extends ")
				param.Constraints = append(param.Constraints, ir.TypeRef{
					Name: constraintType,
				})
			}
		}

		typeParams = append(typeParams, param)
	}

	return typeParams
}

// parseDecorator parses decorator expressions
func (p *ASTParser) parseDecorator(node *sitter.Node) string {
	// TypeScript decorators are like @LogExecution or @Component({...})
	decoratorText := p.nodeText(node)

	// Remove the leading @ symbol if present
	if strings.HasPrefix(decoratorText, "@") {
		decoratorText = decoratorText[1:]
	}

	return decoratorText
}

// Helper functions

func (p *ASTParser) nodeToLocation(node *sitter.Node) ir.Location {
	return ir.Location{
		StartLine:   int(node.StartPoint().Row) + 1,
		StartColumn: int(node.StartPoint().Column) + 1,
		EndLine:     int(node.EndPoint().Row) + 1,
		EndColumn:   int(node.EndPoint().Column) + 1,
		StartByte:   int(node.StartByte()),
		EndByte:     int(node.EndByte()),
	}
}

func (p *ASTParser) nodeText(node *sitter.Node) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	sourceLen := uint32(len(p.source))
	if start > end || end > sourceLen {
		return ""
	}
	return string(p.source[start:end])
}

func (p *ASTParser) findChild(node *sitter.Node, childType string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == childType {
			return child
		}
	}
	return nil
}

func (p *ASTParser) hasModifier(node *sitter.Node, modifier string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == modifier {
			return true
		}
	}
	return false
}

func (p *ASTParser) hasToken(node *sitter.Node, token string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && p.nodeText(child) == token {
			return true
		}
	}
	return false
}

func (p *ASTParser) getVisibility(isExported bool) ir.Visibility {
	if isExported {
		return ir.VisibilityPublic
	}
	// Non-exported items are module-private (internal)
	// They're accessible within the module but not externally
	return ir.VisibilityInternal
}

// analyzeInterfaceSatisfaction analyzes interface implementations
// This enhances explicit implements clauses with structural type analysis
func (p *ASTParser) analyzeInterfaceSatisfaction(file *ir.DistilledFile) {
	// Collect all interfaces and classes
	interfaces := make(map[string]*ir.DistilledInterface)
	classes := make(map[string]*ir.DistilledClass)

	for _, child := range file.Children {
		switch node := child.(type) {
		case *ir.DistilledInterface:
			interfaces[node.Name] = node
		case *ir.DistilledClass:
			classes[node.Name] = node
		}
	}

	// For each class, check if it structurally satisfies interfaces it doesn't explicitly implement
	for _, class := range classes {
		// Get already explicitly implemented interfaces
		explicitImpls := make(map[string]bool)
		for _, impl := range class.Implements {
			explicitImpls[impl.Name] = true
		}

		// Check each interface for structural compatibility
		for intfName, intf := range interfaces {
			if explicitImpls[intfName] {
				continue // Already explicitly implemented
			}

			if p.classImplementsInterface(class, intf) {
				// Add implicit implementation
				class.Implements = append(class.Implements, ir.TypeRef{
					Name: intfName + " (implicit)",
				})
			}
		}
	}
}

// classImplementsInterface checks if a class structurally implements an interface
func (p *ASTParser) classImplementsInterface(class *ir.DistilledClass, intf *ir.DistilledInterface) bool {
	// Create method and property maps for the class
	// Note: This function is disabled in the current implementation
	// as structural typing detection proved too complex
	_ = class
	_ = intf
	return false
	/*
		classMethods := make(map[string]*ir.DistilledFunction)
		classProps := make(map[string]*ir.DistilledField)

		for _, child := range class.Children {
			switch member := child.(type) {
			case *ir.DistilledFunction:
				classMethods[member.Name] = member
			case *ir.DistilledField:
				classProps[member.Name] = member
			}
		}

		// Check if class satisfies all interface requirements
		for _, child := range intf.Children {
			switch requirement := child.(type) {
			case *ir.DistilledFunction:
				// Check if class has matching method
				classMethod, exists := classMethods[requirement.Name]
				if !exists {
					return false
				}

				// Check method signature compatibility
				if !p.methodsCompatible(classMethod, requirement) {
					return false
				}

			case *ir.DistilledField:
				// Check if class has matching property
				classProp, exists := classProps[requirement.Name]
				if !exists {
					return false
				}

				// Check property type compatibility
				if !p.typesCompatible(classProp.Type, requirement.Type) {
					return false
				}
			}
		}

		return true
	}

	// methodsCompatible checks if two methods are compatible
	func (p *ASTParser) methodsCompatible(classMethod, intfMethod *ir.DistilledFunction) bool {
		// Check parameter count
		if len(classMethod.Parameters) != len(intfMethod.Parameters) {
			return false
		}

		// Check parameter types
		for i, classParam := range classMethod.Parameters {
			intfParam := intfMethod.Parameters[i]
			if !p.typesCompatible(&classParam.Type, &intfParam.Type) {
				return false
			}
		}

		// Check return type
		return p.typesCompatible(classMethod.Returns, intfMethod.Returns)
	*/
}

// typesCompatible checks if two types are compatible (simplified structural check)
func (p *ASTParser) typesCompatible(type1, type2 *ir.TypeRef) bool {
	if type1 == nil && type2 == nil {
		return true
	}
	if type1 == nil || type2 == nil {
		return false
	}

	// Simple string comparison for now
	// In a full implementation, this would be more sophisticated
	return strings.TrimSpace(type1.Name) == strings.TrimSpace(type2.Name)
}
