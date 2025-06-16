package python

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASTComparator(t *testing.T) {
	t.Skip("Skipping AST comparator tests - test files missing")
	
	// Skip if Python is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Python3 not available, skipping AST comparison tests")
	}

	p := NewProcessor()
	comparator := NewASTComparator()

	testFiles := []struct {
		name        string
		file        string
		expectMatch bool
		description string
	}{
		{
			name:        "basic_class",
			file:        "../../../testdata/input/basic_class.py",
			expectMatch: true,
			description: "Basic class with methods",
		},
		{
			name:        "complex_imports",
			file:        "../../../testdata/input/complex_imports.py",
			expectMatch: true,
			description: "Various import styles",
		},
		{
			name:        "decorators_and_metadata",
			file:        "../../../testdata/input/decorators_and_metadata.py",
			expectMatch: true,
			description: "Decorated functions and classes",
		},
		{
			name:        "nested_structures",
			file:        "../../../testdata/input/nested_structures.py",
			expectMatch: true,
			description: "Nested classes and functions",
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with our parser
			ourAST, err := p.ProcessFile(tt.file, processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			})
			require.NoError(t, err)
			require.NotNil(t, ourAST)

			// Compare with Python AST
			result, err := comparator.CompareFile(tt.file, ourAST)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Log comparison results
			t.Logf("Comparison for %s:", tt.description)
			t.Logf("  Our nodes: %d, Python nodes: %d", result.OurNodeCount, result.PythonNodeCount)
			
			if len(result.MissingInOurs) > 0 {
				t.Logf("  Missing in ours: %v", result.MissingInOurs)
			}
			if len(result.MissingInPython) > 0 {
				t.Logf("  Missing in Python: %v", result.MissingInPython)
			}
			if len(result.Differences) > 0 {
				t.Logf("  Differences: %v", result.Differences)
			}

			// For now, we don't expect perfect matches due to parser limitations
			// but we should find the main structures
			if tt.expectMatch {
				assert.Empty(t, result.MissingInOurs, "Should not be missing major structures")
			}
		})
	}
}

func TestASTComparatorWithAdvancedFeatures(t *testing.T) {
	t.Skip("Skipping AST comparator advanced tests - test files missing")
	
	// Skip if Python is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Python3 not available, skipping AST comparison tests")
	}

	p := NewProcessor()
	comparator := NewASTComparator()

	// Test with our advanced feature files
	advancedFiles := []string{
		"../../../testdata/input/pattern_matching.py",
		"../../../testdata/input/walrus_operator.py",
		"../../../testdata/input/async_await_syntax.py",
	}

	for _, file := range advancedFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Parse with our parser
			ourAST, err := p.ProcessFile(file, processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			})
			require.NoError(t, err)

			// Compare with Python AST
			result, err := comparator.CompareFile(file, ourAST)
			require.NoError(t, err)

			// Log what we found vs what Python found
			t.Logf("File: %s", filepath.Base(file))
			t.Logf("Our parser found %d nodes, Python found %d nodes", 
				result.OurNodeCount, result.PythonNodeCount)
			
			// We expect to find most structures, even if not all details match
			accuracy := float64(result.OurNodeCount-len(result.MissingInOurs)) / float64(result.PythonNodeCount)
			t.Logf("Accuracy: %.2f%%", accuracy*100)
			
			// Line-based parser should find at least 70% of structures
			assert.Greater(t, accuracy, 0.7, "Parser should find most structures")
		})
	}
}

func TestCompareWithPythonLibraries(t *testing.T) {
	t.Skip("Manual test for comparing with real Python libraries")
	
	// This test can be enabled manually to test against real Python libraries
	// Example usage:
	//
	// 1. Clone a Python library: git clone https://github.com/psf/requests.git
	// 2. Update libraryPath below
	// 3. Run: go test -run TestCompareWithPythonLibraries -v
	
	libraryPath := "/path/to/requests"
	
	p := NewProcessor()
	comparator := NewASTComparator()
	
	// Walk through all Python files
	err := filepath.Walk(libraryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if filepath.Ext(path) == ".py" {
			t.Run(path, func(t *testing.T) {
				// Parse with our parser
				ourAST, err := p.ProcessFile(path, processor.ProcessOptions{
					IncludeComments:       true,
					IncludeImplementation: true,
					IncludeImports:        true,
					IncludePrivate:        true,
				})
				
				if err != nil {
					t.Logf("Failed to parse %s: %v", path, err)
					return
				}
				
				// Compare with Python AST
				result, err := comparator.CompareFile(path, ourAST)
				if err != nil {
					t.Logf("Failed to compare %s: %v", path, err)
					return
				}
				
				// Log statistics
				if !result.Match {
					t.Logf("Mismatch in %s: missing %d items", path, len(result.MissingInOurs))
				}
			})
		}
		
		return nil
	})
	
	require.NoError(t, err)
}