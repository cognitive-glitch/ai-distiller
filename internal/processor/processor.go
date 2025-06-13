package processor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/stripper"
)


// Processor processes files and directories
type Processor struct {
	useTreeSitter bool
}

// New creates a new processor
func New() *Processor {
	return &Processor{
		useTreeSitter: false,
	}
}

// EnableTreeSitter enables tree-sitter parsing for supported languages
func (p *Processor) EnableTreeSitter() {
	p.useTreeSitter = true
}

// ProcessPath processes a file or directory
func (p *Processor) ProcessPath(path string, opts ProcessOptions) (ir.DistilledNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return p.processDirectory(path, opts)
	}
	return p.ProcessFile(path, opts)
}

// ProcessFile processes a single file
func (p *Processor) ProcessFile(filename string, opts ProcessOptions) (*ir.DistilledFile, error) {
	// Get processor for file
	proc, ok := GetByFilename(filename)
	if !ok {
		return nil, fmt.Errorf("no processor found for file: %s", filename)
	}
	
	// Enable tree-sitter if requested and supported
	if p.useTreeSitter && proc.Language() == "python" {
		// Use reflection to call EnableTreeSitter if it exists
		if enabler, ok := proc.(interface{ EnableTreeSitter() }); ok {
			enableTSFunc := enabler.EnableTreeSitter
			enableTSFunc()
		}
	}

	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Process file
	ctx := context.Background()
	result, err := proc.Process(ctx, file, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %w", err)
	}

	// Apply stripping options
	stripOpts := stripper.Options{
		RemovePrivate:         !opts.IncludePrivate,
		RemovePrivateOnly:     opts.RemovePrivateOnly,
		RemoveProtectedOnly:   opts.RemoveProtectedOnly,
		RemoveImplementations: !opts.IncludeImplementation,
		RemoveComments:        !opts.IncludeComments,
		RemoveImports:         !opts.IncludeImports,
	}

	if stripOpts.RemovePrivate || stripOpts.RemovePrivateOnly || stripOpts.RemoveProtectedOnly || 
	   stripOpts.RemoveImplementations || stripOpts.RemoveComments || stripOpts.RemoveImports {
		s := stripper.New(stripOpts)
		strippedNode := result.Accept(s)
		if file, ok := strippedNode.(*ir.DistilledFile); ok {
			return file, nil
		}
		return nil, fmt.Errorf("unexpected node type after stripping")
	}

	return result, nil
}

// Process processes a reader
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Get processor for file
	proc, ok := GetByFilename(filename)
	if !ok {
		return nil, fmt.Errorf("no processor found for file: %s", filename)
	}
	
	// Enable tree-sitter if requested and supported
	if p.useTreeSitter && proc.Language() == "python" {
		// Use reflection to call EnableTreeSitter if it exists
		if enabler, ok := proc.(interface{ EnableTreeSitter() }); ok {
			enableTSFunc := enabler.EnableTreeSitter
			enableTSFunc()
		}
	}

	return proc.Process(ctx, reader, filename)
}

func (p *Processor) processDirectory(dir string, opts ProcessOptions) (*ir.DistilledDirectory, error) {
	result := &ir.DistilledDirectory{
		BaseNode: ir.BaseNode{},
		Path:     dir,
		Children: []ir.DistilledNode{},
	}

	// Walk directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if we have a processor for this file
		if _, ok := GetByFilename(path); !ok {
			return nil
		}

		// Process file
		file, err := p.ProcessFile(path, opts)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		result.Children = append(result.Children, file)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return result, nil
}

// GetSupportedExtensions returns all supported file extensions
func (p *Processor) GetSupportedExtensions() []string {
	var extensions []string
	for _, lang := range List() {
		if proc, ok := Get(lang); ok {
			extensions = append(extensions, proc.SupportedExtensions()...)
		}
	}
	return extensions
}

// CanProcess checks if a file can be processed
func (p *Processor) CanProcess(filename string) bool {
	_, ok := GetByFilename(filename)
	return ok
}

// GetLanguage returns the language for a file
func (p *Processor) GetLanguage(filename string) string {
	proc, ok := GetByFilename(filename)
	if !ok {
		return ""
	}
	return proc.Language()
}

