package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	_ "github.com/janreges/ai-distiller/internal/language" // Register language processors
)

var (
	// Version is set by main.go
	Version string

	// Flags
	outputFile     string
	outputToStdout bool
	outputFormat   string
	stripOptions   []string
	includeGlob    string
	excludeGlob    string
	recursive      bool
	absolutePaths  bool
	strict         bool
	verbosity      int
	useTreeSitter  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aid [path]",
	Short: "AI Distiller - A smart source code summarizer for LLMs",
	Long: `AI Distiller (aid) intelligently "distills" source code from any project 
into a compact, structured format, optimized for the context window of 
Large Language Models (LLMs).

By extracting the essential structure, APIs, and relationships from source code,
AI Distiller creates compact, semantic "blueprints" that enable LLMs to reason 
effectively about complex software projects.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDistiller,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	initFlags()
}

func initFlags() {
	// Output flags
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: .<dir>.[options].aid.txt)")
	rootCmd.Flags().BoolVar(&outputToStdout, "stdout", false, "Print to stdout (in addition to file)")
	rootCmd.Flags().StringVar(&outputFormat, "format", "md", "Output format: md|text|jsonl|json-structured|xml")

	// Processing flags
	rootCmd.Flags().StringSliceVar(&stripOptions, "strip", nil, "Remove items: comments,imports,implementation,non-public")
	rootCmd.Flags().StringVar(&includeGlob, "include", "", "Include file patterns (default: all supported)")
	rootCmd.Flags().StringVar(&excludeGlob, "exclude", "", "Exclude file patterns")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "Process directories recursively")
	rootCmd.Flags().BoolVar(&absolutePaths, "absolute-paths", false, "Use absolute paths in output")
	rootCmd.Flags().BoolVar(&strict, "strict", false, "Fail on first syntax error")

	// General flags
	rootCmd.Flags().CountVarP(&verbosity, "verbose", "v", "Verbose output (use -vv or -vvv for more detail)")
	rootCmd.Flags().Bool("version", false, "Show version information")
	rootCmd.Flags().Bool("help", false, "Show help message")
	
	// Experimental flags
	rootCmd.Flags().BoolVar(&useTreeSitter, "tree-sitter", false, "Use tree-sitter parser (experimental)")

	// Handle version flag specially
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("aid version %s\n", Version)
			os.Exit(0)
		}
	}
}

func runDistiller(cmd *cobra.Command, args []string) error {
	// Get input path
	inputPath := "."
	if len(args) > 0 {
		inputPath = args[0]
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("path does not exist: %s", inputPath)
	}

	// Generate output filename if not specified
	if outputFile == "" {
		outputFile = generateOutputFilename(absPath, stripOptions)
	}

	// Validate output format
	validFormats := []string{"md", "text", "jsonl", "json-structured", "xml"}
	if !contains(validFormats, outputFormat) {
		return fmt.Errorf("invalid output format: %s (valid: %s)", outputFormat, strings.Join(validFormats, ", "))
	}

	// Log configuration if verbose
	if verbosity > 0 {
		fmt.Fprintf(os.Stderr, "AI Distiller %s\n", Version)
		fmt.Fprintf(os.Stderr, "Input: %s\n", absPath)
		fmt.Fprintf(os.Stderr, "Output: %s\n", outputFile)
		fmt.Fprintf(os.Stderr, "Format: %s\n", outputFormat)
		if len(stripOptions) > 0 {
			fmt.Fprintf(os.Stderr, "Strip: %s\n", strings.Join(stripOptions, ", "))
		}
	}

	// Create processor options from flags
	procOpts := processor.ProcessOptions{
		IncludeComments:       !contains(stripOptions, "comments"),
		IncludeImports:        !contains(stripOptions, "imports"),
		IncludeImplementation: !contains(stripOptions, "implementation"),
		IncludePrivate:        !contains(stripOptions, "non-public"),
	}

	// Create the processor
	proc := processor.New()
	
	// Enable tree-sitter if requested
	if useTreeSitter {
		proc.EnableTreeSitter()
		if verbosity > 0 {
			fmt.Fprintf(os.Stderr, "Using tree-sitter parser (experimental)\n")
		}
	}

	// Process the input
	result, err := proc.ProcessPath(absPath, procOpts)
	if err != nil {
		return fmt.Errorf("failed to process: %w", err)
	}

	// Create formatter based on format
	formatterOpts := formatter.Options{}
	formatter, err := formatter.Get(outputFormat, formatterOpts)
	if err != nil {
		return fmt.Errorf("failed to get formatter: %w", err)
	}

	// Write output
	var output strings.Builder
	
	// Handle different result types
	switch r := result.(type) {
	case *ir.DistilledFile:
		if err := formatter.Format(&output, r); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case *ir.DistilledDirectory:
		// Extract files from directory
		var files []*ir.DistilledFile
		for _, child := range r.Children {
			if file, ok := child.(*ir.DistilledFile); ok {
				files = append(files, file)
			}
		}
		if err := formatter.FormatMultiple(&output, files); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	default:
		return fmt.Errorf("unexpected result type: %T", result)
	}

	// Write to file if not stdout-only
	if outputFile != "" && !outputToStdout {
		if err := os.WriteFile(outputFile, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if verbosity > 0 {
			fmt.Fprintf(os.Stderr, "Wrote output to %s\n", outputFile)
		}
	}

	// Write to stdout if requested
	if outputToStdout || outputFile == "" {
		fmt.Print(output.String())
	}

	return nil
}

func generateOutputFilename(path string, stripOptions []string) string {
	// Get directory name
	dirName := filepath.Base(path)
	if dirName == "." || dirName == "/" {
		dirName = "current"
	}

	// Build options suffix
	optionsSuffix := ""
	if len(stripOptions) > 0 {
		// Create abbreviated options
		abbrev := make([]string, 0, len(stripOptions))
		for _, opt := range stripOptions {
			switch opt {
			case "comments":
				abbrev = append(abbrev, "ncom")
			case "imports":
				abbrev = append(abbrev, "nimp")
			case "implementation":
				abbrev = append(abbrev, "nimpl")
			case "non-public":
				abbrev = append(abbrev, "npriv")
			}
		}
		if len(abbrev) > 0 {
			optionsSuffix = "." + strings.Join(abbrev, ".")
		}
	}

	return fmt.Sprintf(".%s%s.aid.txt", dirName, optionsSuffix)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}