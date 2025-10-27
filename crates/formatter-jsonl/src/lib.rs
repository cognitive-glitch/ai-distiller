//! JSONL (JSON Lines) formatter for AI Distiller
//!
//! Outputs one JSON object per line (newline-delimited JSON).
//! Optimized for streaming processing and log aggregation.
//! Always uses compact format (no pretty-printing).

use distiller_core::ir::*;

/// JSONL formatter (always compact, one JSON per line)
pub struct JsonlFormatter;

impl JsonlFormatter {
    /// Create a new JSONL formatter
    pub fn new() -> Self {
        Self
    }

    /// Format a single file as compact JSON
    pub fn format_file(&self, file: &File) -> Result<String, serde_json::Error> {
        serde_json::to_string(file)
    }

    /// Format multiple files as JSONL (one JSON object per line)
    pub fn format_files(&self, files: &[File]) -> Result<String, serde_json::Error> {
        let mut output = String::new();

        for (i, file) in files.iter().enumerate() {
            let json = serde_json::to_string(file)?;
            output.push_str(&json);

            // Add newline after each JSON object (except potentially last)
            if i < files.len() - 1 {
                output.push('\n');
            }
        }

        // Always end with newline for JSONL format
        if !output.is_empty() && !output.ends_with('\n') {
            output.push('\n');
        }

        Ok(output)
    }
}

impl Default for JsonlFormatter {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_jsonl_format_simple() {
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

        let formatter = JsonlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should be valid compact JSON
        assert!(result.contains("\"path\":\"test.py\""));
        assert!(result.contains("\"name\":\"Example\""));
        assert!(result.contains("\"name\":\"__init__\""));
        assert!(result.contains("\"kind\":\"class\""));

        // Should be compact (no newlines, minimal whitespace)
        assert!(!result.contains('\n'));
        assert!(!result.contains("  ")); // No double spaces
    }

    #[test]
    fn test_jsonl_multiple_files() {
        let files = vec![
            File {
                path: "file1.py".to_string(),
                children: vec![Node::Function(Function {
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
                })],
            },
            File {
                path: "file2.py".to_string(),
                children: vec![Node::Function(Function {
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
                })],
            },
        ];

        let formatter = JsonlFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        // Should contain both files
        assert!(result.contains("\"path\":\"file1.py\""));
        assert!(result.contains("\"path\":\"file2.py\""));
        assert!(result.contains("\"name\":\"func1\""));
        assert!(result.contains("\"name\":\"func2\""));

        // Should have exactly 2 lines (one per file)
        let lines: Vec<&str> = result.lines().collect();
        assert_eq!(lines.len(), 2);

        // Each line should be valid JSON
        for line in lines {
            serde_json::from_str::<File>(line).expect("Each line should be valid JSON");
        }

        // Should end with newline
        assert!(result.ends_with('\n'));
    }

    #[test]
    fn test_jsonl_single_line_per_file() {
        let files = vec![
            File {
                path: "a.py".to_string(),
                children: vec![],
            },
            File {
                path: "b.py".to_string(),
                children: vec![],
            },
            File {
                path: "c.py".to_string(),
                children: vec![],
            },
        ];

        let formatter = JsonlFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        // Should have exactly 3 lines
        let lines: Vec<&str> = result.lines().collect();
        assert_eq!(lines.len(), 3);

        // No empty lines
        assert!(!result.contains("\n\n"));

        // Should end with single newline
        assert!(result.ends_with('\n'));
        assert!(!result.ends_with("\n\n"));
    }

    #[test]
    fn test_jsonl_visibility() {
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

        let formatter = JsonlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should contain visibility fields
        assert!(result.contains("\"visibility\":\"private\""));
        assert!(result.contains("\"visibility\":\"public\""));
        assert!(result.contains("\"name\":\"_private\""));
        assert!(result.contains("\"name\":\"public\""));
    }

    #[test]
    fn test_jsonl_type_params() {
        let file = File {
            path: "test.ts".to_string(),
            children: vec![Node::Class(Class {
                name: "Container".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: vec![TypeParam {
                    name: "T".to_string(),
                    constraints: Vec::new(),
                    default: None,
                }],
                extends: Vec::new(),
                implements: Vec::new(),
                children: Vec::new(),
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = JsonlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        // Should contain type parameters
        assert!(result.contains("\"type_params\""));
        assert!(result.contains("\"name\":\"T\""));
    }

    #[test]
    fn test_jsonl_streaming_parse() {
        // Test that output can be parsed line-by-line (streaming)
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

        let formatter = JsonlFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        // Parse each line independently (simulating streaming)
        let mut parsed_count = 0;
        for line in result.lines() {
            let file: File =
                serde_json::from_str(line).expect("Each line should be independently parseable");

            // Verify structure
            assert!(!file.path.is_empty());
            parsed_count += 1;
        }

        // Should have parsed exactly 2 files
        assert_eq!(parsed_count, 2);
    }
}
