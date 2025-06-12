// +build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/language/python"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
)

type TestScenario struct {
	Name        string
	InputFile   string
	Options     processor.ProcessOptions
	StripperOpts stripper.Options
	Formatters  []string
	Description string
}

func main() {
	// Define test scenarios
	scenarios := []TestScenario{
		{
			Name:      "full_output",
			InputFile: "basic_class.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			},
			StripperOpts: stripper.Options{}, // No stripping
			Formatters:   []string{"markdown", "json", "jsonl", "xml"},
			Description:  "Full output with all information preserved",
		},
		{
			Name:      "no_private",
			InputFile: "basic_class.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        false,
			},
			StripperOpts: stripper.Options{
				RemovePrivate: true,
			},
			Formatters:  []string{"markdown", "json"},
			Description: "Remove private members",
		},
		{
			Name:      "no_implementation",
			InputFile: "basic_class.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: false,
				IncludeImports:        true,
				IncludePrivate:        true,
			},
			StripperOpts: stripper.Options{
				RemoveImplementations: true,
			},
			Formatters:  []string{"markdown", "json"},
			Description: "Remove implementations",
		},
		{
			Name:      "minimal",
			InputFile: "basic_class.py",
			Options: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImplementation: false,
				IncludeImports:        false,
				IncludePrivate:        false,
			},
			StripperOpts: stripper.Options{
				RemovePrivate:         true,
				RemoveImplementations: true,
				RemoveComments:        true,
				RemoveImports:         true,
			},
			Formatters:  []string{"markdown", "json"},
			Description: "Minimal output - structure only",
		},
		{
			Name:      "complex_imports",
			InputFile: "complex_imports.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			},
			StripperOpts: stripper.Options{},
			Formatters:   []string{"markdown", "json"},
			Description:  "Test complex import patterns",
		},
		{
			Name:      "decorators",
			InputFile: "decorators_and_metadata.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: true,
				IncludeImports:        true,
				IncludePrivate:        true,
			},
			StripperOpts: stripper.Options{},
			Formatters:   []string{"markdown", "json"},
			Description:  "Test decorators and metadata",
		},
		{
			Name:      "inheritance",
			InputFile: "inheritance_patterns.py",
			Options: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImplementation: false,
				IncludeImports:        true,
				IncludePrivate:        false,
			},
			StripperOpts: stripper.Options{
				RemovePrivate:         true,
				RemoveImplementations: true,
			},
			Formatters:  []string{"markdown", "json"},
			Description: "Test inheritance patterns",
		},
		{
			Name:      "edge_cases",
			InputFile: "edge_cases.py",
			Options: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImplementation: false,
				IncludeImports:        true,
				IncludePrivate:        false,
			},
			StripperOpts: stripper.Options{
				RemovePrivate:         true,
				RemoveImplementations: true,
				RemoveComments:        true,
			},
			Formatters:  []string{"markdown", "json"},
			Description: "Test edge cases and special constructs",
		},
	}

	// Run tests
	for _, scenario := range scenarios {
		log.Printf("Running scenario: %s - %s", scenario.Name, scenario.Description)
		
		if err := runScenario(scenario); err != nil {
			log.Printf("Error in scenario %s: %v", scenario.Name, err)
		}
	}
	
	// Generate report
	generateReport()
}

func runScenario(scenario TestScenario) error {
	// Read input file
	inputPath := filepath.Join("input", scenario.InputFile)
	
	// Process with Python processor
	processor := python.NewProcessor()
	distilled, err := processor.ProcessFile(inputPath, scenario.Options)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}
	
	// Apply stripper if needed
	if scenario.StripperOpts.RemovePrivate || 
	   scenario.StripperOpts.RemoveImplementations ||
	   scenario.StripperOpts.RemoveComments ||
	   scenario.StripperOpts.RemoveImports {
		stripperVisitor := stripper.New(scenario.StripperOpts)
		distilled = distilled.Accept(stripperVisitor).(*ir.DistilledFile)
	}
	
	// Generate outputs in different formats
	for _, formatName := range scenario.Formatters {
		outputPath := filepath.Join("actual", 
			fmt.Sprintf("%s_%s.%s", scenario.Name, scenario.InputFile, formatName))
		
		if err := generateOutput(distilled, formatName, outputPath); err != nil {
			log.Printf("Failed to generate %s output: %v", formatName, err)
		}
	}
	
	return nil
}

func generateOutput(file *ir.DistilledFile, formatName, outputPath string) error {
	// Get formatter
	fmtr, err := formatter.Get(formatName, formatter.Options{
		IncludeLocation: true,
		IncludeMetadata: true,
		Compact:         false,
	})
	if err != nil {
		return err
	}
	
	// Create output file
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()
	
	// Generate output
	return fmtr.Format(output, file)
}

func generateReport() {
	report := &bytes.Buffer{}
	
	fmt.Fprintln(report, "# AI Distiller Test Report")
	fmt.Fprintln(report)
	fmt.Fprintln(report, "## Test Scenarios")
	fmt.Fprintln(report)
	
	// List all generated files
	actualFiles, _ := filepath.Glob("actual/*")
	
	fmt.Fprintln(report, "### Generated Files")
	fmt.Fprintln(report)
	for _, file := range actualFiles {
		fmt.Fprintf(report, "- %s\n", filepath.Base(file))
	}
	
	fmt.Fprintln(report)
	fmt.Fprintln(report, "## Quality Checks")
	fmt.Fprintln(report)
	fmt.Fprintln(report, "### Things to verify:")
	fmt.Fprintln(report, "1. **Structure Preservation**: Classes, functions, and their relationships are correctly captured")
	fmt.Fprintln(report, "2. **Filtering Accuracy**: Private members, implementations, etc. are correctly filtered")
	fmt.Fprintln(report, "3. **Import Handling**: All import types are correctly parsed")
	fmt.Fprintln(report, "4. **Metadata**: Decorators, type hints, and other metadata are preserved")
	fmt.Fprintln(report, "5. **Edge Cases**: Unicode, special methods, async functions are handled")
	fmt.Fprintln(report, "6. **Format Consistency**: Each format represents the same information accurately")
	
	// Save report
	os.WriteFile("test_report.md", report.Bytes(), 0644)
}

// Helper to compare expected vs actual (for future use)
func compareFiles(expected, actual string) error {
	expectedData, err := os.ReadFile(expected)
	if err != nil {
		return err
	}
	
	actualData, err := os.ReadFile(actual)
	if err != nil {
		return err
	}
	
	if !bytes.Equal(expectedData, actualData) {
		return fmt.Errorf("files differ")
	}
	
	return nil
}