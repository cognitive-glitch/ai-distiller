use distiller_core::error::Result;
use distiller_core::{
    error::DistilError,
    ir::{
        Class, Field, File, Function, Import, Interface, Modifier, Node, Parameter, TypeParam,
        TypeRef, Visibility,
    },
    options::ProcessOptions,
    parser::ParserPool,
    processor::language::LanguageProcessor,
};
use std::path::Path;
use std::sync::Arc;

pub struct GoProcessor {
    pool: Arc<ParserPool>,
}

impl GoProcessor {
    /// Create a new Go processor
    ///
    /// # Errors
    ///
    /// Returns an error if parsing or tree-sitter operations fail
    pub fn new() -> Result<Self> {
        Ok(Self {
            pool: Arc::new(ParserPool::default()),
        })
    }

    fn node_text(node: tree_sitter::Node, source: &str) -> String {
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
                        if list_child.kind() == "import_spec"
                            && let Some(import) = self.parse_import_spec(list_child, source)?
                        {
                            imports.push(import);
                        }
                    }
                }
                _ => {}
            }
        }

        Ok(imports)
    }

    #[allow(clippy::unused_self)]
    fn parse_import_spec(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        let mut module = String::new();
        let mut import_type = "import".to_string();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "interpreted_string_literal" => {
                    module = Self::node_text(child, source).trim_matches('"').to_string();
                }
                "package_identifier" => {
                    let alias = Self::node_text(child, source);
                    import_type = format!("import {alias} as");
                }
                "dot" => {
                    import_type = "import .".to_string();
                }
                _ => {}
            }
        }

        if module.is_empty() {
            Ok(None)
        } else {
            Ok(Some(Import {
                import_type,
                module,
                symbols: vec![],
                is_type: false,
                line: Some(node.start_position().row + 1),
            }))
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
                    name = Self::node_text(child, source);
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

    fn parse_struct_fields(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Field>> {
        let mut fields = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "field_declaration"
                && let Some(field) = self.parse_field_declaration(child, source)?
            {
                fields.push(field);
            }
        }

        Ok(fields)
    }

    #[allow(clippy::unused_self)]
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
                    name = Self::node_text(child, source);
                }
                "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
                | "slice_type" | "map_type" => {
                    if name.is_empty() {
                        // Embedded field: type name IS the field name
                        name = Self::node_text(child, source);
                        field_type = Some(TypeRef::new(name.clone()));
                    } else {
                        field_type = Some(TypeRef::new(Self::node_text(child, source)));
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
                    name = Self::node_text(child, source);
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "interface_type" => {
                    let mut interface_cursor = child.walk();
                    for interface_child in child.children(&mut interface_cursor) {
                        if interface_child.kind() == "method_elem"
                            && let Some(method) = self.parse_method_spec(interface_child, source)?
                        {
                            methods.push(method);
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
                    name = Self::node_text(child, source);
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
        let mut has_seen_parameters = false;

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "field_identifier" => {
                    name = Self::node_text(child, source);
                    has_seen_name = true;
                }
                "parameter_list" => {
                    // Go functions can have multiple parameter_list nodes:
                    // 1. Receiver (for methods) - before name
                    // 2. Parameters - after name
                    // 3. Return types - wrapped in parameter_list after parameters
                    if !has_seen_name {
                        // This is a receiver (for methods)
                        let receiver_params = self.parse_parameters(child, source)?;
                        if !receiver_params.is_empty() {
                            receiver_type = Some(receiver_params[0].param_type.clone());
                        }
                    } else if !has_seen_parameters {
                        // This is the actual parameter list
                        parameters = self.parse_parameters(child, source)?;
                        has_seen_parameters = true;
                    }
                    // Skip subsequent parameter_list nodes (return types)
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
            match child.kind() {
                "parameter_declaration" => {
                    parameters.extend(self.parse_parameter_declaration(child, source)?);
                }
                "variadic_parameter_declaration" => {
                    parameters.extend(self.parse_variadic_parameter(child, source)?);
                }
                _ => {}
            }
        }

        Ok(parameters)
    }

    #[allow(clippy::unused_self)]
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
                    names.push(Self::node_text(child, source));
                }
                "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
                | "slice_type" | "map_type" | "channel_type" | "function_type"
                | "interface_type" | "struct_type" => {
                    param_type = Some(TypeRef::new(Self::node_text(child, source)));
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
                param_type: param_type
                    .clone()
                    .unwrap_or_else(|| TypeRef::new(String::new())),
                default_value: None,
                is_variadic: false,
                is_optional: false,
                decorators: vec![],
            })
            .collect())
    }

    #[allow(clippy::unused_self)]
    #[allow(clippy::match_same_arms)]
    fn parse_variadic_parameter(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<Parameter>> {
        let mut name = String::new();
        let mut param_type = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "field_identifier" => {
                    name = Self::node_text(child, source);
                }
                "type_identifier" | "qualified_type" | "pointer_type" | "array_type"
                | "slice_type" | "map_type" | "channel_type" | "function_type"
                | "interface_type" | "struct_type" => {
                    param_type = Some(TypeRef::new(Self::node_text(child, source)));
                }
                "..." => {
                    // Variadic marker, already captured by node kind
                }
                _ => {}
            }
        }

        if name.is_empty() && param_type.is_some() {
            // Unnamed variadic parameter
            return Ok(vec![Parameter {
                name: String::new(),
                param_type: param_type.unwrap(),
                default_value: None,
                is_variadic: true,
                is_optional: false,
                decorators: vec![],
            }]);
        }

        Ok(vec![Parameter {
            name,
            param_type: param_type.unwrap_or_else(|| TypeRef::new(String::new())),
            default_value: None,
            is_variadic: true,
            is_optional: false,
            decorators: vec![],
        }])
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

    #[allow(clippy::unused_self)]
    fn try_extract_type(&self, node: tree_sitter::Node, source: &str) -> Option<TypeRef> {
        match node.kind() {
            "type_identifier" | "qualified_type" | "pointer_type" | "array_type" | "slice_type"
            | "map_type" | "channel_type" | "function_type" | "interface_type" | "struct_type" => {
                Some(TypeRef::new(Self::node_text(node, source)))
            }
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
            if child.kind() == "type_parameter_declaration"
                && let Some(tp) = self.parse_type_parameter(child, source)?
            {
                type_params.push(tp);
            }
        }

        Ok(type_params)
    }

    #[allow(clippy::unused_self)]
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
                        name = Self::node_text(child, source);
                    } else {
                        constraint = Some(TypeRef::new(Self::node_text(child, source)));
                    }
                }
                "qualified_type" | "interface_type" => {
                    constraint = Some(TypeRef::new(Self::node_text(child, source)));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(TypeParam {
            name,
            constraints: constraint.map(|c| vec![c]).unwrap_or_default(),
            default: None,
        }))
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

    fn process_type_spec(
        &self,
        node: tree_sitter::Node,
        source: &str,
        file: &mut File,
    ) -> Result<()> {
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
            .is_some_and(|ext| ext == "go")
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser_guard = self
            .pool
            .acquire("go", || Ok(tree_sitter_go::LANGUAGE.into()))?;
        let parser = parser_guard.get_mut();

        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(path.display().to_string(), "Failed to parse Go source")
        })?;

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

    // ===== Enhanced Test Coverage (11 new tests) =====

    #[test]
    fn test_empty_file() {
        let processor = GoProcessor::new().unwrap();
        let source = "package main\n";
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 0, "Empty file should have no children");
    }

    #[test]
    fn test_multiple_functions() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

