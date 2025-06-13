package formatter

import (
	"io"
	"github.com/janreges/ai-distiller/internal/ir"
)

// LanguageFormatter defines the interface for language-specific text formatting
type LanguageFormatter interface {
	// FormatNode formats a single IR node in language-specific syntax
	FormatNode(w io.Writer, node ir.DistilledNode, indent int) error
	
	// GetLanguage returns the language this formatter supports
	GetLanguage() string
}

// BaseLanguageFormatter provides common functionality for language formatters
type BaseLanguageFormatter struct {
	language string
}

// NewBaseLanguageFormatter creates a new base language formatter
func NewBaseLanguageFormatter(language string) BaseLanguageFormatter {
	return BaseLanguageFormatter{language: language}
}

// GetLanguage returns the supported language
func (f *BaseLanguageFormatter) GetLanguage() string {
	return f.language
}