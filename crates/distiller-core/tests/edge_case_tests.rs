/**
 * Edge Case Testing Suite
 *
 * Tests parser robustness against:
 * 1. Malformed code (syntax errors)
 * 2. Large files (10k+ lines)
 * 3. Unicode characters
 * 4. Syntax edge cases (empty, deeply nested, complex)
 */

use distiller_core::language::LanguageRegistry;
use distiller_core::processor::{DirectoryProcessor, ProcessOptions};
use std::fs;
use std::path::Path;
use std::time::Instant;

fn create_full_registry() -> LanguageRegistry {
    let mut registry = LanguageRegistry::new();

    #[cfg(feature = "lang-python")]
    registry.register(Box::new(lang_python::PythonProcessor::new().unwrap()));

    #[cfg(feature = "lang-typescript")]
    registry.register(Box::new(lang_typescript::TypeScriptProcessor::new().unwrap()));

    #[cfg(feature = "lang-go")]
    registry.register(Box::new(lang_go::GoProcessor::new().unwrap()));

    registry
}

/// Test 1: Malformed Code Handling
/// Parsers should handle syntax errors gracefully without crashing
#[test]
#[cfg(feature = "lang-python")]
fn test_malformed_python() {
    let source = fs::read_to_string("../../testdata/edge-cases/malformed/python_syntax_error.py")
        .expect("Failed to read malformed Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    // Should not panic - tree-sitter handles malformed code
    let result = processor.process(&source, Path::new("test.py"), &opts);

    // Result might be Ok (partial parse) or Err (critical error)
    // Key: should not panic/crash
    match result {
        Ok(file) => {
            println!("✓ Malformed Python: Partial parse successful");
            println!("  Found {} top-level nodes", file.children.len());
        }
        Err(e) => {
            println!("✓ Malformed Python: Error handled gracefully: {}", e);
        }
    }
}

#[test]
#[cfg(feature = "lang-typescript")]
fn test_malformed_typescript() {
    let source =
        fs::read_to_string("../../testdata/edge-cases/malformed/typescript_syntax_error.ts")
            .expect("Failed to read malformed TypeScript file");

    let processor = lang_typescript::TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("test.ts"), &opts);

    match result {
        Ok(file) => {
            println!("✓ Malformed TypeScript: Partial parse successful");
            println!("  Found {} top-level nodes", file.children.len());
        }
        Err(e) => {
            println!("✓ Malformed TypeScript: Error handled gracefully: {}", e);
        }
    }
}

#[test]
#[cfg(feature = "lang-go")]
fn test_malformed_go() {
    let source = fs::read_to_string("../../testdata/edge-cases/malformed/go_syntax_error.go")
        .expect("Failed to read malformed Go file");

    let processor = lang_go::GoProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("test.go"), &opts);

    match result {
        Ok(file) => {
            println!("✓ Malformed Go: Partial parse successful");
            println!("  Found {} top-level nodes", file.children.len());
        }
        Err(e) => {
            println!("✓ Malformed Go: Error handled gracefully: {}", e);
        }
    }
}

/// Test 2: Large File Performance
/// Parser should handle 10k+ line files efficiently
#[test]
#[cfg(feature = "lang-python")]
fn test_large_python_file() {
    let source = fs::read_to_string("../../testdata/edge-cases/large-files/large_python.py")
        .expect("Failed to read large Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    println!("Testing large Python file: {} lines", source.lines().count());

    let start = Instant::now();
    let result = processor.process(&source, Path::new("large.py"), &opts);
    let duration = start.elapsed();

    assert!(result.is_ok(), "Large Python file should parse successfully");

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Large Python: {} classes parsed in {:?}", class_count, duration);
    println!("  Performance: ~{} lines/ms", source.lines().count() / duration.as_millis().max(1) as usize);

    // Performance target: should parse in reasonable time (< 1 second for 15k lines)
    assert!(
        duration.as_secs() < 1,
        "Large file parsing took too long: {:?}",
        duration
    );
}

#[test]
#[cfg(feature = "lang-typescript")]
fn test_large_typescript_file() {
    let source = fs::read_to_string("../../testdata/edge-cases/large-files/large_typescript.ts")
        .expect("Failed to read large TypeScript file");

    let processor = lang_typescript::TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    println!("Testing large TypeScript file: {} lines", source.lines().count());

    let start = Instant::now();
    let result = processor.process(&source, Path::new("large.ts"), &opts);
    let duration = start.elapsed();

    assert!(
        result.is_ok(),
        "Large TypeScript file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Large TypeScript: {} classes parsed in {:?}", class_count, duration);
    println!("  Performance: ~{} lines/ms", source.lines().count() / duration.as_millis().max(1) as usize);

    assert!(
        duration.as_secs() < 1,
        "Large file parsing took too long: {:?}",
        duration
    );
}

#[test]
#[cfg(feature = "lang-go")]
fn test_large_go_file() {
    let source = fs::read_to_string("../../testdata/edge-cases/large-files/large_go.go")
        .expect("Failed to read large Go file");

    let processor = lang_go::GoProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    println!("Testing large Go file: {} lines", source.lines().count());

    let start = Instant::now();
    let result = processor.process(&source, Path::new("large.go"), &opts);
    let duration = start.elapsed();

    assert!(result.is_ok(), "Large Go file should parse successfully");

    let file = result.unwrap();
    let struct_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Large Go: {} structs parsed in {:?}", struct_count, duration);
    println!("  Performance: ~{} lines/ms", source.lines().count() / duration.as_millis().max(1) as usize);

    assert!(
        duration.as_secs() < 1,
        "Large file parsing took too long: {:?}",
        duration
    );
}

/// Test 3: Unicode Character Handling
#[test]
#[cfg(feature = "lang-python")]
fn test_unicode_python() {
    let source = fs::read_to_string("../../testdata/edge-cases/unicode/python_unicode.py")
        .expect("Failed to read Unicode Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("unicode.py"), &opts);

    assert!(result.is_ok(), "Unicode Python file should parse successfully");

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Unicode Python: {} classes with Unicode identifiers", class_count);

    // Should find classes with Unicode names (Пользователь, 🚀Rocket, etc.)
    assert!(class_count >= 5, "Should find at least 5 classes with Unicode names");
}

#[test]
#[cfg(feature = "lang-typescript")]
fn test_unicode_typescript() {
    let source = fs::read_to_string("../../testdata/edge-cases/unicode/typescript_unicode.ts")
        .expect("Failed to read Unicode TypeScript file");

    let processor = lang_typescript::TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("unicode.ts"), &opts);

    assert!(
        result.is_ok(),
        "Unicode TypeScript file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Unicode TypeScript: {} classes with Unicode identifiers", class_count);

    assert!(class_count >= 5, "Should find at least 5 classes with Unicode names");
}

#[test]
#[cfg(feature = "lang-go")]
fn test_unicode_go() {
    let source = fs::read_to_string("../../testdata/edge-cases/unicode/go_unicode.go")
        .expect("Failed to read Unicode Go file");

    let processor = lang_go::GoProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("unicode.go"), &opts);

    assert!(result.is_ok(), "Unicode Go file should parse successfully");

    let file = result.unwrap();
    let struct_count = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    println!("✓ Unicode Go: {} structs with Unicode identifiers", struct_count);

    assert!(struct_count >= 5, "Should find at least 5 structs with Unicode names");
}

