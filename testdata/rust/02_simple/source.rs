// 02_simple.rs
// A test case for simple structs, impls, and traits.

/// Represents a source file to be processed.
// The parser should handle struct definitions with public and private fields.
pub struct SourceFile {
    pub path: String,
    content: String, // This field is private.
    lines_of_code: u32,
}

/// A trait for items that can be summarized.
// This tests the ability to parse trait definitions and their methods.
pub trait Summarizable {
    fn summary(&self) -> String;
    
    /// Default implementation for short summary
    fn short_summary(&self) -> String {
        let full = self.summary();
        if full.len() > 50 {
            format!("{}...", &full[..47])
        } else {
            full
        }
    }
}

// Implementation of methods for the SourceFile struct.
impl SourceFile {
    /// Creates a new SourceFile, demonstrating ownership (takes ownership of path and content).
    pub fn new(path: String, content: String) -> Self {
        let lines_of_code = content.lines().count() as u32;
        Self {
            path,
            content,
            lines_of_code,
        }
    }

    /// A public method to access a derived property.
    pub fn line_count(&self) -> u32 {
        self.lines_of_code
    }

    // A private helper method.
    fn get_file_extension(&self) -> Option<&str> {
        std::path::Path::new(&self.path)
            .extension()
            .and_then(std::ffi::OsStr::to_str)
    }

    /// Internal method for processing
    pub(crate) fn process_internal(&mut self) {
        self.lines_of_code = self.content.lines().count() as u32;
    }

    /// Private validation method
    fn is_valid(&self) -> bool {
        !self.path.is_empty() && !self.content.is_empty()
    }
}

// Implementation of the Summarizable trait for SourceFile.
// This is a critical pattern: `impl Trait for Struct`.
impl Summarizable for SourceFile {
    fn summary(&self) -> String {
        let extension = self.get_file_extension().unwrap_or("unknown");
        format!(
            "File '{}' ({} lines, type: {})",
            self.path, self.lines_of_code, extension
        )
    }
}

/// Additional trait for file operations
pub trait FileOperations {
    type Error;
    
    fn read_content(&self) -> Result<&str, Self::Error>;
    fn write_content(&mut self, content: String) -> Result<(), Self::Error>;
}

/// Error type for file operations
#[derive(Debug)]
pub enum FileError {
    NotFound,
    PermissionDenied,
    InvalidContent,
}

impl FileOperations for SourceFile {
    type Error = FileError;
    
    fn read_content(&self) -> Result<&str, Self::Error> {
        if self.is_valid() {
            Ok(&self.content)
        } else {
            Err(FileError::InvalidContent)
        }
    }
    
    fn write_content(&mut self, content: String) -> Result<(), Self::Error> {
        self.content = content;
        self.process_internal();
        Ok(())
    }
}

fn main() {
    let file = SourceFile::new(
        "src/main.rs".to_string(),
        "fn main() {\n  println!(\"Hello\");\n}".to_string(),
    );
    println!("{}", file.summary());
}