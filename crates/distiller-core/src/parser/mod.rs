//! Parser infrastructure for tree-sitter based parsing
//!
//! This module provides:
//! - Thread-safe parser pooling
//! - Language grammar loading
//! - Source parsing utilities

pub mod pool;

pub use pool::{ParserGuard, ParserPool, PoolStats};
