package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/janreges/ai-distiller/internal/semantic"
)

func main() {
	fmt.Println("=== AI Distiller Semantic Analysis Test ===")
	fmt.Println()

	// Get test file path
	testFile := "./test-data/functional-tests/test_python_complex.py"
	if len(os.Args) > 1 {
		testFile = os.Args[1]
	}

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			"test-data/functional-tests/test_python_complex.py",
			"../test-data/functional-tests/test_python_complex.py",
			"../../test-data/functional-tests/test_python_complex.py",
		}
		
		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				testFile = alt
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("Test file not found: %s\n", testFile)
			fmt.Println("Usage: go run ./cmd/semantic-test [python_file_path]")
			os.Exit(1)
		}
	}

	// Get absolute path
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		fmt.Printf("Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Analyzing file: %s\n", absPath)
	fmt.Println()

	// Create analyzer
	analyzer, err := semantic.NewAnalyzer(filepath.Dir(absPath))
	if err != nil {
		fmt.Printf("Failed to create analyzer: %v\n", err)
		os.Exit(1)
	}

	// Open file
	file, err := os.Open(absPath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Analyze file
	ctx := context.Background()
	analysis, err := analyzer.AnalyzeFile(ctx, file, absPath)
	if err != nil {
		fmt.Printf("Failed to analyze file: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayAnalysisResults(analysis)

	// Optionally save to JSON
	if len(os.Args) > 2 && os.Args[2] == "--save" {
		outputFile := strings.TrimSuffix(absPath, filepath.Ext(absPath)) + "_semantic.json"
		saveAnalysisToJSON(analysis, outputFile)
		fmt.Printf("\nAnalysis saved to: %s\n", outputFile)
	}
}

// displayAnalysisResults shows the analysis results in a formatted way
func displayAnalysisResults(analysis *semantic.FileAnalysis) {
	fmt.Printf("=== Analysis Results ===\n")
	fmt.Printf("File: %s\n", analysis.FilePath)
	fmt.Printf("Language: %s\n", analysis.Language)
	fmt.Printf("Analysis Time: %v\n", analysis.AnalysisTime)
	fmt.Println()

	// Symbol table statistics
	symbolTable := analysis.SymbolTable
	fmt.Printf("=== Symbol Table ===\n")
	fmt.Printf("Total Symbols: %d\n", len(symbolTable.Symbols))
	fmt.Printf("Dependencies: %d\n", len(symbolTable.Dependencies))
	fmt.Println()

	// Group symbols by kind
	symbolsByKind := make(map[semantic.SymbolKind][]*semantic.Symbol)
	for _, symbol := range symbolTable.Symbols {
		symbolsByKind[symbol.Kind] = append(symbolsByKind[symbol.Kind], symbol)
	}

	// Display symbols by category
	categories := []semantic.SymbolKind{
		semantic.SymbolKindClass,
		semantic.SymbolKindFunction,
		semantic.SymbolKindMethod,
		semantic.SymbolKindProperty,
		semantic.SymbolKindVariable,
		semantic.SymbolKindConstant,
	}

	for _, kind := range categories {
		if symbols, exists := symbolsByKind[kind]; exists && len(symbols) > 0 {
			fmt.Printf("=== %s (%d) ===\n", strings.Title(string(kind))+"s", len(symbols))
			for _, symbol := range symbols {
				displaySymbol(symbol)
			}
			fmt.Println()
		}
	}

	// Dependencies
	if len(analysis.Dependencies) > 0 {
		fmt.Printf("=== Dependencies (%d) ===\n", len(analysis.Dependencies))
		for _, dep := range analysis.Dependencies {
			fmt.Printf("  %s: %s", dep.ImportType, dep.TargetModule)
			if len(dep.ImportedSymbols) > 0 && dep.ImportedSymbols[0] != "" {
				fmt.Printf(" [%s]", strings.Join(dep.ImportedSymbols, ", "))
			}
			if dep.ImportAlias != "" {
				fmt.Printf(" as %s", dep.ImportAlias)
			}
			if dep.IsRelative {
				fmt.Printf(" (relative)")
			}
			fmt.Printf(" at line %d\n", dep.Location.StartLine)
		}
		fmt.Println()
	}

	// Call sites
	if len(analysis.CallSites) > 0 {
		fmt.Printf("=== Call Sites (%d) ===\n", len(analysis.CallSites))
		for i, call := range analysis.CallSites {
			if i >= 10 { // Limit output for readability
				fmt.Printf("  ... and %d more calls\n", len(analysis.CallSites)-10)
				break
			}
			fmt.Printf("  %s calls %s at line %d\n", 
				call.CallerID, call.CalleeName, call.Location.StartLine)
		}
		fmt.Println()
	}
}

// displaySymbol shows detailed information about a symbol
func displaySymbol(symbol *semantic.Symbol) {
	fmt.Printf("  %s", symbol.Name)
	
	if symbol.Scope != "" {
		fmt.Printf(" (%s)", symbol.Scope)
	}
	
	if symbol.Signature != "" {
		fmt.Printf(" - %s", symbol.Signature)
	}
	
	fmt.Printf(" [%s]", symbol.Visibility)
	
	if symbol.IsStatic {
		fmt.Printf(" [static]")
	}
	
	if symbol.IsAbstract {
		fmt.Printf(" [abstract]")
	}
	
	if len(symbol.Metadata.Decorators) > 0 {
		fmt.Printf(" @%s", strings.Join(symbol.Metadata.Decorators, " @"))
	}
	
	if len(symbol.Metadata.Extends) > 0 {
		fmt.Printf(" extends %s", strings.Join(symbol.Metadata.Extends, ", "))
	}
	
	fmt.Printf(" at line %d", symbol.Location.StartLine)
	
	if symbol.Metadata.LineCount > 0 {
		fmt.Printf(" (%d lines)", symbol.Metadata.LineCount)
	}
	
	fmt.Println()
}

// saveAnalysisToJSON saves the analysis results to a JSON file
func saveAnalysisToJSON(analysis *semantic.FileAnalysis, outputFile string) {
	// Create a serializable version (without AST and Content)
	serializable := struct {
		FilePath     string                    `json:"file_path"`
		Language     string                    `json:"language"`
		SymbolTable  *semantic.SymbolTable     `json:"symbol_table"`
		Dependencies []semantic.DependencyInfo `json:"dependencies"`
		CallSites    []semantic.CallSite       `json:"call_sites"`
		AnalysisTime string                    `json:"analysis_time"`
	}{
		FilePath:     analysis.FilePath,
		Language:     analysis.Language,
		SymbolTable:  analysis.SymbolTable,
		Dependencies: analysis.Dependencies,
		CallSites:    analysis.CallSites,
		AnalysisTime: analysis.AnalysisTime.String(),
	}

	data, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		fmt.Printf("Failed to write JSON file: %v\n", err)
		return
	}
}