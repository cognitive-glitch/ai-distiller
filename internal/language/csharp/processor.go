package csharp

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

// Processor handles C# source code processing
type Processor struct {
	processor.BaseProcessor
	tsProcessor *TreeSitterProcessor // Reuse parser instance
}

// NewProcessor creates a new C# processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"csharp",
			"1.0.0",
			[]string{".cs"},
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
	// Process using tree-sitter
	result, err := p.Process(ctx, reader, filename)
	if err != nil {
		return nil, err
	}

	// Apply stripper if any options are set
	stripperOpts := stripper.Options{
		RemovePrivate:         !opts.IncludePrivate && !opts.RemovePrivateOnly && !opts.RemoveProtectedOnly,
		RemovePrivateOnly:     opts.RemovePrivateOnly,
		RemoveProtectedOnly:   opts.RemoveProtectedOnly,
		RemoveImplementations: !opts.IncludeImplementation,
		RemoveComments:        !opts.IncludeComments,
		RemoveImports:         !opts.IncludeImports,
	}

	// Only strip if there's something to strip
	if stripperOpts.RemovePrivate || stripperOpts.RemovePrivateOnly || stripperOpts.RemoveProtectedOnly ||
		stripperOpts.RemoveImplementations || stripperOpts.RemoveComments || stripperOpts.RemoveImports {

		s := stripper.New(stripperOpts)
		stripped := result.Accept(s)
		if strippedFile, ok := stripped.(*ir.DistilledFile); ok {
			return strippedFile, nil
		}
	}

	return result, nil
}