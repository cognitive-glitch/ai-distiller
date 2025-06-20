package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/janreges/ai-distiller/internal/ai"
	"github.com/janreges/ai-distiller/internal/aiactions"
	"github.com/janreges/ai-distiller/internal/version"
)

// getVersionInfo returns formatted version string with build date
func getVersionInfo() string {
	versionInfo := fmt.Sprintf("v%s", Version)
	if Version == "" {
		versionInfo = "v" + version.Version
	}
	if version.Date != "unknown" && version.Date != "" {
		if t, err := time.Parse(time.RFC3339, version.Date); err == nil {
			versionInfo += fmt.Sprintf(" built %s", t.Format("2006-01-02"))
		}
	}
	return versionInfo
}

// getHelpTemplate returns the help template with version info
func getHelpTemplate() string {
	versionInfo := getVersionInfo()
	
	return fmt.Sprintf(`{{.Short}} (%s)

AI Distiller transforms source code into optimized formats for Large Language Models.
Compress codebases by 60-90%% while preserving all semantic information needed for AI analysis.
Generate complete AI prompts and workflows - copy the output directly to Gemini 2.5 Pro,
ChatGPT-o3/4o, or Claude for perfect AI-powered code analysis.

USAGE:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

PATH:
  <path>                      Relative or absolute path to source directory or file

QUICK START:
  # Most common usage patterns
  aid ./my-project/src                                       # Basic distillation (public APIs only, no implementation/comments)
  aid ./src --ai-action prompt-for-refactoring-suggestion    # For refactoring
  aid ./src --ai-action prompt-for-security-analysis         # Security analysis
  aid .git --git-limit=50                                    # Git history mode
  aid .git --with-analysis-prompt                            # Git history with AI analysis (quality, patterns, visualizations)

AI ACTIONS:
  Use --ai-action <ACTION> to format output for specific AI tasks:

  prompt-for-refactoring-suggestion     Generate comprehensive refactoring analysis prompt
  prompt-for-complex-codebase-analysis  Generate enterprise-grade codebase overview prompt  
  prompt-for-security-analysis          Generate security audit prompt with OWASP focus
  prompt-for-performance-analysis       Generate performance optimization analysis prompt
  prompt-for-best-practices-analysis    Generate best practices and code quality analysis prompt
  prompt-for-bug-hunting               Generate systematic bug hunting and quality analysis prompt
  prompt-for-single-file-docs          Generate comprehensive single file documentation prompt
  prompt-for-diagrams                  Generate 10 beneficial Mermaid diagrams for architecture and processes
  flow-for-deep-file-to-file-analysis   Generate structured task list for comprehensive analysis
  flow-for-multi-file-docs             Generate structured documentation workflow for multiple files

CORE OPTIONS:
  -o, --output FILE           Output file (default: .aid/ folder or .aid.*.txt)
      --ai-action ACTION      AI analysis action (see list above)
      --ai-output FILE        Output path for AI action (default: action-specific directory/file)
      --format FORMAT         Output format: text|md|jsonl|json-structured|xml (default: text)
      --stdout                Print to stdout (in addition to file output)
  -w, --workers NUM           Parallel workers (0=auto, 1=serial, default: 0)
      --file-path-type TYPE   Path format: relative|absolute (default: relative)

SUMMARY OUTPUT:
  --summary-type TYPE         Summary format: visual-progress-bar|stock-ticker|speedometer-dashboard|
                              minimalist-sparkline|ci-friendly|json|off (default: visual-progress-bar)
      --no-emoji              Disable emojis in summary output

FILTERING (Essential):
  --public/--private/--protected/--internal     Visibility control (0/1, default: public=1)
  --comments/--docstrings/--implementation      Content control (0/1)
  --include/--exclude PATTERNS                  File patterns (e.g., "*.go,*.py")

SPECIAL MODES:
  --raw                       Process text files without parsing (overrides all content filters)
  --lang LANGUAGE             Force language: auto|python|typescript|javascript|go|rust|
                              java|csharp|kotlin|cpp|php|ruby|swift (useful for stdin input)
  aid .git                    Git history analysis mode (shows commit history)
  --with-analysis-prompt      Add comprehensive AI prompt for commit quality analysis, patterns,
                              development timeline visualization, and complexity insights

HELP & DOCUMENTATION:
  aid --help-extended         Complete documentation (man page style)
  aid help actions            Detailed AI actions documentation
  aid help filtering          Complete filtering reference  
  aid help git                Git mode documentation
  aid --cheat                 Quick reference card
{{if .HasAvailableSubCommands}}
AVAILABLE COMMANDS:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

OUTPUT FILE NAMING:
  Default outputs use .aid/ folder or .aid.*.txt files for easy recognition in git status.
  Use --stdout for direct output, but note: large codebases can generate MBs of text
  that exceed AI tool context limits. Some actions (like flow-for-deep-file-to-file-analysis)
  create multiple markdown files and directory structures.

For complete documentation and examples: aid --help-extended

---
AI Distiller (aid) - https://github.com/janreges/ai-distiller
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)
Explore the project on GitHub: https://github.com/janreges/ai-distiller
`, versionInfo)
}

