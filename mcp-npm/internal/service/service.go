package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// DistillerService handles MCP requests by calling the aid binary
type DistillerService struct {
	rootPath   string
	cacheDir   string
	maxFiles   int
	maxTimeout int
	aidBinary  string
}

// NewDistillerService creates a new distiller service
func NewDistillerService(rootPath, cacheDir string, maxFiles, maxTimeout int) (*DistillerService, error) {
	// Find aid binary
	aidBinary, err := findAidBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to find aid binary: %w", err)
	}

	return &DistillerService{
		rootPath:   rootPath,
		cacheDir:   cacheDir,
		maxFiles:   maxFiles,
		maxTimeout: maxTimeout,
		aidBinary:  aidBinary,
	}, nil
}

// findAidBinary locates the aid binary
func findAidBinary() (string, error) {
	// First check for pre-built binary in build directory
	if rootDir := os.Getenv("AID_ROOT"); rootDir != "" {
		buildPath := filepath.Join(rootDir, "build", "aid")
		if _, err := os.Stat(buildPath); err == nil {
			return buildPath, nil
		}
	}

	// Then check in the same directory as the MCP binary
	exePath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exePath)
		aidPath := filepath.Join(dir, "aid")
		if _, err := os.Stat(aidPath); err == nil {
			return aidPath, nil
		}
	}

	// Then check if it's in PATH
	if path, err := exec.LookPath("aid"); err == nil {
		return path, nil
	}

	// Check common locations
	locations := []string{
		"./build/aid",
		"./bin/aid",
		"./aid",
		"../aid",
		"../../aid",
		"../../../build/aid",
		"/usr/local/bin/aid",
		"/usr/bin/aid",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return filepath.Abs(loc)
		}
	}

	return "", fmt.Errorf("aid binary not found in PATH, build directory, or common locations")
}

// sanitizePath ensures the path is within the root directory
func (s *DistillerService) sanitizePath(path string) (string, error) {
	if path == "" || path == "." {
		return s.rootPath, nil
	}

	// Clean and make absolute
	absPath := filepath.Join(s.rootPath, path)
	absPath = filepath.Clean(absPath)

	// Ensure it's within root
	if !strings.HasPrefix(absPath, s.rootPath) {
		return "", fmt.Errorf("path traversal detected: %s", path)
	}

	return absPath, nil
}

