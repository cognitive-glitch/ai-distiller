package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/language/java"
	"github.com/janreges/ai-distiller/internal/language/javascript"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/language/typescript"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Test configuration
type TestConfig struct {
	Language string
	File     string
	Parser   processor.LanguageProcessor
}

// Test result
type TestResult struct {
	Language string
	File     string
	Success  bool
	Error    error
	Stats    ParseStats
}

// Parse statistics
type ParseStats struct {
	Classes    int
	Functions  int
	Fields     int
	Interfaces int
	Imports    int
	Comments   int
	Enums      int
}

func main() {
	fmt.Println("=== AI Distiller Parser Functional Tests ===")
	fmt.Println()

	// Test configurations
	tests := []TestConfig{
		{
			Language: "Java",
			File:     "test_java_complex.java",
			Parser:   java.NewProcessor(),
		},
		{
			Language: "TypeScript",
			File:     "test_typescript_complex.ts",
			Parser:   typescript.NewProcessor(),
		},
		{
			Language: "Python",
			File:     "test_python_complex.py",
			Parser:   python.NewProcessor(),
		},
		{
			Language: "JavaScript",
			File:     "test_javascript_complex.js",
			Parser:   javascript.NewProcessor(),
		},
	}

	// Run tests
	results := make([]TestResult, 0, len(tests))
	for _, test := range tests {
		result := runTest(test)
		results = append(results, result)
		printTestResult(result)
		fmt.Println()
	}

	// Summary
	printSummary(results)

	// Test stripping modes
	fmt.Println("\n=== Testing Stripping Modes ===")
	testStrippingModes(tests)
}

func runTest(config TestConfig) TestResult {
	result := TestResult{
		Language: config.Language,
		File:     config.File,
	}

	// Open test file
	filePath := filepath.Join(".", config.File)
	file, err := os.Open(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file: %w", err)
		return result
	}
	defer file.Close()

	// Parse file
	ctx := context.Background()
	distilledFile, err := config.Parser.Process(ctx, file, filePath)
	if err != nil {
		result.Error = fmt.Errorf("parsing failed: %w", err)
		return result
	}

	// Calculate statistics
	result.Stats = calculateStats(distilledFile)
	result.Success = true

	return result
}

func calculateStats(file *ir.DistilledFile) ParseStats {
	stats := ParseStats{}
	
	countNodes(file.Children, &stats)
	
	return stats
}

func countNodes(nodes []ir.DistilledNode, stats *ParseStats) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *ir.DistilledClass:
			stats.Classes++
			countNodes(n.Children, stats)
		case *ir.DistilledFunction:
			stats.Functions++
		case *ir.DistilledField:
			stats.Fields++
		case *ir.DistilledInterface:
			stats.Interfaces++
			countNodes(n.Children, stats)
		case *ir.DistilledImport:
			stats.Imports++
		case *ir.DistilledComment:
			stats.Comments++
		case *ir.DistilledEnum:
			stats.Enums++
			countNodes(n.Children, stats)
		}
	}
}

func printTestResult(result TestResult) {
	fmt.Printf("Language: %s\n", result.Language)
	fmt.Printf("File: %s\n", result.File)
	
	if result.Success {
		fmt.Printf("Status: ✅ SUCCESS\n")
		fmt.Printf("Statistics:\n")
		fmt.Printf("  Classes: %d\n", result.Stats.Classes)
		fmt.Printf("  Functions: %d\n", result.Stats.Functions)
		fmt.Printf("  Fields: %d\n", result.Stats.Fields)
		fmt.Printf("  Interfaces: %d\n", result.Stats.Interfaces)
		fmt.Printf("  Imports: %d\n", result.Stats.Imports)
		fmt.Printf("  Comments: %d\n", result.Stats.Comments)
		fmt.Printf("  Enums: %d\n", result.Stats.Enums)
	} else {
		fmt.Printf("Status: ❌ FAILED\n")
		fmt.Printf("Error: %v\n", result.Error)
	}
}

func printSummary(results []TestResult) {
	fmt.Println("=== Test Summary ===")
	successful := 0
	for _, result := range results {
		if result.Success {
			successful++
		}
	}
	
	fmt.Printf("Total tests: %d\n", len(results))
	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n", len(results)-successful)
	
	if successful == len(results) {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Println("⚠️  Some tests failed")
	}
}

