// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
	"github.com/google/go-cmp/cmp"
)

// ComprehensiveTestCase represents a test scenario with expected outcomes
type ComprehensiveTestCase struct {
	Name          string
	InputFile     string
	Description   string
	ExpectedNodes []ExpectedNode
	ExpectedErrors []string // Expected parsing errors
}

// ExpectedNode describes what we expect to find in the output
type ExpectedNode struct {
	Type       string // "class", "function", "import"
	Name       string
	Properties map[string]interface{} // Additional properties to check
}

var comprehensiveTests = []ComprehensiveTestCase{
	{
		Name:        "multiline_definitions",
		InputFile:   "input/multiline_definitions.py",
		Description: "Tests multiline function and class definitions",
		ExpectedNodes: []ExpectedNode{
			{
				Type: "function",
				Name: "my_complex_function",
				Properties: map[string]interface{}{
					"param_count": 3,
					"has_return":  true,
					"return_type": "bool",
				},
			},
			{
				Type: "class",
				Name: "MyDerivedClass",
				Properties: map[string]interface{}{
					"extends_count": 2,
				},
			},
			{
				Type: "function",
				Name: "function_with_comments",
				Properties: map[string]interface{}{
					"param_count": 2,
				},
			},
			{
				Type: "function",
				Name: "function_with_nested_types",
				Properties: map[string]interface{}{
					"has_complex_types": true,
				},
			},
			{
				Type: "class",
				Name: "ComplexClass",
				Properties: map[string]interface{}{
					"has_metaclass": true,
				},
			},
			{
				Type: "function",
				Name: "async_function",
				Properties: map[string]interface{}{
					"is_async":       true,
					"has_decorators": true,
				},
			},
		},
	},
	{
		Name:        "multiline_imports",
		InputFile:   "input/multiline_imports.py",
		Description: "Tests various import styles including multiline",
		ExpectedNodes: []ExpectedNode{
			{
				Type: "import",
				Name: "os",
			},
			{
				Type: "import",
				Name: "numpy",
				Properties: map[string]interface{}{
					"has_alias": true,
					"alias":     "np",
				},
			},
			{
				Type: "import",
				Name: "typing",
				Properties: map[string]interface{}{
					"symbols_count": 10, // At least 10 symbols imported
				},
			},
			{
				Type: "import",
				Name: "collections",
				Properties: map[string]interface{}{
					"has_aliases": true,
				},
			},
			{
				Type: "class",
				Name: "ImportantClass",
				Properties: map[string]interface{}{
					"description": "Should be parsed as class, not import",
				},
			},
		},
	},
	{
		Name:        "error_recovery",
		InputFile:   "input/error_recovery.py",
		Description: "Tests parser's ability to recover from syntax errors",
		ExpectedNodes: []ExpectedNode{
			{
				Type: "import",
				Name: "os",
			},
			{
				Type: "function",
				Name: "valid_function",
			},
			{
				Type: "class",
				Name: "ValidClass",
			},
			{
				Type: "import",
				Name: "sys",
			},
			{
				Type: "function",
				Name: "final_function",
			},
			{
				Type: "class",
				Name: "FinalClass",
			},
		},
		ExpectedErrors: []string{
			"MissingColon",    // Class without colon
			"broken_function", // Missing closing parenthesis
			"invalid import",  // Broken import statement
		},
	},
	{
		Name:        "nested_structures",
		InputFile:   "input/nested_structures.py",
		Description: "Tests nested classes and functions with proper hierarchy",
		ExpectedNodes: []ExpectedNode{
			{
				Type: "class",
				Name: "TopLevelClass",
				Properties: map[string]interface{}{
					"has_nested_class": true,
					"method_count":     6, // Including properties and class methods
				},
			},
			{
				Type: "class",
				Name: "NestedClass",
				Properties: map[string]interface{}{
					"is_nested":        true,
					"parent_class":     "TopLevelClass",
					"has_nested_class": true,
				},
			},
			{
				Type: "class",
				Name: "DoublyNestedClass",
				Properties: map[string]interface{}{
					"nesting_level": 2,
				},
			},
			{
				Type: "function",
				Name: "top_level_function",
			},
			{
				Type: "class",
				Name: "AnotherTopLevel",
			},
		},
	},
}

