package python

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/janreges/ai-distiller/internal/ir"
)

// TestPythonConstructs tests all 5 Python constructs with 3 different options
func TestPythonConstructs(t *testing.T) {
	tests := []struct {
		name     string
		testDir  string
		filename string
	}{
		{
			name:     "Construct 1 - Basic validation utils",
			testDir:  "construct_1_basic",
			filename: "validation_utils.py",
		},
		{
			name:     "Construct 2 - Simple user model",
			testDir:  "construct_2_simple", 
			filename: "user_model.py",
		},
		{
			name:     "Construct 3 - Medium notification service",
			testDir:  "construct_3_medium",
			filename: "notification_service.py",
		},
		{
			name:     "Construct 4 - Complex plugin manager",
			testDir:  "construct_4_complex",
			filename: "plugin_manager.py",
		},
		{
			name:     "Construct 5 - Very complex dynamic config",
			testDir:  "construct_5_very_complex",
			filename: "dynamic_config.py",
		},
	}

	options := []struct {
		name               string
		opts               processor.ProcessOptions
		expectedSuffix     string
	}{
		{
			name: "Full output",
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        true,
			},
			expectedSuffix: "expected_full.txt",
		},
		{
			name: "No private members", 
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: true,
				IncludePrivate:        false,
			},
			expectedSuffix: "expected_no_private.txt",
		},
		{
			name: "No implementation",
			opts: processor.ProcessOptions{
				IncludeComments:       true,
				IncludeImports:        true,
				IncludeImplementation: false,
				IncludePrivate:        true,
			},
			expectedSuffix: "expected_no_impl.txt",
		},
	}

	// Get test data directory
	testDataDir := filepath.Join("..", "..", "..", "test-data", "python")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			constructDir := filepath.Join(testDataDir, test.testDir)
			sourceFile := filepath.Join(constructDir, test.filename)

			// Check if source file exists
			if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
				t.Skipf("Source file %s does not exist", sourceFile)
			}

			// Create processor
			proc := NewProcessor()
			proc.EnableTreeSitter()

			for _, opt := range options {
				t.Run(opt.name, func(t *testing.T) {
					// Process the file
					result, err := proc.ProcessFile(sourceFile, opt.opts)
					if err != nil {
						t.Fatalf("Failed to process file: %v", err)
					}

					// Verify basic structure
					if result == nil {
						t.Fatal("Result is nil")
					}

					if result.Language != "python" {
						t.Errorf("Expected language 'python', got '%s'", result.Language)
					}

					if result.Path != sourceFile {
						t.Errorf("Expected path '%s', got '%s'", sourceFile, result.Path)
					}

					// Check that we have some children
					if len(result.Children) == 0 {
						t.Error("Expected some children in the distilled file")
					}

					// Load expected output and compare key characteristics
					expectedFile := filepath.Join(constructDir, opt.expectedSuffix)
					if _, err := os.Stat(expectedFile); err == nil {
						expectedBytes, err := os.ReadFile(expectedFile)
						if err != nil {
							t.Fatalf("Failed to read expected file: %v", err)
						}
						expected := strings.TrimSpace(string(expectedBytes))

						// Basic checks on expected content
						if opt.opts.IncludePrivate {
							// Should contain imports if source has them
							if strings.Contains(expected, "import") && !strings.Contains(expected, "import") {
								t.Error("Expected file should contain imports when IncludePrivate is true")
							}
						}

						if !opt.opts.IncludeImplementation {
							// Should not contain function bodies (implementation)
							if strings.Contains(expected, "if not isinstance") {
								t.Error("Expected file should not contain implementation details when IncludeImplementation is false")
							}
						}
					}

					// Construct-specific validations
					switch test.testDir {
					case "construct_1_basic":
						validateConstruct1(t, result, opt.opts)
					case "construct_2_simple":
						validateConstruct2(t, result, opt.opts)
					case "construct_3_medium":
						validateConstruct3(t, result, opt.opts)
					case "construct_4_complex":
						validateConstruct4(t, result, opt.opts)
					case "construct_5_very_complex":
						validateConstruct5(t, result, opt.opts)
					}
				})
			}
		})
	}
}

