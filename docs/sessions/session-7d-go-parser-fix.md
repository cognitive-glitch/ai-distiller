# Session 7D: Go Parser Fixes (Multiple Return Values + Variadic Parameters)

## Status: Complete ✅

### Summary

Fixed 2 critical bugs in Go parser that were causing test failures:
1. Multiple return values parsed as parameters
2. Variadic parameters not recognized

All 17 Go tests now passing (from 15/17).

### Problems Discovered

While starting Phase C (Testing & Quality), discovered 2 failing Go tests:
- `test_multiple_return_values` - Expected 1 parameter, found 3
- `test_variadic_parameters` - Expected at least 1 parameter, found 0

### Investigation

Used AST debugging (print_tree function) to reveal tree-sitter-go structure:

#### Multiple Return Values Structure

```
function_declaration "func GetUser(id int) (*User, bool, error)"
  identifier "GetUser"
  parameter_list "(id int)"           ← First parameter_list = parameters
    parameter_declaration "id int"
  parameter_list "(*User, bool, error)" ← Second parameter_list = return types!
    parameter_declaration "*User"
    parameter_declaration "bool"
    parameter_declaration "error"
```

**Root Cause**: parse_function was treating BOTH parameter_list nodes as parameters.

#### Variadic Parameters Structure

```
parameter_list "(numbers ...int)"
  variadic_parameter_declaration "numbers ...int"
    identifier "numbers"
    ... "..."
    type_identifier "int"
```

**Root Cause**: parse_parameters only checked for `parameter_declaration`, not `variadic_parameter_declaration`.

### Fixes Implemented

#### Fix 1: parse_function - Multiple Parameter Lists

Added `has_seen_parameters` flag to track state:

```rust
fn parse_function(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
    let mut has_seen_name = false;
    let mut has_seen_parameters = false;  // NEW
    
    for child in node.children(&mut cursor) {
        match child.kind() {
            "parameter_list" => {
                // Go functions can have multiple parameter_list nodes:
                // 1. Receiver (for methods) - before name
                // 2. Parameters - after name
                // 3. Return types - wrapped in parameter_list after parameters
                if !has_seen_name {
                    // This is a receiver (for methods)
                    let receiver_params = self.parse_parameters(child, source)?;
                    if !receiver_params.is_empty() {
                        receiver_type = Some(receiver_params[0].param_type.clone());
                    }
                } else if !has_seen_parameters {
                    // This is the actual parameter list
                    parameters = self.parse_parameters(child, source)?;
                    has_seen_parameters = true;
                }
                // Skip subsequent parameter_list nodes (return types)
            }
            // ...
        }
    }
}
```

**Logic**:
- Before function name → receiver (methods)
- After name, first parameter_list → actual parameters
- After name, subsequent parameter_lists → return types (ignored)

#### Fix 2: parse_parameters - Handle Variadic

Added support for `variadic_parameter_declaration`:

```rust
fn parse_parameters(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Parameter>> {
    let mut parameters = Vec::new();

    let mut cursor = node.walk();
    for child in node.children(&mut cursor) {
        match child.kind() {
            "parameter_declaration" => {
                parameters.extend(self.parse_parameter_declaration(child, source)?);
            }
            "variadic_parameter_declaration" => {  // NEW
                parameters.extend(self.parse_variadic_parameter(child, source)?);
            }
            _ => {}
        }
    }

    Ok(parameters)
}
```

#### Fix 3: New parse_variadic_parameter Method

Dedicated handler for variadic parameters with `is_variadic: true` flag:

```rust
fn parse_variadic_parameter(
    &self,
    node: tree_sitter::Node,
    source: &str,
) -> Result<Vec<Parameter>> {
    let mut name = String::new();
    let mut param_type = None;

    let mut cursor = node.walk();
    for child in node.children(&mut cursor) {
        match child.kind() {
            "identifier" | "field_identifier" => {
                name = self.node_text(child, source);
            }
            "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
            | "slice_type" | "map_type" | "channel_type" | "function_type"
            | "interface_type" | "struct_type" => {
                param_type = Some(TypeRef::new(self.node_text(child, source)));
            }
            "..." => {
                // Variadic marker
            }
            _ => {}
        }
    }

    Ok(vec![Parameter {
        name,
        param_type: param_type.unwrap_or_else(|| TypeRef::new("".to_string())),
        default_value: None,
        is_variadic: true,  // CRITICAL: Mark as variadic
        is_optional: false,
        decorators: vec![],
    }])
}
```

### Test Results

**Before Fixes**:
- 15/17 tests passing
- test_multiple_return_values FAILED (expected 1 param, got 3)
- test_variadic_parameters FAILED (expected >=1 param, got 0)

**After Fixes**:
- 17/17 tests passing ✅
- All workspace tests passing (no failures)

### Files Modified

1. `crates/lang-go/src/lib.rs`:
   - Modified `parse_function` (lines 341-410): Added has_seen_parameters tracking
   - Modified `parse_parameters` (lines 413-430): Handle variadic_parameter_declaration
   - Added `parse_variadic_parameter` (lines 476-522): New method for variadic params
   - Added debug_tests module (lines 1223-1293): AST debugging tools

### Key Learnings

1. **Tree-sitter Multiple Nodes of Same Type**
   - Same node kind can appear multiple times with different meanings
   - Context and order matter: receiver → parameters → return types
   - Need state tracking (has_seen_name, has_seen_parameters) to disambiguate

2. **Similar ≠ Same Node Types**
   - `parameter_declaration` vs `variadic_parameter_declaration`
   - Both represent parameters but have different AST structure
   - Pattern matching must cover all variants

3. **Go's Unique Return Value Syntax**
   - Multiple return values wrapped in parameter_list node
   - Not a separate return_type_list node kind
   - Tree-sitter models it as another parameter_list (syntactically correct)

4. **Debug-Driven Development**
   - AST debugging essential before fixing
   - print_tree function reveals true structure
   - Assumptions about AST structure often wrong

### Implementation Time

- **Investigation**: 15 minutes (AST debugging)
- **Fixes**: 20 minutes (implement + test)
- **Total**: 35 minutes

### Impact

- Unblocks Phase C (Testing & Quality) start
- Establishes clean baseline: **All language processors passing** ✅
- Go parser now handles advanced features correctly

### Next Steps

- Continue with Phase C1: Integration Testing
- Phase C established on solid foundation (no known test failures)
