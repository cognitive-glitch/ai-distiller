//! Markdown formatter for AI Distiller
//!
//! Human-readable Markdown format with syntax highlighting.
//! Wraps the text formatter output in Markdown code blocks with proper language identifiers.

#[allow(clippy::wildcard_imports)]
use distiller_core::ir::*;
use formatter_text::{TextFormatter, TextFormatterOptions};

/// Markdown formatter
pub struct MarkdownFormatter {
    text_formatter: TextFormatter,
}

impl MarkdownFormatter {
    /// Create a new markdown formatter with default options
    #[must_use]
    pub fn new() -> Self {
        Self {
            text_formatter: TextFormatter::new(),
        }
    }

    /// Create a new markdown formatter with custom text formatter options
    #[must_use]
    pub fn with_options(options: TextFormatterOptions) -> Self {
        Self {
            text_formatter: TextFormatter::with_options(options),
        }
    }

    /// Format a single file as Markdown
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_file(&self, file: &File) -> Result<String, std::fmt::Error> {
        // First, format as text
        let text = self.text_formatter.format_file(file)?;

        // Extract content from <file> tags
        let content = self.extract_file_content(&text, &file.path);

        // Wrap in Markdown
        let mut output = String::new();
        output.push_str(&format!("### {}\n\n", file.path));

        // Determine language for syntax highlighting
        let lang = get_language_from_path(&file.path);

        output.push_str(&format!("```{lang}\n"));
        output.push_str(&content);
        if !content.ends_with('\n') {
            output.push('\n');
        }
        output.push_str("```\n");

        Ok(output)
    }

    /// Format multiple files as Markdown
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_files(&self, files: &[File]) -> Result<String, std::fmt::Error> {
        let mut output = String::new();

        for (i, file) in files.iter().enumerate() {
            if i > 0 {
                output.push_str("\n\n");
            }
            output.push_str(&self.format_file(file)?);
        }

        Ok(output)
    }

    /// Extract content from between <file path="..."> and </file> tags
    #[allow(clippy::unused_self)]
    fn extract_file_content(&self, text: &str, path: &str) -> String {
        let start_tag = format!("<file path=\"{path}\">");
        let end_tag = "</file>";

        if let Some(start_idx) = text.find(&start_tag) {
            let content_start = start_idx + start_tag.len();

            // Skip newline after opening tag
            let content_start = if text.as_bytes().get(content_start) == Some(&b'\n') {
                content_start + 1
            } else {
                content_start
            };

            if let Some(end_idx) = text[content_start..].find(end_tag) {
                return text[content_start..content_start + end_idx].to_string();
            }
        }

        // Fallback: return the text as-is
        text.to_string()
    }
}

impl Default for MarkdownFormatter {
    fn default() -> Self {
        Self::new()
    }
}

/// Get language identifier for syntax highlighting based on file extension
fn get_language_from_path(path: &str) -> &str {
    if let Some(ext_start) = path.rfind('.') {
        let ext = &path[ext_start + 1..];
        match ext {
            "py" => "python",
            "go" => "go",
            "ts" | "tsx" => "typescript",
            "js" | "jsx" => "javascript",
            "java" => "java",
            "cs" => "csharp",
            "cpp" | "cc" | "cxx" | "hpp" | "hxx" | "h" => "cpp",
            "rb" => "ruby",
            "rs" => "rust",
            "swift" => "swift",
            "kt" | "kts" => "kotlin",
            "php" => "php",
            "c" => "c",
            _ => ext,
        }
    } else {
        ""
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_markdown_format_simple() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Class(Class {
                name: "Example".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                extends: Vec::new(),
                implements: Vec::new(),
                children: vec![Node::Function(Function {
                    name: "__init__".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: vec![Parameter {
                        name: "self".to_string(),
                        param_type: TypeRef::new("Self"),
                        default_value: None,
                        is_variadic: false,
                        is_optional: false,
                        decorators: Vec::new(),
                    }],
                    return_type: None,
                    implementation: None,
                    line_start: 2,
                    line_end: 3,
                })],
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = MarkdownFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("### test.py"));
        assert!(result.contains("```python"));
        assert!(result.contains("class Example:"));
        assert!(result.contains("def __init__(self: Self)"));
        assert!(result.contains("```"));
    }

    #[test]
    fn test_language_detection() {
        assert_eq!(get_language_from_path("file.py"), "python");
        assert_eq!(get_language_from_path("file.go"), "go");
        assert_eq!(get_language_from_path("file.ts"), "typescript");
        assert_eq!(get_language_from_path("file.tsx"), "typescript");
        assert_eq!(get_language_from_path("file.js"), "javascript");
        assert_eq!(get_language_from_path("file.jsx"), "javascript");
        assert_eq!(get_language_from_path("file.rs"), "rust");
        assert_eq!(get_language_from_path("file.java"), "java");
        assert_eq!(get_language_from_path("file.cpp"), "cpp");
        assert_eq!(get_language_from_path("file.rb"), "ruby");
    }

    #[test]
    fn test_multiple_files() {
        let files = vec![
            File {
                path: "file1.py".to_string(),
                children: vec![Node::Function(Function {
                    name: "hello".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: Vec::new(),
                    return_type: None,
                    implementation: None,
                    line_start: 1,
                    line_end: 2,
                })],
            },
            File {
                path: "file2.py".to_string(),
                children: vec![Node::Function(Function {
                    name: "world".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: Vec::new(),
                    return_type: None,
                    implementation: None,
                    line_start: 1,
                    line_end: 2,
                })],
            },
        ];

        let formatter = MarkdownFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        assert!(result.contains("### file1.py"));
        assert!(result.contains("### file2.py"));
        assert!(result.contains("def hello()"));
        assert!(result.contains("def world()"));
    }

    #[test]
    fn test_private_field_in_markdown() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Class(Class {
                name: "Example".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                extends: Vec::new(),
                implements: Vec::new(),
                children: vec![Node::Field(Field {
                    name: "_private".to_string(),
                    visibility: Visibility::Private,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("str")),
                    default_value: None,
                    line: 2,
                })],
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = MarkdownFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should contain the private field with - prefix
        assert!(result.contains("-_private: str"));
    }
}
