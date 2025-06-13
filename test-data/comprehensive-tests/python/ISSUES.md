# AI Distiller Python Parser Issues

## Critical Bugs Found

### 1. Missing `async` keyword (CRITICAL)
- **Issue**: All `async def` functions are output as regular `def`
- **Impact**: Completely breaks async code semantics
- **Root Cause**: Text formatter doesn't check for `ir.ModifierAsync` in function modifiers
- **Location**: `internal/formatter/text_formatter.go:174`

### 2. Missing metaclass parameter (CRITICAL)
- **Issue**: `metaclass=EnforceAsyncMeta` parameter is completely omitted
- **Impact**: Metaclass logic is bypassed, breaking class validation/transformation
- **Root Cause**: Parser doesn't capture metaclass parameter in class definition

### 3. Missing inheritance syntax in output
- **Issue**: Parent classes not shown in text output (though relationship is captured)
- **Impact**: Makes output invalid Python syntax
- **Root Cause**: Text formatter doesn't output parent classes

### 4. Docstring to comment conversion (HIGH)
- **Issue**: Class docstrings converted to `#` comments
- **Impact**: Changes semantic meaning - docstrings are runtime accessible
- **Example**: `"""A base class..."""` becomes `# A base class...`

### 5. Nested classes not captured (HIGH)
- **Issue**: Nested class `_Formatter` completely missing from output
- **Impact**: Loses important structural information
- **Root Cause**: Line-based parser doesn't handle nested classes

### 6. Inconsistent `__main__` block parsing (MEDIUM)
- **Issue**: Only first line of `if __name__ == "__main__":` block is extracted
- **Impact**: Loses example usage code

### 7. Invalid syntax in stripped output (MEDIUM)
- **Issue**: Functions without implementation lack `: ...` or `: pass`
- **Impact**: Output is not valid Python code
- **Example**: `def foo()` instead of `def foo(): ...`

### 8. Method docstrings removed with implementation (LOW)
- **Issue**: `--strip implementation` removes method docstrings
- **Impact**: Loses API documentation

## Root Cause Analysis

1. **Line-based parser fallback**: When tree-sitter fails, fallback to line parser loses modern Python features
2. **Text formatter limitations**: Missing support for modifiers, inheritance syntax
3. **Parser limitations**: No support for nested classes, metaclass parameters

## Recommendations

1. Fix text formatter to output `async def` when `ir.ModifierAsync` is present
2. Enhance parser to capture metaclass parameters
3. Fix text formatter to output inheritance syntax
4. Add support for nested classes
5. Fix docstring handling to preserve triple quotes
6. Add `: ...` to stripped function definitions for valid syntax