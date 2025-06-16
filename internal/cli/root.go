package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	
	"github.com/spf13/cobra"
	"github.com/janreges/ai-distiller/internal/debug"
	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/language"
	_ "github.com/janreges/ai-distiller/internal/language" // Register language processors
)

var (
	// Version is set by main.go
	Version string

	// Flags
	outputFile       string
	outputToStdout   bool
	outputFormat     string
	stripOptions     []string // Deprecated, kept for backward compatibility
	includeGlob      string
	excludeGlob      string
	recursiveStr     string
	filePathType     string
	relativePathPrefix string
	strict           bool
	verbosity        int
	useTreeSitter    bool
	langOverride     string
	
	// New filtering flags
	includePublic         *bool
	includeProtected      *bool
	includeInternal       *bool
	includePrivate        *bool
	includeComments       *bool
	includeDocstrings     *bool
	includeImplementation *bool
	includeImports        *bool
	includeAnnotations    *bool
	
	// Group flags
	includeList           string
	excludeList           string
	
	// Concurrency flags
	workers               int
	
	// Raw mode flag
	rawMode               bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aid [path]",
	Short: "AI Distiller - Extract essential code structure for LLMs",
	Long: `AI Distiller (aid) intelligently "distills" source code from any project 
into a compact, structured format, optimized for the context window of 
Large Language Models (LLMs).

═══════════════════════════════════════════════════════════════════════════════

BASIC OPTIONS:
  -o, --output <file>          Output file (default: auto-generated)
  --stdout                     Print to stdout instead of file
  --format <type>              Output format: text|md|jsonl|json-structured|xml
                              (default: text)

PATH & OUTPUT CONTROL:
  --file-path-type <type>      How paths appear in output: relative|absolute
                              (default: relative)
  --relative-path-prefix <str> Custom prefix for relative paths (e.g., "src/")
  -r, --recursive              Process directories recursively
                              0/1 (default: 1)

───────────────────────────────────────────────────────────────────────────────

VISIBILITY FILTERING:
  Control which visibility levels are included in output
  
  --public                     Include public members
                              0/1 (default: 1)
  --protected                  Include protected members
                              0/1 (default: 0)
  --internal                   Include internal/package-private members
                              0/1 (default: 0)
  --private                    Include private members
                              0/1 (default: 0)

CONTENT FILTERING:
  Control what code elements are included
  
  --comments                   Include comments
                              0/1 (default: 0)
  --docstrings                 Include documentation
                              0/1 (default: 1)
  --implementation             Include function/method bodies
                              0/1 (default: 0)
  --imports                    Include import statements
                              0/1 (default: 1)
  --annotations                Include decorators/annotations
                              0/1 (default: 1)

ALTERNATIVE FILTERING:
  --include-only <items>       Include ONLY these categories (comma-separated)
  --exclude-items <items>      Exclude these categories (comma-separated)
                              Categories: public,protected,internal,private,
                              comments,docstrings,implementation,imports,annotations

───────────────────────────────────────────────────────────────────────────────

FILE SELECTION:
  --include <pattern>          Include file patterns (e.g., "*.py")
  --exclude <pattern>          Exclude file patterns (e.g., "*test*")

PROCESSING MODE:
  --raw                        Process all text files without parsing
  --lang <language>            Override language detection
                              Languages: auto|python|typescript|javascript|go|ruby|
                              swift|rust|java|csharp|kotlin|cpp|php
                              (default: auto)
  --tree-sitter                Use tree-sitter parser (experimental)

───────────────────────────────────────────────────────────────────────────────

PERFORMANCE:
  -w, --workers <num>          Number of parallel workers
                              0=auto (80% CPU), 1=serial, N=use N workers
                              (default: 0)

DIAGNOSTICS:
  -v, --verbose                Verbose output (use -vv or -vvv for more)
  --strict                     Fail on first syntax error
  --version                    Show version information
  --help                       Show this help message

═══════════════════════════════════════════════════════════════════════════════

EXAMPLES:
  aid                          # Process current dir, public APIs only
  aid src/ --private=1         # Include private members
  aid --file-path-type=absolute # Use absolute paths in output
  aid docs/ --raw              # Process text files without parsing
  aid -w 1                     # Force serial processing
  aid --relative-path-prefix="module/" docs/  # Add custom prefix to paths`,
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
	rootCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: md|text|jsonl|json-structured|xml (default: text)")

	// Legacy processing flags (deprecated)
	rootCmd.Flags().StringSliceVar(&stripOptions, "strip", nil, "DEPRECATED: Use individual filtering flags instead")
	rootCmd.Flags().MarkDeprecated("strip", "use individual filtering flags like --public=1, --private=0, etc.")
	
	// File pattern flags
	rootCmd.Flags().StringVar(&includeGlob, "include", "", "Include file patterns (default: all supported)")
	rootCmd.Flags().StringVar(&excludeGlob, "exclude", "", "Exclude file patterns")
	rootCmd.Flags().StringVarP(&recursiveStr, "recursive", "r", "1", "Process directories recursively (0/1, default: 1)")
	rootCmd.Flags().StringVar(&filePathType, "file-path-type", "relative", "How paths appear in output: relative|absolute (default: relative)")
	rootCmd.Flags().StringVar(&relativePathPrefix, "relative-path-prefix", "", "Custom prefix for relative paths (e.g., \"src/\")")
	rootCmd.Flags().BoolVar(&strict, "strict", false, "Fail on first syntax error")

	// General flags
	rootCmd.Flags().CountVarP(&verbosity, "verbose", "v", "Verbose output (use -vv or -vvv for more detail)")
	rootCmd.Flags().Bool("version", false, "Show version information")
	rootCmd.Flags().Bool("help", false, "Show help message")
	
	// Experimental flags
	rootCmd.Flags().BoolVar(&useTreeSitter, "tree-sitter", false, "Use tree-sitter parser (experimental)")

	// Language override flag
	rootCmd.Flags().StringVar(&langOverride, "lang", "auto", "Override language detection: auto|python|typescript|javascript|go|ruby|swift|rust|java|csharp|kotlin|cpp|php")
	
	// New filtering flags - visibility
	rootCmd.Flags().String("public", "1", "Include public members (0/1, default: 1)")
	rootCmd.Flags().String("protected", "0", "Include protected members (0/1, default: 0)")
	rootCmd.Flags().String("internal", "0", "Include internal/package-private members (0/1, default: 0)")
	rootCmd.Flags().String("private", "0", "Include private members (0/1, default: 0)")
	
	// New filtering flags - content
	rootCmd.Flags().String("comments", "0", "Include comments (0/1, default: 0)")
	rootCmd.Flags().String("docstrings", "1", "Include documentation comments (0/1, default: 1)")
	rootCmd.Flags().String("implementation", "0", "Include function/method bodies (0/1, default: 0)")
	rootCmd.Flags().String("imports", "1", "Include import statements (0/1, default: 1)")
	rootCmd.Flags().String("annotations", "1", "Include decorators/annotations (0/1, default: 1)")
	
	// Group filtering flags
	rootCmd.Flags().StringVar(&includeList, "include-only", "", "Include only these categories (comma-separated)")
	rootCmd.Flags().StringVar(&excludeList, "exclude-items", "", "Exclude these categories (comma-separated)")
	
	// Concurrency flags
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 0, "Number of parallel workers (0=auto/80% CPU cores, 1=serial, default: 0)")
	
	// Raw mode flag
	rootCmd.Flags().BoolVar(&rawMode, "raw", false, "Raw mode: process all text files without parsing (txt, md, json, yaml, etc.)")

	// Handle version flag specially
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("aid version %s\n", Version)
			os.Exit(0)
		}
		
		// Parse boolean flags
		parseBoolFlag(cmd, "public", &includePublic)
		parseBoolFlag(cmd, "protected", &includeProtected)
		parseBoolFlag(cmd, "internal", &includeInternal)
		parseBoolFlag(cmd, "private", &includePrivate)
		parseBoolFlag(cmd, "comments", &includeComments)
		parseBoolFlag(cmd, "docstrings", &includeDocstrings)
		parseBoolFlag(cmd, "implementation", &includeImplementation)
		parseBoolFlag(cmd, "imports", &includeImports)
		parseBoolFlag(cmd, "annotations", &includeAnnotations)
		
		// Validate mutually exclusive flags
		if includeList != "" && excludeList != "" {
			fmt.Fprintf(os.Stderr, "Error: --include-only and --exclude-items are mutually exclusive\n")
			os.Exit(1)
		}
	}
}

