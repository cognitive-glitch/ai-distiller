package stripper

import (
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// Options configures what to strip from the IR
type Options struct {
	RemovePrivate         bool
	RemoveImplementations bool
	RemoveComments        bool
	RemoveImports         bool
}

// Stripper removes specified elements from the IR based on options
type Stripper struct {
	options Options
}

// New creates a new stripper with the given options
func New(options Options) *Stripper {
	return &Stripper{
		options: options,
	}
}

// Visit implements the Visitor interface
func (s *Stripper) Visit(node ir.DistilledNode) ir.DistilledNode {
	if node == nil {
		return nil
	}

	// Process different node types
	switch n := node.(type) {
	case *ir.DistilledFile:
		return s.visitFile(n)
		
	case *ir.DistilledComment:
		if s.options.RemoveComments {
			return nil
		}
		return n
		
	case *ir.DistilledImport:
		if s.options.RemoveImports {
			return nil
		}
		return n
		
	case *ir.DistilledFunction:
		return s.visitFunction(n)
		
	case *ir.DistilledClass:
		return s.visitClass(n)
		
	case *ir.DistilledInterface:
		return s.visitInterface(n)
		
	case *ir.DistilledStruct:
		return s.visitStruct(n)
		
	case *ir.DistilledEnum:
		return s.visitEnum(n)
		
	case *ir.DistilledField:
		return s.visitField(n)
		
	case *ir.DistilledTypeAlias:
		return s.visitTypeAlias(n)
		
	default:
		// For other nodes, visit children
		return s.visitChildren(node)
	}
}

func (s *Stripper) visitFile(n *ir.DistilledFile) *ir.DistilledFile {
	result := &ir.DistilledFile{
		BaseNode: n.BaseNode,
		Path:     n.Path,
		Language: n.Language,
		Version:  n.Version,
		Metadata: n.Metadata,
		Errors:   n.Errors,
	}
	
	// Visit children
	for _, child := range n.Children {
		if visited := child.Accept(s); visited != nil {
			result.Children = append(result.Children, visited)
		}
	}
	
	return result
}

func (s *Stripper) visitFunction(n *ir.DistilledFunction) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Create copy
	result := &ir.DistilledFunction{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
		Modifiers:  n.Modifiers,
		Decorators: n.Decorators,
		TypeParams: n.TypeParams,
		Parameters: n.Parameters,
		Returns:    n.Returns,
		Implementation: n.Implementation,
	}
	
	// Strip implementation if requested
	if s.options.RemoveImplementations {
		result.Implementation = ""
	}
	
	// Functions don't have children in our IR
	
	return result
}

func (s *Stripper) visitClass(n *ir.DistilledClass) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Create copy
	result := &ir.DistilledClass{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
		Modifiers:  n.Modifiers,
		Decorators: n.Decorators,
		TypeParams: n.TypeParams,
		Extends:    n.Extends,
		Implements: n.Implements,
		Mixins:     n.Mixins,
	}
	
	// Visit children
	for _, child := range n.Children {
		if visited := child.Accept(s); visited != nil {
			result.Children = append(result.Children, visited)
		}
	}
	
	return result
}

func (s *Stripper) visitInterface(n *ir.DistilledInterface) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Create copy
	result := &ir.DistilledInterface{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
		TypeParams: n.TypeParams,
		Extends:    n.Extends,
	}
	
	// Visit children
	for _, child := range n.Children {
		if visited := child.Accept(s); visited != nil {
			result.Children = append(result.Children, visited)
		}
	}
	
	return result
}

func (s *Stripper) visitStruct(n *ir.DistilledStruct) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Create copy
	result := &ir.DistilledStruct{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
		TypeParams: n.TypeParams,
	}
	
	// Visit children
	for _, child := range n.Children {
		if visited := child.Accept(s); visited != nil {
			result.Children = append(result.Children, visited)
		}
	}
	
	return result
}

func (s *Stripper) visitEnum(n *ir.DistilledEnum) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Create copy
	result := &ir.DistilledEnum{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
	}
	
	// Visit children
	for _, child := range n.Children {
		if visited := child.Accept(s); visited != nil {
			result.Children = append(result.Children, visited)
		}
	}
	
	return result
}

func (s *Stripper) visitField(n *ir.DistilledField) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Return copy
	return &ir.DistilledField{
		BaseNode:     n.BaseNode,
		Name:         n.Name,
		Visibility:   n.Visibility,
		Modifiers:    n.Modifiers,
		Type:         n.Type,
		DefaultValue: n.DefaultValue,
		Decorators:   n.Decorators,
	}
}

func (s *Stripper) visitTypeAlias(n *ir.DistilledTypeAlias) ir.DistilledNode {
	// Check if should remove private
	if s.options.RemovePrivate && s.isPrivate(n.Name, n.Visibility) {
		return nil
	}
	
	// Return copy
	return &ir.DistilledTypeAlias{
		BaseNode:   n.BaseNode,
		Name:       n.Name,
		Visibility: n.Visibility,
		TypeParams: n.TypeParams,
		Type:       n.Type,
	}
}

func (s *Stripper) visitChildren(node ir.DistilledNode) ir.DistilledNode {
	// For unknown node types, just return as-is
	// In a real implementation, we'd need to handle all node types
	return node
}

// isPrivate checks if a node should be considered private
func (s *Stripper) isPrivate(name string, visibility ir.Visibility) bool {
	// Check explicit visibility first
	switch visibility {
	case ir.VisibilityPrivate, ir.VisibilityInternal, ir.VisibilityFilePrivate:
		return true
	case ir.VisibilityPublic, ir.VisibilityOpen:
		return false
	}

	// If no explicit visibility, use language conventions
	// For now, we only check Python convention (underscore prefix)
	if strings.HasPrefix(name, "_") {
		return true
	}

	// Default to public if no explicit visibility and no underscore
	return false
}