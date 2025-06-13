package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/janreges/ai-distiller/internal/semantic"
)

func main() {
	fmt.Println("=== AI Distiller Semantic Resolver Test (Pass 2) ===")
	fmt.Println()

	// Setup test files
	testDir := "./test-data/semantic-test"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			"test-data/semantic-test",
			"../test-data/semantic-test", 
			"../../test-data/semantic-test",
		}
		
		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				testDir = alt
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("Test directory not found: %s\n", testDir)
			os.Exit(1)
		}
	}

	mainFile := filepath.Join(testDir, "main.py")
	utilsFile := filepath.Join(testDir, "utils.py")

	// Get absolute paths
	absMainFile, err := filepath.Abs(mainFile)
	if err != nil {
		fmt.Printf("Failed to get absolute path for main.py: %v\n", err)
		os.Exit(1)
	}

	absUtilsFile, err := filepath.Abs(utilsFile)
	if err != nil {
		fmt.Printf("Failed to get absolute path for utils.py: %v\n", err)
		os.Exit(1)
	}

	projectRoot, err := filepath.Abs(testDir)
	if err != nil {
		fmt.Printf("Failed to get project root: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project root: %s\n", projectRoot)
	fmt.Printf("Main file: %s\n", absMainFile)
	fmt.Printf("Utils file: %s\n", absUtilsFile)
	fmt.Println()

	// Create analyzer
	analyzer, err := semantic.NewAnalyzer(projectRoot)
	if err != nil {
		fmt.Printf("Failed to create analyzer: %v\n", err)
		os.Exit(1)
	}

	// Analyze main.py (Pass 1)
	fmt.Println("=== Pass 1: Analyzing main.py ===")
	mainAnalysis, err := analyzeFile(analyzer, absMainFile)
	if err != nil {
		fmt.Printf("Failed to analyze main.py: %v\n", err)
		os.Exit(1)
	}
	displayPass1Results("main.py", mainAnalysis)

	// Analyze utils.py (Pass 1)
	fmt.Println("\n=== Pass 1: Analyzing utils.py ===")
	utilsAnalysis, err := analyzeFile(analyzer, absUtilsFile)
	if err != nil {
		fmt.Printf("Failed to analyze utils.py: %v\n", err)
		os.Exit(1)
	}
	displayPass1Results("utils.py", utilsAnalysis)

	// Create resolver (Pass 2)
	fmt.Println("\n=== Pass 2: Resolving symbols ===")
	resolver := semantic.NewResolver(projectRoot)

	// Resolve the two files
	semanticGraph, err := resolver.ResolveFilePair(mainAnalysis, utilsAnalysis)
	if err != nil {
		fmt.Printf("Failed to resolve symbols: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayPass2Results(semanticGraph)
}

// analyzeFile performs Pass 1 analysis on a file
func analyzeFile(analyzer *semantic.Analyzer, filePath string) (*semantic.FileAnalysis, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ctx := context.Background()
	return analyzer.AnalyzeFile(ctx, file, filePath)
}

// displayPass1Results shows Pass 1 analysis results
func displayPass1Results(filename string, analysis *semantic.FileAnalysis) {
	fmt.Printf("File: %s\n", filename)
	fmt.Printf("Symbols: %d\n", len(analysis.SymbolTable.Symbols))
	fmt.Printf("Dependencies: %d\n", len(analysis.Dependencies))
	fmt.Printf("Call Sites: %d\n", len(analysis.CallSites))

	// Show some key symbols
	functions := analysis.SymbolTable.GetSymbolsOfKind(semantic.SymbolKindFunction)
	classes := analysis.SymbolTable.GetSymbolsOfKind(semantic.SymbolKindClass)
	
	if len(classes) > 0 {
		fmt.Printf("Classes: ")
		for i, class := range classes {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", class.Name)
		}
		fmt.Println()
	}

	if len(functions) > 0 {
		fmt.Printf("Functions: ")
		for i, function := range functions {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", function.Name)
		}
		fmt.Println()
	}

	// Show dependencies
	if len(analysis.Dependencies) > 0 {
		fmt.Printf("Imports: ")
		for i, dep := range analysis.Dependencies {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", dep.TargetModule)
		}
		fmt.Println()
	}
}

// displayPass2Results shows Pass 2 resolution results
func displayPass2Results(semanticGraph *semantic.SemanticGraph) {
	fmt.Printf("=== Semantic Graph Results ===\n")
	fmt.Printf("Total Files: %d\n", len(semanticGraph.FileSymbolTables))
	fmt.Printf("Total Symbols: %d\n", semanticGraph.Statistics.TotalSymbols)
	fmt.Printf("Total Call Sites: %d\n", len(semanticGraph.CallSites))
	fmt.Printf("Resolved Calls: %d\n", semanticGraph.Statistics.ResolvedCalls)
	fmt.Printf("Unresolved Calls: %d\n", semanticGraph.Statistics.UnresolvedCalls)
	fmt.Println()

	// Show dependency graph
	fmt.Printf("=== File Dependencies ===\n")
	for file, deps := range semanticGraph.DependencyGraph {
		filename := filepath.Base(file)
		fmt.Printf("%s depends on:\n", filename)
		for _, dep := range deps {
			depName := filepath.Base(dep)
			fmt.Printf("  -> %s\n", depName)
		}
	}
	fmt.Println()

	// Show resolved call graph
	fmt.Printf("=== Call Graph (Resolved) ===\n")
	for callerID, calleeIDs := range semanticGraph.CallGraph {
		callerName := extractSymbolName(string(callerID))
		fmt.Printf("%s calls:\n", callerName)
		for _, calleeID := range calleeIDs {
			calleeName := extractSymbolName(string(calleeID))
			fmt.Printf("  -> %s\n", calleeName)
		}
	}

	// Show unresolved calls
	fmt.Printf("\n=== Unresolved Calls ===\n")
	unresolvedCount := 0
	for _, callSite := range semanticGraph.CallSites {
		if !callSite.IsResolved {
			unresolvedCount++
			if unresolvedCount <= 10 { // Limit output
				callerName := extractSymbolName(string(callSite.CallerID))
				fmt.Printf("%s -> %s (line %d)\n", 
					callerName, callSite.CalleeName, callSite.Location.StartLine)
			}
		}
	}
	if unresolvedCount > 10 {
		fmt.Printf("... and %d more unresolved calls\n", unresolvedCount-10)
	}
}

// extractSymbolName extracts the symbol name from a SymbolID
func extractSymbolName(symbolID string) string {
	parts := strings.Split(symbolID, "::")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (symbol name)
	}
	return symbolID
}