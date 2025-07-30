package java

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaConstructs(t *testing.T) {
	tests := []struct {
		name         string
		construct    string
		options      processor.ProcessOptions
		expectedFile string
	}{
		// Basic construct tests
		{
			name:      "Basic_Default",
			construct: "01_basic",
			options: processor.ProcessOptions{
				IncludePrivate:        false,
				IncludeImplementation: false,
				IncludeComments:       false,
				IncludeDocstrings:     true,
				IncludeImports:        true,
				IncludeFields:         true,  // Fixed: default should include fields
				IncludeMethods:        true,  // Fixed: default should include methods
			},
			expectedFile: "default.txt",
		},
		{
			name:      "Basic_WithImplementation",
			construct: "01_basic",
			options: processor.ProcessOptions{
				IncludePrivate:        false,
				IncludeImplementation: true,
				IncludeComments:       false,
				IncludeDocstrings:     true,
				IncludeImports:        true,
				IncludeFields:         true,  // Fixed: default should include fields
				IncludeMethods:        true,  // Fixed: default should include methods
			},
			expectedFile: "implementation=1.txt",
		},
		{
			name:      "Basic_WithPrivate",
			construct: "01_basic",
			options: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: false,
				IncludeComments:       false,
				IncludeDocstrings:     true,
				IncludeImports:        true,
				IncludeFields:         true,  // Fixed: default should include fields
				IncludeMethods:        true,  // Fixed: default should include methods
			},
			expectedFile: "private=1,protected=1,internal=1,implementation=0.txt",
		},
	}

	p := NewProcessor()
	textFormatter := formatter.NewLanguageAwareTextFormatter(formatter.Options{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read source file
			sourcePath := filepath.Join("../../../testdata/java", tt.construct, "source.java")
			sourceFile, err := os.Open(sourcePath)
			if err != nil {
				t.Fatalf("Failed to open source file: %v", err)
			}
			defer sourceFile.Close()

			// Process the file
			result, err := p.ProcessWithOptions(context.Background(), sourceFile, "source.java", tt.options)
			if err != nil {
				t.Fatalf("Processing failed: %v", err)
			}

			// Format the result
			var output strings.Builder
			if err := textFormatter.Format(&output, result); err != nil {
				t.Fatalf("Formatting failed: %v", err)
			}

			// Read expected output
			expectedPath := filepath.Join("../../../testdata/java", tt.construct, "expected", tt.expectedFile)
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			expected := strings.TrimSpace(string(expectedBytes))
			actual := strings.TrimSpace(output.String())

			if expected != actual {
				t.Errorf("Output mismatch for %s:\nExpected:\n%s\n\nActual:\n%s", tt.name, expected, actual)
			}
		})
	}
}

// TestJavaParserFeatures tests specific Java language features
func TestJavaParserFeatures(t *testing.T) {
	t.Run("Generics", func(t *testing.T) {
		source := `
package test;

public interface Container<T> {
    void add(T item);
    T get(int index);
}

public class StringContainer implements Container<String> {
    public void add(String item) { }
    public String get(int index) { return null; }
}
`
		proc := NewProcessor()
		ctx := context.Background()
		result, err := proc.Process(ctx, strings.NewReader(source), "test.java")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check that we have an interface with type parameters
		var foundInterface bool
		for _, child := range result.Children {
			if iface, ok := child.(*ir.DistilledInterface); ok {
				assert.Equal(t, "Container", iface.Name)
				// TODO: Check type parameters when parser supports them
				foundInterface = true
			}
		}
		assert.True(t, foundInterface, "Interface not found")
	})

	t.Run("Annotations", func(t *testing.T) {
		source := `
package test;

public class Test {
    @Override
    public String toString() {
        return "test";
    }
    
    @Deprecated
    @SuppressWarnings("unchecked")
    public void oldMethod() { }
}
`
		proc := NewProcessor()
		ctx := context.Background()
		result, err := proc.Process(ctx, strings.NewReader(source), "test.java")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Find the class
		var foundClass *ir.DistilledClass
		for _, child := range result.Children {
			if class, ok := child.(*ir.DistilledClass); ok {
				foundClass = class
				break
			}
		}
		require.NotNil(t, foundClass)

		// Check methods have decorators
		var toStringMethod, oldMethod *ir.DistilledFunction
		for _, child := range foundClass.Children {
			if fn, ok := child.(*ir.DistilledFunction); ok {
				if fn.Name == "toString" {
					toStringMethod = fn
				} else if fn.Name == "oldMethod" {
					oldMethod = fn
				}
			}
		}

		require.NotNil(t, toStringMethod)
		assert.Contains(t, toStringMethod.Decorators, "@Override")

		require.NotNil(t, oldMethod)
		assert.Contains(t, oldMethod.Decorators, "@Deprecated")
		assert.Contains(t, oldMethod.Decorators, "@SuppressWarnings(\"unchecked\")")
	})

	t.Run("Sealed Classes", func(t *testing.T) {
		source := `
package test;

public sealed class Shape permits Circle, Square { }
final class Circle extends Shape { }
final class Square extends Shape { }
`
		proc := NewProcessor()
		ctx := context.Background()
		result, err := proc.Process(ctx, strings.NewReader(source), "test.java")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check we found all three classes
		classCount := 0
		for _, child := range result.Children {
			if _, ok := child.(*ir.DistilledClass); ok {
				classCount++
			}
		}
		assert.Equal(t, 3, classCount, "Expected 3 classes")
	})

	t.Run("Records", func(t *testing.T) {
		source := `
package test;

public record Point(int x, int y) {
    public Point {
        if (x < 0 || y < 0) {
            throw new IllegalArgumentException();
        }
    }
}
`
		proc := NewProcessor()
		ctx := context.Background()
		result, err := proc.Process(ctx, strings.NewReader(source), "test.java")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Records are represented as classes with IsRecord flag in Java extensions
		var foundRecord *ir.DistilledClass
		for _, child := range result.Children {
			if class, ok := child.(*ir.DistilledClass); ok {
				if class.Extensions != nil &&
					class.Extensions.Java != nil &&
					class.Extensions.Java.IsRecord {
					foundRecord = class
					break
				}
			}
		}
		assert.NotNil(t, foundRecord, "Record not found")
		assert.Equal(t, "Point", foundRecord.Name)
		assert.True(t, foundRecord.Extensions.Java.IsRecord, "IsRecord flag should be true")
	})
}
