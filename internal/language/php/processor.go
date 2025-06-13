package php

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Processor implements the LanguageProcessor interface for PHP
type Processor struct {
	processor.BaseProcessor
	useTreeSitter bool
}

// NewProcessor creates a new PHP language processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"php",
			"1.0.0",
			[]string{".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".phps", ".inc"},
		),
		useTreeSitter: true, // Default to tree-sitter for PHP
	}
}

// EnableTreeSitter enables tree-sitter based parsing
func (p *Processor) EnableTreeSitter() {
	p.useTreeSitter = true
}

// Process parses PHP source code and returns the IR representation
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Default options
	opts := processor.DefaultProcessOptions()
	return p.ProcessWithOptions(ctx, reader, filename, opts)
}

// ProcessFile processes a file by path
func (p *Processor) ProcessFile(filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read the actual file
	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	ctx := context.Background()
	
	// Use tree-sitter parser
	if p.useTreeSitter {
		processor, err := NewTreeSitterProcessor()
		if err == nil {
			defer processor.Close()
			file, err := processor.ProcessSource(ctx, source, filename)
			if err == nil {
				// Apply options to filter the result
				return p.applyOptions(file, opts), nil
			}
			// Fall back to line-based parser on error
		}
	}
	
	// Fall back to line-based parser
	return p.parseLineBasedPHP(ctx, source, filename, opts)
}

// ProcessWithOptions parses with specific options
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Use tree-sitter parser
	if p.useTreeSitter {
		processor, err := NewTreeSitterProcessor()
		if err == nil {
			defer processor.Close()
			file, err := processor.ProcessSource(ctx, source, filename)
			if err == nil {
				// Apply options to filter the result
				return p.applyOptions(file, opts), nil
			}
			// Fall back to line-based parser on error
		}
	}

	// Fall back to line-based parser
	return p.parseLineBasedPHP(ctx, source, filename, opts)
}

// parseLineBasedPHP provides a simple line-based PHP parser as fallback
func (p *Processor) parseLineBasedPHP(ctx context.Context, source []byte, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// TODO: Implement basic line-based PHP parser
	// For now, return a minimal structure
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   1,
			},
		},
		Path:     filename,
		Language: "php",
		Version:  "8.x",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	return file, nil
}

// applyOptions filters the IR based on processing options
func (p *Processor) applyOptions(file *ir.DistilledFile, opts processor.ProcessOptions) *ir.DistilledFile {
	if !opts.IncludeComments || !opts.IncludeImports || !opts.IncludePrivate || !opts.IncludeImplementation {
		// Create a visitor to filter nodes
		filterVisitor := ir.NewFuncVisitor(func(node ir.DistilledNode) ir.DistilledNode {
			switch n := node.(type) {
			case *ir.DistilledComment:
				if !opts.IncludeComments {
					return nil
				}
			case *ir.DistilledImport:
				if !opts.IncludeImports {
					return nil
				}
			case *ir.DistilledFunction:
				if !opts.IncludePrivate && isPrivate(n.Name) {
					return nil
				}
				if !opts.IncludeImplementation {
					// Clear implementation
					n.Implementation = ""
				}
			case *ir.DistilledClass:
				if !opts.IncludePrivate && isPrivate(n.Name) {
					return nil
				}
			}
			return node
		})

		// Apply filter
		walker := ir.NewWalker(filterVisitor)
		if filtered := walker.Walk(file); filtered != nil {
			return filtered.(*ir.DistilledFile)
		}
	}

	return file
}

func isPrivate(name string) bool {
	// In PHP, names starting with _ are conventionally private
	// But actual visibility is determined by access modifiers
	return len(name) > 0 && name[0] == '_'
}