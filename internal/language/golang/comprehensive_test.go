package golang

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	formatterPkg "github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/processor"
)

func TestGoConstructs(t *testing.T) {
	tests := []struct {
		name         string
		sourceFile   string
		options      processor.ProcessOptions
		expectedFile string
	}{
		// Construct 1: Basic
		{
			name:       "Construct1_Full",
			sourceFile: "construct_1_basic.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct1_full.txt",
		},
		{
			name:       "Construct1_NoImpl",
			sourceFile: "construct_1_basic.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: false,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct1_no_impl.txt",
		},
		{
			name:       "Construct1_NoPrivate",
			sourceFile: "construct_1_basic.go",
			options: processor.ProcessOptions{
				IncludePrivate:       false,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct1_no_private.txt",
		},
		// Construct 2: Simple
		{
			name:       "Construct2_Full",
			sourceFile: "construct_2_simple.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct2_full.txt",
		},
		{
			name:       "Construct2_NoImpl",
			sourceFile: "construct_2_simple.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: false,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct2_no_impl.txt",
		},
		{
			name:       "Construct2_NoPrivate",
			sourceFile: "construct_2_simple.go",
			options: processor.ProcessOptions{
				IncludePrivate:       false,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct2_no_private.txt",
		},
		// Construct 3: Medium
		{
			name:       "Construct3_Full",
			sourceFile: "construct_3_medium.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct3_full.txt",
		},
		{
			name:       "Construct3_NoImpl",
			sourceFile: "construct_3_medium.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: false,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct3_no_impl.txt",
		},
		{
			name:       "Construct3_NoPrivate",
			sourceFile: "construct_3_medium.go",
			options: processor.ProcessOptions{
				IncludePrivate:       false,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct3_no_private.txt",
		},
		// Construct 4: Complex
		{
			name:       "Construct4_Full",
			sourceFile: "construct_4_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct4_full.txt",
		},
		{
			name:       "Construct4_NoImpl",
			sourceFile: "construct_4_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: false,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct4_no_impl.txt",
		},
		{
			name:       "Construct4_NoPrivate",
			sourceFile: "construct_4_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       false,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct4_no_private.txt",
		},
		// Construct 5: Very Complex
		{
			name:       "Construct5_Full",
			sourceFile: "construct_5_very_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct5_full.txt",
		},
		{
			name:       "Construct5_NoImpl",
			sourceFile: "construct_5_very_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       true,
				IncludeImplementation: false,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct5_no_impl.txt",
		},
		{
			name:       "Construct5_NoPrivate",
			sourceFile: "construct_5_very_complex.go",
			options: processor.ProcessOptions{
				IncludePrivate:       false,
				IncludeImplementation: true,
				IncludeComments:      false,
				IncludeImports:       true,
			},
			expectedFile: "expected/construct5_no_private.txt",
		},
	}

	p := NewProcessor()
	textFormatter := formatterPkg.NewLanguageAwareTextFormatter(formatterPkg.Options{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read source file
			sourcePath := filepath.Join("../../../test-data/go", tt.sourceFile)
			sourceFile, err := os.Open(sourcePath)
			if err != nil {
				t.Fatalf("Failed to open source file: %v", err)
			}
			defer sourceFile.Close()

			// Process the file
			result, err := p.ProcessWithOptions(context.Background(), sourceFile, tt.sourceFile, tt.options)
			if err != nil {
				t.Fatalf("Processing failed: %v", err)
			}

			// Format output
			var actualOutput strings.Builder
			err = textFormatter.Format(&actualOutput, result)
			if err != nil {
				t.Fatalf("Formatting failed: %v", err)
			}

			// Read expected output
			expectedPath := filepath.Join("../../../test-data/go", tt.expectedFile)
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				// If expected file doesn't exist, create it with current output
				if os.IsNotExist(err) {
					err = os.MkdirAll(filepath.Dir(expectedPath), 0755)
					if err != nil {
						t.Fatalf("Failed to create directory: %v", err)
					}
					err = os.WriteFile(expectedPath, []byte(actualOutput.String()), 0644)
					if err != nil {
						t.Fatalf("Failed to create expected file: %v", err)
					}
					t.Skipf("Created expected file: %s", expectedPath)
				}
				t.Fatalf("Failed to read expected file: %v", err)
			}

			expected := string(expectedBytes)
			actual := actualOutput.String()

			// Compare outputs
			if actual != expected {
				t.Errorf("Output mismatch\nExpected:\n%s\n\nActual:\n%s", expected, actual)
				
				// Write actual output for debugging
				actualPath := expectedPath + ".actual"
				os.WriteFile(actualPath, []byte(actual), 0644)
				t.Logf("Actual output written to: %s", actualPath)
			}
		})
	}
}

