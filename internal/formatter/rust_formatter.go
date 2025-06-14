package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// RustFormatter implements language-specific formatting for Rust
type RustFormatter struct {
	BaseLanguageFormatter
}

// NewRustFormatter creates a new Rust formatter
func NewRustFormatter() LanguageFormatter {
	return &RustFormatter{
		BaseLanguageFormatter: NewBaseLanguageFormatter("rust"),
	}
}

// FormatNode implements LanguageFormatter
func (f *RustFormatter) FormatNode(w io.Writer, node ir.DistilledNode, indent int) error {
	indentStr := strings.Repeat("    ", indent)

	switch n := node.(type) {
	case *ir.DistilledImport:
		return f.formatImport(w, n, indentStr)
	case *ir.DistilledClass:
		return f.formatStruct(w, n, indent)
	case *ir.DistilledInterface:
		return f.formatTrait(w, n, indent)
	case *ir.DistilledFunction:
		return f.formatFunction(w, n, indentStr)
	case *ir.DistilledField:
		return f.formatField(w, n, indentStr)
	case *ir.DistilledTypeAlias:
		return f.formatTypeAlias(w, n, indentStr)
	case *ir.DistilledComment:
		return f.formatComment(w, n, indentStr)
	default:
		// Skip unknown nodes
		return nil
	}
}

func (f *RustFormatter) formatImport(w io.Writer, imp *ir.DistilledImport, indent string) error {
	// Rust uses "use" for imports
	if len(imp.Symbols) > 0 {
		symbols := make([]string, len(imp.Symbols))
		for i, sym := range imp.Symbols {
			if sym.Alias != "" {
				symbols[i] = fmt.Sprintf("%s as %s", sym.Name, sym.Alias)
			} else {
				symbols[i] = sym.Name
			}
		}
		fmt.Fprintf(w, "%suse %s::{%s};\n", indent, imp.Module, strings.Join(symbols, ", "))
	} else {
		fmt.Fprintf(w, "%suse %s;\n", indent, imp.Module)
	}
	return nil
}

func (f *RustFormatter) formatComment(w io.Writer, comment *ir.DistilledComment, indent string) error {
	lines := strings.Split(comment.Text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
			// Doc comments
			fmt.Fprintf(w, "%s%s\n", indent, line)
		} else if line != "" {
			fmt.Fprintf(w, "%s// %s\n", indent, line)
		} else {
			fmt.Fprintf(w, "%s//\n", indent)
		}
	}
	return nil
}

func (f *RustFormatter) formatStruct(w io.Writer, class *ir.DistilledClass, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Add blank line before struct
	fmt.Fprintln(w)
	
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(class.Visibility)
	
	// Check if this is actually a module (classes without ModifierStruct are modules in Rust)
	isModule := true
	for _, mod := range class.Modifiers {
		if mod == ir.ModifierStruct {
			isModule = false
			break
		}
	}
	
	// Format declaration
	if isModule {
		fmt.Fprintf(w, "%s%smod %s", indentStr, visPrefix, class.Name)
	} else {
		fmt.Fprintf(w, "%s%sstruct %s", indentStr, visPrefix, class.Name)
	}
	
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
	
	// Handle modules vs structs differently
	if isModule {
		// Modules have a body with various items
		if len(class.Children) > 0 {
			fmt.Fprintln(w, " {")
			// Format all children in the module
			for _, child := range class.Children {
				switch n := child.(type) {
				case *ir.DistilledField:
					f.formatField(w, n, indentStr+"    ")
				case *ir.DistilledFunction:
					f.formatFunction(w, n, indentStr+"    ")
				case *ir.DistilledComment:
					f.formatComment(w, n, indentStr+"    ")
				}
			}
			fmt.Fprintf(w, "%s}\n", indentStr)
		} else {
			fmt.Fprintln(w, ";")
		}
	} else {
		// Regular struct handling
		// Check if there are fields
		hasFields := false
		for _, child := range class.Children {
			if _, ok := child.(*ir.DistilledField); ok {
				hasFields = true
				break
			}
		}
		
		if hasFields {
			fmt.Fprintln(w, " {")
			// Format fields
			for _, child := range class.Children {
				if field, ok := child.(*ir.DistilledField); ok {
					f.formatStructField(w, field, indent+1)
				}
			}
			fmt.Fprintf(w, "%s}\n", indentStr)
		} else {
			// Unit struct
			fmt.Fprintln(w, ";")
		}
		
		// Format impl blocks (methods)
		var methods []*ir.DistilledFunction
		for _, child := range class.Children {
			if fn, ok := child.(*ir.DistilledFunction); ok {
				methods = append(methods, fn)
			}
		}
		
		if len(methods) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%simpl %s", indentStr, class.Name)
			
			// Add generic type parameters for impl
			if len(class.TypeParams) > 0 {
				typeParams := make([]string, len(class.TypeParams))
				for i, param := range class.TypeParams {
					typeParams[i] = param.Name
				}
				fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
			}
			
			fmt.Fprintln(w, " {")
			for _, method := range methods {
				f.formatImplMethod(w, method, indent+1)
			}
			fmt.Fprintf(w, "%s}\n", indentStr)
		}
	}
	
	return nil
}

