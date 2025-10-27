//! Intermediate Representation (IR)
//!
//! Language-agnostic representation of code structure.
//! All language processors convert their ASTs to this unified IR.

mod nodes;
mod types;
mod visitor;

pub use nodes::*;
pub use types::*;
pub use visitor::*;
