package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	// First check in the same directory as the MCP binary
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
		"./bin/aid",
		"./aid",
		"../aid",
		"../../aid",
		"/usr/local/bin/aid",
		"/usr/bin/aid",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return filepath.Abs(loc)
		}
	}

	return "", fmt.Errorf("aid binary not found in PATH or common locations")
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

// executeAid runs the aid binary with the given arguments
func (s *DistillerService) executeAid(ctx context.Context, args []string) ([]byte, error) {
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

	// Run command
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("operation timed out after %s", timeout)
		}
		return nil, fmt.Errorf("aid command failed: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// HandleDistillFile processes a single file
func (s *DistillerService) HandleDistillFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}
	
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Check if file exists
	if info, err := os.Stat(absPath); err != nil || info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", filePath)), nil
	}

	// Build aid command
	cmdArgs := []string{absPath, "--stdout"}

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

	// Execute aid
	output, err := s.executeAid(ctx, cmdArgs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to distill file: %s", err)), nil
	}

	// Return result based on format
	if outputFormat == "json" {
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err != nil {
			return mcp.NewToolResultError("Failed to parse JSON output"), nil
		}
		// Convert to JSON string
		jsonBytes, _ := json.Marshal(jsonData)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

// HandleDistillDirectory processes an entire directory
func (s *DistillerService) HandleDistillDirectory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}
	
	dirPath, ok := args["directory_path"].(string)
	if !ok || dirPath == "" {
		return mcp.NewToolResultError("directory_path is required"), nil
	}

	// Sanitize path
	absPath, err := s.sanitizePath(dirPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Check if directory exists
	if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("Directory not found: %s", dirPath)), nil
	}

	// Build aid command
	cmdArgs := []string{absPath, "--stdout"}

	// Add options
	recursive, _ := args["recursive"].(bool)
	if recursive == false {
		// TODO: aid doesn't have a non-recursive flag yet, we'll need to handle this
		// For now, we'll process recursively always
	}

	includePrivate, _ := args["include_private"].(bool)
	includeImpl, _ := args["include_implementation"].(bool)
	includeComments, _ := args["include_comments"].(bool)
	includeImports, _ := args["include_imports"].(bool)
	outputFormat, _ := args["output_format"].(string)
	includePattern, _ := args["include_pattern"].(string)
	excludePattern, _ := args["exclude_pattern"].(string)

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

	// TODO: Handle include/exclude patterns when aid supports them
	_ = includePattern
	_ = excludePattern

	// Execute aid
	output, err := s.executeAid(ctx, cmdArgs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to distill directory: %s", err)), nil
	}

	// Return result based on format
	if outputFormat == "json" {
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err != nil {
			return mcp.NewToolResultError("Failed to parse JSON output"), nil
		}
		// Convert to JSON string
		jsonBytes, _ := json.Marshal(jsonData)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

// HandleListFiles lists files in a directory
func (s *DistillerService) HandleListFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		args = make(map[string]interface{}) // Empty args is ok for listFiles
	}
	
	path, _ := args["path"].(string)
	pattern, _ := args["pattern"].(string)
	recursive, _ := args["recursive"].(bool)

	// Sanitize path
	absPath, err := s.sanitizePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Build find command (using native Go instead of shelling out)
	var files []map[string]interface{}
	
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

		// Add file info
		files = append(files, map[string]interface{}{
			"path":          relPath,
			"size_bytes":    info.Size(),
			"language":      language,
			"last_modified": info.ModTime().Format(time.RFC3339),
		})

		// Limit results
		if len(files) >= s.maxFiles {
			return fmt.Errorf("max files reached")
		}

		return nil
	})

	// Handle error from filepath.Walk
	if walkErr != nil && walkErr.Error() != "max files reached" {
		// Ignore the custom "max files reached" error
	}

	// Create summary
	result := map[string]interface{}{
		"total_files": len(files),
		"files":       files,
		"root_path":   s.rootPath,
	}

	// Convert to JSON string
	jsonBytes, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// HandleGetFileContent reads raw file content
