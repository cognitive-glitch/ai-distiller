package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// CSharpFormatter formats IR nodes as C# code
type CSharpFormatter struct {
	BaseLanguageFormatter
}

// NewCSharpFormatter creates a new C# formatter
func NewCSharpFormatter() *CSharpFormatter {
	return &CSharpFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("csharp"),
	}
}

// FormatNode formats an IR node as C# code
func (f *CSharpFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	switch n := node.(type) {
	case *ir.DistilledImport:
		_, err := fmt.Fprintln(w, f.formatImport(n))
		return err
	case *ir.DistilledClass:
		// Format class declaration
		_, err := fmt.Fprintln(w, f.formatClass(n, indent))
		if err != nil {
			return err
		}
		// Format class members
		for _, child := range n.Children {
			if err := f.FormatNode(w, child, indent+1); err != nil {
				return err
			}
		}
		return nil
	case *ir.DistilledInterface:
		// Format interface declaration
		_, err := fmt.Fprintln(w, f.formatInterface(n, indent))
		if err != nil {
			return err
		}
		// Format interface members
		for _, child := range n.Children {
			if err := f.FormatNode(w, child, indent+1); err != nil {
				return err
			}
		}
		return nil
	case *ir.DistilledStruct:
		// Format struct declaration
		_, err := fmt.Fprintln(w, f.formatStruct(n, indent))
		if err != nil {
			return err
		}
		// Format struct members
		for _, child := range n.Children {
			if err := f.FormatNode(w, child, indent+1); err != nil {
				return err
			}
		}
		return nil
	case *ir.DistilledEnum:
		// Format enum declaration
		_, err := fmt.Fprintln(w, f.formatEnum(n, indent))
		if err != nil {
			return err
		}
		// Format enum members
		for _, child := range n.Children {
			if err := f.FormatNode(w, child, indent+1); err != nil {
				return err
			}
		}
		return nil
	case *ir.DistilledFunction:
		_, err := fmt.Fprintln(w, f.formatFunction(n, indent))
		return err
	case *ir.DistilledField:
		_, err := fmt.Fprintln(w, f.formatField(n, indent))
		return err
	case *ir.DistilledPackage:
		// Format package/namespace declaration
		indentStr := strings.Repeat("    ", indent)
		fmt.Fprintf(w, "%snamespace %s\n", indentStr, n.Name)
		// Format package members
		for _, child := range n.Children {
			if err := f.FormatNode(w, child, indent); err != nil {
				return err
			}
		}
		return nil
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

func (f *CSharpFormatter) formatImport(imp *ir.DistilledImport) string {
	// Check if this is an aliased import via symbols
	if len(imp.Symbols) == 1 && imp.Symbols[0].Alias != "" {
		return fmt.Sprintf("using %s = %s.%s", imp.Symbols[0].Alias, imp.Module, imp.Symbols[0].Name)
	}
	return fmt.Sprintf("using %s", imp.Module)
}

func (f *CSharpFormatter) formatClass(class *ir.DistilledClass, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := getVisibilityPrefix(class.Visibility)
	
	// Collect non-visibility modifiers
	modifiers := []string{}

	// Check modifiers
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierSealed {
			modifiers = append(modifiers, "sealed")
		} else if mod == ir.ModifierPartial {
			modifiers = append(modifiers, "partial")
		}
	}

	// Class declaration
	classDecl := fmt.Sprintf("%s%s", indentStr, visPrefix)
	if len(modifiers) > 0 {
		classDecl += strings.Join(modifiers, " ") + " "
	}
	classDecl += "class " + class.Name

	// Generics
	if len(class.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range class.TypeParams {
			genericParams = append(genericParams, g.Name)
		}
		classDecl += "<" + strings.Join(genericParams, ", ") + ">"
	}

	// Inheritance
	if len(class.Extends) > 0 || len(class.Implements) > 0 {
		bases := []string{}
		for _, ext := range class.Extends {
			bases = append(bases, ext.Name)
		}
		for _, impl := range class.Implements {
			bases = append(bases, impl.Name)
		}
		classDecl += " : " + strings.Join(bases, ", ")
	}

	// Add colon for Python-like syntax
	return classDecl + ":"
}

func (f *CSharpFormatter) formatInterface(intf *ir.DistilledInterface, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := getVisibilityPrefix(intf.Visibility)
	
	// Interface declaration
	intfDecl := fmt.Sprintf("%s%sinterface %s", indentStr, visPrefix, intf.Name)

	// Generics
	if len(intf.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range intf.TypeParams {
			genericParams = append(genericParams, g.Name)
		}
		intfDecl += "<" + strings.Join(genericParams, ", ") + ">"
	}

	// Extends
	if len(intf.Extends) > 0 {
		extends := []string{}
		for _, ext := range intf.Extends {
			extends = append(extends, ext.Name)
		}
		intfDecl += " : " + strings.Join(extends, ", ")
	}

	return intfDecl + ":"
}

func (f *CSharpFormatter) formatStruct(strct *ir.DistilledStruct, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := f.addVisibilityPrefix(strct.Visibility)
	
	// Struct declaration
	structDecl := fmt.Sprintf("%s%sstruct %s", indentStr, visPrefix, strct.Name)

	// Generics
	if len(strct.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range strct.TypeParams {
			genericParams = append(genericParams, g.Name)
		}
		structDecl += "<" + strings.Join(genericParams, ", ") + ">"
	}

	return structDecl + ":"
}

func (f *CSharpFormatter) formatEnum(enum *ir.DistilledEnum, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := f.addVisibilityPrefix(enum.Visibility)
	
	// Enum declaration
	enumDecl := fmt.Sprintf("%s%senum %s", indentStr, visPrefix, enum.Name)

	// Base type
	if enum.Type != nil {
		enumDecl += " : " + enum.Type.Name
	}

	return enumDecl + ":"
}

func (f *CSharpFormatter) formatFunction(fn *ir.DistilledFunction, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := f.addVisibilityPrefix(fn.Visibility)
	
	// Collect non-visibility modifiers
	modifiers := []string{}

	// Check modifiers
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierAsync {
			modifiers = append(modifiers, "async")
		} else if mod == ir.ModifierVirtual {
			modifiers = append(modifiers, "virtual")
		} else if mod == ir.ModifierOverride {
			modifiers = append(modifiers, "override")
		} else if mod == ir.ModifierSealed {
			modifiers = append(modifiers, "sealed")
		}
	}

	// Return type
	returnType := "void"
	if fn.Returns != nil {
		returnType = fn.Returns.Name
	}

	// Function signature
	signature := fmt.Sprintf("%s%s", indentStr, visPrefix)
	if len(modifiers) > 0 {
		signature += strings.Join(modifiers, " ") + " "
	}
	signature += returnType + " " + fn.Name

	// Generics
	if len(fn.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range fn.TypeParams {
			genericParams = append(genericParams, g.Name)
		}
		signature += "<" + strings.Join(genericParams, ", ") + ">"
	}

	// Parameters
	params := []string{}
	for _, p := range fn.Parameters {
		param := ""
		if p.Type.Name != "" {
			param = p.Type.Name + " " + p.Name
		} else {
			param = "dynamic " + p.Name
		}
		if p.DefaultValue != "" {
			param += " = " + p.DefaultValue
		}
		params = append(params, param)
	}
	signature += "(" + strings.Join(params, ", ") + ")"

	return signature
}

