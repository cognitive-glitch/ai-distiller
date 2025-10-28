//! Simplified MCP Server for AI Distiller
//!
//! Provides 4 core operations via JSON-RPC:
//! 1. `distil_directory` - Process entire directory
//! 2. `distil_file` - Process single file
//! 3. `list_dir` - List directory contents with metadata
//! 4. `get_capa` - Get server capabilities

use anyhow::{Context, Result};
use distiller_core::{
    ProcessOptions,
    error::DistilError,
    ir::{File, Node, Visitor},
    processor::Processor,
    stripper::Stripper,
};
use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};

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

/// Maximum request body size (16MB) to prevent memory abuse
const MAX_BODY_BYTES: usize = 16_777_216;

/// JSON-RPC 2.0 standard error codes
/// See: https://www.jsonrpc.org/specification#error_object
#[allow(dead_code)]
const ERROR_PARSE_ERROR: i32 = -32700; // Invalid JSON
#[allow(dead_code)]
const ERROR_INVALID_REQUEST: i32 = -32600; // Invalid Request object
const ERROR_METHOD_NOT_FOUND: i32 = -32601; // Method does not exist
const ERROR_INVALID_PARAMS: i32 = -32602; // Invalid method parameters

/// Server-defined error codes (reserved range: -32000 to -32099)
const ERROR_FILE_NOT_FOUND: i32 = -32001; // File or directory not found
const ERROR_PROCESSING_FAILED: i32 = -32002; // Processing operation failed
#[allow(dead_code)]
const ERROR_PATH_VALIDATION: i32 = -32003; // Path validation failed

/// JSON-RPC error
#[derive(Debug, Serialize)]
struct JsonRpcError {
    code: i32,
    message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<serde_json::Value>,
}

impl JsonRpcError {
    #[allow(dead_code)]
    fn parse_error(message: String) -> Self {
        Self {
            code: ERROR_PARSE_ERROR,
            message,
            data: None,
        }
    }

    fn invalid_params(message: String) -> Self {
        Self {
            code: ERROR_INVALID_PARAMS,
            message,
            data: None,
        }
    }

    fn method_not_found(method: String) -> Self {
        Self {
            code: ERROR_METHOD_NOT_FOUND,
            message: format!("Method not found: {}", method),
            data: None,
        }
    }

    fn file_not_found(path: String) -> Self {
        Self {
            code: ERROR_FILE_NOT_FOUND,
            message: "File or directory not found".to_string(),
            data: Some(serde_json::json!({ "path": path })),
        }
    }

    #[allow(dead_code)]
    fn path_validation_error(message: String) -> Self {
        Self {
            code: ERROR_PATH_VALIDATION,
            message,
            data: None,
        }
    }

    fn processing_failed(message: String, path: Option<String>) -> Self {
        let mut data = serde_json::Map::new();
        if let Some(p) = path {
            data.insert("path".to_string(), serde_json::Value::String(p));
        }
        Self {
            code: ERROR_PROCESSING_FAILED,
            message,
            data: if data.is_empty() {
                None
            } else {
                Some(serde_json::Value::Object(data))
            },
        }
    }
}

/// Parameters for `distil_directory` operation
#[derive(Debug, Clone, Deserialize)]
struct DistilDirectoryParams {
    path: PathBuf,
    #[serde(default)]
    options: DistilOptions,
}

/// Parameters for `distil_file` operation
#[derive(Debug, Clone, Deserialize)]
struct DistilFileParams {
    path: PathBuf,
    #[serde(default)]
    options: DistilOptions,
}

/// Parameters for `list_dir` operation
#[derive(Debug, Clone, Deserialize)]
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
            continue_on_error: false,
        }
    }
}

/// MCP Server state
#[allow(dead_code)]
struct McpServer {
    // Note: Processor is created per-request with custom ProcessOptions
    // Language processor registration is cheap, so we don't cache
}

impl McpServer {
    /// Create new MCP server
    fn new() -> Self {
        Self {}
    }

