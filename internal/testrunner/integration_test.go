package testrunner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/janreges/ai-distiller/internal/testrunner"
)

func TestIntegration(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}
	
	// Find the project root
	projectRoot := findProjectRoot()
	if projectRoot == "" {
		t.Fatal("could not find project root")
	}
	
	// Check if aid binary exists
	aidBinary := filepath.Join(projectRoot, "aid")
	testDataDir := filepath.Join(projectRoot, "testdata")
	
	var runner *testrunner.Runner
	if _, err := os.Stat(aidBinary); os.IsNotExist(err) {
		// Use go run mode
		t.Log("Using 'go run' mode - aid binary not found")
		runner = testrunner.NewWithGoRun(testDataDir, projectRoot)
	} else {
		// Use binary mode
		t.Log("Using compiled aid binary")
		runner = testrunner.New(testDataDir, aidBinary)
	}
	
	// Run all tests
	runner.RunTests(t)
}

// findProjectRoot looks for go.mod to determine project root
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}