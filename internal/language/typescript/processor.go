package typescript

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Processor handles TypeScript source code processing
type Processor struct {
	processor.BaseProcessor
	useTreeSitter bool
}

// NewProcessor creates a new TypeScript processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"typescript",
			"1.0.0",
			[]string{".ts", ".tsx", ".d.ts"},
		),
		useTreeSitter: true, // Default to tree-sitter for TypeScript
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Read source
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Determine if we need TSX parser
	isTSX := false
	for _, ext := range []string{".tsx"} {
		if len(filename) > len(ext) && filename[len(filename)-len(ext):] == ext {
			isTSX = true
			break
		}
	}
	
	// Use enhanced parser
	parser := NewEnhancedParser()
	return parser.ProcessSource(ctx, source, filename, isTSX)
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