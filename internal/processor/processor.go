package processor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/janreges/ai-distiller/internal/debug"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/stripper"
)


// Processor processes files and directories
type Processor struct {
	useTreeSitter bool
	ctx          context.Context
}

// New creates a new processor
func New() *Processor {
	return &Processor{
		useTreeSitter: false,
		ctx:          context.Background(),
	}
}

// NewWithContext creates a new processor with a context
func NewWithContext(ctx context.Context) *Processor {
	return &Processor{
		useTreeSitter: false,
		ctx:          ctx,
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
	dbg := debug.FromContext(p.ctx).WithSubsystem("processor")
	defer dbg.Timing(debug.LevelDetailed, fmt.Sprintf("ProcessFile %s", filepath.Base(filename)))()
	
	dbg.Logf(debug.LevelDetailed, "Processing file: %s", filename)
	
	// Calculate display path based on options
	displayPath := filename
	if opts.FilePathType == "relative" && opts.BasePath != "" {
		// Check if basePath is a file or directory
		baseInfo, err := os.Stat(opts.BasePath)
		var baseDir string
		if err == nil && !baseInfo.IsDir() {
			// If basePath is a file, use its directory
			baseDir = filepath.Dir(opts.BasePath)
		} else {
			// Otherwise use basePath as is
			baseDir = opts.BasePath
		}
		
		// Try to make path relative to base directory
		absBase, err := filepath.Abs(baseDir)
		if err == nil {
			relPath, err := filepath.Rel(absBase, filename)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				displayPath = relPath
				
				// Apply prefix if specified
				if opts.RelativePathPrefix != "" {
					prefix := opts.RelativePathPrefix
					// Ensure prefix ends with separator if not empty and doesn't already
					if !strings.HasSuffix(prefix, "/") && !strings.HasSuffix(prefix, string(filepath.Separator)) {
						prefix += "/"
					}
					displayPath = prefix + displayPath
				}
			}
		}
	}
	var proc LanguageProcessor
	var ok bool
	
	// In raw mode, use RawProcessor for all text files
	if opts.RawMode {
		rawProc := NewRawProcessor()
		// In raw mode, always use the raw processor
		proc = rawProc
		ok = true
	} else {
		// Normal mode - get processor for file
		proc, ok = GetByFilename(filename)
	}
	
	if !ok && !opts.RawMode {
		return nil, fmt.Errorf("no processor found for file: %s", filename)
	}
	
	dbg.Logf(debug.LevelDetailed, "Using %s processor for %s", proc.Language(), filename)
	
	// Enable tree-sitter if requested and supported
	if p.useTreeSitter && proc.Language() == "python" {
		// Use reflection to call EnableTreeSitter if it exists
		if enabler, ok := proc.(interface{ EnableTreeSitter() }); ok {
			enableTSFunc := enabler.EnableTreeSitter
			enableTSFunc()
			dbg.Logf(debug.LevelDetailed, "Enabled tree-sitter for Python")
		}
	}

	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Process file using our debug-enabled context
	// Check if processor supports ProcessWithOptions
	if procWithOpts, ok := proc.(interface {
		ProcessWithOptions(context.Context, io.Reader, string, ProcessOptions) (*ir.DistilledFile, error)
	}); ok {
		result, err := procWithOpts.ProcessWithOptions(p.ctx, file, displayPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to process file: %w", err)
		}
		
		// Dump the IR structure at trace level
		debug.Lazy(p.ctx, debug.LevelTrace, func(d debug.Debugger) {
			d.Dump(debug.LevelTrace, fmt.Sprintf("IR for %s", filepath.Base(filename)), result)
		})
		
		return result, nil
	}
	
	// Fallback to regular Process method
	result, err := proc.Process(p.ctx, file, displayPath)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %w", err)
	}
	
	// Dump the IR structure at trace level
	debug.Lazy(p.ctx, debug.LevelTrace, func(d debug.Debugger) {
		d.Dump(debug.LevelTrace, fmt.Sprintf("IR for %s", filepath.Base(filename)), result)
	})

	// Apply stripping options (but not in raw mode)
	if !opts.RawMode {
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
			dbg.Logf(debug.LevelDetailed, "Applying stripper with options: %+v", stripOpts)
			
			s := stripper.New(stripOpts)
			strippedNode := result.Accept(s)
			if file, ok := strippedNode.(*ir.DistilledFile); ok {
				// Dump stripped result at trace level
				debug.Lazy(p.ctx, debug.LevelTrace, func(d debug.Debugger) {
					d.Dump(debug.LevelTrace, fmt.Sprintf("Stripped IR for %s", filepath.Base(filename)), file)
				})
				return file, nil
			}
			return nil, fmt.Errorf("unexpected node type after stripping")
		}
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
	// Use concurrent processing if workers > 1
	if opts.Workers == 0 || opts.Workers > 1 {
		return p.processDirectoryConcurrent(dir, opts)
	}

	// Calculate display path for directory
	displayPath := dir
	if opts.FilePathType == "relative" && opts.BasePath != "" {
		// Try to make path relative to base path
		absBase, err := filepath.Abs(opts.BasePath)
		if err == nil {
			relPath, err := filepath.Rel(absBase, dir)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				displayPath = relPath
				
				// Apply prefix if specified
				if opts.RelativePathPrefix != "" {
					prefix := opts.RelativePathPrefix
					// Ensure prefix ends with separator if not empty and doesn't already
					if !strings.HasSuffix(prefix, "/") && !strings.HasSuffix(prefix, string(filepath.Separator)) {
						prefix += "/"
					}
					displayPath = prefix + displayPath
				}
			}
		}
	}

	// Serial processing (workers == 1)
	result := &ir.DistilledDirectory{
		BaseNode: ir.BaseNode{},
		Path:     displayPath,
		Children: []ir.DistilledNode{},
	}

	// Walk directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip .aid directories completely
			if filepath.Base(path) == ".aid" {
				return filepath.SkipDir
			}
			
			// If not recursive and not the root directory, skip subdirectories
			if !opts.Recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files containing '.aid.' anywhere in filename
		basename := filepath.Base(path)
		if strings.Contains(basename, ".aid.") {
			return nil
		}

		// Check include/exclude patterns
		if !shouldIncludeFile(path, opts.IncludePatterns, opts.ExcludePatterns) {
			return nil
		}

		// Check if we can process this file
		if opts.RawMode {
			// In raw mode, process all files
			// Skip directories which are already filtered out above
		} else {
			// Normal mode - check if we have a processor
			if _, ok := GetByFilename(path); !ok {
				return nil
			}
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

// shouldIncludeFile checks if a file should be included based on include/exclude patterns
func shouldIncludeFile(filePath string, includePatterns, excludePatterns []string) bool {
	// If exclude patterns are specified and any matches, exclude the file
	for _, excludePattern := range excludePatterns {
		if excludePattern != "" {
			matched, err := filepath.Match(excludePattern, filePath)
			if err == nil && matched {
				return false
			}
			// Also check if the pattern matches the filename only
			matched, err = filepath.Match(excludePattern, filepath.Base(filePath))
			if err == nil && matched {
				return false
			}
		}
	}
	
	// If include patterns are specified, only include files that match any pattern
	if len(includePatterns) > 0 {
		for _, includePattern := range includePatterns {
			if includePattern != "" {
				matched, err := filepath.Match(includePattern, filePath)
				if err == nil && matched {
					return true
				}
				// Also check if the pattern matches the filename only
				matched, err = filepath.Match(includePattern, filepath.Base(filePath))
				if err == nil && matched {
					return true
				}
			}
		}
		// If include patterns are specified but none match, exclude
		return false
	}
	
	// No patterns specified or exclude patterns don't match, include the file
	return true
}

