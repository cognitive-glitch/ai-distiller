package stripper

import (
	"testing"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/stretchr/testify/assert"
)

func TestNewStripper(t *testing.T) {
	opts := Options{
		RemoveComments: true,
		RemovePrivate: true,
	}
	stripper := New(opts)
	
	assert.NotNil(t, stripper)
	assert.Equal(t, opts, stripper.options)
}


func TestIsPrivate(t *testing.T) {
	stripper := New(Options{})
	
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

func TestVisit(t *testing.T) {
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
		options   Options
		checkFunc func(t *testing.T, result *ir.DistilledFile)
	}{
		{
			name: "StripComments",
			options: Options{
				RemoveComments: true,
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
			options: Options{
				RemoveImports: true,
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
			options: Options{
				RemovePrivate: true,
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
			options: Options{
				RemoveImplementations: true,
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
			stripper := New(tt.options)
			walker := ir.NewWalker(stripper)
			result := walker.Walk(file)
			
			assert.NotNil(t, result)
			resultFile := result.(*ir.DistilledFile)
			assert.Equal(t, file.Path, resultFile.Path)
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, resultFile)
			}
		})
	}
}