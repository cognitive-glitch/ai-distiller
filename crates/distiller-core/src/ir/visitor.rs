//! Visitor pattern for IR traversal

use super::nodes::{
    Class, Comment, Directory, Enum, Field, File, Function, Import, Interface, Node, Package,
    RawContent, Struct, TypeAlias,
};

/// Visitor trait for IR node traversal
///
/// Implement this trait to create visitors that can traverse and modify IR trees.
/// Common use cases: stripping content, formatting, analysis.
pub trait Visitor {
    /// Visit a node
    fn visit_node(&mut self, node: &mut Node) {
        match node {
            Node::File(f) => self.visit_file(f),
            Node::Directory(d) => self.visit_directory(d),
            Node::Package(p) => self.visit_package(p),
            Node::Import(i) => self.visit_import(i),
            Node::Class(c) => self.visit_class(c),
            Node::Interface(i) => self.visit_interface(i),
            Node::Struct(s) => self.visit_struct(s),
            Node::Enum(e) => self.visit_enum(e),
            Node::TypeAlias(t) => self.visit_type_alias(t),
            Node::Function(f) => self.visit_function(f),
            Node::Field(f) => self.visit_field(f),
            Node::Comment(c) => self.visit_comment(c),
            Node::RawContent(r) => self.visit_raw_content(r),
        }
    }

    fn visit_file(&mut self, file: &mut File) {
        for child in &mut file.children {
            self.visit_node(child);
        }
    }

    fn visit_directory(&mut self, dir: &mut Directory) {
        for child in &mut dir.children {
            self.visit_node(child);
        }
    }

    fn visit_package(&mut self, pkg: &mut Package) {
        for child in &mut pkg.children {
            self.visit_node(child);
        }
    }

    fn visit_import(&mut self, _import: &mut Import) {}

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

    fn visit_enum(&mut self, enu: &mut Enum) {
        for child in &mut enu.children {
            self.visit_node(child);
        }
    }

    fn visit_type_alias(&mut self, _alias: &mut TypeAlias) {}
    fn visit_function(&mut self, _func: &mut Function) {}
    fn visit_field(&mut self, _field: &mut Field) {}
    fn visit_comment(&mut self, _comment: &mut Comment) {}
    fn visit_raw_content(&mut self, _raw: &mut RawContent) {}
}

impl Node {
    /// Accept a visitor
    pub fn accept<V: Visitor>(&mut self, visitor: &mut V) {
        visitor.visit_node(self);
    }
}
