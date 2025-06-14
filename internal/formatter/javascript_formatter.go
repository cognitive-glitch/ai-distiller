package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// JavaScriptFormatter is a language-specific formatter for JavaScript
type JavaScriptFormatter struct {
	BaseLanguageFormatter
}

// NewJavaScriptFormatter creates a new JavaScript formatter
func NewJavaScriptFormatter() *JavaScriptFormatter {
	return &JavaScriptFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("javascript"),
	}
}

// FormatNode formats a JavaScript node
func (f *JavaScriptFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	switch n := node.(type) {
	case *ir.DistilledImport:
		return f.formatImport(w, n)
	case *ir.DistilledClass:
		return f.formatClass(w, n, indent)
	case *ir.DistilledFunction:
		return f.formatFunction(w, n, indent)
	case *ir.DistilledField:
		return f.formatField(w, n, indent)
	case *ir.DistilledComment:
		return f.formatComment(w, n, indent)
	default:
		// Fallback for unknown nodes
		return nil
	}
}

func (f *JavaScriptFormatter) formatImport(w io.Writer, imp *ir.DistilledImport) error {
	if imp.ImportType == "from" || imp.ImportType == "import" {
		// Handle different import styles
		if len(imp.Symbols) == 0 {
			// Side-effect import: import 'module'
			fmt.Fprintf(w, "import '%s'\n", imp.Module)
		} else {
			// Check for namespace import (* as name)
			hasNamespace := false
			var namespaceAlias string
			for _, sym := range imp.Symbols {
				if sym.Name == "*" && sym.Alias != "" {
					hasNamespace = true
					namespaceAlias = sym.Alias
					break
				}
			}
			
			if hasNamespace {
				fmt.Fprintf(w, "import * as %s from '%s'\n", namespaceAlias, imp.Module)
			} else {
				// Named imports or default import
				defaultImport := ""
				namedImports := []string{}
				
				for _, sym := range imp.Symbols {
					if sym.Alias != "" && sym.Name != "*" {
						namedImports = append(namedImports, fmt.Sprintf("%s as %s", sym.Name, sym.Alias))
					} else if sym.Name != "" && sym.Name != "*" {
						// Check if it's likely a default import (single identifier without braces in original)
						if len(imp.Symbols) == 1 && imp.ImportType == "import" {
							defaultImport = sym.Name
						} else {
							namedImports = append(namedImports, sym.Name)
						}
					}
				}
				
				// Format the import statement
				if defaultImport != "" && len(namedImports) == 0 {
					fmt.Fprintf(w, "import %s from '%s'\n", defaultImport, imp.Module)
				} else if defaultImport != "" && len(namedImports) > 0 {
					fmt.Fprintf(w, "import %s, { %s } from '%s'\n", defaultImport, strings.Join(namedImports, ", "), imp.Module)
				} else if len(namedImports) > 0 {
					fmt.Fprintf(w, "import { %s } from '%s'\n", strings.Join(namedImports, ", "), imp.Module)
				}
			}
		}
	} else {
		// CommonJS require() style - show as comment
		fmt.Fprintf(w, "// require('%s')\n", imp.Module)
	}
	return nil
}

func (f *JavaScriptFormatter) formatClass(w io.Writer, class *ir.DistilledClass, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format visibility prefix
	visPrefix := f.getVisibilityPrefix(class.Visibility)
	
	// Format class declaration
	fmt.Fprintf(w, "\n%s%sclass %s", indentStr, visPrefix, class.Name)
	
	// Add extends clause
	if len(class.Extends) > 0 {
		extends := make([]string, len(class.Extends))
		for i, ext := range class.Extends {
			extends[i] = ext.Name
		}
		fmt.Fprintf(w, " extends %s", strings.Join(extends, ", "))
	}
	
	fmt.Fprintln(w, ":")
	
	// Format class members
	for _, child := range class.Children {
		f.FormatNode(w, child, indent+1)
	}
	
	return nil
}

