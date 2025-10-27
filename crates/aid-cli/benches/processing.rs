// Performance benchmarks for AI Distiller Rust implementation
//
// Benchmarks process real test files across different languages and complexity levels
// to track performance over time and prevent regressions.

use criterion::{BenchmarkId, Criterion, black_box, criterion_group, criterion_main};
use distiller_core::{ProcessOptions, processor::Processor};
use std::path::{Path, PathBuf};

// Language processors
use lang_go::GoProcessor;
use lang_javascript::JavaScriptProcessor;
use lang_python::PythonProcessor;
use lang_typescript::TypeScriptProcessor;

/// Create a properly configured processor with all languages registered
fn create_configured_processor() -> Processor {
    let options = ProcessOptions {
        include_public: true,
        include_protected: false,
        include_internal: false,
        include_private: false,
        include_comments: false,
        include_docstrings: true,
        include_implementation: false,
        include_imports: true,
        include_annotations: true,
        include_fields: true,
        include_methods: true,
        raw_mode: false,
        workers: 0, // Auto
        recursive: true,
        file_path_type: distiller_core::options::PathType::Relative,
        relative_path_prefix: None,
        base_path: Some(PathBuf::from(".")), // Set base path to current directory
        include_patterns: Vec::new(),
        exclude_patterns: Vec::new(),
    };

    let mut processor = Processor::new(options);

    // Register all language processors
    processor.register_language(Box::new(PythonProcessor::new().expect("Python")));
    processor.register_language(Box::new(TypeScriptProcessor::new().expect("TypeScript")));
    processor.register_language(Box::new(JavaScriptProcessor::new().expect("JavaScript")));
    processor.register_language(Box::new(GoProcessor::new().expect("Go")));

    processor
}

/// Benchmark single file processing across Python complexity levels
fn bench_python_complexity(c: &mut Criterion) {
    let mut group = c.benchmark_group("python_complexity");

    let test_files = vec![
        ("01_basic", "testdata/python/01_basic/source.py"),
        ("02_simple", "testdata/python/02_simple/source.py"),
        ("03_medium", "testdata/python/03_medium/source.py"),
        ("04_complex", "testdata/python/04_complex/source.py"),
        (
            "05_very_complex",
            "testdata/python/05_very_complex/source.py",
        ),
    ];

    for (name, file_path) in test_files {
        let processor = create_configured_processor();
        let path = Path::new(file_path);

        group.bench_with_input(BenchmarkId::from_parameter(name), &path, |b, path| {
            b.iter(|| {
                processor
                    .process_path(black_box(path))
                    .expect("Processing failed")
            });
        });
    }

    group.finish();
}

/// Benchmark TypeScript processing
fn bench_typescript_processing(c: &mut Criterion) {
    let mut group = c.benchmark_group("typescript");

    let test_files = vec![
        ("01_basic", "testdata/typescript/01_basic/source.ts"),
        ("03_medium", "testdata/typescript/03_medium/source.ts"),
        (
            "05_very_complex",
            "testdata/typescript/05_very_complex/source.ts",
        ),
    ];

    for (name, file_path) in test_files {
        let processor = create_configured_processor();
        let path = Path::new(file_path);

        group.bench_with_input(BenchmarkId::from_parameter(name), &path, |b, path| {
            b.iter(|| {
                processor
                    .process_path(black_box(path))
                    .expect("Processing failed")
            });
        });
    }

    group.finish();
}

/// Benchmark Go processing
fn bench_go_processing(c: &mut Criterion) {
    let mut group = c.benchmark_group("go");

    let test_files = vec![
        ("01_basic", "testdata/go/01_basic/source.go"),
        ("03_medium", "testdata/go/03_medium/source.go"),
        ("05_very_complex", "testdata/go/05_very_complex/source.go"),
    ];

    for (name, file_path) in test_files {
        let processor = create_configured_processor();
        let path = Path::new(file_path);

        group.bench_with_input(BenchmarkId::from_parameter(name), &path, |b, path| {
            b.iter(|| {
                processor
                    .process_path(black_box(path))
                    .expect("Processing failed")
            });
        });
    }

    group.finish();
}

/// Benchmark directory processing (real-world scenario)
fn bench_directory_processing(c: &mut Criterion) {
    let mut group = c.benchmark_group("directory");

    // React app with 3 TypeScript files
    let processor = create_configured_processor();
    let path = Path::new("testdata/real-world/react-app");

    group.bench_function("react_app", |b| {
        b.iter(|| {
            processor
                .process_path(black_box(path))
                .expect("Processing failed")
        });
    });

    group.finish();
}

criterion_group!(
    benches,
    bench_python_complexity,
    bench_typescript_processing,
    bench_go_processing,
    bench_directory_processing
);
criterion_main!(benches);
