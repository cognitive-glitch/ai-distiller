package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/janreges/ai-distiller/mcp-npm/internal/service"
	"github.com/janreges/ai-distiller/internal/project"
)

const (
	serverName    = "AI Distiller MCP"
	serverVersion = "1.0.0"
)

var (
	rootPath   string
	cacheDir   string
	maxFiles   int
	maxTimeout int
)

func main() {
	// Parse command line flags
	flag.StringVar(&rootPath, "root", "", "Root directory for analysis (defaults to current directory)")
	flag.StringVar(&cacheDir, "cache-dir", "", "Cache directory (defaults to ~/.cache/aid)")
	flag.IntVar(&maxFiles, "max-files", 10000, "Maximum number of files to process in a single request")
	flag.IntVar(&maxTimeout, "max-timeout", 60, "Maximum timeout in seconds for operations")
	flag.Parse()

	// Set defaults
	if rootPath == "" {
		if envRoot := os.Getenv("AID_ROOT"); envRoot != "" {
			rootPath = envRoot
		} else {
			rootPath, _ = os.Getwd()
		}
	}
	rootPath, _ = filepath.Abs(rootPath)

	if cacheDir == "" {
		// Try to get project root cache directory
		if projectCacheDir := getProjectCacheDir(); projectCacheDir != "" {
			cacheDir = projectCacheDir
		} else {
			// Fallback to user home cache
			home, _ := os.UserHomeDir()
			cacheDir = filepath.Join(home, ".cache", "aid")
		}
	}

	// Create cache directory if it doesn't exist
	os.MkdirAll(cacheDir, 0755)

	// Initialize MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Initialize service
	svc, err := service.NewDistillerService(rootPath, cacheDir, maxFiles, maxTimeout)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Register tools
	registerTools(s, svc)

	// Log startup info
	log.Printf("Starting %s v%s", serverName, serverVersion)
	log.Printf("Root path: %s", rootPath)
	log.Printf("Cache directory: %s", cacheDir)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getProjectCacheDir attempts to get the project root cache directory
func getProjectCacheDir() string {
	cacheDir, err := project.GetCacheDir()
	if err != nil {
		return ""
	}
	return cacheDir
}

func registerTools(s *server.MCPServer, svc *service.DistillerService) {
	// Core Tool: aid_analyze - The base engine for all AI analysis tasks
	registerAidAnalyze(s, svc)
	
	// Specialized Tools (Phase 0 rollout - 5 most important)
	registerAidHuntBugs(s, svc)
	registerAidSuggestRefactoring(s, svc)
	registerAidGenerateDiagram(s, svc)
	registerAidAnalyzeSecurity(s, svc)
	registerAidGenerateDocs(s, svc)
	
	// File Operations (backwards compatibility)
	registerFileOperations(s, svc)
	
	// Meta Tools
	registerMetaTools(s, svc)
}

// Core base tool for all AI analysis
func registerAidAnalyze(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_analyze",
		mcp.WithDescription("Core AI Distiller analysis engine with automatic pagination for large outputs. Use specialized tools (aid_hunt_bugs, aid_suggest_refactoring, etc.) when available. This tool directly maps to aid --ai-action for advanced or custom analysis flows. Responses are automatically paginated when exceeding ~20000 tokens.\n\nIMPORTANT: This tool generates analysis files on disk. The output includes file paths to the generated analysis. For best results, read these files directly instead of trying to process the entire analysis through MCP responses."),
		mcp.WithString("ai_action",
			mcp.Required(),
			mcp.Description("The specific AI action to execute"),
			mcp.Enum(
				"flow-for-deep-file-to-file-analysis",
				"flow-for-multi-file-docs",
				"prompt-for-refactoring-suggestion",
				"prompt-for-complex-codebase-analysis",
				"prompt-for-security-analysis",
				"prompt-for-performance-analysis",
				"prompt-for-best-practices-analysis",
				"prompt-for-bug-hunting",
				"prompt-for-single-file-docs",
				"prompt-for-diagrams",
			),
		),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or file to analyze (relative to project root)"),
		),
		mcp.WithString("user_query",
			mcp.Description("Optional specific question or instruction to guide the analysis"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format (default: md)"),
			mcp.Enum("md", "text", "json"),
		),
		mcp.WithBoolean("include_private",
			mcp.Description("LEGACY: Include all non-public members. For new code, use granular visibility controls"),
		),
		mcp.WithBoolean("include_public",
			mcp.Description("Include public members (default: true)"),
		),
		mcp.WithBoolean("include_protected",
			mcp.Description("Include protected members (default: false)"),
		),
		mcp.WithBoolean("include_internal",
			mcp.Description("Include internal/package-private members (default: false)"),
		),
		mcp.WithBoolean("include_private_new",
			mcp.Description("Include strictly private members (default: false)"),
		),
		mcp.WithBoolean("include_implementation",
			mcp.Description("Include function bodies for detailed analysis (default: false)"),
		),
		mcp.WithBoolean("include_comments",
			mcp.Description("Include comments (default: false)"),
		),
		mcp.WithBoolean("include_docstrings",
			mcp.Description("Include documentation comments (default: true)"),
		),
		mcp.WithBoolean("include_imports",
			mcp.Description("Include import statements (default: true)"),
		),
		mcp.WithBoolean("include_annotations",
			mcp.Description("Include decorators/annotations (default: true)"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts' (comma-separated)"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude: '*test*,*spec*' (comma-separated)"),
		),
	)
	s.AddTool(tool, svc.HandleAidAnalyze)
}

// Specialized tool for bug hunting
func registerAidHuntBugs(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_hunt_bugs",
		mcp.WithDescription("Systematically scans code files to identify potential bugs, logical errors, race conditions, and quality issues. Use when you suspect hidden bugs or want a comprehensive code health check. Returns detailed bug analysis with explanations and fix suggestions.\n\nOUTPUT: Generates a detailed markdown file with bug analysis. The response includes the file path - read this file directly for the complete analysis rather than processing through MCP pagination."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or file to scan for bugs"),
		),
		mcp.WithString("focus_area",
			mcp.Description("Specific area to focus on (e.g., 'null pointer exceptions', 'race conditions', 'data validation')"),
		),
		mcp.WithBoolean("include_private",
			mcp.Description("Include private code in analysis (recommended for thorough bug hunting, default: true)"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts'"),
		),
		mcp.WithString("exclude_patterns", 
			mcp.Description("File patterns to exclude: '*test*' (tests often have intentional edge cases)"),
		),
	)
	s.AddTool(tool, svc.HandleAidHuntBugs)
}

// Specialized tool for refactoring suggestions
func registerAidSuggestRefactoring(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_suggest_refactoring",
		mcp.WithDescription("Analyzes code to identify and suggest specific refactoring opportunities with concrete examples. Use to improve code quality, readability, maintainability, or performance. Returns actionable refactoring suggestions with before/after code examples.\n\nOUTPUT: Generates a comprehensive refactoring analysis markdown file. The response includes the file path - read this file directly for detailed suggestions with code examples."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or file to analyze for refactoring opportunities"),
		),
		mcp.WithString("refactoring_goal",
			mcp.Required(),
			mcp.Description("Main objective (e.g., 'improve readability', 'reduce complexity in calculate_totals function', 'make more testable', 'optimize performance')"),
		),
		mcp.WithBoolean("include_implementation",
			mcp.Description("Include function bodies for detailed analysis (recommended, default: true)"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts'"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude: '*test*,*spec*'"),
		),
	)
	s.AddTool(tool, svc.HandleAidSuggestRefactoring)
}

// Specialized tool for diagram generation
func registerAidGenerateDiagram(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_generate_diagram",
		mcp.WithDescription("Generates architectural diagrams from source code using Mermaid format. Creates 10 beneficial diagrams including flowcharts, sequence diagrams, class diagrams, and architecture overviews. Perfect for understanding complex systems and documenting architecture.\n\nOUTPUT: Generates a markdown file with multiple Mermaid diagrams. The response includes the file path - read this file to view and render all diagrams."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or files to generate diagrams from"),
		),
		mcp.WithString("diagram_focus",
			mcp.Description("What aspect to focus on (e.g., 'user authentication flow', 'data processing pipeline', 'class relationships', 'overall architecture')"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts'"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude: '*test*,*spec*'"),
		),
	)
	s.AddTool(tool, svc.HandleAidGenerateDiagram)
}

// Specialized tool for security analysis
func registerAidAnalyzeSecurity(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_analyze_security",
		mcp.WithDescription("Performs comprehensive security analysis with OWASP Top 10 focus. Identifies potential vulnerabilities, security anti-patterns, and weak points. Use for security audits, compliance checks, or before production deployment. Returns security findings with risk levels and remediation steps.\n\nOUTPUT: Generates a detailed security audit markdown file. The response includes the file path - read this file for complete vulnerability analysis and remediation recommendations."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or file to analyze for security issues"),
		),
		mcp.WithString("security_focus",
			mcp.Description("Specific security area (e.g., 'input validation', 'authentication', 'data encryption', 'SQL injection')"),
		),
		mcp.WithBoolean("include_private",
			mcp.Description("Include private code (recommended for comprehensive security analysis, default: true)"),
		),
		mcp.WithBoolean("include_implementation",
			mcp.Description("Include function bodies (essential for security analysis, default: true)"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts'"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude (usually none for security analysis)"),
		),
	)
	s.AddTool(tool, svc.HandleAidAnalyzeSecurity)
}

// Specialized tool for documentation generation
func registerAidGenerateDocs(s *server.MCPServer, svc *service.DistillerService) {
	tool := mcp.NewTool("aid_generate_docs",
		mcp.WithDescription("Generates comprehensive documentation for source code including API references, usage examples, and developer guides. Creates structured documentation workflows for single files or entire projects. Perfect for creating technical documentation from code.\n\nOUTPUT: Generates one or more markdown documentation files. The response includes file paths - read these files directly for the complete documentation."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Path to directory or file to generate documentation for"),
		),
		mcp.WithString("doc_type",
			mcp.Description("Type of documentation"),
			mcp.Enum("single-file-docs", "multi-file-docs", "api-reference"),
		),
		mcp.WithString("audience",
			mcp.Description("Target audience (e.g., 'developers', 'end-users', 'API consumers', 'maintainers')"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include: '*.go,*.py,*.ts'"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude: '*test*,*spec*,*internal*'"),
		),
	)
	s.AddTool(tool, svc.HandleAidGenerateDocs)
}

// File operations for backwards compatibility
func registerFileOperations(s *server.MCPServer, svc *service.DistillerService) {
	// distill_file tool (backwards compatibility)
	distillFileTool := mcp.NewTool("distill_file",
		mcp.WithDescription("Extracts essential code structure from a single file. Legacy tool - prefer aid_analyze or specialized tools for new workflows.\n\nNOTE: For files that might exceed token limits, the tool will warn you. Consider using more restrictive parameters (include_implementation=false, include_private=false) or using aid_analyze tools that save results to files.\n\nVISIBILITY: Use either legacy include_private OR new granular controls (include_public/protected/internal/private_new)"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to source file"),
		),
		mcp.WithBoolean("include_private",
			mcp.Description("LEGACY: Include all non-public members (private+protected+internal). For new code, use granular controls"),
		),
		mcp.WithBoolean("include_public",
			mcp.Description("Include public members (default: true)"),
		),
		mcp.WithBoolean("include_protected",
			mcp.Description("Include protected members (default: false)"),
		),
		mcp.WithBoolean("include_internal",
			mcp.Description("Include internal/package-private members (default: false)"),
		),
		mcp.WithBoolean("include_private_new",
			mcp.Description("Include strictly private members (default: false). Use this instead of legacy include_private"),
		),
		mcp.WithBoolean("include_implementation",
			mcp.Description("Include function/method bodies (default: false)"),
		),
		mcp.WithBoolean("include_comments",
			mcp.Description("Include comments (default: false)"),
		),
		mcp.WithBoolean("include_docstrings",
			mcp.Description("Include documentation comments (default: true)"),
		),
		mcp.WithBoolean("include_imports",
			mcp.Description("Include import statements (default: true)"),
		),
		mcp.WithBoolean("include_annotations",
			mcp.Description("Include decorators/annotations (default: true)"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format"),
			mcp.Enum("text", "md", "json"),
		),
	)
	s.AddTool(distillFileTool, svc.HandleDistillFile)

	// distill_directory tool (backwards compatibility with pagination support)
	distillDirTool := mcp.NewTool("distill_directory",
		mcp.WithDescription("Extracts code structure from directories with automatic pagination for large results. Returns paginated responses when content exceeds ~20000 tokens. Use page_token to get subsequent pages.\n\nFILE PATTERN FILTERING:\n- Include patterns: \"*.go\", \"*.go,*.py\", \"src/**/*.js\"\n- Exclude patterns: \"*_test.go\", \"vendor/**\", \"*/tmp/*\", \"node_modules/**\"\n- Supports multiple patterns (comma-separated)\n- Supports ** for recursive directory matching\n- Examples:\n  - Exclude tests: exclude_patterns=\"*_test.go,*.spec.js\"\n  - Only source files: include_patterns=\"src/**/*.ts\", exclude_patterns=\"**/*.d.ts\"\n  - Skip dependencies: exclude_patterns=\"vendor/**,node_modules/**,build/**\"\n\nCACHING STRATEGY for large codebases:\n- First page: Call with no_cache=true to ensure fresh data and populate cache\n- Subsequent pages: Use cached data (default) for consistency\n- Cache TTL: 5 minutes\n- Alternative: For very large analyses, consider using aid_analyze which saves results to files that can be read directly"),
		mcp.WithString("directory_path",
			mcp.Required(),
			mcp.Description("Path to directory"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Process recursively (default: true)"),
		),
		mcp.WithBoolean("include_private",
			mcp.Description("LEGACY: Include all non-public members (private+protected+internal). For new code, use granular controls"),
		),
		mcp.WithBoolean("include_public",
			mcp.Description("Include public members (default: true)"),
		),
		mcp.WithBoolean("include_protected",
			mcp.Description("Include protected members (default: false)"),
		),
		mcp.WithBoolean("include_internal",
			mcp.Description("Include internal/package-private members (default: false)"),
		),
		mcp.WithBoolean("include_private_new",
			mcp.Description("Include strictly private members (default: false). Use this instead of legacy include_private"),
		),
		mcp.WithBoolean("include_implementation",
			mcp.Description("Include function/method bodies (default: false)"),
		),
		mcp.WithBoolean("include_comments",
			mcp.Description("Include comments (default: false)"),
		),
		mcp.WithBoolean("include_docstrings",
			mcp.Description("Include documentation comments (default: true)"),
		),
		mcp.WithBoolean("include_imports",
			mcp.Description("Include import statements (default: true)"),
		),
		mcp.WithBoolean("include_annotations",
			mcp.Description("Include decorators/annotations (default: true)"),
		),
		mcp.WithString("include_patterns",
			mcp.Description("File patterns to include (comma-separated): '*.go,*.py' or 'src/**/*.js'"),
		),
		mcp.WithString("exclude_patterns",
			mcp.Description("File patterns to exclude (comma-separated): '*_test.go,vendor/**'"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format"),
			mcp.Enum("text", "md", "json"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Maximum tokens per page (1000-20000, default: 20000)"),
		),
		mcp.WithString("page_token",
			mcp.Description("Token for retrieving next page of results"),
		),
		mcp.WithBoolean("no_cache",
			mcp.Description("Disable caching (default: false, cache TTL: 5 minutes)"),
		),
		mcp.WithNumber("workers",
			mcp.Description("Number of parallel workers (0=auto which is 80% CPU cores, 1=serial, 2+=specific count)"),
		),
	)
	s.AddTool(distillDirTool, svc.HandleDistillDirectory)

	// list_files tool
	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription("Lists project files with language detection and statistics."),
		mcp.WithString("path",
			mcp.Description("Directory path to scan"),
		),
		mcp.WithString("pattern",
			mcp.Description("File pattern filter"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Scan recursively (default: true)"),
		),
	)
	s.AddTool(listFilesTool, svc.HandleListFiles)
}

// Meta tools
func registerMetaTools(s *server.MCPServer, svc *service.DistillerService) {
	// analyze_git_history tool - Enhanced version
	gitHistoryTool := mcp.NewTool("analyze_git_history",
		mcp.WithDescription("Analyzes git commit history with AI insights using AI Distiller's git mode. Provides comprehensive commit analysis, developer patterns, and optional AI-ready analysis prompts.\n\nFEATURES:\n- Supports all aid git mode capabilities\n- Configurable commit limit (default: 200)\n- Optional AI analysis prompt for quality scoring and visualizations\n- Clean formatted output optimized for AI consumption\n\nUSE CASES:\n- PR/MR automated reviews and impact analysis\n- Release note generation\n- Developer productivity insights\n- Code quality trends over time"),
		mcp.WithNumber("git_limit",
			mcp.Description("Number of commits to analyze (0 = all commits, default: 200)"),
		),
		mcp.WithBoolean("with_analysis_prompt",
			mcp.Description("Include comprehensive AI analysis prompt for commit quality scoring, timeline visualizations, and pattern detection (default: true)"),
		),
		mcp.WithString("repository_path",
			mcp.Description("Path to repository (default: current directory). Must contain .git directory"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format: text (default), json"),
			mcp.Enum("text", "json"),
		),
	)
	s.AddTool(gitHistoryTool, svc.HandleAnalyzeGitHistory)

	// get_capabilities tool
	capabilitiesTool := mcp.NewTool("get_capabilities",
		mcp.WithDescription("Returns comprehensive information about AI Distiller capabilities, supported languages, and available tools."),
	)
	s.AddTool(capabilitiesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		capabilities := map[string]interface{}{
			"server_name":    serverName,
			"server_version": serverVersion,
			"root_path":      rootPath,
			"cache_dir":      cacheDir,
			"tools": map[string]interface{}{
				"specialized": []string{
					"aid_hunt_bugs",
					"aid_suggest_refactoring", 
					"aid_generate_diagram",
					"aid_analyze_security",
					"aid_generate_docs",
				},
				"core": []string{
					"aid_analyze",
				},
				"legacy": []string{
					"distill_file",
					"distill_directory",
					"list_files",
				},
				"meta": []string{
					"get_capabilities",
					"analyze_git_history",
				},
			},
			"ai_actions": []string{
				"flow-for-deep-file-to-file-analysis",
				"flow-for-multi-file-docs", 
				"prompt-for-refactoring-suggestion",
				"prompt-for-complex-codebase-analysis",
				"prompt-for-security-analysis",
				"prompt-for-performance-analysis",
				"prompt-for-best-practices-analysis",
				"prompt-for-bug-hunting",
				"prompt-for-single-file-docs",
				"prompt-for-diagrams",
			},
			"supported_languages": []string{
				"python", "typescript", "javascript", "go", "java",
				"csharp", "rust", "ruby", "swift", "kotlin", "php", "cpp", "c",
			},
			"supported_formats": []string{
				"text", "md", "json", "xml", "jsonl",
			},
			"features": []string{
				"ai_actions", "pattern_filtering", "specialized_analysis",
				"diagram_generation", "security_analysis", "bug_hunting",
				"refactoring_suggestions", "documentation_generation",
				"pagination", "caching", "git_mode_analysis",
			},
			"caching_strategy": map[string]interface{}{
				"ttl_seconds": 300,
				"cache_dir": filepath.Join(cacheDir, "mcp"),
				"recommendations": []string{
					"For large codebases: First call with no_cache=true to ensure fresh data",
					"Subsequent pages will use cache for consistency",
					"For AI analysis tools: Read generated files directly from disk",
					"Cache is automatically invalidated after 5 minutes",
				},
			},
			"pagination": map[string]interface{}{
				"default_page_size": 20000,
				"max_page_size": 20000,
				"token_limit": 25000,
				"usage": "Use page_token from response to get next page",
			},
		}
		jsonBytes, _ := json.Marshal(capabilities)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}