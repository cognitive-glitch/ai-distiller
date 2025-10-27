//! File and directory processing
//!
//! The processor is responsible for:
//! - Walking directories with .gitignore support
//! - Detecting file languages
//! - Dispatching to language processors
//! - Parallel processing with rayon

pub mod directory;
pub mod language;

pub use directory::{DirectoryProcessor, LanguageRegistry};
pub use language::LanguageProcessor;

use crate::{ProcessOptions, Result, ir::Node};
use std::path::Path;

/// Main processor for files and directories
pub struct Processor {
    options: ProcessOptions,
    language_registry: LanguageRegistry,
}

impl Processor {
    /// Create a new processor with options
    #[must_use]
    pub fn new(options: ProcessOptions) -> Self {
        Self {
            options,
            language_registry: LanguageRegistry::new(),
        }
    }

    /// Create processor with default options
    #[must_use]
    pub fn with_defaults() -> Self {
        Self::new(ProcessOptions::default())
    }

    /// Register a language processor
    pub fn register_language(&mut self, processor: Box<dyn language::LanguageProcessor>) {
        self.language_registry.register(processor);
    }

    /// Process a file or directory
    ///
    /// Automatically detects whether the path is a file or directory
    /// and dispatches to the appropriate processor.
    ///
    /// # Errors
    ///
    /// Returns an error if the path does not exist, is not accessible, or processing fails.
    pub fn process_path(&self, path: &Path) -> Result<Node> {
        if path.is_dir() {
            // Process directory
            let dir_processor = DirectoryProcessor::new(self.options.clone());
            let directory = dir_processor.process(path, &self.language_registry)?;
            Ok(Node::Directory(directory))
        } else if path.is_file() {
            // Process single file
            let file = self.process_single_file(path)?;
            Ok(Node::File(file))
        } else {
            Err(crate::error::DistilError::InvalidConfig(format!(
                "Path does not exist or is not accessible: {}",
                path.display()
            )))
        }
    }

    /// Process a single file
    fn process_single_file(&self, path: &Path) -> Result<crate::ir::File> {
        use crate::error::DistilError;

        // Find processor for this file
        let processor = self.language_registry.find_processor(path).ok_or_else(|| {
            DistilError::UnsupportedLanguage {
                path: path.display().to_string(),
                lang: path
                    .extension()
                    .and_then(|s| s.to_str())
                    .unwrap_or("unknown")
                    .to_string(),
            }
        })?;

        // Read file
        let source = std::fs::read_to_string(path).map_err(DistilError::Io)?;

        // Process with language-specific processor
        processor.process(&source, path, &self.options)
    }

    /// Get reference to language registry (for testing/inspection)
    #[must_use]
    pub fn language_registry(&self) -> &LanguageRegistry {
        &self.language_registry
    }

    /// Get reference to options (for testing/inspection)
    #[must_use]
    pub fn options(&self) -> &ProcessOptions {
        &self.options
    }
}

impl Default for Processor {
    fn default() -> Self {
        Self::with_defaults()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let opts = ProcessOptions::default();
        let processor = Processor::new(opts.clone());

        assert_eq!(processor.options().workers, opts.workers);
    }

    #[test]
    fn test_processor_with_defaults() {
        let processor = Processor::with_defaults();

        assert_eq!(processor.options().workers, 0); // Auto
        assert!(processor.options().recursive);
    }

    #[test]
    fn test_invalid_path_error() {
        let processor = Processor::with_defaults();
        let result = processor.process_path(Path::new("/tmp/nonexistent_xyz_12345"));

        assert!(result.is_err());
    }
}