// Extended help content for --help-extended
const extendedHelpContent = `AI DISTILLER - COMPLETE REFERENCE

NAME
    aid - AI Distiller: Extract essential code structure for LLMs

SYNOPSIS
    aid [OPTIONS] <path>

DESCRIPTION
    AI Distiller transforms source code into optimized formats for Large Language Models.
    It analyzes codebases, applies configurable filtering, and generates either compressed
    code representations or complete AI analysis prompts. The output can be directly
    copied to AI tools like Gemini 2.5 Pro (1M context), ChatGPT-o3/4o, or Claude for perfect
    AI-powered code analysis, refactoring, security audits, and architectural reviews.
    
    Key capabilities:
    • Compress codebases by 60-90%% while preserving semantic information
    • Generate complete AI prompts with embedded code for specific analysis tasks
    • Support for 15+ programming languages with intelligent parsing
    • Flexible filtering by visibility, content type, and file patterns

CORE CONCEPTS

    Distillation: The process of extracting essential code structure while removing
    noise (comments, implementation details, private members) that aren't needed
    for AI analysis.

    AI Actions: Pre-configured output formats optimized for specific AI tasks like
    refactoring, security analysis, or performance optimization.

    Visibility Filtering: Control which code elements are included based on their
    access level (public, private, protected, internal).

AI ACTIONS (--ai-action)

    prompt-for-refactoring-suggestion
        Generates a comprehensive prompt for AI-powered refactoring analysis.
        Includes context awareness, effort sizing, validation steps, and risk
        assessment. Output optimized for models like GPT-4, Claude, or Gemini.
        
        Example: aid --ai-action prompt-for-refactoring-suggestion ./src

    prompt-for-complex-codebase-analysis  
        Creates enterprise-grade codebase analysis prompt with architecture
        diagrams, compliance sections, and evidence-based findings. Includes
        coverage gaps and limitation acknowledgments.
        
        Example: aid --ai-action prompt-for-complex-codebase-analysis ./

    prompt-for-security-analysis
        Generates security audit prompt with OWASP Top 10 focus, static vs 
        dynamic analysis boundaries, and evidence checklists. Includes SARIF
        output integration for CI/CD pipelines.
        
        Example: aid --ai-action prompt-for-security-analysis ./src

    prompt-for-performance-analysis
        Creates performance optimization analysis prompt with static analysis
        constraints, profiling guidance, and enterprise scalability considerations.
        Focuses on algorithmic complexity and resource utilization patterns.
        
        Example: aid --ai-action prompt-for-performance-analysis ./src

OPTIONS

Primary Options:
    <path>                      Path to source file or directory [required]
    -o, --output FILE           Write output to file (default: .aid/ folder or .aid.*.txt)
    --ai-action ACTION          Use predefined AI action configuration
    --ai-output FILE            Custom output path for AI action
    --format FORMAT             Output format (text|md|jsonl|json-structured|xml)

Visibility Filtering:
    --public 0|1               Include public members (default: 1)
    --protected 0|1            Include protected members (default: 0) 
    --internal 0|1             Include internal/package-private members (default: 0)
    --private 0|1              Include private members (default: 0)

Content Filtering:
    --comments 0|1             Include comments (default: 0)
    --docstrings 0|1           Include documentation comments (default: 1)
    --implementation 0|1       Include function/method bodies (default: 0)
    --imports 0|1              Include import/require statements (default: 1)
    --annotations 0|1          Include decorators/annotations (default: 1)

Alternative Filtering:
    --include-only CATEGORIES   Include ONLY these categories (comma-separated)
    --exclude-items CATEGORIES  Exclude these categories (comma-separated)
                               Categories: public,protected,internal,private,
                               comments,docstrings,implementation,imports,annotations

File Selection:
    --include PATTERNS         Include file patterns (e.g., "*.py,*.go")
    --exclude PATTERNS         Exclude file patterns (e.g., "*test*,*.json")

Processing Options:
    --raw                      Process all text files without parsing
    --lang LANGUAGE            Override language detection
    --tree-sitter              Use tree-sitter parser (experimental)
    -r, --recursive 0|1        Process directories recursively (default: 1)

Path Control:
    --file-path-type TYPE      Path format: relative|absolute (default: relative)
    --relative-path-prefix STR Custom prefix for relative paths

Git Mode (when path is .git):
    --git-limit NUM            Limit number of commits (default: 200, 0=all)
    --with-analysis-prompt     Prepend AI analysis prompt for comprehensive insights:
                              - Commit quality scoring (0-100%) with individual feedback
                              - Development timeline Gantt charts (10-25 items)
                              - Complexity analysis with Mermaid visualizations
                              - Bug pattern detection and problem area identification

Performance:
    -w, --workers NUM          Parallel workers (0=auto, 1=serial, default: 0)

Summary Output:
    --summary-type TYPE        Summary output format (default: visual-progress-bar)
                               visual-progress-bar: Progress bar with compression ratio (default)
                               stock-ticker: Stock market style (📊 AID 97.5% ▲)
                               speedometer-dashboard: Multi-line dashboard with metrics
                               minimalist-sparkline: Compact single line format
                               ci-friendly: Clean format for CI/CD pipelines
                               json: Machine-readable JSON
                               off: Disable summary output
    --no-emoji                 Disable emojis in summary output

Diagnostics:
    -v, --verbose              Verbose output (-vv, -vvv for more detail)
    --strict                   Fail on first syntax error
    --version                  Show version information

SUPPORTED LANGUAGES

    auto-detected: python, typescript, javascript, go, ruby, swift, rust, 
    java, csharp, kotlin, cpp, php

EXAMPLES

Basic Usage:
    aid ./src                            # Distill src directory, public APIs only
    aid main.py --implementation=1       # Include function bodies
    aid . --private=1 --protected=1      # Include all visibility levels

AI-Powered Analysis:
    aid --ai-action prompt-for-refactoring-suggestion ./src
    aid --ai-action prompt-for-security-analysis . --private=1
    aid --ai-action prompt-for-performance-analysis ./core

Output Control:
    aid ./src --format=md -o analysis.md
    aid ./src --stdout | pbcopy          # Copy to clipboard (macOS)
    aid ./src --format=jsonl > data.jsonl

Filtering Examples:
    aid --include "*.go,*.py" ./          # Only Go and Python files
    aid --exclude "*test*,*spec*" ./      # Exclude test files
    aid --include-only public,imports ./  # Only public APIs and imports
    aid --exclude-items comments,implementation ./

Git Analysis:
    aid .git --git-limit=100             # Last 100 commits
    aid .git --with-analysis-prompt      # With AI analysis guidance

Advanced:
    aid --lang=python ./mixed-repo       # Force Python parsing
    aid --raw ./docs                     # Process as plain text
    aid -w 1 ./large-project            # Single-threaded processing

FILES

    Output Files:
        Default pattern: .aid.<dirname>.[options].txt
        Example: .aid.myproject.pub.txt (public only)
                 .aid.myproject.priv.prot.impl.txt (private, protected, implementation)

    Configuration:
        No configuration file support - all options via command line

EXIT STATUS
    0    Success
    1    General error (file not found, parse error, etc.)
    2    Invalid arguments or conflicting options

SEE ALSO
    aid help actions     - Detailed AI actions documentation
    aid help filtering   - Complete filtering reference
    aid help git         - Git mode documentation
    aid --cheat          - Quick reference card

    Online documentation: https://github.com/janreges/ai-distiller

AUTHOR
    AI Distiller development team

COPYRIGHT
    Licensed under MIT License
`

