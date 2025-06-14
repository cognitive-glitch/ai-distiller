package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// SwiftFormatter implements language-specific formatting for Swift
type SwiftFormatter struct {
	BaseLanguageFormatter
}

// NewSwiftFormatter creates a new Swift formatter
func NewSwiftFormatter() LanguageFormatter {
	return &SwiftFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("swift"),
	}
}

// FormatNode implements LanguageFormatter
func (f *SwiftFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	indentStr := strings.Repeat("    ", indent)

	switch n := node.(type) {
	case *ir.DistilledImport:
		return f.formatImport(w, n, indentStr)
	case *ir.DistilledClass:
		return f.formatClass(w, n, indent)
	case *ir.DistilledInterface:
		return f.formatProtocol(w, n, indent)
	case *ir.DistilledFunction:
		return f.formatFunction(w, n, indentStr)
	case *ir.DistilledField:
		return f.formatField(w, n, indentStr)
	case *ir.DistilledTypeAlias:
		return f.formatTypeAlias(w, n, indentStr)
	case *ir.DistilledComment:
		return f.formatComment(w, n, indentStr)
	case *ir.DistilledEnum:
		return f.formatEnum(w, n, indent)
	default:
		// Skip unknown nodes
		return nil
	}
}

func (f *SwiftFormatter) formatImport(w io.Writer, imp *ir.DistilledImport, indent string) error {
	fmt.Fprintf(w, "%simport %s\n", indent, imp.Module)
	return nil
}

func (f *SwiftFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent string) error {
	lines := strings.Split(comment.Text, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Fprintf(w, "%s// %s\n", indent, line)
		} else {
			fmt.Fprintf(w, "%s//\n", indent)
		}
	}
	return nil
}

