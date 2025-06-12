package stripper

import (
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Stripper removes specified elements from the IR based on options
type Stripper struct {
	options processor.ProcessOptions
}

// NewStripper creates a new stripper with the given options
func NewStripper(options processor.ProcessOptions) *Stripper {
	return &Stripper{
		options: options,
	}
}

// Strip applies stripping rules to the IR tree
func (s *Stripper) Strip(file *ir.DistilledFile) *ir.DistilledFile {
	// Create a visitor that filters nodes based on options
	visitor := ir.NewFuncVisitor(func(node ir.DistilledNode) ir.DistilledNode {
		// Skip nodes based on options
		switch n := node.(type) {
		case *ir.DistilledComment:
			if !s.options.IncludeComments {
				return nil // Remove comment nodes
			}

		case *ir.DistilledImport:
			if !s.options.IncludeImports {
				return nil // Remove import nodes
			}

		case *ir.DistilledFunction:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private functions
			}
			
			// Strip implementation if requested
			if !s.options.IncludeImplementation && n.Implementation != "" {
				// Create a copy without implementation
				stripped := *n
				stripped.Implementation = ""
				return &stripped
			}

		case *ir.DistilledClass:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private classes
			}

		case *ir.DistilledInterface:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private interfaces
			}

		case *ir.DistilledStruct:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private structs
			}

		case *ir.DistilledEnum:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private enums
			}

		case *ir.DistilledField:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private fields
			}

		case *ir.DistilledTypeAlias:
			// Check visibility
			if !s.options.IncludePrivate && s.isPrivate(n.Name, n.Visibility) {
				return nil // Remove private type aliases
			}
		}

		return node // Keep the node
	})

	// Apply the visitor
	walker := ir.NewWalker(visitor)
	result := walker.Walk(file)
	
	if result == nil {
		return file // Return original if walk failed
	}
	
	return result.(*ir.DistilledFile)
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
	// Go convention would need language context to apply correctly
	if strings.HasPrefix(name, "_") {
		return true
	}

	// Default to public if no explicit visibility and no underscore
	return false
}

// StripOptions represents what to strip from the code
type StripOptions struct {
	Comments        bool
	Imports         bool
	Implementation  bool
	NonPublic       bool
}

// FromStrings converts string options to StripOptions
func FromStrings(options []string) StripOptions {
	opts := StripOptions{}
	
	for _, opt := range options {
		switch opt {
		case "comments":
			opts.Comments = true
		case "imports":
			opts.Imports = true
		case "implementation":
			opts.Implementation = true
		case "non-public":
			opts.NonPublic = true
		}
	}
	
	return opts
}

// ToProcessOptions converts StripOptions to ProcessOptions
func (s StripOptions) ToProcessOptions() processor.ProcessOptions {
	return processor.ProcessOptions{
		IncludeImplementation: !s.Implementation,
		IncludeComments:       !s.Comments,
		IncludeImports:        !s.Imports,
		IncludePrivate:        !s.NonPublic,
		MaxDepth:              100,
		Strict:                false,
		SymbolResolution:      true,
		IncludeLineNumbers:    true,
	}
}