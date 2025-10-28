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

/// Extract File nodes from an IR Node (recursive for Directory)
pub fn extract_files(node: &Node) -> Vec<File> {
    let mut files = Vec::new();

    match node {
        Node::File(file) => {
            files.push(file.clone());
        }
        Node::Directory(dir) => {
            for child in &dir.children {
                files.extend(extract_files(child));
            }
        }
        _ => {}
    }

    files
}
