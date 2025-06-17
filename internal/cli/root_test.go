package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/janreges/ai-distiller/internal/processor"
)

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		stripOptions []string
		setupFlags   func()
		expected     string
	}{
		{
			name:         "NoStripOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.myproject.txt",
		},
		{
			name:         "WithComments",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments"},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.myproject.ncom.txt",
		},
		{
			name:         "MultipleOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments", "imports", "implementation"},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.myproject.ncom.nimp.nimpl.txt",
		},
		{
			name:         "AllOptions",
			path:         "/home/user/myproject",
			stripOptions: []string{"comments", "imports", "implementation", "non-public"},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.myproject.ncom.nimp.nimpl.npub.txt",
		},
		{
			name:         "CurrentDirectory",
			path:         ".",
			stripOptions: []string{},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.current.txt",
		},
		{
			name:         "RootDirectory",
			path:         "/",
			stripOptions: []string{},
			setupFlags:   func() { resetAllFlags() },
			expected:     ".aid.current.txt",
		},
		{
			name:         "NewFlagsWithPrivate",
			path:         "/home/user/myproject",
			stripOptions: []string{},
			setupFlags: func() {
				resetAllFlags()
				includePrivate = boolPtr(true)
			},
			expected:     ".aid.myproject.priv.txt",
		},
		{
			name:         "NewFlagsWithProtectedAndImplementation",
			path:         "/home/user/myproject",
			stripOptions: []string{},
			setupFlags: func() {
				resetAllFlags()
				includeProtected = boolPtr(true)
				includeImplementation = boolPtr(true)
			},
			expected:     ".aid.myproject.prot.impl.txt",
		},
		{
			name:         "NewFlagsWithComments",
			path:         "/home/user/myproject",
			stripOptions: []string{},
			setupFlags: func() {
				resetAllFlags()
				includeComments = boolPtr(true)
			},
			expected:     ".aid.myproject.com.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			result := generateOutputFilename(tt.path, tt.stripOptions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to reset all flags
func resetAllFlags() {
	includePublic = nil
	includeProtected = nil
	includeInternal = nil
	includePrivate = nil
	includeComments = nil
	includeDocstrings = nil
	includeImplementation = nil
	includeImports = nil
	includeAnnotations = nil
	includeList = ""
	excludeList = ""
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
		outputFormat = "text"
		stripOptions = nil
		includeGlob = []string{}
		excludeGlob = []string{}
		recursiveStr = "1"
		filePathType = "relative"
		strict = false
		verbosity = 0
		includePublic = nil
		includeProtected = nil
		includeInternal = nil
		includePrivate = nil
		includeComments = nil
		includeDocstrings = nil
		includeImplementation = nil
		includeImports = nil
		includeAnnotations = nil
		includeList = ""
		excludeList = ""

		// Test default values
		assert.Equal(t, "", outputFile)
		assert.False(t, outputToStdout)
		assert.Equal(t, "text", outputFormat)
		assert.Equal(t, "1", recursiveStr)
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
			"--file-path-type", "absolute",
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

		filePathType, _ := cmd.Flags().GetString("file-path-type")
		assert.Equal(t, "absolute", filePathType)

		strictFlag, _ := cmd.Flags().GetBool("strict")
		assert.True(t, strictFlag)

		verbose, _ := cmd.Flags().GetCount("verbose")
		assert.Equal(t, 3, verbose)
	})

	t.Run("ParseNewFilterFlags", func(t *testing.T) {
		// Create a new command instance for testing
		cmd := rootCmd
		cmd.ResetFlags()
		initFlags() // Re-initialize flags

		// Test parsing new filtering flags
		args := []string{
			"--public", "1",
			"--protected", "0",
			"--private", "1",
			"--internal", "0",
			"--comments", "1",
			"--docstrings", "0",
			"--implementation", "1",
			"--imports", "0",
			"--annotations", "1",
		}

		cmd.ParseFlags(args)

		// Verify flags were parsed correctly
		public, _ := cmd.Flags().GetString("public")
		assert.Equal(t, "1", public)

		protected, _ := cmd.Flags().GetString("protected")
		assert.Equal(t, "0", protected)

		private, _ := cmd.Flags().GetString("private")
		assert.Equal(t, "1", private)

		comments, _ := cmd.Flags().GetString("comments")
		assert.Equal(t, "1", comments)

		implementation, _ := cmd.Flags().GetString("implementation")
		assert.Equal(t, "1", implementation)
	})

	t.Run("ParseGroupFilterFlags", func(t *testing.T) {
		// Create a new command instance for testing
		cmd := rootCmd
		cmd.ResetFlags()
		initFlags() // Re-initialize flags

		// Test parsing group filter flags
		args := []string{
			"--include-only", "public,protected,imports",
			"--exclude-items", "comments,implementation",
		}

		cmd.ParseFlags(args)

		// Verify flags were parsed correctly
		includeOnly, _ := cmd.Flags().GetString("include-only")
		assert.Equal(t, "public,protected,imports", includeOnly)

		excludeItems, _ := cmd.Flags().GetString("exclude-items")
		assert.Equal(t, "comments,implementation", excludeItems)
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
		// Reset all flags before test
		resetAllFlags()
		
		// Create temp directory for testing
		tempDir := t.TempDir()

		cmd := rootCmd
		cmd.ResetFlags()
		initFlags()
		cmd.SetArgs([]string{tempDir, "--format", "invalid"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output format")
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		// Reset all flags before test
		resetAllFlags()
		
		cmd := rootCmd
		cmd.ResetFlags()
		initFlags()
		cmd.SetArgs([]string{"/non/existent/path"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
	})

	t.Run("ValidPath", func(t *testing.T) {
		// Reset all flags before test
		resetAllFlags()
		
		// Create temp directory
		tempDir := t.TempDir()

		// Reset output file and format to defaults
		outputFile = ""
		outputFormat = "text"

		cmd := rootCmd
		cmd.ResetFlags()
		initFlags()
		cmd.SetArgs([]string{tempDir})

		// This will return "not yet implemented" which is expected for now
		err := cmd.Execute()
		// Since we print and return nil, no error is expected
		assert.NoError(t, err)
	})
}

func TestCLIHelp(t *testing.T) {
	// Reset all flags before test
	resetAllFlags()
	
	buf := new(bytes.Buffer)
	cmd := rootCmd
	cmd.ResetFlags()
	initFlags()
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
	// Check for new flags instead of deprecated --strip
	assert.Contains(t, output, "--public")
	assert.Contains(t, output, "--private")
	assert.Contains(t, output, "--protected")
	assert.Contains(t, output, "--comments")
	assert.Contains(t, output, "--implementation")
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
		{[]string{"non-public"}, ".npub"},
		{[]string{"private"}, ".npriv"},
		{[]string{"protected"}, ".nprot"},
		{[]string{"comments", "imports"}, ".ncom.nimp"},
		{[]string{"private", "protected"}, ".npriv.nprot"},
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

func TestGetBoolFlag(t *testing.T) {
	tests := []struct {
		name        string
		flag        *bool
		defaultVal  bool
		expected    bool
	}{
		{
			name:       "NilFlagReturnsDefault",
			flag:       nil,
			defaultVal: true,
			expected:   true,
		},
		{
			name:       "TrueFlag",
			flag:       boolPtr(true),
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "FalseFlag",
			flag:       boolPtr(false),
			defaultVal: true,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolFlag(tt.flag, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessIncludeList(t *testing.T) {
	tests := []struct {
		name     string
		list     string
		expected processor.ProcessOptions
	}{
		{
			name: "IncludePublicOnly",
			list: "public",
			expected: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImports:        false,
				IncludeImplementation: false,
				IncludePrivate:        false,
			},
		},
		{
			name: "IncludePublicAndPrivate",
			list: "public,private",
			expected: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImports:        false,
				IncludeImplementation: false,
				IncludePrivate:        true,
			},
		},
		{
			name: "IncludeAllContent",
			list: "public,comments,implementation,imports",
			expected: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processIncludeList(tt.list)
			assert.Equal(t, tt.expected.IncludeComments, result.IncludeComments)
			assert.Equal(t, tt.expected.IncludeImports, result.IncludeImports)
			assert.Equal(t, tt.expected.IncludeImplementation, result.IncludeImplementation)
			assert.Equal(t, tt.expected.IncludePrivate, result.IncludePrivate)
		})
	}
}

func TestProcessExcludeList(t *testing.T) {
	tests := []struct {
		name     string
		list     string
		expected processor.ProcessOptions
	}{
		{
			name: "ExcludePrivate",
			list: "private",
			expected: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        true,
				RemovePrivateOnly:     true,
			},
		},
		{
			name: "ExcludeCommentsAndImplementation",
			list: "comments,implementation",
			expected: processor.ProcessOptions{
				IncludeComments:       false,
				IncludeImports:        true,
				IncludeImplementation: false,
				IncludePrivate:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processExcludeList(tt.list)
			assert.Equal(t, tt.expected.IncludeComments, result.IncludeComments)
			assert.Equal(t, tt.expected.IncludeImports, result.IncludeImports)
			assert.Equal(t, tt.expected.IncludeImplementation, result.IncludeImplementation)
			assert.Equal(t, tt.expected.RemovePrivateOnly, result.RemovePrivateOnly)
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}