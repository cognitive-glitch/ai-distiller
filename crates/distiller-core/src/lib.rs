//! # distiller-core
//!
//! Core library for AI Distiller - extracts essential code structure from large codebases
//! for LLM consumption.
//!
//! ## Architecture
//!
//! - **IR (Intermediate Representation)**: Language-agnostic AST representation
//! - **Processor**: File and directory processing with rayon parallelism
//! - **Stripper**: Visitor-based filtering of IR nodes
//! - **Language Processors**: Per-language parsers using tree-sitter
//!
//! ## Concurrency Model
//!
//! This crate uses **rayon** for CPU parallelism, NOT tokio/async.
//! All operations are synchronous for simplicity and performance.

pub mod error;
pub mod ir;
pub mod options;
pub mod processor;

// Re-exports
pub use error::{DistilError, Result};
pub use options::ProcessOptions;
