package cpp

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

// Processor handles C++ source code processing
type Processor struct {
	processor.BaseProcessor
	tsProcessor *TreeSitterProcessor // Reuse parser instance
}

// NewProcessor creates a new C++ processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"cpp",
			"1.0.0",
			[]string{".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp", ".hh", ".hxx", ".h++"},
		),
		tsProcessor: NewTreeSitterProcessor(), // Initialize once
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Use the shared tree-sitter parser instance
	return p.tsProcessor.ProcessSource(ctx, source, filename)
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Use the shared tree-sitter parser instance
	file, err := p.tsProcessor.ProcessSource(ctx, source, filename)
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