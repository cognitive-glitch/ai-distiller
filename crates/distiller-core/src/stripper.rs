//! Stripper visitor for filtering IR nodes
//!
//! Applies ProcessOptions to filter IR based on visibility and content preferences.

use crate::{
    ProcessOptions,
    ir::{
        Class, Enum, Field, File, Function, Interface, Node, Package, Struct, TypeAlias,
        Visibility, Visitor,
    },
};

/// Stripper visitor - filters IR nodes based on ProcessOptions
pub struct Stripper {
    options: ProcessOptions,
}

impl Stripper {
    #[must_use]
    pub fn new(options: ProcessOptions) -> Self {
        Self { options }
    }

    /// Check if a visibility level should be included
    fn should_include_visibility(&self, visibility: Visibility) -> bool {
        match visibility {
            Visibility::Public => self.options.include_public,
            Visibility::Protected => self.options.include_protected,
            Visibility::Internal => self.options.include_internal,
            Visibility::Private => self.options.include_private,
        }
    }

    /// Check if a node should be included based on type and options
    fn should_include_node(&self, node: &Node) -> bool {
        match node {
            Node::Import(_) => self.options.include_imports,
            Node::Comment(c) => {
                if c.format == "doc" {
                    self.options.include_docstrings
                } else {
                    self.options.include_comments
                }
            }
            Node::Function(f) => {
                self.options.include_methods && self.should_include_visibility(f.visibility)
            }
            Node::Field(f) => {
                self.options.include_fields && self.should_include_visibility(f.visibility)
            }
            Node::Class(c) => self.should_include_visibility(c.visibility),
            Node::Interface(i) => self.should_include_visibility(i.visibility),
            Node::Struct(s) => self.should_include_visibility(s.visibility),
            Node::Enum(e) => self.should_include_visibility(e.visibility),
            Node::TypeAlias(t) => self.should_include_visibility(t.visibility),
            _ => true, // Include other node types by default
        }
    }

    /// Filter decorators if annotations are disabled
    fn filter_decorators(&self, decorators: &mut Vec<String>) {
        if !self.options.include_annotations {
            decorators.clear();
        }
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
            Node::Function(f) => self.visit_function(f),
            Node::Field(f) => self.visit_field(f),
            Node::TypeAlias(t) => self.visit_type_alias(t),
            _ => {}
        }
    }

    fn visit_file(&mut self, file: &mut File) {
        // Filter children based on options
        file.children
            .retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut file.children {
            self.visit_node(child);
        }
    }

    fn visit_package(&mut self, package: &mut Package) {
        // Filter children
        package
            .children
            .retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut package.children {
            self.visit_node(child);
        }
    }

    fn visit_class(&mut self, class: &mut Class) {
        // Filter decorators
        self.filter_decorators(&mut class.decorators);

        // Filter children
        class
            .children
            .retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut class.children {
            self.visit_node(child);
        }
    }

    fn visit_interface(&mut self, interface: &mut Interface) {
        // Filter children
        interface
            .children
            .retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut interface.children {
            self.visit_node(child);
        }
    }

    fn visit_struct(&mut self, strukt: &mut Struct) {
        // Filter children
        strukt
            .children
            .retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut strukt.children {
            self.visit_node(child);
        }
    }

    fn visit_enum(&mut self, enm: &mut Enum) {
        // Filter children (enum variants)
        enm.children.retain(|child| self.should_include_node(child));

        // Recurse into remaining children
        for child in &mut enm.children {
            self.visit_node(child);
        }
    }

    fn visit_function(&mut self, function: &mut Function) {
        // Filter decorators
        self.filter_decorators(&mut function.decorators);

        // Remove implementation if not included
        if !self.options.include_implementation {
            function.implementation = None;
        }
    }

    fn visit_field(&mut self, _field: &mut Field) {
        // Fields don't have children, nothing to do
    }

    fn visit_type_alias(&mut self, _type_alias: &mut TypeAlias) {
        // Type aliases don't have children, nothing to do
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ir::{Function, Visibility};

    #[test]
    fn test_stripper_creation() {
        let opts = ProcessOptions::default();
        let _stripper = Stripper::new(opts);
    }

    #[test]
    fn test_visibility_filtering() {
        let opts = ProcessOptions {
            include_public: true,
            include_private: false,
            ..Default::default()
        };
        let stripper = Stripper::new(opts);

        assert!(stripper.should_include_visibility(Visibility::Public));
        assert!(!stripper.should_include_visibility(Visibility::Private));
    }

    #[test]
    fn test_implementation_removal() {
        let mut opts = ProcessOptions::default();
        opts.include_implementation = false;

        let mut stripper = Stripper::new(opts);
        let mut func = Function {
            name: "test".to_string(),
            visibility: Visibility::Public,
            modifiers: vec![],
            decorators: vec![],
            type_params: vec![],
            parameters: vec![],
            return_type: None,
            implementation: Some("return 42;".to_string()),
            line_start: 1,
            line_end: 3,
        };

        stripper.visit_function(&mut func);
        assert!(func.implementation.is_none());
    }
}
