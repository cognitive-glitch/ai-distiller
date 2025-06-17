package ruby

import (
	"context"
	"fmt"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
	sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
)

// TreeSitterProcessor uses tree-sitter for Ruby parsing
type TreeSitterProcessor struct {
	parser *sitter.Parser
}

// NewTreeSitterProcessor creates a new tree-sitter based processor
func NewTreeSitterProcessor() *TreeSitterProcessor {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tree_sitter_ruby.Language()))

	return &TreeSitterProcessor{
		parser: parser,
	}
}

// ProcessSource processes Ruby source code using tree-sitter
func (p *TreeSitterProcessor) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	// Parse the source code
	tree, err := p.parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ruby code: %w", err)
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
		Language: "ruby",
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
	case "class":
		p.processClass(node, source, file, parent)
	case "singleton_class":
		p.processSingletonClass(node, source, file, parent)
	case "module":
		p.processModule(node, source, file, parent)
	case "method", "singleton_method":
		p.processMethod(node, source, file, parent)
	case "assignment":
		p.processAssignment(node, source, file, parent)
	case "call":
		p.processCall(node, source, file, parent)
	case "alias":
		p.processAlias(node, source, file, parent)
	case "require", "require_relative", "load":
		p.processRequire(node, source, file)
	default:
		// Recursively process children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			p.processNode(child, source, file, parent)
		}
	}
}

// processClass handles class definitions
func (p *TreeSitterProcessor) processClass(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Visibility: ir.VisibilityPublic, // Classes are public by default in Ruby
		Children:   []ir.DistilledNode{},
	}

	// Extract class name and superclass
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "constant":
			class.Name = string(source[child.StartByte():child.EndByte()])
		case "superclass":
			// Find the constant within superclass
			for j := 0; j < int(child.ChildCount()); j++ {
				superChild := child.Child(j)
				if superChild.Type() == "constant" || superChild.Type() == "scope_resolution" {
					superclassName := string(source[superChild.StartByte():superChild.EndByte()])
					class.Extends = append(class.Extends, ir.TypeRef{Name: superclassName})
					break
				}
			}
		case "body_statement", "statements":
			// Process class body
			p.processClassBody(child, source, file, class)
		}
	}

	p.addToParent(file, parent, class)
}

// processSingletonClass handles class << self blocks
func (p *TreeSitterProcessor) processSingletonClass(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Create a special class to represent singleton methods
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Name:       "<<self>>",
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{ir.ModifierStatic}, // Mark as static to indicate singleton
		Children:   []ir.DistilledNode{},
	}

	// Process singleton class body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "body_statement" || child.Type() == "statements" {
			p.processClassBody(child, source, file, class)
		}
	}

	p.addToParent(file, parent, class)
}

// processModule handles module definitions
func (p *TreeSitterProcessor) processModule(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Modules are represented as interfaces in the IR
	module := &ir.DistilledInterface{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Visibility: ir.VisibilityPublic,
		Children:   []ir.DistilledNode{},
	}

	// Extract module name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "constant", "scope_resolution":
			module.Name = string(source[child.StartByte():child.EndByte()])
		case "body_statement", "statements":
			// Process module body
			p.processModuleBody(child, source, file, module)
		}
	}

	p.addToParent(file, parent, module)
}

// processMethod handles method definitions
func (p *TreeSitterProcessor) processMethod(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Parameters: []ir.Parameter{},
		Modifiers:  []ir.Modifier{},
	}

	// Determine visibility (default is public for methods)
	fn.Visibility = p.getCurrentVisibility(parent)

	// Check if it's a singleton method (class method)
	if node.Type() == "singleton_method" {
		fn.Modifiers = append(fn.Modifiers, ir.ModifierStatic)
	}

	// Extract method components
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier", "constant", "operator":
			fn.Name = string(source[child.StartByte():child.EndByte()])
		case "method_parameters", "parameters":
			p.extractParameters(child, source, fn)
		case "body_statement", "statements":
			// Extract method body if needed
			fn.Implementation = string(source[child.StartByte():child.EndByte()])
		}
	}

	// Handle special method names
	if fn.Name == "initialize" {
		fn.Name = "initialize" // Constructor
	}

	p.addToParent(file, parent, fn)
}

