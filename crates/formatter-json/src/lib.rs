//! JSON formatter for AI Distiller
//!
//! Structured JSON format for tools and programmatic processing.
//! Provides both pretty-printed and compact JSON output.

use distiller_core::ir::*;
use serde_json;

/// JSON formatter options
#[derive(Debug, Clone)]
pub struct JsonFormatterOptions {
    /// Pretty-print the JSON output
    pub pretty: bool,
}

impl Default for JsonFormatterOptions {
    fn default() -> Self {
        Self { pretty: true }
    }
}

/// JSON formatter
pub struct JsonFormatter {
    options: JsonFormatterOptions,
}

impl JsonFormatter {
    /// Create a new JSON formatter with default options (pretty-printed)
    pub fn new() -> Self {
        Self {
            options: JsonFormatterOptions::default(),
        }
    }

    /// Create a new JSON formatter with custom options
    pub fn with_options(options: JsonFormatterOptions) -> Self {
        Self { options }
    }

    /// Format a single file as JSON
    pub fn format_file(&self, file: &File) -> Result<String, serde_json::Error> {
        if self.options.pretty {
            serde_json::to_string_pretty(file)
        } else {
            serde_json::to_string(file)
        }
    }

    /// Format multiple files as JSON array
    pub fn format_files(&self, files: &[File]) -> Result<String, serde_json::Error> {
        if self.options.pretty {
            serde_json::to_string_pretty(files)
        } else {
            serde_json::to_string(files)
        }
    }
}

impl Default for JsonFormatter {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_json_format_simple() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![
                Node::Class(Class {
                    name: "Example".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    extends: Vec::new(),
                    implements: Vec::new(),
                    children: vec![
                        Node::Function(Function {
                            name: "__init__".to_string(),
                            visibility: Visibility::Public,
                            modifiers: Vec::new(),
                            decorators: Vec::new(),
                            type_params: Vec::new(),
                            parameters: vec![
                                Parameter {
                                    name: "self".to_string(),
                                    param_type: TypeRef::new("Self"),
                                    default_value: None,
                                    is_variadic: false,
                                    is_optional: false,
                                    decorators: Vec::new(),
                                },
                            ],
                            return_type: None,
                            implementation: None,
                            line_start: 2,
                            line_end: 3,
                        }),
                    ],
                    line_start: 1,
                    line_end: 3,
                }),
            ],
        };

        let formatter = JsonFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should be valid JSON
        assert!(result.contains("\"path\": \"test.py\""));
        assert!(result.contains("\"name\": \"Example\""));
        assert!(result.contains("\"name\": \"__init__\""));
        assert!(result.contains("\"kind\": \"class\""));

        // Should be pretty-printed (contains newlines)
        assert!(result.contains('\n'));
    }

    #[test]
    fn test_json_compact() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![
                Node::Function(Function {
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
                }),
            ],
        };

        let options = JsonFormatterOptions { pretty: false };
        let formatter = JsonFormatter::with_options(options);
        let result = formatter.format_file(&file).unwrap();

        // Should be valid JSON
        assert!(result.contains("\"path\":\"test.py\""));
        assert!(result.contains("\"name\":\"hello\""));

        // Should be compact (minimal whitespace)
        assert!(!result.contains("  ")); // No double spaces
    }

    #[test]
    fn test_json_multiple_files() {
        let files = vec![
            File {
                path: "file1.py".to_string(),
                children: vec![
                    Node::Function(Function {
                        name: "func1".to_string(),
                        visibility: Visibility::Public,
                        modifiers: Vec::new(),
                        decorators: Vec::new(),
                        type_params: Vec::new(),
                        parameters: Vec::new(),
                        return_type: None,
                        implementation: None,
                        line_start: 1,
                        line_end: 2,
                    }),
                ],
            },
            File {
                path: "file2.py".to_string(),
                children: vec![
                    Node::Function(Function {
                        name: "func2".to_string(),
                        visibility: Visibility::Public,
                        modifiers: Vec::new(),
                        decorators: Vec::new(),
                        type_params: Vec::new(),
                        parameters: Vec::new(),
                        return_type: None,
                        implementation: None,
                        line_start: 1,
                        line_end: 2,
                    }),
                ],
            },
        ];

        let formatter = JsonFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        // Should be JSON array
        assert!(result.starts_with('['));
        assert!(result.ends_with("]\n") || result.ends_with(']'));

        // Should contain both files
        assert!(result.contains("\"path\": \"file1.py\""));
        assert!(result.contains("\"path\": \"file2.py\""));
        assert!(result.contains("\"name\": \"func1\""));
        assert!(result.contains("\"name\": \"func2\""));
    }

    #[test]
    fn test_json_visibility() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![
                Node::Field(Field {
                    name: "_private".to_string(),
                    visibility: Visibility::Private,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("str")),
                    default_value: None,
                    line: 1,
                }),
                Node::Field(Field {
                    name: "public".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("int")),
                    default_value: None,
                    line: 2,
                }),
            ],
        };

        let formatter = JsonFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should contain visibility fields
        assert!(result.contains("\"visibility\": \"private\""));
        assert!(result.contains("\"visibility\": \"public\""));
        assert!(result.contains("\"name\": \"_private\""));
        assert!(result.contains("\"name\": \"public\""));
    }

    #[test]
    fn test_json_type_params() {
        let file = File {
            path: "test.ts".to_string(),
            children: vec![
                Node::Class(Class {
                    name: "Container".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: vec![
                        TypeParam {
                            name: "T".to_string(),
                            constraints: Vec::new(),
                            default: None,
                        },
                    ],
                    extends: Vec::new(),
                    implements: Vec::new(),
                    children: Vec::new(),
                    line_start: 1,
                    line_end: 3,
                }),
            ],
        };

        let formatter = JsonFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should contain type parameters
        assert!(result.contains("\"type_params\""));
        assert!(result.contains("\"name\": \"T\""));
    }
}
