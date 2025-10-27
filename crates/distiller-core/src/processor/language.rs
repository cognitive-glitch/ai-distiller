//! Language processor trait
//!
//! All language-specific processors implement this trait.
//! Each language (Python, TypeScript, etc.) has its own crate.

use crate::{ProcessOptions, Result, ir::File};
use std::path::Path;

/// Trait for language-specific processors
///
/// Each language processor:
/// - Identifies files it can process (by extension)
/// - Parses source code using tree-sitter
/// - Converts tree-sitter AST to our IR
///
/// **IMPORTANT**: This trait is SYNCHRONOUS (no async/await).
/// Use rayon for parallelism at the processor level.
pub trait LanguageProcessor: Send + Sync {
    /// Get the language name
    fn language(&self) -> &'static str;

    /// Get supported file extensions
    fn supported_extensions(&self) -> &'static [&'static str];

    /// Check if this processor can handle a file
    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| self.supported_extensions().contains(&ext))
            .unwrap_or(false)
    }

    /// Process a source file
    ///
    /// This method is SYNCHRONOUS for simplicity and performance.
    /// Parsing is CPU-bound, so async provides no benefit.
    fn process(&self, source: &str, path: &Path, opts: &ProcessOptions) -> Result<File>;
}
