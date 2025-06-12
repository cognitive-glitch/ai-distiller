package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"fmt"
)

// TestCLIIntegration tests the CLI with real commands
func TestCLIIntegration(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../ai-distiller", "../cmd/ai-distiller/main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v\n%s", err, output)
	}

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output string)
	}{
		{
			name: "distill_basic_class_json",
			args: []string{"input/basic_class.py", "--format", "json"},
			validate: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
				
				// Check structure
				if lang, ok := data["language"].(string); !ok || lang != "python" {
					t.Errorf("Expected language to be python, got %v", data["language"])
				}
				
				if path, ok := data["path"].(string); !ok || !strings.Contains(path, "basic_class.py") {
					t.Errorf("Expected path to contain basic_class.py, got %v", data["path"])
				}
			},
		},
		{
			name: "distill_no_private_markdown",
			args: []string{"input/basic_class.py", "--format", "markdown"}, // Default is no private
			validate: func(t *testing.T, output string) {
				// Should not contain private method
				if strings.Contains(output, "_calculate_id") {
					t.Error("Output should not contain private method _calculate_id")
				}
				
				// Should contain public methods
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain public method get_info")
				}
			},
		},
		{
			name: "distill_no_implementation_json",
			args: []string{"input/basic_class.py", "--format", "json"},
			validate: func(t *testing.T, output string) {
				// Check that implementations are empty
				if strings.Contains(output, `"implementation":`) && strings.Contains(output, "return f") {
					t.Error("Output should not contain implementation details")
				}
			},
		},
		{
			name: "distill_minimal_markdown",
			args: []string{"input/basic_class.py", "--format", "markdown", "--compact"},
			validate: func(t *testing.T, output string) {
				// Should only have structure
				if strings.Contains(output, "```") {
					t.Error("Minimal output should not contain code blocks")
				}
				
				// Should still have class and function names
				if !strings.Contains(output, "Person") {
					t.Error("Minimal output should contain class name")
				}
			},
		},
		{
			name: "distill_complex_imports",
			args: []string{"input/complex_imports.py", "--format", "markdown", "--imports"},
			validate: func(t *testing.T, output string) {
				// Check various import styles
				imports := []string{
					"import `os`",
					"from `typing` import",
					"import `numpy`",
					"from `pandas` import `DataFrame` as `DF`",
				}
				
				for _, imp := range imports {
					if !strings.Contains(output, imp) {
						t.Errorf("Output should contain import: %s", imp)
					}
				}
			},
		},
		{
			name: "distill_directory_multiple_files",
			args: []string{"input/", "--format", "jsonl"},
			validate: func(t *testing.T, output string) {
				// JSONL should have one object per line
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 4 { // We have at least 5 test files
					t.Errorf("Expected at least 4 files, got %d lines", len(lines))
				}
				
				// Each line should be valid JSON
				for i, line := range lines {
					var data map[string]interface{}
					if err := json.Unmarshal([]byte(line), &data); err != nil {
						t.Errorf("Line %d is not valid JSON: %v", i+1, err)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../ai-distiller", tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\n%s", err, output)
			}
			
			tt.validate(t, string(output))
		})
	}
}

// TestOptionsInteraction tests how different options interact
func TestOptionsInteraction(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldHave  []string
		shouldNotHave []string
	}{
		{
			name: "no_private_with_no_implementation",
			args: []string{"distill", "input/basic_class.py", "--format", "markdown", "--no-private", "--no-implementation"},
			shouldHave: []string{
				"get_info",
				"__init__",
			},
			shouldNotHave: []string{
				"_calculate_id",
				"```", // no code blocks
			},
		},
		{
			name: "no_imports_option",
			args: []string{"distill", "input/complex_imports.py", "--format", "markdown", "--no-imports"},
			shouldHave: []string{
				"Container",
			},
			shouldNotHave: []string{
				"📥 **Import**",
				"import `os`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../ai-distiller", tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\n%s", err, output)
			}
			
			outputStr := string(output)
			
			for _, should := range tt.shouldHave {
				if !strings.Contains(outputStr, should) {
					t.Errorf("Output should contain %q", should)
				}
			}
			
			for _, shouldNot := range tt.shouldNotHave {
				if strings.Contains(outputStr, shouldNot) {
					t.Errorf("Output should NOT contain %q", shouldNot)
				}
			}
		})
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "nonexistent_file",
			args:        []string{"distill", "nonexistent.py"},
			shouldError: true,
			errorMsg:    "failed to read file",
		},
		{
			name:        "invalid_format",
			args:        []string{"distill", "input/basic_class.py", "--format", "invalid"},
			shouldError: true,
			errorMsg:    "formatter not found",
		},
		{
			name:        "directory_without_python_files",
			args:        []string{"distill", "../expected/"}, // empty directory
			shouldError: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../ai-distiller", tt.args...)
			output, err := cmd.CombinedOutput()
			
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but command succeeded")
				}
				if tt.errorMsg != "" && !strings.Contains(strings.ToLower(string(output)), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %s", tt.errorMsg, output)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v\n%s", err, output)
				}
			}
		})
	}
}

// TestPerformance tests processing speed
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a large Python file
	largeFile := "input/large_test.py"
	content := `# Large test file
import os
import sys

"""
This is a large test file for performance testing.
"""

`
	// Generate many classes and functions
	for i := 0; i < 100; i++ {
		content += fmt.Sprintf(`
class TestClass%d:
    """Test class %d"""
    
    def __init__(self):
        self.value = %d
    
    def method1(self, x: int) -> int:
        """Method 1"""
        return x + self.value
    
    def method2(self, y: str) -> str:
        """Method 2"""
        return y * self.value
    
    def _private_method(self):
        """Private method"""
        pass

`, i, i, i)
	}

	if err := os.WriteFile(largeFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	defer os.Remove(largeFile)

	// Time the processing
	start := time.Now()
	cmd := exec.Command("../ai-distiller", "distill", largeFile, "--format", "json")
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to process large file: %v\n%s", err, output)
	}

	// Check performance
	if duration > 5*time.Second {
		t.Errorf("Processing took too long: %v", duration)
	}

	// Verify output
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	// Should have 100 classes
	if stats, ok := data["stats"].(map[string]interface{}); ok {
		if classes, ok := stats["class"].(float64); ok {
			if int(classes) != 100 {
				t.Errorf("Expected 100 classes, got %d", int(classes))
			}
		}
	}
}

// BenchmarkDistiller benchmarks the distiller
func BenchmarkDistiller(b *testing.B) {
	cmd := exec.Command("go", "build", "-o", "../ai-distiller", "../cmd/ai-distiller/main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build CLI: %v\n%s", err, output)
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("../ai-distiller", "distill", "input/basic_class.py", "--format", "json")
		if _, err := cmd.CombinedOutput(); err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}