// initializeHelpSystem sets up custom help templates and commands
func initializeHelpSystem() {
	// Set custom help template
	rootCmd.SetHelpTemplate(getHelpTemplate())
	
	// Add extended help functionality
	rootCmd.Flags().Bool("help-extended", false, "Show extended help documentation")
	rootCmd.Flags().Bool("cheat", false, "Show quick reference card")
	
	// Add help subcommands
	helpCmd := &cobra.Command{
		Use:   "help [topic]",
		Short: "Get detailed help on specific topics",
		Long:  "Get detailed help documentation for specific topics like AI actions, filtering, or git mode.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runHelpCommand,
	}
	
	rootCmd.AddCommand(helpCmd)
	
	// Handle special help flags in PreRun
	originalPreRun := rootCmd.PreRun
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check for extended help
		if extended, _ := cmd.Flags().GetBool("help-extended"); extended {
			showExtendedHelp()
			os.Exit(0)
		}
		
		// Check for cheat sheet
		if cheat, _ := cmd.Flags().GetBool("cheat"); cheat {
			showCheatSheet()
			os.Exit(0)
		}
		
		// Call original PreRun if it exists
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}
}

// runHelpCommand handles topical help commands
func runHelpCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	
	topic := args[0]
	switch topic {
	case "actions":
		showAIActionsHelp()
	case "filtering":
		showFilteringHelp()
	case "git":
		showGitHelp()
	default:
		return fmt.Errorf("unknown help topic: %s\nAvailable topics: actions, filtering, git", topic)
	}
	
	return nil
}

