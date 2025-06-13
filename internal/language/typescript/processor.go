package typescript

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
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
	
	// Use tree-sitter AST parser
	parser := NewASTParser()
	return parser.ProcessSource(ctx, source, filename, isTSX)
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// First, process without stripping
	result, err := p.Process(ctx, reader, filename)
	if err != nil {
		return nil, err
	}
	
	// Apply stripping if any options are set
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

// EnableTreeSitter enables tree-sitter parsing
func (p *Processor) EnableTreeSitter() {
	p.useTreeSitter = true
}

// DisableTreeSitter disables tree-sitter parsing
func (p *Processor) DisableTreeSitter() {
	p.useTreeSitter = false
}

// ProcessFile processes a file by path
func (p *Processor) ProcessFile(filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.ProcessWithOptions(context.Background(), file, filename, opts)
}