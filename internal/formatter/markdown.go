package formatter

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// MarkdownFormatter formats IR as Markdown by wrapping text formatter output
type MarkdownFormatter struct {
	BaseFormatter
	textFormatter *TextFormatter
}

// NewMarkdownFormatter creates a new Markdown formatter
func NewMarkdownFormatter(options Options) *MarkdownFormatter {
	return &MarkdownFormatter{
		BaseFormatter: NewBaseFormatter(options),
		textFormatter: NewTextFormatter(options).(*TextFormatter),
	}
}

// Extension returns the file extension for Markdown
func (f *MarkdownFormatter) Extension() string {
	return ".md"
}

// Format writes a single file as Markdown
func (f *MarkdownFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	// First, format as text
	var textBuf bytes.Buffer
	if err := f.textFormatter.Format(&textBuf, file); err != nil {
		return err
	}
	
	// Convert text format to markdown
	text := textBuf.String()
	
	// Extract file content from <file> tags
	fileRe := regexp.MustCompile(`(?s)<file path="([^"]+)">\s*(.*?)\s*</file>`)
	matches := fileRe.FindStringSubmatch(text)
	
	if len(matches) > 2 {
		path := matches[1]
		content := matches[2]
		
		// Write as markdown
		fmt.Fprintf(w, "### %s\n\n", path)
		
		// Determine language for syntax highlighting
		lang := f.getLanguageIdentifier(file.Language)
		
		fmt.Fprintf(w, "```%s\n", lang)
		fmt.Fprint(w, content)
		if !strings.HasSuffix(content, "\n") {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "```")
	} else {
		// Fallback - just write the text as-is
		fmt.Fprint(w, text)
	}
	
	return nil
}

// FormatMultiple writes multiple files as Markdown
func (f *MarkdownFormatter) FormatMultiple(w io.Writer, files []*ir.DistilledFile) error {
	for i, file := range files {
		if i > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		if err := f.Format(w, file); err != nil {
			return err
		}
	}
	return nil
}

// getLanguageIdentifier returns the correct language identifier for syntax highlighting
func (f *MarkdownFormatter) getLanguageIdentifier(language string) string {
	// Map our language names to common markdown identifiers
	switch strings.ToLower(language) {
	case "python":
		return "python"
	case "go", "golang":
		return "go"
	case "typescript":
		return "typescript"
	case "javascript":
		return "javascript"
	case "java":
		return "java"
	case "csharp", "c#":
		return "csharp"
	case "cpp", "c++":
		return "cpp"
	case "ruby":
		return "ruby"
	case "rust":
		return "rust"
	case "swift":
		return "swift"
	case "kotlin":
		return "kotlin"
	case "php":
		return "php"
	default:
		return language
	}
}