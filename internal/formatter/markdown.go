package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// MarkdownFormatter formats IR as Markdown
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

// Format writes a single file as Markdown
func (f *MarkdownFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	// Write file header
	fmt.Fprintf(w, "# %s\n\n", file.Path)

	if f.options.IncludeMetadata && file.Metadata != nil {
		fmt.Fprintf(w, "**Language:** %s\n", file.Language)
		fmt.Fprintf(w, "**Size:** %d bytes\n", file.Metadata.Size)
		if file.Metadata.LastModified.Unix() > 0 {
			fmt.Fprintf(w, "**Modified:** %s\n", file.Metadata.LastModified.Format("2006-01-02 15:04:05"))
		}
		fmt.Fprintln(w)
	}

	// Write errors if any
	if len(file.Errors) > 0 {
		fmt.Fprintf(w, "## ⚠️ Errors (%d)\n\n", len(file.Errors))
		for _, err := range file.Errors {
			f.formatError(w, &err)
		}
		fmt.Fprintln(w)
	}

	// Write nodes
	if len(file.Children) > 0 {
		fmt.Fprintln(w, "## Structure")
		fmt.Fprintln(w)
		for _, node := range file.Children {
			f.formatNode(w, node, 0)
		}
	}

	return nil
}

// FormatMultiple writes multiple files as Markdown
func (f *MarkdownFormatter) FormatMultiple(w io.Writer, files []*ir.DistilledFile) error {
	for i, file := range files {
		if i > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "---")
			fmt.Fprintln(w)
		}
		if err := f.Format(w, file); err != nil {
			return err
		}
	}
	return nil
}

// formatNode formats a single node
func (f *MarkdownFormatter) formatNode(w io.Writer, node ir.DistilledNode, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *ir.DistilledPackage:
		fmt.Fprintf(w, "%s📦 **Package** `%s`\n", indent, n.Name)
		f.formatChildren(w, n.Children, depth+1)

	case *ir.DistilledImport:
		fmt.Fprintf(w, "%s📥 **Import** ", indent)
		if n.ImportType == "from" {
			fmt.Fprintf(w, "from `%s` import ", n.Module)
			names := make([]string, len(n.Symbols))
			for i, sym := range n.Symbols {
				if sym.Alias != "" {
					names[i] = fmt.Sprintf("`%s` as `%s`", sym.Name, sym.Alias)
				} else {
					names[i] = fmt.Sprintf("`%s`", sym.Name)
				}
			}
			fmt.Fprintf(w, "%s", strings.Join(names, ", "))
		} else {
			fmt.Fprintf(w, "`%s`", n.Module)
		}
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)

	case *ir.DistilledClass:
		fmt.Fprintf(w, "%s🏛️ **Class** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		f.formatModifiers(w, n.Modifiers)
		if len(n.Extends) > 0 || len(n.Implements) > 0 {
			fmt.Fprint(w, " (")
			if len(n.Extends) > 0 {
				fmt.Fprintf(w, "extends %s", f.formatTypeRefs(n.Extends))
			}
			if len(n.Implements) > 0 {
				if len(n.Extends) > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprintf(w, "implements %s", f.formatTypeRefs(n.Implements))
			}
			fmt.Fprint(w, ")")
		}
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)
		f.formatChildren(w, n.Children, depth+1)

	case *ir.DistilledInterface:
		fmt.Fprintf(w, "%s🔌 **Interface** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		if len(n.Extends) > 0 {
			fmt.Fprintf(w, " extends %s", f.formatTypeRefs(n.Extends))
		}
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)
		f.formatChildren(w, n.Children, depth+1)

	case *ir.DistilledStruct:
		fmt.Fprintf(w, "%s📐 **Struct** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)
		f.formatChildren(w, n.Children, depth+1)

	case *ir.DistilledEnum:
		fmt.Fprintf(w, "%s🎲 **Enum** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)
		f.formatChildren(w, n.Children, depth+1)

	case *ir.DistilledFunction:
		fmt.Fprintf(w, "%s🔧 **Function** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		f.formatModifiers(w, n.Modifiers)

		// Format parameters
		fmt.Fprint(w, "(")
		params := make([]string, len(n.Parameters))
		for i, p := range n.Parameters {
			params[i] = fmt.Sprintf("`%s`", p.Name)
			if p.Type.Name != "" && p.Type.Name != "Any" {
				params[i] += fmt.Sprintf(": `%s`", p.Type.Name)
			}
		}
		fmt.Fprintf(w, "%s)", strings.Join(params, ", "))

		// Format return type
		if n.Returns != nil && n.Returns.Name != "" {
			fmt.Fprintf(w, " → `%s`", n.Returns.Name)
		}

		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)

		// Format implementation if included
		if n.Implementation != "" && !f.options.Compact {
			fmt.Fprintf(w, "%s  ```\n", indent)
			lines := strings.Split(n.Implementation, "\n")
			for _, line := range lines {
				fmt.Fprintf(w, "%s  %s\n", indent, line)
			}
			fmt.Fprintf(w, "%s  ```\n", indent)
		}

	case *ir.DistilledField:
		fmt.Fprintf(w, "%s📊 **Field** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		f.formatModifiers(w, n.Modifiers)
		if n.Type != nil && n.Type.Name != "" {
			fmt.Fprintf(w, ": `%s`", n.Type.Name)
		}
		if n.DefaultValue != "" {
			fmt.Fprintf(w, " = `%s`", n.DefaultValue)
		}
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)

	case *ir.DistilledTypeAlias:
		fmt.Fprintf(w, "%s🏷️ **Type** `%s`", indent, n.Name)
		f.formatVisibility(w, n.Visibility)
		fmt.Fprintf(w, " = `%s`", n.Type.Name)
		f.formatLocation(w, n, depth)
		fmt.Fprintln(w)

	case *ir.DistilledComment:
		if !f.options.Compact {
			fmt.Fprintf(w, "%s💬 ", indent)
			if n.Format == "doc" {
				fmt.Fprint(w, "**Doc:** ")
			}
			fmt.Fprintf(w, "*%s*", strings.TrimSpace(n.Text))
			f.formatLocation(w, n, depth)
			fmt.Fprintln(w)
		}

	case *ir.DistilledError:
		f.formatError(w, n)

	default:
		// Generic node handling
		fmt.Fprintf(w, "%s• %s", indent, node.GetNodeKind())
		f.formatLocation(w, node, depth)
		fmt.Fprintln(w)
	}
}

