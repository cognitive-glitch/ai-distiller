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

use crate::{ir::Node, ProcessOptions, Result};
use std::path::Path;

/// Main processor for files and directories
pub struct Processor {
    options: ProcessOptions,
    language_registry: LanguageRegistry,
}

impl Processor {
    /// Create a new processor with options
    pub fn new(options: ProcessOptions) -> Self {
        Self {
            options,
            language_registry: LanguageRegistry::new(),
        }
    }

    /// Create processor with default options
    pub fn with_defaults() -> Self {
        Self::new(ProcessOptions::default())
    }

    /// Process a file or directory
    ///
    /// Automatically detects whether the path is a file or directory
    /// and dispatches to the appropriate processor.
    pub fn process_path(&self, path: &Path) -> Result<Node> {
        if path.is_dir() {
            // Process directory
            let dir_processor = DirectoryProcessor::new(self.options.clone());
            let directory = dir_processor.process(path, &self.language_registry)?;
            Ok(Node::Directory(directory))
        } else if path.is_file() {
            // Process single file
            // TODO: Implement single file processing
            todo!("Single file processing - Phase 2.4")
        } else {
            Err(crate::error::DistilError::InvalidConfig(format!(
                "Path does not exist or is not accessible: {}",
                path.display()
            )))
        }
    }

    /// Get reference to language registry (for testing/inspection)
    pub fn language_registry(&self) -> &LanguageRegistry {
        &self.language_registry
    }

    /// Get reference to options (for testing/inspection)
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
