package formatter

import (
	"fmt"
	"strings"
	"sync"
)

// Registry manages available formatters
type Registry struct {
	mu         sync.RWMutex
	formatters map[string]func(Options) Formatter
}

// defaultRegistry is the global formatter registry
var defaultRegistry = &Registry{
	formatters: make(map[string]func(Options) Formatter),
}

// Register adds a formatter to the registry
func (r *Registry) Register(name string, factory func(Options) Formatter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name = strings.ToLower(name)
	if _, exists := r.formatters[name]; exists {
		return fmt.Errorf("formatter %q already registered", name)
	}
	
	r.formatters[name] = factory
	return nil
}

// Get returns a formatter by name
func (r *Registry) Get(name string, options Options) (Formatter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	name = strings.ToLower(name)
	factory, exists := r.formatters[name]
	if !exists {
		return nil, fmt.Errorf("formatter %q not found", name)
	}
	
	return factory(options), nil
}

// List returns all registered formatter names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.formatters))
	for name := range r.formatters {
		names = append(names, name)
	}
	return names
}

// Register adds a formatter to the default registry
func Register(name string, factory func(Options) Formatter) error {
	return defaultRegistry.Register(name, factory)
}

// Get returns a formatter from the default registry
func Get(name string, options Options) (Formatter, error) {
	return defaultRegistry.Get(name, options)
}

// List returns all formatters in the default registry
func List() []string {
	return defaultRegistry.List()
}

// init registers all built-in formatters
func init() {
	// Register built-in formatters
	Register("markdown", func(opts Options) Formatter {
		return NewMarkdownFormatter(opts)
	})
	Register("md", func(opts Options) Formatter {
		return NewMarkdownFormatter(opts)
	})
	
	Register("jsonl", func(opts Options) Formatter {
		return NewJSONLFormatter(opts)
	})
	Register("json-lines", func(opts Options) Formatter {
		return NewJSONLFormatter(opts)
	})
	
	Register("xml", func(opts Options) Formatter {
		return NewXMLFormatter(opts)
	})
	
	Register("json", func(opts Options) Formatter {
		return NewJSONStructuredFormatter(opts)
	})
	Register("json-structured", func(opts Options) Formatter {
		return NewJSONStructuredFormatter(opts)
	})
}