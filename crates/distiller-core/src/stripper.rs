//! Stripper visitor for filtering IR nodes
//!
//! Minimal implementation for Phase 2 - will be enhanced when language processors are added.

use crate::{
    ProcessOptions,
    ir::{Class, Enum, File, Function, Interface, Node, Package, Struct, Visitor},
};

/// Stripper visitor (placeholder for Phase 2)
pub struct Stripper {
    _options: ProcessOptions,
}

impl Stripper {
    #[must_use]
    pub fn new(options: ProcessOptions) -> Self {
        Self { _options: options }
    }
}

impl Visitor for Stripper {
    fn visit_node(&mut self, node: &mut Node) {
        match node {
            Node::File(f) => self.visit_file(f),
            Node::Class(c) => self.visit_class(c),
            Node::Interface(i) => self.visit_interface(i),
            Node::Struct(s) => self.visit_struct(s),
            Node::Enum(e) => self.visit_enum(e),
            Node::Package(p) => self.visit_package(p),
            _ => {}
        }
    }

    fn visit_file(&mut self, file: &mut File) {
        // Recurse into children
        for child in &mut file.children {
            self.visit_node(child);
        }
    }

    fn visit_package(&mut self, package: &mut Package) {
        for child in &mut package.children {
            self.visit_node(child);
        }
    }

    fn visit_class(&mut self, class: &mut Class) {
        for child in &mut class.children {
            self.visit_node(child);
        }
    }

    fn visit_interface(&mut self, interface: &mut Interface) {
        for child in &mut interface.children {
            self.visit_node(child);
        }
    }

    fn visit_struct(&mut self, strukt: &mut Struct) {
        for child in &mut strukt.children {
            self.visit_node(child);
        }
    }

    fn visit_enum(&mut self, enm: &mut Enum) {
        for child in &mut enm.children {
            self.visit_node(child);
        }
    }

    fn visit_function(&mut self, _function: &mut Function) {
        // Functions don't have children in current IR
    }
}

/// Apply stripper to a node tree (placeholder)
#[must_use]
pub fn strip(node: &Node, _options: &ProcessOptions) -> Node {
    node.clone()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stripper_creation() {
        let opts = ProcessOptions::default();
        let _stripper = Stripper::new(opts);
        // Basic creation test
    }

    #[test]
    fn test_strip_identity() {
        let file = File {
            path: "test.rs".to_string(),
            children: vec![],
        };
        let opts = ProcessOptions::default();
        let stripped = strip(&Node::File(file.clone()), &opts);

        if let Node::File(f) = stripped {
            assert_eq!(f.path, file.path);
        }
    }
}
