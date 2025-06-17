package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/janreges/ai-distiller/poc/wasm/internal/parser"
)

func main() {
	// Start time for measuring startup
	startTime := time.Now()

	// Parse flags
	var (
		helpFlag    = flag.Bool("help", false, "Show help message")
		versionFlag = flag.Bool("version", false, "Show version")
		inputFile   = flag.String("file", "", "Python file to parse")
	)
	flag.Parse()

	// Handle help and version
	if *helpFlag {
		fmt.Println("WASM Tree-sitter PoC")
		fmt.Println("\nUsage:")
		fmt.Println("  poc-wasm [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Printf("\nStartup time: %v\n", time.Since(startTime))
		return
	}

	if *versionFlag {
		fmt.Println("poc-wasm version 0.1.0")
		return
	}

	// If no file specified, use test data
	if *inputFile == "" {
		*inputFile = "../testdata/simple.py"
	}

	// Read file
	source, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Create parser (includes WASM initialization)
	initStart := time.Now()
	p, err := parser.NewWASMPythonParser()
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()
	initTime := time.Since(initStart)

	// Parse file
	parseStart := time.Now()
	tree, err := p.Parse(source)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}
	parseTime := time.Since(parseStart)

	// Print results
	fmt.Printf("File: %s\n", *inputFile)
	fmt.Printf("Size: %d bytes\n", len(source))
	fmt.Printf("WASM init time: %v\n", initTime)
	fmt.Printf("Parse time: %v\n", parseTime)
	fmt.Printf("Tree root type: %s\n", tree.RootType)
	fmt.Printf("Node count: %d\n", tree.NodeCount)
	fmt.Printf("Has errors: %v\n", tree.HasErrors)
	fmt.Printf("\nTotal startup + parse time: %v\n", time.Since(startTime))

	// Check if we meet performance requirements
	if time.Since(startTime) < 50*time.Millisecond {
		fmt.Println("✓ Meets <50ms startup requirement")
	} else {
		fmt.Println("✗ Exceeds 50ms startup requirement")
	}

	// Check parse performance vs CGo baseline
	// Assuming CGo baseline is ~1ms for simple.py
	baselineParseTime := 1 * time.Millisecond
	if parseTime <= baselineParseTime*2 {
		fmt.Printf("✓ Parse performance within 2x of baseline (%v vs %v)\n", parseTime, baselineParseTime)
	} else {
		fmt.Printf("✗ Parse performance exceeds 2x baseline (%v vs %v)\n", parseTime, baselineParseTime)
	}
}