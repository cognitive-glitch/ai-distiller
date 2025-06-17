package language

import (
	"github.com/janreges/ai-distiller/internal/processor"
)

// MustRegisterAll registers all processors and panics on error
func MustRegisterAll() {
	if err := RegisterAll(); err != nil {
		panic(err)
	}
}

// GetProcessor returns a language processor by name
func GetProcessor(language string) (processor.LanguageProcessor, bool) {
	return processor.Get(language)
}

func init() {
	// Register all language processors at startup
	MustRegisterAll()
}