// formatChildren formats child nodes
func (f *MarkdownFormatter) formatChildren(w io.Writer, children []ir.DistilledNode, depth int) {
	for _, child := range children {
		f.formatNode(w, child, depth)
	}
}

// formatError formats an error node
func (f *MarkdownFormatter) formatError(w io.Writer, err *ir.DistilledError) {
	icon := "⚠️"
	if err.Severity == "error" {
		icon = "❌"
	}
	severity := err.Severity
	if len(severity) > 0 {
		severity = strings.ToUpper(severity[:1]) + severity[1:]
	}
	fmt.Fprintf(w, "%s **%s**: %s", icon, severity, err.Message)
	if err.Code != "" {
		fmt.Fprintf(w, " [%s]", err.Code)
	}
	if f.options.IncludeLocation {
		loc := err.GetLocation()
		fmt.Fprintf(w, " (line %d)", loc.StartLine)
	}
	fmt.Fprintln(w)
}

// formatVisibility adds visibility info
func (f *MarkdownFormatter) formatVisibility(w io.Writer, vis ir.Visibility) {
	if vis != "" && vis != ir.VisibilityPublic {
		fmt.Fprintf(w, " _%s_", vis)
	}
}

// formatModifiers adds modifier info
func (f *MarkdownFormatter) formatModifiers(w io.Writer, mods []ir.Modifier) {
	for _, mod := range mods {
		fmt.Fprintf(w, " _%s_", mod)
	}
}

// formatTypeRefs formats type references
func (f *MarkdownFormatter) formatTypeRefs(refs []ir.TypeRef) string {
	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = fmt.Sprintf("`%s`", ref.Name)
	}
	return strings.Join(names, ", ")
}

// formatLocation adds location info if enabled
func (f *MarkdownFormatter) formatLocation(w io.Writer, node ir.DistilledNode, depth int) {
	if f.options.IncludeLocation {
		loc := node.GetLocation()
		if loc.StartLine > 0 {
			fmt.Fprintf(w, " <sub>L%d", loc.StartLine)
			if loc.EndLine > loc.StartLine {
				fmt.Fprintf(w, "-%d", loc.EndLine)
			}
			fmt.Fprint(w, "</sub>")
		}
	}
}
