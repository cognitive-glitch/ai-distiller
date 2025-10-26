//! Error types for AI Distiller
//!
//! Uses `thiserror` for ergonomic error handling with proper context.

use std::path::PathBuf;

/// Result type alias for distiller operations
pub type Result<T> = std::result::Result<T, DistilError>;

/// Core error type for AI Distiller operations
#[derive(thiserror::Error, Debug)]
pub enum DistilError {
    /// I/O errors (file reading, writing, etc.)
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    /// Unsupported language for a given file
    #[error("Unsupported language for {path}: {lang}")]
    UnsupportedLanguage { path: String, lang: String },

    /// Parse error during tree-sitter processing
    #[error("Parse error in {path}: {message}")]
    Parse { path: String, message: String },

    /// Tree-sitter specific errors
    #[error("Tree-sitter error: {0}")]
    TreeSitter(String),

    /// Invalid configuration or options
    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    /// File not found
    #[error("File not found: {0}")]
    FileNotFound(PathBuf),

    /// Directory traversal error
    #[error("Directory traversal error: {0}")]
    WalkDir(String),

    /// Serialization/deserialization errors
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
}

impl DistilError {
    /// Create a parse error with context
    pub fn parse_error(path: impl Into<String>, message: impl Into<String>) -> Self {
        Self::Parse {
            path: path.into(),
            message: message.into(),
        }
    }

    /// Create an unsupported language error
    pub fn unsupported_language(path: impl Into<String>, lang: impl Into<String>) -> Self {
        Self::UnsupportedLanguage {
            path: path.into(),
            lang: lang.into(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_error_display() {
        let err = DistilError::parse_error("test.py", "Invalid syntax");
        assert_eq!(
            err.to_string(),
            "Parse error in test.py: Invalid syntax"
        );
    }

    #[test]
    fn test_unsupported_language() {
        let err = DistilError::unsupported_language("test.xyz", "xyz");
        assert_eq!(
            err.to_string(),
            "Unsupported language for test.xyz: xyz"
        );
    }
}
