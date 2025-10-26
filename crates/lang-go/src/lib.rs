use distiller_core::error::Result;
use distiller_core::{
    error::DistilError,
    ir::{
        Class, Field, File, Function, Import, Interface, Modifier, Node,
        Parameter, TypeParam, TypeRef, Visibility,
    },
    options::ProcessOptions,
    processor::language::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use tree_sitter::Parser;

pub struct GoProcessor {
    parser: Mutex<Parser>,
}

impl GoProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        let language = tree_sitter_go::LANGUAGE;
        parser
            .set_language(&language.into())
            .map_err(|e| DistilError::parse_error("", format!("Failed to set Go language: {}", e)))?;

        Ok(Self {
            parser: Mutex::new(parser),
        })
    }

    fn node_text(&self, node: tree_sitter::Node, source: &str) -> String {
        let start = node.start_byte();
        let end = node.end_byte();
        let source_len = source.len();
        if start > end || end > source_len {
            return String::new();
        }
        source[start..end].to_string()
    }

    fn parse_imports(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Import>> {
        let mut imports = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "import_declaration" {
                imports.extend(self.parse_import_declaration(child, source)?);
            }
        }

        Ok(imports)
    }

    fn parse_import_declaration(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<Import>> {
        let mut imports = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "import_spec" => {
                    if let Some(import) = self.parse_import_spec(child, source)? {
                        imports.push(import);
                    }
                }
                "import_spec_list" => {
                    let mut list_cursor = child.walk();
                    for list_child in child.children(&mut list_cursor) {
                        if list_child.kind() == "import_spec" {
                            if let Some(import) = self.parse_import_spec(list_child, source)? {
                                imports.push(import);
                            }
                        }
                    }
                }
                _ => {}
            }
        }

        Ok(imports)
    }

    fn parse_import_spec(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        let mut module = String::new();
        let mut import_type = "import".to_string();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "interpreted_string_literal" => {
                    module = self.node_text(child, source).trim_matches('"').to_string();
                }
                "package_identifier" => {
                    let alias = self.node_text(child, source);
                    import_type = format!("import {} as", alias);
                }
                "dot" => {
                    import_type = "import .".to_string();
                }
                _ => {}
            }
        }

        if !module.is_empty() {
            Ok(Some(Import {
                import_type,
                module,
                symbols: vec![],
                is_type: false,
                line: Some(node.start_position().row + 1),
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_struct(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut fields = Vec::new();
        let mut type_params = Vec::new();

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "struct_type" => {
                    let mut struct_cursor = child.walk();
                    for struct_child in child.children(&mut struct_cursor) {
                        if struct_child.kind() == "field_declaration_list" {
                            fields = self.parse_struct_fields(struct_child, source)?;
                        }
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if name.chars().next().unwrap_or('a').is_uppercase() {
            Visibility::Public
        } else {
            Visibility::Internal
        };

        Ok(Some(Class {
            name,
            visibility,
            decorators: vec![],
            type_params,
            extends: vec![],
            implements: vec![],
            children: fields.into_iter().map(Node::Field).collect(),
            modifiers: vec![],
            line_start,
            line_end,
        }))
    }

    fn parse_struct_fields(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<Field>> {
        let mut fields = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "field_declaration" {
                if let Some(field) = self.parse_field_declaration(child, source)? {
                    fields.push(field);
                }
            }
        }

        Ok(fields)
    }

    fn parse_field_declaration(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;

        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
                | "slice_type" | "map_type" => {
                    if name.is_empty() {
                        // Embedded field: type name IS the field name
                        name = self.node_text(child, source);
                        field_type = Some(TypeRef::new(name.clone()));
                    } else {
                        field_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if name.chars().next().unwrap_or('a').is_uppercase() {
            Visibility::Public
        } else {
            Visibility::Internal
        };

        Ok(Some(Field {
            name,
            visibility,
            field_type,
            modifiers: vec![],
            default_value: None,
            line,
        }))
    }

    fn parse_interface(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Interface>> {
        let mut name = String::new();
        let mut methods = Vec::new();
        let mut type_params = Vec::new();

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "interface_type" => {
                    let mut interface_cursor = child.walk();
                    for interface_child in child.children(&mut interface_cursor) {
                        if interface_child.kind() == "method_elem" {
                            if let Some(method) = self.parse_method_spec(interface_child, source)? {
                                methods.push(method);
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

        let visibility = if name.chars().next().unwrap_or('a').is_uppercase() {
            Visibility::Public
        } else {
            Visibility::Internal
        };

        Ok(Some(Interface {
            name,
            visibility,
            type_params,
            extends: vec![],
            children: methods.into_iter().map(Node::Function).collect(),
            line_start,
            line_end,
        }))
    }


    fn parse_method_spec(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_identifier" => {
                    name = self.node_text(child, source);
                }
                "parameter_list" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "parameter_declaration" => {
                    if let Some(type_ref) = self.extract_type_from_parameter(child, source)? {
                        return_type = Some(type_ref);
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if name.chars().next().unwrap_or('a').is_uppercase() {
            Visibility::Public
        } else {
            Visibility::Internal
        };

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type,
            decorators: vec![],
            type_params: vec![],
            modifiers: vec![],
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_function(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let mut type_params = Vec::new();
        let mut receiver_type = None;
        let mut has_seen_name = false;

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "field_identifier" => {
                    name = self.node_text(child, source);
                    has_seen_name = true;
                }
                "parameter_list" => {
                    // If we haven't seen the name yet, this is a receiver (for methods)
                    // If we have seen the name, this is the parameter list
                    if !has_seen_name {
                        let receiver_params = self.parse_parameters(child, source)?;
                        if !receiver_params.is_empty() {
                            receiver_type = Some(receiver_params[0].param_type.clone());
                        }
                    } else {
                        parameters = self.parse_parameters(child, source)?;
                    }
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                _ => {
                    if let Some(type_ref) = self.try_extract_type(child, source) {
                        return_type = Some(type_ref);
                    }
                }
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if name.chars().next().unwrap_or('a').is_uppercase() {
            Visibility::Public
        } else {
            Visibility::Internal
        };

        let modifiers = if receiver_type.is_some() {
            vec![Modifier::Static]
        } else {
            vec![]
        };

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type,
            decorators: vec![],
            type_params,
            modifiers,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_parameters(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Parameter>> {
        let mut parameters = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "parameter_declaration" {
                parameters.extend(self.parse_parameter_declaration(child, source)?);
            }
        }

        Ok(parameters)
    }

    fn parse_parameter_declaration(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<Parameter>> {
        let mut names = Vec::new();
        let mut param_type = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "field_identifier" => {
                    names.push(self.node_text(child, source));
                }
                "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
                | "slice_type" | "map_type" | "channel_type" | "function_type"
                | "interface_type" | "struct_type" => {
                    param_type = Some(TypeRef::new(self.node_text(child, source)));
                }
                _ => {}
            }
        }

        // If no names found but type exists, it's an unnamed parameter
        if names.is_empty() && param_type.is_some() {
            return Ok(vec![Parameter {
                name: String::new(),
                param_type: param_type.unwrap(),
                default_value: None,
                is_variadic: false,
                is_optional: false,
                decorators: vec![],
            }]);
        }

        // Multiple names with same type (e.g., "a, b, c int")
        Ok(names
            .into_iter()
            .map(|name| Parameter {
                name,
                param_type: param_type.clone().unwrap_or_else(|| TypeRef::new("".to_string())),
                default_value: None,
                is_variadic: false,
                is_optional: false,
                decorators: vec![],
            })
            .collect())
    }

    fn extract_type_from_parameter(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<TypeRef>> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if let Some(type_ref) = self.try_extract_type(child, source) {
                return Ok(Some(type_ref));
            }
        }
        Ok(None)
    }

    fn try_extract_type(&self, node: tree_sitter::Node, source: &str) -> Option<TypeRef> {
        match node.kind() {
            "type_identifier" | "qualified_type" | "pointer_type" | "array_type" | "slice_type"
            | "map_type" | "channel_type" | "function_type" | "interface_type"
            | "struct_type" => Some(TypeRef::new(self.node_text(node, source))),
            _ => None,
        }
    }

    fn parse_type_parameters(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<TypeParam>> {
        let mut type_params = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter_declaration" {
                if let Some(tp) = self.parse_type_parameter(child, source)? {
                    type_params.push(tp);
                }
            }
        }

        Ok(type_params)
    }

    fn parse_type_parameter(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<TypeParam>> {
        let mut name = String::new();
        let mut constraint = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    } else {
                        constraint = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                "qualified_type" | "interface_type" => {
                    constraint = Some(TypeRef::new(self.node_text(child, source)));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(TypeParam { name, constraints: constraint.map(|c| vec![c]).unwrap_or_default(), default: None }))
    }

    fn process_node(&self, node: tree_sitter::Node, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "type_declaration" => {
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    if child.kind() == "type_spec" {
                        self.process_type_spec(child, source, file)?;
                    }
                }
            }
            "function_declaration" | "method_declaration" => {
                if let Some(func) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(func));
                }
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

    fn process_type_spec(&self, node: tree_sitter::Node, source: &str, file: &mut File) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "struct_type" => {
                    let parent = node;
                    if let Some(struct_node) = self.parse_struct(parent, source)? {
                        file.children.push(Node::Class(struct_node));
                    }
                }
                "interface_type" => {
                    let parent = node;
                    if let Some(interface_node) = self.parse_interface(parent, source)? {
                        file.children.push(Node::Interface(interface_node));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }
}

impl LanguageProcessor for GoProcessor {
    fn language(&self) -> &'static str {
        "go"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["go"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| ext == "go")
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser
            .parse(source, None)
            .ok_or_else(|| DistilError::parse_error(path.display().to_string(), "Failed to parse Go source"))?;

        let root_node = tree.root_node();

        let mut file = File {
            path: path.display().to_string(),
            children: vec![],
        };

        let imports = self.parse_imports(root_node, source)?;
        for import in imports {
            file.children.push(Node::Import(import));
        }

        self.process_node(root_node, source, &mut file)?;

        Ok(file)
    }
}

impl Default for GoProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create GoProcessor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let processor = GoProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_supported_extensions() {
        let processor = GoProcessor::new().unwrap();
        let extensions = processor.supported_extensions();
        assert_eq!(extensions, &["go"]);
    }

    #[test]
    fn test_can_process() {
        let processor = GoProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("main.go")));
        assert!(!processor.can_process(Path::new("main.py")));
    }

    #[test]
    fn test_import_statements() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

import "fmt"
import "os"
import (
    "context"
    "sync"
    stdlib "github.com/user/lib"
)
"#;

        let result = processor.process(source, Path::new("test.go"), &ProcessOptions::default());
        assert!(result.is_ok());

        let file = result.unwrap();
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

        assert_eq!(imports.len(), 5);
        assert_eq!(imports[0].module, "fmt");
        assert_eq!(imports[1].module, "os");
        assert_eq!(imports[2].module, "context");
        assert_eq!(imports[3].module, "sync");
        assert_eq!(imports[4].module, "github.com/user/lib");
        assert!(imports[4].import_type.contains("stdlib"));
    }

    #[test]
    fn test_struct_with_methods() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type User struct {
    ID   int
    Name string
    email string
}

func (u *User) GetName() string {
    return u.Name
}

func (u *User) setEmail(email string) {
    u.email = email
}
"#;

        let result = processor.process(source, Path::new("test.go"), &ProcessOptions::default());
        assert!(result.is_ok());

        let file = result.unwrap();


        let structs: Vec<_> = file
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

        assert_eq!(structs.len(), 1);
        assert_eq!(structs[0].name, "User");
        assert_eq!(structs[0].visibility, Visibility::Public);
        assert_eq!(structs[0].children.len(), 3);

        let methods: Vec<_> = file
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

        assert_eq!(methods.len(), 2);
        assert_eq!(methods[0].name, "GetName");
        assert_eq!(methods[0].visibility, Visibility::Public);
        assert_eq!(methods[1].name, "setEmail");
        assert_eq!(methods[1].visibility, Visibility::Internal);
    }

    #[test]
    fn test_interface_with_generics() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type Repository[T any] interface {
    FindByID(id int) T
    Save(entity T) error
    Delete(id int) error
}
"#;

        let result = processor.process(source, Path::new("test.go"), &ProcessOptions::default());
        assert!(result.is_ok());

        let file = result.unwrap();

        let interfaces: Vec<_> = file
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

        assert_eq!(interfaces.len(), 1);
        assert_eq!(interfaces[0].name, "Repository");
        assert_eq!(interfaces[0].type_params.len(), 1);
        assert_eq!(interfaces[0].type_params[0].name, "T");
        assert_eq!(interfaces[0].children.len(), 3);
    }
}
