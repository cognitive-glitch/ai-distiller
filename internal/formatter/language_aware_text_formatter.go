package formatter

import (
	"fmt"
	"io"
	"sync"

	"github.com/janreges/ai-distiller/internal/ir"
)

// LanguageAwareTextFormatter is a text formatter that uses language-specific formatters
type LanguageAwareTextFormatter struct {
	BaseFormatter
	formatters map[string]LanguageFormatter
	mu         sync.RWMutex
}

// NewLanguageAwareTextFormatter creates a new language-aware text formatter
func NewLanguageAwareTextFormatter(options Options) *LanguageAwareTextFormatter {
	f := &LanguageAwareTextFormatter{
		BaseFormatter: NewBaseFormatter(options),
		formatters:    make(map[string]LanguageFormatter),
	}
	
	// Register built-in language formatters
	f.RegisterLanguageFormatter("java", NewJavaFormatter())
	f.RegisterLanguageFormatter("go", NewGoFormatter())
	// Python formatter can be added later by refactoring existing code
	// f.RegisterLanguageFormatter("python", NewPythonFormatter())
	
	return f
}

// RegisterLanguageFormatter registers a language-specific formatter
func (f *LanguageAwareTextFormatter) RegisterLanguageFormatter(language string, formatter LanguageFormatter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.formatters[language] = formatter
}

// Format implements formatter.Formatter
func (f *LanguageAwareTextFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	// Write file header
	fmt.Fprintf(w, "<file path=\"%s\">\n", file.Path)
	
	// Get language-specific formatter
	langFormatter := f.getLanguageFormatter(file.Language)
	
	// Reset formatter state for new file
	if langFormatter != nil {
		langFormatter.Reset()
	}
	
	// Write file contents
	for _, child := range file.Children {
		if langFormatter != nil {
			if err := langFormatter.FormatNode(w, child, 0); err != nil {
				return err
			}
		} else {
			// Fallback to generic formatting
			if err := f.formatNodeGeneric(w, child, 0); err != nil {
				return err
			}
		}
	}
	
	// For Go formatter, ensure import block is closed
	if goFormatter, ok := langFormatter.(*GoFormatter); ok && goFormatter.lastWasImport {
		fmt.Fprintln(w, ")")
	}
	
	// Write file footer
	fmt.Fprintln(w, "</file>")
	
	return nil
}

// FormatMultiple implements formatter.Formatter
func (f *LanguageAwareTextFormatter) FormatMultiple(w io.Writer, files []*ir.DistilledFile) error {
	for i, file := range files {
		if err := f.Format(w, file); err != nil {
			return err
		}
		if i < len(files)-1 {
			fmt.Fprintln(w) // Add blank line between files
		}
	}
	return nil
}

// Extension implements formatter.Formatter
func (f *LanguageAwareTextFormatter) Extension() string {
	return "txt"
}

// FormatError implements formatter.Formatter
func (f *LanguageAwareTextFormatter) FormatError(w io.Writer, err error) error {
	fmt.Fprintf(w, "ERROR: %v\n", err)
	return nil
}

func (f *LanguageAwareTextFormatter) getLanguageFormatter(language string) LanguageFormatter {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.formatters[language]
}

// formatNodeGeneric provides a generic fallback for unsupported languages
func (f *LanguageAwareTextFormatter) formatNodeGeneric(w io.Writer, node ir.DistilledNode, indent int) error {
	// This is a simplified generic formatter
	// In a real implementation, this could be more sophisticated
	
	switch n := node.(type) {
	case *ir.DistilledImport:
		fmt.Fprintf(w, "import %s\n", n.Module)
	case *ir.DistilledClass:
		fmt.Fprintf(w, "\nclass %s:\n", n.Name)
		for _, child := range n.Children {
			f.formatNodeGeneric(w, child, indent+1)
		}
	case *ir.DistilledFunction:
		fmt.Fprintf(w, "    function %s()\n", n.Name)
		if n.Implementation != "" {
			fmt.Fprintln(w, "        // implementation")
		}
	case *ir.DistilledField:
		fmt.Fprintf(w, "    field %s\n", n.Name)
	case *ir.DistilledComment:
		fmt.Fprintf(w, "// %s\n", n.Text)
	default:
		// Skip unknown nodes
	}
	
	return nil
}