// showAIActionsHelp displays detailed AI actions documentation
func showAIActionsHelp() {
	output := `AI ACTIONS - DETAILED REFERENCE

AI Actions are pre-configured analysis workflows that format the distilled output
for specific AI/LLM tasks. Each action includes optimized prompts, filtering
settings, and output formats.

AVAILABLE ACTIONS:

prompt-for-refactoring-suggestion
    PURPOSE: Generate comprehensive refactoring analysis prompts
    
    WHAT IT INCLUDES:
    • Context awareness with business constraints inference
    • Effort sizing and validation steps for each recommendation  
    • Risk assessment with rollback plans
    • Evidence-based findings with file:line references
    • ROI-focused prioritization framework
    
    FILTERING: Includes implementation details for analysis
    OUTPUT: Markdown prompt optimized for ChatGPT-o3/4o, Claude, Gemini
    
    EXAMPLE:
        aid ./src --ai-action prompt-for-refactoring-suggestion
        # Creates ./.aid/REFACTORING-SUGGESTION.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./src --ai-action prompt-for-refactoring-suggestion --raw
        # Includes full source code bodies (for smaller codebases or large AI context)

prompt-for-complex-codebase-analysis
    PURPOSE: Generate enterprise-grade codebase overview prompts
    
    WHAT IT INCLUDES:
    • Project context inference and assumption tracking
    • Enterprise concerns (compliance, governance, scalability)
    • Evidence-based findings with confidence levels
    • Coverage gaps and static analysis limitations
    • Architectural recommendations with dependency tracking
    
    FILTERING: Focuses on public APIs and structure
    OUTPUT: Comprehensive markdown analysis prompt
    
    EXAMPLE:
        aid ./ --ai-action prompt-for-complex-codebase-analysis
        # Creates ./.aid/COMPLEX-CODEBASE-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./lib --ai-action prompt-for-complex-codebase-analysis --raw
        # Includes full source code bodies (for smaller codebases or large AI context)

prompt-for-security-analysis
    PURPOSE: Generate security audit prompts with OWASP focus
    
    WHAT IT INCLUDES:
    • Static vs dynamic analysis boundaries with clear limitations
    • OWASP Top 10 comprehensive coverage
    • Evidence checklists with confidence tagging
    • False-positive mitigation strategies
    • SARIF output integration for CI/CD
    
    FILTERING: Includes all code for comprehensive security review
    OUTPUT: Security-focused analysis prompt with validation frameworks
    
    EXAMPLE:
        aid ./src --ai-action prompt-for-security-analysis --private=1
        # Creates ./.aid/SECURITY-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./src --ai-action prompt-for-security-analysis --private=1 --raw
        # Includes full source code bodies (for comprehensive security analysis)

prompt-for-performance-analysis
    PURPOSE: Generate performance optimization analysis prompts
    
    WHAT IT INCLUDES:
    • Static analysis constraints (no runtime speculation)
    • Algorithmic complexity analysis with evidence requirements
    • Profiling guidance to bridge static→dynamic analysis
    • Enterprise scalability considerations
    • Performance anti-pattern detection
    
    FILTERING: Includes implementation for complexity analysis
    OUTPUT: Performance-focused prompt with testing guidance
    
    EXAMPLE:
        aid ./core --ai-action prompt-for-performance-analysis
        # Creates ./.aid/PERFORMANCE-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./core --ai-action prompt-for-performance-analysis --raw
        # Includes full source code bodies (for detailed performance analysis)

prompt-for-best-practices-analysis
    PURPOSE: Generate best practices and code quality analysis prompts
    
    WHAT IT INCLUDES:
    • Code organization and structure assessment
    • Industry best practices compliance evaluation
    • Quality metrics and maintainability analysis
    • Team collaboration and workflow assessment
    • Implementation roadmap with prioritized improvements
    
    FILTERING: Includes implementation details for thorough analysis
    OUTPUT: Best practices analysis prompt with actionable recommendations
    
    EXAMPLE:
        aid ./src --ai-action prompt-for-best-practices-analysis
        # Creates ./.aid/BEST-PRACTICES-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./src --ai-action prompt-for-best-practices-analysis --raw
        # Includes full source code bodies (for comprehensive best practices analysis)

prompt-for-bug-hunting
    PURPOSE: Generate systematic bug hunting and quality analysis prompts
    
    WHAT IT INCLUDES:
    • Logic errors and edge case detection
    • Memory and resource management issues
    • Concurrency and threading problems
    • Input validation and error handling gaps
    • Systematic bug categorization and prioritization
    
    FILTERING: Includes all code and implementation details
    OUTPUT: Bug hunting analysis prompt with systematic detection methodology
    
    EXAMPLE:
        aid ./src --ai-action prompt-for-bug-hunting --private=1
        # Creates ./.aid/BUG-HUNTING-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./src --ai-action prompt-for-bug-hunting --private=1 --raw
        # Includes full source code bodies (for comprehensive bug detection)

prompt-for-single-file-docs
    PURPOSE: Generate comprehensive single file documentation prompts
    
    WHAT IT INCLUDES:
    • Complete API reference documentation
    • Usage examples and integration patterns
    • Implementation details and design decisions
    • Testing and troubleshooting guidance
    • Developer-friendly documentation structure
    
    FILTERING: Focuses on public APIs with usage examples
    OUTPUT: Single file documentation prompt with complete coverage
    
    EXAMPLE:
        aid ./src/main.py --ai-action prompt-for-single-file-docs
        # Creates ./.aid/SINGLE-FILE-DOCS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./src/api.ts --ai-action prompt-for-single-file-docs --raw
        # Includes full source code bodies (for detailed file documentation)

prompt-for-diagrams
    PURPOSE: Generate 10 beneficial Mermaid diagrams for architecture and process visualization
    
    WHAT IT INCLUDES:
    • Analysis strategy for both source code and text/documentation content
    • Systematic diagram selection based on coverage, clarity, and uniqueness
    • GitHub-compatible Mermaid syntax for all diagram types
    • Comprehensive prompt for creating flowcharts, sequence diagrams, class diagrams, etc.
    • Support for architectural overviews, data flows, and process workflows
    
    FILTERING: Uses default filtering to focus on key architectural elements
    OUTPUT: Comprehensive Mermaid diagram generation prompt with 10 distinct diagram specifications
    
    EXAMPLE:
        aid ./src --ai-action prompt-for-diagrams
        # Creates ./.aid/DIAGRAMS-ANALYSIS.YYYY-MM-DD.HH-MM-SS.basename.md
        
        aid ./docs --ai-action prompt-for-diagrams --raw
        # Includes full content for comprehensive diagram analysis from documentation

flow-for-deep-file-to-file-analysis
    PURPOSE: Generate structured task list for comprehensive file-by-file analysis
    
    WHAT IT INCLUDES:
    • Creates .aid/ directory with analysis infrastructure
    • Generates ANALYSIS-TASK-LIST.md with sequential file-by-file tasks
    • Template files for comprehensive project analysis workflow
    • Ensures consistent analysis methodology across all files
    
    FILTERING: Uses default public API filtering for overview
    OUTPUT: Directory structure with task lists and analysis templates
    
    EXAMPLE:
        aid ./src --ai-action flow-for-deep-file-to-file-analysis
        # Creates ./.aid/ directory with ANALYSIS-TASK-LIST.md and templates
        
        aid ./src --ai-action flow-for-deep-file-to-file-analysis --raw
        # Full source code in workflow (for comprehensive file-by-file analysis)

flow-for-multi-file-docs
    PURPOSE: Generate structured documentation workflow for multiple files
    
    WHAT IT INCLUDES:
    • Creates comprehensive documentation task list
    • Generates documentation index and navigation
    • Template files for API reference and developer guides
    • Individual file documentation templates
    • Quality assurance and review workflows
    
    FILTERING: Uses default public API filtering for documentation
    OUTPUT: Complete documentation workflow with templates and guides
    
    EXAMPLE:
        aid ./src --ai-action flow-for-multi-file-docs
        # Creates ./.aid/ directory with DOCS-TASK-LIST.md and documentation templates
        
        aid ./src --ai-action flow-for-multi-file-docs --raw
        # Full source code in documentation workflow (for detailed docs)

--RAW MODE FOR AI ACTIONS:

    Adding --raw flag includes full source code bodies in AI prompts for comprehensive analysis.
    
    CONTEXT SIZE CONSIDERATIONS:
    • Large codebases: Analyze specific parts/folders that fit in AI context, or use default
      filtering (public APIs only, no implementation/comments) which may be insufficient 
      for some analysis types but fits in smaller contexts
    • Small codebases: Use --raw for full source code analysis
    • Recommended: Gemini 2.5 Pro with 1M context window for largest codebase capacity

CUSTOMIZING AI ACTION OUTPUT:

    --ai-output PATH        Custom output file path
                           Supports template variables:
                           %%YYYY-MM-DD%% - Current date
                           %%HH-MM-SS%%   - Current time  
                           %%%folder-basename%% - Directory name
    
    EXAMPLES:
        aid ./src --ai-action prompt-for-refactoring-suggestion \
            --ai-output "./docs/refactoring-%%YYYY-MM-DD%%.md"
            
        aid ./ --ai-action prompt-for-security-analysis \
            --ai-output "./security-audit.md"

COMBINING WITH FILTERING:

    AI actions can be combined with custom filtering for specialized analysis:
    
    # Security analysis with all visibility levels
    aid ./ --ai-action prompt-for-security-analysis --private=1 --protected=1
    
    # Performance analysis without comments
    aid ./ --ai-action prompt-for-performance-analysis --comments=0
    
    # Refactoring analysis for specific file types
    aid ./ --ai-action prompt-for-refactoring-suggestion --include "*.go,*.py"
    
    # Full source code analysis (large context required - use Gemini 2.5 Pro recommended)
    aid ./small-module --ai-action prompt-for-security-analysis --raw

WORKFLOW INTEGRATION:

    # CI/CD Security Pipeline
    aid --ai-action prompt-for-security-analysis . > security-prompt.md
    # Feed security-prompt.md + distilled code to LLM
    
    # Code Review Assistant
    aid --ai-action prompt-for-refactoring-suggestion ./changed-files > review.md
    # Use review.md as context for AI-powered code review

OUTPUT FILE NAMING & MANAGEMENT:

AI Distiller uses intelligent file naming to avoid conflicts and enable easy management:

DEFAULT FILE PATTERNS:
    Regular distillation: .aid.{basename}.{options}.txt
    AI actions: .aid/{ACTION-NAME}.{timestamp}.{basename}.md
    Flow actions: .aid/ directory with multiple files and subdirectories

RATIONALE:
    • .aid prefix ensures easy recognition in git status
    • Can be easily added to .gitignore if desired
    • Avoids requiring users to specify output files for every operation
    • Large codebases can generate MBs of output - files prevent context overflow

STDOUT CONSIDERATIONS:
    --stdout works for all operations but use carefully:
    • Small projects: safe for direct AI tool input
    • Large codebases: may generate MBs of text exceeding AI context limits
    • Flow actions: only shows summary, full output still goes to .aid/ directory

EXAMPLES:
    aid ./src                    → .aid.src.pub.txt
    aid ./src --implementation=1 → .aid.src.pub.impl.txt
    aid ./src --ai-action prompt-for-refactoring-suggestion → .aid/REFACTORING-SUGGESTION.2025-06-17.14-30-00.src.md
    aid ./src --ai-action flow-for-deep-file-to-file-analysis → .aid/ directory with task lists and templates

For more examples: aid --help-extended

---
AI Distiller (aid) - https://github.com/janreges/ai-distiller
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)
Explore the project on GitHub: https://github.com/janreges/ai-distiller
`
	fmt.Print(output)
}

