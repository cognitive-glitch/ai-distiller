use distiller_core::{
    error::{DistilError, Result},
    ir::{
        Class, Field, File, Function, Import, Interface, Modifier, Node, Parameter, TypeRef,
        Visibility,
    },
    options::ProcessOptions,
    processor::language::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;

pub struct RustProcessor {
    parser: Arc<Mutex<tree_sitter::Parser>>,
}

impl RustProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = tree_sitter::Parser::new();
        parser
            .set_language(&tree_sitter_rust::LANGUAGE.into())
            .map_err(|e| DistilError::TreeSitter(format!("Failed to set Rust language: {}", e)))?;
        Ok(Self {
            parser: Arc::new(Mutex::new(parser)),
        })
    }

    fn node_text(&self, node: tree_sitter::Node, source: &str) -> String {
        if node.start_byte() > node.end_byte() || node.end_byte() > source.len() {
            return String::new();
        }
        source[node.start_byte()..node.end_byte()].to_string()
    }

    fn parse_visibility(&self, node: tree_sitter::Node, source: &str) -> Visibility {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "visibility_modifier" {
                let text = self.node_text(child, source);
                return match text.as_str() {
                    "pub" => Visibility::Public,
                    text if text.contains("pub(crate)") => Visibility::Internal,
                    text if text.contains("pub(super)") || text.contains("pub(in ") => {
                        Visibility::Protected
                    }
                    _ => Visibility::Public,
                };
            }
        }
        Visibility::Private // Default in Rust
    }

    fn parse_field(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = TypeRef::new("".to_string());
        let visibility = self.parse_visibility(node, source);
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_identifier" | "primitive_type" | "generic_type" => {
                    field_type = TypeRef::new(self.node_text(child, source));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Field {
            name,
            visibility,
            field_type: Some(field_type),
            modifiers: vec![],
            default_value: None,
            line,
        }))
    }

    fn parse_parameters(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Parameter>> {
        let mut parameters = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "parameter" || child.kind() == "self_parameter" {
                if let Some(param) = self.parse_parameter(child, source)? {
                    parameters.push(param);
                }
            }
        }

        Ok(parameters)
    }

    fn parse_parameter(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Parameter>> {
        let mut name = String::new();
        let mut param_type = TypeRef::new("".to_string());

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "self" => {
                    name = "self".to_string();
                }
                "type_identifier" | "primitive_type" | "reference_type" | "generic_type" => {
                    param_type = TypeRef::new(self.node_text(child, source));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Parameter {
            name,
            param_type,
            is_variadic: false,
            is_optional: false,
            decorators: vec![],
            default_value: None,
        }))
    }

    fn parse_impl_block(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<(String, Vec<Function>)> {
        let mut type_name = String::new();
        let mut methods = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if type_name.is_empty() {
                        type_name = self.node_text(child, source);
                    }
                }
                "declaration_list" => {
                    let mut decl_cursor = child.walk();
                    for decl_child in child.children(&mut decl_cursor) {
                        if decl_child.kind() == "function_item" {
                            if let Some(method) = self.parse_function(decl_child, source)? {
                                methods.push(method);
                            }
                        }
                    }
                }
                _ => {}
            }
        }

        Ok((type_name, methods))
    }

    fn parse_use(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        let line = node.start_position().row + 1;

        // Extract module path
        let mut module = String::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "use_clause" || child.kind() == "scoped_identifier" {
                module = self.node_text(child, source);
                break;
            }
        }

        Ok(Some(Import {
            import_type: "use".to_string(),
            module,
            symbols: vec![],
            is_type: false,
            line: Some(line),
        }))
    }

    fn parse_struct(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut fields = Vec::new();
        let visibility = self.parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "field_declaration_list" => {
                    let mut field_cursor = child.walk();
                    for field_child in child.children(&mut field_cursor) {
                        if field_child.kind() == "field_declaration" {
                            if let Some(field) = self.parse_field(field_child, source)? {
                                fields.push(field);
                            }
                        }
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Class {
            name,
            visibility,
            extends: vec![],
            implements: vec![],
            type_params: vec![],
            decorators: vec![],
            modifiers: vec![],
            children: fields.into_iter().map(Node::Field).collect(),
            line_start,
            line_end,
        }))
    }

    fn parse_trait(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Interface>> {
        let mut name = String::new();
        let visibility = self.parse_visibility(node, source);
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_identifier" {
                name = self.node_text(child, source);
                break;
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Interface {
            name,
            visibility,
            extends: vec![],
            type_params: vec![],
            children: vec![],
            line_start,
            line_end,
        }))
    }

    fn parse_function(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let mut is_async = false;
        let visibility = self.parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "type_identifier" | "primitive_type" | "generic_type" => {
                    // This might be return type
                    return_type = Some(TypeRef::new(self.node_text(child, source)));
                }
                "function_modifiers" => {
                    // Check for async in function_modifiers
                    let text = self.node_text(child, source);
                    if text.contains("async") {
                        is_async = true;
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let mut modifiers = vec![];
        if is_async {
            modifiers.push(Modifier::Async);
        }

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type,
            decorators: vec![],
            type_params: vec![],
            modifiers,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn process_node(&self, node: tree_sitter::Node, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "use_declaration" => {
                if let Some(import) = self.parse_use(node, source)? {
                    file.children.push(Node::Import(import));
                }
            }
            "struct_item" => {
                if let Some(struct_def) = self.parse_struct(node, source)? {
                    file.children.push(Node::Class(struct_def));
                }
            }
            "trait_item" => {
                if let Some(trait_def) = self.parse_trait(node, source)? {
                    file.children.push(Node::Interface(trait_def));
                }
            }
            "function_item" => {
                if let Some(func) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(func));
                }
            }
            "impl_item" => {
                // Handle impl blocks in second pass
            }
            _ => {
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.process_node(child, source, file)?;
                }
            }
        }
        Ok(())
    }

    fn associate_impl_blocks(
        &self,
        node: tree_sitter::Node,
        source: &str,
        file: &mut File,
    ) -> Result<()> {
        match node.kind() {
            "impl_item" => {
                let (type_name, methods) = self.parse_impl_block(node, source)?;
                if !type_name.is_empty() {
                    // Find the struct and add methods
                    for child in &mut file.children {
                        if let Node::Class(ref mut class) = child {
                            if class.name == type_name {
                                // Add methods to the struct
                                for method in methods {
                                    class.children.push(Node::Function(method));
                                }
                                break;
                            }
                        }
                    }
                }
            }
            _ => {
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.associate_impl_blocks(child, source, file)?;
                }
            }
        }
        Ok(())
    }
}

