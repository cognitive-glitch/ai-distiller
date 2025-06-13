package golang

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// ASTParser uses Go's native AST parser for processing Go source code
type ASTParser struct {
	source   []byte
	filename string
	fset     *token.FileSet
}

// NewASTParser creates a new AST parser
func NewASTParser() *ASTParser {
	return &ASTParser{
		fset: token.NewFileSet(),
	}
}

// ProcessSource processes Go source code using go/parser
func (p *ASTParser) ProcessSource(ctx context.Context, source []byte, filename string) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename

	// Parse the source code
	file, err := parser.ParseFile(p.fset, filename, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Create distilled file
	distilledFile := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(file.Pos(), file.End()),
		},
		Path:     filename,
		Language: "go",
		Version:  "1.0",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}

	// Process package declaration
	if file.Name != nil {
		pkg := &ir.DistilledPackage{
			BaseNode: ir.BaseNode{
				Location: p.getLocation(file.Name.Pos(), file.Name.End()),
			},
			Name: file.Name.Name,
		}
		distilledFile.Children = append(distilledFile.Children, pkg)
	}

	// Process imports
	for _, spec := range file.Imports {
		imp := p.processImport(spec)
		if imp != nil {
			distilledFile.Children = append(distilledFile.Children, imp)
		}
	}

	// Process declarations
	for _, decl := range file.Decls {
		nodes := p.processDecl(decl)
		distilledFile.Children = append(distilledFile.Children, nodes...)
	}

	return distilledFile, nil
}

// processImport processes an import specification
func (p *ASTParser) processImport(spec *ast.ImportSpec) *ir.DistilledImport {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(spec.Pos(), spec.End()),
		},
		ImportType: "import",
		Symbols:    []ir.ImportedSymbol{},
	}

	// Get the import path without quotes
	path := strings.Trim(spec.Path.Value, `"`)
	imp.Module = path

	// Handle aliased imports
	if spec.Name != nil {
		alias := spec.Name.Name
		if alias == "." {
			// Dot import
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  ".",
				Alias: "",
			})
		} else if alias == "_" {
			// Blank import
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  "_",
				Alias: "",
			})
		} else {
			// Named import
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  path,
				Alias: alias,
			})
		}
	} else {
		// Standard import: use the last part of the path as the name
		parts := strings.Split(path, "/")
		pkgName := parts[len(parts)-1]
		// Handle versioned paths like "gopkg.in/yaml.v2"
		if strings.HasPrefix(pkgName, "v") && len(pkgName) <= 3 && len(parts) > 1 {
			pkgName = parts[len(parts)-2]
		}
		imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
			Name: pkgName,
		})
	}

	return imp
}

// processDecl processes a declaration
func (p *ASTParser) processDecl(decl ast.Decl) []ir.DistilledNode {
	var nodes []ir.DistilledNode

	switch d := decl.(type) {
	case *ast.FuncDecl:
		fn := p.processFunction(d)
		if fn != nil {
			nodes = append(nodes, fn)
		}
	case *ast.GenDecl:
		nodes = append(nodes, p.processGenDecl(d)...)
	}

	return nodes
}

// processGenDecl processes a general declaration (type, const, var)
func (p *ASTParser) processGenDecl(decl *ast.GenDecl) []ir.DistilledNode {
	var nodes []ir.DistilledNode

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			node := p.processTypeSpec(s)
			if node != nil {
				nodes = append(nodes, node)
			}
		case *ast.ValueSpec:
			// Process const/var declarations
			for i, name := range s.Names {
				if name.Name == "_" {
					continue
				}

				field := &ir.DistilledField{
					BaseNode: ir.BaseNode{
						Location: p.getLocation(name.Pos(), name.End()),
					},
					Name:       name.Name,
					Visibility: p.getVisibility(name.Name),
					Modifiers:  []ir.Modifier{},
				}

				// Determine if const or var
				if decl.Tok == token.CONST {
					field.Modifiers = append(field.Modifiers, ir.ModifierFinal)
				}

				// Get type if specified
				if s.Type != nil {
					field.Type = &ir.TypeRef{Name: p.typeToString(s.Type)}
				}

				// Get value if specified
				if s.Values != nil && i < len(s.Values) {
					field.DefaultValue = p.exprToString(s.Values[i])
				}

				nodes = append(nodes, field)
			}
		}
	}

	return nodes
}

// processTypeSpec processes a type specification
func (p *ASTParser) processTypeSpec(spec *ast.TypeSpec) ir.DistilledNode {
	switch t := spec.Type.(type) {
	case *ast.StructType:
		return p.processStruct(spec.Name.Name, t)
	case *ast.InterfaceType:
		return p.processInterface(spec.Name.Name, t)
	default:
		// Type alias
		alias := &ir.DistilledTypeAlias{
			BaseNode: ir.BaseNode{
				Location: p.getLocation(spec.Pos(), spec.End()),
			},
			Name:       spec.Name.Name,
			Visibility: p.getVisibility(spec.Name.Name),
			Type:       ir.TypeRef{Name: p.typeToString(spec.Type)},
		}
		return alias
	}
}