func Add(x int, y int) int {
    return x + y
}

func Greet(name string, age int) string {
    return "Hello"
}

func Process(data []byte, count int) ([]byte, error) {
    return data, nil
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 3);

        // Validate Add function
        assert_eq!(functions[0].name, "Add");
        assert_eq!(functions[0].parameters.len(), 2);
        assert_eq!(functions[0].parameters[0].name, "x");
        assert_eq!(functions[0].parameters[0].param_type.name, "int");
        assert_eq!(functions[0].parameters[1].name, "y");
        assert_eq!(functions[0].parameters[1].param_type.name, "int");

        // Validate Greet function
        assert_eq!(functions[1].name, "Greet");
        assert_eq!(functions[1].parameters.len(), 2);
        assert_eq!(functions[1].parameters[0].name, "name");
        assert_eq!(functions[1].parameters[0].param_type.name, "string");
    }

    #[test]
    fn test_struct_methods() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type Counter struct {
    value int
}

func (c Counter) Get() int {
    return c.value
}

func (c *Counter) Increment() {
    c.value++
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

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
        assert_eq!(structs[0].name, "Counter");

        let methods: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(methods.len(), 2);
        assert_eq!(methods[0].name, "Get");
        assert_eq!(methods[1].name, "Increment");

        // Both methods should have Static modifier (receiver methods)
        assert!(methods[0].modifiers.contains(&Modifier::Static));
        assert!(methods[1].modifiers.contains(&Modifier::Static));
    }

    #[test]
    fn test_interface_definition() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type Writer interface {
    Write(data []byte) (int, error)
    Close() error
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

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
        assert_eq!(interfaces[0].name, "Writer");
        assert_eq!(interfaces[0].visibility, Visibility::Public);
        assert_eq!(interfaces[0].children.len(), 2);

        // Validate interface methods
        if let Node::Function(write_method) = &interfaces[0].children[0] {
            assert_eq!(write_method.name, "Write");
        } else {
            panic!("Expected function node for Write method");
        }

        if let Node::Function(close_method) = &interfaces[0].children[1] {
            assert_eq!(close_method.name, "Close");
        } else {
            panic!("Expected function node for Close method");
        }
    }

    #[test]
    fn test_embedded_struct() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type Base struct {
    ID int
}

