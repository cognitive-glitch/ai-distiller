# Test Coverage Enhancement Plan

## Executive Summary

**Current State**: 106 tests across 12 languages (6-10 tests per language)
**Target State**: 180+ tests with comprehensive coverage (15+ tests per language)
**Gap**: Missing edge cases, error handling, feature-specific tests

## Current Test Analysis

### Per-Language Test Count (Before Enhancement)
| Language   | Current Tests | Quality | Coverage Gaps |
|----------|---------------|---------|---------------|
| Python      | 6 tests | Basic | Decorators, async, pattern matching, type hints |
| TypeScript  | 6 tests | Basic | Advanced types, generics, decorators, namespaces |
| Go          | 6 tests | Basic | Interfaces, embedding, generics, error handling |
| Rust        | 6 tests | Basic | Traits, lifetimes, macros, async |
| Swift       | 7 tests | Basic | **Extensions (CRITICAL)**, protocols, optionals |
| Ruby        | 6 tests | Basic | Blocks, modules, metaprogramming |
| Java        | 8 tests | Medium | Annotations, records, sealed classes |
| C#          | 9 tests | Medium | Records, nullable refs, pattern matching |
| Kotlin      | 9 tests | Medium | **Generics, return types**, coroutines, extensions |
| C++         | 10 tests | Good | Templates, concepts, RAII |
| JavaScript  | 6 tests | Basic | Async/await, classes, modules |
| PHP         | 10 tests | Good | **8.x features** (enums, attributes, readonly) |

## Test Categories to Add

### Category 1: Core Feature Tests (All Languages)
- ✅ Processor creation
- ✅ Extension detection
- ✅ Visibility detection
- ✅ Simple functions
- ✅ Simple classes
- ✅ Import statements
- ⚠️ **Generic/Type parameters** (missing in most)
- ⚠️ **Return types** (partially covered)
- ⚠️ **Decorators/Annotations** (missing in most)
- ⚠️ **Inheritance** (missing explicit tests)
- ⚠️ **Interface implementation** (missing)
- ⚠️ **Field parsing** (minimal coverage)
- ⚠️ **Modifiers** (static, abstract, final, etc.)

### Category 2: Language-Specific Advanced Features
**Python**:
- Pattern matching (match/case)
- Type hints (Union, Optional, Generic)
- Async/await functions
- Property decorators (@property, @staticmethod)
- Context managers
- Multiple inheritance

**TypeScript**:
- Advanced types (conditional, mapped, template literal)
- Type aliases
- Enums
- Namespaces
- Generic constraints
- Optional chaining

**Go**:
- Interface satisfaction
- Struct embedding
- Generics with constraints
- Receiver methods
- Error handling patterns

**Rust**:
- Trait implementations
- Lifetimes
- Macros
- Async/await
- Associated types

**Swift**:
- **Extensions (CRITICAL - completely missing)**
- Protocol conformance
- Optionals
- Closures
- Property wrappers

**Kotlin**:
- **Generic parameters (CRITICAL - missing)**
- **Return types (CRITICAL - missing)**
- Coroutines
- Extensions
- Data classes

