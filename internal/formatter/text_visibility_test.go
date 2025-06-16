package formatter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/stretchr/testify/assert"
)

func TestTextFormatter_VisibilityPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		node     ir.DistilledNode
		expected string
	}{
		{
			name: "public function no prefix",
			node: &ir.DistilledFile{
				Path: "test.py",
				Children: []ir.DistilledNode{
					&ir.DistilledFunction{
						Name:       "public_func",
						Visibility: ir.VisibilityPublic,
					},
				},
			},
			expected: "public_func()",
		},
		{
			name: "private function with prefix",
			node: &ir.DistilledFile{
				Path: "test.py",
				Children: []ir.DistilledNode{
					&ir.DistilledFunction{
						Name:       "private_func",
						Visibility: ir.VisibilityPrivate,
					},
				},
			},
			expected: "-private_func()",
		},
		{
			name: "protected function with prefix",
			node: &ir.DistilledFile{
				Path: "test.py",
				Children: []ir.DistilledNode{
					&ir.DistilledFunction{
						Name:       "protected_func",
						Visibility: ir.VisibilityProtected,
					},
				},
			},
			expected: "*protected_func()",
		},
		{
			name: "mixed visibility in class",
			node: &ir.DistilledFile{
				Path: "test.php",
				Children: []ir.DistilledNode{
					&ir.DistilledClass{
						Name:       "TestClass",
						Visibility: ir.VisibilityPublic,
						Children: []ir.DistilledNode{
							&ir.DistilledField{
								Name:       "publicField",
								Visibility: ir.VisibilityPublic,
								Type:       &ir.TypeRef{Name: "string"},
							},
							&ir.DistilledField{
								Name:       "protectedField",
								Visibility: ir.VisibilityProtected,
								Type:       &ir.TypeRef{Name: "int"},
							},
							&ir.DistilledField{
								Name:       "privateField",
								Visibility: ir.VisibilityPrivate,
								Type:       &ir.TypeRef{Name: "bool"},
							},
							&ir.DistilledFunction{
								Name:       "publicMethod",
								Visibility: ir.VisibilityPublic,
							},
							&ir.DistilledFunction{
								Name:       "protectedMethod",
								Visibility: ir.VisibilityProtected,
							},
							&ir.DistilledFunction{
								Name:       "privateMethod",
								Visibility: ir.VisibilityPrivate,
							},
						},
					},
				},
			},
			expected: `class TestClass:
    publicField: string
    *protectedField: int
    -privateField: bool
    publicMethod()
    *protectedMethod()
    -privateMethod()`,
		},
		{
			name: "internal visibility distinct from private",
			node: &ir.DistilledFile{
				Path: "test.go",
				Children: []ir.DistilledNode{
					&ir.DistilledFunction{
						Name:       "internalFunc",
						Visibility: ir.VisibilityInternal,
					},
				},
			},
			expected: "~internalFunc()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewTextFormatter(Options{})
			var buf bytes.Buffer
			
			if file, ok := tt.node.(*ir.DistilledFile); ok {
				err := formatter.Format(&buf, file)
				assert.NoError(t, err)
			}
			
			output := strings.TrimSpace(buf.String())
			// Remove file tags for comparison
			output = strings.Replace(output, "<file path=\""+tt.node.(*ir.DistilledFile).Path+"\">", "", 1)
			output = strings.Replace(output, "</file>", "", 1)
			output = strings.TrimSpace(output)
			
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestGetVisibilityPrefix(t *testing.T) {
	tests := []struct {
		visibility ir.Visibility
		expected   string
	}{
		{ir.VisibilityPublic, ""},       // No prefix for public
		{ir.VisibilityPrivate, "-"},     // Private
		{ir.VisibilityProtected, "*"},   // Protected
		{ir.VisibilityInternal, "~"},    // UML package/internal
		{ir.VisibilityFilePrivate, "-"}, // Swift fileprivate -> similar to private
		{ir.VisibilityOpen, ""},         // Swift open -> treat as public
		{ir.VisibilityProtectedInternal, "*~"}, // C# protected internal -> combination
		{ir.VisibilityPrivateProtected, "-*"},  // C# private protected -> combination
		{"", ""},                        // Default to public (no prefix)
	}

	for _, tt := range tests {
		t.Run(string(tt.visibility), func(t *testing.T) {
			result := getVisibilityPrefix(tt.visibility)
			assert.Equal(t, tt.expected, result)
		})
	}
}