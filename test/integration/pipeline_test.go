package integration

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineSimpleFile(t *testing.T) {
	// Test the complete pipeline with a simple Python file
	testFile := filepath.Join("..", "testdata", "simple.py")
	
	// Step 1: Process the file
	pythonProcessor := python.NewProcessor()
	
	processOpts := processor.ProcessOptions{
		IncludeComments:       true,
		IncludeImplementation: true,
		IncludeImports:        true,
		IncludePrivate:        true,
		SymbolResolution:      true,
	}
	
	distilledFile, err := pythonProcessor.ProcessFile(testFile, processOpts)
	require.NoError(t, err)
	require.NotNil(t, distilledFile)
	
	// Verify basic structure
	assert.Equal(t, testFile, distilledFile.Path)
	assert.Equal(t, "python", distilledFile.Language)
	assert.NotEmpty(t, distilledFile.Children)
	
	// Step 2: Apply stripper visitor with different configurations
	tests := []struct {
		name    string
		options stripper.Options
		check   func(t *testing.T, result *ir.DistilledFile)
	}{
		{
			name: "strip private members",
			options: stripper.Options{
				RemovePrivate: true,
			},
			check: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have _history field or _transform method
				hasPrivate := false
				ir.Walk(result, func(node ir.DistilledNode) bool {
					switch n := node.(type) {
					case *ir.DistilledField:
						if strings.HasPrefix(n.Name, "_") {
							hasPrivate = true
						}
					case *ir.DistilledFunction:
						if strings.HasPrefix(n.Name, "_") && n.Name != "__init__" {
							hasPrivate = true
						}
					}
					return true
				})
				assert.False(t, hasPrivate, "Should not have private members")
			},
		},
		{
			name: "strip implementations",
			options: stripper.Options{
				RemoveImplementations: true,
			},
			check: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have implementations
				hasImpl := false
				ir.Walk(result, func(node ir.DistilledNode) bool {
					if fn, ok := node.(*ir.DistilledFunction); ok {
						if fn.Implementation != "" {
							hasImpl = true
						}
					}
					return true
				})
				assert.False(t, hasImpl, "Should not have implementations")
			},
		},
		{
			name: "strip comments",
			options: stripper.Options{
				RemoveComments: true,
			},
			check: func(t *testing.T, result *ir.DistilledFile) {
				// Should not have comments
				hasComments := false
				ir.Walk(result, func(node ir.DistilledNode) bool {
					if _, ok := node.(*ir.DistilledComment); ok {
						hasComments = true
					}
					return true
				})
				assert.False(t, hasComments, "Should not have comments")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripperVisitor := stripper.New(tt.options)
			strippedFile := distilledFile.Accept(stripperVisitor).(*ir.DistilledFile)
			tt.check(t, strippedFile)
		})
	}
	
	// Step 3: Format the output
	formatTests := []struct {
		name      string
		formatter string
		options   formatter.Options
		contains  []string
	}{
		{
			name:      "markdown output",
			formatter: "markdown",
			options: formatter.Options{
				IncludeLocation: true,
			},
			contains: []string{
				"# " + testFile,
				"## Structure",
				"🏛️ **Class** `ExampleClass`",
				"🔧 **Function** `example_method`",
				"<sub>L", // Location info
			},
		},
		{
			name:      "json output",
			formatter: "json",
			options:   formatter.Options{},
			contains: []string{
				`"type": "file"`,
				`"path": "` + testFile + `"`,
				`"classes"`,
				`"functions"`,
				`"ExampleClass"`,
			},
		},
		{
			name:      "xml output",
			formatter: "xml",
			options:   formatter.Options{},
			contains: []string{
				`<?xml version="1.0" encoding="UTF-8"?>`,
				`<distilled>`,
				`<class name="ExampleClass"`,
				`<function name="example_method"`,
				`</distilled>`,
			},
		},
		{
			name:      "jsonl output",
			formatter: "jsonl",
			options:   formatter.Options{},
			contains: []string{
				`"type":"file"`,
				`"type":"class"`,
				`"type":"function"`,
			},
		},
	}
	
	for _, tt := range formatTests {
		t.Run(tt.name, func(t *testing.T) {
			fmt, err := formatter.Get(tt.formatter, tt.options)
			require.NoError(t, err)
			
			var buf bytes.Buffer
			err = fmt.Format(&buf, distilledFile)
			require.NoError(t, err)
			
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestPipelineComplexFile(t *testing.T) {
	// Test with complex Python file
	testFile := filepath.Join("..", "testdata", "complex.py")
	
	pythonProcessor := python.NewProcessor()
	processOpts := processor.ProcessOptions{
		IncludeComments:       true,
		IncludeImplementation: true,
		IncludeImports:        true,
		IncludePrivate:        true,
		SymbolResolution:      true,
	}
	
	distilledFile, err := pythonProcessor.ProcessFile(testFile, processOpts)
	require.NoError(t, err)
	
	// Since we're using mock data, just check basic structure
	var classCount, functionCount int
	
	ir.Walk(distilledFile, func(node ir.DistilledNode) bool {
		switch node.(type) {
		case *ir.DistilledClass:
			classCount++
		case *ir.DistilledFunction:
			functionCount++
		}
		return true
	})
	
	// Mock returns at least one class and function
	assert.GreaterOrEqual(t, classCount, 1, "Should have at least one class")
	assert.GreaterOrEqual(t, functionCount, 1, "Should have at least one function")
}

func TestPipelineMultipleFiles(t *testing.T) {
	// Test formatting multiple files
	files := []string{
		filepath.Join("..", "testdata", "simple.py"),
		filepath.Join("..", "testdata", "complex.py"),
	}
	
	pythonProcessor := python.NewProcessor()
	processOpts := processor.ProcessOptions{
		IncludeComments:       false,
		IncludeImplementation: false,
		IncludeImports:        true,
	}
	
	var distilledFiles []*ir.DistilledFile
	for _, file := range files {
		df, err := pythonProcessor.ProcessFile(file, processOpts)
		require.NoError(t, err)
		distilledFiles = append(distilledFiles, df)
	}
	
	// Test multi-file formatting
	fmt, err := formatter.Get("json", formatter.Options{})
	require.NoError(t, err)
	
	var buf bytes.Buffer
	err = fmt.FormatMultiple(&buf, distilledFiles)
	require.NoError(t, err)
	
	output := buf.String()
	assert.Contains(t, output, `"type": "project"`)
	assert.Contains(t, output, `"files"`)
	assert.Contains(t, output, `"total_stats"`)
	assert.Contains(t, output, "simple.py")
	assert.Contains(t, output, "complex.py")
}

func TestPipelineErrorHandling(t *testing.T) {
	// Test error handling
	pythonProcessor := python.NewProcessor()
	
	// Non-existent file
	_, err := pythonProcessor.ProcessFile("nonexistent.py", processor.ProcessOptions{})
	assert.Error(t, err)
	
	// Invalid formatter
	_, err = formatter.Get("invalid", formatter.Options{})
	assert.Error(t, err)
}

// Helper function for debugging
func printIR(t *testing.T, file *ir.DistilledFile) {
	t.Helper()
	fmt, _ := formatter.Get("markdown", formatter.Options{
		IncludeLocation: true,
		Compact:         false,
	})
	var buf bytes.Buffer
	fmt.Format(&buf, file)
	t.Log(buf.String())
}