func (f *CSharpFormatter) formatField(field *ir.DistilledField, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Get visibility prefix
	visPrefix := f.addVisibilityPrefix(field.Visibility)
	
	// Collect non-visibility modifiers
	modifiers := []string{}

	// Check modifiers
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierReadonly {
			modifiers = append(modifiers, "readonly")
		} else if mod == ir.ModifierConst {
			modifiers = append(modifiers, "const")
		}
	}

	// Field declaration
	fieldDecl := fmt.Sprintf("%s%s", indentStr, visPrefix)
	if len(modifiers) > 0 {
		fieldDecl += strings.Join(modifiers, " ") + " "
	}

	// Type
	if field.Type != nil && field.Type.Name != "" {
		fieldDecl += field.Type.Name + " "
	} else {
		fieldDecl += "dynamic "
	}

	fieldDecl += field.Name

	// Initializer
	if field.DefaultValue != "" {
		fieldDecl += " = " + field.DefaultValue
	}

	// Remove semicolon for Python-like syntax
	// fieldDecl += ";"

	return fieldDecl
}

func (f *CSharpFormatter) addVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPublic:
		return "" // No prefix for public
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "*"
	case ir.VisibilityInternal:
		return "~"
	case ir.VisibilityProtectedInternal:
		return "*~" // C# protected internal
	case ir.VisibilityPrivateProtected:
		return "-*" // C# private protected
	default:
		return "~" // default to internal
	}
}