// showFilteringHelp displays complete filtering documentation
func showFilteringHelp() {
	fmt.Print(`FILTERING - COMPLETE REFERENCE

AI Distiller provides flexible filtering to control exactly what code elements
are included in the distilled output. This allows you to focus on specific
aspects like public APIs, implementation details, or security-relevant code.

VISIBILITY FILTERING:

Controls which code elements are included based on their access level:

    --public 0|1           Include public members (default: 1)
                          • Public functions, classes, methods
                          • Exported symbols (Go)
                          • Non-prefixed symbols (Python)
    
    --protected 0|1        Include protected members (default: 0)
                          • Protected methods (Java, C#, C++)
                          • _single_underscore methods (Python convention)
    
    --internal 0|1         Include internal/package-private members (default: 0)
                          • Package-private (Java)
                          • lowercase symbols (Go)
                          • module-level privates
    
    --private 0|1          Include private members (default: 0)
                          • Private methods/fields
                          • __double_underscore methods (Python)
                          • #private fields (JavaScript)

CONTENT FILTERING:

Controls what parts of the code structure are included:

    --comments 0|1         Include comments (default: 0)
                          • Single-line comments (// # ;)
                          • Block comments (/* */ """ )
                          • Inline comments
    
    --docstrings 0|1       Include documentation comments (default: 1)
                          • JSDoc comments
                          • Python docstrings
                          • Go package comments
                          • XML documentation (C#)
    
    --implementation 0|1   Include function/method bodies (default: 0)
                          • Function implementations
                          • Method bodies
                          • Constructor implementations
                          • Property getters/setters
    
    --imports 0|1          Include import statements (default: 1)
                          • import/require statements
                          • using directives
                          • #include directives
    
    --annotations 0|1      Include decorators/annotations (default: 1)
                          • Python decorators (@property)
                          • Java annotations (@Override)
                          • C# attributes ([Serializable])

ALTERNATIVE FILTERING SYNTAX:

    --include-only CATEGORIES    Include ONLY specified categories
    --exclude-items CATEGORIES   Exclude specified categories
    
    Categories: public, protected, internal, private, comments, docstrings,
               implementation, imports, annotations
    
    EXAMPLES:
        --include-only public,imports         # Only public APIs and imports
        --exclude-items comments,private      # Everything except comments and private
        --include-only public,protected,implementation  # Public/protected with bodies

FILE PATTERN FILTERING:

Controls which files are processed:

    --include PATTERNS     Include files matching patterns (comma-separated)
    --exclude PATTERNS     Exclude files matching patterns (comma-separated)
    
    Pattern syntax:
        *.ext              Files with specific extension
        **/pattern         Recursive directory matching
        dir/*              Files in specific directory
        *test*             Files containing "test"
    
    EXAMPLES:
        --include "*.go,*.py"              # Only Go and Python files
        --exclude "*test*,*spec*,*.json"   # Exclude test and config files
        --include "src/**/*.ts"            # TypeScript files in src/
        --exclude "vendor/*,node_modules/*" # Exclude dependency directories

PRACTICAL FILTERING COMBINATIONS:

Public API Documentation:
    aid ./src --include-only public,docstrings,imports
    # Perfect for generating API documentation

Security Analysis (All Code):
    aid ./ --public=1 --private=1 --protected=1 --internal=1 --implementation=1
    # Include everything for comprehensive security review

Refactoring Analysis:
    aid ./src --implementation=1 --comments=0
    # Focus on code structure and implementation

Architecture Overview:
    aid ./ --include-only public,protected,imports --exclude "*test*"
    # High-level structure without implementation details

Performance Analysis:
    aid ./ --implementation=1 --include "*.go,*.py" --exclude "*test*"
    # Implementation details for performance-critical languages

Clean Interface Export:
    aid ./lib --include-only public,docstrings
    # Clean API surface for library documentation

DEBUGGING FILTERING:

Use verbose mode to see what's being filtered:

    aid ./ -v --include-only public         # See what gets included
    aid ./ -vv --exclude "*test*"           # See exclusion details
    aid ./ -vvv --implementation=1          # Full filtering trace

FILTERING VALIDATION:

Check your filtering with dry-run approach:

    aid ./ --stdout --format=text           # Quick preview
    aid ./ --format=md                      # Formatted preview
    
Common mistakes:
    • Forgetting --private=1 for security analysis
    • Including --comments=1 when not needed (increases token count)
    • Not excluding test files (--exclude "*test*,*spec*")
    • Using --implementation=1 for API documentation

For complete examples: aid --help-extended

---
AI Distiller (aid) - https://github.com/janreges/ai-distiller
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)
Explore the project on GitHub: https://github.com/janreges/ai-distiller
`)
}

