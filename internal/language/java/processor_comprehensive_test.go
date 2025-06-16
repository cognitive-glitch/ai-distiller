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

// TestJavaConstructs tests all 5 Java constructs with 3 different options
func TestJavaConstructs(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "Construct 1 - Basic",
			filename: "Basic.java",
		},
		{
			name:     "Construct 2 - SimpleOOP",
			filename: "SimpleOOP.java",
		},
		{
			name:     "Construct 3 - GenericsAndInterfaces",
			filename: "GenericsAndInterfaces.java",
		},
		{
			name:     "Construct 4 - ModernJava",
			filename: "ModernJava.java",
		},
		{
			name:     "Construct 5 - Advanced",
			filename: "Advanced.java",
		},
	}

	options := []struct {
		name           string
		opts           processor.ProcessOptions
		expectedSuffix string
	}{
		{
			name: "Full output",
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        true,
			},
			expectedSuffix: "_expected_full.txt",
		},
		{
			name: "No private members",
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        false,
			},
			expectedSuffix: "_expected_no_private.txt",
		},
		{
			name: "No implementation",
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: false,
				IncludePrivate:        true,
			},
			expectedSuffix: "_expected_no_impl.txt",
		},
	}

	// Get test data directory
	testDataDir := filepath.Join("..", "..", "..", "test-data", "java")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := filepath.Join(testDataDir, test.filename)

			// Check if source file exists
			if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
				t.Skipf("Source file %s does not exist", sourceFile)
			}

			// Create processor
			proc := NewProcessor()

			for _, opt := range options {
				t.Run(opt.name, func(t *testing.T) {
					// Read source file
					sourceContent, err := os.ReadFile(sourceFile)
					require.NoError(t, err)

					// Process the file
					ctx := context.Background()
					result, err := proc.ProcessWithOptions(ctx, strings.NewReader(string(sourceContent)), test.filename, opt.opts)
					require.NoError(t, err)
					require.NotNil(t, result)

					// Format using text formatter
					textFormatter := formatter.NewLanguageAwareTextFormatter(formatter.Options{})
					
					var output strings.Builder
					err = textFormatter.Format(&output, result)
					require.NoError(t, err)

					// Get expected output file
					baseName := strings.TrimSuffix(test.filename, ".java")
					expectedFile := filepath.Join(testDataDir, baseName+opt.expectedSuffix)

					// Check if expected file exists
					if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
						// If expected file doesn't exist, write the actual output for manual review
						actualFile := filepath.Join(testDataDir, baseName+opt.expectedSuffix+".actual")
						err = os.WriteFile(actualFile, []byte(output.String()), 0644)
						require.NoError(t, err)
						t.Skipf("Expected file %s does not exist. Actual output written to %s", expectedFile, actualFile)
					}

					// Read expected output
					expectedContent, err := os.ReadFile(expectedFile)
					require.NoError(t, err)

					// Compare outputs
					assert.Equal(t, strings.TrimSpace(string(expectedContent)), strings.TrimSpace(output.String()),
						"Output mismatch for %s with %s", test.filename, opt.name)
				})
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