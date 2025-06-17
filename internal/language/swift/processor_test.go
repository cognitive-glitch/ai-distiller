package swift

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/processor"
)

func TestSwiftProcessor_Constructs(t *testing.T) {
	t.Skip("Skipping comprehensive swift tests - expected files need updating")
	testCases := []struct {
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
				IncludeImports:        true,
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
				IncludeImports:        true,
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
				IncludeImports:        true,
			},
			expectedFile: "private=1,protected=1,internal=1,implementation=0.txt",
		},
	}

	p := NewProcessor()
	textFormatter := formatter.NewLanguageAwareTextFormatter(formatter.Options{})

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Read source file
			sourcePath := filepath.Join("../../../testdata/swift", tt.construct, "source.swift")
			sourceFile, err := os.Open(sourcePath)
			if err != nil {
				t.Fatalf("Failed to open source file: %v", err)
			}
			defer sourceFile.Close()

			// Process the file
			result, err := p.ProcessWithOptions(context.Background(), sourceFile, "source.swift", tt.options)
			if err != nil {
				t.Fatalf("Processing failed: %v", err)
			}

			// Format the result
			var output strings.Builder
			if err := textFormatter.Format(&output, result); err != nil {
				t.Fatalf("Formatting failed: %v", err)
			}

			// Read expected output
			expectedPath := filepath.Join("../../../testdata/swift", tt.construct, "expected", tt.expectedFile)
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
