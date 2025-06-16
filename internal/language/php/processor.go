package php

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
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
		treeparser, err := NewTreeSitterProcessor()
		if err != nil {
			// Tree-sitter processor creation failed
			fmt.Fprintf(os.Stderr, "PHP: Failed to create tree-sitter processor: %v\n", err)
			return p.parseLineBasedPHP(ctx, source, filename, opts)
		}
		defer treeparser.Close()
		file, err := treeparser.ProcessSource(ctx, source, filename)
		if err != nil {
			// Fall back to line-based parser on error
			fmt.Fprintf(os.Stderr, "PHP: Tree-sitter parse failed: %v\n", err)
			return p.parseLineBasedPHP(ctx, source, filename, opts)
		}
		
		// Apply standardized stripper for filtering
		stripperOpts := opts.ToStripperOptions()
		
		// Debug
		// fmt.Printf("DEBUG: IncludeComments=%v, RemoveComments=%v\n", opts.IncludeComments, stripperOpts.RemoveComments)
		
		// Only apply stripper if we need to remove something
		if stripperOpts.HasAnyOption() {
			s := stripper.New(stripperOpts)
			stripped := file.Accept(s)
			return stripped.(*ir.DistilledFile), nil
		}
		
		return file, nil
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
		treeparser, err := NewTreeSitterProcessor()
		if err != nil {
			// Tree-sitter processor creation failed
			fmt.Fprintf(os.Stderr, "PHP: Failed to create tree-sitter processor: %v\n", err)
			return p.parseLineBasedPHP(ctx, source, filename, opts)
		}
		defer treeparser.Close()
		file, err := treeparser.ProcessSource(ctx, source, filename)
		if err != nil {
			// Fall back to line-based parser on error
			fmt.Fprintf(os.Stderr, "PHP: Tree-sitter parse failed: %v\n", err)
			return p.parseLineBasedPHP(ctx, source, filename, opts)
		}
		
		// Apply standardized stripper for filtering
		stripperOpts := opts.ToStripperOptions()
		
		// Debug
		// fmt.Printf("DEBUG: IncludeComments=%v, RemoveComments=%v\n", opts.IncludeComments, stripperOpts.RemoveComments)
		
		// Only apply stripper if we need to remove something
		if stripperOpts.HasAnyOption() {
			s := stripper.New(stripperOpts)
			stripped := file.Accept(s)
			return stripped.(*ir.DistilledFile), nil
		}
		
		return file, nil
	}

	// Fall back to line-based parser
	return p.parseLineBasedPHP(ctx, source, filename, opts)
}

// parseLineBasedPHP provides a simple line-based PHP parser as fallback
func (p *Processor) parseLineBasedPHP(ctx context.Context, source []byte, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Create a minimal file structure with an error
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   len(strings.Split(string(source), "\n")),
			},
		},
		Path:     filename,
		Language: "php",
		Version:  "8.x",
		Children: []ir.DistilledNode{},
		Errors: []ir.DistilledError{
			{
				BaseNode: ir.BaseNode{
					Location: ir.Location{
						StartLine: 1,
						EndLine:   1,
					},
				},
				Message:  "PHP parser fallback: tree-sitter parser not available",
				Severity: "warning",
			},
		},
	}
	
	return file, nil
}

