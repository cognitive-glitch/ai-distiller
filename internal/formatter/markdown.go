package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// MarkdownFormatter formats IR as clean, standard Markdown with code blocks
type MarkdownFormatter struct {
	BaseFormatter
}

// NewMarkdownFormatter creates a new Markdown formatter
func NewMarkdownFormatter(options Options) *MarkdownFormatter {
	return &MarkdownFormatter{
		BaseFormatter: NewBaseFormatter(options),
	}
}

// Extension returns the file extension for Markdown
func (f *MarkdownFormatter) Extension() string {
	return ".md"
}

// Format writes a single file as clean Markdown
func (f *MarkdownFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	// Write file header
	fmt.Fprintf(w, "# %s\n\n", file.Path)

	// Determine language for syntax highlighting
	lang := f.getLanguageIdentifier(file.Language)

	// Write the code block
	fmt.Fprintf(w, "```%s\n", lang)
	
	// Format all nodes as simplified code representation
	for _, node := range file.Children {
		f.formatNode(w, node, "")
	}
	
	fmt.Fprintln(w, "```")

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

// formatNode formats a node in a simplified code representation
func (f *MarkdownFormatter) formatNode(w io.Writer, node ir.DistilledNode, indent string) {
	switch n := node.(type) {
	case *ir.DistilledPackage:
		fmt.Fprintf(w, "%spackage %s\n\n", indent, n.Name)
		f.formatChildren(w, n.Children, indent)

	case *ir.DistilledImport:
		if n.ImportType == "from" {
			fmt.Fprintf(w, "%sfrom %s import ", indent, n.Module)
			names := make([]string, len(n.Symbols))
			for i, sym := range n.Symbols {
				if sym.Alias != "" {
					names[i] = fmt.Sprintf("%s as %s", sym.Name, sym.Alias)
				} else {
					names[i] = sym.Name
				}
			}
			fmt.Fprintf(w, "%s\n", strings.Join(names, ", "))
		} else {
			fmt.Fprintf(w, "%simport %s\n", indent, n.Module)
		}

	case *ir.DistilledComment:
		// Skip comments in markdown - they'll be part of docstrings or ignored
		return

	case *ir.DistilledClass:
		vis := f.getVisibilityMarker(n.Visibility)
		fmt.Fprintf(w, "%s%sclass %s", indent, vis, n.Name)
		
		// Add inheritance
		if len(n.Extends) > 0 || len(n.Implements) > 0 {
			parts := []string{}
			for _, ext := range n.Extends {
				parts = append(parts, ext.Name)
			}
			for _, impl := range n.Implements {
				parts = append(parts, impl.Name)
			}
			fmt.Fprintf(w, "(%s)", strings.Join(parts, ", "))
		}
		fmt.Fprintln(w, ":")
		
		// Add children
		if len(n.Children) > 0 {
			f.formatChildren(w, n.Children, indent+"    ")
		}

	case *ir.DistilledInterface:
		vis := f.getVisibilityMarker(n.Visibility)
		fmt.Fprintf(w, "%s%sinterface %s", indent, vis, n.Name)
		
		if len(n.Extends) > 0 {
			names := make([]string, len(n.Extends))
			for i, ext := range n.Extends {
				names[i] = ext.Name
			}
			fmt.Fprintf(w, " extends %s", strings.Join(names, ", "))
		}
		fmt.Fprintln(w, ":")
		
		if len(n.Children) > 0 {
			f.formatChildren(w, n.Children, indent+"    ")
		}

	case *ir.DistilledStruct:
		vis := f.getVisibilityMarker(n.Visibility)
		fmt.Fprintf(w, "%s%sstruct %s:\n", indent, vis, n.Name)
		if len(n.Children) > 0 {
			f.formatChildren(w, n.Children, indent+"    ")
		}

	case *ir.DistilledEnum:
		vis := f.getVisibilityMarker(n.Visibility)
		fmt.Fprintf(w, "%s%senum %s:\n", indent, vis, n.Name)
		if len(n.Children) > 0 {
			f.formatChildren(w, n.Children, indent+"    ")
		}


	case *ir.DistilledFunction:
		vis := f.getVisibilityMarker(n.Visibility)
		
		// Modifiers
		mods := ""
		if len(n.Modifiers) > 0 {
			modStrings := make([]string, len(n.Modifiers))
			for i, mod := range n.Modifiers {
				modStrings[i] = string(mod)
			}
			mods = strings.Join(modStrings, " ") + " "
		}
		
		fmt.Fprintf(w, "%s%s%s%s(", indent, vis, mods, n.Name)
		
		// Parameters
		params := make([]string, len(n.Parameters))
		for i, p := range n.Parameters {
			params[i] = p.Name
			if p.Type.Name != "" && p.Type.Name != "Any" {
				params[i] += ": " + p.Type.Name
			}
		}
		fmt.Fprintf(w, "%s)", strings.Join(params, ", "))
		
		// Return type
		if n.Returns != nil && n.Returns.Name != "" {
			fmt.Fprintf(w, " -> %s", n.Returns.Name)
		}
		
		// Implementation
		if n.Implementation != "" && !f.options.Compact {
			fmt.Fprintln(w, ":")
			lines := strings.Split(strings.TrimRight(n.Implementation, "\n"), "\n")
			for _, line := range lines {
				fmt.Fprintf(w, "%s%s\n", indent+"    ", line)
			}
		} else {
			fmt.Fprintln(w)
		}

	case *ir.DistilledField:
		vis := f.getVisibilityMarker(n.Visibility)
		
		// Modifiers
		mods := ""
		if len(n.Modifiers) > 0 {
			modStrings := make([]string, len(n.Modifiers))
			for i, mod := range n.Modifiers {
				modStrings[i] = string(mod)
			}
			mods = strings.Join(modStrings, " ") + " "
		}
		
		fmt.Fprintf(w, "%s%s%s%s", indent, vis, mods, n.Name)
		
		if n.Type != nil && n.Type.Name != "" {
			fmt.Fprintf(w, ": %s", n.Type.Name)
		}
		
		if n.DefaultValue != "" {
			fmt.Fprintf(w, " = %s", n.DefaultValue)
		}
		
		fmt.Fprintln(w)


	case *ir.DistilledTypeAlias:
		vis := f.getVisibilityMarker(n.Visibility)
		fmt.Fprintf(w, "%s%stype %s = %s\n", indent, vis, n.Name, n.Type.Name)

	default:
		// For other node types, just recurse into children if they have any
		var children []ir.DistilledNode
		switch n := node.(type) {
		case *ir.DistilledDirectory:
			children = n.Children
		case *ir.DistilledFile:
			children = n.Children
		}
		if len(children) > 0 {
			f.formatChildren(w, children, indent)
		}
	}
}

// formatChildren formats child nodes
func (f *MarkdownFormatter) formatChildren(w io.Writer, children []ir.DistilledNode, indent string) {
	for _, child := range children {
		f.formatNode(w, child, indent)
	}
}

// getVisibilityMarker returns the visibility marker for the code
func (f *MarkdownFormatter) getVisibilityMarker(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "*"
	case ir.VisibilityInternal:
		return "~"
	default:
		return ""
	}
}