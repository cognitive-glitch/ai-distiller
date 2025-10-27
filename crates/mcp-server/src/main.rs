//! Simplified MCP Server for AI Distiller
//!
//! Provides 4 core operations via JSON-RPC:
//! 1. distil_directory - Process entire directory
//! 2. distil_file - Process single file
//! 3. list_dir - List directory contents with metadata
//! 4. get_capa - Get server capabilities

use anyhow::{Context, Result};
use distiller_core::{ProcessOptions, ir::*, processor::Processor};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};

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

// Formatters
use formatter_json::JsonFormatter;
use formatter_jsonl::JsonlFormatter;
use formatter_markdown::MarkdownFormatter;
use formatter_text::TextFormatter;
use formatter_xml::XmlFormatter;

/// JSON-RPC request
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct JsonRpcRequest {
    jsonrpc: String,
    id: serde_json::Value,
    method: String,
    params: Option<serde_json::Value>,
}

/// JSON-RPC response
#[derive(Debug, Serialize)]
struct JsonRpcResponse {
    jsonrpc: String,
    id: serde_json::Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<JsonRpcError>,
}

/// JSON-RPC error
#[derive(Debug, Serialize)]
struct JsonRpcError {
    code: i32,
    message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<serde_json::Value>,
}

/// Parameters for distil_directory operation
#[derive(Debug, Deserialize)]
struct DistilDirectoryParams {
    path: PathBuf,
    #[serde(default)]
    options: DistilOptions,
}

/// Parameters for distil_file operation
#[derive(Debug, Deserialize)]
struct DistilFileParams {
    path: PathBuf,
    #[serde(default)]
    options: DistilOptions,
}

/// Parameters for list_dir operation
#[derive(Debug, Deserialize)]
struct ListDirParams {
    path: PathBuf,
    #[serde(default)]
    filters: Option<Vec<String>>,
}

/// Distillation options (simplified from CLI)
#[derive(Debug, Clone, Deserialize, Default)]
struct DistilOptions {
    #[serde(default = "default_true")]
    include_public: bool,
    #[serde(default)]
    include_protected: bool,
    #[serde(default)]
    include_internal: bool,
    #[serde(default)]
    include_private: bool,

    #[serde(default)]
    include_comments: bool,
    #[serde(default = "default_true")]
    include_docstrings: bool,
    #[serde(default)]
    include_implementation: bool,
    #[serde(default = "default_true")]
    include_imports: bool,
    #[serde(default = "default_true")]
    include_annotations: bool,
    #[serde(default = "default_true")]
    include_fields: bool,
    #[serde(default = "default_true")]
    include_methods: bool,

    #[serde(default)]
    format: String, // "text", "md", "json", "jsonl", "xml"
}

fn default_true() -> bool {
    true
}

impl From<DistilOptions> for ProcessOptions {
    fn from(opts: DistilOptions) -> Self {
        ProcessOptions {
            include_public: opts.include_public,
            include_protected: opts.include_protected,
            include_internal: opts.include_internal,
            include_private: opts.include_private,
            include_comments: opts.include_comments,
            include_docstrings: opts.include_docstrings,
            include_implementation: opts.include_implementation,
            include_imports: opts.include_imports,
            include_annotations: opts.include_annotations,
            include_fields: opts.include_fields,
            include_methods: opts.include_methods,
            raw_mode: false,
            workers: 0, // Auto
            recursive: true,
            file_path_type: distiller_core::options::PathType::Relative,
            relative_path_prefix: None,
            base_path: None,
            include_patterns: Vec::new(),
            exclude_patterns: Vec::new(),
        }
    }
}

/// MCP Server state
#[allow(dead_code)]
struct McpServer {
    processor: Processor,
}

impl McpServer {
    /// Create new MCP server with all language processors registered
    fn new() -> Self {
        let mut processor = Processor::new(ProcessOptions::default());
        register_all_languages(&mut processor);

        Self { processor }
    }

    /// Handle distil_directory operation
    async fn handle_distil_directory(&self, params: DistilDirectoryParams) -> Result<String> {
        let path = &params.path;
        if !path.exists() {
            anyhow::bail!("Path does not exist: {}", path.display());
        }
        if !path.is_dir() {
            anyhow::bail!("Path is not a directory: {}", path.display());
        }

        // Update processor options
        let proc_opts: ProcessOptions = params.options.clone().into();
        let processor = Processor::new(proc_opts);
        let mut processor = processor;
        register_all_languages(&mut processor);

        // Process directory
        let node = processor.process_path(path)
            .context("Failed to process directory")?;

        // Extract files
        let files = extract_files(&node);

        if files.is_empty() {
            anyhow::bail!("No files found in directory");
        }

        // Format output
        let format = if params.options.format.is_empty() {
            "text"
        } else {
            &params.options.format
        };

        let output = self.format_files(&files, format)?;
        Ok(output)
    }