func (f *RustFormatter) formatStructField(w io.Writer, field *ir.DistilledField, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Get visibility
	vis := ""
	if field.Visibility == ir.VisibilityPublic {
		vis = "pub "
	}
	
	typeName := ""
	if field.Type != nil {
		typeName = field.Type.Name
	}
	fmt.Fprintf(w, "%s%s%s: %s,\n", indentStr, vis, field.Name, typeName)
	return nil
}

func (f *RustFormatter) formatTrait(w io.Writer, trait *ir.DistilledInterface, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Add blank line before trait
	fmt.Fprintln(w)
	
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(trait.Visibility)
	
	// Format trait declaration
	fmt.Fprintf(w, "%s%strait %s", indentStr, visPrefix, trait.Name)
	
	// Add generic type parameters
	if len(trait.TypeParams) > 0 {
		typeParams := make([]string, len(trait.TypeParams))
		for i, param := range trait.TypeParams {
			typeParams[i] = param.Name
			if len(param.Constraints) > 0 {
				typeParams[i] += ": " + param.Constraints[0].Name
			}
		}
		fmt.Fprintf(w, "<%s>", strings.Join(typeParams, ", "))
	}
	
	// Add supertrait bounds
	if len(trait.Extends) > 0 {
		extends := make([]string, len(trait.Extends))
		for i, ext := range trait.Extends {
			extends[i] = ext.Name
		}
		fmt.Fprintf(w, ": %s", strings.Join(extends, " + "))
	}
	
	fmt.Fprintln(w, " {")
	
	// Format trait members
	for _, child := range trait.Children {
		if fn, ok := child.(*ir.DistilledFunction); ok {
			f.formatTraitMethod(w, fn, indent+1)
		}
	}
	
	fmt.Fprintf(w, "%s}\n", indentStr)
	
	return nil
}