func runDistiller(cmd *cobra.Command, args []string) error {
	// Create debugger based on verbosity level
	dbg := debug.New(os.Stderr, verbosity)
	ctx := debug.NewContext(context.Background(), dbg)
	
	// Log startup info
	dbg.Logf(debug.LevelBasic, "AI Distiller %s starting", Version)
	
	// Check if stdin is available (not a TTY) or explicitly requested with "-"
	stdinAvailable := false
	if len(args) > 0 && args[0] == "-" {
		stdinAvailable = true
	} else {
		// Check if stdin is piped
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinAvailable = true
		}
	}
	
	// Handle stdin input
	if stdinAvailable {
		return processStdinWithContext(ctx)
	}
	
	// Handle file/directory input
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

	// Generate output filename if not specified and not using stdout
	if outputFile == "" && !outputToStdout {
		outputFile = generateOutputFilename(absPath, stripOptions)
	}

	// Validate output format
	validFormats := []string{"md", "text", "jsonl", "json-structured", "xml"}
	if !contains(validFormats, outputFormat) {
		return fmt.Errorf("invalid output format: %s (valid: %s)", outputFormat, strings.Join(validFormats, ", "))
	}

	// Log configuration using debugger
	dbg.Logf(debug.LevelBasic, "Input: %s", absPath)
	dbg.Logf(debug.LevelBasic, "Output: %s", outputFile)
	dbg.Logf(debug.LevelBasic, "Format: %s", outputFormat)
	
	if len(stripOptions) > 0 {
		dbg.Logf(debug.LevelBasic, "Strip (deprecated): %s", strings.Join(stripOptions, ", "))
	}
	
	// Log detailed configuration at level 2
	dbg.Logf(debug.LevelDetailed, "Visibility: public=%v, protected=%v, internal=%v, private=%v",
		getBoolFlag(includePublic, true),
		getBoolFlag(includeProtected, false),
		getBoolFlag(includeInternal, false),
		getBoolFlag(includePrivate, false))
	dbg.Logf(debug.LevelDetailed, "Content: comments=%v, docstrings=%v, implementation=%v, imports=%v, annotations=%v",
		getBoolFlag(includeComments, false),
		getBoolFlag(includeDocstrings, true),
		getBoolFlag(includeImplementation, false),
		getBoolFlag(includeImports, true),
		getBoolFlag(includeAnnotations, true))

	// Create processor options from flags
	procOpts := createProcessOptionsFromFlags()

	// Create the processor with context
	proc := processor.NewWithContext(ctx)
	
	// Enable tree-sitter if requested
	if useTreeSitter {
		proc.EnableTreeSitter()
		dbg.Logf(debug.LevelBasic, "Using tree-sitter parser (experimental)")
	}

	// Log workers configuration
	actualWorkers := workers
	if actualWorkers == 0 {
		actualWorkers = int(float64(runtime.NumCPU()) * 0.8)
		if actualWorkers < 1 {
			actualWorkers = 1
		}
	}
	if workers != 1 {
		dbg.Logf(debug.LevelBasic, "Using %d parallel workers (%d CPU cores available)", actualWorkers, runtime.NumCPU())
	}

	// Set base path information
	procOpts.BasePath = inputPath
	procOpts.FilePathType = filePathType
	procOpts.RelativePathPrefix = relativePathPrefix
	
	// If user provided absolute path and didn't specify file-path-type, default to absolute
	if filepath.IsAbs(inputPath) && !cmd.Flags().Changed("file-path-type") {
		procOpts.FilePathType = "absolute"
	}
	
	// Process the input
	result, err := proc.ProcessPath(absPath, procOpts)
	if err != nil {
		return fmt.Errorf("failed to process: %w", err)
	}
	if result == nil {
		return fmt.Errorf("no result returned from processing")
	}

	// Create formatter based on format
	formatterOpts := formatter.Options{}
	formatter, err := formatter.Get(outputFormat, formatterOpts)
	if err != nil {
		return fmt.Errorf("failed to get formatter: %w", err)
	}

	// Write output
	var output strings.Builder
	
	// Log formatting phase
	dbg.Logf(debug.LevelDetailed, "Starting formatting phase with %s formatter", outputFormat)
	defer dbg.Timing(debug.LevelDetailed, fmt.Sprintf("formatting to %s", outputFormat))()
	
	// Handle different result types
	switch r := result.(type) {
	case *ir.DistilledFile:
		dbg.Logf(debug.LevelDetailed, "Formatting single file: %s", r.Path)
		
		// Dump IR being formatted at trace level
		debug.Lazy(ctx, debug.LevelTrace, func(d debug.Debugger) {
			d.Dump(debug.LevelTrace, "IR being formatted", r)
		})
		
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
		
		dbg.Logf(debug.LevelDetailed, "Formatting %d files from directory", len(files))
		
		if err := formatter.FormatMultiple(&output, files); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	default:
		return fmt.Errorf("unexpected result type: %T", result)
	}
	
	dbg.Logf(debug.LevelDetailed, "Formatted output size: %d bytes", output.Len())

	// Write to file if not stdout-only
	if outputFile != "" && !outputToStdout {
		if err := os.WriteFile(outputFile, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		dbg.Logf(debug.LevelBasic, "Wrote output to %s", outputFile)
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

	// Build options suffix based on what's excluded from defaults
	var abbrev []string
	
	// Check if using new flag system
	if len(stripOptions) == 0 {
		// New flag system - check what differs from defaults
		if !getBoolFlag(includePublic, true) {
			abbrev = append(abbrev, "npub")
		}
		if getBoolFlag(includeProtected, false) {
			abbrev = append(abbrev, "prot")
		}
		if getBoolFlag(includeInternal, false) {
			abbrev = append(abbrev, "int")
		}
		if getBoolFlag(includePrivate, false) {
			abbrev = append(abbrev, "priv")
		}
		if getBoolFlag(includeComments, false) {
			abbrev = append(abbrev, "com")
		}
		if !getBoolFlag(includeDocstrings, true) {
			abbrev = append(abbrev, "ndoc")
		}
		if getBoolFlag(includeImplementation, false) {
			abbrev = append(abbrev, "impl")
		}
		if !getBoolFlag(includeImports, true) {
			abbrev = append(abbrev, "nimp")
		}
		if !getBoolFlag(includeAnnotations, true) {
			abbrev = append(abbrev, "nann")
		}
	} else {
		// Legacy --strip system
		for _, opt := range stripOptions {
			switch opt {
			case "comments":
				abbrev = append(abbrev, "ncom")
			case "imports":
				abbrev = append(abbrev, "nimp")
			case "implementation":
				abbrev = append(abbrev, "nimpl")
			case "non-public":
				abbrev = append(abbrev, "npub")
			case "private":
				abbrev = append(abbrev, "npriv")
			case "protected":
				abbrev = append(abbrev, "nprot")
			}
		}
	}
	
	optionsSuffix := ""
	if len(abbrev) > 0 {
		optionsSuffix = "." + strings.Join(abbrev, ".")
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

// processStdinWithContext handles input from stdin with debugging context
func processStdinWithContext(ctx context.Context) error {
	dbg := debug.FromContext(ctx).WithSubsystem("stdin")
	
	// Force stdout output when reading from stdin
	outputToStdout = true
	
	dbg.Logf(debug.LevelBasic, "Processing input from stdin")
	
	// Read stdin into buffer for language detection
	var buffer bytes.Buffer
	tee := io.TeeReader(os.Stdin, &buffer)
	
	// Read up to 64KB for detection
	detectBuf := make([]byte, 64*1024)
	n, _ := tee.Read(detectBuf)
	detectBuf = detectBuf[:n]
	
	dbg.Logf(debug.LevelDetailed, "Read %d bytes for language detection", n)
	
	// Determine language
	lang := langOverride
	if lang == "auto" {
		detector := language.NewDetector()
		detectedLang, err := detector.DetectFromReader(bytes.NewReader(detectBuf))
		if err != nil {
			return fmt.Errorf("could not auto-detect language from stdin. Please specify with --lang flag")
		}
		lang = detectedLang
		dbg.Logf(debug.LevelBasic, "Detected language: %s", lang)
	} else {
		dbg.Logf(debug.LevelBasic, "Using specified language: %s", lang)
	}
	
	// Read the rest of stdin
	remainingBytes, _ := io.ReadAll(tee)
	fullContent := append(detectBuf, remainingBytes...)
	
	// Get language processor
	langProc, ok := language.GetProcessor(lang)
	if !ok {
		return fmt.Errorf("no processor found for language: %s", lang)
	}
	
	// Create processor options from flags
	procOpts := createProcessOptionsFromFlags()
	
	// Process the input with our debug-enabled context
	result, err := langProc.ProcessWithOptions(ctx, bytes.NewReader(fullContent), "stdin", procOpts)
	if err != nil {
		return fmt.Errorf("failed to process stdin: %w", err)
	}
	
	// Create formatter based on format
	formatterOpts := formatter.Options{}
	formatter, err := formatter.Get(outputFormat, formatterOpts)
	if err != nil {
		return fmt.Errorf("failed to get formatter: %w", err)
	}
	
	// Format and output
	var output strings.Builder
	if err := formatter.Format(&output, result); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	
	// Always write to stdout for stdin input
	fmt.Print(output.String())
	
	return nil
}

// parseBoolFlag parses a string flag as boolean (0/1)
func parseBoolFlag(cmd *cobra.Command, name string, target **bool) {
	if val, err := cmd.Flags().GetString(name); err == nil && val != "" {
		// fmt.Printf("DEBUG parseBoolFlag: %s = %s\n", name, val)
		b, err := strconv.ParseBool(val)
		if err != nil {
			// Try parsing as 0/1
			if val == "0" {
				b = false
			} else if val == "1" {
				b = true
			} else {
				fmt.Fprintf(os.Stderr, "Error: --%s must be 0 or 1, got %q\n", name, val)
				os.Exit(1)
			}
		}
		*target = &b
	}
}

// createProcessOptionsFromFlags creates ProcessOptions from the new flag system
func createProcessOptionsFromFlags() processor.ProcessOptions {
	opts := processor.ProcessOptions{}
	
	// Handle legacy --strip flag if present
	if len(stripOptions) > 0 {
		if verbosity > 0 {
			fmt.Fprintf(os.Stderr, "Warning: --strip is deprecated. Use individual filtering flags instead.\n")
		}
		// Apply legacy strip options
		opts.IncludeComments = !contains(stripOptions, "comments")
		opts.IncludeImports = !contains(stripOptions, "imports")
		opts.IncludeImplementation = !contains(stripOptions, "implementation")
		opts.IncludePrivate = !contains(stripOptions, "non-public")
		opts.RemovePrivateOnly = contains(stripOptions, "private")
		opts.RemoveProtectedOnly = contains(stripOptions, "protected")
		opts.Workers = workers
		opts.RawMode = rawMode
		return opts
	}
	
	// Process include/exclude lists if provided
	if includeList != "" {
		opts = processIncludeList(includeList)
		opts.Workers = workers
		opts.RawMode = rawMode
		return opts
	}
	if excludeList != "" {
		opts = processExcludeList(excludeList)
		opts.Workers = workers
		opts.RawMode = rawMode
		return opts
	}
	
	// Use individual flags with defaults
	// fmt.Printf("DEBUG: includeComments=%v\n", includeComments)
	opts.IncludeComments = getBoolFlag(includeComments, false)
	opts.IncludeImports = getBoolFlag(includeImports, true)
	opts.IncludeImplementation = getBoolFlag(includeImplementation, false)
	
	// Handle visibility flags
	includePublicVal := getBoolFlag(includePublic, true)
	includeProtectedVal := getBoolFlag(includeProtected, false)
	includeInternalVal := getBoolFlag(includeInternal, false)
	includePrivateVal := getBoolFlag(includePrivate, false)
	
	// Convert to stripper options
	// If only public is included, remove all non-public
	if includePublicVal && !includeProtectedVal && !includeInternalVal && !includePrivateVal {
		opts.IncludePrivate = false
	} else {
		opts.IncludePrivate = true
		// Set specific removal flags based on what's NOT included
		opts.RemovePrivateOnly = !includePrivateVal
		opts.RemoveProtectedOnly = !includeProtectedVal
		opts.RemoveInternalOnly = !includeInternalVal
	}
	
	// Handle docstrings and annotations
	opts.IncludeDocstrings = getBoolFlag(includeDocstrings, true)
	opts.IncludeAnnotations = getBoolFlag(includeAnnotations, true)
	
	// Set workers value
	opts.Workers = workers
	
	// Set raw mode
	opts.RawMode = rawMode
	
	// Set recursive - parse from global recursiveStr
	opts.Recursive = recursiveStr != "0"
	
	return opts
}

// getBoolFlag returns the value of a bool flag or its default
func getBoolFlag(flag *bool, defaultVal bool) bool {
	if flag == nil {
		return defaultVal
	}
	return *flag
}

// processIncludeList processes the --include-only flag
func processIncludeList(list string) processor.ProcessOptions {
	opts := processor.ProcessOptions{
		// Start with everything excluded
		IncludeComments: false,
		IncludeImports: false,
		IncludeImplementation: false,
		IncludePrivate: false,
	}
	
	items := strings.Split(list, ",")
	for _, item := range items {
		switch strings.TrimSpace(item) {
		case "public":
			// This is the default, nothing to change
		case "protected":
			opts.IncludePrivate = true
			opts.RemovePrivateOnly = true
		case "internal":
			opts.IncludePrivate = true
			opts.RemoveProtectedOnly = true
		case "private":
			opts.IncludePrivate = true
		case "comments":
			opts.IncludeComments = true
		case "docstrings":
			// TODO: Implement separate docstring handling
			opts.IncludeComments = true
		case "implementation":
			opts.IncludeImplementation = true
		case "imports":
			opts.IncludeImports = true
		case "annotations":
			// TODO: Implement annotation handling
		}
	}
	
	return opts
}

// processExcludeList processes the --exclude-items flag
func processExcludeList(list string) processor.ProcessOptions {
	opts := processor.ProcessOptions{
		// Start with everything included
		IncludeComments: true,
		IncludeImports: true,
		IncludeImplementation: true,
		IncludePrivate: true,
	}
	
	items := strings.Split(list, ",")
	for _, item := range items {
		switch strings.TrimSpace(item) {
		case "private":
			opts.RemovePrivateOnly = true
		case "protected":
			opts.RemoveProtectedOnly = true
		case "internal":
			// Internal is often grouped with private in many languages
			opts.RemovePrivateOnly = true
		case "comments":
			opts.IncludeComments = false
		case "docstrings":
			// TODO: Implement separate docstring handling
			opts.IncludeComments = false
		case "implementation":
			opts.IncludeImplementation = false
		case "imports":
			opts.IncludeImports = false
		case "annotations":
			// TODO: Implement annotation handling
		}
	}
	
	return opts
}