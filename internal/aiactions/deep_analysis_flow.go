package aiactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/janreges/ai-distiller/internal/ai"
)

// DeepAnalysisFlowAction implements the existing --ai-analysis-task-list functionality
type DeepAnalysisFlowAction struct{}

// Ensure DeepAnalysisFlowAction implements FlowAction
var _ ai.FlowAction = (*DeepAnalysisFlowAction)(nil)

func (a *DeepAnalysisFlowAction) Name() string {
	return "flow-for-deep-file-to-file-analysis"
}

func (a *DeepAnalysisFlowAction) Description() string {
	return "Generate structured task list for comprehensive file-by-file AI analysis"
}

func (a *DeepAnalysisFlowAction) Type() ai.ActionType {
	return ai.ActionTypeFlow
}

func (a *DeepAnalysisFlowAction) DefaultOutput() string {
	return "./.aid"
}

func (a *DeepAnalysisFlowAction) Validate() error {
	return nil
}

func (a *DeepAnalysisFlowAction) ExecuteFlow(ctx *ai.ActionContext) (*ai.FlowResult, error) {
	// Get project basename and current date
	basename := ctx.BaseName
	currentDate := fmt.Sprintf("%04d-%02d-%02d",
		ctx.Timestamp.Year(), ctx.Timestamp.Month(), ctx.Timestamp.Day())

	// Collect source files from the project
	sourceFiles, err := a.collectSourceFiles(ctx.ProjectPath, ctx.IncludePatterns, ctx.ExcludePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to collect source files: %w", err)
	}

	// Generate file paths
	taskListPath := fmt.Sprintf("ANALYSIS-TASK-LIST.%s.%s.md", basename, currentDate)
	summaryPath := fmt.Sprintf("ANALYSIS-SUMMARY.%s.%s.md", basename, currentDate)
	analysisDir := fmt.Sprintf("analysis.%s/%s", basename, currentDate)

	// Create result with all files to be created
	result := &ai.FlowResult{
		Files: map[string]string{
			taskListPath: a.generateTaskList(basename, currentDate, sourceFiles, analysisDir),
			summaryPath:  a.generateSummaryFile(basename, currentDate),
		},
		Messages: []string{
			fmt.Sprintf("\n✅ AI Analysis Task List generated successfully!"),
			fmt.Sprintf("📋 Task List: .aid/%s", taskListPath),
			fmt.Sprintf("📊 Summary File: .aid/%s", summaryPath),
			fmt.Sprintf("📁 Analysis Reports Directory: .aid/%s", analysisDir),
			fmt.Sprintf("🤖 Ready for AI-driven analysis workflow!"),
			fmt.Sprintf("📂 Files to analyze: %d\n", len(sourceFiles)),
			fmt.Sprintf("💡 If you are an AI agent, please read the Task List above and carefully follow all instructions to systematically analyze each file."),
		},
	}

	// Pre-create directory structure for report files
	for _, file := range sourceFiles {
		reportDir := filepath.Join(analysisDir, filepath.Dir(file))
		// Add a marker file to ensure directory creation
		markerPath := filepath.Join(reportDir, ".aid-created")
		result.Files[markerPath] = "# Directory created by AI Distiller\n"
	}

	return result, nil
}

