package testrunner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestCase represents a single test case discovered from the filesystem
type TestCase struct {
	Language      string
	ScenarioName  string
	SourceFile    string
	ExpectedFile  string
	Flags         []string
}

// Runner is the main test runner that discovers and executes tests
type Runner struct {
	testDataDir string
	aidBinary   string
	updateMode  bool
	useGoRun    bool   // Whether to use "go run" instead of binary
	projectRoot string // Project root for go run mode
}

// New creates a new test runner
func New(testDataDir, aidBinary string) *Runner {
	return &Runner{
		testDataDir: testDataDir,
		aidBinary:   aidBinary,
		updateMode:  os.Getenv("UPDATE_EXPECTED") == "true",
	}
}

// NewWithGoRun creates a test runner that uses "go run" instead of a binary
func NewWithGoRun(testDataDir, projectRoot string) *Runner {
	return &Runner{
		testDataDir: testDataDir,
		projectRoot: projectRoot,
		updateMode:  os.Getenv("UPDATE_EXPECTED") == "true",
		useGoRun:    true,
	}
}

// parseFlags extracts flags from expected filename based on parameter encoding
// e.g., "Test5-Complex.implementation=0.comments=0.public=1.expected" -> ["--implementation=0", "--comments=0", "--public=1"]
func parseParametersFromFilename(filename string) []string {
	// Remove the .expected extension
	name := strings.TrimSuffix(filename, ".expected")
	
	// Find first dot to separate test name from parameters
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		// No parameters, use defaults
		return []string{}
	}
	
	var flags []string
	// Start from second part (skip test name)
	for i := 1; i < len(parts); i++ {
		param := parts[i]
		// Check if it's a parameter=value format
		if strings.Contains(param, "=") {
			flags = append(flags, "--"+param)
		}
	}
	
	return flags
}


// DiscoverTests walks the testdata directory and finds all test cases
func (r *Runner) DiscoverTests() ([]TestCase, error) {
	var tests []TestCase
	
	// Walk through each language directory
	entries, err := os.ReadDir(r.testDataDir)
	if err != nil {
		return nil, fmt.Errorf("reading testdata dir: %w", err)
	}
	
	for _, langEntry := range entries {
		if !langEntry.IsDir() {
			continue
		}
		
		language := langEntry.Name()
		langDir := filepath.Join(r.testDataDir, language)
		
		// Walk through each scenario directory
		scenarios, err := os.ReadDir(langDir)
		if err != nil {
			return nil, fmt.Errorf("reading language dir %s: %w", language, err)
		}
		
		for _, scenarioEntry := range scenarios {
			if !scenarioEntry.IsDir() {
				continue
			}
			
			scenarioName := scenarioEntry.Name()
			scenarioDir := filepath.Join(langDir, scenarioName)
			
			// Find source file
			sourceFile, err := findSourceFile(scenarioDir, language)
			if err != nil {
				continue // Skip scenarios without source files
			}
			
			// Find expected files
			expectedDir := filepath.Join(scenarioDir, "expected")
			if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
				// If no expected directory, assume default.expected
				tests = append(tests, TestCase{
					Language:     language,
					ScenarioName: scenarioName,
					SourceFile:   sourceFile,
					ExpectedFile: filepath.Join(scenarioDir, "default.expected"),
					Flags:        []string{},
				})
				continue
			}
			
			// Read all expected files
			expectedFiles, err := os.ReadDir(expectedDir)
			if err != nil {
				return nil, fmt.Errorf("reading expected dir for %s/%s: %w", language, scenarioName, err)
			}
			
			for _, expFile := range expectedFiles {
				if !strings.HasSuffix(expFile.Name(), ".expected") {
					continue
				}
				
				// Use new parameter-based parsing
				flags := parseParametersFromFilename(expFile.Name())
				tests = append(tests, TestCase{
					Language:     language,
					ScenarioName: scenarioName,
					SourceFile:   sourceFile,
					ExpectedFile: filepath.Join(expectedDir, expFile.Name()),
					Flags:        flags,
				})
			}
		}
	}
	
	return tests, nil
}

// findSourceFile finds the source file for a given language in a scenario directory
func findSourceFile(dir, language string) (string, error) {
	extensions := map[string][]string{
		"go":         {".go"},
		"python":     {".py"},
		"typescript": {".ts", ".tsx"},
		"javascript": {".js", ".jsx"},
		"java":       {".java"},
		"csharp":     {".cs"},
		"cpp":        {".cpp", ".cc", ".cxx", ".hpp", ".h"},
		"rust":       {".rs"},
		"swift":      {".swift"},
		"kotlin":     {".kt"},
		"php":        {".php"},
		"ruby":       {".rb"},
	}
	
	exts, ok := extensions[language]
	if !ok {
		return "", fmt.Errorf("unknown language: %s", language)
	}
	
	// Look for source.{ext} first
	for _, ext := range exts {
		sourcePath := filepath.Join(dir, "source"+ext)
		if _, err := os.Stat(sourcePath); err == nil {
			return sourcePath, nil
		}
	}
	
	// If not found, look for any file with the right extension
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		for _, ext := range exts {
			if strings.HasSuffix(entry.Name(), ext) {
				return filepath.Join(dir, entry.Name()), nil
			}
		}
	}
	
	return "", fmt.Errorf("no source file found for language %s in %s", language, dir)
}

