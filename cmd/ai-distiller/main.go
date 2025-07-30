package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	_ "github.com/janreges/ai-distiller/internal/language" // Register all language processors
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/stripper"
	"github.com/spf13/cobra"
)

var (
	// Command flags
	outputFormat          string
	outputFile            string
	includePrivate        bool
	includeImplementation bool
	includeComments       bool
	includeImports        bool
	includeLocation       bool
	includeMetadata       bool
	compactOutput         bool
	verbose               bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ai-distiller [files...]",
		Short: "Distill source code into compact representations for LLMs",
		Long: `AI Distiller processes source code files and extracts their essential structure,
making them suitable for use with Large Language Models by removing unnecessary details
while preserving important semantic information.`,
		Example: `  # Distill a single Python file to markdown
  ai-distiller example.py

  # Distill multiple files to JSON
  ai-distiller --format json-structured src/*.py

  # Strip private members and implementations
  ai-distiller --no-private --no-implementation module.py

  # Output to a file
  ai-distiller --output distilled.md src/**/*.py`,
		Args: cobra.MinimumNArgs(1),
		RunE: runDistill,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "markdown", 
		"Output format: markdown, json, jsonl, xml")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "",
		"Output file (default: stdout)")
	rootCmd.Flags().BoolVar(&includePrivate, "private", false,
		"Include private members")
	rootCmd.Flags().BoolVar(&includeImplementation, "implementation", true,
		"Include function/method implementations")
	rootCmd.Flags().BoolVar(&includeComments, "comments", true,
		"Include comments and docstrings")
	rootCmd.Flags().BoolVar(&includeImports, "imports", true,
		"Include import statements")
	rootCmd.Flags().BoolVar(&includeLocation, "location", false,
		"Include source location information")
	rootCmd.Flags().BoolVar(&includeMetadata, "metadata", false,
		"Include file metadata")
	rootCmd.Flags().BoolVar(&compactOutput, "compact", false,
		"Produce compact output")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose output")

	// Add subcommands
	rootCmd.AddCommand(createListCommand())
	rootCmd.AddCommand(createVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDistill(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no input files specified")
	}

	// Expand glob patterns
	var inputFiles []string
	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		if len(matches) == 0 {
			// If no matches, treat as literal filename
			inputFiles = append(inputFiles, pattern)
		} else {
			inputFiles = append(inputFiles, matches...)
		}
	}

	// Process files
	var distilledFiles []*ir.DistilledFile
	
	for _, file := range inputFiles {
		if verbose {
			fmt.Fprintf(os.Stderr, "Processing %s...\n", file)
		}

		// Detect language and get processor
		proc, err := getProcessorForFile(file)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", file, err)
			}
			continue
		}

		// Process file
		processOpts := processor.ProcessOptions{
			IncludeComments:       includeComments,
			IncludeImplementation: includeImplementation,
			IncludeImports:        includeImports,
			IncludePrivate:        includePrivate,
			SymbolResolution:      true,
		}

		// Open file for processing
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", file, err)
		}
		defer f.Close()
		
		distilled, err := proc.ProcessWithOptions(cmd.Context(), f, file, processOpts)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", file, err)
		}

		// Apply stripper if needed
		if !includePrivate || !includeImplementation || !includeComments || !includeImports {
			stripOpts := stripper.Options{
				RemovePrivate:          !includePrivate,
				RemoveImplementations:  !includeImplementation,
				RemoveComments:         !includeComments,
				RemoveImports:          !includeImports,
			}
			stripperVisitor := stripper.New(stripOpts)
			distilled = distilled.Accept(stripperVisitor).(*ir.DistilledFile)
		}

		distilledFiles = append(distilledFiles, distilled)
	}

	if len(distilledFiles) == 0 {
		return fmt.Errorf("no files were successfully processed")
	}

	// Get formatter
	formatOpts := formatter.Options{
		IncludeLocation: includeLocation,
		IncludeMetadata: includeMetadata,
		Compact:         compactOutput,
		AbsolutePaths:   false,
		SortNodes:       false,
	}

	fmtr, err := formatter.Get(outputFormat, formatOpts)
	if err != nil {
		return fmt.Errorf("invalid format %q: %w", outputFormat, err)
	}

	// Determine output destination
	var output *os.File
	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// Format output
	if len(distilledFiles) == 1 {
		err = fmtr.Format(output, distilledFiles[0])
	} else {
		err = fmtr.FormatMultiple(output, distilledFiles)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if verbose && outputFile != "" {
		fmt.Fprintf(os.Stderr, "Output written to %s\n", outputFile)
	}

	return nil
}

func getProcessorForFile(filename string) (processor.LanguageProcessor, error) {
	proc, ok := processor.GetByFilename(filename)
	if !ok {
		ext := filepath.Ext(filename)
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
	return proc, nil
}

func createListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available output formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			formats := formatter.List()
			fmt.Println("Available output formats:")
			for _, format := range formats {
				fmt.Printf("  - %s\n", format)
			}
			return nil
		},
	}
}

func createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("AI Distiller v0.1.0")
		},
	}
}