package rust

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Processor handles Rust source code processing
type Processor struct {
	processor.BaseProcessor
	useTreeSitter bool
}

// NewProcessor creates a new Rust processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"rust",
			"1.0.0",
			[]string{".rs"},
		),
		useTreeSitter: false, // Disabled until WASM/native issues are resolved
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
		// Tree-sitter support is temporarily disabled due to integration issues
		// TODO: Fix WASM processor or native tree-sitter integration
	}

	// Use line-based parser
	parser := NewLineParser(source, filename)
	return parser.Parse(), nil
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