    /// Handle distil_file operation
    async fn handle_distil_file(&self, params: DistilFileParams) -> Result<String> {
        let path = &params.path;
        if !path.exists() {
            anyhow::bail!("File does not exist: {}", path.display());
        }
        if !path.is_file() {
            anyhow::bail!("Path is not a file: {}", path.display());
        }

        // Update processor options
        let proc_opts: ProcessOptions = params.options.clone().into();
        let processor = Processor::new(proc_opts);
        let mut processor = processor;
        register_all_languages(&mut processor);

        // Process file
        let node = processor.process_path(path)
            .context("Failed to process file")?;

        // Extract files
        let files = extract_files(&node);

        if files.is_empty() {
            anyhow::bail!("Failed to process file");
        }

        // Format output
        let format = if params.options.format.is_empty() {
            "text"
        } else {
            &params.options.format
        };

        let output = self.format_files(&files, format)?;
        Ok(output)
    }

    /// Handle list_dir operation
    async fn handle_list_dir(&self, params: ListDirParams) -> Result<Vec<FileInfo>> {
        let path = &params.path;
        if !path.exists() {
            anyhow::bail!("Path does not exist: {}", path.display());
        }
        if !path.is_dir() {
            anyhow::bail!("Path is not a directory: {}", path.display());
        }

        let mut entries = Vec::new();

        for entry in std::fs::read_dir(path)? {
            let entry = entry?;
            let path = entry.path();
            let metadata = entry.metadata()?;

            // Check filters if provided
            if let Some(ref filters) = params.filters {
                let filename = path.file_name()
                    .and_then(|n| n.to_str())
                    .unwrap_or("");

                let matches = filters.iter().any(|filter| {
                    filename.contains(filter)
                });

                if !matches {
                    continue;
                }
            }

            let info = FileInfo {
                path: path.display().to_string(),
                is_file: metadata.is_file(),
                is_dir: metadata.is_dir(),
                size: metadata.len(),
            };
            entries.push(info);
        }

        Ok(entries)
    }

    /// Handle get_capa operation
    async fn handle_get_capa(&self) -> Result<ServerCapabilities> {
        Ok(ServerCapabilities {
            version: env!("CARGO_PKG_VERSION").to_string(),
            operations: vec![
                "distil_directory".to_string(),
                "distil_file".to_string(),
                "list_dir".to_string(),
                "get_capa".to_string(),
            ],
            supported_languages: vec![
                "Python", "TypeScript", "JavaScript", "Go", "Rust",
                "Java", "Kotlin", "Swift", "Ruby", "PHP", "C#", "C++", "C",
            ].into_iter().map(String::from).collect(),
            supported_formats: vec![
                "text", "md", "json", "jsonl", "xml",
            ].into_iter().map(String::from).collect(),
        })
    }

    /// Format files using specified formatter
    fn format_files(&self, files: &[File], format: &str) -> Result<String> {
        match format {
            "text" => {
                let formatter = TextFormatter::new();
                formatter.format_files(files)
                    .context("Failed to format as text")
            }
            "md" | "markdown" => {
                let formatter = MarkdownFormatter::new();
                formatter.format_files(files)
                    .context("Failed to format as markdown")
            }
            "json" => {
                let formatter = JsonFormatter::new();
                formatter.format_files(files)
                    .context("Failed to format as JSON")
            }
            "jsonl" => {
                let formatter = JsonlFormatter::new();
                formatter.format_files(files)
                    .context("Failed to format as JSONL")
            }
            "xml" => {
                let formatter = XmlFormatter::new();
                formatter.format_files(files)
                    .context("Failed to format as XML")
            }
            _ => anyhow::bail!("Unsupported format: {}", format),
        }
    }
}

/// File information for list_dir response
#[derive(Debug, Serialize)]
struct FileInfo {
    path: String,
    is_file: bool,
    is_dir: bool,
    size: u64,
}

/// Server capabilities
#[derive(Debug, Serialize)]
struct ServerCapabilities {
    version: String,
    operations: Vec<String>,
    supported_languages: Vec<String>,
    supported_formats: Vec<String>,
}

/// Extract File nodes from an IR Node (recursive for Directory)
fn extract_files(node: &Node) -> Vec<File> {
    let mut files = Vec::new();

    match node {
        Node::File(file) => {
            files.push(file.clone());
        }
        Node::Directory(dir) => {
            for child in &dir.children {
                files.extend(extract_files(child));
            }
        }
        _ => {
            // Other node types don't contain files
        }
    }

    files
}

/// Register all supported language processors
fn register_all_languages(processor: &mut Processor) {
    // Python
    processor.register_language(Box::new(PythonProcessor::new().expect("Failed to create PythonProcessor")));

    // TypeScript/JavaScript
    processor.register_language(Box::new(TypeScriptProcessor::new().expect("Failed to create TypeScriptProcessor")));
    processor.register_language(Box::new(JavaScriptProcessor::new().expect("Failed to create JavaScriptProcessor")));

    // Systems languages
    processor.register_language(Box::new(RustProcessor::new().expect("Failed to create RustProcessor")));
    processor.register_language(Box::new(CppProcessor::new().expect("Failed to create CppProcessor")));
    processor.register_language(Box::new(CProcessor::new().expect("Failed to create CProcessor")));
    processor.register_language(Box::new(GoProcessor::new().expect("Failed to create GoProcessor")));

    // JVM languages
    processor.register_language(Box::new(JavaProcessor::new().expect("Failed to create JavaProcessor")));
    processor.register_language(Box::new(KotlinProcessor::new().expect("Failed to create KotlinProcessor")));

    // .NET languages
    processor.register_language(Box::new(CSharpProcessor::new().expect("Failed to create CSharpProcessor")));

    // Other languages
    processor.register_language(Box::new(SwiftProcessor::new().expect("Failed to create SwiftProcessor")));
    processor.register_language(Box::new(RubyProcessor::new().expect("Failed to create RubyProcessor")));
    processor.register_language(Box::new(PhpProcessor::new().expect("Failed to create PhpProcessor")));
}

