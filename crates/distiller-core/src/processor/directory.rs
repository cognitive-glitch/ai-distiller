//! Directory processing with rayon parallelism
//!
//! Processes entire directory trees in parallel while maintaining file order.
//! Respects .gitignore patterns and provides progress tracking.

use crate::{
    ProcessOptions,
    error::{DistilError, Result},
    ir::{Directory, File, Node},
};
use ignore::WalkBuilder;
use rayon::prelude::*;
use std::path::{Path, PathBuf};
use std::sync::Arc;

/// Result of processing a single file
struct FileResult {
    result: Result<File>,
    index: usize, // Original order
}

/// Directory processor that uses rayon for parallelism
pub struct DirectoryProcessor {
    /// Process options (visibility, content, parallelism)
    options: Arc<ProcessOptions>,
}

impl DirectoryProcessor {
    /// Create a new directory processor
    #[must_use]
    pub fn new(options: ProcessOptions) -> Self {
        Self {
            options: Arc::new(options),
        }
    }

    /// Process a directory tree
    ///
    /// Returns a Directory node containing all processed files.
    /// Files are processed in parallel using rayon, but results maintain
    /// their original discovery order.
    ///
    /// # Arguments
    /// * `path` - Root directory to process
    /// * `language_registry` - Registry of language processors
    ///
    /// # Errors
    /// * If directory doesn't exist or isn't readable
    /// * If any file fails to parse (can be relaxed in future)
    pub fn process<P: AsRef<Path>>(
        &self,
        path: P,
        language_registry: &LanguageRegistry,
    ) -> Result<Directory> {
        let path = path.as_ref();

        if !path.is_dir() {
            return Err(DistilError::InvalidConfig(format!(
                "Path is not a directory: {}",
                path.display()
            )));
        }

        // Collect files to process
        let files = self.discover_files(path)?;

        // Process files in parallel
        let results = self.process_files(&files, language_registry)?;

        // Build directory structure
        Ok(Directory {
            path: path.to_string_lossy().into_owned(),
            children: results.into_iter().map(Node::File).collect(),
        })
    }

    /// Discover files in directory respecting .gitignore
    fn discover_files(&self, root: &Path) -> Result<Vec<(PathBuf, usize)>> {
        let mut builder = WalkBuilder::new(root);

        // Configure walker
        builder
            .standard_filters(true) // Respect .gitignore, .ignore, .git/info/exclude
            .hidden(false) // Include hidden files (for now)
            .follow_links(false) // Don't follow symlinks (avoid cycles)
            .max_depth(if self.options.recursive {
                None
            } else {
                Some(1)
            });

        // Build walker and collect files
        let walker = builder.build();

        let mut files = Vec::new();
        let mut index = 0;

        for entry in walker {
            let entry = entry.map_err(|e| DistilError::Io(std::io::Error::other(e.to_string())))?;

            let path = entry.path();

            // Only process regular files
            if path.is_file() {
                files.push((path.to_path_buf(), index));
                index += 1;
            }
        }

        Ok(files)
    }

    /// Process files in parallel using rayon
    fn process_files(
        &self,
        files: &[(PathBuf, usize)],
        language_registry: &LanguageRegistry,
    ) -> Result<Vec<File>> {
        let opts = self.options.clone();

        // Process in parallel
        let mut results: Vec<FileResult> = files
            .par_iter()
            .map(|(path, index)| {
                let result = Self::process_single_file(path, language_registry, &opts);

                FileResult {
                    result,
                    index: *index,
                }
            })
            .collect();

        // Sort by original discovery order
        results.sort_by_key(|r| r.index);

        // Extract results, propagating errors
        results
            .into_iter()
            .map(|r| match r.result {
                Ok(file) => Ok(file),
                Err(e) => Err(e),
            })
            .collect()
    }

    /// Process a single file
    fn process_single_file(
        path: &Path,
        language_registry: &LanguageRegistry,
        opts: &ProcessOptions,
    ) -> Result<File> {
        // Find processor for this file
        let processor = language_registry.find_processor(path).ok_or_else(|| {
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
        processor.process(&source, path, opts)
    }
}

/// Registry of language processors
///
/// Stores language processors and finds the appropriate one for a given file.
pub struct LanguageRegistry {
    processors: Vec<Box<dyn super::language::LanguageProcessor>>,
}

impl LanguageRegistry {
    /// Create an empty registry
    #[must_use]
    pub fn new() -> Self {
        Self {
            processors: Vec::new(),
        }
    }

    /// Register a language processor
    pub fn register(&mut self, processor: Box<dyn super::language::LanguageProcessor>) {
        self.processors.push(processor);
    }

    /// Find a processor that can handle this file
    pub(crate) fn find_processor(
        &self,
        path: &Path,
    ) -> Option<&dyn super::language::LanguageProcessor> {
        for processor in &self.processors {
            if processor.can_process(path) {
                return Some(processor.as_ref());
            }
        }
        None
    }
}

impl Default for LanguageRegistry {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let opts = ProcessOptions::default();
        let processor = DirectoryProcessor::new(opts);

        assert!(Arc::strong_count(&processor.options) >= 1);
    }

    #[test]
    fn test_registry_creation() {
        let registry = LanguageRegistry::new();
        let result = registry.find_processor(Path::new("test.py"));

        assert!(result.is_none()); // No processors registered yet
    }

    #[test]
    fn test_non_directory_error() {
        let opts = ProcessOptions::default();
        let processor = DirectoryProcessor::new(opts);
        let registry = LanguageRegistry::new();

        // Try to process a non-existent path
        let result = processor.process("/tmp/nonexistent_file_12345.txt", &registry);

        assert!(result.is_err());
    }

    // Integration tests will be added when we have actual language processors
}
