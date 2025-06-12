package stripper

import (
	"testing"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/stretchr/testify/assert"
)

func TestNewStripper(t *testing.T) {
	opts := processor.DefaultProcessOptions()
	stripper := NewStripper(opts)
	
	assert.NotNil(t, stripper)
	assert.Equal(t, opts, stripper.options)
}

func TestFromStrings(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		expected StripOptions
	}{
		{
			name:    "Empty",
			options: []string{},
			expected: StripOptions{},
		},
		{
			name:    "Comments",
			options: []string{"comments"},
			expected: StripOptions{Comments: true},
		},
		{
			name:    "All",
			options: []string{"comments", "imports", "implementation", "non-public"},
			expected: StripOptions{
				Comments:       true,
				Imports:        true,
				Implementation: true,
				NonPublic:      true,
			},
		},
		{
			name:    "Unknown",
			options: []string{"unknown", "comments"},
			expected: StripOptions{Comments: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromStrings(tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToProcessOptions(t *testing.T) {
	stripOpts := StripOptions{
		Comments:       true,
		Imports:        true,
		Implementation: true,
		NonPublic:      true,
	}

	procOpts := stripOpts.ToProcessOptions()
	
	assert.False(t, procOpts.IncludeComments)
	assert.False(t, procOpts.IncludeImports)
	assert.False(t, procOpts.IncludeImplementation)
	assert.False(t, procOpts.IncludePrivate)
	assert.Equal(t, 100, procOpts.MaxDepth)
	assert.True(t, procOpts.SymbolResolution)
	assert.True(t, procOpts.IncludeLineNumbers)
}

func TestIsPrivate(t *testing.T) {
	stripper := NewStripper(processor.DefaultProcessOptions())
	
	tests := []struct {
		name       string
		nodeName   string
		visibility ir.Visibility
		expected   bool
	}{
		// Explicit visibility
		{"ExplicitPrivate", "MyClass", ir.VisibilityPrivate, true},
		{"ExplicitInternal", "MyClass", ir.VisibilityInternal, true},
		{"ExplicitFilePrivate", "MyClass", ir.VisibilityFilePrivate, true},
		{"ExplicitPublic", "MyClass", ir.VisibilityPublic, false},
		{"ExplicitOpen", "MyClass", ir.VisibilityOpen, false},
		
		// Python convention
		{"PythonPrivate", "_private_func", "", true},
		{"PythonDunder", "__init__", "", true},
		{"PythonPublic", "public_func", "", false},
		
		// Go convention would need language context
		{"GoLowercase", "privateFunc", "", false}, // Without language context, defaults to public
		{"GoUppercase", "PublicFunc", "", false},
		
		// Edge cases
		{"EmptyName", "", "", false},
		{"Underscore", "_", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripper.isPrivate(tt.nodeName, tt.visibility)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStrip(t *testing.T) {
	// Create a test file with various nodes
	file := &ir.DistilledFile{
		Path:     "test.py",
		Language: "python",
		Children: []ir.DistilledNode{
			&ir.DistilledComment{
				BaseNode: ir.BaseNode{},
				Text:     "File comment",
			},
			&ir.DistilledImport{
				BaseNode: ir.BaseNode{},
				Module:   "typing",
			},
			&ir.DistilledFunction{
				BaseNode:       ir.BaseNode{},
				Name:           "public_function",
				Visibility:     ir.VisibilityPublic,
				Implementation: "return 42",
			},
			&ir.DistilledFunction{
				BaseNode:       ir.BaseNode{},
				Name:           "_private_function",
				Visibility:     "",
				Implementation: "return secret",
			},
			&ir.DistilledClass{
				BaseNode:   ir.BaseNode{},
				Name:       "PublicClass",
				Visibility: ir.VisibilityPublic,
				Children: []ir.DistilledNode{
					&ir.DistilledField{
						BaseNode:   ir.BaseNode{},
						Name:       "_private_field",
						Visibility: "",
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		options   processor.ProcessOptions
		checkFunc func(t *testing.T, result *ir.DistilledFile)
	}{
		{
			name: "StripComments",
			options: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImports:        true,
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have comments
				for _, child := range result.Children {
					_, isComment := child.(*ir.DistilledComment)
					assert.False(t, isComment, "Should not have comments")
				}
			},
		},
		{
			name: "StripImports",
			options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        false,
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have imports
				for _, child := range result.Children {
					_, isImport := child.(*ir.DistilledImport)
					assert.False(t, isImport, "Should not have imports")
				}
			},
		},
		{
			name: "StripPrivate",
			options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludePrivate:        false,
				IncludeImplementation: true,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have private functions
				for _, child := range result.Children {
					if fn, ok := child.(*ir.DistilledFunction); ok {
						assert.NotEqual(t, "_private_function", fn.Name)
					}
				}
			},
		},
		{
			name: "StripImplementation",
			options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludePrivate:        true,
				IncludeImplementation: false,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Functions should have empty implementation
				for _, child := range result.Children {
					if fn, ok := child.(*ir.DistilledFunction); ok {
						assert.Empty(t, fn.Implementation)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripper := NewStripper(tt.options)
			result := stripper.Strip(file)
			
			assert.NotNil(t, result)
			assert.Equal(t, file.Path, result.Path)
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}