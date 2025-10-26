//! Python language processor using tree-sitter
//!
//! Parses Python source code into IR nodes, handling:
//! - Classes and methods
//! - Functions and decorators
//! - Import statements
//! - Field assignments
//! - Visibility detection (_private, __dunder__)

use distiller_core::{
    error::{DistilError, Result},
    ir::{
        Class, Field, File, Function, Import, ImportedSymbol, Modifier, Node, Parameter, TypeRef,
        Visibility,
    },
    options::ProcessOptions,
    processor::language::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use tree_sitter::Parser;

/// Python language processor
pub struct PythonProcessor {
    parser: Mutex<Parser>,
}

impl PythonProcessor {
    /// Create a new Python processor
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_python::LANGUAGE.into())
            .map_err(|e| {
                DistilError::TreeSitter(format!("Failed to set Python language: {}", e))
            })?;

        Ok(Self {
            parser: Mutex::new(parser),
        })
    }

    /// Parse source code into IR
    fn parse_source(&self, source: &str, filename: &str) -> Result<File> {
        let mut parser = self.parser.lock();

        let tree = parser
            .parse(source, None)
            .ok_or_else(|| DistilError::parse_error(filename, "Failed to parse Python source"))?;

        let root = tree.root_node();

        let mut file = File {
            path: filename.to_string(),
            children: Vec::new(),
        };

        // Process all top-level nodes
        self.process_node(root, source, &mut file)?;

        Ok(file)
    }

    /// Process a tree-sitter node recursively
    fn process_node(&self, node: tree_sitter::Node, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "module" => {
                // Process all children of module
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.process_node(child, source, file)?;
                }
            }
            "import_statement" | "import_from_statement" => {
                if let Some(import) = self.parse_import(node, source)? {
                    file.children.push(Node::Import(import));
                }
            }
            "class_definition" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "function_definition" => {
                if let Some(function) = self.parse_function(node, source, Visibility::Public)? {
                    file.children.push(Node::Function(function));
                }
            }
            "decorated_definition" => {
                // Handle @decorator syntax
                if let Some(decorated_node) = self.parse_decorated(node, source)? {
                    file.children.push(decorated_node);
                }
            }
            _ => {
                // Recurse into other nodes
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.process_node(child, source, file)?;
                }
            }
        }

        Ok(())
    }

    /// Parse an import statement
    fn parse_import(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        match node.kind() {
            "import_statement" => {
                // import foo, bar
                let text = self.node_text(node, source);
                let module = text
                    .strip_prefix("import ")
                    .unwrap_or(&text)
                    .trim()
                    .to_string();

                Ok(Some(Import {
                    import_type: "import".to_string(),
                    module,
                    symbols: Vec::new(),
                    is_type: false,
                    line: Some(node.start_position().row + 1),
                }))
            }
            "import_from_statement" => {
                // from foo import bar, baz
                let mut module = String::new();
                let mut symbols = Vec::new();
                let mut seen_import_keyword = false;

                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    match child.kind() {
                        "import" => {
                            seen_import_keyword = true;
                        }
                        "dotted_name" => {
                            if !seen_import_keyword {
                                // This is the module name
                                module = self.node_text(child, source);
                            } else {
                                // This is an imported symbol
                                symbols.push(ImportedSymbol {
                                    name: self.node_text(child, source),
                                    alias: None,
                                });
                            }
                        }
                        "aliased_import" => {
                            symbols.push(ImportedSymbol {
                                name: self.node_text(child, source),
                                alias: None,
                            });
                        }
                        _ => {}
                    }
                }

                Ok(Some(Import {
                    import_type: "from".to_string(),
                    module,
                    symbols,
                    is_type: false,
                    line: Some(node.start_position().row + 1),
                }))
            }
            _ => Ok(None),
        }
    }

    /// Parse a class definition
    fn parse_class(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Class>> {
        let mut class = Class {
            name: String::new(),
            visibility: Visibility::Public,
            modifiers: Vec::new(),
            decorators: Vec::new(),
            type_params: Vec::new(),
            extends: Vec::new(),
            implements: Vec::new(),
            children: Vec::new(),
            line_start: node.start_position().row + 1,
            line_end: node.end_position().row + 1,
        };

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    class.name = self.node_text(child, source);
                    // Detect visibility from name
                    class.visibility = self.detect_visibility(&class.name);
                }
                "argument_list" => {
                    // Parse base classes
                    self.parse_base_classes(child, source, &mut class)?;
                }
                "block" => {
                    // Parse class body
                    self.parse_class_body(child, source, &mut class)?;
                }
                _ => {}
            }
        }

        if class.name.is_empty() {
            return Ok(None);
        }

        Ok(Some(class))
    }

    /// Parse base classes from argument list
    fn parse_base_classes(
        &self,
        node: tree_sitter::Node,
        source: &str,
        class: &mut Class,
    ) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "identifier" || child.kind() == "attribute" {
                let base_name = self.node_text(child, source);
                class.extends.push(TypeRef::new(base_name));
            }
        }
        Ok(())
    }

    /// Parse class body (methods and fields)
    fn parse_class_body(
        &self,
        node: tree_sitter::Node,
        source: &str,
        class: &mut Class,
    ) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "function_definition" => {
                    // Parse method
                    let visibility = self.detect_visibility_from_node(child, source);
                    if let Some(function) = self.parse_function(child, source, visibility)? {
                        class.children.push(Node::Function(function));
                    }
                }
                "decorated_definition" => {
                    // Handle @decorator syntax on methods
                    if let Some(decorated_node) = self.parse_decorated(child, source)? {
                        class.children.push(decorated_node);
                    }
                }
                "expression_statement" => {
                    // Parse field assignments (self.field = value)
                    if let Some(field) = self.parse_field_assignment(child, source)? {
                        class.children.push(Node::Field(field));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }

    /// Parse a function/method definition
    fn parse_function(
        &self,
        node: tree_sitter::Node,
        source: &str,
        visibility: Visibility,
    ) -> Result<Option<Function>> {
        let mut function = Function {
            name: String::new(),
            visibility,
            modifiers: Vec::new(),
            decorators: Vec::new(),
            type_params: Vec::new(),
            parameters: Vec::new(),
            return_type: None,
            implementation: None,
            line_start: node.start_position().row + 1,
            line_end: node.end_position().row + 1,
        };

        // Check for async modifier
        let mut cursor = node.walk();
        if let Some(first) = node.child(0) {
            if first.kind() == "async" {
                function.modifiers.push(Modifier::Async);
            }
        }

        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    function.name = self.node_text(child, source);
                    // Override visibility if needed
                    if function.visibility == Visibility::Public {
                        function.visibility = self.detect_visibility(&function.name);
                    }
                }
                "parameters" => {
                    self.parse_parameters(child, source, &mut function)?;
                }
                "type" => {
                    // Return type annotation
                    let return_type = self.node_text(child, source);
                    function.return_type = Some(TypeRef::new(return_type));
                }
                "block" => {
                    // Function body
                    function.implementation = Some(self.node_text(child, source));
                }
                _ => {}
            }
        }

        if function.name.is_empty() {
            return Ok(None);
        }

        Ok(Some(function))
    }

    /// Parse function parameters
    fn parse_parameters(
        &self,
        node: tree_sitter::Node,
        source: &str,
        function: &mut Function,
    ) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if let Some(param) = self.parse_parameter(child, source)? {
                function.parameters.push(param);
            }
        }
        Ok(())
    }

    /// Parse a single parameter
    fn parse_parameter(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Parameter>> {
        match node.kind() {
            "identifier" => {
                let name = self.node_text(node, source);
                // Skip 'self' and 'cls' special parameters
                if name == "self" || name == "cls" {
                    return Ok(None);
                }

                Ok(Some(Parameter {
                    name,
                    param_type: TypeRef::new("Any"),
                    default_value: None,
                    is_variadic: false,
                    is_optional: false,
                    decorators: Vec::new(),
                }))
            }
            "typed_parameter" | "default_parameter" => {
                let mut param = Parameter {
                    name: String::new(),
                    param_type: TypeRef::new("Any"),
                    default_value: None,
                    is_variadic: false,
                    is_optional: false,
                    decorators: Vec::new(),
                };

                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    match child.kind() {
                        "identifier" => {
                            param.name = self.node_text(child, source);
                        }
                        "type" => {
                            let type_name = self.node_text(child, source);
                            param.param_type = TypeRef::new(type_name);
                        }
                        _ => {
                            // Could be default value
                            if !child.kind().starts_with('(') && !child.kind().starts_with(')') {
                                param.default_value = Some(self.node_text(child, source));
                            }
                        }
                    }
                }

                if param.name.is_empty() || param.name == "self" || param.name == "cls" {
                    Ok(None)
                } else {
                    Ok(Some(param))
                }
            }
            _ => Ok(None),
        }
    }

    /// Parse decorated definition (class or function with decorators)
    fn parse_decorated(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Node>> {
        let mut decorators = Vec::new();
        let mut definition_node = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "decorator" => {
                    let decorator_text = self.node_text(child, source);
                    decorators.push(decorator_text);
                }
                "class_definition" => {
                    definition_node = Some(child);
                }
                "function_definition" => {
                    definition_node = Some(child);
                }
                _ => {}
            }
        }

        if let Some(def_node) = definition_node {
            match def_node.kind() {
                "class_definition" => {
                    if let Some(mut class) = self.parse_class(def_node, source)? {
                        class.decorators = decorators;
                        return Ok(Some(Node::Class(class)));
                    }
                }
                "function_definition" => {
                    let visibility = self.detect_visibility_from_node(def_node, source);
                    if let Some(mut function) = self.parse_function(def_node, source, visibility)? {
                        function.decorators = decorators;
                        return Ok(Some(Node::Function(function)));
                    }
                }
                _ => {}
            }
        }

        Ok(None)
    }

    /// Parse field assignment (self.field = value)
    fn parse_field_assignment(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<Field>> {
        // Look for patterns like: self.field_name = value
        let text = self.node_text(node, source);
        if !text.starts_with("self.") {
            return Ok(None);
        }

        // Extract field name
        if let Some(eq_pos) = text.find('=') {
            let field_part = &text[5..eq_pos].trim();
            let field_name = field_part.to_string();

            Ok(Some(Field {
                name: field_name.clone(),
                visibility: self.detect_visibility(&field_name),
                modifiers: Vec::new(),
                field_type: None,
                default_value: None,
                line: node.start_position().row + 1,
            }))
        } else {
            Ok(None)
        }
    }

    /// Detect visibility from name conventions
    fn detect_visibility(&self, name: &str) -> Visibility {
        if name.starts_with("__") && name.ends_with("__") {
            // __dunder__ methods are public API
            Visibility::Public
        } else if name.starts_with("__") {
            // __private (name mangling)
            Visibility::Private
        } else if name.starts_with('_') {
            // _protected
            Visibility::Protected
        } else {
            Visibility::Public
        }
    }

    /// Detect visibility from function node
    fn detect_visibility_from_node(&self, node: tree_sitter::Node, source: &str) -> Visibility {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "identifier" {
                let name = self.node_text(child, source);
                return self.detect_visibility(&name);
            }
        }
        Visibility::Public
    }

    /// Get text content of a node
    fn node_text(&self, node: tree_sitter::Node, source: &str) -> String {
        let start = node.start_byte();
        let end = node.end_byte();
        if start > end || end > source.len() {
            return String::new();
        }
        source[start..end].to_string()
    }
}

