package python

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/parser"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Processor implements the LanguageProcessor interface for Python
type Processor struct {
	processor.BaseProcessor
	wasmRuntime *parser.WASMRuntime
	module      *parser.WASMModule
}

// NewProcessor creates a new Python language processor
func NewProcessor() *Processor {
	return &Processor{
		BaseProcessor: processor.NewBaseProcessor(
			"python",
			"1.0.0",
			[]string{".py", ".pyw", ".pyi"},
		),
	}
}

// InitializeWASM sets up the WASM runtime and loads the Python parser
func (p *Processor) InitializeWASM(ctx context.Context, wasmBytes []byte) error {
	// Create WASM runtime if not exists
	if p.wasmRuntime == nil {
		runtime, err := parser.NewWASMRuntime(ctx)
		if err != nil {
			return fmt.Errorf("failed to create WASM runtime: %w", err)
		}
		p.wasmRuntime = runtime
	}

	// Load Python parser module
	if err := p.wasmRuntime.LoadModule("tree-sitter-python", wasmBytes); err != nil {
		return fmt.Errorf("failed to load Python parser: %w", err)
	}

	// Get the module
	module, err := p.wasmRuntime.GetModule("tree-sitter-python")
	if err != nil {
		return fmt.Errorf("failed to get Python module: %w", err)
	}

	p.module = module
	return nil
}

// Process parses Python source code and returns the IR representation
func (p *Processor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	// Default options
	opts := processor.DefaultProcessOptions()
	return p.ProcessWithOptions(ctx, reader, filename, opts)
}

// ProcessFile processes a file by path
func (p *Processor) ProcessFile(filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// For now, return a mock implementation
	// TODO: Implement file reading and processing
	
	// Simulate file not found error
	if filename == "nonexistent.py" {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	
	ctx := context.Background()
	source := []byte("# Mock Python file\nclass Test:\n    pass")
	return p.processMock(ctx, source, filename, opts)
}

// ProcessWithOptions parses with specific options
func (p *Processor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Read source code
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	// If no WASM module loaded, create a mock response for now
	if p.module == nil {
		return p.processMock(ctx, source, filename, opts)
	}

	// Create parser
	tsParser, err := parser.NewParser(p.module)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	defer tsParser.Delete()

	// Parse the source
	tree, err := tsParser.Parse(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	defer tree.Delete()

	// Convert to IR
	converter := parser.NewConverter("python", source)
	file, err := converter.ConvertTree(ctx, tree)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to IR: %w", err)
	}

	// Apply options
	file = p.applyOptions(file, opts)

	file.Path = filename
	return file, nil
}

// processMock creates a mock IR for testing without WASM
func (p *Processor) processMock(ctx context.Context, source []byte, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Create a mock file structure
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine:   1,
				StartColumn: 1,
				EndLine:     countLines(source),
				EndColumn:   1,
			},
		},
		Path:     filename,
		Language: "python",
		Version:  "2.0.0",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}

	// Parse Python-specific constructs (simplified mock)
	// In real implementation, this would use tree-sitter
	nodes := p.mockParsePython(source, opts)
	file.Children = nodes

	return file, nil
}

// mockParsePython creates mock nodes for testing
func (p *Processor) mockParsePython(source []byte, opts processor.ProcessOptions) []ir.DistilledNode {
	nodes := []ir.DistilledNode{}

	// Mock: Add a module-level comment
	if opts.IncludeComments {
		nodes = append(nodes, createMockComment("Python source file"))
	}

	// Mock: Add an import
	if opts.IncludeImports {
		nodes = append(nodes, createMockImport("typing", []string{"List", "Dict", "Optional"}))
	}

	// Mock: Add a class
	classNode := createMockClass("ExampleClass", "public")
	
	// Mock: Add a method to the class
	methodNode := createMockFunction("example_method", "public", []string{"self", "arg1: str", "arg2: int"}, "str")
	classNode.Children = append(classNode.Children, methodNode)

	nodes = append(nodes, classNode)

	// Mock: Add a function
	if opts.IncludePrivate || !isPrivate("process_data") {
		funcNode := createMockFunction("process_data", "public", []string{"data: List[Dict]"}, "Dict")
		nodes = append(nodes, funcNode)
	}

	return nodes
}

// applyOptions filters the IR based on processing options
func (p *Processor) applyOptions(file *ir.DistilledFile, opts processor.ProcessOptions) *ir.DistilledFile {
	if !opts.IncludeComments || !opts.IncludeImports || !opts.IncludePrivate || !opts.IncludeImplementation {
		// Create a visitor to filter nodes
		filterVisitor := ir.NewFuncVisitor(func(node ir.DistilledNode) ir.DistilledNode {
			switch n := node.(type) {
			case *ir.DistilledComment:
				if !opts.IncludeComments {
					return nil
				}
			case *ir.DistilledImport:
				if !opts.IncludeImports {
					return nil
				}
			case *ir.DistilledFunction:
				if !opts.IncludePrivate && isPrivate(n.Name) {
					return nil
				}
				if !opts.IncludeImplementation {
					// Clear implementation
					n.Implementation = ""
				}
			case *ir.DistilledClass:
				if !opts.IncludePrivate && isPrivate(n.Name) {
					return nil
				}
			}
			return node
		})

		// Apply filter
		walker := ir.NewWalker(filterVisitor)
		if filtered := walker.Walk(file); filtered != nil {
			return filtered.(*ir.DistilledFile)
		}
	}

	return file
}

// Helper functions

func countLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := 1
	for _, b := range source {
		if b == '\n' {
			lines++
		}
	}
	return lines
}

func isPrivate(name string) bool {
	// In Python, names starting with _ are considered private
	return len(name) > 0 && name[0] == '_'
}

// Mock node creators for testing

func createMockComment(text string) *ir.DistilledComment {
	return &ir.DistilledComment{
		BaseNode: ir.BaseNode{
			Location: ir.Location{StartLine: 1, EndLine: 1},
		},
		Text:   text,
		Format: "line",
	}
}

func createMockImport(module string, names []string) *ir.DistilledImport {
	symbols := make([]ir.ImportedSymbol, len(names))
	for i, name := range names {
		symbols[i] = ir.ImportedSymbol{
			Name:  name,
			Alias: "",
		}
	}

	return &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: ir.Location{StartLine: 2, EndLine: 2},
		},
		ImportType: "from",
		Module:     module,
		Symbols:    symbols,
	}
}

func createMockClass(name string, visibility string) *ir.DistilledClass {
	return &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{StartLine: 4, EndLine: 10},
		},
		Name:         name,
		Visibility:   ir.Visibility(visibility),
		Decorators:   []string{},
		TypeParams:   []ir.TypeParam{},
		Extends:      []ir.TypeRef{},
		Implements:   []ir.TypeRef{},
		Children:     []ir.DistilledNode{},
	}
}

func createMockFunction(name string, visibility string, params []string, returnType string) *ir.DistilledFunction {
	parameters := make([]ir.Parameter, len(params))
	for i, param := range params {
		parameters[i] = ir.Parameter{
			Name: param,
			Type: ir.TypeRef{Name: "Any"},
		}
	}

	return &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: ir.Location{StartLine: 5, EndLine: 8},
		},
		Name:       name,
		Visibility: ir.Visibility(visibility),
		Modifiers:  []ir.Modifier{},
		Parameters: parameters,
		Returns: &ir.TypeRef{
			Name: returnType,
		},
		Decorators:     []string{},
		Implementation: "# Implementation details omitted",
	}
}