    /// Create processor with all language processors registered
    fn create_processor(&self, options: ProcessOptions) -> Processor {
        let mut processor = Processor::new(options);
        register_all_languages(&mut processor);
        processor
    }

    /// Validate path to prevent directory traversal attacks
    fn validate_path(&self, path: &Path) -> Result<PathBuf> {
        // Canonicalize to resolve symlinks and .. components
        let canonical = path
            .canonicalize()
            .map_err(|_| DistilError::FileNotFound(path.to_path_buf()))?;

        // Additional security: ensure path doesn't escape to sensitive directories
        let path_str = canonical.to_string_lossy();
        if path_str.contains("/etc/") || path_str.contains("/sys/") || path_str.contains("/proc/") {
            return Err(DistilError::InvalidConfig(
                "Access to system directories is not allowed".to_string(),
            )
            .into());
        }

        Ok(canonical)
    }

    /// Handle `distil_directory` operation
    async fn handle_distil_directory(&self, params: DistilDirectoryParams) -> Result<String> {
        let path = &params.path;

        // Validate path for security
        let validated_path = self.validate_path(path)?;

        if !validated_path.is_dir() {
            anyhow::bail!("Path is not a directory: {}", validated_path.display());
        }

        // Update processor options
        let proc_opts: ProcessOptions = params.options.clone().into();
        let processor = self.create_processor(proc_opts.clone());

        // Process directory
        let mut node = processor
            .process_path(&validated_path)
            .context("Failed to process directory")?;

        // Apply stripper to match CLI filtering behavior
        let mut stripper = Stripper::new(proc_opts);
        stripper.visit_node(&mut node);

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

    /// Handle `distil_file` operation
    async fn handle_distil_file(&self, params: DistilFileParams) -> Result<String> {
        let path = &params.path;

        // Validate path for security
        let validated_path = self.validate_path(path)?;

        if !validated_path.is_file() {
            anyhow::bail!("Path is not a file: {}", validated_path.display());
        }

        // Update processor options
        let proc_opts: ProcessOptions = params.options.clone().into();
        let processor = self.create_processor(proc_opts.clone());

        // Process file
        let mut node = processor
            .process_path(&validated_path)
            .context("Failed to process file")?;

        // Apply stripper to match CLI filtering behavior
        let mut stripper = Stripper::new(proc_opts);
        stripper.visit_node(&mut node);

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

    /// Handle `list_dir` operation
    async fn handle_list_dir(&self, params: ListDirParams) -> Result<Vec<FileInfo>> {
        let path = &params.path;

        // Validate path for security
        let validated_path = self.validate_path(path)?;

        if !validated_path.is_dir() {
            anyhow::bail!("Path is not a directory: {}", validated_path.display());
        }

        let mut entries = Vec::new();

        for entry in std::fs::read_dir(&validated_path)? {
            let entry = entry?;
            let path = entry.path();
            let metadata = entry.metadata()?;

            // Check filters if provided
            if let Some(ref filters) = params.filters {
                let filename = path.file_name().and_then(|n| n.to_str()).unwrap_or("");

                let matches = filters.iter().any(|filter| filename.contains(filter));

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

    /// Handle `get_capa` operation
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
                "Python",
                "TypeScript",
                "JavaScript",
                "Go",
                "Rust",
                "Java",
                "Kotlin",
                "Swift",
                "Ruby",
                "PHP",
                "C#",
                "C++",
                "C",
            ]
            .into_iter()
            .map(String::from)
            .collect(),
            supported_formats: vec!["text", "md", "json", "jsonl", "xml"]
                .into_iter()
                .map(String::from)
                .collect(),
        })
    }

    /// Format files using specified formatter
    fn format_files(&self, files: &[File], format: &str) -> Result<String> {
        match format {
            "text" => {
                let formatter = TextFormatter::new();
                formatter
                    .format_files(files)
                    .context("Failed to format as text")
            }
            "md" | "markdown" => {
                let formatter = MarkdownFormatter::new();
                formatter
                    .format_files(files)
                    .context("Failed to format as markdown")
            }
            "json" => {
                let formatter = JsonFormatter::new();
                formatter
                    .format_files(files)
                    .context("Failed to format as JSON")
            }
            "jsonl" => {
                let formatter = JsonlFormatter::new();
                formatter
                    .format_files(files)
                    .context("Failed to format as JSONL")
            }
            "xml" => {
                let formatter = XmlFormatter::new();
                formatter
                    .format_files(files)
                    .context("Failed to format as XML")
            }
            _ => anyhow::bail!("Unsupported format: {format}"),
        }
    }
}

/// File information for `list_dir` response
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

/// Helper function to send a JSON-RPC response with Content-Length framing
async fn send_response(stdout: &mut tokio::io::Stdout, response: &JsonRpcResponse) -> Result<()> {
    let response_json = serde_json::to_string(response)?;
    let content_length = response_json.len();

    // Write Content-Length header
    stdout
        .write_all(format!("Content-Length: {}\r\n\r\n", content_length).as_bytes())
        .await?;

    // Write JSON body
    stdout.write_all(response_json.as_bytes()).await?;
    stdout.flush().await?;
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    // Setup logging with unified helper
    distiller_core::logging::init_logging_from_env("info");

    log::info!("🚀 MCP Server v{} starting...", env!("CARGO_PKG_VERSION"));

    let server = McpServer::new();
    log::info!("✅ Server initialized with 13 language processors");
    log::info!("📡 Listening for JSON-RPC requests on stdin...");

    // Read JSON-RPC requests from stdin with Content-Length framing
    let stdin = tokio::io::stdin();
    let mut reader = BufReader::new(stdin);
    let mut stdout = tokio::io::stdout();

    loop {
        // Read Content-Length header
        let mut header_line = String::new();
        match reader.read_line(&mut header_line).await {
            Ok(0) => {
                // EOF
                log::info!("📪 Received EOF, shutting down...");
                break;
            }
            Ok(_) => {
                let header_line = header_line.trim();

                // Skip empty lines
                if header_line.is_empty() {
                    continue;
                }

                // Parse Content-Length header
                let content_length =
                    if let Some(len_str) = header_line.strip_prefix("Content-Length:") {
                        match len_str.trim().parse::<usize>() {
                            Ok(len) => len,
                            Err(e) => {
                                log::error!("❌ Failed to parse Content-Length: {e}");
                                continue;
                            }
                        }
                    } else {
                        log::error!("❌ Expected Content-Length header, got: {}", header_line);
                        continue;
                    };

                // Validate body size to prevent memory abuse
                if content_length > MAX_BODY_BYTES {
                    log::warn!(
                        "⚠️  Request body too large: {} bytes (max: {} bytes)",
                        content_length,
                        MAX_BODY_BYTES
                    );
                    let error_response = JsonRpcResponse {
                        jsonrpc: "2.0".to_string(),
                        id: serde_json::Value::Null,
                        result: None,
                        error: Some(JsonRpcError::invalid_params(format!(
                            "Request body too large: {} bytes (max: {} bytes)",
                            content_length, MAX_BODY_BYTES
                        ))),
                    };
                    if let Err(e) = send_response(&mut stdout, &error_response).await {
                        log::error!("❌ Failed to send error response: {e}");
                    }
                    // Skip reading the oversized body
                    continue;
                }

                // Read blank line after header
                let mut blank_line = String::new();
                if let Err(e) = reader.read_line(&mut blank_line).await {
                    log::error!("❌ Failed to read blank line: {e}");
                    break;
                }

                // Read exactly content_length bytes for the JSON body
                let mut body_buf = vec![0u8; content_length];
                if let Err(e) = reader.read_exact(&mut body_buf).await {
                    log::error!("❌ Failed to read message body: {e}");
                    break;
                }

                let body = match String::from_utf8(body_buf) {
                    Ok(s) => s,
                    Err(e) => {
                        log::error!("❌ Invalid UTF-8 in message body: {e}");
                        continue;
                    }
                };

                // Parse JSON-RPC request
                let request: JsonRpcRequest = match serde_json::from_str(&body) {
                    Ok(req) => req,
                    Err(e) => {
                        log::error!("❌ Failed to parse JSON-RPC request: {e}");
                        continue;
                    }
                };

                log::info!(
                    "📥 Received request: method={}, id={:?}",
                    request.method,
                    request.id
                );

                // Handle request based on method
                let response = match request.method.as_str() {
                    "distil_directory" => {
                        // Parse params
                        let params: DistilDirectoryParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null),
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
                                        message: format!("Invalid params: {e}"),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_distil_directory(params.clone()).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::Value::String(result)),
                                error: None,
                            },
                            Err(e) => {
                                let error = match e.downcast_ref::<DistilError>() {
                                    Some(DistilError::FileNotFound(_)) => {
                                        JsonRpcError::file_not_found(
                                            params.path.display().to_string(),
                                        )
                                    }
                                    _ => JsonRpcError::processing_failed(
                                        e.to_string(),
                                        Some(params.path.display().to_string()),
                                    ),
                                };
                                JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id,
                                    result: None,
                                    error: Some(error),
                                }
                            }
                        }
                    }
                    "distil_file" => {
                        // Parse params
                        let params: DistilFileParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null),
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
                                        message: format!("Invalid params: {e}"),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_distil_file(params.clone()).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::Value::String(result)),
                                error: None,
                            },
                            Err(e) => {
                                let error = match e.downcast_ref::<DistilError>() {
                                    Some(DistilError::FileNotFound(_)) => {
                                        JsonRpcError::file_not_found(
                                            params.path.display().to_string(),
                                        )
                                    }
                                    _ => JsonRpcError::processing_failed(
                                        e.to_string(),
                                        Some(params.path.display().to_string()),
                                    ),
                                };
                                JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id,
                                    result: None,
                                    error: Some(error),
                                }
                            }
                        }
                    }
                    "list_dir" => {
                        // Parse params
                        let params: ListDirParams = match serde_json::from_value(
                            request.params.unwrap_or(serde_json::Value::Null),
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
                                        message: format!("Invalid params: {e}"),
                                        data: None,
                                    }),
                                };
                                send_response(&mut stdout, &error_response).await?;
                                log::info!("📤 Sent error response for id={:?}", error_response.id);
                                continue;
                            }
                        };

                        // Handle operation
                        match server.handle_list_dir(params.clone()).await {
                            Ok(result) => JsonRpcResponse {
                                jsonrpc: "2.0".to_string(),
                                id: request.id,
                                result: Some(serde_json::to_value(result).unwrap()),
                                error: None,
                            },
                            Err(e) => {
                                let error = match e.downcast_ref::<DistilError>() {
                                    Some(DistilError::FileNotFound(_)) => {
                                        JsonRpcError::file_not_found(
                                            params.path.display().to_string(),
                                        )
                                    }
                                    _ => JsonRpcError::processing_failed(
                                        e.to_string(),
                                        Some(params.path.display().to_string()),
                                    ),
                                };
                                JsonRpcResponse {
                                    jsonrpc: "2.0".to_string(),
                                    id: request.id,
                                    result: None,
                                    error: Some(error),
                                }
                            }
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
                                error: Some(JsonRpcError::processing_failed(e.to_string(), None)),
                            },
                        }
                    }
                    _ => JsonRpcResponse {
                        jsonrpc: "2.0".to_string(),
                        id: request.id,
                        result: None,
                        error: Some(JsonRpcError::method_not_found(request.method.clone())),
                    },
                };

                // Send response
                send_response(&mut stdout, &response).await?;
                log::info!("📤 Sent response for id={:?}", response.id);
            }
            Err(e) => {
                log::error!("❌ Failed to read from stdin: {e}");
                break;
            }
        }
    }

    log::info!("👋 MCP Server shutting down...");
    Ok(())
}
