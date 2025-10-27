//! Integration tests for DirectoryProcessor
//!
//! Tests multi-file, multi-language processing scenarios.

use distiller_core::{
    ir::Node,
    options::ProcessOptions,
    processor::directory::{DirectoryProcessor, LanguageRegistry},
};
use std::path::Path;

/// Create a fully-configured language registry
fn create_full_registry() -> LanguageRegistry {
    let mut registry = LanguageRegistry::new();

    // Register all available language processors
    registry.register(Box::new(lang_python::PythonProcessor::new().unwrap()));
    registry.register(Box::new(
        lang_typescript::TypeScriptProcessor::new().unwrap(),
    ));
    registry.register(Box::new(lang_go::GoProcessor::new().unwrap()));
    registry.register(Box::new(lang_c::CProcessor::new().unwrap()));

    registry
}

#[test]
fn test_mixed_language_directory() {
    let registry = create_full_registry();
    let opts = ProcessOptions::default();
    let processor = DirectoryProcessor::new(opts);

    let test_dir = Path::new("../../../testdata/integration/mixed");

    if !test_dir.exists() {
        eprintln!("Skipping test: directory not found: {}", test_dir.display());
        return;
    }

    let result = processor.process(test_dir, &registry);

    match result {
        Ok(directory) => {
            // Should process Python, TypeScript, and Go files
            let file_count = directory.children.len();
            assert!(
                file_count >= 3,
                "Expected at least 3 files (py, ts, go), got {}",
                file_count
            );

            // Verify all files were processed
            let has_python = directory.children.iter().any(|node| {
                if let Node::File(f) = node {
                    f.path.ends_with(".py")
                } else {
                    false
                }
            });

            let has_typescript = directory.children.iter().any(|node| {
                if let Node::File(f) = node {
                    f.path.ends_with(".ts")
                } else {
                    false
                }
            });

            let has_go = directory.children.iter().any(|node| {
                if let Node::File(f) = node {
                    f.path.ends_with(".go")
                } else {
                    false
                }
            });

            assert!(has_python, "Should process Python file");
            assert!(has_typescript, "Should process TypeScript file");
            assert!(has_go, "Should process Go file");

            println!(
                "✅ Mixed language processing: {} files processed",
                file_count
            );
        }
        Err(e) => {
            panic!("Failed to process mixed directory: {}", e);
        }
    }
}

#[test]
fn test_option_combinations() {
    let registry = create_full_registry();

    // Test 1: Default options (public only)
    let opts_default = ProcessOptions::default();
    let processor_default = DirectoryProcessor::new(opts_default);

    // Test 2: With implementations
    let opts_impl = ProcessOptions {
        include_implementation: true,
        ..Default::default()
    };
    let processor_impl = DirectoryProcessor::new(opts_impl);

    // Test 3: With private members
    let opts_private = ProcessOptions {
        include_private: true,
        ..Default::default()
    };
    let processor_private = DirectoryProcessor::new(opts_private);

    let test_dir = Path::new("../../../testdata/python/01_basic");

    if !test_dir.exists() {
        eprintln!("Skipping test: directory not found");
        return;
    }

    // Process with different options
    let result_default = processor_default.process(test_dir, &registry);
    let result_impl = processor_impl.process(test_dir, &registry);
    let result_private = processor_private.process(test_dir, &registry);

    assert!(result_default.is_ok(), "Default options should work");
    assert!(result_impl.is_ok(), "Implementation options should work");
    assert!(result_private.is_ok(), "Private options should work");

    println!("✅ Option combinations work correctly");
}

#[test]
fn test_empty_directory_handling() {
    let registry = create_full_registry();
    let opts = ProcessOptions::default();
    let processor = DirectoryProcessor::new(opts);

    // Create a temporary empty directory
    let temp_dir = std::env::temp_dir().join("aid_test_empty");
    let _ = std::fs::create_dir(&temp_dir);

    let result = processor.process(&temp_dir, &registry);

    // Cleanup
    let _ = std::fs::remove_dir(&temp_dir);

    match result {
        Ok(directory) => {
            assert_eq!(
                directory.children.len(),
                0,
                "Empty directory should have 0 files"
            );
            println!("✅ Empty directory handled correctly");
        }
        Err(e) => {
            panic!("Empty directory should not error: {}", e);
        }
    }
}