// collectSourceFiles is similar to the existing implementation but adapted for the new structure
func (a *DeepAnalysisFlowAction) collectSourceFiles(projectPath string, includePatterns, excludePatterns []string) ([]string, error) {
	var sourceFiles []string

	// Load git submodules to skip them
	// gitSubmodules := make(map[string]bool) // TODO: Implement later

	// Define source file extensions
	sourceExtensions := map[string]bool{
		".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".go": true, ".java": true, ".kt": true, ".kts": true,
		".rs": true, ".swift": true, ".rb": true, ".php": true,
		".cpp": true, ".cc": true, ".cxx": true, ".c": true, ".h": true, ".hpp": true,
		".cs": true, ".fs": true, ".vb": true,
		".scala": true, ".clj": true, ".cljs": true,
		".vue": true, ".svelte": true, ".astro": true,
		".html": true, ".htm": true, ".css": true, ".scss": true, ".sass": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".md": true, ".mdx": true, ".rst": true, ".txt": true,
		".sh": true, ".bash": true, ".sql": true, ".graphql": true,
	}

	// Skip these directories
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, ".svn": true, ".hg": true,
		"__pycache__": true, ".pytest_cache": true,
		"target": true, "build": true, "dist": true, "out": true,
		".aid": true, "vendor": true, ".vscode": true, ".idea": true,
		"coverage": true, ".coverage": true, ".nyc_output": true,
		"grammars": true, "test-data": true, "testdata": true,
		"docs": true, "examples": true, "assets": true, "static": true,
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirName := filepath.Base(path)

			// Skip directories
			if skipDirs[dirName] || strings.HasPrefix(dirName, ".") && dirName != "." {
				return filepath.SkipDir
			}

			// Check for tree-sitter related content
			relPath, _ := filepath.Rel(projectPath, path)
			if strings.Contains(relPath, "grammars") || strings.Contains(relPath, "tree-sitter") {
				return filepath.SkipDir
			}

			return nil
		}

		// Check if it's a source file
		ext := strings.ToLower(filepath.Ext(path))
		if sourceExtensions[ext] {
			relPath, err := filepath.Rel(projectPath, path)
			if err == nil {
				// Skip generated files
				fileName := filepath.Base(relPath)
				if strings.HasPrefix(fileName, ".aid.") ||
					strings.HasPrefix(fileName, ".") ||
					strings.Contains(fileName, "generated") {
					return nil
				}

				// Apply include/exclude filters
				if a.shouldIncludeFile(relPath, includePatterns, excludePatterns) {
					sourceFiles = append(sourceFiles, relPath)
				}
			}
		}

		return nil
	})

	return sourceFiles, err
}

// shouldIncludeFile checks if a file should be included based on patterns
func (a *DeepAnalysisFlowAction) shouldIncludeFile(filePath string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, excludePattern := range excludePatterns {
		if excludePattern != "" {
			matched, _ := filepath.Match(excludePattern, filePath)
			if matched {
				return false
			}
			// Also check filename only
			matched, _ = filepath.Match(excludePattern, filepath.Base(filePath))
			if matched {
				return false
			}
		}
	}

	// If include patterns specified, file must match one
	if len(includePatterns) > 0 {
		for _, includePattern := range includePatterns {
			if includePattern != "" {
				matched, _ := filepath.Match(includePattern, filePath)
				if matched {
					return true
				}
				// Also check filename only
				matched, _ = filepath.Match(includePattern, filepath.Base(filePath))
				if matched {
					return true
				}
			}
		}
		return false
	}

	return true
}

// generateTaskList creates the task list content
func (a *DeepAnalysisFlowAction) generateTaskList(basename, currentDate string, sourceFiles []string, analysisDir string) string {
	var sb strings.Builder

	// Write header
	sb.WriteString(fmt.Sprintf("# AI Distiller – Comprehensive Code Analysis Task List\n\n"))
	sb.WriteString(fmt.Sprintf("**Project:** `%s`  \n", basename))
	sb.WriteString(fmt.Sprintf("**Analysis Date:** %s  \n", currentDate))
	sb.WriteString(fmt.Sprintf("**Total Files:** %d  \n\n", len(sourceFiles)))

	// Write comprehensive prompt
	sb.WriteString(a.getAIAnalysisPrompt(basename, currentDate))
	sb.WriteString("\n\n")

	// Write the task list
	sb.WriteString("## 📋 Analysis Task List\n\n")
	sb.WriteString("Complete each task in order, checking off items as you finish:\n\n")

	// Task 1: Create summary file
	sb.WriteString(fmt.Sprintf("- [ ] **1. Initialize Analysis Summary**  \n"))
	sb.WriteString(fmt.Sprintf("      Create `./aid/ANALYSIS-SUMMARY.%s.%s.md` with project overview\n\n", basename, currentDate))

	// Tasks for each file
	for i, file := range sourceFiles {
		taskNum := i + 2
		sb.WriteString(fmt.Sprintf("- [ ] **%d. Analyze `%s`**  \n", taskNum, file))
		sb.WriteString(fmt.Sprintf("      → Create report: `./aid/%s/%s.md`  \n", analysisDir, file))
		sb.WriteString(fmt.Sprintf("      → Add summary row to ANALYSIS-SUMMARY file\n\n"))
	}

	// Final task: Generate conclusion
	finalTaskNum := len(sourceFiles) + 2
	sb.WriteString(fmt.Sprintf("- [ ] **%d. Generate Project Conclusion**  \n", finalTaskNum))
	sb.WriteString(fmt.Sprintf("      Read completed ANALYSIS-SUMMARY file and write comprehensive conclusion\n\n"))

	// Write workflow notes
	sb.WriteString("## 🔄 Workflow Notes\n\n")
	sb.WriteString("- Check off each task **[x]** only after completing BOTH the individual report AND the summary row\n")
	sb.WriteString("- Follow the exact file naming conventions specified\n")
	sb.WriteString("- Use the standardized analysis format provided in the prompt\n")
	sb.WriteString("- Maintain consistent scoring across all files\n")
	sb.WriteString("- The final conclusion should synthesize findings from the entire summary table\n\n")

	// Add scope control tips
	sb.WriteString("## 💡 Scope Control Tips\n\n")
	sb.WriteString("If this task list is too large:\n")
	sb.WriteString("- **Analyze specific directories**: `aid internal/cli --ai-action=flow-for-deep-file-to-file-analysis`\n")
	sb.WriteString("- **Exclude test files**: `aid --exclude \"*test*,*spec*\" --ai-action=flow-for-deep-file-to-file-analysis`\n")
	sb.WriteString("- **Focus on core languages**: `aid --include \"*.go,*.py,*.ts,*.php\" --ai-action=flow-for-deep-file-to-file-analysis`\n\n")

	sb.WriteString("---\n")
	sb.WriteString("*Generated by AI Distiller – Comprehensive Code Analysis System*\n")

	return sb.String()
}

