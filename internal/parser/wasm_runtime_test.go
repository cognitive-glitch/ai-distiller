package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWASMRuntime(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWASMRuntime(ctx)
	
	require.NoError(t, err)
	assert.NotNil(t, runtime)
	
	// Clean up
	err = runtime.Close()
	assert.NoError(t, err)
}

func TestWASMRuntimeLoadModule(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWASMRuntime(ctx)
	require.NoError(t, err)
	defer runtime.Close()
	
	// Test with empty module (will fail but tests the flow)
	err = runtime.LoadModule("test", []byte{})
	assert.Error(t, err) // Empty module should fail
	
	// Test duplicate module
	// First, we'd need a valid WASM module
	// For now, we just test that the registry works
	modules := runtime.modules
	assert.Empty(t, modules)
}

func TestWASMRuntimeGetModule(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWASMRuntime(ctx)
	require.NoError(t, err)
	defer runtime.Close()
	
	// Test getting non-existent module
	module, err := runtime.GetModule("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, module)
	assert.Contains(t, err.Error(), "not found")
}

func TestWASMModuleLoadFunctions(t *testing.T) {
	// This test would require a real WASM module
	// For now, we just test that the structure is correct
	module := &WASMModule{
		Name:     "test",
		Language: "test",
	}
	
	assert.Equal(t, "test", module.Name)
	assert.Equal(t, "test", module.Language)
	assert.Nil(t, module.TreeSitterLanguage)
}