type Extended struct {
    Base
    Name string
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

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

        assert_eq!(structs.len(), 2);

        // Validate Base struct
        assert_eq!(structs[0].name, "Base");
        assert_eq!(structs[0].children.len(), 1);

        // Validate Extended struct with embedded field
        assert_eq!(structs[1].name, "Extended");
        assert_eq!(structs[1].children.len(), 2);

        // Check for embedded Base field
        let fields: Vec<_> = structs[1]
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Field(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(fields.len(), 2);
        assert!(
            fields.iter().any(|f| f.name == "Base"),
            "Expected embedded Base field"
        );
        assert!(
            fields.iter().any(|f| f.name == "Name"),
            "Expected Name field"
        );
    }

    #[test]
    fn test_generic_function() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

func Max[T comparable](a, b T) T {
    if a > b {
        return a
    }
    return b
}

func Map[T any, R any](slice []T, fn func(T) R) []R {
    result := make([]R, len(slice))
    return result
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 2);

        // Validate Max function with type parameters
        assert_eq!(functions[0].name, "Max");
        assert!(
            !functions[0].type_params.is_empty(),
            "Expected type parameters for generic function"
        );

        // Validate Map function with multiple type parameters
        assert_eq!(functions[1].name, "Map");
        assert!(
            !functions[1].type_params.is_empty(),
            "Expected type parameters for generic function"
        );
    }

    #[test]
    fn test_package_level_variables() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

var GlobalCounter int = 0
var (
    AppName string = "MyApp"
    Version string = "1.0.0"
)
"#;
        let opts = ProcessOptions::default();

        // Note: Current implementation may not parse package-level vars as separate nodes
        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();
        assert!(file.path.contains("test.go"));

        // This test validates the processor doesn't crash on package-level variables
        // Full var parsing may be added in future enhancements
    }

    #[test]
    fn test_unexported_types() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type publicStruct struct {
    Public int
    private int
}

