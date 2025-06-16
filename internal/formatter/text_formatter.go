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
	// fmt.Printf("File has %d children\n", len(file.Children))
	for _, child := range file.Children {
		// fmt.Printf("  Child %d: type=%T\n", i, child)
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
	case *ir.DistilledRawContent:
		return f.formatRawContent(w, n, indentStr)
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
	
	// Format API docblock if present (PHP)
	if class.APIDocblock != "" {
		// Add a newline before the docblock if at top level
		if indent == 0 {
			fmt.Fprintln(w)
		}
		// Output the docblock before the class
		lines := strings.Split(class.APIDocblock, "\n")
		for _, line := range lines {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}
	
	// Format decorators/attributes
	for _, dec := range class.Decorators {
		fmt.Fprintf(w, "%s@%s\n", indentStr, dec)
	}
	
	// Format class declaration - check if it's a struct
	classType := "class"
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierStruct {
			classType = "struct"
			break
		}
	}
	// Add newline before class if no docblock or decorators
	if class.APIDocblock == "" && len(class.Decorators) == 0 {
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "%s%s %s", indentStr, classType, class.Name)
	
	// Add base classes if any
	if len(class.Extends) > 0 {
		bases := make([]string, len(class.Extends))
		for i, base := range class.Extends {
			bases[i] = base.Name
		}
		fmt.Fprintf(w, "(%s)", strings.Join(bases, ", "))
	}
	
	// Add implements relationships
	if len(class.Implements) > 0 {
		implements := make([]string, len(class.Implements))
		for i, impl := range class.Implements {
			implements[i] = impl.Name
		}
		if len(class.Extends) > 0 {
			fmt.Fprintf(w, " implements %s", strings.Join(implements, ", "))
		} else {
			fmt.Fprintf(w, " implements %s", strings.Join(implements, ", "))
		}
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
	
	// Check if this is a magic method from PHP docblock
	isMethodFromDocblock := false
	if fn.Extensions != nil && fn.Extensions.PHP != nil {
		if fn.Extensions.PHP.Origin == ir.FieldOriginDocblock {
			isMethodFromDocblock = true
		}
	}
	
	// Format function signature with visibility prefix and modifiers
	visPrefix := getVisibilityPrefix(fn.Visibility)
	modifiers := formatModifiers(fn.Modifiers)
	
	if isMethodFromDocblock {
		// For magic methods, prefix with "method"
		fmt.Fprintf(w, "%smethod %s%s(", indent, modifiers, fn.Name)
	} else {
		fmt.Fprintf(w, "%s%s%s%s(", indent, visPrefix, modifiers, fn.Name)
	}
	
	// Format parameters
	// fmt.Printf("Function %s has %d parameters\n", fn.Name, len(fn.Parameters))
	params := make([]string, 0, len(fn.Parameters))
	for _, param := range fn.Parameters {
		// fmt.Printf("  Param: name=%q, type=%q\n", param.Name, param.Type.Name)
		if param.Name == "" {
			continue // Skip empty parameters
		}
		paramStr := param.Name
		if param.Type.Name != "" {
			paramStr += ": " + param.Type.Name
		}
		if param.DefaultValue != "" {
			paramStr += " = " + param.DefaultValue
		}
		params = append(params, paramStr)
	}
	fmt.Fprintf(w, "%s)", strings.Join(params, ", "))
	
	// Format return type
	if fn.Returns != nil {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Add source annotation comment for magic methods
	if isMethodFromDocblock && fn.Extensions.PHP.SourceAnnotation != "" {
		fmt.Fprintf(w, " // %s", fn.Extensions.PHP.SourceAnnotation)
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
	
	// Check if this is a magic property from PHP docblock
	isPropertyFromDocblock := false
	var accessMode string
	if field.Extensions != nil && field.Extensions.PHP != nil {
		if field.Extensions.PHP.Origin == ir.FieldOriginDocblock {
			isPropertyFromDocblock = true
			switch field.Extensions.PHP.AccessMode {
			case ir.FieldAccessReadOnly:
				accessMode = "property-read "
			case ir.FieldAccessWriteOnly:
				accessMode = "property-write "
			default:
				accessMode = "property "
			}
		}
	}
	
	// Format field with visibility prefix
	visPrefix := getVisibilityPrefix(field.Visibility)
	modifiers := formatModifiers(field.Modifiers)
	
	if isPropertyFromDocblock {
		// For magic properties, show property/property-read/property-write
		fmt.Fprintf(w, "%s%s%s%s", indent, accessMode, modifiers, field.Name)
	} else {
		fmt.Fprintf(w, "%s%s%s%s", indent, visPrefix, modifiers, field.Name)
	}
	
	// Add type if present
	if field.Type != nil {
		fmt.Fprintf(w, ": %s", field.Type.Name)
	}
	
	// Add default value if present
	if field.DefaultValue != "" {
		fmt.Fprintf(w, " = %s", field.DefaultValue)
	}
	
	// Add source annotation comment for magic properties
	if isPropertyFromDocblock && field.Extensions.PHP.SourceAnnotation != "" {
		fmt.Fprintf(w, " // %s", field.Extensions.PHP.SourceAnnotation)
	}
	
	fmt.Fprintln(w)
	return nil
}

func (f *TextFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent string) error {
	// Just output the comment text as-is for language-agnostic output
	fmt.Fprintf(w, "%s%s\n", indent, comment.Text)
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
	modifiers := formatModifiers(alias.Modifiers)
	
	// Check for export modifier
	hasExport := false
	for _, mod := range alias.Modifiers {
		if mod == ir.ModifierExport {
			hasExport = true
			break
		}
	}
	
	if hasExport {
		fmt.Fprintf(w, "%sexport %s%stype %s = %s\n", indent, visPrefix, modifiers, alias.Name, alias.Type.Name)
	} else {
		fmt.Fprintf(w, "%s%s%stype %s = %s\n", indent, visPrefix, modifiers, alias.Name, alias.Type.Name)
	}
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

func (f *TextFormatter) formatRawContent(w io.Writer, n *ir.DistilledRawContent, indent string) error {
	// Output raw content as-is without any processing
	fmt.Fprint(w, n.Content)
	// Ensure there's a newline at the end if content doesn't have one
	if len(n.Content) > 0 && n.Content[len(n.Content)-1] != '\n' {
		fmt.Fprintln(w)
	}
	return nil
}

// getVisibilityPrefix converts visibility to UML-style prefix
func getVisibilityPrefix(vis ir.Visibility) string {
	// For now, keep using prefixes in the base text formatter
	// The language-aware formatter will use full keywords
	switch vis {
	case ir.VisibilityPublic:
		return ""   // No prefix for public
	case ir.VisibilityPrivate:
		return "-"  // Private
	case ir.VisibilityProtected:
		return "*"  // Protected
	case ir.VisibilityInternal:
		return "~"  // UML package/internal
	case ir.VisibilityFilePrivate:
		return "-"  // Swift fileprivate -> similar to private
	case ir.VisibilityOpen:
		return ""   // Swift open -> treat as public
	case ir.VisibilityProtectedInternal:
		return "*~" // C# protected internal -> combination
	case ir.VisibilityPrivateProtected:
		return "-*" // C# private protected -> combination
	default:
		return ""   // Default to public (no prefix)
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
		case ir.ModifierEmbedded:
			mods = append(mods, "embeds")
		case ir.ModifierMutating:
			mods = append(mods, "mutating")
		case ir.ModifierActor:
			mods = append(mods, "actor")
		// Skip ModifierStruct as it's handled in formatClass
		}
	}
	
	if len(mods) > 0 {
		return strings.Join(mods, " ") + " "
	}
	return ""
}