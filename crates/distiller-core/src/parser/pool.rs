//! Thread-safe parser pool for tree-sitter parsers
//!
//! Manages a pool of tree-sitter parsers per language to avoid
//! expensive parser creation on every file. Uses `parking_lot` for
//! efficient locking.

use crate::error::{DistilError, Result};
use parking_lot::Mutex;
use std::collections::HashMap;
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};
use tree_sitter::{Language, Parser};

/// Parser pool statistics
#[derive(Debug, Clone, Default)]
pub struct PoolStats {
    /// Number of times a parser was acquired from pool
    pub hits: usize,
    /// Number of times a new parser was created
    pub misses: usize,
    /// Total parsers created
    pub created: usize,
    /// Total parsers reused
    pub reused: usize,
}

/// A thread-safe pool of tree-sitter parsers
///
/// Parsers are expensive to create, so we pool them per language.
/// Each language gets its own stack of available parsers.
#[derive(Clone)]
pub struct ParserPool {
    inner: Arc<Mutex<PoolInner>>,
    stats: Arc<PoolStatsInner>,
}

struct PoolStatsInner {
    hits: AtomicUsize,
    misses: AtomicUsize,
    created: AtomicUsize,
    reused: AtomicUsize,
}

impl Default for PoolStatsInner {
    fn default() -> Self {
        Self {
            hits: AtomicUsize::new(0),
            misses: AtomicUsize::new(0),
            created: AtomicUsize::new(0),
            reused: AtomicUsize::new(0),
        }
    }
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
    #[must_use]
    pub fn new(max_per_language: usize) -> Self {
        Self {
            inner: Arc::new(Mutex::new(PoolInner {
                pools: HashMap::new(),
                max_per_language: max_per_language.max(1),
            })),
            stats: Arc::new(PoolStatsInner::default()),
        }
    }

    /// Get current pool statistics
    #[must_use]
    pub fn stats(&self) -> PoolStats {
        PoolStats {
            hits: self.stats.hits.load(Ordering::Relaxed),
            misses: self.stats.misses.load(Ordering::Relaxed),
            created: self.stats.created.load(Ordering::Relaxed),
            reused: self.stats.reused.load(Ordering::Relaxed),
        }
    }

    /// Acquire a parser for the given language
    ///
    /// Returns a guard that automatically returns the parser when dropped.
    ///
    /// # Arguments
    /// * `language_name` - Name of the language (e.g., "python", "rust")
    /// * `language_fn` - Function to get tree-sitter Language (only called if needed)
    ///
    /// # Errors
    ///
    /// Returns an error if the language cannot be loaded or set on the parser.
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

            self.stats.hits.fetch_add(1, Ordering::Relaxed);
            self.stats.reused.fetch_add(1, Ordering::Relaxed);

            return Ok(ParserGuard {
                parser: Some(parser),
                language_name: language_name.to_string(),
                pool: self.clone(),
            });
        }

        // No available parser, create new one
        self.stats.misses.fetch_add(1, Ordering::Relaxed);
        self.stats.created.fetch_add(1, Ordering::Relaxed);

        drop(inner); // Release lock during expensive operation

        let language = language_fn()?;
        let mut parser = Parser::new();

        parser
            .set_language(&language)
            .map_err(|e| DistilError::TreeSitter(format!("Failed to set language: {e}")))?;

        Ok(ParserGuard {
            parser: Some(parser),
            language_name: language_name.to_string(),
            pool: self.clone(),
        })
    }

    /// Return a parser to the pool
    ///
    /// Called automatically by `ParserGuard::drop`, but can be called manually
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
    ///
    /// # Panics
    ///
    /// Panics if the parser has already been returned to the pool (should never happen).
    pub fn get_mut(&mut self) -> &mut Parser {
        self.parser
            .as_mut()
            .expect("Parser should always be Some during lifetime")
    }

    /// Get immutable reference to the parser
    ///
    /// # Panics
    ///
    /// Panics if the parser has already been returned to the pool (should never happen).
    #[must_use]
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_pool_creation() {
        let pool = ParserPool::new(10);
        let stats = pool.stats();

        assert_eq!(stats.hits, 0);
        assert_eq!(stats.misses, 0);
        assert_eq!(stats.created, 0);
        assert_eq!(stats.reused, 0);
    }

    #[test]
    fn test_pool_default() {
        let pool = ParserPool::default();
        let stats = pool.stats();

        assert_eq!(stats.hits, 0);
        assert_eq!(stats.misses, 0);
    }

    #[test]
    fn test_pool_stats_empty() {
        let pool = ParserPool::new(5);
        let stats = pool.stats();

        assert_eq!(stats.hits, 0);
        assert_eq!(stats.misses, 0);
        assert_eq!(stats.created, 0);
        assert_eq!(stats.reused, 0);
    }

    // Note: Full integration tests require actual tree-sitter language libraries
    // These will be added when language processors are implemented
}