// showGitHelp displays git mode documentation
func showGitHelp() {
	fmt.Print(`GIT MODE - COMPLETE REFERENCE

When you target a .git directory, AI Distiller switches to a special git analysis
mode that formats commit history for LLM consumption. This is useful for
understanding project evolution, generating release notes, or analyzing
development patterns.

ACTIVATION:

Git mode is automatically activated when the target path is a .git directory:

    aid .git                    # Analyze current repository
    aid /path/to/repo/.git      # Analyze specific repository

OUTPUT FORMAT:

Git mode outputs commit history in a clean, LLM-optimized format:

    [commit_hash] YYYY-MM-DD HH:MM:SS | author_name | commit_subject
        commit_body_line_1
        commit_body_line_2
        (indented for readability)

OPTIONS:

    --git-limit NUM            Limit number of commits (default: 200)
                              0 = all commits
                              Useful for large repositories
    
    --with-analysis-prompt     Prepend comprehensive AI analysis prompt
                              Guides LLM to generate insights about:
                              • Contributor statistics and expertise areas
                              • Commit message quality analysis with scores
                              • Timeline visualization with Gantt charts
                              • Complexity and bug pattern analysis
                              • Development patterns and problem areas
                              • Actionable recommendations for improvement

EXAMPLES:

Basic Git History:
    aid .git                           # Last 200 commits
    aid .git --git-limit=50            # Last 50 commits  
    aid .git --git-limit=0             # All commits (careful with large repos)

With AI Analysis Guidance:
    aid .git --with-analysis-prompt    # Include analysis instructions
    aid .git --git-limit=100 --with-analysis-prompt  # Last 100 + analysis

Output Control:
    aid .git -o history.md             # Save to file
    aid .git --stdout                  # Print to stdout
    aid .git --format=md               # Markdown format (default for git mode)

TYPICAL WORKFLOWS:

Release Notes Generation:
    # Get commits since last release
    aid .git --git-limit=50 --with-analysis-prompt > release-analysis.md
    # Feed to LLM: "Generate release notes from this commit history"

Project Understanding:
    aid .git --git-limit=200 --with-analysis-prompt > project-evolution.md
    # Feed to LLM: "Explain this project's evolution and key architectural decisions"

Contributor Analysis:
    aid .git --git-limit=1000 --with-analysis-prompt > team-analysis.md  
    # Feed to LLM: "Analyze development patterns and contributor expertise"

CODE REVIEW INTEGRATION:

Combine git mode with regular distillation for comprehensive analysis:

    # 1. Get recent changes
    aid .git --git-limit=20 > recent-changes.md
    
    # 2. Get current codebase structure  
    aid --ai-action prompt-for-refactoring-suggestion . > code-analysis.md
    
    # 3. Feed both to LLM for comprehensive review

LIMITATIONS:

    • Only analyzes commit metadata and messages, not file changes
    • Respects git log order (chronological, newest first)
    • Does not include diff information (use git diff separately if needed)
    • Performance: Large repositories (>10k commits) may be slow

ANALYSIS PROMPT FEATURES:

When using --with-analysis-prompt, the output includes guidance for:

    Contributor Statistics:
    • Who are the main contributors?
    • What are their areas of expertise?
    • How has team composition changed?

    Commit Message Quality:
    • Overall project quality score (0-100%)
    • Individual contributor scores and feedback
    • Best/worst examples for learning
    • Personalized improvement recommendations

    Timeline Visualization:
    • Development phases with Gantt charts
    • Major features and release cycles
    • Bug fix periods and refactoring sprints
    • 10-25 timeline items based on project scope

    Complexity & Problem Areas:
    • Files with highest change frequency
    • Error-prone modules identification
    • Visualization with pie charts and flowcharts
    • Design challenge detection

    Bug Pattern Analysis:
    • Most problematic features
    • Common bug types and frequencies
    • Correlation between features and bugs
    • Visual bug distribution diagrams

    Evolution Insights:
    • Technology and architecture shifts
    • Technical debt indicators
    • Code health trends over time
    • Future development recommendations

For complete examples: aid --help-extended

---
AI Distiller (aid) - https://github.com/janreges/ai-distiller
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)
Explore the project on GitHub: https://github.com/janreges/ai-distiller
`)
}