func TestComprehensivePythonParsing(t *testing.T) {
	pythonProc := python.NewProcessor()

	for _, tc := range comprehensiveTests {
		t.Run(tc.Name, func(t *testing.T) {
			// Process the file
			file, err := pythonProc.ProcessFile(tc.InputFile, processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			})
			
			if err != nil && len(tc.ExpectedErrors) == 0 {
				t.Fatalf("Unexpected error processing %s: %v", tc.InputFile, err)
			}

			// Check expected nodes
			for _, expected := range tc.ExpectedNodes {
				found := findNode(file, expected.Type, expected.Name)
				if found == nil {
					t.Errorf("Expected %s '%s' not found", expected.Type, expected.Name)
					continue
				}

				// Check additional properties
				for prop, expectedValue := range expected.Properties {
					actualValue := getNodeProperty(found, prop)
					if !compareValues(expectedValue, actualValue) {
						t.Errorf("Property %s of %s '%s': expected %v, got %v",
							prop, expected.Type, expected.Name, expectedValue, actualValue)
					}
				}
			}

			// If we expect errors, check that they were handled gracefully
			if len(tc.ExpectedErrors) > 0 {
				// For now, just ensure the parser didn't crash
				// In the future, check file.Errors when implemented
				t.Logf("Parser handled %d expected error cases", len(tc.ExpectedErrors))
			}
		})
	}
}

// TestGoldenFiles tests against known good outputs
func TestGoldenFiles(t *testing.T) {
	goldenDir := "golden"
	inputDir := "input"

	// Skip if golden directory doesn't exist
	if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
		t.Skip("Golden files directory not found")
	}

	pythonProc := python.NewProcessor()

	// Process each input file and compare with golden
	inputFiles, err := filepath.Glob(filepath.Join(inputDir, "*.py"))
	if err != nil {
		t.Fatal(err)
	}

	for _, inputFile := range inputFiles {
		baseName := filepath.Base(inputFile)
		goldenFile := filepath.Join(goldenDir, strings.TrimSuffix(baseName, ".py")+".json")

		t.Run(baseName, func(t *testing.T) {
			// Skip if no golden file exists
			if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
				t.Skip("No golden file for " + baseName)
			}

			// Process input
			file, err := pythonProc.ProcessFile(inputFile, processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Convert to JSON
			actual, err := json.MarshalIndent(file, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			// Read golden file
			expected, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatal(err)
			}

			// Compare
			if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
				t.Errorf("Output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// Update golden files with -update flag
var updateGolden = flag.Bool("update", false, "update golden files")

func TestUpdateGoldenFiles(t *testing.T) {
	if !*updateGolden {
		t.Skip("Use -update flag to update golden files")
	}

	// Implementation would update golden files
	// This is a placeholder for the actual implementation
}

// Helper functions

func findNode(file *ir.DistilledFile, nodeType, name string) ir.DistilledNode {
	var result ir.DistilledNode
	ir.Walk(file, func(n ir.DistilledNode) bool {
		switch nodeType {
		case "class":
			if class, ok := n.(*ir.DistilledClass); ok && class.Name == name {
				result = n
				return false
			}
		case "function":
			if fn, ok := n.(*ir.DistilledFunction); ok && fn.Name == name {
				result = n
				return false
			}
		case "import":
			if imp, ok := n.(*ir.DistilledImport); ok && imp.Module == name {
				result = n
				return false
			}
		}
		return true
	})
	return result
}

func getNodeProperty(node ir.DistilledNode, property string) interface{} {
	// Implementation would extract specific properties from nodes
	// This is simplified for the example
	switch property {
	case "param_count":
		if fn, ok := node.(*ir.DistilledFunction); ok {
			return len(fn.Parameters)
		}
	case "extends_count":
		if class, ok := node.(*ir.DistilledClass); ok {
			return len(class.Extends)
		}
	case "is_async":
		if fn, ok := node.(*ir.DistilledFunction); ok {
			for _, mod := range fn.Modifiers {
				if mod == ir.ModifierAsync {
					return true
				}
			}
		}
	}
	return nil
}

func compareValues(expected, actual interface{}) bool {
	// Simple comparison, could be enhanced
	return fmt.Sprintf("%v", expected) == fmt.Sprintf("%v", actual)
}