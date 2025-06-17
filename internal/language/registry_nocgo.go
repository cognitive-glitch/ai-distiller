//go:build !cgo
// +build !cgo

package language

import (
	"github.com/janreges/ai-distiller/internal/language/golang"
	"github.com/janreges/ai-distiller/internal/processor"
)

// RegisterAll registers only non-CGO processors when CGO is disabled
func RegisterAll() error {
	// Only register Go processor which doesn't require CGO
	goProc := golang.NewProcessor()
	if err := processor.Register(goProc); err != nil {
		return err
	}

	// Register stub processors for other languages
	RegisterTreeSitterProcessors()

	return nil
}