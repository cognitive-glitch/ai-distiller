package ai

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExpandTemplate expands template variables in a path template
func ExpandTemplate(template string, ctx *ActionContext) string {
	result := template
	
	// Time-based replacements
	now := ctx.Timestamp
	result = strings.ReplaceAll(result, "%YYYY%", fmt.Sprintf("%04d", now.Year()))
	result = strings.ReplaceAll(result, "%MM%", fmt.Sprintf("%02d", now.Month()))
	result = strings.ReplaceAll(result, "%DD%", fmt.Sprintf("%02d", now.Day()))
	result = strings.ReplaceAll(result, "%HH%", fmt.Sprintf("%02d", now.Hour()))
	result = strings.ReplaceAll(result, "%mm%", fmt.Sprintf("%02d", now.Minute()))
	result = strings.ReplaceAll(result, "%SS%", fmt.Sprintf("%02d", now.Second()))
	
	// Combined date/time formats
	result = strings.ReplaceAll(result, "%YYYY-MM-DD%", now.Format("2006-01-02"))
	result = strings.ReplaceAll(result, "%HH-MM-SS%", now.Format("15-04-05"))
	result = strings.ReplaceAll(result, "%YYYY-MM-DD.HH-MM-SS%", now.Format("2006-01-02.15-04-05"))
	
	// Project-based replacements
	baseName := ctx.BaseName
	if baseName == "" {
		baseName = "project"
	}
	result = strings.ReplaceAll(result, "%folder-basename%", baseName)
	result = strings.ReplaceAll(result, "%project%", baseName)
	
	// Clean up any double slashes
	result = filepath.Clean(result)
	
	return result
}

// ValidateOutputPath validates that an output path is safe
func ValidateOutputPath(path string, projectPath string) error {
	// Ensure the path is not absolute (unless it's within project)
	if filepath.IsAbs(path) {
		absProject, _ := filepath.Abs(projectPath)
		absPath, _ := filepath.Abs(path)
		
		// Check if the absolute path is within the project directory
		if !strings.HasPrefix(absPath, absProject) {
			return fmt.Errorf("output path must be within project directory")
		}
	}
	
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		cleaned := filepath.Clean(path)
		if strings.Contains(cleaned, "..") {
			return fmt.Errorf("output path cannot contain '..' for security reasons")
		}
	}
	
	return nil
}