/// Test 4: Syntax Edge Cases
#[test]
#[cfg(feature = "lang-python")]
fn test_empty_python_file() {
    let source = fs::read_to_string("../../testdata/edge-cases/syntax-edge/empty.py")
        .expect("Failed to read empty Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("empty.py"), &opts);

    assert!(result.is_ok(), "Empty Python file should parse successfully");

    let file = result.unwrap();

    println!("✓ Empty Python file: {} nodes", file.children.len());

    // Empty file should have 0 or very few nodes (maybe just imports/comments)
    assert!(
        file.children.len() <= 1,
        "Empty file should have minimal nodes"
    );
}

#[test]
#[cfg(feature = "lang-python")]
fn test_only_comments_python() {
    let source = fs::read_to_string("../../testdata/edge-cases/syntax-edge/only_comments.py")
        .expect("Failed to read comments-only Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions {
        include_comments: true,
        ..Default::default()
    };

    let result = processor.process(&source, Path::new("comments.py"), &opts);

    assert!(
        result.is_ok(),
        "Comments-only Python file should parse successfully"
    );

    let file = result.unwrap();

    println!("✓ Comments-only Python: {} nodes", file.children.len());

    // Should have minimal actual code nodes
    let code_nodes = file
        .children
        .iter()
        .filter(|n| !matches!(n, distiller_core::ir::Node::Comment(_)))
        .count();

    assert_eq!(code_nodes, 0, "Should have no code nodes, only comments");
}

#[test]
#[cfg(feature = "lang-python")]
fn test_deeply_nested_python() {
    let source = fs::read_to_string("../../testdata/edge-cases/syntax-edge/deeply_nested.py")
        .expect("Failed to read deeply nested Python file");

    let processor = lang_python::PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("nested.py"), &opts);

    assert!(
        result.is_ok(),
        "Deeply nested Python file should parse successfully"
    );

    let file = result.unwrap();

    println!("✓ Deeply nested Python: {} top-level nodes", file.children.len());

    // Should handle deep nesting without stack overflow
    assert!(file.children.len() >= 2, "Should find Level1 class and complex_nesting function");
}

#[test]
#[cfg(feature = "lang-typescript")]
fn test_complex_generics_typescript() {
    let source = fs::read_to_string("../../testdata/edge-cases/syntax-edge/complex_generics.ts")
        .expect("Failed to read complex generics TypeScript file");

    let processor = lang_typescript::TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("generics.ts"), &opts);

    assert!(
        result.is_ok(),
        "Complex generics TypeScript file should parse successfully"
    );

    let file = result.unwrap();

    println!("✓ Complex generics TypeScript: {} top-level nodes", file.children.len());

    // Should handle complex generic constraints
    let classes = file
        .children
        .iter()
        .filter(|n| matches!(n, distiller_core::ir::Node::Class(_)))
        .count();

    assert!(classes >= 2, "Should find GenericManager and GenericStatic classes");
}

/// Test 5: Directory-level Edge Cases
#[test]
fn test_mixed_edge_cases_directory() {
    let registry = create_full_registry();
    let opts = ProcessOptions::default();
    let processor = DirectoryProcessor::new(opts);

    // Test processing entire edge-cases directory
    let result = processor.process("../../testdata/edge-cases", &registry);

    assert!(
        result.is_ok(),
        "Edge cases directory should process successfully"
    );

    let file = result.unwrap();

    println!("✓ Edge cases directory: {} files processed", file.children.len());

    // Should process multiple edge case files
    assert!(
        file.children.len() >= 10,
        "Should find multiple edge case files"
    );
}