func (f *SwiftFormatter) formatClass(w io.Writer, class *ir.DistilledClass, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Add blank line before class
	fmt.Fprintln(w)
	
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(class.Visibility)
	
	// Format modifiers
	modifiers := ""
	for _, mod := range class.Modifiers {
		switch mod {
		case ir.ModifierFinal:
			modifiers += "final "
		// case ir.ModifierOpen:
		//	modifiers += "open "
		}
	}
	
	// Format class declaration
	fmt.Fprintf(w, "%s%s%sclass %s", indentStr, visPrefix, modifiers, class.Name)
	
	// Add generic type parameters
	if len(class.TypeParams) > 0 {
		typeParams := make([]string, len(class.TypeParams))
		for i, param := range class.TypeParams {
			typeParams[i] = param.Name
			if len(param.Constraints) > 0 {
				typeParams[i] += ": " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
	}
	
	// Add inheritance
	var inheritance []string
	if len(class.Extends) > 0 {
		for _, ext := range class.Extends {
			inheritance = append(inheritance, ext.Name)
		}
	}
	if len(class.Implements) > 0 {
		for _, impl := range class.Implements {
			inheritance = append(inheritance, impl.Name)
		}
	}
	
	if len(inheritance) > 0 {
		fmt.Fprintf(w, ": %s", strings.Join(inheritance, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format class members
	for _, child := range class.Children {
		f.FormatNode(w, child, indent+1)
	}
	
	return nil
}

func (f *SwiftFormatter) formatProtocol(w io.Writer, intf *ir.DistilledInterface, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Add blank line before protocol
	fmt.Fprintln(w)
	
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(intf.Visibility)
	
	// Format protocol declaration
	fmt.Fprintf(w, "%s%sprotocol %s", indentStr, visPrefix, intf.Name)
	
	// Add generic type parameters
	if len(intf.TypeParams) > 0 {
		typeParams := make([]string, len(intf.TypeParams))
		for i, param := range intf.TypeParams {
			typeParams[i] = param.Name
			if len(param.Constraints) > 0 {
				typeParams[i] += ": " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
	}
	
	// Add inheritance
	if len(intf.Extends) > 0 {
		extends := make([]string, len(intf.Extends))
		for i, ext := range intf.Extends {
			extends[i] = ext.Name
		}
		fmt.Fprintf(w, ": %s", strings.Join(extends, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format protocol members
	for _, child := range intf.Children {
		f.FormatNode(w, child, indent+1)
	}
	
	return nil
}

func (f *SwiftFormatter) formatFunction(w io.Writer, fn *ir.DistilledFunction, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(fn.Visibility)
	
	// Format modifiers
	modifiers := ""
	for _, mod := range fn.Modifiers {
		switch mod {
		case ir.ModifierStatic:
			modifiers += "static "
		case ir.ModifierMutating:
			modifiers += "mutating "
		case ir.ModifierOverride:
			modifiers += "override "
		case ir.ModifierFinal:
			modifiers += "final "
		case ir.ModifierAsync:
			modifiers += "async "
		// case ir.ModifierThrows:
		//	modifiers += "throws "
		}
	}
	
	// Check if it's an initializer
	if fn.Name == "init" {
		fmt.Fprintf(w, "%s%s%sinit", indent, visPrefix, modifiers)
	} else {
		fmt.Fprintf(w, "%s%s%sfunc %s", indent, visPrefix, modifiers, fn.Name)
	}
	
	// Add generic type parameters
	if len(fn.TypeParams) > 0 {
		typeParams := make([]string, len(fn.TypeParams))
		for i, param := range fn.TypeParams {
			typeParams[i] = param.Name
			if len(param.Constraints) > 0 {
				typeParams[i] += ": " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
	}
	
	// Parameters
	fmt.Fprintf(w, "(")
	f.formatParameters(w, fn.Parameters)
	fmt.Fprintf(w, ")")
	
	// Return type
	if fn.Returns != nil && fn.Returns.Name != "" && fn.Returns.Name != "Void" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Implementation
	if fn.Implementation != "" {
		fmt.Fprintln(w, ":")
		impl := strings.TrimSpace(fn.Implementation)
		for _, line := range strings.Split(impl, "\n") {
			fmt.Fprintf(w, "%s    %s\n", indent, line)
		}
	} else {
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *SwiftFormatter) formatField(w io.Writer, field *ir.DistilledField, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(field.Visibility)
	
	// Determine var/let
	varType := "var"
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierFinal || mod == ir.ModifierReadonly {
			varType = "let"
			break
		}
	}
	
	// Check for static
	isStatic := false
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierStatic {
			isStatic = true
			break
		}
	}
	
	// Format field
	if isStatic {
		fmt.Fprintf(w, "%s%sstatic %s %s", indent, visPrefix, varType, field.Name)
	} else {
		fmt.Fprintf(w, "%s%s%s %s", indent, visPrefix, varType, field.Name)
	}
	
	// Add type if specified
	if field.Type != nil && field.Type.Name != "" {
		fmt.Fprintf(w, ": %s", field.Type.Name)
	}
	
	// Add default value if specified
	if field.DefaultValue != "" {
		fmt.Fprintf(w, " = %s", field.DefaultValue)
	}
	
	fmt.Fprintln(w)
	return nil
}

func (f *SwiftFormatter) formatTypeAlias(w io.Writer, alias *ir.DistilledTypeAlias, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(alias.Visibility)
	
	fmt.Fprintf(w, "\n%s%stypealias %s", indent, visPrefix, alias.Name)
	
	// Add generic type parameters
	if len(alias.TypeParams) > 0 {
		typeParams := make([]string, len(alias.TypeParams))
		for i, param := range alias.TypeParams {
			typeParams[i] = param.Name
			if len(param.Constraints) > 0 {
				typeParams[i] += ": " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
	}
	
	fmt.Fprintf(w, " = %s\n", alias.Type.Name)
	
	return nil
}

func (f *SwiftFormatter) formatParameters(w io.Writer, params []ir.Parameter) {
	for i, param := range params {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		
		// External parameter name would go here if supported
		// if param.ExternalName != "" && param.ExternalName != param.Name {
		//	if param.ExternalName == "_" {
		//		fmt.Fprintf(w, "_ ")
		//	} else {
		//		fmt.Fprintf(w, "%s ", param.ExternalName)
		//	}
		// }
		
		// Parameter name
		fmt.Fprintf(w, "%s", param.Name)
		
		// Type
		if param.Type.Name != "" {
			fmt.Fprintf(w, ": %s", param.Type.Name)
		}
		
		// Default value
		if param.DefaultValue != "" {
			fmt.Fprintf(w, " = %s", param.DefaultValue)
		}
		
		// Variadic
		if param.IsVariadic {
			fmt.Fprintf(w, "...")
		}
	}
}

func (f *SwiftFormatter) formatEnum(w io.Writer, enum *ir.DistilledEnum, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Add blank line before enum
	fmt.Fprintln(w)
	
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(enum.Visibility)
	
	// Format enum declaration
	fmt.Fprintf(w, "%s%senum %s", indentStr, visPrefix, enum.Name)
	
	// Check if enum has a raw value type
	if enum.Type != nil && enum.Type.Name != "" {
		fmt.Fprintf(w, ": %s", enum.Type.Name)
	}
	
	fmt.Fprintln(w, ":")
	
	// Format enum cases
	for _, child := range enum.Children {
		switch c := child.(type) {
		case *ir.DistilledField:
			// Enum case
			fmt.Fprintf(w, "%s    case %s", indentStr, c.Name)
			if c.DefaultValue != "" {
				fmt.Fprintf(w, " = %s", c.DefaultValue)
			} else if c.Type != nil && c.Type.Name != "" {
				// Associated values
				fmt.Fprintf(w, "%s", c.Type.Name)
			}
			fmt.Fprintln(w)
		default:
			// Other members (methods, etc.)
			f.FormatNode(w, child, indent+1)
		}
	}
	
	return nil
}

func (f *SwiftFormatter) getVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "*"
	case ir.VisibilityPublic:
		return "" // No prefix for public
	case ir.VisibilityInternal:
		return "~"
	case ir.VisibilityFilePrivate:
		return "-" // File-private is similar to private
	case ir.VisibilityOpen:
		return "" // Open is similar to public
	default:
		return ""
	}
}