func testStrippingModes(tests []TestConfig) {
	strippingTests := []struct {
		Name string
		Opts processor.ProcessOptions
	}{
		{
			Name: "Strip Implementation",
			Opts: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        true,
				IncludeComments:       true,
				IncludeImports:        true,
			},
		},
		{
			Name: "Strip Private",
			Opts: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        false,
				IncludeComments:       true,
				IncludeImports:        true,
			},
		},
		{
			Name: "Strip Comments",
			Opts: processor.ProcessOptions{
				IncludeImplementation: true,
				IncludePrivate:        true,
				IncludeComments:       false,
				IncludeImports:        true,
			},
		},
		{
			Name: "Strip All Non-Essential",
			Opts: processor.ProcessOptions{
				IncludeImplementation: false,
				IncludePrivate:        false,
				IncludeComments:       false,
				IncludeImports:        true,
			},
		},
	}

	for _, stripTest := range strippingTests {
		fmt.Printf("\n--- %s ---\n", stripTest.Name)
		
		for _, test := range tests {
			result := runStrippingTest(test, stripTest.Opts)
			fmt.Printf("%s: ", test.Language)
			if result.Success {
				fmt.Printf("✅ Classes: %d, Functions: %d, Fields: %d\n", 
					result.Stats.Classes, result.Stats.Functions, result.Stats.Fields)
			} else {
				fmt.Printf("❌ Error: %v\n", result.Error)
			}
		}
	}
}

func runStrippingTest(config TestConfig, opts processor.ProcessOptions) TestResult {
	result := TestResult{
		Language: config.Language,
		File:     config.File,
	}

	// Open test file
	filePath := filepath.Join(".", config.File)
	file, err := os.Open(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file: %w", err)
		return result
	}
	defer file.Close()

	// Check if parser supports ProcessWithOptions
	if processorWithOpts, ok := config.Parser.(interface {
		ProcessWithOptions(context.Context, io.Reader, string, processor.ProcessOptions) (*ir.DistilledFile, error)
	}); ok {
		// Parse file with options
		ctx := context.Background()
		distilledFile, err := processorWithOpts.ProcessWithOptions(ctx, file, filePath, opts)
		if err != nil {
			result.Error = fmt.Errorf("parsing with options failed: %w", err)
			return result
		}

		// Calculate statistics
		result.Stats = calculateStats(distilledFile)
		result.Success = true
	} else {
		result.Error = fmt.Errorf("parser does not support ProcessWithOptions")
	}

	return result
}

// validateExpectedStructures verifies that expected constructs are found
func validateExpectedStructures(language string, stats ParseStats) []string {
	var issues []string

	switch language {
	case "Java":
		if stats.Classes < 3 {
			issues = append(issues, "Expected at least 3 classes (ComplexJavaClass, DataProcessor interface, Status enum)")
		}
		if stats.Functions < 10 {
			issues = append(issues, "Expected at least 10 methods/functions")
		}
		if stats.Imports < 3 {
			issues = append(issues, "Expected at least 3 import statements")
		}
		if stats.Enums < 1 {
			issues = append(issues, "Expected at least 1 enum")
		}
		if stats.Interfaces < 1 {
			issues = append(issues, "Expected at least 1 interface")
		}

	case "TypeScript":
		if stats.Classes < 2 {
			issues = append(issues, "Expected at least 2 classes")
		}
		if stats.Functions < 8 {
			issues = append(issues, "Expected at least 8 functions/methods")
		}
		if stats.Interfaces < 3 {
			issues = append(issues, "Expected at least 3 interfaces")
		}
		if stats.Imports < 2 {
			issues = append(issues, "Expected at least 2 import statements")
		}

	case "Python":
		if stats.Classes < 5 {
			issues = append(issues, "Expected at least 5 classes")
		}
		if stats.Functions < 15 {
			issues = append(issues, "Expected at least 15 functions/methods")
		}
		if stats.Imports < 5 {
			issues = append(issues, "Expected at least 5 import statements")
		}

	case "JavaScript":
		if stats.Classes < 3 {
			issues = append(issues, "Expected at least 3 classes")
		}
		if stats.Functions < 10 {
			issues = append(issues, "Expected at least 10 functions/methods")
		}
		if stats.Imports < 2 {
			issues = append(issues, "Expected at least 2 import statements")
		}
	}

	return issues
}