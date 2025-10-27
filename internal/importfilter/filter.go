package importfilter

import (
	"fmt"
	"sync"
)

// ImportFilter defines the interface for language-specific import filtering
type ImportFilter interface {
	// FilterUnusedImports removes unused imports from distilled code
	// Returns filtered code and list of removed imports for logging
	FilterUnusedImports(code string, debugLevel int) (string, []string, error)

	// Language returns the language this filter handles
	Language() string
}

// ImportStatement represents a parsed import statement
type ImportStatement struct {
	FullLine      string            // Original import line(s)
	StartLine     int               // Starting line number (1-based)
	EndLine       int               // Ending line number (1-based)
	ImportedNames []string          // Names imported (for specific imports)
	IsWildcard    bool              // true for "import *" or "from X import *"
	IsSideEffect  bool              // true for side-effect imports like "import 'polyfill'"
	Module        string            // Module/package name
	Aliases       map[string]string // Mapping of alias->original name
}

// Registry for language-specific filters
var (
	filtersMu sync.RWMutex
	filters   = make(map[string]ImportFilter)
)

// Register adds a filter for a specific language
func Register(language string, filter ImportFilter) {
	filtersMu.Lock()
	defer filtersMu.Unlock()
	filters[language] = filter
}

// GetFilter returns the filter for a specific language
func GetFilter(language string) (ImportFilter, error) {
	filtersMu.RLock()
	defer filtersMu.RUnlock()

	filter, ok := filters[language]
	if !ok {
		return nil, fmt.Errorf("no import filter registered for language: %s", language)
	}

	return filter, nil
}

// FilterImports is the main entry point for filtering imports
func FilterImports(code string, language string, debugLevel int) (string, []string, error) {
	filter, err := GetFilter(language)
	if err != nil {
		// If no filter is registered, return code unchanged
		return code, nil, nil
	}

	return filter.FilterUnusedImports(code, debugLevel)
}