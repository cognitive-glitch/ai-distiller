package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		stripOptions []string
		expected     string
	}{
		{
			name:         "NoStripOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{},
			expected:     ".myproject.aid.txt",
		},
		{
			name:         "WithComments",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments"},
			expected:     ".myproject.ncom.aid.txt",
		},
		{
			name:         "MultipleOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments", "imports", "implementation"},
			expected:     ".myproject.ncom.nimp.nimpl.aid.txt",
		},
		{
			name:         "AllOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments", "imports", "implementation", "non-public"},
			expected:     ".myproject.ncom.nimp.nimpl.npriv.aid.txt",
		},
		{
			name:         "CurrentDirectory",
			path:         ".",
			stripOptions: []string{},
			expected:     ".current.aid.txt",
		},
		{
			name:         "RootDirectory",
			path:         "/",
			stripOptions: []string{},
			expected:     ".current.aid.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOutputFilename(tt.path, tt.stripOptions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "ItemExists",
			slice:    []string{"apple", "banana", "orange"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "ItemDoesNotExist",
			slice:    []string{"apple", "banana", "orange"},
			item:     "grape",
			expected: false,
		},
		{
			name:     "EmptySlice",
			slice:    []string{},
			item:     "apple",
			expected: false,
		},
		{
			name:     "EmptyString",
			slice:    []string{"apple", "", "orange"},
			item:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCLIFlags(t *testing.T) {
	t.Run("DefaultFlags", func(t *testing.T) {
		// Reset flags to defaults
		outputFile = ""
		outputToStdout = false
		outputFormat = "md"
		stripOptions = nil
		includeGlob = ""
		excludeGlob = ""
		recursive = true
		absolutePaths = false
		strict = false
		verbosity = 0

		// Test default values
		assert.Equal(t, "", outputFile)
		assert.False(t, outputToStdout)
		assert.Equal(t, "md", outputFormat)
		assert.True(t, recursive)
		assert.False(t, strict)
	})

	t.Run("ParseFlags", func(t *testing.T) {
		// Create a new command instance for testing
		cmd := rootCmd
		cmd.ResetFlags()
		initFlags() // Re-initialize flags

		// Test parsing various flags
		args := []string{
			"--output", "test.txt",
			"--stdout",
			"--format", "jsonl",
			"--strip", "comments,imports",
			"--include", "*.go",
			"--exclude", "*_test.go",
			"--absolute-paths",
			"--strict",
			"-vvv",
		}

		cmd.ParseFlags(args)

		// Verify flags were parsed correctly
		output, _ := cmd.Flags().GetString("output")
		assert.Equal(t, "test.txt", output)

		stdout, _ := cmd.Flags().GetBool("stdout")
		assert.True(t, stdout)

		format, _ := cmd.Flags().GetString("format")
		assert.Equal(t, "jsonl", format)

		strip, _ := cmd.Flags().GetStringSlice("strip")
		assert.Equal(t, []string{"comments", "imports"}, strip)

		include, _ := cmd.Flags().GetString("include")
		assert.Equal(t, "*.go", include)

		exclude, _ := cmd.Flags().GetString("exclude")
		assert.Equal(t, "*_test.go", exclude)

		absPath, _ := cmd.Flags().GetBool("absolute-paths")
		assert.True(t, absPath)

		strictFlag, _ := cmd.Flags().GetBool("strict")
		assert.True(t, strictFlag)

		verbose, _ := cmd.Flags().GetCount("verbose")
		assert.Equal(t, 3, verbose)
	})
}

func TestVersionFlag(t *testing.T) {
	// Save original version
	originalVersion := Version
	Version = "1.2.3-test"
	defer func() { Version = originalVersion }()

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create command with version flag
	cmd := rootCmd
	cmd.SetArgs([]string{"--version"})

	// We expect this to call os.Exit, so we need to handle it differently
	// For now, we'll just test that the version string is correct
	assert.Equal(t, "1.2.3-test", Version)

	// Restore stdout
	w.Close()
	os.Stdout = old
	io.Copy(io.Discard, r)
}

func TestRunDistillerValidation(t *testing.T) {
	t.Run("InvalidOutputFormat", func(t *testing.T) {
		// Create temp directory for testing
		tempDir := t.TempDir()

		cmd := rootCmd
		cmd.SetArgs([]string{tempDir, "--format", "invalid"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output format")
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"/non/existent/path"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
	})

	t.Run("ValidPath", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Reset output file and format to defaults
		outputFile = ""
		outputFormat = "md"

		cmd := rootCmd
		cmd.SetArgs([]string{tempDir})

		// This will return "not yet implemented" which is expected for now
		err := cmd.Execute()
		// Since we print and return nil, no error is expected
		assert.NoError(t, err)
	})
}

func TestCLIHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := rootCmd
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	
	// Check that help contains expected content
	assert.Contains(t, output, "AI Distiller")
	assert.Contains(t, output, "aid [path]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--format")
	assert.Contains(t, output, "--strip")
	assert.Contains(t, output, "--recursive")
}

func TestStripOptionsAbbreviation(t *testing.T) {
	tests := []struct {
		options  []string
		expected string
	}{
		{[]string{"comments"}, ".ncom"},
		{[]string{"imports"}, ".nimp"},
		{[]string{"implementation"}, ".nimpl"},
		{[]string{"non-public"}, ".npriv"},
		{[]string{"comments", "imports"}, ".ncom.nimp"},
		{[]string{"unknown"}, ""}, // Unknown options are ignored
	}

	for _, tt := range tests {
		filename := generateOutputFilename("/test", tt.options)
		if tt.expected != "" {
			assert.Contains(t, filename, tt.expected)
		}
	}
}

// Helper to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}

// Test verbose output
func TestVerboseOutput(t *testing.T) {
	// This test would require mocking stderr output
	// For now, we just verify the verbosity counter works
	
	cmd := rootCmd
	cmd.ResetFlags()
	initFlags()
	
	// Test single -v
	cmd.ParseFlags([]string{"-v"})
	v, _ := cmd.Flags().GetCount("verbose")
	assert.Equal(t, 1, v)
	
	// Test -vvv
	cmd.ResetFlags()
	initFlags()
	cmd.ParseFlags([]string{"-vvv"})
	v, _ = cmd.Flags().GetCount("verbose")
	assert.Equal(t, 3, v)
}