// RunTest executes a single test case
func (r *Runner) RunTest(tc TestCase) error {
	// Build command
	var cmd *exec.Cmd
	
	if r.useGoRun {
		// Use go run mode
		goArgs := []string{"run", "./cmd/aid"}
		goArgs = append(goArgs, tc.SourceFile)
		goArgs = append(goArgs, tc.Flags...)
		goArgs = append(goArgs, "--format", "text", "--stdout")
		
		cmd = exec.Command("go", goArgs...)
		cmd.Dir = r.projectRoot
	} else {
		// Use binary mode
		args := append([]string{tc.SourceFile}, tc.Flags...)
		args = append(args, "--format", "text", "--stdout")
		
		cmd = exec.Command(r.aidBinary, args...)
	}
	
	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running aid: %w\nstderr: %s", err, stderr.String())
	}
	
	// Get actual output
	actual := stdout.Bytes()
	
	if r.updateMode {
		// Update expected file
		return os.WriteFile(tc.ExpectedFile, actual, 0644)
	}
	
	// Read expected output
	expected, err := os.ReadFile(tc.ExpectedFile)
	if err != nil {
		if os.IsNotExist(err) && r.updateMode {
			// Create the expected file in update mode
			return os.WriteFile(tc.ExpectedFile, actual, 0644)
		}
		return fmt.Errorf("reading expected file: %w", err)
	}
	
	// Compare outputs
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("output mismatch:\nEXPECTED:\n%s\nACTUAL:\n%s", expected, actual)
	}
	
	return nil
}

// RunTests executes all discovered tests
func (r *Runner) RunTests(t *testing.T) {
	tests, err := r.DiscoverTests()
	if err != nil {
		t.Fatalf("discovering tests: %v", err)
	}
	
	if len(tests) == 0 {
		t.Log("No tests discovered")
		return
	}
	
	for _, tc := range tests {
		testName := fmt.Sprintf("%s/%s/%s", tc.Language, tc.ScenarioName, filepath.Base(tc.ExpectedFile))
		t.Run(testName, func(t *testing.T) {
			if err := r.RunTest(tc); err != nil {
				t.Errorf("test failed: %v", err)
			}
		})
	}
}

// GenerateExpectedFiles is a helper to generate expected files for existing source files
func (r *Runner) GenerateExpectedFiles(language, scenario string, flagSets [][]string) error {
	scenarioDir := filepath.Join(r.testDataDir, language, scenario)
	sourceFile, err := findSourceFile(scenarioDir, language)
	if err != nil {
		return fmt.Errorf("finding source file: %w", err)
	}
	
	expectedDir := filepath.Join(scenarioDir, "expected")
	if err := os.MkdirAll(expectedDir, 0755); err != nil {
		return fmt.Errorf("creating expected dir: %w", err)
	}
	
	for _, flags := range flagSets {
		// Determine filename from flags
		filename := r.flagsToFilename(flags)
		
		// Run aid with these flags
		args := append([]string{sourceFile}, flags...)
		args = append(args, "--format", "text", "--stdout")
		
		cmd := exec.Command(r.aidBinary, args...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("running aid with flags %v: %w", flags, err)
		}
		
		// Write expected file
		expectedFile := filepath.Join(expectedDir, filename)
		if err := os.WriteFile(expectedFile, output, 0644); err != nil {
			return fmt.Errorf("writing expected file %s: %w", expectedFile, err)
		}
	}
	
	return nil
}

// flagsToFilename converts a set of flags to an expected filename
func (r *Runner) flagsToFilename(flags []string) string {
	if len(flags) == 0 {
		return "default.expected"
	}
	
	// Convert flags to parameter format
	var params []string
	for _, flag := range flags {
		// Remove -- prefix
		param := strings.TrimPrefix(flag, "--")
		params = append(params, param)
	}
	
	// Sort for consistent naming
	sort.Strings(params)
	
	return "test." + strings.Join(params, ".") + ".expected"
}

// Helper functions

func equalFlags(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	aMap := make(map[string]bool)
	for _, flag := range a {
		aMap[flag] = true
	}
	
	for _, flag := range b {
		if !aMap[flag] {
			return false
		}
	}
	
	return true
}

func containsAll(flags, subset []string) bool {
	flagMap := make(map[string]bool)
	for _, flag := range flags {
		flagMap[flag] = true
	}
	
	for _, flag := range subset {
		if !flagMap[flag] {
			return false
		}
	}
	
	return true
}