func (s *DistillerService) HandleGetFileContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}
	
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	startLine, _ := args["start_line"].(float64)
	endLine, _ := args["end_line"].(float64)

	// Sanitize path
	absPath, err := s.sanitizePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %s", err)), nil
	}

	// Check file size
	info, err := os.Stat(absPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("File not found: %s", filePath)), nil
	}

	// Limit file size (10MB)
	if info.Size() > 10*1024*1024 {
		return mcp.NewToolResultError("File too large (max 10MB)"), nil
	}

	// Read file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %s", err)), nil
	}

	// Handle line range if specified
	if startLine > 0 || endLine > 0 {
		lines := strings.Split(string(content), "\n")
		
		start := int(startLine) - 1
		if start < 0 {
			start = 0
		}
		
		end := int(endLine)
		if end <= 0 || end > len(lines) {
			end = len(lines)
		}

		if start < len(lines) && start < end {
			selectedLines := lines[start:end]
			content = []byte(strings.Join(selectedLines, "\n"))
		}
	}

	return mcp.NewToolResultText(string(content)), nil
}

// HandleSearch searches the codebase
func (s *DistillerService) HandleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments"), nil
	}
	
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	mode, _ := args["mode"].(string)
	if mode == "" {
		mode = "literal"
	}

	caseSensitive, _ := args["case_sensitive"].(bool)
	searchPath, _ := args["path"].(string)
	includePattern, _ := args["include_pattern"].(string)
	excludePattern, _ := args["exclude_pattern"].(string)
	maxResults, _ := args["max_results"].(float64)
	if maxResults <= 0 {
		maxResults = 100
	}

	// Sanitize search path
	var absPath string
	var err error
	if searchPath != "" {
		absPath, err = s.sanitizePath(searchPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %s", err)), nil
		}
	} else {
		absPath = s.rootPath
	}

	// Build ripgrep command
	rgArgs := []string{
		"--json",
		"--max-count", fmt.Sprintf("%d", int(maxResults)),
	}

	if !caseSensitive {
		rgArgs = append(rgArgs, "--ignore-case")
	}

	if mode == "regex" {
		rgArgs = append(rgArgs, "--regexp")
	} else {
		rgArgs = append(rgArgs, "--fixed-strings")
	}

	if includePattern != "" {
		rgArgs = append(rgArgs, "--glob", includePattern)
	}

	if excludePattern != "" {
		rgArgs = append(rgArgs, "--glob", "!"+excludePattern)
	}

	rgArgs = append(rgArgs, query, absPath)

	// Execute ripgrep
	cmd := exec.CommandContext(ctx, "rg", rgArgs...)
	output, _ := cmd.Output() // Ignore error as rg returns non-zero for no matches

	// Parse ripgrep JSON output
	var results []map[string]interface{}
	scanner := strings.Split(string(output), "\n")
	
	for _, line := range scanner {
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry["type"] == "match" {
			data := entry["data"].(map[string]interface{})
			path := data["path"].(map[string]interface{})
			
			// Get relative path
			filePath := path["text"].(string)
			relPath, _ := filepath.Rel(s.rootPath, filePath)

			results = append(results, map[string]interface{}{
				"file":        relPath,
				"line_number": data["line_number"],
				"line":        data["lines"].(map[string]interface{})["text"],
			})
		}
	}

	// Convert to JSON string
	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"query":         query,
		"mode":          mode,
		"total_matches": len(results),
		"matches":       results,
	})
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// detectLanguage detects the programming language from file extension
func detectLanguage(ext string) string {
	languageMap := map[string]string{
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".go":   "go",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".cs":   "csharp",
		".rb":   "ruby",
		".rs":   "rust",
		".swift": "swift",
		".kt":   "kotlin",
		".php":  "php",
		".r":    "r",
		".m":    "objective-c",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "unknown"
}