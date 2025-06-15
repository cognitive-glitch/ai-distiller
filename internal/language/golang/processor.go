package golang

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/debug"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

// Processor handles Go source code processing
type Processor struct {
	processor.BaseProcessor
}

// NewProcessor creates a new Go processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"go",
			"1.0.0",
			[]string{".go"},
		),
	}
}

// Process implements processor.LanguageProcessor
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Read source
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// Use AST parser
	parser := NewASTParser()
	return parser.ProcessSource(ctx, source, filename)
}

// ProcessWithOptions implements processor.LanguageProcessor
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	dbg := debug.FromContext(ctx).WithSubsystem("golang")
	defer dbg.Timing(debug.LevelDetailed, "ProcessWithOptions")()
	
	dbg.Logf(debug.LevelDetailed, "Processing %s with Go AST parser", filename)
	
	// First, process without stripping
	result, err := p.Process(ctx, reader, filename)
	if err != nil {
		return nil, err
	}
	
	// Dump raw IR at trace level
	debug.Lazy(ctx, debug.LevelTrace, func(d debug.Debugger) {
		d.Dump(debug.LevelTrace, "Raw Go IR before stripping", result)
	})
	
	// Apply stripping if any options are set
	stripperOpts := opts.ToStripperOptions()
	
	// Only strip if there's something to strip
	if stripperOpts.HasAnyOption() {
		dbg.Logf(debug.LevelDetailed, "Applying stripper with options: %+v", stripperOpts)
		
		s := stripper.New(stripperOpts)
		stripped := result.Accept(s)
		if strippedFile, ok := stripped.(*ir.DistilledFile); ok {
			// Dump stripped IR at trace level
			debug.Lazy(ctx, debug.LevelTrace, func(d debug.Debugger) {
				d.Dump(debug.LevelTrace, "Stripped Go IR", strippedFile)
			})
			return strippedFile, nil
		}
	}
	
	return result, nil
}