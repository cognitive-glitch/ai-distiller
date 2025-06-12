package parser

import (
	"context"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMRuntime manages WASM modules for tree-sitter parsers
type WASMRuntime struct {
	ctx      context.Context
	runtime  wazero.Runtime
	modules  map[string]*WASMModule
	mu       sync.RWMutex
}

// WASMModule represents a loaded tree-sitter WASM module
type WASMModule struct {
	Name     string
	Language string
	Module   api.Module
	Compiled wazero.CompiledModule
	
	// Function references
	TreeSitterLanguage api.Function
	ParserNew          api.Function
	ParserDelete       api.Function
	ParserParse        api.Function
	ParserSetLanguage  api.Function
	TreeDelete         api.Function
	TreeRootNode       api.Function
	NodeString         api.Function
	NodeChildCount     api.Function
	NodeChild          api.Function
	NodeType           api.Function
	NodeStartByte      api.Function
	NodeEndByte        api.Function
	NodeStartPoint     api.Function
	NodeEndPoint       api.Function
	NodeIsNamed        api.Function
	NodeIsError        api.Function
	NodeFieldNameForChild api.Function
	NodeChildByFieldName  api.Function
}

// NewWASMRuntime creates a new WASM runtime for tree-sitter parsers
func NewWASMRuntime(ctx context.Context) (*WASMRuntime, error) {
	r := wazero.NewRuntime(ctx)
	
	// Instantiate WASI (required by emscripten-compiled modules)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	
	// Define host functions that tree-sitter might need
	if err := defineHostFunctions(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to define host functions: %w", err)
	}
	
	return &WASMRuntime{
		ctx:     ctx,
		runtime: r,
		modules: make(map[string]*WASMModule),
	}, nil
}

// LoadModule loads a tree-sitter WASM module
func (w *WASMRuntime) LoadModule(name string, wasmBytes []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Check if already loaded
	if _, exists := w.modules[name]; exists {
		return fmt.Errorf("module %s already loaded", name)
	}
	
	// Compile the module
	compiled, err := w.runtime.CompileModule(w.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile module %s: %w", name, err)
	}
	
	// Instantiate the module
	module, err := w.runtime.InstantiateModule(w.ctx, compiled, wazero.NewModuleConfig().
		WithName(name))
	if err != nil {
		return fmt.Errorf("failed to instantiate module %s: %w", name, err)
	}
	
	// Create module wrapper
	wasmModule := &WASMModule{
		Name:     name,
		Module:   module,
		Compiled: compiled,
	}
	
	// Get function exports
	if err := wasmModule.loadFunctions(); err != nil {
		module.Close(w.ctx)
		return fmt.Errorf("failed to load functions for module %s: %w", name, err)
	}
	
	w.modules[name] = wasmModule
	return nil
}

// GetModule returns a loaded WASM module
func (w *WASMRuntime) GetModule(name string) (*WASMModule, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	module, exists := w.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}
	
	return module, nil
}

// Close shuts down the WASM runtime
func (w *WASMRuntime) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Close all modules
	for name, module := range w.modules {
		if err := module.Module.Close(w.ctx); err != nil {
			// Log error but continue closing others
			fmt.Printf("Error closing module %s: %v\n", name, err)
		}
	}
	
	// Clear modules map
	w.modules = make(map[string]*WASMModule)
	
	// Close runtime
	return w.runtime.Close(w.ctx)
}

// loadFunctions loads function exports from the module
func (m *WASMModule) loadFunctions() error {
	// Load TreeSitterLanguage function
	if fn := m.Module.ExportedFunction("tree_sitter_" + m.Language); fn != nil {
		m.TreeSitterLanguage = fn
	}
	
	// Load parser functions
	if fn := m.Module.ExportedFunction("ts_parser_new"); fn != nil {
		m.ParserNew = fn
	}
	if fn := m.Module.ExportedFunction("ts_parser_delete"); fn != nil {
		m.ParserDelete = fn
	}
	if fn := m.Module.ExportedFunction("ts_parser_parse"); fn != nil {
		m.ParserParse = fn
	}
	if fn := m.Module.ExportedFunction("ts_parser_set_language"); fn != nil {
		m.ParserSetLanguage = fn
	}
	
	// Load tree functions
	if fn := m.Module.ExportedFunction("ts_tree_delete"); fn != nil {
		m.TreeDelete = fn
	}
	if fn := m.Module.ExportedFunction("ts_tree_root_node"); fn != nil {
		m.TreeRootNode = fn
	}
	
	// Load node functions
	if fn := m.Module.ExportedFunction("ts_node_string"); fn != nil {
		m.NodeString = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_child_count"); fn != nil {
		m.NodeChildCount = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_child"); fn != nil {
		m.NodeChild = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_type"); fn != nil {
		m.NodeType = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_start_byte"); fn != nil {
		m.NodeStartByte = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_end_byte"); fn != nil {
		m.NodeEndByte = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_start_point"); fn != nil {
		m.NodeStartPoint = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_end_point"); fn != nil {
		m.NodeEndPoint = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_is_named"); fn != nil {
		m.NodeIsNamed = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_is_error"); fn != nil {
		m.NodeIsError = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_field_name_for_child"); fn != nil {
		m.NodeFieldNameForChild = fn
	}
	if fn := m.Module.ExportedFunction("ts_node_child_by_field_name"); fn != nil {
		m.NodeChildByFieldName = fn
	}
	
	
	// Check that we have at least the language function
	if m.TreeSitterLanguage == nil {
		// Try alternative naming
		if fn := m.Module.ExportedFunction("tree_sitter_language"); fn != nil {
			m.TreeSitterLanguage = fn
		} else {
			return fmt.Errorf("tree-sitter language function not found")
		}
	}
	
	return nil
}

// defineHostFunctions defines functions that tree-sitter WASM modules might call
func defineHostFunctions(ctx context.Context, r wazero.Runtime) error {
	// Define memory management functions that tree-sitter expects
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, size uint32) uint32 {
			// Simple malloc implementation
			// In production, this would need proper memory management
			return size // Return the size as the "address" for simplicity
		}).
		Export("malloc").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, ptr uint32) {
			// Simple free implementation (no-op for now)
		}).
		Export("free").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, ptr uint32, c int32, size uint32) uint32 {
			// Simple memset implementation
			return ptr
		}).
		Export("memset").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, dest, src, size uint32) uint32 {
			// Simple memcpy implementation
			return dest
		}).
		Export("memcpy").
		Instantiate(ctx)
		
	return err
}