// processStruct processes a struct type
func (p *ASTParser) processStruct(name string, structType *ast.StructType) *ir.DistilledClass {
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(structType.Pos(), structType.End()),
		},
		Name:       name,
		Visibility: p.getVisibility(name),
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}

	// Process fields
	for _, field := range structType.Fields.List {
		// Get field type
		fieldType := p.typeToString(field.Type)

		if len(field.Names) == 0 {
			// Embedded field
			embedField := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: p.getLocation(field.Pos(), field.End()),
				},
				Name:       fieldType, // Use type name as field name
				Type:       &ir.TypeRef{Name: fieldType},
				Visibility: ir.VisibilityPublic,
				Modifiers:  []ir.Modifier{ir.ModifierStatic}, // Mark as embedded
			}
			class.Children = append(class.Children, embedField)
		} else {
			// Named fields
			for _, name := range field.Names {
				distilledField := &ir.DistilledField{
					BaseNode: ir.BaseNode{
						Location: p.getLocation(name.Pos(), name.End()),
					},
					Name:       name.Name,
					Type:       &ir.TypeRef{Name: fieldType},
					Visibility: p.getVisibility(name.Name),
				}

				// TODO: Process tags when annotation support is added

				class.Children = append(class.Children, distilledField)
			}
		}
	}

	return class
}

// collectUnionTypes recursively collects types from a type union expression
func (p *ASTParser) collectUnionTypes(expr ast.Expr) []string {
	if binExpr, ok := expr.(*ast.BinaryExpr); ok && binExpr.Op == token.OR {
		var types []string
		types = append(types, p.collectUnionTypes(binExpr.X)...)
		types = append(types, p.collectUnionTypes(binExpr.Y)...)
		return types
	}
	// Base case: a single type identifier
	return []string{p.typeToString(expr)}
}

// processInterface processes an interface type
func (p *ASTParser) processInterface(name string, interfaceType *ast.InterfaceType) *ir.DistilledInterface {
	intf := &ir.DistilledInterface{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(interfaceType.Pos(), interfaceType.End()),
		},
		Name:       name,
		Visibility: p.getVisibility(name),
		Children:   []ir.DistilledNode{},
	}

	// Process methods
	for _, method := range interfaceType.Methods.List {
		if len(method.Names) == 0 {
			// Embedded interface or type union
			types := p.collectUnionTypes(method.Type)
			for _, t := range types {
				intf.Extends = append(intf.Extends, ir.TypeRef{Name: t})
			}
		} else {
			// Method declaration
			for _, name := range method.Names {
				if funcType, ok := method.Type.(*ast.FuncType); ok {
					fn := p.processFuncType(name.Name, funcType)
					fn.Visibility = ir.VisibilityPublic
					intf.Children = append(intf.Children, fn)
				}
			}
		}
	}

	return intf
}

// processFunction processes a function declaration
func (p *ASTParser) processFunction(fn *ast.FuncDecl) *ir.DistilledFunction {
	distilledFn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(fn.Pos(), fn.End()),
		},
		Name:       fn.Name.Name,
		Visibility: p.getVisibility(fn.Name.Name),
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
	}

	// Check if it's a method (has receiver)
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		recvType := p.typeToString(recv.Type)
		
		// Add receiver as special parameter
		recvParam := ir.Parameter{
			Name: "receiver",
			Type: ir.TypeRef{Name: recvType},
		}
		if len(recv.Names) > 0 {
			recvParam.Name = recv.Names[0].Name
		}
		distilledFn.Parameters = append(distilledFn.Parameters, recvParam)
		
		// Mark as method
		distilledFn.Modifiers = append(distilledFn.Modifiers, ir.ModifierAbstract) // Using Abstract as "method" marker
	}

	// Process parameters
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			paramType := p.typeToString(param.Type)
			if len(param.Names) == 0 {
				// Unnamed parameter
				distilledFn.Parameters = append(distilledFn.Parameters, ir.Parameter{
					Type: ir.TypeRef{Name: paramType},
				})
			} else {
				// Named parameters
				for _, name := range param.Names {
					distilledFn.Parameters = append(distilledFn.Parameters, ir.Parameter{
						Name: name.Name,
						Type: ir.TypeRef{Name: paramType},
					})
				}
			}
		}
	}

	// Process return types
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		if len(fn.Type.Results.List) == 1 && len(fn.Type.Results.List[0].Names) == 0 {
			// Single unnamed return
			distilledFn.Returns = &ir.TypeRef{Name: p.typeToString(fn.Type.Results.List[0].Type)}
		} else {
			// Multiple returns or named returns
			var returns []string
			for _, result := range fn.Type.Results.List {
				resultType := p.typeToString(result.Type)
				if len(result.Names) == 0 {
					returns = append(returns, resultType)
				} else {
					for _, name := range result.Names {
						returns = append(returns, name.Name+" "+resultType)
					}
				}
			}
			distilledFn.Returns = &ir.TypeRef{Name: "(" + strings.Join(returns, ", ") + ")"}
		}
	}

	// Extract function body if exists
	if fn.Body != nil {
		// Count interesting constructs
		var goroutines, channels, defers int
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.GoStmt:
				goroutines++
			case *ast.SendStmt:
				channels++
			case *ast.DeferStmt:
				defers++
			}
			return true
		})

		// Add modifiers for interesting constructs
		if goroutines > 0 {
			distilledFn.Modifiers = append(distilledFn.Modifiers, ir.ModifierAsync) // Using Async for goroutines
		}
	}

	return distilledFn
}

