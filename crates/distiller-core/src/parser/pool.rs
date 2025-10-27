//! Thread-safe parser pool for tree-sitter parsers
//!
//! Manages a pool of tree-sitter parsers per language to avoid
//! expensive parser creation on every file. Uses parking_lot for
//! efficient locking.

use crate::error::{DistilError, Result};
use parking_lot::Mutex;
use std::collections::HashMap;
use std::sync::Arc;
use tree_sitter::{Language, Parser};

/// A thread-safe pool of tree-sitter parsers
///
/// Parsers are expensive to create, so we pool them per language.
/// Each language gets its own stack of available parsers.
#[derive(Clone)]
pub struct ParserPool {
    inner: Arc<Mutex<PoolInner>>,
}

struct PoolInner {
    /// Available parsers per language name
    pools: HashMap<String, Vec<Parser>>,
    /// Maximum parsers per language (prevents unbounded growth)
    max_per_language: usize,
}

impl ParserPool {
    /// Create a new parser pool
    ///
    /// # Arguments
    /// * `max_per_language` - Maximum parsers to cache per language (default: 32)
    pub fn new(max_per_language: usize) -> Self {
        Self {
            inner: Arc::new(Mutex::new(PoolInner {
                pools: HashMap::new(),
                max_per_language: max_per_language.max(1),
            })),
        }
    }

    /// Acquire a parser for the given language
    ///
    /// Returns a guard that automatically returns the parser when dropped.
    ///
    /// # Arguments
    /// * `language_name` - Name of the language (e.g., "python", "rust")
    /// * `language_fn` - Function to get tree-sitter Language (only called if needed)
    pub fn acquire<F>(&self, language_name: &str, language_fn: F) -> Result<ParserGuard>
    where
        F: FnOnce() -> Result<Language>,
    {
        let mut inner = self.inner.lock();

        // Try to get existing parser from pool
        if let Some(pool) = inner.pools.get_mut(language_name)
            && let Some(mut parser) = pool.pop()
        {
            // Reset parser state for reuse
            parser.reset();

            return Ok(ParserGuard {
                parser: Some(parser),
                language_name: language_name.to_string(),
                pool: self.clone(),
            });
        }

        // No available parser, create new one
        drop(inner); // Release lock during expensive operation

        let language = language_fn()?;
        let mut parser = Parser::new();

        parser
            .set_language(&language)
            .map_err(|e| DistilError::TreeSitter(format!("Failed to set language: {}", e)))?;

        Ok(ParserGuard {
            parser: Some(parser),
            language_name: language_name.to_string(),
            pool: self.clone(),
        })
    }

    /// Return a parser to the pool
    ///
    /// Called automatically by ParserGuard::drop, but can be called manually
    /// for early return.
    fn release(&self, language_name: String, parser: Parser) {
        let mut inner = self.inner.lock();

        // Capture max before mutable borrow
        let max = inner.max_per_language;
        let pool = inner.pools.entry(language_name).or_default();

        // Only keep parser if we haven't exceeded max
        if pool.len() < max {
            pool.push(parser);
        }
        // Otherwise drop the parser (happens automatically)
    }

    /// Get current pool statistics (for debugging/monitoring)
    pub fn stats(&self) -> PoolStats {
        let inner = self.inner.lock();

        let mut stats = PoolStats {
            languages: Vec::new(),
            total_parsers: 0,
        };

        for (lang, pool) in &inner.pools {
            stats.languages.push(LanguageStats {
                name: lang.clone(),
                available: pool.len(),
            });
            stats.total_parsers += pool.len();
        }

        stats.languages.sort_by(|a, b| a.name.cmp(&b.name));
        stats
    }
}

impl Default for ParserPool {
    /// Default: 32 parsers per language (good for high parallelism)
    fn default() -> Self {
        Self::new(32)
    }
}

/// RAII guard for a parser
///
/// Automatically returns parser to pool when dropped.
pub struct ParserGuard {
    parser: Option<Parser>,
    language_name: String,
    pool: ParserPool,
}

impl ParserGuard {
    /// Get mutable reference to the parser
    pub fn get_mut(&mut self) -> &mut Parser {
        self.parser
            .as_mut()
            .expect("Parser should always be Some during lifetime")
    }

    /// Get immutable reference to the parser
    pub fn get(&self) -> &Parser {
        self.parser
            .as_ref()
            .expect("Parser should always be Some during lifetime")
    }
}

impl Drop for ParserGuard {
    fn drop(&mut self) {
        if let Some(parser) = self.parser.take() {
            self.pool.release(self.language_name.clone(), parser);
        }
    }
}

/// Statistics about parser pool usage
#[derive(Debug, Clone)]
pub struct PoolStats {
    pub languages: Vec<LanguageStats>,
    pub total_parsers: usize,
}

#[derive(Debug, Clone)]
pub struct LanguageStats {
    pub name: String,
    pub available: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_pool_creation() {
        let pool = ParserPool::new(10);
        let stats = pool.stats();

        assert_eq!(stats.total_parsers, 0);
        assert_eq!(stats.languages.len(), 0);
    }

    #[test]
    fn test_pool_default() {
        let pool = ParserPool::default();
        let stats = pool.stats();

        assert_eq!(stats.total_parsers, 0);
    }

    #[test]
    fn test_pool_stats_empty() {
        let pool = ParserPool::new(5);
        let stats = pool.stats();

        assert_eq!(stats.total_parsers, 0);
        assert_eq!(stats.languages.len(), 0);
    }

    // Note: Full integration tests require actual tree-sitter language libraries
    // These will be added when language processors are implemented
}
