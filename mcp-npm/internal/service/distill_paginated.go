package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// PaginatedDistillDirectoryResponse represents paginated response for directory distillation
type PaginatedDistillDirectoryResponse struct {
	DirectoryPath   string                 `json:"directory_path"`
	OutputFormat    string                 `json:"output_format"`
	Content         interface{}            `json:"content"`
	EstimatedTokens int                    `json:"estimated_tokens"`
	NextPageToken   *string                `json:"next_page_token,omitempty"`
	CurrentPage     int                    `json:"current_page"`
	TotalPages      int                    `json:"total_pages,omitempty"`
	Statistics      map[string]interface{} `json:"statistics"`
	FilterSettings  map[string]interface{} `json:"filter_settings"`
}

// HandleDistillDirectoryPaginated is the new paginated version of HandleDistillDirectory
func (s *DistillerService) HandleDistillDirectoryPaginated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Extract pagination parameters
	pageTokenStr, _ := args["page_token"].(string)
	pageSize := 20000 // default
	if ps, ok := args["page_size"].(float64); ok {
		pageSize = int(ps)
	}
	noCache, _ := args["no_cache"].(bool)
	
	// Decode page token
	pageToken, err := DecodePageToken(pageTokenStr)
	if err != nil {
		return s.createErrorResponse(toolName, "INVALID_PAGE_TOKEN", fmt.Sprintf("Invalid page token: %s", err)), nil
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

	// Generate cache key
	cacheKey := s.cache.GenerateCacheKey(toolName, args)
	
	// Try to get from cache (only for subsequent pages, not first page)
	var units []ContentUnit
	var dirHash string
	
	// Use cache only for subsequent pages (when pageToken is present)
	if pageToken != nil && !noCache {
		if cached, err := s.cache.Get(cacheKey); err == nil && cached != nil {
			units = cached.Units
			dirHash = cached.DirectoryHash
			log.Printf("[%s] Using cached response for page %d (key: %s)", toolName, pageToken.PageNumber+1, cacheKey)
		}
	}
	
	// If not cached, execute aid command
	if units == nil {
		log.Printf("[%s] Executing aid command (cache miss or disabled)", toolName)
		
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
		
		// Split into units
		units = SplitIntoUnits(string(output), outputFormat)
		
		// Calculate directory hash
		paths := make([]string, len(units))
		for i, unit := range units {
			paths[i] = unit.Path
		}
		dirHash = CalculateDirectoryHash(paths)
		
		// Cache the result (if not disabled)
		if !noCache {
			metadata := map[string]interface{}{
				"directory_path": dirPath,
				"output_format":  outputFormat,
				"file_count":     len(units),
			}
			if err := s.cache.Put(cacheKey, units, dirHash, metadata); err != nil {
				log.Printf("[%s] Failed to cache response: %v", toolName, err)
			}
		}
	}
	
	// Create paginator
	paginator := NewPaginator(units, dirHash)
	
	// Get page
	result, err := paginator.GetPage(pageToken, pageSize)
	if err != nil {
		return s.createErrorResponse(toolName, "PAGINATION_ERROR", fmt.Sprintf("Pagination failed: %s", err)), nil
	}
	
	// Format content for response
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "text"
	}
	
	var content interface{}
	if outputFormat == "text" {
		// Reconstruct text format from units
		textContent := ""
		for _, unit := range result.Units {
			textContent += unit.Content.(string)
		}
		content = textContent
	} else {
		// For other formats, return the units directly
		content = result.Units
	}
	
	// Count files in directory for statistics
	fileCount := len(units)
	
	// Create structured response
	response := PaginatedDistillDirectoryResponse{
		DirectoryPath:   dirPath,
		OutputFormat:    outputFormat,
		Content:         content,
		EstimatedTokens: result.EstimatedTokens,
		NextPageToken:   result.NextPageToken,
		CurrentPage:     result.CurrentPage,
		TotalPages:      result.TotalPages,
		Statistics: map[string]interface{}{
			"cached":              units != nil && pageToken == nil && !noCache,
			"total_files":         fileCount,
			"files_in_page":       len(result.Units),
			"last_modified":       info.ModTime().Format(time.RFC3339),
		},
		FilterSettings: map[string]interface{}{
			"include_patterns":       args["include_patterns"],
			"exclude_patterns":       args["exclude_patterns"],
			"include_private":        args["include_private"],
			"include_implementation": args["include_implementation"],
			"recursive":              args["recursive"],
			"page_size":              pageSize,
		},
	}
	
	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}

// HandleDistillFilePaginated is the paginated version of HandleDistillFile
func (s *DistillerService) HandleDistillFilePaginated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// For single files, check if the output would be too large
	// and suggest using pagination parameters if needed
	
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

	// Estimate tokens
	estimatedTokens := EstimateTokens(string(output))
	
	// If output is too large, return warning
	if estimatedTokens > 20000 {
		return s.createErrorResponse(
			toolName, 
			"OUTPUT_TOO_LARGE", 
			fmt.Sprintf("Output contains ~%d tokens which exceeds safe limits. Consider using more restrictive parameters (e.g., include_implementation=false, include_private=false) or splitting the file.", estimatedTokens),
		), nil
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

	// Create structured response with token estimate
	response := struct {
		FilePath        string                 `json:"file_path"`
		FileSize        int64                  `json:"file_size_bytes"`
		Language        string                 `json:"detected_language"`
		OutputFormat    string                 `json:"output_format"`
		Content         interface{}            `json:"content"`
		EstimatedTokens int                    `json:"estimated_tokens"`
		Statistics      map[string]interface{} `json:"statistics"`
	}{
		FilePath:        filePath,
		FileSize:        info.Size(),
		Language:        detectLanguage(filepath.Ext(absPath)),
		OutputFormat:    outputFormat,
		Content:         content,
		EstimatedTokens: estimatedTokens,
		Statistics: map[string]interface{}{
			"output_size_bytes": len(output),
			"last_modified":     info.ModTime().Format(time.RFC3339),
		},
	}

	return s.createSuccessResponse(toolName, response, time.Since(startTime)), nil
}