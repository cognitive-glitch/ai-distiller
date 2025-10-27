# Enhancement Wave 1: Top 5 Comprehensive Upgrade

**Start Date**: 2025-10-27
**Estimated Duration**: 23 hours
**Status**: IN PROGRESS

---

## Execution Order (Optimized)

### 1. Kotlin Generics & Return Types ⚡ (3 hours)
**Priority**: Quick Win | **Complexity**: LOW
- Fix generic parameter extraction
- Add return type parsing
- Update tests (2 new tests)
- **Files**: `crates/lang-kotlin/src/lib.rs`

### 2. Swift Extensions ⚠️ (3 hours)
**Priority**: Critical | **Complexity**: MEDIUM
- Add extension_declaration parsing
- Extract extended type and protocols
- Mark with "extension" decorator
- Update tests (2 new tests)
- **Files**: `crates/lang-swift/src/lib.rs`

### 3. Python Pattern Matching 🔥 (4 hours)
**Priority**: High | **Complexity**: MEDIUM
- Add match_statement node parsing
- Extract case patterns
- Handle guard clauses
- Update tests (3 new tests)
- **Files**: `crates/lang-python/src/lib.rs`

### 4. PHP 8.x Features 🔥 (5 hours)
**Priority**: High | **Complexity**: MEDIUM-HIGH
- Add enum_declaration parsing
- Add attribute parsing (#[...])
- Add readonly property detection
- Add union type support
- Update tests (4 new tests)
- **Files**: `crates/lang-php/src/lib.rs`

### 5. TypeScript Advanced Types 🔥 (8 hours)
**Priority**: High | **Complexity**: HIGH
- Add type_alias declarations
- Add enum_declaration parsing
- Add conditional type detection
- Add mapped type detection
- Add template literal type support
- Update tests (5 new tests)
- **Files**: `crates/lang-typescript/src/lib.rs`

---

## Implementation Checklist

### Enhancement 1: Kotlin ⚡
- [ ] Debug Kotlin AST for generics (type_parameters node)
- [ ] Extract generic parameters in parse_function
- [ ] Extract return type (type_identifier after `:`)
- [ ] Add test_generic_function
- [ ] Add test_return_types
- [ ] Verify all existing tests still pass

### Enhancement 2: Swift ⚠️
- [ ] Debug Swift AST for extensions (extension_declaration)
- [ ] Parse extension name and extended type
- [ ] Parse protocol conformances
- [ ] Extract extension methods/properties
- [ ] Mark with ["extension"] decorator
- [ ] Add test_protocol_extension
- [ ] Add test_type_extension
- [ ] Verify all existing tests still pass

### Enhancement 3: Python 🔥
- [ ] Debug Python AST for match statements (match_statement)
- [ ] Parse match expression
- [ ] Parse case clauses (case_clause)
- [ ] Handle pattern types (literal, wildcard, capture, etc.)
- [ ] Add test_simple_match
- [ ] Add test_complex_match
- [ ] Add test_match_with_guards
- [ ] Verify all existing tests still pass

### Enhancement 4: PHP 🔥
- [ ] Debug PHP AST for enums (enum_declaration)
- [ ] Parse enum cases (enum_case)
- [ ] Parse attributes (attribute)
- [ ] Parse readonly keyword
- [ ] Parse union types (union_type)
- [ ] Add test_enum_declaration
- [ ] Add test_attributes
- [ ] Add test_readonly_properties
- [ ] Add test_union_types
- [ ] Verify all existing tests still pass

### Enhancement 5: TypeScript 🔥
- [ ] Debug TypeScript AST for type aliases (type_alias_declaration)
- [ ] Debug enum declarations (enum_declaration)
- [ ] Parse conditional types (conditional_type)
- [ ] Parse mapped types (mapped_type)
- [ ] Parse template literal types (template_type)
- [ ] Add test_type_alias
- [ ] Add test_enum
- [ ] Add test_conditional_types
- [ ] Add test_mapped_types
- [ ] Add test_template_literals
- [ ] Verify all existing tests still pass

---

## Expected LOC Changes

| Language | Current LOC | Added LOC | New LOC | Tests Added |
|----------|-------------|-----------|---------|-------------|
| Kotlin | 584 | +80 | 664 | +2 (11 total) |
| Swift | 629 | +120 | 749 | +2 (9 total) |
| Python | 644 | +150 | 794 | +3 (9 total) |
| PHP | 714 | +180 | 894 | +4 (14 total) |
| TypeScript | 1040 | +200 | 1240 | +5 (11 total) |
| **Total** | **3,611** | **+730** | **4,341** | **+16 (54 total)** |

---

## Test Strategy

Each enhancement will add comprehensive tests:

1. **Basic test**: Simple usage of the new feature
2. **Complex test**: Advanced usage with edge cases
3. **Integration test**: Feature combined with existing features

**Total New Tests**: 16 (across 5 languages)
**Expected Total**: 106 → 122 tests

---

## Quality Gates

Before marking each enhancement complete:

✅ **Code Quality**:
- Zero clippy warnings
- Proper error handling
- Consistent with existing patterns

✅ **Testing**:
- All new tests passing
- All existing tests still passing
- Edge cases covered

✅ **Documentation**:
- Code comments for complex logic
- Test descriptions clear

---

## Risk Mitigation

**Low Risk**:
- Kotlin (existing TODOs, straightforward)
- Swift (standard pattern, similar to classes)

**Medium Risk**:
- Python (new AST nodes, but well-documented)
- PHP (tree-sitter support solid)

**Higher Risk**:
- TypeScript (complex type system, many interactions)

**Mitigation Strategy**:
- Debug AST structure first (create temp debug programs)
- Implement incrementally (one feature at a time)
- Test thoroughly after each addition
- Keep git commits granular for easy rollback

---

## Success Criteria

**Individual Enhancement Complete**:
- Feature fully implemented
- New tests added and passing
- All existing tests still passing
- Zero clippy warnings
- Code formatted with rustfmt

**Wave 1 Complete**:
- All 5 enhancements implemented
- 16+ new tests passing
- Total test count: 122+
- Zero clippy warnings across all crates
- Documentation updated
- Commit pushed to remote

---

## Timeline

**Optimistic** (if everything goes smoothly): 18 hours
**Realistic** (expected with debugging): 23 hours
**Pessimistic** (if significant issues): 28 hours

**Target Completion**: This session + 1-2 follow-up sessions

---

## Notes

- Focus on correctness over perfection
- Each enhancement is independently valuable
- Can commit after each enhancement for granular history
- Will create session summary after completion

---

Last updated: 2025-10-27
