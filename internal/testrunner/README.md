# AI Distiller Unified Test System

## Overview

This directory contains the unified test runner for AI Distiller that automatically discovers and executes integration tests across all supported programming languages.

## Test Structure

All integration tests follow this structure:

```
testdata/
├── <language>/
│   ├── <NN_scenario_name>/
│   │   ├── source.<ext>           # Main source file to test
│   │   ├── helper.<ext>           # Optional additional files
│   │   └── expected/
│   │       ├── default.expected   # Expected output with default flags
│   │       ├── public.expected    # Expected output with --public=1 only
│   │       └── public.no_impl.expected  # Multiple flags
│   └── ...more scenarios
└── ...more languages
```

## Naming Conventions

### Scenario Directories
- Format: `NN_descriptive_name` (e.g., `01_simple_function`, `02_class_inheritance`)
- Numbers provide ordering but are not required
- Use descriptive names that explain what's being tested

### Source Files
- Primary source file should be named `source.<ext>` (e.g., `source.go`, `source.py`)
- Additional helper files can have any name

### Expected Files
- Located in `expected/` subdirectory
- Two naming formats supported:

#### 1. Parameter-based format (RECOMMENDED)
- Format: `test.<param1=value>.<param2=value>.expected`
- Parameters are extracted and converted to CLI flags
- Examples:
  - `test.implementation=0.expected` → `--implementation=0`
  - `test.public=1.private=0.expected` → `--public=1 --private=0`
  - `test.implementation=0.comments=0.public=1.expected` → Multiple flags

#### 2. Simple format
- `default.expected` - No flags, default behavior

### Parameter Examples

| Filename | CLI Flags Generated |
|----------|-------------------|
| `default.expected` | (none) |
| `test.implementation=0.expected` | `--implementation=0` |
| `test.public=1.protected=0.internal=0.private=0.expected` | `--public=1 --protected=0 --internal=0 --private=0` |
| `test.comments=0.docstrings=0.expected` | `--comments=0 --docstrings=0` |
| `test.implementation=0.public=1.expected` | `--implementation=0 --public=1` |

## Running Tests

### Run All Tests
```bash
go test ./internal/testrunner
```

### Run Specific Language Tests
```bash
go test ./internal/testrunner -run "TestIntegration/python"
```

### Update Expected Files
```bash
UPDATE_EXPECTED=true go test ./internal/testrunner
```

This will regenerate all `.expected` files based on current aid output.

## Adding New Tests

### 1. Create Test Structure
```bash
mkdir -p testdata/python/10_new_feature/expected
```

### 2. Add Source File
Create `testdata/python/10_new_feature/source.py`:
```python
class Example:
    def public_method(self):
        return "public"
    
    def _private_method(self):
        return "private"
```

### 3. Generate Expected Files

Option A: Run with UPDATE_EXPECTED=true (RECOMMENDED)
```bash
UPDATE_EXPECTED=true go test ./internal/testrunner -run "TestIntegration/python/10_new_feature"
```

Option B: Manually run aid and save output
```bash
# Default behavior
aid testdata/python/10_new_feature/source.py --format text > testdata/python/10_new_feature/expected/default.expected

# With specific parameters
aid testdata/python/10_new_feature/source.py --format text --implementation=0 > testdata/python/10_new_feature/expected/test.implementation=0.expected

# Multiple parameters (sorted alphabetically in filename)
aid testdata/python/10_new_feature/source.py --format text --public=1 --protected=0 --internal=0 --private=0 --implementation=0 > testdata/python/10_new_feature/expected/test.implementation=0.internal=0.private=0.protected=0.public=1.expected
```

## Migration from Old Tests

Use the migration helper to convert existing tests:

```bash
go run internal/testrunner/migrate.go
```

This will:
1. Find all existing test files matching known patterns
2. Create the new directory structure under `testdata/`
3. Copy and rename files according to new conventions
4. Transform expected filenames to the new tag-based system

## Best Practices

1. **Test Real-World Patterns**: Each test should represent code patterns developers actually write
2. **Progressive Complexity**: Number tests from simple (01) to complex (10+)
3. **Document Edge Cases**: Add comments in source files explaining what edge case is being tested
4. **Language-Specific Features**: Don't force all languages to have identical tests - test what's idiomatic
5. **Minimal Expected Variants**: Only create expected files for meaningful flag combinations

## Common Flag Combinations

These are the most useful combinations to test:

1. `default.expected` - Baseline behavior
2. `public.expected` - API documentation use case
3. `no_impl.expected` - Header/interface generation
4. `public.no_impl.expected` - Clean API reference
5. `all.expected` - Complete code analysis

## Debugging Failed Tests

When a test fails, the error message will show:
- The exact aid command that was run
- The expected output
- The actual output
- A diff between them

To debug:
1. Run the aid command manually to see full output
2. Check if the source file has syntax errors
3. Verify the expected file matches your intentions
4. Use UPDATE_EXPECTED=true to see what aid currently produces

## CI Integration

Add to your CI workflow:

```yaml
- name: Run Integration Tests
  run: go test ./internal/testrunner -v

- name: Check Test Coverage
  run: go test ./internal/testrunner -run "TestIntegration" -count=1
```