/// Helper function to send a JSON-RPC response
async fn send_response(stdout: &mut tokio::io::Stdout, response: &JsonRpcResponse) -> Result<()> {
    let response_json = serde_json::to_string(response)?;
    stdout.write_all(response_json.as_bytes()).await?;
    stdout.write_all(b"\n").await?;
    stdout.flush().await?;
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    // Setup logging
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("info")).init();

    log::info!("🚀 MCP Server v{} starting...", env!("CARGO_PKG_VERSION"));

    let server = McpServer::new();
    log::info!("✅ Server initialized with 13 language processors");
    log::info!("📡 Listening for JSON-RPC requests on stdin...");

    // Read JSON-RPC requests from stdin
    let stdin = tokio::io::stdin();
    let mut reader = BufReader::new(stdin);
    let mut stdout = tokio::io::stdout();

    let mut line = String::new();

    loop {
        line.clear();

        match reader.read_line(&mut line).await {
            Ok(0) => {
                // EOF
                log::info!("📪 Received EOF, shutting down...");
                break;
            }
            Ok(_) => {
                let line = line.trim();
                if line.is_empty() {
                    continue;
                }

                // Parse JSON-RPC request
                let request: JsonRpcRequest = match serde_json::from_str(line) {
                    Ok(req) => req,
                    Err(e) => {
                        log::error!("❌ Failed to parse JSON-RPC request: {}", e);
                        continue;
                    }
                };

                log::info!("📥 Received request: method={}, id={:?}", request.method, request.id);

                // Handle request based on method
                let response = match request.method.as_str() {
                    "distil_directory" => {
                        // Parse params
                        let params: DistilDirectoryParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null)
                        ) {
                            Ok(p) => p,
                            Err(e) => {
                                // Send error response and continue to next request
                                let error_response = JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id.clone(),
                                    result: None,
                                    error: Some(JsonRpcError {
                                        code: -32602,
                                        message: format!("Invalid params: {}", e),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_distil_directory(params).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::Value::String(result)),
                                error: None,
                            },
                            Err(e) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: None,
                                error: Some(JsonRpcError {
                                    code: -32000,
                                    message: e.to_string(),
                                    data: None,
                                }),
                            },
                        }
                    }
                    "distil_file" => {
                        // Parse params
                        let params: DistilFileParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null)
                        ) {
                            Ok(p) => p,
                            Err(e) => {
                                // Send error response and continue to next request
                                let error_response = JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id.clone(),
                                    result: None,
                                    error: Some(JsonRpcError {
                                        code: -32602,
                                        message: format!("Invalid params: {}", e),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_distil_file(params).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::Value::String(result)),
                                error: None,
                            },
                            Err(e) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: None,
                                error: Some(JsonRpcError {
                                    code: -32000,
                                    message: e.to_string(),
                                    data: None,
                                }),
                            },
                        }
                    }
                    "list_dir" => {
                        // Parse params
                        let params: ListDirParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null)
                        ) {
                            Ok(p) => p,
                            Err(e) => {
                                // Send error response and continue to next request
                                let error_response = JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id.clone(),
                                    result: None,
                                    error: Some(JsonRpcError {
                                        code: -32602,
                                        message: format!("Invalid params: {}", e),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_list_dir(params).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::to_value(result).unwrap()),
                                error: None,
                            },
                            Err(e) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: None,
                                error: Some(JsonRpcError {
                                    code: -32000,
                                    message: e.to_string(),
                                    data: None,
                                }),
                            },
                        }
                    }
                    "get_capa" => {
                        // No params needed for get_capa
                        match server.handle_get_capa().await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::to_value(result).unwrap()),
                                error: None,
                            },
                            Err(e) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: None,
                                error: Some(JsonRpcError {
                                    code: -32000,
                                    message: e.to_string(),
                                    data: None,
                                }),
                            },
                        }
                    }
                    _ => JsonRpcResponse {
                        jsonrpc: "2.0".to_string(),
                        id: request.id,
                        result: None,
                        error: Some(JsonRpcError {
                            code: -32601,
                            message: format!("Method not found: {}", request.method),
                            data: None,
                        }),
                    },
                };

                // Send response
                send_response(&mut stdout, &response).await?;
                log::info!("📤 Sent response for id={:?}", response.id);
            }
            Err(e) => {
                log::error!("❌ Failed to read from stdin: {}", e);
                break;
            }
        }
    }

    log::info!("👋 MCP Server shutting down...");
    Ok(())
}