// showCheatSheet displays a quick reference card
func showCheatSheet() {
	fmt.Print(`AI DISTILLER - QUICK REFERENCE

Transform source code into AI-optimized formats. Compress codebases 60-90%% while preserving
semantic information. Generate complete prompts - copy output directly to Gemini 2.5 Pro, 
ChatGPT-o3/4o, or Claude for perfect AI-powered analysis.

BASIC USAGE:
  aid ./project/src                       # Basic distillation (public APIs only, no implementation/comments)
  aid ./src --stdout                      # Print to terminal  
  aid ./src --format=md -o analysis.md    # Markdown output

AI ACTIONS:
  --ai-action prompt-for-refactoring-suggestion     # Refactoring analysis
  --ai-action prompt-for-complex-codebase-analysis  # Architecture overview  
  --ai-action prompt-for-security-analysis          # Security audit
  --ai-action prompt-for-performance-analysis       # Performance optimization
  
  Add --raw for full source code (large context needed - use Gemini 2.5 Pro)

ESSENTIAL FILTERING:
  --public=1 --private=1 --protected=1    # Include all visibility levels
  --implementation=1                      # Include function bodies
  --comments=1                           # Include comments
  --include "*.go,*.py"                  # File patterns
  --exclude "*test*"                     # Exclude tests

QUICK COMBINATIONS:
  # API Documentation
  aid ./src --include-only public,docstrings,imports

  # Security Analysis (everything)
  aid ./ --ai-action prompt-for-security-analysis --private=1

  # Performance Analysis
  aid ./ --ai-action prompt-for-performance-analysis --implementation=1

  # Git History with Comprehensive Analysis
  aid .git --git-limit=50 --with-analysis-prompt
  # Includes: commit quality scores, Gantt charts, bug patterns, complexity analysis

OUTPUT FORMATS:
  --format text        # Compact (default)
  --format md          # Human-readable markdown
  --format jsonl       # One JSON per file
  --format json-structured  # Rich semantic data

SUMMARY FORMATS:
  --summary-type visual-progress-bar   # Visual progress bar: [████████░░] 97.5% saved! (default)
  --summary-type stock-ticker          # Stock ticker: 📊 AID 97.5% ▲ | 5MB→128KB | ~1.2M tokens saved
  --summary-type speedometer-dashboard # Dashboard with dual progress bars
  --summary-type minimalist-sparkline  # Minimalist: ✨ 5MB → 128KB (97.5% saved)
  --summary-type ci-friendly           # CI/log friendly: [aid] ✓ 97.5% saved | 5MB → 128KB | 450ms
  --summary-type json                  # Machine-readable JSON output
  --summary-type off                   # Disable summary output
  --no-emoji                           # Remove emojis from summary

OUTPUT FILES:
  Default: .aid/ folder or .aid.*.txt (easy git recognition)
  --stdout available but large codebases may exceed AI context limits

HELP:
  aid --help-extended  # Complete documentation
  aid help actions     # AI actions details
  aid help filtering   # Filtering reference
  aid help git         # Git mode help

---
AI Distiller (aid) - https://github.com/janreges/ai-distiller
Authored by Claude Code & Ján Regeš from SiteOne (Czech Republic)
Explore the project on GitHub: https://github.com/janreges/ai-distiller
`)
}