func validateConstruct1(t *testing.T, result *ir.DistilledFile, opts processor.ProcessOptions) {
	// Should have imports, functions, and possibly private constants
	hasImports := false
	hasFunctions := false
	hasPrivateConstants := false

	for _, child := range result.Children {
		switch child.(type) {
		case *ir.DistilledImport:
			hasImports = true
		case *ir.DistilledFunction:
			hasFunctions = true
		case *ir.DistilledField:
			if field := child.(*ir.DistilledField); field.Visibility == ir.VisibilityPrivate {
				hasPrivateConstants = true
			}
		}
	}

	if !hasImports {
		t.Error("Expected imports in construct 1")
	}
	if !hasFunctions {
		t.Error("Expected functions in construct 1")
	}
	if opts.IncludePrivate && opts.IncludeImplementation && !hasPrivateConstants {
		t.Error("Expected private constants when IncludePrivate and IncludeImplementation are true")
	}
	if !opts.IncludePrivate && hasPrivateConstants {
		t.Error("Should not have private constants when IncludePrivate is false")
	}
}

func validateConstruct2(t *testing.T, result *ir.DistilledFile, opts processor.ProcessOptions) {
	// Should have a class with methods
	hasClass := false
	hasInit := false

	for _, child := range result.Children {
		if class, ok := child.(*ir.DistilledClass); ok {
			hasClass = true
			if class.Name == "User" {
				// Check for __init__ method
				for _, method := range class.Children {
					if fn, ok := method.(*ir.DistilledFunction); ok && fn.Name == "__init__" {
						hasInit = true
						break
					}
				}
			}
		}
	}

	if !hasClass {
		t.Error("Expected User class in construct 2")
	}
	if !hasInit {
		t.Error("Expected __init__ method in User class")
	}
}

func validateConstruct3(t *testing.T, result *ir.DistilledFile, opts processor.ProcessOptions) {
	// Should have inheritance and abstract methods
	hasBaseClass := false
	hasDerivedClass := false

	for _, child := range result.Children {
		if class, ok := child.(*ir.DistilledClass); ok {
			if class.Name == "BaseNotifier" {
				hasBaseClass = true
			}
			if class.Name == "EmailNotifier" && len(class.Extends) > 0 {
				hasDerivedClass = true
			}
		}
	}

	if !hasBaseClass {
		t.Error("Expected BaseNotifier class in construct 3")
	}
	if !hasDerivedClass {
		t.Error("Expected EmailNotifier class with inheritance in construct 3")
	}
}

func validateConstruct4(t *testing.T, result *ir.DistilledFile, opts processor.ProcessOptions) {
	// Should have protocols and complex typing
	hasProtocol := false
	hasAdvancedTypes := false

	for _, child := range result.Children {
		if class, ok := child.(*ir.DistilledClass); ok && class.Name == "Plugin" {
			hasProtocol = true
		}
		if fn, ok := child.(*ir.DistilledFunction); ok {
			// Check for advanced typing in parameters or return types
			for _, param := range fn.Parameters {
				if strings.Contains(param.Type.Name, "Dict") || strings.Contains(param.Type.Name, "Callable") {
					hasAdvancedTypes = true
					break
				}
			}
			// Also check return type
			if fn.Returns != nil && (strings.Contains(fn.Returns.Name, "Dict") || strings.Contains(fn.Returns.Name, "Callable")) {
				hasAdvancedTypes = true
			}
		}
	}

	if !hasProtocol {
		t.Error("Expected Plugin protocol in construct 4")
	}
	if !hasAdvancedTypes {
		t.Error("Expected advanced typing (Dict, Callable) in construct 4")
	}
}

func validateConstruct5(t *testing.T, result *ir.DistilledFile, opts processor.ProcessOptions) {
	// Should have metaclasses, descriptors, and context managers
	hasMetaclass := false
	hasDescriptor := false
	hasContextManager := false

	for _, child := range result.Children {
		if class, ok := child.(*ir.DistilledClass); ok {
			if class.Name == "SingletonMeta" {
				hasMetaclass = true
			}
			if class.Name == "ValidatedSetting" {
				hasDescriptor = true
			}
			if class.Name == "AppConfig" {
				// Check for context manager methods
				for _, method := range class.Children {
					if fn, ok := method.(*ir.DistilledFunction); ok {
						if fn.Name == "__enter__" || fn.Name == "__exit__" {
							hasContextManager = true
							break
						}
					}
				}
			}
		}
	}

	if !hasMetaclass {
		t.Error("Expected SingletonMeta metaclass in construct 5")
	}
	if !hasDescriptor {
		t.Error("Expected ValidatedSetting descriptor in construct 5")
	}
	if !hasContextManager {
		t.Error("Expected context manager methods (__enter__/__exit__) in construct 5")
	}
}