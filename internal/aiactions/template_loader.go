package aiactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// TemplateData contains data for template rendering
type TemplateData struct {
	ProjectName  string
	AnalysisDate string
}

// LoadTemplate loads and renders a template from the templates directory
func LoadTemplate(templateName string, data TemplateData) (string, error) {
	// Get executable path to find templates directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Try relative to executable first (for production)
	templatesDir := filepath.Join(filepath.Dir(execPath), "internal", "aiactions", "templates")
	templatePath := filepath.Join(templatesDir, templateName+".md")
	
	// If not found, try relative to current directory (for development)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatesDir = filepath.Join("internal", "aiactions", "templates")
		templatePath = filepath.Join(templatesDir, templateName+".md")
	}
	
	// Load template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	
	// Parse and execute template
	tmpl, err := template.New(templateName).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// CreateTemplateData creates template data with common values
func CreateTemplateData(projectName string) TemplateData {
	return TemplateData{
		ProjectName:  projectName,
		AnalysisDate: time.Now().Format("2006-01-02"),
	}
}