// getAIActionsList returns a list of available AI actions with descriptions
func getAIActionsList() []string {
	registry := ai.NewActionRegistry()
	
	// Register actions to get the list
	aiactions.Register(registry)
	
	var actions []string
	actionList := registry.List()
	for _, action := range actionList {
		description := action.Description()
		actions = append(actions, fmt.Sprintf("  %-35s %s", action.Name(), description))
	}
	
	return actions
}

// showExtendedHelp displays extended help, loading from docs file if available
func showExtendedHelp() {
	// Try to load from docs/COMMAND-LINE-OPTIONS.md first
	execPath, err := os.Executable()
	if err == nil {
		// Determine the project root (executable should be in project root during development)
		projectRoot := filepath.Dir(execPath)
		docsPath := filepath.Join(projectRoot, "docs", "COMMAND-LINE-OPTIONS.md")
		
		// If not found, try relative to current directory (for development)
		if _, err := os.Stat(docsPath); os.IsNotExist(err) {
			docsPath = filepath.Join("docs", "COMMAND-LINE-OPTIONS.md")
		}
		
		if content, err := os.ReadFile(docsPath); err == nil {
			fmt.Print(string(content))
			return
		}
	}
	
	// Fallback to embedded content
	fmt.Print(extendedHelpContent)
}