func publicFunc() {}
func privateFunc() {}

type internalInterface interface {
    method()
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        // Validate visibility detection
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
        assert_eq!(structs[0].name, "publicStruct");
        assert_eq!(
            structs[0].visibility,
            Visibility::Internal,
            "Lowercase struct should be internal/unexported"
        );

        // Check field visibility
        let fields: Vec<_> = structs[0]
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Field(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(
            fields[0].visibility,
            Visibility::Public,
            "Uppercase field should be public"
        );
        assert_eq!(
            fields[1].visibility,
            Visibility::Internal,
            "Lowercase field should be internal"
        );

        // Validate functions
        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 2);
        assert_eq!(
            functions[0].visibility,
            Visibility::Internal,
            "Lowercase function should be internal"
        );
        assert_eq!(
            functions[1].visibility,
            Visibility::Internal,
            "Lowercase function should be internal"
        );
    }

    #[test]
    fn test_multiple_return_values() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func GetUser(id int) (*User, bool, error) {
    return nil, false, nil
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 2);

        // Validate Divide function
        assert_eq!(functions[0].name, "Divide");
        assert_eq!(functions[0].parameters.len(), 2);

        // Validate GetUser function
        assert_eq!(functions[1].name, "GetUser");
        assert_eq!(functions[1].parameters.len(), 1);
    }

    #[test]
    fn test_pointer_receiver() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

type Buffer struct {
    data []byte
}

func (b *Buffer) Write(p []byte) (int, error) {
    b.data = append(b.data, p...)
    return len(p), nil
}

func (b *Buffer) Reset() {
    b.data = b.data[:0]
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        let methods: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(methods.len(), 2);

        // Both methods have pointer receivers
        assert_eq!(methods[0].name, "Write");
        assert!(
            methods[0].modifiers.contains(&Modifier::Static),
            "Pointer receiver method should have Static modifier"
        );

        assert_eq!(methods[1].name, "Reset");
        assert!(
            methods[1].modifiers.contains(&Modifier::Static),
            "Pointer receiver method should have Static modifier"
        );
    }

    #[test]
    fn test_variadic_parameters() {
        let processor = GoProcessor::new().unwrap();
        let source = r#"
package main

func Sum(numbers ...int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}

func Printf(format string, args ...interface{}) (int, error) {
    return 0, nil
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.go"), &opts)
            .unwrap();

        let functions: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(functions.len(), 2);

        // Validate Sum function with variadic parameter
        assert_eq!(functions[0].name, "Sum");
        assert!(
            !functions[0].parameters.is_empty(),
            "Expected at least one parameter"
        );

        // Validate Printf function with format string and variadic args
        assert_eq!(functions[1].name, "Printf");
        assert!(
            !functions[1].parameters.is_empty(),
            "Expected at least format parameter"
        );
    }
}

#[cfg(test)]
mod debug_tests {
    use super::*;
    use tree_sitter::Node as TSNode;