// getCacheKey generates a cache key for the given parameters
func (s *DistillerService) getCacheKey(tool string, params map[string]interface{}) string {
	// Create deterministic key
	parts := []string{tool}
	
	// Add sorted parameters
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	data := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ToolResponse represents the unified response structure for all tools
type ToolResponse struct {
	Status   string                 `json:"status"`
	ToolName string                 `json:"tool_name"`
	Data     interface{}           `json:"data,omitempty"`
	Error    *ToolError            `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ToolError represents structured error information
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// executeAid runs the aid binary with the given arguments and tracks performance
func (s *DistillerService) executeAid(ctx context.Context, toolName string, args []string) ([]byte, error) {
	startTime := time.Now()
	
	// Create command with timeout
	timeout := time.Duration(s.maxTimeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.aidBinary, args...)
	cmd.Dir = s.rootPath

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Log command for debugging
	log.Printf("[%s] Executing aid: %s %v", toolName, s.aidBinary, args)

	// Run command
	err := cmd.Run()
	duration := time.Since(startTime)
	
	// Log performance metrics
	outputSize := stdout.Len()
	log.Printf("[%s] Performance: duration=%dms, output_size=%d bytes", 
		toolName, duration.Milliseconds(), outputSize)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[%s] ERROR: Timeout after %s", toolName, timeout)
			return nil, fmt.Errorf("operation timed out after %s", timeout)
		}
		log.Printf("[%s] ERROR: Command failed: %v, stderr: %s", toolName, err, stderr.String())
		return nil, fmt.Errorf("aid command failed: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// createSuccessResponse creates a standardized success response
func (s *DistillerService) createSuccessResponse(toolName string, data interface{}, duration time.Duration) *mcp.CallToolResult {
	response := ToolResponse{
		Status:   "success",
		ToolName: toolName,
		Data:     data,
		Metadata: map[string]interface{}{
			"execution_duration_ms": duration.Milliseconds(),
			"server_version":        "1.1.0",
		},
	}
	
	jsonBytes, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonBytes))
}

// createErrorResponse creates a standardized error response
func (s *DistillerService) createErrorResponse(toolName string, code string, message string) *mcp.CallToolResult {
	response := ToolResponse{
		Status:   "error",
		ToolName: toolName,
		Error: &ToolError{
			Code:    code,
			Message: message,
		},
		Metadata: map[string]interface{}{
			"server_version": "1.1.0",
		},
	}
	
	jsonBytes, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonBytes))
}

// buildAidArgs builds aid command arguments from parameters
func (s *DistillerService) buildAidArgs(args map[string]interface{}, basePath string) []string {
	cmdArgs := []string{basePath, "--stdout"}

	// Add options based on parameters
	includePrivate, _ := args["include_private"].(bool)
	includeImpl, _ := args["include_implementation"].(bool)
	includeComments, _ := args["include_comments"].(bool)
	includeImports, _ := args["include_imports"].(bool)
	outputFormat, _ := args["output_format"].(string)

	// Set visibility flags
	if includePrivate {
		cmdArgs = append(cmdArgs, "--private=1", "--protected=1", "--internal=1")
	}

	// Set content flags
	if includeImpl {
		cmdArgs = append(cmdArgs, "--implementation=1")
	}
	if includeComments {
		cmdArgs = append(cmdArgs, "--comments=1")
	}
	if includeImports == false {
		cmdArgs = append(cmdArgs, "--imports=0")
	}

	// Set format
	if outputFormat != "" && outputFormat != "text" {
		cmdArgs = append(cmdArgs, "--format", outputFormat)
	}

	// Add include/exclude patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	return cmdArgs
}

// DistillFileResponse represents structured response for file distillation
type DistillFileResponse struct {
	FilePath     string                 `json:"file_path"`
	FileSize     int64                  `json:"file_size_bytes"`
	Language     string                 `json:"detected_language"`
	OutputFormat string                 `json:"output_format"`
	Content      interface{}           `json:"content"`
	Statistics   map[string]interface{} `json:"statistics"`
}

// HandleDistillFile processes a single file using AI Distiller (aid)
func (s *DistillerService) HandleDistillFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "distill_file"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return s.createErrorResponse(toolName, "MISSING_FILE_PATH", "file_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(filePath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid file path: %s", err)), nil
	}

	// Check if file exists and get info
	info, err := os.Stat(absPath)
	if err != nil {
		return s.createErrorResponse(toolName, "FILE_NOT_FOUND", fmt.Sprintf("File not found: %s", filePath)), nil
	}
	if info.IsDir() {
		return s.createErrorResponse(toolName, "PATH_IS_DIRECTORY", fmt.Sprintf("Path points to directory, not file: %s", filePath)), nil
	}

	// Build and execute aid command
	cmdArgs := s.buildAidArgs(args, absPath)
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "DISTILLATION_FAILED", fmt.Sprintf("Failed to distill file: %s", err)), nil
	}

	// Parse output based on format
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "text"
	}

	var content interface{}
	if outputFormat == "json" {
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err != nil {
			return s.createErrorResponse(toolName, "JSON_PARSE_ERROR", "Failed to parse JSON output from aid"), nil
		}
		content = jsonData
	} else {
		content = string(output)
	}

	// Create structured response
	response := DistillFileResponse{
		FilePath:     filePath,
		FileSize:     info.Size(),
		Language:     detectLanguage(filepath.Ext(absPath)),
		OutputFormat: outputFormat,
		Content:      content,
		Statistics: map[string]interface{}{
			"output_size_bytes": len(output),
			"last_modified":     info.ModTime().Format(time.RFC3339),
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// DistillDirectoryResponse represents structured response for directory distillation
type DistillDirectoryResponse struct {
	DirectoryPath   string                 `json:"directory_path"`
	OutputFormat    string                 `json:"output_format"`
	Content         interface{}           `json:"content"`
	Statistics      map[string]interface{} `json:"statistics"`
	FilterSettings  map[string]interface{} `json:"filter_settings"`
}

// HandleDistillDirectory processes an entire directory using AI Distiller (aid) 
func (s *DistillerService) HandleDistillDirectory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "distill_directory"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	dirPath, ok := args["directory_path"].(string)
	if !ok || dirPath == "" {
		return s.createErrorResponse(toolName, "MISSING_DIRECTORY_PATH", "directory_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(dirPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid directory path: %s", err)), nil
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return s.createErrorResponse(toolName, "DIRECTORY_NOT_FOUND", fmt.Sprintf("Directory not found: %s", dirPath)), nil
	}
	if !info.IsDir() {
		return s.createErrorResponse(toolName, "PATH_IS_FILE", fmt.Sprintf("Path points to file, not directory: %s", dirPath)), nil
	}

	// Build and execute aid command
	cmdArgs := s.buildAidArgs(args, absPath)
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "DISTILLATION_FAILED", fmt.Sprintf("Failed to distill directory: %s", err)), nil
	}

	// Parse output based on format
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "text"
	}

	var content interface{}
	if outputFormat == "json" {
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err != nil {
			return s.createErrorResponse(toolName, "JSON_PARSE_ERROR", "Failed to parse JSON output from aid"), nil
		}
		content = jsonData
	} else {
		content = string(output)
	}

	// Count files in directory for statistics
	fileCount := s.countFilesInDirectory(absPath)

	// Create structured response
	response := DistillDirectoryResponse{
		DirectoryPath: dirPath,
		OutputFormat:  outputFormat,
		Content:       content,
		Statistics: map[string]interface{}{
			"output_size_bytes":   len(output),
			"total_files_scanned": fileCount,
			"last_modified":       info.ModTime().Format(time.RFC3339),
		},
		FilterSettings: map[string]interface{}{
			"include_patterns":     args["include_patterns"],
			"exclude_patterns":     args["exclude_patterns"],
			"include_private":      args["include_private"],
			"include_implementation": args["include_implementation"],
			"recursive":            args["recursive"],
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// countFilesInDirectory counts the number of files in a directory (for statistics)
func (s *DistillerService) countFilesInDirectory(dirPath string) int {
	count := 0
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

// AITaskListResponse represents structured response for AI task list generation
type AITaskListResponse struct {
	AnalysisPath    string                 `json:"analysis_path"`
	TaskList        interface{}           `json:"task_list"`
	Statistics      map[string]interface{} `json:"statistics"`
	FilterSettings  map[string]interface{} `json:"filter_settings"`
}

// HandleGenerateAITaskList generates comprehensive AI-driven task lists using AI Distiller (aid) --ai-analysis-task-list
func (s *DistillerService) HandleProposeCodeAnalysisPlan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "propose_code_analysis_plan"
	
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Get path (default to current directory)
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}

	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid analysis path: %s", err)), nil
	}

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		return s.createErrorResponse(toolName, "PATH_NOT_FOUND", fmt.Sprintf("Analysis path not found: %s", path)), nil
	}

	// Build aid command for code analysis plan generation
	cmdArgs := []string{absPath, "--ai-analysis-task-list"}

	// Add include/exclude patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "TASK_GENERATION_FAILED", fmt.Sprintf("Failed to generate AI task list: %s", err)), nil
	}

	// Try to parse as structured content (if aid returns structured data)
	var taskList interface{}
	taskList = string(output) // Default to raw text
	
	// Attempt JSON parsing if it looks like JSON
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "{") || strings.HasPrefix(outputStr, "[") {
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err == nil {
			taskList = jsonData
		}
	}

	// Create structured response
	response := map[string]interface{}{
		"analysis_path":       path,
		"code_analysis_plan": taskList,
		"statistics": map[string]interface{}{
			"output_size_bytes": len(output),
			"task_count":        strings.Count(string(output), "\n"),
		},
		"filter_settings": map[string]interface{}{
			"include_patterns": args["include_patterns"],
			"exclude_patterns": args["exclude_patterns"],
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// GitHistoryResponse represents structured response for git history analysis
type GitHistoryResponse struct {
	CommitLimit       int                    `json:"commit_limit"`
	AnalysisPrompt    string                 `json:"analysis_prompt,omitempty"`
	CommitHistory     interface{}           `json:"commit_history"`
	Statistics        map[string]interface{} `json:"statistics"`
	WithAnalysisPrompt bool                  `json:"with_analysis_prompt"`
}

// HandleAnalyzeGitHistory analyzes git commit history with AI insights using AI Distiller (aid)
func (s *DistillerService) HandleAnalyzeGitHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "analyze_git_history"
	
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Build aid command for git history analysis
	cmdArgs := []string{".git", "--stdout"}

	// Add git-specific options
	gitLimit, _ := args["git_limit"].(float64)
	if gitLimit <= 0 {
		gitLimit = 200 // Default
	}
	cmdArgs = append(cmdArgs, "--git-limit", fmt.Sprintf("%d", int(gitLimit)))

	withAnalysisPrompt, exists := args["with_analysis_prompt"].(bool)
	if !exists || withAnalysisPrompt { // Default true
		cmdArgs = append(cmdArgs, "--with-analysis-prompt")
	}

	// Execute aid
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "GIT_ANALYSIS_FAILED", fmt.Sprintf("Failed to analyze git history: %s", err)), nil
	}

	// Parse output - separate analysis prompt from commit history if present
	outputStr := string(output)
	var analysisPrompt string
	var commitHistory interface{}
	
	if withAnalysisPrompt {
		// Look for analysis prompt separator or patterns
		lines := strings.Split(outputStr, "\n")
		promptEnd := -1
		for i, line := range lines {
			if strings.Contains(line, "[") && strings.Contains(line, "]") && 
			   (strings.Contains(line, "20") || strings.Contains(line, "19")) { // Date pattern
				promptEnd = i
				break
			}
		}
		
		if promptEnd > 0 {
			analysisPrompt = strings.Join(lines[:promptEnd], "\n")
			commitHistory = strings.Join(lines[promptEnd:], "\n")
		} else {
			commitHistory = outputStr
		}
	} else {
		commitHistory = outputStr
	}

	// Count commits for statistics
	commitCount := strings.Count(outputStr, "[")
	if commitCount == 0 {
		commitCount = strings.Count(outputStr, "\n")
	}

	// Create structured response
	response := GitHistoryResponse{
		CommitLimit:       int(gitLimit),
		AnalysisPrompt:    analysisPrompt,
		CommitHistory:     commitHistory,
		WithAnalysisPrompt: withAnalysisPrompt,
		Statistics: map[string]interface{}{
			"output_size_bytes": len(output),
			"commit_count":      commitCount,
			"has_analysis_prompt": analysisPrompt != "",
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// ListFilesResponse represents structured response for file listing
type ListFilesResponse struct {
	SearchPath      string                 `json:"search_path"`
	Pattern         string                 `json:"pattern,omitempty"`
	Recursive       bool                   `json:"recursive"`
	Files           []FileInfo            `json:"files"`
	Statistics      map[string]interface{} `json:"statistics"`
	LanguageBreakdown map[string]int       `json:"language_breakdown"`
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path         string `json:"path"`
	SizeBytes    int64  `json:"size_bytes"`
	Language     string `json:"language"`
	LastModified string `json:"last_modified"`
}

// HandleListFiles lists files in a directory with enhanced metadata using AI Distiller patterns
func (s *DistillerService) HandleListFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "list_files"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{}) // Empty args is ok for listFiles
	}
	
	path, _ := args["path"].(string)
	pattern, _ := args["pattern"].(string)
	recursive, _ := args["recursive"].(bool)
	if recursive == false && args["recursive"] == nil {
		recursive = true // Default to true
	}

	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid search path: %s", err)), nil
	}

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		return s.createErrorResponse(toolName, "PATH_NOT_FOUND", fmt.Sprintf("Search path not found: %s", path)), nil
	}

	// Build find command (using native Go instead of shelling out)
	var files []FileInfo
	languageCount := make(map[string]int)
	var totalSize int64
	
	walkErr := filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			if !recursive && filePath != absPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check pattern if provided
		if pattern != "" {
			matched, _ := filepath.Match(pattern, filepath.Base(filePath))
			if !matched {
				return nil
			}
		}

		// Get relative path
		relPath, _ := filepath.Rel(s.rootPath, filePath)

		// Detect language
		ext := filepath.Ext(filePath)
		language := detectLanguage(ext)
		languageCount[language]++
		totalSize += info.Size()

		// Add file info
		files = append(files, FileInfo{
			Path:         relPath,
			SizeBytes:    info.Size(),
			Language:     language,
			LastModified: info.ModTime().Format(time.RFC3339),
		})

		// Limit results
		if len(files) >= s.maxFiles {
			return fmt.Errorf("max files reached")
		}

		return nil
	})

	// Handle error from filepath.Walk
	if walkErr != nil && walkErr.Error() != "max files reached" {
		log.Printf("[%s] Walk error (ignored): %v", toolName, walkErr)
	}

	// Create structured response
	response := ListFilesResponse{
		SearchPath:        path,
		Pattern:           pattern,
		Recursive:         recursive,
		Files:             files,
		LanguageBreakdown: languageCount,
		Statistics: map[string]interface{}{
			"total_files":      len(files),
			"total_size_bytes": totalSize,
			"max_files_limit":  s.maxFiles,
			"search_truncated": len(files) >= s.maxFiles,
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// ReadFileResponse - REMOVED (not needed)
// type ReadFileResponse struct {
// 	FilePath     string                 `json:"file_path"`
// 	FileSize     int64                  `json:"file_size_bytes"`
// 	Language     string                 `json:"detected_language"`
// 	Level        string                 `json:"level"`
// 	Content      interface{}            `json:"content"`
// 	Statistics   map[string]interface{} `json:"statistics"`
// }

// HandleReadFile - REMOVED per user request
// Use distill_file instead for semantic code analysis
func (s *DistillerService) HandleReadFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createErrorResponse("read_file", "TOOL_REMOVED", "This tool has been removed. Please use distill_file instead."), nil
}

// createFileSummary creates a natural language summary from distilled JSON data
func (s *DistillerService) createFileSummary(filePath string, data map[string]interface{}) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("File: %s", filepath.Base(filePath)))
	
	// Count different types of elements
	var classCount, functionCount, importCount int
	
	// This is a simplified version - in real implementation, you'd parse the JSON structure properly
	if classes, ok := data["classes"].([]interface{}); ok {
		classCount = len(classes)
	}
	if functions, ok := data["functions"].([]interface{}); ok {
		functionCount = len(functions)
	}
	if imports, ok := data["imports"].([]interface{}); ok {
		importCount = len(imports)
	}
	
	if classCount > 0 {
		parts = append(parts, fmt.Sprintf("Contains %d class(es)", classCount))
	}
	if functionCount > 0 {
		parts = append(parts, fmt.Sprintf("Contains %d function(s)", functionCount))
	}
	if importCount > 0 {
		parts = append(parts, fmt.Sprintf("Imports %d module(s)", importCount))
	}
	
	return strings.Join(parts, ". ")
}

// SearchResponse - REMOVED (not needed)
// type SearchResponse struct {
// 	Query          string                 `json:"query"`
// 	SearchMode     string                 `json:"search_mode"`
// 	SearchPath     string                 `json:"search_path"`
// 	CaseSensitive  bool                   `json:"case_sensitive"`
// 	Matches        []SearchMatch         `json:"matches"`
// 	Statistics     map[string]interface{} `json:"statistics"`
// 	FilterSettings map[string]interface{} `json:"filter_settings"`
// }
//
// // SearchMatch - REMOVED (not needed)
// type SearchMatch struct {
// 	File       string `json:"file"`
// 	LineNumber int    `json:"line_number"`
// 	Line       string `json:"line"`
// 	Language   string `json:"language"`
// }

// HandleFindCode performs semantic code search for definitions, references, and patterns
// HandleFindCode - REMOVED per user request
// Aid currently has no special semantic search features
func (s *DistillerService) HandleFindCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createErrorResponse("find_code", "TOOL_REMOVED", "This tool has been removed. Aid currently has no special semantic search features."), nil
}

// LogAnalysisResponse represents structured response for log analysis (WOW FEATURE!)
type LogAnalysisResponse struct {
	Summary         LogAnalysisSummary     `json:"summary"`
	LogFiles        []LogFileResult       `json:"log_files"`
	AIAnalysisPrompt string               `json:"ai_analysis_prompt,omitempty"`
	OutputFormat    string                `json:"output_format"`
}

// LogAnalysisSummary provides overview statistics for log analysis
type LogAnalysisSummary struct {
	TotalLogFiles     int    `json:"total_log_files"`
	SearchPath        string `json:"search_path"`
	MaxFiles          int    `json:"max_files"`
	LinesPerFile      int    `json:"lines_per_file"`
	TotalEntries      int    `json:"total_entries"`
	AnalysisTimestamp string `json:"analysis_timestamp"`
}

// LogFileResult represents analysis results for a single log file
type LogFileResult struct {
	SourceFile      string     `json:"source_file"`
	AbsolutePath    string     `json:"absolute_path"`
	SizeBytes       int64      `json:"size_bytes"`
	LastModified    string     `json:"last_modified"`
	LogType         string     `json:"log_type"`
	LinesExtracted  int        `json:"lines_extracted"`
	Entries         []LogEntry `json:"entries"`
}

// LogEntry represents a single log entry with metadata
type LogEntry struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	SourceFile string `json:"source_file"`
	Timestamp  string `json:"timestamp,omitempty"`
	LogLevel   string `json:"log_level,omitempty"`
}

// HandleAnalyzeLogs performs comprehensive log analysis - finds X newest log files and extracts Y last lines from each (WOW FEATURE!)
func (s *DistillerService) HandleAnalyzeLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "analyze_logs"
	
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	// Extract parameters with defaults
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}

	maxFiles, _ := args["max_files"].(float64)
	if maxFiles <= 0 {
		maxFiles = 10 // Default X=10
	}

	linesPerFile, _ := args["lines_per_file"].(float64)
	if linesPerFile <= 0 {
		linesPerFile = 100 // Default Y=100
	}

	includeAnalysisPrompt, _ := args["include_analysis_prompt"].(bool)
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "ndjson" // Default to NDJSON for structured data
	}

	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid log search path: %s", err)), nil
	}

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		return s.createErrorResponse(toolName, "PATH_NOT_FOUND", fmt.Sprintf("Log search path not found: %s", path)), nil
	}

	// Find log files recursively
	logFiles, err := s.findLogFiles(absPath, int(maxFiles))
	if err != nil {
		return s.createErrorResponse(toolName, "LOG_DISCOVERY_FAILED", fmt.Sprintf("Failed to find log files: %s", err)), nil
	}

	if len(logFiles) == 0 {
		return s.createErrorResponse(toolName, "NO_LOGS_FOUND", "No log files found in the specified path"), nil
	}

	// Extract lines from each log file
	var logFileResults []LogFileResult
	for _, logFile := range logFiles {
		lines, err := s.extractLastLines(logFile.Path, int(linesPerFile))
		if err != nil {
			// Skip files with errors but continue processing
			log.Printf("[%s] Failed to extract lines from %s: %v", toolName, logFile.Path, err)
			continue
		}

		// Create metadata for this log file
		relPath, _ := filepath.Rel(s.rootPath, logFile.Path)
		var entries []LogEntry

		// Process each line with metadata
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			entry := LogEntry{
				LineNumber: i + (s.getFileLineCount(logFile.Path) - len(lines) + 1),
				Content:    line,
				SourceFile: relPath,
				Timestamp:  extractTimestamp(line),
				LogLevel:   extractLogLevel(line),
			}

			entries = append(entries, entry)
		}

		logFileResult := LogFileResult{
			SourceFile:     relPath,
			AbsolutePath:   logFile.Path,
			SizeBytes:      logFile.Size,
			LastModified:   logFile.ModTime.Format(time.RFC3339),
			LogType:        detectLogType(logFile.Path),
			LinesExtracted: len(lines),
			Entries:        entries,
		}

		logFileResults = append(logFileResults, logFileResult)
	}

	// Calculate total entries
	totalEntries := 0
	for _, result := range logFileResults {
		totalEntries += len(result.Entries)
	}

	// Build structured response
	response := LogAnalysisResponse{
		Summary: LogAnalysisSummary{
			TotalLogFiles:     len(logFileResults),
			SearchPath:        path,
			MaxFiles:          int(maxFiles),
			LinesPerFile:      int(linesPerFile),
			TotalEntries:      totalEntries,
			AnalysisTimestamp: time.Now().Format(time.RFC3339),
		},
		LogFiles:     logFileResults,
		OutputFormat: outputFormat,
	}

	// Add AI analysis prompt if requested
	if includeAnalysisPrompt {
		response.AIAnalysisPrompt = generateLogAnalysisPrompt(logFileResults, int(maxFiles), int(linesPerFile))
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// generateLogAnalysisPrompt creates comprehensive AI analysis prompt (updated signature)
func generateLogAnalysisPrompt(logFiles []LogFileResult, maxFiles, linesPerFile int) string {
	return fmt.Sprintf(`# Comprehensive Log Analysis Task

## Context
You have been provided with log data from %d newest log files (max %d requested), with up to %d lines extracted from each file. Your task is to perform a thorough analysis to identify patterns, anomalies, issues, and insights.

## Analysis Framework

### 1. **Error & Issue Detection**
- Identify all ERROR, FATAL, CRITICAL level entries
- Look for stack traces, exception patterns, and failure indicators
- Flag unusual error frequencies or patterns
- Note any cascading failures or error storms

### 2. **Performance & Resource Analysis**
- Identify slow operations, timeouts, or performance degradation
- Look for memory issues, CPU spikes, or resource exhaustion
- Analyze response times and throughput patterns
- Flag any performance anomalies or bottlenecks

### 3. **Security & Access Analysis**
- Identify suspicious access patterns or authentication failures
- Look for potential security threats or breaches
- Analyze access logs for unusual user behavior
- Flag any security-related warnings or alerts

### 4. **Operational Insights**
- Analyze deployment patterns, restarts, or configuration changes
- Identify system health indicators and status changes
- Look for maintenance windows or planned operations
- Note any infrastructure or environment issues

### 5. **Temporal Pattern Analysis**
- Identify time-based patterns (peak usage, quiet periods)
- Look for correlations between different log sources
- Analyze event sequences and dependencies
- Flag any unusual timing patterns

### 6. **Anomaly Detection**
- Identify entries that deviate from normal patterns
- Look for unusual data values, formats, or structures
- Flag unexpected system behavior or responses
- Note any configuration or environmental anomalies

## Output Requirements

Please provide:

1. **Executive Summary** (2-3 sentences of key findings)
2. **Critical Issues** (any urgent problems requiring immediate attention)
3. **Error Analysis** (detailed breakdown of errors and their potential causes)
4. **Performance Insights** (performance trends and optimization opportunities)
5. **Security Assessment** (security-related findings and recommendations)
6. **Operational Recommendations** (actionable steps for improvement)
7. **Notable Patterns** (interesting trends or correlations discovered)
8. **Risk Assessment** (potential risks and their severity levels)

## Analysis Guidelines

- **Be Specific**: Reference actual log entries with timestamps when relevant
- **Prioritize by Impact**: Focus on issues that affect system reliability or security
- **Provide Context**: Explain why certain patterns or entries are significant
- **Suggest Actions**: Offer concrete recommendations for addressing identified issues
- **Consider Dependencies**: Look for relationships between different log sources
- **Think Systematically**: Consider both immediate issues and longer-term trends

Begin your analysis now, processing the provided log data comprehensively.`, 
		len(logFiles), maxFiles, linesPerFile)
}

// findLogFiles recursively finds log files and returns the newest ones
func (s *DistillerService) findLogFiles(rootPath string, maxFiles int) ([]fileInfo, error) {
	var allLogFiles []fileInfo

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is a log file
		if isLogFile(path, info) {
			allLogFiles = append(allLogFiles, fileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(allLogFiles)-1; i++ {
		for j := i + 1; j < len(allLogFiles); j++ {
			if allLogFiles[i].ModTime.Before(allLogFiles[j].ModTime) {
				allLogFiles[i], allLogFiles[j] = allLogFiles[j], allLogFiles[i]
			}
		}
	}

	// Return top maxFiles
	if len(allLogFiles) > maxFiles {
		allLogFiles = allLogFiles[:maxFiles]
	}

	return allLogFiles, nil
}

// fileInfo holds file metadata
type fileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// isLogFile determines if a file is a log file based on multiple criteria
func isLogFile(path string, info os.FileInfo) bool {
	// Skip binary files and very large files (>100MB)
	if info.Size() > 100*1024*1024 {
		return false
	}

	fileName := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))

	// Check by extension
	logExtensions := []string{".log", ".logs", ".txt", ".out", ".err"}
	for _, logExt := range logExtensions {
		if ext == logExt {
			return true
		}
	}

	// Check by filename patterns
	logPatterns := []string{
		"log", "logs", "access", "error", "debug", "trace", "audit",
		"syslog", "messages", "journal", "output", "console",
	}
	for _, pattern := range logPatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	// Check for rotated log patterns (file.log.1, file.log.gz, etc.)
	if strings.Contains(fileName, ".log.") {
		return true
	}

	return false
}

// extractLastLines reads the last N lines from a file efficiently
func (s *DistillerService) extractLastLines(filePath string, numLines int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// For efficiency, read from the end if file is large
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var lines []string
	if stat.Size() > 1024*1024 { // If file > 1MB, use tail-like approach
		lines, err = s.tailFile(file, numLines)
	} else {
		// For smaller files, read all and take last lines
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		allLines := strings.Split(string(content), "\n")
		start := len(allLines) - numLines
		if start < 0 {
			start = 0
		}
		lines = allLines[start:]
	}

	return lines, err
}

// tailFile implements efficient tail functionality for large files
func (s *DistillerService) tailFile(file *os.File, numLines int) ([]string, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Start from end and read backwards in chunks
	fileSize := stat.Size()
	bufSize := int64(4096)
	pos := fileSize
	var lines []string
	var buffer []byte

	for len(lines) < numLines && pos > 0 {
		// Determine read position and size
		readSize := bufSize
		if pos < bufSize {
			readSize = pos
		}
		pos -= readSize

		// Read chunk
		chunk := make([]byte, readSize)
		_, err := file.ReadAt(chunk, pos)
		if err != nil {
			return nil, err
		}

		// Prepend to buffer
		buffer = append(chunk, buffer...)

		// Split into lines
		allLines := strings.Split(string(buffer), "\n")
		if len(allLines) > 1 {
			// Keep last partial line in buffer for next iteration
			buffer = []byte(allLines[0])
			lines = append(allLines[1:], lines...)
		}
	}

	// Take only the requested number of lines
	if len(lines) > numLines {
		lines = lines[len(lines)-numLines:]
	}

	return lines, nil
}

// getFileLineCount efficiently counts total lines in a file
func (s *DistillerService) getFileLineCount(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := strings.Split
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	return len(scanner(string(content), "\n"))
}

// detectLogType determines the type of log file
func detectLogType(path string) string {
	fileName := strings.ToLower(filepath.Base(path))
	
	typePatterns := map[string]string{
		"access":    "access_log",
		"error":     "error_log", 
		"nginx":     "nginx_log",
		"apache":    "apache_log",
		"syslog":    "system_log",
		"messages":  "system_log",
		"debug":     "debug_log",
		"trace":     "trace_log",
		"audit":     "audit_log",
		"console":   "console_log",
		"output":    "application_log",
	}

	for pattern, logType := range typePatterns {
		if strings.Contains(fileName, pattern) {
			return logType
		}
	}

	return "generic_log"
}

// extractTimestamp attempts to extract timestamp from log line
func extractTimestamp(line string) string {
	// Common timestamp patterns
	patterns := []string{
		`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`,     // ISO format
		`\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2}`,        // Apache format
		`\w{3} \d{2} \d{2}:\d{2}:\d{2}`,              // Syslog format
		`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`,        // Standard format
	}

	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			// Extract first 20 characters that might contain timestamp
			if len(line) >= 20 {
				return strings.TrimSpace(line[:20])
			}
		}
	}

	return ""
}

// extractLogLevel attempts to extract log level from log line  
func extractLogLevel(line string) string {
	line = strings.ToUpper(line)
	levels := []string{"ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE", "FATAL", "CRITICAL"}
	
	for _, level := range levels {
		if strings.Contains(line, level) {
			return level
		}
	}

	return ""
}


// detectLanguage detects the programming language from file extension
func detectLanguage(ext string) string {
	languageMap := map[string]string{
		".py":    "python",
		".js":    "javascript", 
		".ts":    "typescript",
		".go":    "go",
		".java":  "java",
		".cpp":   "cpp",
		".c":     "c",
		".cs":    "csharp",
		".rb":    "ruby",
		".rs":    "rust",
		".swift": "swift",
		".kt":    "kotlin",
		".php":   "php",
		".r":     "r",
		".m":     "objective-c",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "unknown"
}

// HandleExplainCodeStructure generates high-level overview with AI prompt
func (s *DistillerService) HandleExplainCodeStructure(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "explain_code_structure"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	
	// Get path (default to current directory)
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}
	
	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}
	
	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		return s.createErrorResponse(toolName, "PATH_NOT_FOUND", fmt.Sprintf("Path not found: %s", path)), nil
	}
	
	// Build aid command - use default behavior (public only, no implementation, with comments/docstrings)
	cmdArgs := []string{absPath, "--stdout", "--comments=1", "--docstrings=1"}
	
	// Add include/exclude patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}
	
	// Execute aid
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "ANALYSIS_FAILED", fmt.Sprintf("Failed to analyze code structure: %s", err)), nil
	}
	
	// Generate comprehensive AI prompt
	prompt := fmt.Sprintf(`# AI Code Structure Explanation Request

## CONTEXT
This analysis was generated by AI Distiller (aid). It contains a high-level, "public API" view of the directory "%s", including all public functions, classes, methods, and their documentation. Implementation details have been omitted to provide a concise architectural overview.

- **Target Path:** %s
- **Analysis Timestamp:** %s

## DISTILLED CODE OVERVIEW
%s

## TASK FOR AI AGENT
As an expert software architect, your task is to synthesize the provided code overview into a comprehensive architectural summary.

### 1. Identify Core Components
List the primary classes, modules, and functions found in the codebase.

### 2. Describe Responsibilities
For each component, describe its main purpose and responsibilities based on its name and documentation.

### 3. Infer Relationships
Explain how these components likely interact with each other to fulfill the overall functionality of the module.

### 4. Architecture Patterns
Identify any common design patterns or architectural styles evident in the code structure.

### 5. Generate a Mermaid Diagram
Create a mermaid.js graph diagram (using the 'graph TD' syntax) to visually represent the component relationships. Enclose it in a 'mermaid' code block.

### 6. Key Insights
Provide 3-5 key insights about the codebase architecture that would be valuable for a developer new to this project.

Please structure your response with clear headings for each section.`, path, path, time.Now().Format(time.RFC3339), string(output))
	
	// Create structured response
	response := map[string]interface{}{
		"analysis_path":     path,
		"distilled_content": string(output),
		"ai_prompt":        prompt,
		"statistics": map[string]interface{}{
			"content_size_bytes": len(output),
			"prompt_size_bytes":  len(prompt),
			"total_size_bytes":   len(output) + len(prompt),
		},
		"filter_settings": map[string]interface{}{
			"include_patterns": args["include_patterns"],
			"exclude_patterns": args["exclude_patterns"],
		},
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleSuggestRefactoring analyzes code with implementation and generates refactoring prompt
func (s *DistillerService) HandleSuggestRefactoring(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "suggest_refactoring"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	
	// Get path (required - can be file or directory)
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return s.createErrorResponse(toolName, "MISSING_PATH", "path parameter is required"), nil
	}
	
	// Get refactoring goal
	goal, _ := args["goal"].(string)
	if goal == "" {
		goal = "general code quality improvements"
	}
	
	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}
	
	// Verify path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return s.createErrorResponse(toolName, "PATH_NOT_FOUND", fmt.Sprintf("Path not found: %s", path)), nil
	}
	
	// Build aid command - include implementation for refactoring analysis
	cmdArgs := []string{absPath, "--stdout", "--implementation=1", "--private=1", "--protected=1", "--internal=1"}
	
	// Add include/exclude patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}
	
	// Execute aid
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "ANALYSIS_FAILED", fmt.Sprintf("Failed to analyze code: %s", err)), nil
	}
	
	// Check if output is too large (token limit protection)
	const maxBytes = 500000 // ~100k tokens
	if len(output) > maxBytes {
		return s.createErrorResponse(toolName, "CONTENT_TOO_LARGE", 
			fmt.Sprintf("Content size (%d bytes) exceeds limit (%d bytes). Please analyze a smaller scope.", len(output), maxBytes)), nil
	}
	
	// Determine if it's a file or directory
	targetType := "file"
	if info.IsDir() {
		targetType = "directory"
	}
	
	// Generate comprehensive refactoring prompt
	prompt := fmt.Sprintf(`# AI Refactoring Analysis Request

## CONTEXT
This analysis was generated by AI Distiller (aid), a tool that extracts and structures code for AI review. The user is requesting refactoring suggestions for a specific part of their codebase. The source code provided below is complete for the specified %s, including all implementation details.

- **User's Refactoring Goal:** %s
- **Target Path:** %s
- **Analysis Timestamp:** %s

## SOURCE CODE WITH FULL IMPLEMENTATION
%s

## TASK FOR AI AGENT
As an expert software engineer specializing in code quality and refactoring, analyze the provided source code and suggest improvements.

### 1. Code Analysis
Perform a thorough analysis of the code, identifying:
- Code smells and anti-patterns
- Complexity issues
- Maintainability concerns
- Performance bottlenecks
- Security vulnerabilities
- Testing gaps

### 2. Refactoring Opportunities
Based on your analysis, identify the top 5-10 refactoring opportunities, prioritized by impact and feasibility.

### 3. Detailed Refactoring Plan
For the top 3 most important refactorings:
- Explain what needs to be changed and why
- Provide the refactored code using GitHub-style diff format (with + and - markers)
- Estimate the effort required (low/medium/high)
- List any risks or considerations

### 4. Best Practices Alignment
Explain how your suggested refactorings align with:
- SOLID principles
- Clean Code principles
- Language-specific best practices
- Common design patterns

### 5. Testing Recommendations
Suggest any new tests that should be added to ensure the refactored code maintains correctness.

Please be specific and actionable in your recommendations, focusing on changes that will have the most positive impact on code quality and maintainability.`, targetType, goal, path, time.Now().Format(time.RFC3339), string(output))
	
	// Create structured response
	response := map[string]interface{}{
		"analysis_path":     path,
		"refactoring_goal":  goal,
		"source_content":    string(output),
		"ai_prompt":        prompt,
		"statistics": map[string]interface{}{
			"content_size_bytes": len(output),
			"prompt_size_bytes":  len(prompt),
			"total_size_bytes":   len(output) + len(prompt),
			"is_directory":       info.IsDir(),
		},
		"filter_settings": map[string]interface{}{
			"include_patterns": args["include_patterns"],
			"exclude_patterns": args["exclude_patterns"],
		},
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleGetDependencies analyzes dependencies for a given symbol
// HandleGetDependencies - REMOVED per user request
// Aid currently has no special dependency analysis features
func (s *DistillerService) HandleGetDependencies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createErrorResponse("get_dependencies", "TOOL_REMOVED", "This tool has been removed. Aid currently has no special dependency analysis features."), nil
}

// NEW AI ACTION HANDLERS

// HandleAidAnalyze - Core handler for the base aid_analyze tool
func (s *DistillerService) HandleAidAnalyze(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_analyze"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	aiAction, ok := args["ai_action"].(string)
	if !ok || aiAction == "" {
		return s.createErrorResponse(toolName, "MISSING_AI_ACTION", "ai_action parameter is required"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build aid command arguments
	cmdArgs := []string{absPath, "--ai-action", aiAction, "--stdout"}
	
	// Add optional parameters
	if userQuery, ok := args["user_query"].(string); ok && userQuery != "" {
		// For now, store user query in metadata - future enhancement could pass it to aid
		log.Printf("[%s] User query: %s", toolName, userQuery)
	}
	
	if outputFormat, ok := args["output_format"].(string); ok && outputFormat != "" && outputFormat != "md" {
		cmdArgs = append(cmdArgs, "--format", outputFormat)
	}
	
	if includePrivate, _ := args["include_private"].(bool); includePrivate {
		cmdArgs = append(cmdArgs, "--private=1", "--protected=1", "--internal=1")
	}
	
	if includeImpl, _ := args["include_implementation"].(bool); includeImpl {
		cmdArgs = append(cmdArgs, "--implementation=1")
	}
	
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute analysis: %s", err)), nil
	}

	// Create response
	response := map[string]interface{}{
		"ai_action":    aiAction,
		"target_path":  targetPath,
		"output":      string(output),
		"output_size": len(output),
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleAidHuntBugs - Specialized handler for bug hunting
func (s *DistillerService) HandleAidHuntBugs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_hunt_bugs"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build aid command arguments for bug hunting
	cmdArgs := []string{absPath, "--ai-action", "prompt-for-bug-hunting", "--stdout"}
	
	// Default to include private code for thorough bug hunting
	includePrivate := true
	if val, ok := args["include_private"].(bool); ok {
		includePrivate = val
	}
	if includePrivate {
		cmdArgs = append(cmdArgs, "--private=1", "--protected=1", "--internal=1")
	}
	
	// Always include implementation for bug hunting
	cmdArgs = append(cmdArgs, "--implementation=1")
	
	// Add patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute bug hunting: %s", err)), nil
	}

	// Create response with focus area context
	response := map[string]interface{}{
		"target_path":  targetPath,
		"focus_area":   args["focus_area"],
		"output":      string(output),
		"output_size": len(output),
		"analysis_type": "bug_hunting",
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleAidSuggestRefactoring - Specialized handler for refactoring suggestions
func (s *DistillerService) HandleAidSuggestRefactoring(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_suggest_refactoring"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}
	
	refactoringGoal, ok := args["refactoring_goal"].(string)
	if !ok || refactoringGoal == "" {
		return s.createErrorResponse(toolName, "MISSING_REFACTORING_GOAL", "refactoring_goal parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build aid command arguments for refactoring
	cmdArgs := []string{absPath, "--ai-action", "prompt-for-refactoring-suggestion", "--stdout"}
	
	// Default to include implementation for refactoring analysis
	includeImpl := true
	if val, ok := args["include_implementation"].(bool); ok {
		includeImpl = val
	}
	if includeImpl {
		cmdArgs = append(cmdArgs, "--implementation=1")
	}
	
	// Add patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute refactoring analysis: %s", err)), nil
	}

	// Create response with refactoring goal context
	response := map[string]interface{}{
		"target_path":       targetPath,
		"refactoring_goal":  refactoringGoal,
		"output":           string(output),
		"output_size":      len(output),
		"analysis_type":    "refactoring_suggestions",
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleAidGenerateDiagram - Specialized handler for diagram generation
func (s *DistillerService) HandleAidGenerateDiagram(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_generate_diagram"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build aid command arguments for diagram generation
	cmdArgs := []string{absPath, "--ai-action", "prompt-for-diagrams", "--stdout"}
	
	// Add patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute diagram generation: %s", err)), nil
	}

	// Create response with diagram focus context
	response := map[string]interface{}{
		"target_path":    targetPath,
		"diagram_focus":  args["diagram_focus"],
		"output":        string(output),
		"output_size":   len(output),
		"analysis_type": "diagram_generation",
		"format":        "mermaid",
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleAidAnalyzeSecurity - Specialized handler for security analysis
func (s *DistillerService) HandleAidAnalyzeSecurity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_analyze_security"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build aid command arguments for security analysis
	cmdArgs := []string{absPath, "--ai-action", "prompt-for-security-analysis", "--stdout"}
	
	// Default to include private code and implementation for security analysis
	includePrivate := true
	if val, ok := args["include_private"].(bool); ok {
		includePrivate = val
	}
	if includePrivate {
		cmdArgs = append(cmdArgs, "--private=1", "--protected=1", "--internal=1")
	}
	
	includeImpl := true
	if val, ok := args["include_implementation"].(bool); ok {
		includeImpl = val
	}
	if includeImpl {
		cmdArgs = append(cmdArgs, "--implementation=1")
	}
	
	// Add patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute security analysis: %s", err)), nil
	}

	// Create response with security focus context
	response := map[string]interface{}{
		"target_path":     targetPath,
		"security_focus":  args["security_focus"],
		"output":         string(output),
		"output_size":    len(output),
		"analysis_type":  "security_analysis",
		"framework":      "OWASP_Top_10",
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleAidGenerateDocs - Specialized handler for documentation generation
func (s *DistillerService) HandleAidGenerateDocs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	toolName := "aid_generate_docs"
	
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return s.createErrorResponse(toolName, "INVALID_ARGUMENTS", "Invalid arguments provided"), nil
	}
	
	targetPath, ok := args["target_path"].(string)
	if !ok || targetPath == "" {
		return s.createErrorResponse(toolName, "MISSING_TARGET_PATH", "target_path parameter is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(targetPath)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PATH", fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Determine AI action based on doc_type
	docType, _ := args["doc_type"].(string)
	aiAction := "prompt-for-single-file-docs" // default
	if docType == "multi-file-docs" {
		aiAction = "flow-for-multi-file-docs"
	}

	// Build aid command arguments for documentation generation
	cmdArgs := []string{absPath, "--ai-action", aiAction, "--stdout"}
	
	// Add patterns
	if includePatterns, ok := args["include_patterns"].(string); ok && includePatterns != "" {
		cmdArgs = append(cmdArgs, "--include", includePatterns)
	}
	
	if excludePatterns, ok := args["exclude_patterns"].(string); ok && excludePatterns != "" {
		cmdArgs = append(cmdArgs, "--exclude", excludePatterns)
	}

	// Execute aid command
	output, err := s.executeAid(ctx, toolName, cmdArgs)
	if err != nil {
		return s.createErrorResponse(toolName, "EXECUTION_ERROR", fmt.Sprintf("Failed to execute documentation generation: %s", err)), nil
	}

	// Create response with documentation context
	response := map[string]interface{}{
		"target_path":    targetPath,
		"doc_type":       docType,
		"audience":       args["audience"],
		"output":        string(output),
		"output_size":   len(output),
		"analysis_type": "documentation_generation",
		"ai_action":     aiAction,
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}
