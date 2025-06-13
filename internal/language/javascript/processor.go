package javascript

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Processor handles JavaScript source code processing
type Processor struct {
	processor.BaseProcessor
	useTreeSitter bool
}

// NewProcessor creates a new JavaScript processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"javascript",
			"1.0.0",
			[]string{".js", ".mjs", ".cjs", ".jsx"},
		),
		useTreeSitter: true, // Default to tree-sitter for JavaScript
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Read source
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Try tree-sitter first if enabled
	if p.useTreeSitter {
		tsProcessor, err := NewTreeSitterProcessor()
		if err == nil {
			result, err := tsProcessor.ProcessSource(ctx, source, filename)
			if err == nil {
				return result, nil
			}
			// Fall through to line-based parser on error
		}
	}

	// Fallback to simple line-based parser
	// TODO: Implement basic line-based parser as fallback
	return &ir.DistilledFile{
		Path:     filename,
		Language: "javascript",
		Version:  "ES2022",
		Children: []ir.DistilledNode{},
		Errors: []ir.DistilledError{
			{
				BaseNode: ir.BaseNode{
					Location: ir.Location{StartLine: 1},
				},
				Message:  "Line-based parser not yet implemented for JavaScript",
				Severity: "warning",
			},
		},
	}, nil
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// For now, just use Process
	// TODO: Implement options handling
	return p.Process(ctx, reader, filename)
}

// EnableTreeSitter enables tree-sitter parsing
func (p *Processor) EnableTreeSitter() {
	p.useTreeSitter = true
}

// DisableTreeSitter disables tree-sitter parsing
func (p *Processor) DisableTreeSitter() {
	p.useTreeSitter = false
}