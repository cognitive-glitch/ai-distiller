package ruby

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

// Processor handles Ruby source code processing
type Processor struct {
	processor.BaseProcessor
}

// NewProcessor creates a new Ruby processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"ruby",
			"1.0.0",
			[]string{".rb", ".rake", ".gemspec"},
		),
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Use tree-sitter parser
	parser := NewTreeSitterProcessor()
	return parser.ProcessSource(ctx, source, filename)
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Create tree-sitter parser instance
	parser := NewTreeSitterProcessor()
	file, err := parser.ProcessSource(ctx, source, filename)
	if err != nil {
		return nil, err
	}
	
	// Apply stripper if any options are set
	stripperOpts := opts.ToStripperOptions()
	
	// Only strip if there's something to strip
	if stripperOpts.HasAnyOption() {
		s := stripper.New(stripperOpts)
		stripped := file.Accept(s)
		if strippedFile, ok := stripped.(*ir.DistilledFile); ok {
			return strippedFile, nil
		}
	}
	
	return file, nil
}