package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// CppFormatter formats IR nodes as C++ code
type CppFormatter struct {
	BaseLanguageFormatter
}

// NewCppFormatter creates a new C++ formatter
func NewCppFormatter() *CppFormatter {
	return &CppFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("cpp"),
	}
}

// FormatNode formats an IR node as C++ code
func (f *CppFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	switch n := node.(type) {
	case *ir.DistilledImport:
		_, err := fmt.Fprintln(w, f.formatImport(n))
		return err
	case *ir.DistilledClass:
		_, err := fmt.Fprintln(w, f.formatClass(n, indent))
		return err
	case *ir.DistilledStruct:
		_, err := fmt.Fprintln(w, f.formatStruct(n, indent))
		return err
	case *ir.DistilledEnum:
		_, err := fmt.Fprintln(w, f.formatEnum(n, indent))
		return err
	case *ir.DistilledFunction:
		_, err := fmt.Fprintln(w, f.formatFunction(n, indent))
		return err
	case *ir.DistilledField:
		_, err := fmt.Fprintln(w, f.formatField(n, indent))
		return err
	default:
		// For nodes with children, process them recursively
		children := node.GetChildren()
		for _, child := range children {
			if err := f.FormatNode(w, child, indent); err != nil {
				return err
			}
		}
		return nil
	}
}

func (f *CppFormatter) formatImport(imp *ir.DistilledImport) string {
	// C++ uses #include
	if strings.HasPrefix(imp.Module, "<") && strings.HasSuffix(imp.Module, ">") {
		return fmt.Sprintf("#include %s", imp.Module)
	} else if strings.HasPrefix(imp.Module, "\"") && strings.HasSuffix(imp.Module, "\"") {
		return fmt.Sprintf("#include %s", imp.Module)
	} else {
		// Assume user header
		return fmt.Sprintf("#include \"%s\"", imp.Module)
	}
}

func (f *CppFormatter) formatClass(class *ir.DistilledClass, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	var parts []string

	// Class type is always "class" for DistilledClass
	classType := "class"

	// Template parameters
	if len(class.TypeParams) > 0 {
		templateParams := []string{}
		for _, g := range class.TypeParams {
			templateParam := "typename " + g.Name
			if len(g.Constraints) > 0 {
				// C++ concepts or requires clause
				templateParam = g.Constraints[0].Name + " " + g.Name
			}
			templateParams = append(templateParams, templateParam)
		}
		parts = append(parts, indentStr+"template<"+strings.Join(templateParams, ", ")+">")
	}

	// Class declaration
	classDecl := classType + " " + class.Name

	// Inheritance
	if len(class.Extends) > 0 || len(class.Implements) > 0 {
		bases := []string{}
		for _, ext := range class.Extends {
			// Assume public inheritance by default
			base := "public " + ext.Name
			bases = append(bases, base)
		}
		// C++ doesn't distinguish implements, all are base classes
		for _, impl := range class.Implements {
			bases = append(bases, "public "+impl.Name)
		}
		classDecl += " : " + strings.Join(bases, ", ")
	}

	parts = append(parts, indentStr+classDecl+" {")

	// Format children (methods, fields, etc.)
	for _, child := range class.Children {
		switch n := child.(type) {
		case *ir.DistilledFunction:
			parts = append(parts, f.formatFunction(n, indent+1))
		case *ir.DistilledField:
			parts = append(parts, f.formatField(n, indent+1))
		}
	}

	return strings.Join(parts, "\n")
}

func (f *CppFormatter) formatStruct(strct *ir.DistilledStruct, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	var parts []string

	// Template parameters
	if len(strct.TypeParams) > 0 {
		templateParams := []string{}
		for _, g := range strct.TypeParams {
			templateParam := "typename " + g.Name
			if len(g.Constraints) > 0 {
				templateParam = g.Constraints[0].Name + " " + g.Name
			}
			templateParams = append(templateParams, templateParam)
		}
		parts = append(parts, indentStr+"template<"+strings.Join(templateParams, ", ")+">")
	}

	// Struct declaration
	structDecl := "struct " + strct.Name

	parts = append(parts, indentStr+structDecl+" {")

	// Format children (methods, fields, etc.)
	for _, child := range strct.Children {
		switch n := child.(type) {
		case *ir.DistilledFunction:
			parts = append(parts, f.formatFunction(n, indent+1))
		case *ir.DistilledField:
			parts = append(parts, f.formatField(n, indent+1))
		}
	}

	return strings.Join(parts, "\n")
}

func (f *CppFormatter) formatEnum(enum *ir.DistilledEnum, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Enum declaration
	enumDecl := "enum class " + enum.Name

	// Base type
	if enum.Type != nil {
		enumDecl += " : " + enum.Type.Name
	}

	var parts []string
	parts = append(parts, indentStr+enumDecl+" {")
	
	// Format enum values
	for _, child := range enum.Children {
		if field, ok := child.(*ir.DistilledField); ok {
			valueStr := indentStr + "    " + field.Name
			if field.DefaultValue != "" {
				valueStr += " = " + field.DefaultValue
			}
			parts = append(parts, valueStr)
		}
	}
	
	return strings.Join(parts, "\n")
}