#[test]
fn test_non_directory_error() {
    let registry = create_full_registry();
    let opts = ProcessOptions::default();
    let processor = DirectoryProcessor::new(opts);

    // Try to process a file (not a directory)
    let result = processor.process(Path::new("Cargo.toml"), &registry);

    assert!(result.is_err(), "Processing a file should error");
    println!("✅ Non-directory error handling works");
}

#[test]
fn test_parallel_processing_consistency() {
    let registry = create_full_registry();

    // Process the same directory multiple times
    // Results should be consistent (rayon parallelism shouldn't cause randomness)
    let test_dir = Path::new("../../../testdata/python");

    if !test_dir.exists() {
        eprintln!("Skipping test: directory not found");
        return;
    }

    let opts = ProcessOptions::default();

    let processor1 = DirectoryProcessor::new(opts.clone());
    let processor2 = DirectoryProcessor::new(opts.clone());
    let processor3 = DirectoryProcessor::new(opts);

    let result1 = processor1.process(test_dir, &registry).unwrap();
    let result2 = processor2.process(test_dir, &registry).unwrap();
    let result3 = processor3.process(test_dir, &registry).unwrap();

    // File counts should be identical
    assert_eq!(
        result1.children.len(),
        result2.children.len(),
        "Parallel processing should be deterministic"
    );
    assert_eq!(
        result2.children.len(),
        result3.children.len(),
        "Parallel processing should be deterministic"
    );

    // File order should be preserved (sorted by discovery order)
    for (i, (node1, node2)) in result1
        .children
        .iter()
        .zip(result2.children.iter())
        .enumerate()
    {
        if let (Node::File(f1), Node::File(f2)) = (node1, node2) {
            assert_eq!(
                f1.path, f2.path,
                "File order should be consistent at index {}",
                i
            );
        }
    }

    println!("✅ Parallel processing is consistent across runs");
}

#[test]
fn test_recursive_vs_non_recursive() {
    let registry = create_full_registry();

    // Test recursive processing (default)
    let opts_recursive = ProcessOptions {
        recursive: true,
        ..Default::default()
    };
    let processor_recursive = DirectoryProcessor::new(opts_recursive);

    // Test non-recursive processing
    let opts_non_recursive = ProcessOptions {
        recursive: false,
        ..Default::default()
    };
    let processor_non_recursive = DirectoryProcessor::new(opts_non_recursive);

    let test_dir = Path::new("../../../testdata/python");

    if !test_dir.exists() {
        eprintln!("Skipping test: directory not found");
        return;
    }

    let result_recursive = processor_recursive.process(test_dir, &registry);
    let result_non_recursive = processor_non_recursive.process(test_dir, &registry);

    if let (Ok(dir_recursive), Ok(dir_non_recursive)) = (result_recursive, result_non_recursive) {
        // Recursive should find more files (subdirectories)
        assert!(
            dir_recursive.children.len() >= dir_non_recursive.children.len(),
            "Recursive should find at least as many files as non-recursive"
        );

        println!(
            "✅ Recursive: {} files, Non-recursive: {} files",
            dir_recursive.children.len(),
            dir_non_recursive.children.len()
        );
    }
}

#[test]
fn test_c_language_processing() {
    let registry = create_full_registry();
    let opts = ProcessOptions::default();
    let processor = DirectoryProcessor::new(opts);

    let test_dir = Path::new("../../../testdata/c");

    if !test_dir.exists() {
        eprintln!("Skipping test: C testdata directory not found");
        return;
    }

    let result = processor.process(test_dir, &registry);

    match result {
        Ok(directory) => {
            let file_count = directory.children.len();
            assert!(
                file_count >= 5,
                "Expected at least 5 C test files, got {}",
                file_count
            );

            // Verify C files were processed
            let c_files: Vec<_> = directory
                .children
                .iter()
                .filter_map(|node| {
                    if let Node::File(f) = node {
                        if f.path.ends_with(".c") {
                            Some(f)
                        } else {
                            None
                        }
                    } else {
                        None
                    }
                })
                .collect();

            assert!(!c_files.is_empty(), "Should process at least one C file");

            // Verify C files contain expected structures
            for file in c_files {
                let has_content = !file.children.is_empty();
                assert!(
                    has_content,
                    "C file {} should have parsed content",
                    file.path
                );
            }

            println!("✅ C language processing: {} files processed", file_count);
        }
        Err(e) => {
            panic!("Failed to process C directory: {}", e);
        }
    }
}
