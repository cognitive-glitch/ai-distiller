package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// TextFormatter implements the text output formatter for AI consumption
type TextFormatter struct {
	BaseFormatter
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(options Options) Formatter {
	return &TextFormatter{
		BaseFormatter: NewBaseFormatter(options),
	}
}

// Format implements formatter.Formatter
func (f *TextFormatter) Format(w io.Writer, file *ir.DistilledFile) error {
	// Write file header
	fmt.Fprintf(w, "<file path=\"%s\">\n", file.Path)
	
	// Write file contents
	for _, child := range file.Children {
		if err := f.formatNode(w, child, 0); err != nil {
			return err
		}
	}
	
	// Write file footer
	fmt.Fprintln(w, "</file>")
	
	return nil
}

// FormatMultiple implements formatter.Formatter
func (f *TextFormatter) FormatMultiple(w io.Writer, files []*ir.DistilledFile) error {
	for _, file := range files {
		if err := f.Format(w, file); err != nil {
			return err
		}
		fmt.Fprintln(w) // Add blank line between files
	}
	return nil
}

// Extension implements formatter.Formatter
func (f *TextFormatter) Extension() string {
	return "txt"
}

// FormatError implements formatter.Formatter
func (f *TextFormatter) FormatError(w io.Writer, err error) error {
	fmt.Fprintf(w, "ERROR: %v\n", err)
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
	case *ir.DistilledInterface:
		return f.formatInterface(w, n, indent)
	case *ir.DistilledTypeAlias:
		return f.formatTypeAlias(w, n, indentStr)
	case *ir.DistilledEnum:
		return f.formatEnum(w, n, indent)
	case *ir.DistilledPackage:
		return f.formatPackage(w, n, indentStr)
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
	
	// Format function signature with visibility prefix and def keyword
	visPrefix := getVisibilityPrefix(fn.Visibility)
	fmt.Fprintf(w, "%s%sdef %s(", indent, visPrefix, fn.Name)
	
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
	if fn.Returns != nil {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Format implementation
	if fn.Implementation != "" {
		fmt.Fprintln(w, ":")
		// Indent implementation
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
	// Format decorators
	for _, dec := range field.Decorators {
		fmt.Fprintf(w, "%s@%s\n", indent, dec)
	}
	
	// Format field with visibility prefix
	visPrefix := getVisibilityPrefix(field.Visibility)
	modifiers := formatModifiers(field.Modifiers)
	
	fmt.Fprintf(w, "%s%s%s%s", indent, visPrefix, modifiers, field.Name)
	
	// Add type if present
	if field.Type != nil {
		fmt.Fprintf(w, ": %s", field.Type.Name)
	}
	
	// Add default value if present
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

func (f *TextFormatter) formatInterface(w io.Writer, intf *ir.DistilledInterface, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format interface declaration
	fmt.Fprintf(w, "\n%sinterface %s", indentStr, intf.Name)
	
	// Add extends if any
	if len(intf.Extends) > 0 {
		bases := make([]string, len(intf.Extends))
		for i, base := range intf.Extends {
			bases[i] = base.Name
		}
		fmt.Fprintf(w, "(%s)", strings.Join(bases, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format interface body
	if len(intf.Children) > 0 {
		for _, child := range intf.Children {
			if err := f.formatNode(w, child, indent+1); err != nil {
				return err
			}
		}
	} else {
		fmt.Fprintf(w, "%s    pass\n", indentStr)
	}
	
	return nil
}

func (f *TextFormatter) formatTypeAlias(w io.Writer, alias *ir.DistilledTypeAlias, indent string) error {
	visPrefix := getVisibilityPrefix(alias.Visibility)
	fmt.Fprintf(w, "%s%stype %s = %s\n", indent, visPrefix, alias.Name, alias.Type.Name)
	return nil
}

func (f *TextFormatter) formatEnum(w io.Writer, enum *ir.DistilledEnum, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format enum declaration
	fmt.Fprintf(w, "\n%senum %s:\n", indentStr, enum.Name)
	
	// Format enum members
	if len(enum.Children) > 0 {
		for _, child := range enum.Children {
			if err := f.formatNode(w, child, indent+1); err != nil {
				return err
			}
		}
	} else {
		fmt.Fprintf(w, "%s    pass\n", indentStr)
	}
	
	return nil
}

func (f *TextFormatter) formatPackage(w io.Writer, pkg *ir.DistilledPackage, indent string) error {
	fmt.Fprintf(w, "%spackage %s\n", indent, pkg.Name)
	
	// Format package children
	for _, child := range pkg.Children {
		if err := f.formatNode(w, child, len(indent)/4); err != nil {
			return err
		}
	}
	
	return nil
}

// getVisibilityPrefix converts visibility to UML-style prefix
func getVisibilityPrefix(vis ir.Visibility) string {
	switch vis {
	case ir.VisibilityPublic:
		return "+"
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "#"
	case ir.VisibilityInternal:
		return "~"
	default:
		return "+"
	}
}

// formatModifiers formats modifiers as a string
func formatModifiers(modifiers []ir.Modifier) string {
	if len(modifiers) == 0 {
		return ""
	}
	
	var mods []string
	for _, mod := range modifiers {
		switch mod {
		case ir.ModifierStatic:
			mods = append(mods, "static")
		case ir.ModifierFinal:
			mods = append(mods, "final")
		case ir.ModifierAbstract:
			mods = append(mods, "abstract")
		case ir.ModifierAsync:
			mods = append(mods, "async")
		case ir.ModifierReadonly:
			mods = append(mods, "readonly")
		}
	}
	
	if len(mods) > 0 {
		return strings.Join(mods, " ") + " "
	}
	return ""
}