// processAssignment handles constant and class variable assignments
func (p *TreeSitterProcessor) processAssignment(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Look for constant assignments
	var name string
	var isConstant bool
	var isClassVar bool
	var isInstanceVar bool

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "constant":
			name = string(source[child.StartByte():child.EndByte()])
			isConstant = true
		case "class_variable":
			name = string(source[child.StartByte():child.EndByte()])
			isClassVar = true
		case "instance_variable":
			name = string(source[child.StartByte():child.EndByte()])
			isInstanceVar = true
		}
	}

	if name != "" && (isConstant || isClassVar || isInstanceVar) {
		field := &ir.DistilledField{
			BaseNode: ir.BaseNode{
				Location: p.nodeLocation(node),
			},
			Name:       name,
			Visibility: ir.VisibilityPublic,
			Modifiers:  []ir.Modifier{},
		}

		if isConstant {
			field.Modifiers = append(field.Modifiers, ir.ModifierFinal)
		}
		if isClassVar {
			field.Modifiers = append(field.Modifiers, ir.ModifierStatic)
		}

		p.addToParent(file, parent, field)
	}
}

// processCall handles method calls that might be DSL methods
func (p *TreeSitterProcessor) processCall(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Look for common metaprogramming patterns
	var methodName string
	var arguments []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			methodName = string(source[child.StartByte():child.EndByte()])
		} else if child.Type() == "argument_list" {
			// Extract arguments
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)
				if arg.Type() == "symbol" || arg.Type() == "string" {
					arguments = append(arguments, string(source[arg.StartByte():arg.EndByte()]))
				}
			}
		}
	}

	// Handle common metaprogramming methods
	switch methodName {
	case "attr_reader", "attr_writer", "attr_accessor":
		p.processAttrMethods(methodName, arguments, node, file, parent)
	case "include", "prepend", "extend":
		p.processModuleInclusion(methodName, arguments, node, file, parent)
	case "private", "protected", "public":
		// These affect visibility of subsequent methods
		// This is simplified - real Ruby is more complex
	case "alias_method":
		if len(arguments) >= 2 {
			p.processAliasMethod(arguments[0], arguments[1], node, file, parent)
		}
	}
}

// processAttrMethods handles attr_reader, attr_writer, attr_accessor
func (p *TreeSitterProcessor) processAttrMethods(methodName string, arguments []string, node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	for _, arg := range arguments {
		// Remove quotes and colons
		attrName := strings.Trim(arg, "\":' ")

		// Create getter
		if methodName == "attr_reader" || methodName == "attr_accessor" {
			getter := &ir.DistilledFunction{
				BaseNode: ir.BaseNode{
					Location: p.nodeLocation(node),
				},
				Name:       attrName,
				Visibility: ir.VisibilityPublic,
				Returns:    &ir.TypeRef{Name: "Object"},
			}
			p.addToParent(file, parent, getter)
		}

		// Create setter
		if methodName == "attr_writer" || methodName == "attr_accessor" {
			setter := &ir.DistilledFunction{
				BaseNode: ir.BaseNode{
					Location: p.nodeLocation(node),
				},
				Name:       attrName + "=",
				Visibility: ir.VisibilityPublic,
				Parameters: []ir.Parameter{
					{Name: "value", Type: ir.TypeRef{Name: "Object"}},
				},
			}
			p.addToParent(file, parent, setter)
		}

		// Also create the instance variable
		field := &ir.DistilledField{
			BaseNode: ir.BaseNode{
				Location: p.nodeLocation(node),
			},
			Name:       "@" + attrName,
			Visibility: ir.VisibilityPrivate,
			Type:       &ir.TypeRef{Name: "Object"},
		}
		p.addToParent(file, parent, field)
	}
}

// processModuleInclusion handles include, prepend, extend
func (p *TreeSitterProcessor) processModuleInclusion(methodName string, arguments []string, node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// In Ruby, include/prepend add module methods as instance methods
	// extend adds module methods as class methods
	// For simplicity, we'll add them to implements
	if class, ok := parent.(*ir.DistilledClass); ok {
		for _, arg := range arguments {
			moduleName := strings.Trim(arg, "\"' ")
			class.Implements = append(class.Implements, ir.TypeRef{Name: moduleName})
		}
	}
}

// processAlias handles alias statements
func (p *TreeSitterProcessor) processAlias(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	var newName, oldName string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "symbol" {
			text := string(source[child.StartByte():child.EndByte()])
			if newName == "" {
				newName = strings.TrimPrefix(text, ":")
			} else if oldName == "" {
				oldName = strings.TrimPrefix(text, ":")
			}
		}
	}

	if newName != "" && oldName != "" {
		p.processAliasMethod(newName, oldName, node, file, parent)
	}
}

// processAliasMethod creates an alias method
func (p *TreeSitterProcessor) processAliasMethod(newName, oldName string, node *sitter.Node, file *ir.DistilledFile, parent ir.DistilledNode) {
	// Create a function that represents the alias
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		Name:           strings.TrimPrefix(newName, ":"),
		Visibility:     ir.VisibilityPublic,
		Implementation: fmt.Sprintf("alias for %s", oldName),
	}
	p.addToParent(file, parent, fn)
}

