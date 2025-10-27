package parser

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Embed the tree-sitter Python WASM module
// In a real implementation, this would be built by the setup script
//go:embed tree_sitter_python.wasm
var treeSitterPythonWASM []byte

// WASMPythonParser wraps a WASM-based tree-sitter parser
type WASMPythonParser struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	module   api.Module

	// Function references
	parseFunc api.Function
	initFunc  api.Function
}

// NewWASMPythonParser creates a new WASM-based Python parser
func NewWASMPythonParser() (*WASMPythonParser, error) {
	ctx := context.Background()

	// Create runtime with WASI support
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI (needed by emscripten-compiled modules)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// For this PoC, we'll simulate loading a WASM module
	// In reality, this would load the actual tree-sitter-python.wasm
	if len(treeSitterPythonWASM) == 0 {
		// Simulate a minimal WASM module for the PoC
		return &WASMPythonParser{
			runtime: r,
		}, nil
	}

	// Compile the module
	compiled, err := r.CompileModule(ctx, treeSitterPythonWASM)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate the module
	module, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	// Get function exports
	parseFunc := module.ExportedFunction("tree_sitter_parse")
	if parseFunc == nil {
		module.Close(ctx)
		r.Close(ctx)
		return nil, fmt.Errorf("parse function not found in WASM module")
	}

	return &WASMPythonParser{
		runtime:   r,
		compiled:  compiled,
		module:    module,
		parseFunc: parseFunc,
	}, nil
}

// Parse parses Python source code using the WASM module
func (p *WASMPythonParser) Parse(source []byte) (*ParseTree, error) {
	// For this PoC, we simulate parsing
	// In a real implementation, this would:
	// 1. Allocate memory in WASM for the source
	// 2. Copy source to WASM memory
	// 3. Call the parse function
	// 4. Read the resulting tree from WASM memory

	// Simulate some processing time
	nodeCount := len(source) / 10 // Rough estimate

	// Check for basic syntax errors (very simplified)
	hasErrors := false
	openParens := 0
	for _, b := range source {
		switch b {
		case '(':
			openParens++
		case ')':
			openParens--
			if openParens < 0 {
				hasErrors = true
			}
		}
	}
	if openParens != 0 {
		hasErrors = true
	}

	return &ParseTree{
		RootType:  "module",
		NodeCount: nodeCount,
		HasErrors: hasErrors,
	}, nil
}

// Close cleans up the WASM runtime
func (p *WASMPythonParser) Close() error {
	ctx := context.Background()

	if p.module != nil {
		p.module.Close(ctx)
	}

	if p.runtime != nil {
		return p.runtime.Close(ctx)
	}

	return nil
}