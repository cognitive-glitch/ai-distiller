package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// TextFormatter formats IR as ultra-compact plain text for AI consumption
type TextFormatter struct {
	options Options
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(opts Options) Formatter {
	return &TextFormatter{
		options: opts,
	}
}

// getVisibilityPrefix returns a UML notation prefix based on visibility
func getVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPrivate, ir.VisibilityInternal, ir.VisibilityFilePrivate:
		return "-"
	case ir.VisibilityProtected:
		return "#"
	case ir.VisibilityPublic, ir.VisibilityOpen:
		return "+"
	default:
		return "+" // default to public
	}
}

// Format formats a distilled file as text
func (f *TextFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	return f.formatFile(w, file)
}

// FormatMultiple formats multiple files as text
func (f *TextFormatter) FormatMultiple(w io.Writer, files []*ir.DistilledFile) error {
	for i, file := range files {
		if i > 0 {
			fmt.Fprintln(w) // Empty line between files
		}
		if err := f.Format(w, file); err != nil {
			return err
		}
	}
	return nil
}

// Extension returns the recommended file extension
func (f *TextFormatter) Extension() string {
	return ".txt"
}

func (f *TextFormatter) formatFile(w io.Writer, file *ir.DistilledFile) error {
	// Start file tag
	fmt.Fprintf(w, "<file path=\"%s\">\n", file.Path)

	// Format all children
	for _, child := range file.Children {
		if err := f.formatNode(w, child, 0); err != nil {
			return err
		}
	}

	// End file tag
	fmt.Fprintln(w, "</file>")
	return nil
}

func (f *TextFormatter) formatDirectory(w io.Writer, dir *ir.DistilledDirectory) error {
	// Format each file in the directory
	for i, child := range dir.Children {
		if file, ok := child.(*ir.DistilledFile); ok {
			if i > 0 {
				fmt.Fprintln(w) // Empty line between files
			}
			if err := f.Format(w, file); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *TextFormatter) formatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	indentStr := strings.Repeat("    ", indent)

	switch n := node.(type) {
	case *ir.DistilledImport:
		return f.formatImport(w, n, indentStr)
	case *ir.DistilledClass:
		return f.formatClass(w, n, indent)
	case *ir.DistilledFunction:
		return f.formatFunction(w, n, indentStr)
	case *ir.DistilledField:
		return f.formatField(w, n, indentStr)
	case *ir.DistilledComment:
		return f.formatComment(w, n, indentStr)
	default:
		// Skip unknown nodes
		return nil
	}
}

func (f *TextFormatter) formatImport(w io.Writer, imp *ir.DistilledImport, indent string) error {
	if imp.ImportType == "from" {
		if len(imp.Symbols) > 0 {
			symbols := make([]string, len(imp.Symbols))
			for i, sym := range imp.Symbols {
				if sym.Alias != "" {
					symbols[i] = fmt.Sprintf("%s as %s", sym.Name, sym.Alias)
				} else {
					symbols[i] = sym.Name
				}
			}
			fmt.Fprintf(w, "from %s import %s\n", imp.Module, strings.Join(symbols, ", "))
		} else {
			fmt.Fprintf(w, "from %s import *\n", imp.Module)
		}
	} else {
		// For simple imports, check if there's an alias in symbols
		var alias string
		if len(imp.Symbols) > 0 && imp.Symbols[0].Alias != "" {
			alias = imp.Symbols[0].Alias
		}
		
		if alias != "" {
			fmt.Fprintf(w, "import %s as %s\n", imp.Module, alias)
		} else {
			fmt.Fprintf(w, "import %s\n", imp.Module)
		}
	}
	return nil
}

func (f *TextFormatter) formatClass(w io.Writer, class *ir.DistilledClass, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format class declaration
	fmt.Fprintf(w, "\n%sclass %s", indentStr, class.Name)
	
	// Add base classes if any
	if len(class.Extends) > 0 {
		bases := make([]string, len(class.Extends))
		for i, base := range class.Extends {
			bases[i] = base.Name
		}
		fmt.Fprintf(w, "(%s)", strings.Join(bases, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format class body
	if len(class.Children) > 0 {
		for _, child := range class.Children {
			if err := f.formatNode(w, child, indent+1); err != nil {
				return err
			}
		}
	} else {
		fmt.Fprintf(w, "%s    pass\n", indentStr)
	}
	
	return nil
}

func (f *TextFormatter) formatFunction(w io.Writer, fn *ir.DistilledFunction, indent string) error {
	// Format decorators
	for _, dec := range fn.Decorators {
		fmt.Fprintf(w, "%s@%s\n", indent, dec)
	}
	
	// Format function signature with visibility prefix (no 'def' needed)
	visPrefix := getVisibilityPrefix(fn.Visibility)
	fmt.Fprintf(w, "%s%s%s(", indent, visPrefix, fn.Name)
	
	// Format parameters
	params := make([]string, len(fn.Parameters))
	for i, param := range fn.Parameters {
		paramStr := param.Name
		if param.Type.Name != "" {
			paramStr += ": " + param.Type.Name
		}
		if param.DefaultValue != "" {
			paramStr += " = " + param.DefaultValue
		}
		params[i] = paramStr
	}
	fmt.Fprintf(w, "%s)", strings.Join(params, ", "))
	
	// Format return type
	if fn.Returns != nil && fn.Returns.Name != "" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Format body
	if fn.Implementation != "" {
		fmt.Fprintln(w, ":")
		// Add proper indentation to implementation
		lines := strings.Split(fn.Implementation, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Fprintf(w, "%s    %s\n", indent, line)
			}
		}
	} else {
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *TextFormatter) formatField(w io.Writer, field *ir.DistilledField, indent string) error {
	visPrefix := getVisibilityPrefix(field.Visibility)
	fmt.Fprintf(w, "%s%s%s", indent, visPrefix, field.Name)
	if field.Type != nil && field.Type.Name != "" {
		fmt.Fprintf(w, ": %s", field.Type.Name)
	}
	if field.DefaultValue != "" {
		fmt.Fprintf(w, " = %s", field.DefaultValue)
	}
	fmt.Fprintln(w)
	return nil
}

func (f *TextFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent string) error {
	// Format as Python comment or docstring
	lines := strings.Split(comment.Text, "\n")
	if len(lines) > 1 || strings.Contains(comment.Text, "\n") {
		// Multi-line comment/docstring
		fmt.Fprintf(w, "%s\"\"\"%s\"\"\"\n", indent, comment.Text)
	} else {
		// Single line comment
		fmt.Fprintf(w, "%s# %s\n", indent, comment.Text)
	}
	
	return nil
}