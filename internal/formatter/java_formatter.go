package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// JavaFormatter implements language-specific formatting for Java
type JavaFormatter struct {
	BaseLanguageFormatter
}

// NewJavaFormatter creates a new Java formatter
func NewJavaFormatter() *JavaFormatter {
	return &JavaFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("java"),
	}
}

// FormatNode implements LanguageFormatter
func (f *JavaFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
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
	case *ir.DistilledEnum:
		return f.formatEnum(w, n, indent)
	case *ir.DistilledPackage:
		return f.formatPackage(w, n, indentStr)
	default:
		// Skip unknown nodes
		return nil
	}
}

func (f *JavaFormatter) formatImport(w io.Writer, imp *ir.DistilledImport, indent string) error {
	// Java imports are simpler than Python
	if imp.Module != "" {
		fmt.Fprintf(w, "%simport %s\n", indent, imp.Module)
	}
	return nil
}

func (f *JavaFormatter) formatPackage(w io.Writer, pkg *ir.DistilledPackage, indent string) error {
	fmt.Fprintf(w, "%spackage %s\n", indent, pkg.Name)
	return nil
}

func (f *JavaFormatter) formatClass(w io.Writer, class *ir.DistilledClass, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format class declaration
	fmt.Fprintf(w, "\n%s", indentStr)
	
	// Add visibility prefix
	visPrefix := getVisibilityPrefix(class.Visibility)
	fmt.Fprintf(w, "%s", visPrefix)
	
	// Add modifiers (but not visibility keywords)
	for _, mod := range class.Modifiers {
		switch mod {
		case ir.ModifierStatic:
			fmt.Fprintf(w, "static ")
		case ir.ModifierFinal:
			fmt.Fprintf(w, "final ")
		case ir.ModifierAbstract:
			fmt.Fprintf(w, "abstract ")
		}
	}
	
	// If it's a record (has ModifierData), format as record
	isRecord := false
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierData {
			isRecord = true
			break
		}
	}
	
	if isRecord {
		fmt.Fprintf(w, "record %s", class.Name)
	} else {
		fmt.Fprintf(w, "class %s", class.Name)
	}
	
	// Add generics if present
	if len(class.TypeParams) > 0 {
		params := make([]string, len(class.TypeParams))
		for i, param := range class.TypeParams {
			params[i] = param.Name
			if len(param.Constraints) > 0 {
				params[i] += " extends " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(params, ", "))
	}
	
	// Add extends
	if len(class.Extends) > 0 {
		fmt.Fprintf(w, " extends %s", class.Extends[0].Name)
	}
	
	// Add implements
	if len(class.Implements) > 0 {
		implements := make([]string, len(class.Implements))
		for i, impl := range class.Implements {
			implements[i] = impl.Name
		}
		fmt.Fprintf(w, " implements %s", strings.Join(implements, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format class body
	for _, child := range class.Children {
		if err := f.FormatNode(w, child, indent+1); err != nil {
			return err
		}
	}
	
	return nil
}

func (f *JavaFormatter) formatInterface(w io.Writer, intf *ir.DistilledInterface, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format interface declaration
	fmt.Fprintf(w, "\n%s", indentStr)
	
	// Add visibility prefix
	visPrefix := getVisibilityPrefix(intf.Visibility)
	fmt.Fprintf(w, "%s", visPrefix)
	
	fmt.Fprintf(w, "interface %s", intf.Name)
	
	// Add generics if present
	if len(intf.TypeParams) > 0 {
		params := make([]string, len(intf.TypeParams))
		for i, param := range intf.TypeParams {
			params[i] = param.Name
		}
		fmt.Fprintf(w, "<%s>", strings.Join(params, ", "))
	}
	
	// Add extends
	if len(intf.Extends) > 0 {
		extends := make([]string, len(intf.Extends))
		for i, ext := range intf.Extends {
			extends[i] = ext.Name
		}
		fmt.Fprintf(w, " extends %s", strings.Join(extends, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format interface body
	for _, child := range intf.Children {
		if err := f.FormatNode(w, child, indent+1); err != nil {
			return err
		}
	}
	
	return nil
}

func (f *JavaFormatter) formatEnum(w io.Writer, enum *ir.DistilledEnum, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format enum declaration
	fmt.Fprintf(w, "\n%s", indentStr)
	
	// Add visibility prefix
	visPrefix := getVisibilityPrefix(enum.Visibility)
	fmt.Fprintf(w, "%s", visPrefix)
	
	fmt.Fprintf(w, "enum %s", enum.Name)
	
	// Enums in IR don't have Implements field
	// TODO: Add support if needed
	
	fmt.Fprintln(w, ":")
	
	// Format enum values and body
	for _, child := range enum.Children {
		if err := f.FormatNode(w, child, indent+1); err != nil {
			return err
		}
	}
	
	return nil
}

func (f *JavaFormatter) formatFunction(w io.Writer, fn *ir.DistilledFunction, indentStr string) error {
	// Format annotations
	for _, dec := range fn.Decorators {
		// Remove @ if already present to avoid @@
		decStr := strings.TrimPrefix(dec, "@")
		fmt.Fprintf(w, "%s@%s\n", indentStr, decStr)
	}
	
	// Format method signature
	fmt.Fprintf(w, "%s", indentStr)
	
	// Add visibility prefix
	visPrefix := getVisibilityPrefix(fn.Visibility)
	fmt.Fprintf(w, "%s", visPrefix)
	
	// Add modifiers (but not visibility keywords)
	for _, mod := range fn.Modifiers {
		switch mod {
		case ir.ModifierStatic:
			fmt.Fprintf(w, "static ")
		case ir.ModifierFinal:
			fmt.Fprintf(w, "final ")
		case ir.ModifierAbstract:
			fmt.Fprintf(w, "abstract ")
		// Synchronize and native are not in IR yet
		}
	}
	
	// Add generics if present
	if len(fn.TypeParams) > 0 {
		params := make([]string, len(fn.TypeParams))
		for i, param := range fn.TypeParams {
			params[i] = param.Name
			if len(param.Constraints) > 0 {
				bounds := make([]string, len(param.Constraints))
				for j, bound := range param.Constraints {
					bounds[j] = bound.Name
				}
				params[i] += " extends " + strings.Join(bounds, " & ")
			}
		}
		fmt.Fprintf(w, "<%s> ", strings.Join(params, ", "))
	}
	
	// Add return type (constructors don't have return types)
	if fn.Returns != nil && fn.Name != "<init>" {
		fmt.Fprintf(w, "%s ", formatTypeRef(fn.Returns))
	}
	
	// Method name
	fmt.Fprintf(w, "%s(", fn.Name)
	
	// Parameters
	params := make([]string, len(fn.Parameters))
	for i, param := range fn.Parameters {
		paramStr := ""
		// Add parameter decorators (annotations)
		for _, dec := range param.Decorators {
			paramStr += "@" + dec + " "
		}
		paramStr += formatTypeRef(&param.Type) + " "
		paramStr += param.Name
		params[i] = paramStr
	}
	fmt.Fprintf(w, "%s)", strings.Join(params, ", "))
	
	// Throws clause
	if len(fn.Throws) > 0 {
		throws := make([]string, len(fn.Throws))
		for i, t := range fn.Throws {
			throws[i] = t.Name
		}
		fmt.Fprintf(w, " throws %s", strings.Join(throws, ", "))
	}
	
	// Format implementation
	if fn.Implementation != "" {
		fmt.Fprintln(w, ":")
		lines := strings.Split(strings.TrimSpace(fn.Implementation), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Fprintf(w, "%s    %s\n", indentStr, line)
			}
		}
	} else {
		// No implementation - abstract method or interface method
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *JavaFormatter) formatField(w io.Writer, field *ir.DistilledField, indent string) error {
	// Format field with Java syntax
	fmt.Fprintf(w, "%s", indent)
	
	// Add visibility prefix
	visPrefix := getVisibilityPrefix(field.Visibility)
	fmt.Fprintf(w, "%s", visPrefix)
	
	// Add modifiers (but not visibility keywords)
	for _, mod := range field.Modifiers {
		switch mod {
		case ir.ModifierStatic:
			fmt.Fprintf(w, "static ")
		case ir.ModifierFinal:
			fmt.Fprintf(w, "final ")
		case ir.ModifierVolatile:
			fmt.Fprintf(w, "volatile ")
		case ir.ModifierTransient:
			fmt.Fprintf(w, "transient ")
		}
	}
	
	// Add type
	if field.Type != nil {
		fmt.Fprintf(w, "%s ", formatTypeRef(field.Type))
	}
	
	// Add name
	fmt.Fprintf(w, "%s", field.Name)
	
	// Add default value if present
	if field.DefaultValue != "" {
		fmt.Fprintf(w, " = %s", field.DefaultValue)
	}
	
	fmt.Fprintln(w, ";")
	return nil
}

func (f *JavaFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent string) error {
	lines := strings.Split(comment.Text, "\n")
	if comment.Format == "doc" || strings.HasPrefix(comment.Text, "/**") {
		// Javadoc comment
		fmt.Fprintf(w, "%s/**\n", indent)
		for _, line := range lines {
			if !strings.HasPrefix(line, "/**") && !strings.HasSuffix(line, "*/") {
				fmt.Fprintf(w, "%s * %s\n", indent, line)
			}
		}
		fmt.Fprintf(w, "%s */\n", indent)
	} else if len(lines) > 1 {
		// Multi-line comment
		fmt.Fprintf(w, "%s/*\n", indent)
		for _, line := range lines {
			fmt.Fprintf(w, "%s * %s\n", indent, line)
		}
		fmt.Fprintf(w, "%s */\n", indent)
	} else {
		// Single line comment
		fmt.Fprintf(w, "%s// %s\n", indent, comment.Text)
	}
	
	return nil
}

// formatTypeRef formats a type reference for Java
func formatTypeRef(ref *ir.TypeRef) string {
	if ref == nil {
		return ""
	}
	
	result := ref.Name
	
	// Add generics if present
	if len(ref.TypeArgs) > 0 {
		args := make([]string, len(ref.TypeArgs))
		for i, arg := range ref.TypeArgs {
			args[i] = formatTypeRef(&arg)
		}
		result += "<" + strings.Join(args, ", ") + ">"
	}
	
	// Handle array types
	if ref.IsArray {
		result += "[]"
	}
	
	return result
}