func (f *CppFormatter) formatFunction(fn *ir.DistilledFunction, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Template parameters
	var templateLine string
	if len(fn.TypeParams) > 0 {
		templateParams := []string{}
		for _, g := range fn.TypeParams {
			templateParam := "typename " + g.Name
			if len(g.Constraints) > 0 {
				templateParam = g.Constraints[0].Name + " " + g.Name
			}
			templateParams = append(templateParams, templateParam)
		}
		templateLine = indentStr + "template<" + strings.Join(templateParams, ", ") + ">\n"
	}

	// Function modifiers
	modifiers := []string{}
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierVirtual {
			modifiers = append(modifiers, "virtual")
		} else if mod == ir.ModifierInline {
			modifiers = append(modifiers, "inline")
		} else if mod == ir.ModifierExtern {
			modifiers = append(modifiers, "extern")
		}
	}

	// Return type
	returnType := "void"
	if fn.Returns != nil {
		returnType = fn.Returns.Name
	}

	// Function signature
	signature := ""
	if len(modifiers) > 0 {
		signature = strings.Join(modifiers, " ") + " "
	}
	signature += returnType + " " + fn.Name

	// Parameters
	params := []string{}
	for _, p := range fn.Parameters {
		param := ""
		if p.Type.Name != "" {
			param = p.Type.Name + " " + p.Name
		} else {
			param = "auto " + p.Name
		}
		if p.DefaultValue != "" {
			param += " = " + p.DefaultValue
		}
		params = append(params, param)
	}
	signature += "(" + strings.Join(params, ", ") + ")"

	// Function qualifiers - check modifiers
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierConst {
			signature += " const"
		} else if mod == ir.ModifierOverride {
			signature += " override"
		} else if mod == ir.ModifierFinal {
			signature += " final"
		} else if mod == ir.ModifierAbstract {
			signature += " = 0"
		}
	}

	// Add implementation if present
	if fn.Implementation != "" {
		signature += " {\n"
		// Strip leading and trailing braces from implementation if present
		impl := fn.Implementation
		lines := strings.Split(impl, "\n")
		
		// Find first and last non-empty lines
		firstNonEmpty := -1
		lastNonEmpty := -1
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				if firstNonEmpty == -1 {
					firstNonEmpty = i
				}
				lastNonEmpty = i
			}
		}
		
		// Check if first and last lines are braces
		if firstNonEmpty >= 0 && lastNonEmpty >= 0 && firstNonEmpty < lastNonEmpty {
			firstLine := strings.TrimSpace(lines[firstNonEmpty])
			lastLine := strings.TrimSpace(lines[lastNonEmpty])
			if firstLine == "{" && lastLine == "}" {
				// Remove brace lines
				lines = lines[firstNonEmpty+1:lastNonEmpty]
			}
		}
		
		// Join back and add
		impl = strings.Join(lines, "\n")
		impl = strings.TrimSpace(impl)
		
		// If implementation is empty after stripping braces, don't add anything
		if impl == "" {
			// Don't add implementation block
		} else {
			signature += impl
			if !strings.HasSuffix(impl, "\n") {
				signature += "\n"
			}
			signature += indentStr + "}"
		}
	}

	// Top-level functions (indent == 0) don't have visibility prefix
	if indent == 0 {
		result := templateLine + signature
		return result
	}
	result := templateLine + indentStr + f.addVisibilityPrefix(fn.Visibility) + signature
	return result
}

func (f *CppFormatter) formatField(field *ir.DistilledField, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Field modifiers
	modifiers := []string{}
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierConst || mod == ir.ModifierReadonly {
			modifiers = append(modifiers, "const")
		} else if mod == ir.ModifierMutable {
			modifiers = append(modifiers, "mutable")
		} else if mod == ir.ModifierVolatile {
			modifiers = append(modifiers, "volatile")
		}
	}

	// Field declaration
	fieldDecl := ""
	if len(modifiers) > 0 {
		fieldDecl = strings.Join(modifiers, " ") + " "
	}

	// Type
	if field.Type != nil && field.Type.Name != "" {
		fieldDecl += field.Type.Name + " "
	} else {
		fieldDecl += "auto "
	}

	fieldDecl += field.Name

	// Initializer
	if field.DefaultValue != "" {
		fieldDecl += " = " + field.DefaultValue
	}

	fieldDecl += ";"

	return indentStr + f.addVisibilityPrefix(field.Visibility) + fieldDecl
}

func (f *CppFormatter) addVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPublic:
		return "" // No prefix for public
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "*"
	case ir.VisibilityInternal:
		return "~"
	default:
		return "-" // Default is private in C++ classes
	}
}
