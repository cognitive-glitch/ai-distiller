package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/janreges/ai-distiller/poc/cgo/internal/parser"
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
		fmt.Println("CGo Tree-sitter PoC")
		fmt.Println("\nUsage:")
		fmt.Println("  poc-cgo [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Printf("\nStartup time: %v\n", time.Since(startTime))
		return
	}

	if *versionFlag {
		fmt.Println("poc-cgo version 0.1.0")
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

	// Create parser
	p, err := parser.NewPythonParser()
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

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
	fmt.Printf("Parse time: %v\n", parseTime)
	fmt.Printf("Tree: %s\n", tree.RootNode().String())
	fmt.Printf("Node count: %d\n", countNodes(tree.RootNode()))
	fmt.Printf("\nTotal startup + parse time: %v\n", time.Since(startTime))

	// Check if we meet performance requirements
	if time.Since(startTime) < 50*time.Millisecond {
		fmt.Println("✓ Meets <50ms startup requirement")
	} else {
		fmt.Println("✗ Exceeds 50ms startup requirement")
	}
}

func countNodes(node *parser.Node) int {
	count := 1
	for i := 0; i < int(node.ChildCount()); i++ {
		count += countNodes(node.Child(i))
	}
	return count
}