**PHP**:
- **PHP 8.x enums**
- **Attributes (#[...])**
- **Readonly properties**
- **Union types**
- Named arguments

### Category 3: Edge Cases & Error Handling
- Empty files
- Syntax errors (graceful handling)
- Nested classes
- Complex generics
- Multiple inheritance
- Circular dependencies
- Very long identifiers
- Unicode in identifiers
- Comments handling

### Category 4: Integration Tests
- Multiple classes in one file
- Mixed visibility levels
- Complex inheritance hierarchies
- Interface + implementation
- Decorator stacking

## Enhancement Targets

### Phase 1: Core Features (Add 5 tests per language = 60 tests)
**Priority**: HIGH | **Duration**: 4 hours
- Generic/type parameters test
- Return types comprehensive test
- Inheritance test
- Interface implementation test
- Multiple modifiers test

### Phase 2: Language-Specific Features (Add 3-5 tests per language = 48 tests)
**Priority**: HIGH | **Duration**: 6 hours
- Python: Pattern matching, async, type hints
- TypeScript: Advanced types, enums, namespaces
- Go: Interface satisfaction, generics
- Rust: Traits, lifetimes, async
- Swift: **Extensions (CRITICAL)**, protocols
- Kotlin: **Generics, return types**
- PHP: **8.x features**
- Others: Language-specific critical features

### Phase 3: Edge Cases (Add 2-3 tests per language = 30 tests)
**Priority**: MEDIUM | **Duration**: 3 hours
- Empty file handling
- Nested structures
- Complex scenarios
- Unicode identifiers
- Error recovery

## Implementation Strategy

### Systematic Approach (Per Language)
1. **Read** current tests in `crates/lang-X/src/lib.rs`
2. **Identify** missing coverage from categories above
3. **Write** 10-15 new comprehensive tests
4. **Run** `cargo test -p lang-X` to verify
5. **Document** coverage improvements
6. **Commit** per-language with clear message

### Test Naming Convention
```rust
#[test]
fn test_<feature>_<scenario>() {
    // Examples:
    // test_generics_single_param()
    // test_generics_multiple_params()
    // test_generics_with_constraints()
    // test_return_type_simple()
    // test_return_type_generic()
    // test_inheritance_single()
    // test_inheritance_multiple()
    // test_interface_implementation()
    // test_decorators_stacked()
    // test_modifiers_static_final()
}
```

### Test Quality Standards
- **Assertions**: Use specific assertions, not just `!is_empty()`
- **Validate**: Check names, visibility, modifiers, types, parameters
- **Edge Cases**: Test boundary conditions
- **Error Messages**: Clear panic messages
- **Documentation**: Comment complex test scenarios

### Example Enhanced Test
```rust
#[test]
fn test_generics_with_constraints() {
    let processor = GoProcessor::new().unwrap();
    let source = r#"
        func Process[T Comparable, V any](key T, value V) V {
            return value
        }
    "#;
    let opts = ProcessOptions::default();
    let file = processor.process(source, Path::new("test.go"), &opts).unwrap();

    assert_eq!(file.children.len(), 1);
    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "Process");
        assert_eq!(func.type_params.len(), 2);

        // Validate first type param
        assert_eq!(func.type_params[0].name, "T");
        assert_eq!(func.type_params[0].constraint, Some("Comparable".to_string()));

        // Validate second type param
        assert_eq!(func.type_params[1].name, "V");
        assert_eq!(func.type_params[1].constraint, Some("any".to_string()));

        // Validate parameters
        assert_eq!(func.parameters.len(), 2);
        assert_eq!(func.parameters[0].name, "key");
        assert_eq!(func.parameters[0].param_type.name, "T");

        // Validate return type
        assert!(func.return_type.is_some());
        assert_eq!(func.return_type.as_ref().unwrap().name, "V");
    } else {
        panic!("Expected function node, got {:?}", file.children[0]);
    }
}
```

## Success Metrics

### Quantitative
- ✅ Total tests: 106 → 180+ (70% increase)
- ✅ Average tests per language: 9 → 15+ (67% increase)
- ✅ Code coverage: TBD (measure with cargo-tarpaulin)
- ✅ All tests passing

### Qualitative
- ✅ All core features have dedicated tests
- ✅ All language-specific features tested
- ✅ Edge cases covered
- ✅ Clear, documented test cases
- ✅ Easy to add new tests (good patterns established)

## Timeline

| Phase | Duration | Tests Added | Languages |
|-------|----------|-------------|-----------|
| Phase 1: Core Features | 4 hours | 60 tests | All 12 |
| Phase 2: Lang-Specific | 6 hours | 48 tests | All 12 |
| Phase 3: Edge Cases | 3 hours | 30 tests | All 12 |
| **Total** | **13 hours** | **138 tests** | **12 languages** |

**Final Count**: 106 + 138 = **244 total tests**

## Risk Assessment

### Low Risk
- Adding tests won't break existing functionality
- Incremental approach allows rollback
- Each language independent

### Medium Risk
- Time investment (13 hours)
- May discover bugs in existing parsers
- Need to maintain test data

### Mitigation
- Test after each language enhancement
- Fix bugs as discovered
- Document any parser changes needed

## Next Actions

1. **Start with Python** (foundational, well-understood)
2. **Move to Kotlin** (has identified gaps - generics, return types)
3. **Then Swift** (CRITICAL - missing extensions)
4. **Continue systematically** through remaining languages
5. **Document** improvements in STATUS.md
6. **Commit** after each language with detailed message
