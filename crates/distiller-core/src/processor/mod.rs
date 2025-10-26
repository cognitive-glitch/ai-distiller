//! File and directory processing
//!
//! The processor is responsible for:
//! - Walking directories
//! - Detecting file languages
//! - Dispatching to language processors
//! - Parallel processing with rayon

use crate::{ir::Node, Result};
use std::path::Path;

pub mod language;

pub use language::LanguageProcessor;

/// Main processor for files and directories
pub struct Processor {
    // Processor configuration will be added in next phase
}

impl Processor {
    /// Create a new processor
    pub fn new() -> Self {
        Self {}
    }

    /// Process a file or directory
    pub fn process_path(&self, _path: &Path) -> Result<Node> {
        // Implementation will be added in Phase 2
        todo!("process_path implementation")
    }
}

impl Default for Processor {
    fn default() -> Self {
        Self::new()
    }
}
