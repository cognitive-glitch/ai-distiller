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


func TestShouldRemoveByVisibility(t *testing.T) {
	tests := []struct {
		name       string
		options    Options
		nodeName   string
		visibility ir.Visibility
		expected   bool
	}{
		// Legacy RemovePrivate behavior (removes both private and protected)
		{"LegacyPrivate", Options{RemovePrivate: true}, "MyClass", ir.VisibilityPrivate, true},
		{"LegacyProtected", Options{RemovePrivate: true}, "MyClass", ir.VisibilityProtected, true},
		{"LegacyInternal", Options{RemovePrivate: true}, "MyClass", ir.VisibilityInternal, true},
		{"LegacyFilePrivate", Options{RemovePrivate: true}, "MyClass", ir.VisibilityFilePrivate, true},
		{"LegacyPublic", Options{RemovePrivate: true}, "MyClass", ir.VisibilityPublic, false},
		{"LegacyOpen", Options{RemovePrivate: true}, "MyClass", ir.VisibilityOpen, false},

		// RemovePrivateOnly behavior
		{"PrivateOnlyPrivate", Options{RemovePrivateOnly: true}, "MyClass", ir.VisibilityPrivate, true},
		{"PrivateOnlyProtected", Options{RemovePrivateOnly: true}, "MyClass", ir.VisibilityProtected, false},
		{"PrivateOnlyPublic", Options{RemovePrivateOnly: true}, "MyClass", ir.VisibilityPublic, false},

		// RemoveProtectedOnly behavior
		{"ProtectedOnlyPrivate", Options{RemoveProtectedOnly: true}, "MyClass", ir.VisibilityPrivate, false},
		{"ProtectedOnlyProtected", Options{RemoveProtectedOnly: true}, "MyClass", ir.VisibilityProtected, true},
		{"ProtectedOnlyPublic", Options{RemoveProtectedOnly: true}, "MyClass", ir.VisibilityPublic, false},

		// Python convention with RemovePrivate
		{"PythonPrivate", Options{RemovePrivate: true}, "_private_func", "", true},
		{"PythonDunder", Options{RemovePrivate: true}, "__init__", "", true},
		{"PythonPublic", Options{RemovePrivate: true}, "public_func", "", false},

		// Python convention with RemovePrivateOnly
		{"PythonPrivateOnly", Options{RemovePrivateOnly: true}, "_private_func", "", true},
		{"PythonPublicOnly", Options{RemovePrivateOnly: true}, "public_func", "", false},

		// No removal options
		{"NoRemoval", Options{}, "MyClass", ir.VisibilityPrivate, false},

		// Edge cases
		{"EmptyName", Options{RemovePrivate: true}, "", "", false},
		{"Underscore", Options{RemovePrivate: true}, "_", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripper := New(tt.options)
			result := stripper.shouldRemoveByVisibility(tt.nodeName, tt.visibility)
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
					&ir.DistilledField{
						BaseNode:   ir.BaseNode{},
						Name:       "protected_field",
						Visibility: ir.VisibilityProtected,
					},
					&ir.DistilledFunction{
						BaseNode:   ir.BaseNode{},
						Name:       "protected_method",
						Visibility: ir.VisibilityProtected,
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
		{
			name: "StripPrivateOnly",
			options: Options{
				RemovePrivateOnly: true,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have private members but should keep protected
				hasProtected := false
				hasPrivate := false

				for _, child := range result.Children {
					if class, ok := child.(*ir.DistilledClass); ok {
						for _, member := range class.Children {
							if field, ok := member.(*ir.DistilledField); ok {
								if field.Name == "protected_field" {
									hasProtected = true
								}
								if field.Name == "_private_field" {
									hasPrivate = true
								}
							}
						}
					}
				}

				assert.True(t, hasProtected, "Should keep protected members")
				assert.False(t, hasPrivate, "Should remove private members")
			},
		},
		{
			name: "StripProtectedOnly",
			options: Options{
				RemoveProtectedOnly: true,
			},
			checkFunc: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have protected members but should keep private
				hasProtected := false
				hasPrivate := false

				for _, child := range result.Children {
					if class, ok := child.(*ir.DistilledClass); ok {
						for _, member := range class.Children {
							if field, ok := member.(*ir.DistilledField); ok {
								if field.Name == "protected_field" {
									hasProtected = true
								}
								if field.Name == "_private_field" {
									hasPrivate = true
								}
							}
						}
					}
				}

				assert.False(t, hasProtected, "Should remove protected members")
				assert.True(t, hasPrivate, "Should keep private members")
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