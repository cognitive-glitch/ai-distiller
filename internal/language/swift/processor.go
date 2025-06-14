package swift

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

// Processor handles Swift source code processing
type Processor struct {
	processor.BaseProcessor
}

// NewProcessor creates a new Swift processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"swift",
			"1.0.0",
			[]string{".swift"},
		),
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Use ProcessWithOptions with default options
	return p.ProcessWithOptions(ctx, reader, filename, processor.ProcessOptions{
		IncludePrivate:        true,
		IncludeImplementation: true,
		IncludeComments:       true,
		IncludeImports:        true,
	})
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Try tree-sitter parser first
	treeparser, err := NewTreeSitterProcessor()
	if err == nil {
		defer treeparser.Close()
		file, err := treeparser.ProcessSource(ctx, source, filename)
		if err != nil {
			// Fall back to line-based parser on error
			fmt.Fprintf(os.Stderr, "warning: Swift tree-sitter for %s failed with error: %v. Falling back to line parser.\n", filename, err)
			parser := NewLineParser(source, filename)
			return parser.Parse(), nil
		}
		
		// Apply standardized stripper for filtering
		stripperOpts := opts.ToStripperOptions()
		
		// Only apply stripper if we need to remove something
		if stripperOpts.HasAnyOption() {
			s := stripper.New(stripperOpts)
			stripped := file.Accept(s)
			return stripped.(*ir.DistilledFile), nil
		}
		
		return file, nil
	}

	// Fall back to line-based parser
	parser := NewLineParser(source, filename)
	return parser.Parse(), nil
}