func (f *JavaScriptFormatter) formatFunction(w io.Writer, fn *ir.DistilledFunction, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format visibility prefix
	visPrefix := f.getVisibilityPrefix(fn.Visibility)
	
	// Handle different function types
	if strings.HasPrefix(fn.Name, "get ") {
		// Getter
		fmt.Fprintf(w, "%s%sget %s()", indentStr, visPrefix, strings.TrimPrefix(fn.Name, "get "))
	} else if strings.HasPrefix(fn.Name, "set ") {
		// Setter
		setterName := strings.TrimPrefix(fn.Name, "set ")
		params := f.formatParameters(fn.Parameters)
		if params == "" {
			params = "value" // Default parameter for setter
		}
		fmt.Fprintf(w, "%s%sset %s(%s)", indentStr, visPrefix, setterName, params)
	} else if fn.Name == "constructor" {
		// Constructor
		fmt.Fprintf(w, "%s%sconstructor(%s)", indentStr, visPrefix, f.formatParameters(fn.Parameters))
	} else {
		// Regular function or method
		// modifiers := ""
		isAsync := false
		isStatic := false
		isGenerator := strings.HasPrefix(fn.Name, "*")
		functionName := fn.Name
		
		if isGenerator {
			functionName = strings.TrimPrefix(functionName, "*")
		}
		
		for _, mod := range fn.Modifiers {
			if mod == ir.ModifierAsync {
				isAsync = true
			} else if mod == ir.ModifierStatic {
				isStatic = true
			}
		}
		
		// Check if it's a top-level const function
		isConst := false
		for _, mod := range fn.Modifiers {
			if mod == ir.ModifierFinal {
				isConst = true
				break
			}
		}
		
		// Format based on context
		if indent == 0 && isConst {
			// Top-level const arrow function
			fmt.Fprintf(w, "%s%sconst %s = (%s)", indentStr, visPrefix, functionName, f.formatParameters(fn.Parameters))
		} else if indent == 0 {
			// Top-level function declaration
			prefixParts := []string{}
			if visPrefix != "" {
				prefixParts = append(prefixParts, strings.TrimSpace(visPrefix))
			}
			if isAsync {
				prefixParts = append(prefixParts, "async")
			}
			if isGenerator {
				prefixParts = append(prefixParts, "*")
			}
			prefix := ""
			if len(prefixParts) > 0 {
				prefix = strings.Join(prefixParts, " ") + " "
			}
			fmt.Fprintf(w, "%s%sfunction %s(%s)", indentStr, prefix, functionName, f.formatParameters(fn.Parameters))
		} else {
			// Method inside a class
			prefixParts := []string{}
			if visPrefix != "" {
				prefixParts = append(prefixParts, strings.TrimSpace(visPrefix))
			}
			if isStatic {
				prefixParts = append(prefixParts, "static")
			}
			if isAsync {
				prefixParts = append(prefixParts, "async")
			}
			if isGenerator {
				prefixParts = append(prefixParts, "*")
			}
			prefix := ""
			if len(prefixParts) > 0 {
				prefix = strings.Join(prefixParts, " ") + " "
			}
			fmt.Fprintf(w, "%s%s%s(%s)", indentStr, prefix, functionName, f.formatParameters(fn.Parameters))
		}
	}
	
	// Add return type if available (from JSDoc)
	if fn.Returns != nil && fn.Returns.Name != "" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	fmt.Fprintln(w)
	
	// Implementation
	if fn.Implementation != "" {
		fmt.Fprintf(w, "%s    // implementation\n", indentStr)
	}
	
	return nil
}

func (f *JavaScriptFormatter) formatField(w io.Writer, field *ir.DistilledField, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Format visibility prefix
	visPrefix := f.getVisibilityPrefix(field.Visibility)
	
	// Check for modifiers
	isStatic := false
	isConst := false
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierStatic {
			isStatic = true
		} else if mod == ir.ModifierFinal {
			isConst = true
		}
	}
	
	if indent == 0 {
		// Top-level variable declaration
		varType := "let"
		if isConst {
			varType = "const"
		}
		
		fmt.Fprintf(w, "%s%s %s", visPrefix, varType, field.Name)
		
		// Add type annotation if available (from JSDoc)
		if field.Type != nil && field.Type.Name != "" {
			fmt.Fprintf(w, ": %s", field.Type.Name)
		}
		
		// Add default value
		if field.DefaultValue != "" {
			fmt.Fprintf(w, " = %s", field.DefaultValue)
		}
		
		fmt.Fprintln(w)
	} else {
		// Class field
		prefixParts := []string{}
		if visPrefix != "" {
			prefixParts = append(prefixParts, strings.TrimSpace(visPrefix))
		}
		if isStatic {
			prefixParts = append(prefixParts, "static")
		}
		prefix := ""
		if len(prefixParts) > 0 {
			prefix = strings.Join(prefixParts, " ") + " "
		}
		
		fmt.Fprintf(w, "%s%s%s", indentStr, prefix, field.Name)
		
		// Add type annotation if available (from JSDoc)
		if field.Type != nil && field.Type.Name != "" {
			fmt.Fprintf(w, ": %s", field.Type.Name)
		}
		
		// Add default value
		if field.DefaultValue != "" {
			fmt.Fprintf(w, " = %s", field.DefaultValue)
		}
		
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *JavaScriptFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Handle special comment formats
	switch comment.Format {
	case "export":
		// Export-related comments (e.g., module.exports)
		fmt.Fprintf(w, "%s// %s\n", indentStr, comment.Text)
	case "block":
		// Block comments (e.g., static initialization blocks)
		fmt.Fprintf(w, "%s// %s\n", indentStr, comment.Text)
	default:
		// Regular comments
		if strings.Contains(comment.Text, "\n") {
			// Multi-line comment
			lines := strings.Split(comment.Text, "\n")
			for _, line := range lines {
				fmt.Fprintf(w, "%s// %s\n", indentStr, strings.TrimSpace(line))
			}
		} else {
			fmt.Fprintf(w, "%s// %s\n", indentStr, comment.Text)
		}
	}
	
	return nil
}

// Helper methods

func (f *JavaScriptFormatter) getVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPrivate:
		return "- " // Private methods/fields (# prefix or _convention)
	case ir.VisibilityProtected:
		return "# " // Protected (convention-based)
	case ir.VisibilityPublic:
		return "+ " // Explicitly public
	default:
		return "" // Default visibility (implicitly public)
	}
}

func (f *JavaScriptFormatter) formatParameters(params []ir.Parameter) string {
	if len(params) == 0 {
		return ""
	}
	
	paramStrs := make([]string, 0, len(params))
	for _, param := range params {
		if param.Name == "" {
			continue
		}
		
		paramStr := param.Name
		
		// Add type annotation if available (from JSDoc)
		if param.Type.Name != "" {
			paramStr += ": " + param.Type.Name
		}
		
		// Add default value
		if param.DefaultValue != "" {
			paramStr += " = " + param.DefaultValue
		}
		
		paramStrs = append(paramStrs, paramStr)
	}
	
	return strings.Join(paramStrs, ", ")
}