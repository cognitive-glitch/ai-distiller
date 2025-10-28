//! # distiller-core
//!
//! Core library for AI Distiller - extracts essential code structure from large codebases
//! for LLM consumption.
//!
//! ## Architecture
//!
//! - **IR (Intermediate Representation)**: Language-agnostic AST representation
//! - **Parser**: Thread-safe tree-sitter parser pooling
//! - **Processor**: File and directory processing with rayon parallelism
//! - **Stripper**: Visitor-based filtering of IR nodes
//! - **Language Processors**: Per-language parsers using tree-sitter
//!
//! ## Concurrency Model
//!
//! This crate uses **rayon** for CPU parallelism, NOT tokio/async.
//! All operations are synchronous for simplicity and performance.
//!
//! **IMPORTANT**: No tokio or async types are allowed in this crate.

#![forbid(unsafe_code)]

pub mod error;
pub mod ir;
pub mod logging;
pub mod options;
pub mod parser;
pub mod processor;
pub mod stripper;

// Re-exports
pub use error::{DistilError, Result};
pub use options::ProcessOptions;
pub use parser::ParserPool;
pub use stripper::Stripper;
