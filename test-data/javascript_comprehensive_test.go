package test_data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		constructFile  string
		stripOptions   processor.ProcessOptions
		expectedFile   string
		description    string
	}{
		// Construct 1: Scope & Closure Gauntlet
		{
			name:          "Construct1_Full",
			constructFile: "javascript/construct_1_scope.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct1_full.txt",
			description:  "Full output with all scope, hoisting, and closure patterns",
		},
		{
			name:          "Construct1_NoPrivate",
			constructFile: "javascript/construct_1_scope.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct1_no_private.txt",
			description:  "Public members only",
		},
		{
			name:          "Construct1_NoImplementation",
			constructFile: "javascript/construct_1_scope.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct1_no_impl.txt",
			description:  "Signatures only, no implementations",
		},
		// Construct 2: The Asynchronous Labyrinth
		{
			name:          "Construct2_Full",
			constructFile: "javascript/construct_2_async.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct2_full.txt",
			description:  "Full async/await, generators, and Promise patterns",
		},
		{
			name:          "Construct2_NoPrivate",
			constructFile: "javascript/construct_2_async.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct2_no_private.txt",
			description:  "Public async members only",
		},
		{
			name:          "Construct2_NoImplementation",
			constructFile: "javascript/construct_2_async.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct2_no_impl.txt",
			description:  "Async signatures only",
		},
		// Construct 3: The Module System Bridge
		{
			name:          "Construct3_Full",
			constructFile: "javascript/construct_3_modules.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct3_full.txt",
			description:  "Full ESM module with CJS interop",
		},
		{
			name:          "Construct3_NoPrivate",
			constructFile: "javascript/construct_3_modules.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct3_no_private.txt",
			description:  "Public exports only",
		},
		{
			name:          "Construct3_NoImplementation",
			constructFile: "javascript/construct_3_modules.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct3_no_impl.txt",
			description:  "Module structure only",
		},
		// Construct 4: React/JSX Component Tree
		{
			name:          "Construct4_Full",
			constructFile: "javascript/construct_4_react.jsx",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct4_full.txt",
			description:  "Full React components with hooks and JSX",
		},
		{
			name:          "Construct4_NoPrivate",
			constructFile: "javascript/construct_4_react.jsx",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct4_no_private.txt",
			description:  "Public React components only",
		},
		{
			name:          "Construct4_NoImplementation",
			constructFile: "javascript/construct_4_react.jsx",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct4_no_impl.txt",
			description:  "React component signatures only",
		},
		// Construct 5: The Metaprogramming Minefield
		{
			name:          "Construct5_Full",
			constructFile: "javascript/construct_5_meta.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct5_full.txt",
			description:  "Full metaprogramming with Proxy, Symbol, and Reflect",
		},
		{
			name:          "Construct5_NoPrivate",
			constructFile: "javascript/construct_5_meta.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct5_no_private.txt",
			description:  "Public metaprogramming constructs only",
		},
		{
			name:          "Construct5_NoImplementation",
			constructFile: "javascript/construct_5_meta.js",
			stripOptions: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expectedFile: "javascript/expected/construct5_no_impl.txt",
			description:  "Metaprogramming signatures only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the construct source file
			sourceFile := filepath.Join("test-data", tt.constructFile)
			source, err := os.ReadFile(sourceFile)
			require.NoError(t, err, "Failed to read construct file")

			// Process the file
			result, err := processJavaScriptFile(t, string(source), tt.constructFile, tt.stripOptions)
			require.NoError(t, err, "Failed to process JavaScript file")

			// Read expected output
			expectedFile := filepath.Join("test-data", tt.expectedFile)
			expected, err := os.ReadFile(expectedFile)
			require.NoError(t, err, "Failed to read expected file")

			// Normalize line endings
			actualNormalized := normalizeLineEndings(result)
			expectedNormalized := normalizeLineEndings(string(expected))

			// Compare outputs
			if actualNormalized != expectedNormalized {
				t.Errorf("Output mismatch for %s\nExpected:\n%s\nActual:\n%s",
					tt.description, expectedNormalized, actualNormalized)

				// Save actual output for debugging
				debugFile := filepath.Join("test-data", "javascript", "debug",
					strings.ReplaceAll(tt.name, "/", "_")+".actual.txt")
				os.MkdirAll(filepath.Dir(debugFile), 0755)
				os.WriteFile(debugFile, []byte(result), 0644)
			}
		})
	}
}

func processJavaScriptFile(t *testing.T, source, filename string, opts processor.ProcessOptions) (string, error) {
	// Get the JavaScript processor
	proc := processor.GetByLanguage("javascript")
	require.NotNil(t, proc, "JavaScript processor not found")

	// Process the file
	ctx := context.Background()
	reader := strings.NewReader(source)
	file, err := proc.ProcessWithOptions(ctx, reader, filename, opts)
	if err != nil {
		return "", err
	}

	// Format output as text
	var buf strings.Builder
	formatter := &TextFormatter{}
	err = formatter.Format(&buf, file)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func normalizeLineEndings(s string) string {
	// Normalize all line endings to \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Trim trailing whitespace
	s = strings.TrimSpace(s)
	return s
}