// processFuncType processes a function type (for interface methods)
func (p *ASTParser) processFuncType(name string, funcType *ast.FuncType) *ir.DistilledFunction {
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: p.getLocation(funcType.Pos(), funcType.End()),
		},
		Name:       name,
		Parameters: []ir.Parameter{},
	}

	// Process parameters
	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			paramType := p.typeToString(param.Type)
			if len(param.Names) == 0 {
				fn.Parameters = append(fn.Parameters, ir.Parameter{
					Type: ir.TypeRef{Name: paramType},
				})
			} else {
				for _, name := range param.Names {
					fn.Parameters = append(fn.Parameters, ir.Parameter{
						Name: name.Name,
						Type: ir.TypeRef{Name: paramType},
					})
				}
			}
		}
	}

	// Process return types
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		if len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
			fn.Returns = &ir.TypeRef{Name: p.typeToString(funcType.Results.List[0].Type)}
		} else {
			var returns []string
			for _, result := range funcType.Results.List {
				returns = append(returns, p.typeToString(result.Type))
			}
			fn.Returns = &ir.TypeRef{Name: "(" + strings.Join(returns, ", ") + ")"}
		}
	}

	return fn
}

// Helper methods

// getLocation converts token positions to IR location
func (p *ASTParser) getLocation(start, end token.Pos) ir.Location {
	startPos := p.fset.Position(start)
	endPos := p.fset.Position(end)
	return ir.Location{
		StartLine:   startPos.Line,
		StartColumn: startPos.Column,
		EndLine:     endPos.Line,
		EndColumn:   endPos.Column,
	}
}

// getVisibility determines visibility based on name
func (p *ASTParser) getVisibility(name string) ir.Visibility {
	if name == "" || name[0] >= 'A' && name[0] <= 'Z' {
		return ir.VisibilityPublic
	}
	return ir.VisibilityPrivate
}

// typeToString converts an AST expression to a string representation
func (p *ASTParser) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + p.typeToString(t.Elt)
		}
		return "[" + p.exprToString(t.Len) + "]" + p.typeToString(t.Elt)
	case *ast.SliceExpr:
		return "[]" + p.typeToString(t.X)
	case *ast.MapType:
		return "map[" + p.typeToString(t.Key) + "]" + p.typeToString(t.Value)
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + p.typeToString(t.Value)
		case ast.RECV:
			return "<-chan " + p.typeToString(t.Value)
		default:
			return "chan " + p.typeToString(t.Value)
		}
	case *ast.FuncType:
		return "func" // Simplified
	case *ast.InterfaceType:
		if len(t.Methods.List) == 0 {
			return "interface{}"
		}
		return "interface{...}"
	case *ast.StructType:
		if len(t.Fields.List) == 0 {
			return "struct{}"
		}
		return "struct{...}"
	case *ast.SelectorExpr:
		return p.typeToString(t.X) + "." + t.Sel.Name
	case *ast.Ellipsis:
		return "..." + p.typeToString(t.Elt)
	default:
		// Fallback: get the source text
		start := p.fset.Position(expr.Pos()).Offset
		end := p.fset.Position(expr.End()).Offset
		if start >= 0 && end <= len(p.source) && start < end {
			return string(p.source[start:end])
		}
		return "unknown"
	}
}

// exprToString converts an expression to string (for default values)
func (p *ASTParser) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	case *ast.CompositeLit:
		return p.typeToString(e.Type) + "{...}"
	default:
		// Get source text
		start := p.fset.Position(expr.Pos()).Offset
		end := p.fset.Position(expr.End()).Offset
		if start >= 0 && end <= len(p.source) && start < end {
			return string(p.source[start:end])
		}
		return ""
	}
}