//! AI Distiller CLI
//!
//! Extract code structure for AI consumption.

use anyhow::{Context, Result};
use clap::{Parser, ValueEnum};
use distiller_core::{ProcessOptions, processor::Processor};
use std::path::{Path, PathBuf};

// Language processors
use lang_c::CProcessor;
use lang_cpp::CppProcessor;
use lang_csharp::CSharpProcessor;
use lang_go::GoProcessor;
use lang_java::JavaProcessor;
use lang_javascript::JavaScriptProcessor;
use lang_kotlin::KotlinProcessor;
use lang_php::PhpProcessor;
use lang_python::PythonProcessor;
use lang_ruby::RubyProcessor;
use lang_rust::RustProcessor;
use lang_swift::SwiftProcessor;
use lang_typescript::TypeScriptProcessor;

/// Output format selection
#[derive(Debug, Clone, Copy, ValueEnum)]
enum Format {
    /// Ultra-compact text format (best for AI)
    Text,
    /// Markdown format with syntax highlighting
    Md,
    /// JSON format with structured data
    Json,
    /// JSON Lines format (one object per line)
    Jsonl,
    /// XML format for legacy tools
    Xml,
}

impl Default for Format {
    fn default() -> Self {
        Self::Text
    }
}

#[derive(Parser, Debug)]
#[command(
    name = "aid",
    version,
    about = "AI Distiller - Extract code structure for AI",
    long_about = "AI Distiller extracts essential code structure from large codebases,\n\
                  making them digestible for LLMs by removing unnecessary details while\n\
                  preserving semantic information."
)]
struct Args {
    /// Path to file or directory
    #[arg(value_name = "PATH")]
    path: Option<PathBuf>,

    // Output options
    /// Output format
    #[arg(short = 'f', long, value_enum, default_value = "text")]
    format: Format,

    /// Output file (auto-generated if not specified)
    #[arg(short, long)]
    output: Option<PathBuf>,

    /// Print to stdout instead of file
    #[arg(long)]
    stdout: bool,

    // Visibility filtering
    /// Include public members
    #[arg(long, default_value = "true")]
    public: bool,

    /// Include protected members
    #[arg(long, default_value = "true")]
    protected: bool,

    /// Include internal/package-private members
    #[arg(long)]
    internal: bool,

    /// Include private members
    #[arg(long)]
    private: bool,

    // Content filtering
    /// Include regular comments
    #[arg(long, default_value = "true")]
    comments: bool,

    /// Include documentation comments/docstrings
    #[arg(long, default_value = "true")]
    docstrings: bool,

    /// Include function/method implementations
    #[arg(long, default_value = "true")]
    implementation: bool,

    /// Include import statements
    #[arg(long, default_value = "true")]
    imports: bool,

    /// Include annotations/decorators
    #[arg(long, default_value = "true")]
    annotations: bool,

    /// Include class fields/properties
    #[arg(long, default_value = "true")]
    fields: bool,

    /// Include methods/functions
    #[arg(long, default_value = "true")]
    methods: bool,

    // Processing options
    /// Number of worker threads (0 = auto: 80% CPU cores)
    #[arg(short = 'w', long, default_value = "0")]
    workers: usize,

    /// Process directories recursively
    #[arg(short = 'r', long, default_value = "true")]
    recursive: bool,

    /// Include only files matching these patterns (comma-separated)
    #[arg(long)]
    include: Option<String>,

    /// Exclude files matching these patterns (comma-separated)
    #[arg(long)]
    exclude: Option<String>,

    /// Continue processing on errors (collect partial results)
    #[arg(long)]
    continue_on_error: bool,

    // Formatter-specific options
    /// Pretty-print JSON output (JSON formatter only)
    #[arg(long)]
    pretty: bool,

    /// XML indentation spaces (XML formatter only)
    #[arg(long, default_value = "2")]
    indent: usize,

    /// Verbosity level (-v, -vv, -vvv)
    #[arg(short, long, action = clap::ArgAction::Count)]
    verbose: u8,
}

impl Args {
    /// Convert CLI args to `ProcessOptions`
    fn to_process_options(&self) -> ProcessOptions {
        let mut options = ProcessOptions {
            include_public: self.public,
            include_protected: self.protected,
            include_internal: self.internal,
            include_private: self.private,
            include_comments: self.comments,
            include_docstrings: self.docstrings,
            include_implementation: self.implementation,
            include_imports: self.imports,
            include_annotations: self.annotations,
            include_fields: self.fields,
            include_methods: self.methods,
            workers: self.workers,
            recursive: self.recursive,
            continue_on_error: self.continue_on_error,
            ..Default::default()
        };

        // Pattern filtering
        if let Some(ref include) = self.include {
            options.include_patterns = include.split(',').map(|s| s.trim().to_string()).collect();
        }
        if let Some(ref exclude) = self.exclude {
            options.exclude_patterns = exclude.split(',').map(|s| s.trim().to_string()).collect();
        }

        options
    }
}

