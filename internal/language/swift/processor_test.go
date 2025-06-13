package swift

import (
	"context"
	"fmt"
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

func TestSwiftProcessor_Constructs(t *testing.T) {
	testCases := []struct {
		name     string
		file     string
		opts     processor.ProcessOptions
		expected string
	}{
		// Construct 1: Basic Fundamentals
		{
			name: "construct1_full",
			file: "construct1_basic.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct1_basic_full.txt",
		},
		{
			name: "construct1_no_private",
			file: "construct1_basic.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        false,
				IncludeImplementation: true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct1_basic_no_private.txt",
		},
		{
			name: "construct1_no_impl",
			file: "construct1_basic.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct1_basic_no_impl.txt",
		},
		
		// Construct 2: Value Types & State
		{
			name: "construct2_full",
			file: "construct2_value_types.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct2_value_types_full.txt",
		},
		{
			name: "construct2_no_private",
			file: "construct2_value_types.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        false,
				IncludeImplementation: true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct2_value_types_no_private.txt",
		},
		{
			name: "construct2_no_impl",
			file: "construct2_value_types.swift",
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
			expected: "expected/construct2_value_types_no_impl.txt",
		},
	}

	proc := NewProcessor()
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read test file
			testFile := filepath.Join("../../../test-data/swift", tc.file)
			reader, err := os.Open(testFile)
			require.NoError(t, err)
			defer reader.Close()

			// Process the file
			result, err := proc.ProcessWithOptions(ctx, reader, tc.file, tc.opts)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Read expected output
			expectedFile := filepath.Join("../../../test-data/swift", tc.expected)
			expectedContent, err := os.ReadFile(expectedFile)
			require.NoError(t, err)

			// Compare with expected output
			// Note: This is a simplified comparison. In real tests, we would
			// serialize the IR to text format and compare
			actualContent := formatDistilledFile(result, tc.opts)
			assert.Equal(t, strings.TrimSpace(string(expectedContent)), strings.TrimSpace(actualContent))
		})
	}
}

// formatDistilledFile converts the distilled file to text format for comparison
func formatDistilledFile(file *ir.DistilledFile, opts processor.ProcessOptions) string {
	var sb strings.Builder
	formatter := formatter.NewTextFormatter(formatter.Options{})
	err := formatter.Format(&sb, file)
	if err != nil {
		return fmt.Sprintf("Error formatting: %v", err)
	}
	return sb.String()
}