func (f *RustFormatter) formatFunction(w io.Writer, fn *ir.DistilledFunction, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(fn.Visibility)
	
	// Check for async
	isAsync := false
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierAsync {
			isAsync = true
			break
		}
	}
	
	// Format function signature
	if isAsync {
		fmt.Fprintf(w, "\n%s%sasync fn %s", indent, visPrefix, fn.Name)
	} else {
		fmt.Fprintf(w, "\n%s%sfn %s", indent, visPrefix, fn.Name)
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
	if fn.Returns != nil && fn.Returns.Name != "" && fn.Returns.Name != "()" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Implementation
	if fn.Implementation != "" {
		fmt.Fprintln(w, " {")
		impl := strings.TrimSpace(fn.Implementation)
		for _, line := range strings.Split(impl, "\n") {
			fmt.Fprintf(w, "%s    %s\n", indent, line)
		}
		fmt.Fprintf(w, "%s}\n", indent)
	} else {
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *RustFormatter) formatImplMethod(w io.Writer, fn *ir.DistilledFunction, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Check for visibility
	vis := ""
	if fn.Visibility == ir.VisibilityPublic {
		vis = "pub "
	}
	
	// Check for async
	isAsync := false
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierAsync {
			isAsync = true
			break
		}
	}
	
	// Format method signature
	if isAsync {
		fmt.Fprintf(w, "%s%sasync fn %s", indentStr, vis, fn.Name)
	} else {
		fmt.Fprintf(w, "%s%sfn %s", indentStr, vis, fn.Name)
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
	if fn.Returns != nil && fn.Returns.Name != "" && fn.Returns.Name != "()" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Implementation
	if fn.Implementation != "" {
		fmt.Fprintln(w, " {")
		impl := strings.TrimSpace(fn.Implementation)
		for _, line := range strings.Split(impl, "\n") {
			fmt.Fprintf(w, "%s    %s\n", indentStr, line)
		}
		fmt.Fprintf(w, "%s}\n", indentStr)
	} else {
		fmt.Fprintln(w)
	}
	
	return nil
}

func (f *RustFormatter) formatTraitMethod(w io.Writer, fn *ir.DistilledFunction, indent int) error {
	indentStr := strings.Repeat("    ", indent)
	
	// Check for async
	isAsync := false
	for _, mod := range fn.Modifiers {
		if mod == ir.ModifierAsync {
			isAsync = true
			break
		}
	}
	
	// Format method signature
	if isAsync {
		fmt.Fprintf(w, "%sasync fn %s", indentStr, fn.Name)
	} else {
		fmt.Fprintf(w, "%sfn %s", indentStr, fn.Name)
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
	if fn.Returns != nil && fn.Returns.Name != "" && fn.Returns.Name != "()" {
		fmt.Fprintf(w, " -> %s", fn.Returns.Name)
	}
	
	// Default implementation or just signature
	if fn.Implementation != "" {
		fmt.Fprintln(w, " {")
		impl := strings.TrimSpace(fn.Implementation)
		for _, line := range strings.Split(impl, "\n") {
			fmt.Fprintf(w, "%s    %s\n", indentStr, line)
		}
		fmt.Fprintf(w, "%s}\n", indentStr)
	} else {
		fmt.Fprintln(w, ";")
	}
	
	return nil
}

func (f *RustFormatter) formatField(w io.Writer, field *ir.DistilledField, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(field.Visibility)
	
	// Check if it's a constant
	isConst := false
	isStatic := false
	for _, mod := range field.Modifiers {
		if mod == ir.ModifierFinal {
			isConst = true
		}
		if mod == ir.ModifierStatic {
			isStatic = true
		}
	}
	
	// Format field/constant
	typeName := ""
	if field.Type != nil {
		typeName = field.Type.Name
	}
	if isConst {
		fmt.Fprintf(w, "\n%s%sconst %s: %s", indent, visPrefix, strings.ToUpper(field.Name), typeName)
	} else if isStatic {
		fmt.Fprintf(w, "\n%s%sstatic %s: %s", indent, visPrefix, field.Name, typeName)
	} else {
		// Regular field (shouldn't appear at top level in Rust)
		fmt.Fprintf(w, "%s%slet %s", indent, visPrefix, field.Name)
		if field.Type != nil && field.Type.Name != "" {
			fmt.Fprintf(w, ": %s", field.Type.Name)
		}
	}
	
	// Add value if specified
	if field.DefaultValue != "" {
		fmt.Fprintf(w, " = %s", field.DefaultValue)
	}
	
	fmt.Fprintln(w, ";")
	return nil
}

func (f *RustFormatter) formatTypeAlias(w io.Writer, alias *ir.DistilledTypeAlias, indent string) error {
	// Get visibility prefix
	visPrefix := f.getVisibilityPrefix(alias.Visibility)
	
	fmt.Fprintf(w, "\n%s%stype %s", indent, visPrefix, alias.Name)
	
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
	
	fmt.Fprintf(w, " = %s;\n", alias.Type.Name)
	
	return nil
}

func (f *RustFormatter) formatParameters(w io.Writer, params []ir.Parameter) {
	for i, param := range params {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		
		// Check for self parameter
		if param.Name == "self" {
			// Check for &self, &mut self, etc.
			if strings.Contains(param.Type.Name, "&") {
				fmt.Fprintf(w, "%s", param.Type.Name)
			} else {
				fmt.Fprintf(w, "self")
			}
		} else {
			// Regular parameter
			fmt.Fprintf(w, "%s: %s", param.Name, param.Type.Name)
		}
	}
}

func (f *RustFormatter) getVisibilityPrefix(visibility ir.Visibility) string {
	switch visibility {
	case ir.VisibilityPrivate:
		return "-"
	case ir.VisibilityProtected:
		return "*"
	case ir.VisibilityPublic:
		return "" // No prefix for public
	case ir.VisibilityInternal:
		return "~"
	default:
		return ""
	}
}