fn main() -> Result<()> {
    let args = Args::parse();

    // Setup logging with unified helper
    distiller_core::logging::init_logging(args.verbose);

    log::info!("🦀 AI Distiller v{} (Rust)", env!("CARGO_PKG_VERSION"));

    // Require path argument
    let path = args.path.as_ref().context("PATH argument is required")?;

    // Validate path exists
    if !path.exists() {
        anyhow::bail!("Path does not exist: {}", path.display());
    }

    log::info!("Processing: {}", path.display());
    log::debug!("Format: {:?}", args.format);
    log::debug!("Workers: {}", args.workers);

    // Step 1: Create processor with options
    let options = args.to_process_options();
    let processor = Processor::new(options.clone());

    // Register all language processors
    let mut processor = processor;
    register_all_languages(&mut processor);

    // Step 2: Process path to get IR
    let mut node = processor
        .process_path(path)
        .context("Failed to process path")?;

    // Step 2.5: Apply stripper to filter IR based on options (skip if raw_mode)
    if !options.raw_mode {
        use distiller_core::ir::Visitor;
        use distiller_core::stripper::Stripper;
        let mut stripper = Stripper::new(options.clone());
        stripper.visit_node(&mut node);
    }

    // Step 3: Extract files from IR node
    let files = distiller_core::ir::extract_files(&node);

    if files.is_empty() {
        anyhow::bail!("No files found to format");
    }

    log::info!("Formatting {} file(s)...", files.len());

    // Step 4: Format output based on selected format
    let output = match args.format {
        Format::Text => {
            use formatter_text::{TextFormatter, TextFormatterOptions};
            let formatter_opts = TextFormatterOptions {
                include_implementation: options.include_implementation,
            };
            let formatter = TextFormatter::with_options(formatter_opts);
            formatter
                .format_files(&files)
                .context("Failed to format as text")?
        }
        Format::Md => {
            use formatter_markdown::MarkdownFormatter;
            use formatter_text::TextFormatterOptions;
            let formatter_opts = TextFormatterOptions {
                include_implementation: options.include_implementation,
            };
            let formatter = MarkdownFormatter::with_options(formatter_opts);
            formatter
                .format_files(&files)
                .context("Failed to format as markdown")?
        }
        Format::Json => {
            use formatter_json::{JsonFormatter, JsonFormatterOptions};
            let opts = JsonFormatterOptions {
                pretty: args.pretty,
            };
            let formatter = JsonFormatter::with_options(opts);
            formatter
                .format_files(&files)
                .context("Failed to format as JSON")?
        }
        Format::Jsonl => {
            use formatter_jsonl::JsonlFormatter;
            let formatter = JsonlFormatter::new();
            formatter
                .format_files(&files)
                .context("Failed to format as JSONL")?
        }
        Format::Xml => {
            use formatter_xml::{XmlFormatter, XmlFormatterOptions};
            let opts = XmlFormatterOptions {
                indent: args.indent > 0,
                indent_size: args.indent,
            };
            let formatter = XmlFormatter::with_options(opts);
            formatter
                .format_files(&files)
                .context("Failed to format as XML")?
        }
    };

    // Step 5: Write output
    if args.stdout {
        println!("{output}");
        log::info!("Output written to stdout");
    } else {
        let output_path = if let Some(ref path) = args.output {
            path.clone()
        } else {
            // Auto-generate output filename
            generate_output_path(path, args.format)?
        };

        std::fs::write(&output_path, output).context(format!(
            "Failed to write output to {}",
            output_path.display()
        ))?;

        println!("✨ Output written to: {}", output_path.display());
        log::info!("Output written to: {}", output_path.display());
    }

    Ok(())
}

/// Generate automatic output filename based on input path and format
fn generate_output_path(input: &Path, format: Format) -> Result<PathBuf> {
    let extension = match format {
        Format::Text => "txt",
        Format::Md => "md",
        Format::Json => "json",
        Format::Jsonl => "jsonl",
        Format::Xml => "xml",
    };

    let basename = if input.is_file() {
        input
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("output")
    } else {
        input
            .file_name()
            .and_then(|s| s.to_str())
            .unwrap_or("output")
    };

    // Add timestamp to avoid collisions
    let timestamp = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    // Create .aid/ directory if it doesn't exist
    let aid_dir = PathBuf::from(".aid");
    if !aid_dir.exists() {
        std::fs::create_dir(&aid_dir).context("Failed to create .aid/ directory")?;
    }

    Ok(aid_dir.join(format!("{basename}.{timestamp}.{extension}")))
}

/// Register all supported language processors
fn register_all_languages(processor: &mut Processor) {
    // Python
    processor.register_language(Box::new(
        PythonProcessor::new().expect("Failed to create PythonProcessor"),
    ));

    // TypeScript/JavaScript
    processor.register_language(Box::new(
        TypeScriptProcessor::new().expect("Failed to create TypeScriptProcessor"),
    ));
    processor.register_language(Box::new(
        JavaScriptProcessor::new().expect("Failed to create JavaScriptProcessor"),
    ));

    // Systems languages
    processor.register_language(Box::new(
        RustProcessor::new().expect("Failed to create RustProcessor"),
    ));
    processor.register_language(Box::new(
        CppProcessor::new().expect("Failed to create CppProcessor"),
    ));
    processor.register_language(Box::new(
        CProcessor::new().expect("Failed to create CProcessor"),
    ));
    processor.register_language(Box::new(
        GoProcessor::new().expect("Failed to create GoProcessor"),
    ));

    // JVM languages
    processor.register_language(Box::new(
        JavaProcessor::new().expect("Failed to create JavaProcessor"),
    ));
    processor.register_language(Box::new(
        KotlinProcessor::new().expect("Failed to create KotlinProcessor"),
    ));

    // .NET languages
    processor.register_language(Box::new(
        CSharpProcessor::new().expect("Failed to create CSharpProcessor"),
    ));

    // Other languages
    processor.register_language(Box::new(
        SwiftProcessor::new().expect("Failed to create SwiftProcessor"),
    ));
    processor.register_language(Box::new(
        RubyProcessor::new().expect("Failed to create RubyProcessor"),
    ));
    processor.register_language(Box::new(
        PhpProcessor::new().expect("Failed to create PhpProcessor"),
    ));
}