// processRequire handles require/require_relative/load statements
func (p *TreeSitterProcessor) processRequire(node *sitter.Node, source []byte, file *ir.DistilledFile) {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: p.nodeLocation(node),
		},
		ImportType: node.Type(),
		Symbols:    []ir.ImportedSymbol{},
	}

	// Find the required file
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "argument_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)
				if arg.Type() == "string" {
					// Extract the string content
					content := string(source[arg.StartByte():arg.EndByte()])
					// Remove quotes
					content = strings.Trim(content, "\"'")
					imp.Module = content
					imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
						Name: content,
					})
					break
				}
			}
		}
	}

	file.Children = append(file.Children, imp)
}

// processClassBody processes the body of a class
func (p *TreeSitterProcessor) processClassBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
	}
}

// processModuleBody processes the body of a module
func (p *TreeSitterProcessor) processModuleBody(node *sitter.Node, source []byte, file *ir.DistilledFile, parent ir.DistilledNode) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.processNode(child, source, file, parent)
	}
}

// Helper methods

// nodeLocation converts tree-sitter node position to IR location
func (p *TreeSitterProcessor) nodeLocation(node *sitter.Node) ir.Location {
	return ir.Location{
		StartLine:   int(node.StartPoint().Row) + 1,
		StartColumn: int(node.StartPoint().Column) + 1,
		EndLine:     int(node.EndPoint().Row) + 1,
		EndColumn:   int(node.EndPoint().Column) + 1,
	}
}

// extractParameters extracts method parameters
func (p *TreeSitterProcessor) extractParameters(node *sitter.Node, source []byte, fn *ir.DistilledFunction) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			// Simple parameter
			param := ir.Parameter{
				Name: string(source[child.StartByte():child.EndByte()]),
				Type: ir.TypeRef{Name: "Object"}, // Ruby is dynamically typed
			}
			fn.Parameters = append(fn.Parameters, param)
		case "optional_parameter":
			// Parameter with default value
			var paramName string
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" && paramName == "" {
					paramName = string(source[subChild.StartByte():subChild.EndByte()])
				}
			}
			if paramName != "" {
				param := ir.Parameter{
					Name:         paramName,
					Type:         ir.TypeRef{Name: "Object"},
					DefaultValue: "...", // Simplified
					IsOptional:   true,
				}
				fn.Parameters = append(fn.Parameters, param)
			}
		case "keyword_parameter":
			// Keyword parameter (name:)
			var paramName string
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					paramName = string(source[subChild.StartByte():subChild.EndByte()])
					break
				}
			}
			if paramName != "" {
				param := ir.Parameter{
					Name: paramName + ":",
					Type: ir.TypeRef{Name: "Object"},
				}
				fn.Parameters = append(fn.Parameters, param)
			}
		case "splat_parameter":
			// *args
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					param := ir.Parameter{
						Name:       "*" + string(source[subChild.StartByte():subChild.EndByte()]),
						Type:       ir.TypeRef{Name: "Array"},
						IsVariadic: true,
					}
					fn.Parameters = append(fn.Parameters, param)
					break
				}
			}
		case "hash_splat_parameter":
			// **kwargs
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					param := ir.Parameter{
						Name: "**" + string(source[subChild.StartByte():subChild.EndByte()]),
						Type: ir.TypeRef{Name: "Hash"},
					}
					fn.Parameters = append(fn.Parameters, param)
					break
				}
			}
		case "block_parameter":
			// &block
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" {
					param := ir.Parameter{
						Name: "&" + string(source[subChild.StartByte():subChild.EndByte()]),
						Type: ir.TypeRef{Name: "Proc"},
					}
					fn.Parameters = append(fn.Parameters, param)
					break
				}
			}
		}
	}
}

// getCurrentVisibility determines the current visibility context
func (p *TreeSitterProcessor) getCurrentVisibility(parent ir.DistilledNode) ir.Visibility {
	// In Ruby, methods are public by default
	// This is a simplified implementation
	return ir.VisibilityPublic
}

// addToParent adds a node to its parent or to the file
func (p *TreeSitterProcessor) addToParent(file *ir.DistilledFile, parent ir.DistilledNode, child ir.DistilledNode) {
	if parent != nil {
		switch p := parent.(type) {
		case *ir.DistilledClass:
			p.Children = append(p.Children, child)
		case *ir.DistilledInterface:
			p.Children = append(p.Children, child)
		default:
			file.Children = append(file.Children, child)
		}
	} else {
		file.Children = append(file.Children, child)
	}
}
