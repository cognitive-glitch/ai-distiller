package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// PHPFormatter formats IR nodes as PHP code
type PHPFormatter struct {
	BaseLanguageFormatter
}

// NewPHPFormatter creates a new PHP formatter
func NewPHPFormatter() *PHPFormatter {
	return &PHPFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("php"),
	}
}

// FormatNode formats an IR node as PHP code
func (f *PHPFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
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

func (f *PHPFormatter) formatImport(imp *ir.DistilledImport) string {
	if imp.ImportType == "namespace" {
		return fmt.Sprintf("namespace %s;", imp.Module)
	}
	
	// Use statement with alias
	if len(imp.Symbols) == 1 && imp.Symbols[0].Alias != "" {
		return fmt.Sprintf("use %s\\%s as %s;", imp.Module, imp.Symbols[0].Name, imp.Symbols[0].Alias)
	}
	
	// Check for function/const imports
	if strings.Contains(imp.Module, "function ") {
		return fmt.Sprintf("use %s;", imp.Module)
	}
	if strings.Contains(imp.Module, "const ") {
		return fmt.Sprintf("use %s;", imp.Module)
	}
	
	return fmt.Sprintf("use %s;", imp.Module)
}

func (f *PHPFormatter) formatClass(class *ir.DistilledClass, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	var parts []string

	// Class modifiers
	modifiers := []string{}
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierFinal {
			modifiers = append(modifiers, "final")
		}
	}

	// Class type is always "class" for DistilledClass
	classType := "class"

	// Class declaration
	classDecl := ""
	if len(modifiers) > 0 {
		classDecl = strings.Join(modifiers, " ") + " "
	}
	classDecl += classType + " " + class.Name

	// Inheritance
	if len(class.Extends) > 0 {
		// Classes can only extend one class
		classDecl += " extends " + class.Extends[0].Name
	}

	// Implements
	if len(class.Implements) > 0 {
		implements := []string{}
		for _, impl := range class.Implements {
			implements = append(implements, impl.Name)
		}
		classDecl += " implements " + strings.Join(implements, ", ")
	}

	parts = append(parts, f.addVisibilityPrefix(class.Visibility)+indentStr+classDecl+" {")

	return strings.Join(parts, "\n")
}

func (f *PHPFormatter) formatInterface(intf *ir.DistilledInterface, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	var parts []string

	// Interface declaration
	intfDecl := "interface " + intf.Name

	// Extends
	if len(intf.Extends) > 0 {
		extends := []string{}
		for _, ext := range intf.Extends {
			extends = append(extends, ext.Name)
		}
		intfDecl += " extends " + strings.Join(extends, ", ")
	}

	parts = append(parts, f.addVisibilityPrefix(intf.Visibility)+indentStr+intfDecl+" {")

	return strings.Join(parts, "\n")
}

func (f *PHPFormatter) formatEnum(enum *ir.DistilledEnum, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Enum declaration (PHP 8.1+)
	enumDecl := "enum " + enum.Name

	// Base type (backed enum)
	if enum.Type != nil && enum.Type.Name != "" {
		enumDecl += ": " + enum.Type.Name
	}

	return f.addVisibilityPrefix(enum.Visibility) + indentStr + enumDecl + " {"  
}

func (f *PHPFormatter) formatFunction(fn *ir.DistilledFunction, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if fn.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if fn.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if fn.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	}

	// Check modifiers
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierAbstract {
			modifiers = append(modifiers, "abstract")
		} else if mod == ir.ModifierFinal {
			modifiers = append(modifiers, "final")
		}
	}

	// Function signature
	signature := ""
	if len(modifiers) > 0 {
		signature = strings.Join(modifiers, " ") + " "
	}
	signature += "function " + fn.Name

	// Parameters
	params := []string{}
	for _, p := range fn.Parameters {
		param := ""
		
		// Type hint
		if p.Type.Name != "" {
			typeName := p.Type.Name
			// Handle nullable types
			if p.IsOptional && !strings.HasPrefix(typeName, "?") {
				typeName = "?" + typeName
			}
			param = typeName + " "
		}
		
		// Parameter name (PHP variables start with $)
		paramName := p.Name
		if !strings.HasPrefix(paramName, "$") {
			paramName = "$" + paramName
		}
		param += paramName
		
		// Default value
		if p.DefaultValue != "" {
			param += " = " + p.DefaultValue
		}
		
		params = append(params, param)
	}
	signature += "(" + strings.Join(params, ", ") + ")"

	// Return type
	if fn.Returns != nil {
		returnType := fn.Returns.Name
		if returnType == "never" {
			returnType = "never"
		} else if returnType == "" {
			returnType = "void"
		}
		signature += ": " + returnType
	}

	return f.addVisibilityPrefix(fn.Visibility) + indentStr + signature
}

func (f *PHPFormatter) formatField(field *ir.DistilledField, indent int) string {
	indentStr := strings.Repeat("    ", indent)

	// Access modifiers
	modifiers := []string{}
	if field.Visibility == ir.VisibilityPublic {
		modifiers = append(modifiers, "public")
	} else if field.Visibility == ir.VisibilityProtected {
		modifiers = append(modifiers, "protected")
	} else if field.Visibility == ir.VisibilityPrivate {
		modifiers = append(modifiers, "private")
	}

	// Check modifiers
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierStatic {
			modifiers = append(modifiers, "static")
		} else if mod == ir.ModifierReadonly {
			modifiers = append(modifiers, "readonly")
		}
	}

	// Field declaration
	fieldDecl := ""
	if len(modifiers) > 0 {
		fieldDecl = strings.Join(modifiers, " ") + " "
	}

	// Type (PHP 7.4+ typed properties)
	if field.Type != nil && field.Type.Name != "" {
		fieldDecl += field.Type.Name + " "
	}

	// Property name (PHP variables start with $)
	fieldName := field.Name
	if !strings.HasPrefix(fieldName, "$") {
		fieldName = "$" + fieldName
	}
	fieldDecl += fieldName

	// Initializer
	if field.DefaultValue != "" {
		fieldDecl += " = " + field.DefaultValue
	}

	fieldDecl += ";"

	return f.addVisibilityPrefix(field.Visibility) + indentStr + fieldDecl
}

func (f *PHPFormatter) addVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPublic:
		return "+ "
	case ir.VisibilityPrivate:
		return "- "
	case ir.VisibilityProtected:
		return "# "
	default:
		return "+ " // Default is public in PHP
	}
}