impl LanguageProcessor for RustProcessor {
    fn language(&self) -> &'static str {
        "rust"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["rs"]
    }

    fn can_process(&self, path: &Path) -> bool {
        if let Some(ext) = path.extension() {
            if let Some(ext_str) = ext.to_str() {
                return self.supported_extensions().contains(&ext_str);
            }
        }
        false
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(
                path.to_string_lossy().as_ref(),
                "Failed to parse Rust source",
            )
        })?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: vec![],
        };

        // First pass: collect structs, traits, functions, imports
        self.process_node(tree.root_node(), source, &mut file)?;

        // Second pass: associate impl blocks with structs
        self.associate_impl_blocks(tree.root_node(), source, &mut file)?;

        Ok(file)
    }
}

impl Default for RustProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create RustProcessor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let processor = RustProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = RustProcessor::new().unwrap();
        let extensions = processor.supported_extensions();
        assert_eq!(extensions, &["rs"]);

        assert!(processor.can_process(Path::new("main.rs")));
        assert!(processor.can_process(Path::new("lib.rs")));
        assert!(!processor.can_process(Path::new("main.go")));
        assert!(!processor.can_process(Path::new("test.js")));
    }

    #[test]
    fn test_use_statements() {
        let processor = RustProcessor::new().unwrap();
        let source = r#"
use std::collections::HashMap;
use std::io::{Read, Write};
use super::models::User;
use crate::utils::*;
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.rs"), &opts)
            .unwrap();

        let imports: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Import(imp) = n {
                    Some(imp)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(imports.len(), 4);
    }

    #[test]
    fn test_struct_with_impl() {
        let processor = RustProcessor::new().unwrap();
        let source = r#"
pub struct User {
    pub id: u64,
    pub name: String,
    email: String,
}

impl User {
    pub fn new(id: u64, name: String, email: String) -> Self {
        Self { id, name, email }
    }

    pub(crate) fn get_email(&self) -> &str {
        &self.email
    }

    fn validate(&self) -> bool {
        !self.email.is_empty()
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.rs"), &opts)
            .unwrap();

        let classes: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Class(cls) = n {
                    Some(cls)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(classes.len(), 1);
        assert_eq!(classes[0].name, "User");
        assert_eq!(classes[0].visibility, Visibility::Public);

        // Should have 3 methods
        let methods: Vec<_> = classes[0]
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(func) = n {
                    Some(func)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(methods.len(), 3);
        assert_eq!(methods[0].name, "new");
        assert_eq!(methods[0].visibility, Visibility::Public);
        assert_eq!(methods[1].name, "get_email");
        assert_eq!(methods[1].visibility, Visibility::Internal);
        assert_eq!(methods[2].name, "validate");
        assert_eq!(methods[2].visibility, Visibility::Private);
    }

    #[test]
    fn test_trait_definitions() {
        let processor = RustProcessor::new().unwrap();
        let source = r#"
pub trait Validator {
    fn validate(&self) -> bool;
    fn is_valid(&self) -> bool {
        self.validate()
    }
}

trait Display {
    fn display(&self) -> String;
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.rs"), &opts)
            .unwrap();

        let traits: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Interface(iface) = n {
                    Some(iface)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(traits.len(), 2);
        assert_eq!(traits[0].name, "Validator");
        assert_eq!(traits[0].visibility, Visibility::Public);
        assert_eq!(traits[1].name, "Display");
        assert_eq!(traits[1].visibility, Visibility::Private);
    }

    #[test]
    fn test_function_declarations() {
        let processor = RustProcessor::new().unwrap();
        let source = r#"
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}

pub(crate) fn internal_helper() -> String {
    String::from("helper")
}

fn private_compute(x: f64, y: f64) -> f64 {
    x * y
}

pub async fn fetch_data(url: &str) -> Result<String, Error> {
    // async implementation
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.rs"), &opts)
            .unwrap();

        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(func) = n {
                    Some(func)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 4);

        // add - public
        assert_eq!(functions[0].name, "add");
        assert_eq!(functions[0].visibility, Visibility::Public);
        assert_eq!(functions[0].parameters.len(), 2);

        // internal_helper - pub(crate)
        assert_eq!(functions[1].name, "internal_helper");
        assert_eq!(functions[1].visibility, Visibility::Internal);

        // private_compute - private
        assert_eq!(functions[2].name, "private_compute");
        assert_eq!(functions[2].visibility, Visibility::Private);

        // fetch_data - async
        assert_eq!(functions[3].name, "fetch_data");
        assert_eq!(functions[3].visibility, Visibility::Public);
        assert!(functions[3].modifiers.contains(&Modifier::Async));
    }
}
