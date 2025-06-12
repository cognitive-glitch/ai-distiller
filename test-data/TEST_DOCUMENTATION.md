# AI Distiller Test Documentation

## Overview
Comprehensive test suite for AI Distiller that validates real behavior with various Python files and command-line options.

## Test Types

### 1. Unit Tests (`distiller_test.go`)
Direct testing of the Python processor and formatters without CLI.

**Test Cases:**
- `full_output_basic_class` - Tests complete parsing with all features enabled
- `no_private_basic_class` - Validates private member filtering (`_` prefix)
- `no_implementation_basic_class` - Tests implementation stripping
- `complex_imports_parsing` - Validates import statement parsing
- `decorators_parsing` - Tests decorator preservation
- `inheritance_parsing` - Tests class inheritance detection
- `edge_cases_unicode` - Tests unicode identifiers and special constructs

**Validators:**
- `validateClassCount` - Verifies correct number of classes
- `validateFunctionCount` - Verifies correct number of functions
- `validateImportCount` - Verifies correct number of imports
- `validateHasClass` - Checks if specific class exists
- `validateHasFunction` - Checks if specific function exists
- `validateNoFunction` - Ensures function is filtered out
- `validateFunctionHasImplementation` - Verifies implementation is preserved
- `validateFunctionNoImplementation` - Verifies implementation is stripped
- `validateClassExtends` - Validates inheritance relationships
- `validateFunctionHasDecorator` - Checks decorator preservation

### 2. Integration Tests (`integration_test.go`)
Tests the complete CLI with real command execution.

**Test Scenarios:**
- `distill_basic_class_json` - Tests JSON output format
- `distill_no_private_markdown` - Tests --no-private flag
- `distill_no_implementation_json` - Tests --no-implementation flag
- `distill_minimal_markdown` - Tests --minimal flag
- `distill_complex_imports` - Tests import parsing
- `distill_directory_multiple_files` - Tests directory processing

**Option Interaction Tests:**
- Combining `--no-private` with `--no-implementation`
- Testing `--no-imports` option

**Error Handling Tests:**
- Nonexistent file handling
- Invalid format handling
- Empty directory handling

### 3. Format Validation (`format_validator.go`)
Ensures consistency across all output formats (JSON, XML, Markdown, JSONL).

**Validations:**
- Same number of classes, functions, imports in all formats
- No data loss between formats
- Consistent structure representation

### 4. Test Runner (`test_runner.go`)
Executes multiple scenarios with different configurations.

**Scenarios:**
1. `full_output` - All information preserved
2. `no_private` - Private members filtered
3. `no_implementation` - Implementation bodies removed
4. `minimal` - Structure only
5. `complex_imports` - Import handling
6. `decorators` - Decorator preservation
7. `inheritance` - OOP constructs
8. `edge_cases` - Special Python features

## Running Tests

### Quick Start
```bash
# Run all tests
make test-all

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Quick test of basic functionality
make test-quick
```

### Individual Test Commands
```bash
# Unit tests only
go test -v distiller_test.go

# Integration tests (requires built CLI)
go test -v integration_test.go

# Format validation
go run format_validator.go

# Run all test scenarios
go run test_runner.go
```

### Benchmarking
```bash
# Run performance benchmarks
make benchmark
```

## Test Data Structure
```
test-data/
├── input/                    # Test Python files
│   ├── basic_class.py       # Standard OOP patterns
│   ├── complex_imports.py   # Import variations
│   ├── decorators_and_metadata.py  # Decorators, dataclasses
│   ├── inheritance_patterns.py     # Class inheritance
│   └── edge_cases.py        # Unicode, async, metaclasses
├── actual/                   # Generated outputs (git-ignored)
├── distiller_test.go        # Unit tests
├── integration_test.go      # CLI integration tests
├── format_validator.go      # Cross-format validation
├── test_runner.go           # Scenario executor
└── Makefile                 # Test automation
```

## Test Results Interpretation

### Passing Tests ✅
- Basic class/function parsing
- Private member filtering
- Implementation stripping
- Format consistency

### Known Limitations ⚠️
Due to the simple line-based parser (pending tree-sitter WASM):
- Complex multi-line imports may not parse correctly
- Some decorators might be missed
- Unicode handling depends on proper encoding
- Nested classes/functions have limited support

## Adding New Tests

### 1. Add Test Case to `distiller_test.go`
```go
{
    Name:      "your_test_name",
    InputFile: "input/your_test_file.py",
    Options: processor.ProcessOptions{
        // Configure options
    },
    Validators: []Validator{
        // Add validators
    },
}
```

### 2. Create Validator Functions
```go
func validateYourCondition(expected string) Validator {
    return func(t *testing.T, file *ir.DistilledFile, tc TestCase) {
        // Your validation logic
    }
}
```

### 3. Add Integration Test
```go
{
    name: "your_integration_test",
    args: []string{"distill", "input/file.py", "--your-flag"},
    validate: func(t *testing.T, output string) {
        // Validate CLI output
    },
}
```

## Continuous Improvement
1. When tree-sitter WASM is integrated, all edge case tests should pass
2. Add more language-specific test cases
3. Implement automated regression testing
4. Add performance benchmarks for large files