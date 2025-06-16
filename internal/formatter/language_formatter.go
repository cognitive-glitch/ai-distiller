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
	
	// Reset resets the formatter state for a new file (optional)
	Reset()
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

// Reset provides a default no-op implementation
func (f *BaseLanguageFormatter) Reset() {
	// Default implementation does nothing
}

// GetVisibilityKeyword returns the language-specific visibility keyword
// This is a helper that each language formatter can override
func (f *BaseLanguageFormatter) GetVisibilityKeyword(visibility ir.Visibility, language string) string {
	// Default implementation for most languages
	switch language {
	case "go":
		// Go uses uppercase/lowercase naming convention, no keywords
		return ""
	case "python", "ruby":
		// Python and Ruby don't have visibility keywords
		return ""
	default:
		// Most languages use standard keywords
		switch visibility {
		case ir.VisibilityPublic:
			return "public"
		case ir.VisibilityPrivate:
			return "private"
		case ir.VisibilityProtected:
			return "protected"
		case ir.VisibilityInternal:
			if language == "csharp" {
				return "internal"
			} else if language == "swift" {
				return "internal"
			} else if language == "kotlin" {
				return "internal"
			}
			return "" // Other languages might not have internal
		case ir.VisibilityFilePrivate:
			if language == "swift" {
				return "fileprivate"
			}
			return "private"
		case ir.VisibilityOpen:
			if language == "swift" {
				return "open"
			}
			return "public"
		case ir.VisibilityProtectedInternal:
			if language == "csharp" {
				return "protected internal"
			}
			return "protected"
		case ir.VisibilityPrivateProtected:
			if language == "csharp" {
				return "private protected"
			}
			return "private"
		default:
			return ""
		}
	}
}