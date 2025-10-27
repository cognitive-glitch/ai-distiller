package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/language/csharp"
	"github.com/janreges/ai-distiller/internal/language/java"
	"github.com/janreges/ai-distiller/internal/language/javascript"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/language/rust"
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

	// Get test data directory
	testDataDir := filepath.Join(".", "test-data", "functional-tests")
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		testDataDir = filepath.Join("..", "..", "test-data", "functional-tests")
	}

	// Test configurations
	tests := []TestConfig{
		{
			Language: "Java",
			File:     filepath.Join(testDataDir, "test_java_complex.java"),
			Parser:   java.NewProcessor(),
		},
		{
			Language: "TypeScript",
			File:     filepath.Join(testDataDir, "test_typescript_complex.ts"),
			Parser:   typescript.NewProcessor(),
		},
		{
			Language: "Python",
			File:     filepath.Join(testDataDir, "test_python_complex.py"),
			Parser:   python.NewProcessor(),
		},
		{
			Language: "JavaScript",
			File:     filepath.Join(testDataDir, "test_javascript_complex.js"),
			Parser:   javascript.NewProcessor(),
		},
		{
			Language: "C#",
			File:     filepath.Join(testDataDir, "test_csharp_complex.cs"),
			Parser:   csharp.NewProcessor(),
		},
		{
			Language: "Rust",
			File:     filepath.Join(testDataDir, "test_rust_complex.rs"),
			Parser:   rust.NewProcessor(),
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
		File:     filepath.Base(config.File),
	}

	// Open test file
	file, err := os.Open(config.File)
	if err != nil {
		result.Error = fmt.Errorf("failed to open file: %w", err)
		return result
	}
	defer file.Close()

	// Parse file
	ctx := context.Background()
	distilledFile, err := config.Parser.Process(ctx, file, config.File)
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
		File:     filepath.Base(config.File),
	}

	// Open test file
	file, err := os.Open(config.File)
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
		distilledFile, err := processorWithOpts.ProcessWithOptions(ctx, file, config.File, opts)
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