    fn print_tree(node: TSNode, source: &str, depth: usize) {
        let indent = "  ".repeat(depth);
        let kind = node.kind();

        let start = node.start_byte();
        let end = node.end_byte();
        let text = if end > start && end <= source.len() {
            &source[start..end]
        } else {
            ""
        };

        let text_preview = if text.len() > 60 {
            format!("{}...", &text[..60].replace('\n', "\\n"))
        } else {
            text.replace('\n', "\\n")
        };

        eprintln!("{}[{}] \"{}\"", indent, kind, text_preview);

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            print_tree(child, source, depth + 1);
        }
    }

    #[test]
    #[ignore]
    fn debug_multiple_return_values() {
        let source = r#"package main

func GetUser(id int) (*User, bool, error) {
    return nil, false, nil
}"#;

        let processor = GoProcessor::new().unwrap();
        let mut parser_guard = processor
            .pool
            .acquire("go", || Ok(tree_sitter_go::LANGUAGE.into()))
            .unwrap();
        let parser = parser_guard.get_mut();
        let tree = parser.parse(source, None).unwrap();
        let root = tree.root_node();

        eprintln!("\n=== Go Multiple Return Values AST ===\n");
        print_tree(root, source, 0);

        panic!("Debug output - check stderr");
    }

    #[test]
    #[ignore]
    fn debug_variadic_parameters() {
        let source = r#"package main

func Sum(numbers ...int) int {
    return 0
}"#;

        let processor = GoProcessor::new().unwrap();
        let mut parser_guard = processor
            .pool
            .acquire("go", || Ok(tree_sitter_go::LANGUAGE.into()))
            .unwrap();
        let parser = parser_guard.get_mut();
        let tree = parser.parse(source, None).unwrap();
        let root = tree.root_node();

        eprintln!("\n=== Go Variadic Parameters AST ===\n");
        print_tree(root, source, 0);

        panic!("Debug output - check stderr");
    }

    // ===== EDGE CASE TESTS (Phase C3) =====

    #[test]
    fn test_malformed_go() {
        let source =
            std::fs::read_to_string("../../testdata/edge-cases/malformed/go_syntax_error.go")
                .expect("Failed to read malformed Go file");

        let processor = GoProcessor::new().unwrap();
        let opts = ProcessOptions::default();

        // Should not panic - tree-sitter handles malformed code
        let result = processor.process(&source, Path::new("error.go"), &opts);

        match result {
            Ok(file) => {
                println!("✓ Malformed Go: Partial parse successful");
                println!("  Found {} top-level nodes", file.children.len());
                // Tree-sitter should recover and parse valid nodes
                assert!(
                    !file.children.is_empty(),
                    "Should find at least some valid nodes"
                );
            }
            Err(e) => {
                println!("✓ Malformed Go: Error handled gracefully: {}", e);
                // As long as it doesn't panic, we're good
            }
        }
    }

    #[test]
    fn test_unicode_go() {
        let source = std::fs::read_to_string("../../testdata/edge-cases/unicode/go_unicode.go")
            .expect("Failed to read Unicode Go file");

        let processor = GoProcessor::new().unwrap();
        let opts = ProcessOptions::default();

        let result = processor.process(&source, Path::new("unicode.go"), &opts);

        assert!(result.is_ok(), "Unicode Go file should parse successfully");

        let file = result.unwrap();
        let struct_count = file
            .children
            .iter()
            .filter(|n| matches!(n, Node::Class(_)))
            .count();

        println!(
            "✓ Unicode Go: {} structs with Unicode identifiers",
            struct_count
        );

        // Should find structs with Unicode names
        assert!(
            struct_count >= 5,
            "Should find at least 5 structs with Unicode names"
        );
    }

    #[test]
    fn test_large_go_file() {
        let source = std::fs::read_to_string("../../testdata/edge-cases/large-files/large_go.go")
            .expect("Failed to read large Go file");

        let processor = GoProcessor::new().unwrap();
        let opts = ProcessOptions::default();

        println!("Testing large Go file: {} lines", source.lines().count());

        let start = std::time::Instant::now();
        let result = processor.process(&source, Path::new("large.go"), &opts);
        let duration = start.elapsed();

        assert!(result.is_ok(), "Large Go file should parse successfully");

        let file = result.unwrap();
        let struct_count = file
            .children
            .iter()
            .filter(|n| matches!(n, Node::Class(_)))
            .count();

        println!(
            "✓ Large Go: {} structs parsed in {:?}",
            struct_count, duration
        );
        println!(
            "  Performance: ~{} lines/ms",
            source.lines().count() / duration.as_millis().max(1) as usize
        );

        // Performance target: should parse in reasonable time (< 1 second for 17k lines)
        assert!(
            duration.as_secs() < 1,
            "Large file parsing took too long: {:?}",
            duration
        );
    }
}
