package language

import (
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/processor"
)

// RegisterAll registers all built-in language processors
func RegisterAll() error {
	// Register Python processor
	pythonProc, err := python.NewProcessor()
	if err != nil {
		return err
	}
	if err := processor.Register(pythonProc); err != nil {
		return err
	}

	// TODO: Register other language processors
	// - Go
	// - JavaScript/TypeScript
	// - Java
	// - C#
	// - Rust
	// - etc.

	return nil
}

// MustRegisterAll registers all processors and panics on error
func MustRegisterAll() {
	if err := RegisterAll(); err != nil {
		panic(err)
	}
}