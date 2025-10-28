//! Processing options and configuration
//!
//! Defines how files should be processed and what content to include/exclude.

use std::path::PathBuf;

/// Path type for output file paths
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum PathType {
    /// Absolute file paths
    Absolute,
    /// Relative file paths
    Relative,
}

impl Default for PathType {
    fn default() -> Self {
        Self::Relative
    }
}

/// Core processing options
///
/// Controls what content is included in the distilled output and how
/// processing should be performed.
#[derive(Debug, Clone)]
#[allow(clippy::struct_excessive_bools)]
pub struct ProcessOptions {
    // Visibility filtering
    /// Include public members (default: true)
    pub include_public: bool,
    /// Include protected members (default: false)
    pub include_protected: bool,
    /// Include internal/package-private members (default: false)
    pub include_internal: bool,
    /// Include private members (default: false)
    pub include_private: bool,

    // Content filtering
    /// Include regular comments (default: false)
    pub include_comments: bool,
    /// Include documentation comments/docstrings (default: true)
    pub include_docstrings: bool,
    /// Include function/method implementations (default: false)
    pub include_implementation: bool,
    /// Include import statements (default: true)
    pub include_imports: bool,
    /// Include annotations/decorators (default: true)
    pub include_annotations: bool,
    /// Include class fields/properties (default: true)
    pub include_fields: bool,
    /// Include methods/functions (default: true)
    pub include_methods: bool,

    // Processing configuration
    /// Raw mode - process all files as text (default: false)
    pub raw_mode: bool,
    /// Number of worker threads (0 = auto: 80% of CPU cores)
    pub workers: usize,
    /// Process directories recursively (default: true)
    pub recursive: bool,

    // Path configuration
    /// How to format file paths in output
    pub file_path_type: PathType,
    /// Prefix for relative paths
    pub relative_path_prefix: Option<String>,
    /// Base path for relative path calculation
    pub base_path: Option<PathBuf>,

    // Pattern filtering
    /// Include only files matching these patterns (empty = include all)
    pub include_patterns: Vec<String>,
    /// Exclude files matching these patterns
    pub exclude_patterns: Vec<String>,

    // Error handling
    /// Continue processing on file errors (collect partial results)
    pub continue_on_error: bool,
}

impl Default for ProcessOptions {
    fn default() -> Self {
        Self {
            // Default: public APIs only
            include_public: true,
            include_protected: false,
            include_internal: false,
            include_private: false,

            // Default: signatures with docstrings
            include_comments: false,
            include_docstrings: true,
            include_implementation: false,
            include_imports: true,
            include_annotations: true,
            include_fields: true,
            include_methods: true,

            // Default: parallel processing
            raw_mode: false,
            workers: 0, // Auto-detect
            recursive: true,

            // Default: relative paths
            file_path_type: PathType::Relative,
            relative_path_prefix: None,
            base_path: None,

            // Default: no pattern filtering
            include_patterns: Vec::new(),
            exclude_patterns: Vec::new(),

            // Default: fail on first error
            continue_on_error: false,
        }
    }
}

impl ProcessOptions {
    /// Create a new builder for `ProcessOptions`
    #[must_use]
    pub fn builder() -> ProcessOptionsBuilder {
        ProcessOptionsBuilder::default()
    }

    /// Get the number of worker threads to use
    ///
    /// Returns 0 if auto-detection should be used (80% of CPU cores).
    #[must_use]
    pub fn worker_count(&self) -> usize {
        if self.workers == 0 {
            // Auto: 80% of available parallelism
            let cpus = num_cpus::get();
            (cpus * 4 / 5).max(1)
        } else {
            self.workers
        }
    }

    /// Check if any visibility filters are enabled
    #[must_use]
    pub fn has_visibility_filters(&self) -> bool {
        !self.include_public
            || self.include_protected
            || self.include_internal
            || self.include_private
    }

    /// Check if content should be stripped (not in raw mode)
    #[must_use]
    pub fn should_strip_content(&self) -> bool {
        !self.raw_mode
            && (!self.include_comments
                || !self.include_implementation
                || self.has_visibility_filters())
    }
}

/// Builder for `ProcessOptions`
#[derive(Default)]
#[allow(clippy::struct_excessive_bools)]
pub struct ProcessOptionsBuilder {
    options: ProcessOptions,
}

impl ProcessOptionsBuilder {
    #[must_use]
    pub fn include_public(mut self, value: bool) -> Self {
        self.options.include_public = value;
        self
    }

    #[must_use]
    pub fn include_private(mut self, value: bool) -> Self {
        self.options.include_private = value;
        self
    }

    #[must_use]
    pub fn include_protected(mut self, value: bool) -> Self {
        self.options.include_protected = value;
        self
    }

    #[must_use]
    pub fn include_internal(mut self, value: bool) -> Self {
        self.options.include_internal = value;
        self
    }

    #[must_use]
    pub fn include_implementation(mut self, value: bool) -> Self {
        self.options.include_implementation = value;
        self
    }

    #[must_use]
    pub fn include_comments(mut self, value: bool) -> Self {
        self.options.include_comments = value;
        self
    }

    #[must_use]
    pub fn include_annotations(mut self, value: bool) -> Self {
        self.options.include_annotations = value;
        self
    }

    #[must_use]
    pub fn include_fields(mut self, value: bool) -> Self {
        self.options.include_fields = value;
        self
    }

    #[must_use]
    pub fn include_methods(mut self, value: bool) -> Self {
        self.options.include_methods = value;
        self
    }

    #[must_use]
    pub fn file_path_type(mut self, path_type: PathType) -> Self {
        self.options.file_path_type = path_type;
        self
    }

    #[must_use]
    pub fn relative_path_prefix(mut self, prefix: Option<String>) -> Self {
        self.options.relative_path_prefix = prefix;
        self
    }

    #[must_use]
    pub fn base_path(mut self, path: Option<PathBuf>) -> Self {
        self.options.base_path = path;
        self
    }

    #[must_use]
    pub fn workers(mut self, count: usize) -> Self {
        self.options.workers = count;
        self
    }

    #[must_use]
    pub fn recursive(mut self, value: bool) -> Self {
        self.options.recursive = value;
        self
    }

    #[must_use]
    pub fn include_patterns(mut self, patterns: Vec<String>) -> Self {
        self.options.include_patterns = patterns;
        self
    }

    #[must_use]
    pub fn exclude_patterns(mut self, patterns: Vec<String>) -> Self {
        self.options.exclude_patterns = patterns;
        self
    }

    #[must_use]
    pub fn build(self) -> ProcessOptions {
        self.options
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_options() {
        let opts = ProcessOptions::default();
        assert!(opts.include_public);
        assert!(!opts.include_private);
        assert!(!opts.include_implementation);
        assert!(opts.include_docstrings);
    }

    #[test]
    fn test_builder() {
        let opts = ProcessOptions::builder()
            .include_private(true)
            .include_implementation(true)
            .workers(4)
            .build();

        assert!(opts.include_private);
        assert!(opts.include_implementation);
        assert_eq!(opts.workers, 4);
    }

    #[test]
    fn test_worker_count_auto() {
        let opts = ProcessOptions::default();
        let count = opts.worker_count();
        assert!(count > 0);
        assert!(count <= num_cpus::get());
    }

    #[test]
    fn test_worker_count_explicit() {
        let opts = ProcessOptions::builder().workers(8).build();
        assert_eq!(opts.worker_count(), 8);
    }
}