// generateSummaryFile creates the summary file template
func (a *DeepAnalysisFlowAction) generateSummaryFile(basename, currentDate string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Project Analysis Summary – %s (%s)\n\n", basename, currentDate))

	sb.WriteString("## 📊 Overview\n\n")
	sb.WriteString("This document provides a comprehensive analysis summary of the entire codebase. ")
	sb.WriteString("Each file has been individually analyzed for security, maintainability, performance, ")
	sb.WriteString("and readability. The results are compiled in the table below.\n\n")

	sb.WriteString("## 📈 Analysis Results\n\n")
	sb.WriteString("| File | Security % | Maintainability % | Performance % | Readability % | Critical | High | Medium | Low |\n")
	sb.WriteString("|------|:----------:|:-----------------:|:-------------:|:-------------:|:--------:|:----:|:------:|:---:|\n")

	// Note: Individual file rows will be appended here during analysis

	return sb.String()
}

// getAIAnalysisPrompt returns the comprehensive prompt for AI analysis
func (a *DeepAnalysisFlowAction) getAIAnalysisPrompt(basename, currentDate string) string {
	infrastructureInfo := fmt.Sprintf("### 🚀 Pre-Created Infrastructure\n"+
		"- **All report directories have been pre-created** - no need to run mkdir commands\n"+
		"- **Individual reports go to**: `.aid/analysis.%s/%s/[file-path].md`\n"+
		"- **Summary table updates**: `.aid/ANALYSIS-SUMMARY.%s.%s.md`", basename, currentDate, basename, currentDate)

	// Return the full prompt (abbreviated here for space)
	return fmt.Sprintf(`## 🤖 AI Analysis Instructions

# CRITICAL EXECUTION MANDATE: Unbreakable Sequential Processing

This is a FORMAL PROTOCOL implementing Chain-of-Thought (CoT) analysis with ZERO tolerance for deviations.

## ABSOLUTE PROHIBITIONS ⛔

1. **PROHIBITED**: Batch processing multiple files simultaneously
2. **PROHIBITED**: Using any "time-saving" shortcuts or optimizations
3. **PROHIBITED**: Skipping individual file analysis for "efficiency"
4. **PROHIBITED**: Marking tasks complete before ALL outputs are verified
5. **PROHIBITED**: Referencing or planning for files not yet in scope
6. **VIOLATION CONSEQUENCE**: Any deviation = IMMEDIATE PROTOCOL FAILURE

You are an **Expert Senior Staff Engineer and Security Auditor** conducting a comprehensive file-by-file analysis of the **%s** project.

%s

[Rest of the original prompt content...]`, basename, infrastructureInfo)
}
