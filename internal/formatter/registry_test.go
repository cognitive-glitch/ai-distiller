package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	// Create a new registry for testing
	r := &Registry{
		formatters: make(map[string]func(Options) Formatter),
	}

	// Test registration
	err := r.Register("test", func(opts Options) Formatter {
		return NewMarkdownFormatter(opts)
	})
	require.NoError(t, err)

	// Test duplicate registration
	err = r.Register("test", func(opts Options) Formatter {
		return NewMarkdownFormatter(opts)
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test case insensitive
	err = r.Register("TEST", func(opts Options) Formatter {
		return NewMarkdownFormatter(opts)
	})
	assert.Error(t, err)

	// Test Get
	formatter, err := r.Get("test", Options{})
	require.NoError(t, err)
	assert.NotNil(t, formatter)
	assert.IsType(t, &MarkdownFormatter{}, formatter)

	// Test Get case insensitive
	formatter, err = r.Get("TEST", Options{})
	require.NoError(t, err)
	assert.NotNil(t, formatter)

	// Test Get not found
	_, err = r.Get("nonexistent", Options{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test List
	r.Register("another", func(opts Options) Formatter {
		return NewXMLFormatter(opts)
	})

	names := r.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "test")
	assert.Contains(t, names, "another")
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default formatters are registered
	names := List()
	assert.Contains(t, names, "markdown")
	assert.Contains(t, names, "md")
	assert.Contains(t, names, "jsonl")
	assert.Contains(t, names, "json-lines")
	assert.Contains(t, names, "xml")
	assert.Contains(t, names, "json")
	assert.Contains(t, names, "json-structured")

	// Test getting formatters
	tests := []struct {
		name     string
		expected interface{}
	}{
		{"markdown", &MarkdownFormatter{}},
		{"md", &MarkdownFormatter{}},
		{"jsonl", &JSONLFormatter{}},
		{"json-lines", &JSONLFormatter{}},
		{"xml", &XMLFormatter{}},
		{"json", &JSONStructuredFormatter{}},
		{"json-structured", &JSONStructuredFormatter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := Get(tt.name, Options{})
			require.NoError(t, err)
			assert.IsType(t, tt.expected, formatter)
		})
	}

	// Test with options
	formatter, err := Get("markdown", Options{
		IncludeLocation: true,
		Compact:         true,
	})
	require.NoError(t, err)
	mdFormatter := formatter.(*MarkdownFormatter)
	assert.True(t, mdFormatter.options.IncludeLocation)
	assert.True(t, mdFormatter.options.Compact)
}
