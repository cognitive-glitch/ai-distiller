package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// KotlinFormatter formats IR nodes as Kotlin code
type KotlinFormatter struct {
	BaseLanguageFormatter
}

// NewKotlinFormatter creates a new Kotlin formatter
func NewKotlinFormatter() *KotlinFormatter {
	return &KotlinFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("kotlin"),
	}
}

// FormatNode formats an IR node as Kotlin code
func (f *KotlinFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	switch n := node.(type) {
	case *ir.DistilledImport:
		_, err := fmt.Fprintln(w, f.formatImport(n))
		return err
	case *ir.DistilledClass:
		_, err := fmt.Fprintln(w, f.formatClass(n, indent))
		return err
	case *ir.DistilledInterface:
		_, err := fmt.Fprintln(w, f.formatInterface(n, indent))
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

func (f *KotlinFormatter) formatImport(imp *ir.DistilledImport) string {
	// Check if this is an aliased import via symbols
	if len(imp.Symbols) == 1 && imp.Symbols[0].Alias != "" {
		return fmt.Sprintf("import %s.%s as %s", imp.Module, imp.Symbols[0].Name, imp.Symbols[0].Alias)
	}
	return fmt.Sprintf("import %s", imp.Module)
}

func (f *KotlinFormatter) formatClass(class *ir.DistilledClass, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	var parts []string

	// Access modifier
	modifiers := []string{}
	if class.Visibility == ir.VisibilityPublic {
		// public is default in Kotlin, don't add it
	} else if class.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if class.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else if class.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Check modifiers
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierFinal {
			modifiers = append(modifiers, "final")
		} else if mod == ir.ModifierData {
			modifiers = append(modifiers, "data")
		} else if mod == ir.ModifierSealed {
			modifiers = append(modifiers, "sealed")
		}
	}

	// Check if it's a data class
	isDataClass := false
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierData {
			isDataClass = true
			break
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
			genericParam := g.Name
			if len(g.Constraints) > 0 {
				genericParam += " : " + g.Constraints[0].Name
			}
			genericParams = append(genericParams, genericParam)
		}
		classDecl += "<" + strings.Join(genericParams, ", ") + ">"
	}

	// Constructor parameters (for data classes)
	if isDataClass && len(class.Children) > 0 {
		params := []string{}
		for _, child := range class.Children {
			if fieldNode, ok := child.(*ir.DistilledField); ok {
				param := ""
				// Check if field has readonly modifier
				hasReadonly := false
				for _, mod := range fieldNode.Modifiers {
					if mod == ir.ModifierReadonly {
						hasReadonly = true
						break
					}
				}
				if hasReadonly {
					param = "val "
				} else {
					param = "var "
				}
				param += fieldNode.Name
				if fieldNode.Type != nil && fieldNode.Type.Name != "" {
					param += ": " + fieldNode.Type.Name
				}
				params = append(params, param)
			}
		}
		if len(params) > 0 {
			classDecl += "(" + strings.Join(params, ", ") + ")"
		}
	}

	// Inheritance
	if len(class.Extends) > 0 || len(class.Implements) > 0 {
		bases := []string{}
		for _, ext := range class.Extends {
			base := ext.Name
			if ext.Name != "Any" { // Don't show default superclass
				bases = append(bases, base+"()")
			}
		}
		for _, impl := range class.Implements {
			bases = append(bases, impl.Name)
		}
		if len(bases) > 0 {
			classDecl += " : " + strings.Join(bases, ", ")
		}
	}

	parts = append(parts, f.addVisibilityPrefix(class.Visibility)+indentStr+classDecl)

	return strings.Join(parts, "\n")
}

func (f *KotlinFormatter) formatInterface(intf *ir.DistilledInterface, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if intf.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else if intf.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}
	// public is default in Kotlin

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
			genericParam := g.Name
			if len(g.Constraints) > 0 {
				genericParam += " : " + g.Constraints[0].Name
			}
			genericParams = append(genericParams, genericParam)
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

	return f.addVisibilityPrefix(intf.Visibility) + indentStr + intfDecl
}

func (f *KotlinFormatter) formatEnum(enum *ir.DistilledEnum, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifier
	modifiers := []string{}
	if enum.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else if enum.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Enum declaration
	enumDecl := strings.Join(modifiers, " ")
	if enumDecl != "" {
		enumDecl += " "
	}
	enumDecl += "enum class " + enum.Name

	return f.addVisibilityPrefix(enum.Visibility) + indentStr + enumDecl
}

func (f *KotlinFormatter) formatFunction(fn *ir.DistilledFunction, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if fn.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if fn.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else if fn.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}
	// public is default in Kotlin

	// Check modifiers
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierAsync {
			modifiers = append(modifiers, "suspend")
		} else if mod == ir.ModifierOverride {
			modifiers = append(modifiers, "override")
		} else if mod == ir.ModifierFinal {
			modifiers = append(modifiers, "final")
		} else if mod == ir.ModifierInline {
			modifiers = append(modifiers, "inline")
		}
	}

	// Function keyword
	funKeyword := "fun"

	// Function signature
	signature := strings.Join(modifiers, " ")
	if signature != "" {
		signature += " "
	}
	signature += funKeyword + " "

	// Generics
	if len(fn.TypeParams) > 0 {
		genericParams := []string{}
		for _, g := range fn.TypeParams {
			genericParam := g.Name
			if len(g.Constraints) > 0 {
				genericParam += " : " + g.Constraints[0].Name
			}
			genericParams = append(genericParams, genericParam)
		}
		signature += "<" + strings.Join(genericParams, ", ") + "> "
	}

	signature += fn.Name

	// Parameters
	params := []string{}
	for _, p := range fn.Parameters {
		param := p.Name
		if p.Type.Name != "" {
			param += ": " + p.Type.Name
		}
		if p.DefaultValue != "" {
			param += " = " + p.DefaultValue
		}
		params = append(params, param)
	}
	signature += "(" + strings.Join(params, ", ") + ")"

	// Return type
	if fn.Returns != nil && fn.Returns.Name != "Unit" {
		signature += ": " + fn.Returns.Name
	}

	return f.addVisibilityPrefix(fn.Visibility) + indentStr + signature
}

func (f *KotlinFormatter) formatField(field *ir.DistilledField, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if field.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if field.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	} else if field.Visibility == ir.VisibilityInternal {
		modifiers = append(modifiers, "internal")
	}

	// Property keyword
	hasReadonly := false
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierReadonly || mod == ir.ModifierFinal {
			hasReadonly = true
			break
		}
	}
	if hasReadonly {
		modifiers = append(modifiers, "val")
	} else {
		modifiers = append(modifiers, "var")
	}

	// Field declaration
	fieldDecl := strings.Join(modifiers, " ") + " " + field.Name

	// Type
	if field.Type != nil {
		fieldDecl += ": " + field.Type.Name
	}

	// Initializer
	if field.DefaultValue != "" {
		fieldDecl += " = " + field.DefaultValue
	}

	return f.addVisibilityPrefix(field.Visibility) + indentStr + fieldDecl
}

func (f *KotlinFormatter) addVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPublic:
		return "+ "
	case ir.VisibilityPrivate:
		return "- "
	case ir.VisibilityProtected:
		return "# "
	case ir.VisibilityInternal:
		return "~ "
	default:
		return "+ " // Default is public in Kotlin
	}
}
