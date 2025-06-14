// +build ignore

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"fmt"
)

// TestAIDCLIIntegration tests the aid CLI with real commands
func TestAIDCLIIntegration(t *testing.T) {
	// Build the aid CLI first
	cmd := exec.Command("go", "build", "-o", "../aid", "../cmd/aid/main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("Skipping aid CLI tests (build failed): %v\n%s", err, output)
	}

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output string)
	}{
		{
			name: "distill_basic_class_markdown",
			args: []string{"input/basic_class.py", "--format", "md"},
			validate: func(t *testing.T, output string) {
				// Should contain class and methods
				if !strings.Contains(output, "Person") {
					t.Error("Output should contain Person class")
				}
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain get_info method")
				}
			},
		},
		{
			name: "distill_no_private_markdown",
			args: []string{"input/basic_class.py", "--format", "md", "--private", "0", "--protected", "0"},
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
			args: []string{"input/basic_class.py", "--format", "json-structured", "--implementation", "0"},
			validate: func(t *testing.T, output string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(output), &data); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
				
				// Check that implementations are not included
				jsonStr := string(output)
				if strings.Contains(jsonStr, `"implementation"`) {
					t.Error("Output should not contain implementation field")
				}
			},
		},
		{
			name: "distill_minimal_markdown",
			args: []string{"input/basic_class.py", "--format", "md", "--comments", "0", "--implementation", "0"},
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
			args: []string{"input/complex_imports.py", "--format", "md"},
			validate: func(t *testing.T, output string) {
				// Check various import styles
				imports := []string{
					"import `os`",
					"from `typing` import",
					"import `numpy`",
				}
				
				for _, imp := range imports {
					if !strings.Contains(output, imp) {
						t.Errorf("Output should contain import: %s", imp)
					}
				}
			},
		},
		{
			name: "distill_directory_jsonl",
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
		// New tests for individual filtering flags
		{
			name: "distill_with_new_flags_public_only",
			args: []string{"input/basic_class.py", "--format", "text", "--public", "1", "--private", "0", "--protected", "0"},
			validate: func(t *testing.T, output string) {
				// Should contain public method
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain public method get_info")
				}
				// Should not contain private method
				if strings.Contains(output, "_calculate_id") {
					t.Error("Output should not contain private method _calculate_id")
				}
			},
		},
		{
			name: "distill_with_new_flags_all_visibility",
			args: []string{"input/basic_class.py", "--format", "text", "--public", "1", "--private", "1", "--protected", "1"},
			validate: func(t *testing.T, output string) {
				// Should contain both public and private methods
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain public method get_info")
				}
				if !strings.Contains(output, "_calculate_id") {
					t.Error("Output should contain private method _calculate_id")
				}
			},
		},
		{
			name: "distill_with_include_only",
			args: []string{"input/basic_class.py", "--format", "text", "--include-only", "public,imports"},
			validate: func(t *testing.T, output string) {
				// Should contain imports
				if !strings.Contains(output, "datetime") {
					t.Error("Output should contain imports")
				}
				// Should contain public members only
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain public method")
				}
				// Should not contain private
				if strings.Contains(output, "_calculate_id") {
					t.Error("Output should not contain private method")
				}
			},
		},
		{
			name: "distill_with_exclude_items",
			args: []string{"input/basic_class.py", "--format", "text", "--exclude-items", "private,comments"},
			validate: func(t *testing.T, output string) {
				// Should not contain private
				if strings.Contains(output, "_calculate_id") {
					t.Error("Output should not contain private method")
				}
				// Should contain public
				if !strings.Contains(output, "get_info") {
					t.Error("Output should contain public method")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use --stdout to get output directly
			args := append([]string{"--stdout"}, tt.args...)
			cmd := exec.Command("../aid", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\n%s", err, output)
			}
			
			tt.validate(t, string(output))
		})
	}
}

// TestOptionsInteraction tests how different strip options interact
func TestOptionsInteraction(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldHave  []string
		shouldNotHave []string
	}{
		{
			name: "strip_multiple_options",
			args: []string{"--stdout", "input/basic_class.py", "--format", "md", "--private", "0", "--implementation", "0"},
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
			name: "strip_imports",
			args: []string{"--stdout", "input/complex_imports.py", "--format", "md", "--imports", "0"},
			shouldHave: []string{
				"Container",
			},
			shouldNotHave: []string{
				"📥 **Import**",
				"import `os`",
			},
		},
		// Test new group flags
		{
			name: "new_group_include_only",
			args: []string{"--stdout", "input/basic_class.py", "--format", "text", "--include-only", "public"},
			shouldHave: []string{
				"get_info",
				"Person",
			},
			shouldNotHave: []string{
				"_calculate_id",
				"datetime", // imports not included
			},
		},
		{
			name: "new_group_exclude_items", 
			args: []string{"--stdout", "input/basic_class.py", "--format", "text", "--exclude-items", "implementation,comments"},
			shouldHave: []string{
				"get_info",
				"_calculate_id",
				"datetime", // imports included
			},
			shouldNotHave: []string{
				"self.age", // implementation detail
			},
		},
	}

	// Skip if aid is not built
	if _, err := os.Stat("../aid"); os.IsNotExist(err) {
		t.Skip("aid CLI not built, skipping interaction tests")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../aid", tt.args...)
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
			args:        []string{"nonexistent.py"},
			shouldError: true,
			errorMsg:    "does not exist",
		},
		{
			name:        "invalid_format",
			args:        []string{"input/basic_class.py", "--format", "invalid"},
			shouldError: true,
			errorMsg:    "invalid output format",
		},
	}

	// Skip if aid is not built
	if _, err := os.Stat("../aid"); os.IsNotExist(err) {
		t.Skip("aid CLI not built, skipping error tests")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../aid", tt.args...)
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