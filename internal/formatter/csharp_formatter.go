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
		return fmt.Sprintf("using %s = %s.%s;", imp.Symbols[0].Alias, imp.Module, imp.Symbols[0].Name)
	}
	return fmt.Sprintf("using %s;", imp.Module)
}

func (f *CSharpFormatter) formatClass(class *ir.DistilledClass, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if class.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if class.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if class.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else {
		modifiers = append(modifiers, "internal")
	}

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
	classDecl := strings.Join(modifiers, " ")
	if classDecl != "" {
		classDecl += " "
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

	return indentStr + f.addVisibilityPrefix(class.Visibility) + classDecl
}

func (f *CSharpFormatter) formatInterface(intf *ir.DistilledInterface, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if intf.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if intf.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Interface declaration
	intfDecl := strings.Join(modifiers, " ")
	if intfDecl != "" {
		intfDecl += " "
	}
	intfDecl += "interface " + intf.Name

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

	return indentStr + f.addVisibilityPrefix(intf.Visibility) + intfDecl
}

func (f *CSharpFormatter) formatStruct(strct *ir.DistilledStruct, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if strct.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if strct.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Struct declaration
	structDecl := strings.Join(modifiers, " ")
	if structDecl != "" {
		structDecl += " "
	}
	structDecl += "struct " + strct.Name

	// Generics
	if len(strct.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range strct.TypeParams {
			genericParams = append(genericParams, g.Name)
		}
		structDecl += "<" + strings.Join(genericParams, ", ") + ">"
	}

	return indentStr + f.addVisibilityPrefix(strct.Visibility) + structDecl
}

func (f *CSharpFormatter) formatEnum(enum *ir.DistilledEnum, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if enum.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if enum.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Enum declaration
	enumDecl := strings.Join(modifiers, " ")
	if enumDecl != "" {
		enumDecl += " "
	}
	enumDecl += "enum " + enum.Name

	// Base type
	if enum.Type != nil {
		enumDecl += " : " + enum.Type.Name
	}

	return indentStr + f.addVisibilityPrefix(enum.Visibility) + enumDecl
}

func (f *CSharpFormatter) formatFunction(fn *ir.DistilledFunction, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if fn.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if fn.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if fn.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else {
		modifiers = append(modifiers, "internal")
	}

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
	signature := strings.Join(modifiers, " ")
	if signature != "" {
		signature += " "
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

	return indentStr + f.addVisibilityPrefix(fn.Visibility) + signature
}

func (f *CSharpFormatter) formatField(field *ir.DistilledField, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if field.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if field.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if field.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else {
		modifiers = append(modifiers, "internal")
	}

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
	fieldDecl := strings.Join(modifiers, " ")
	if fieldDecl != "" {
		fieldDecl += " "
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

	fieldDecl += ";"

	return indentStr + f.addVisibilityPrefix(field.Visibility) + fieldDecl
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