impl Default for PythonProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create default PythonProcessor")
    }
}

impl LanguageProcessor for PythonProcessor {
    fn language(&self) -> &'static str {
        "python"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["py", "pyw"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| ext == "py" || ext == "pyw")
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let filename = path.to_string_lossy().into_owned();
        self.parse_source(source, &filename)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let processor = PythonProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_supported_extensions() {
        let processor = PythonProcessor::new().unwrap();
        assert_eq!(processor.supported_extensions(), &["py", "pyw"]);
        assert!(processor.can_process(Path::new("test.py")));
        assert!(processor.can_process(Path::new("test.pyw")));
        assert!(!processor.can_process(Path::new("test.js")));
    }

    #[test]
    fn test_visibility_detection() {
        let processor = PythonProcessor::new().unwrap();

        assert_eq!(
            processor.detect_visibility("public_method"),
            Visibility::Public
        );
        assert_eq!(
            processor.detect_visibility("_protected_method"),
            Visibility::Protected
        );
        assert_eq!(
            processor.detect_visibility("__private_method"),
            Visibility::Private
        );
        assert_eq!(processor.detect_visibility("__init__"), Visibility::Public);
        assert_eq!(processor.detect_visibility("__str__"), Visibility::Public);
    }

    #[test]
    fn test_simple_function() {
        let processor = PythonProcessor::new().unwrap();
        let source = "def hello():\n    pass";
        let options = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.py"), &options)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Function(func) = &file.children[0] {
            assert_eq!(func.name, "hello");
            assert_eq!(func.visibility, Visibility::Public);
        } else {
            panic!("Expected function node");
        }
    }

    #[test]
    fn test_simple_class() {
        let processor = PythonProcessor::new().unwrap();
        let source = "class MyClass:\n    pass";
        let options = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.py"), &options)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "MyClass");
            assert_eq!(class.visibility, Visibility::Public);
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_import_statements() {
        let processor = PythonProcessor::new().unwrap();
        let source = "import os\nfrom typing import List";
        let options = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.py"), &options)
            .unwrap();
        assert_eq!(file.children.len(), 2);

        if let Node::Import(import) = &file.children[0] {
            assert_eq!(import.module, "os");
            assert_eq!(import.import_type, "import");
        } else {
            panic!("Expected import node");
        }

        if let Node::Import(import) = &file.children[1] {
            assert_eq!(import.module, "typing");
            assert_eq!(import.import_type, "from");
        } else {
            panic!("Expected import node");
        }
    }
}
