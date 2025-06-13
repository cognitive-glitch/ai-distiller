package language

import (
	"github.com/janreges/ai-distiller/internal/language/javascript"
	"github.com/janreges/ai-distiller/internal/language/php"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/language/typescript"
	"github.com/janreges/ai-distiller/internal/processor"
)

// RegisterAll registers all built-in language processors
func RegisterAll() error {
	// Register Python processor
	pythonProc := python.NewProcessor()
	if err := processor.Register(pythonProc); err != nil {
		return err
	}

	// Register PHP processor
	phpProc := php.NewProcessor()
	if err := processor.Register(phpProc); err != nil {
		return err
	}

	// Register JavaScript processor
	jsProc := javascript.NewProcessor()
	if err := processor.Register(jsProc); err != nil {
		return err
	}

	// Register TypeScript processor
	tsProc := typescript.NewProcessor()
	if err := processor.Register(tsProc); err != nil {
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

func init() {
	// Register all language processors at startup
	MustRegisterAll()
}