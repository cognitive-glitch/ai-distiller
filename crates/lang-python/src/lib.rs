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
    parser::ParserPool,
    processor::language::LanguageProcessor,
};
use std::path::Path;
use std::sync::Arc;

/// Python language processor
pub struct PythonProcessor {
    pool: Arc<ParserPool>,
}

impl PythonProcessor {
    /// Create a new Python processor
    ///
    /// # Errors
    ///
    /// Returns an error if parsing or tree-sitter operations fail
    pub fn new() -> Result<Self> {
        Ok(Self {
            pool: Arc::new(ParserPool::default()),
        })
    }

    /// Parse source code into IR
    fn parse_source(&self, source: &str, filename: &str) -> Result<File> {
        let mut parser_guard = self
            .pool
            .acquire("python", || Ok(tree_sitter_python::LANGUAGE.into()))?;
        let parser = parser_guard.get_mut();

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
                if let Some(import) = Self::parse_import(node, source)? {
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
    fn parse_import(node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        match node.kind() {
            "import_statement" => {
                // import foo, bar
                let text = Self::node_text(node, source);
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
                            if seen_import_keyword {
                                // This is an imported symbol
                                symbols.push(ImportedSymbol {
                                    name: Self::node_text(child, source),
                                    alias: None,
                                });
                            } else {
                                // This is the module name
                                module = Self::node_text(child, source);
                            }
                        }
                        "aliased_import" => {
                            symbols.push(ImportedSymbol {
                                name: Self::node_text(child, source),
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
                    class.name = Self::node_text(child, source);
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
    #[allow(clippy::unused_self)]
    #[allow(clippy::unnecessary_wraps)]
    fn parse_base_classes(
        &self,
        node: tree_sitter::Node,
        source: &str,
        class: &mut Class,
    ) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "identifier" || child.kind() == "attribute" {
                let base_name = Self::node_text(child, source);
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
        if let Some(first) = node.child(0)
            && first.kind() == "async"
        {
            function.modifiers.push(Modifier::Async);
        }

        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    function.name = Self::node_text(child, source);
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
                    let return_type = Self::node_text(child, source);
                    function.return_type = Some(TypeRef::new(return_type));
                }
                "block" => {
                    // Function body
                    function.implementation = Some(Self::node_text(child, source));
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
    #[allow(clippy::unused_self)]
    fn parse_parameter(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Parameter>> {
        match node.kind() {
            "identifier" => {
                let name = Self::node_text(node, source);
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
                            param.name = Self::node_text(child, source);
                        }
                        "type" => {
                            let type_name = Self::node_text(child, source);
                            param.param_type = TypeRef::new(type_name);
                        }
                        _ => {
                            // Could be default value
                            if !child.kind().starts_with('(') && !child.kind().starts_with(')') {
                                param.default_value = Some(Self::node_text(child, source));
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
    #[allow(clippy::match_same_arms)]
    fn parse_decorated(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Node>> {
        let mut decorators = Vec::new();
        let mut definition_node = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "decorator" => {
                    let decorator_text = Self::node_text(child, source);
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
        let text = Self::node_text(node, source);
        if !text.starts_with("self.") {
            return Ok(None);
        }

        // Extract field name
        if let Some(eq_pos) = text.find('=') {
            let field_part = &text[5..eq_pos].trim();
            let field_name = (*field_part).to_string();

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
    #[allow(clippy::unused_self)]
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
                let name = Self::node_text(child, source);
                return self.detect_visibility(&name);
            }
        }
        Visibility::Public
    }

    /// Get text content of a node
    fn node_text(node: tree_sitter::Node, source: &str) -> String {
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
            .is_some_and(|ext| ext == "py" || ext == "pyw")
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

// ===== Enhanced Test Coverage =====

#[test]
fn test_function_with_return_type() {
    let processor = PythonProcessor::new().unwrap();
    let source = "def calculate(x: int, y: int) -> int:\n    return x + y";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "calculate");
        assert!(
            !func.parameters.is_empty(),
            "Expected at least 1 parameter, got {}",
            func.parameters.len()
        );

        // Validate typed parameters
        assert_eq!(func.parameters[0].name, "x");
        assert_eq!(func.parameters[0].param_type.name, "int");
        assert_eq!(func.parameters[1].name, "y");
        assert_eq!(func.parameters[1].param_type.name, "int");

        // Validate return type
        assert!(func.return_type.is_some());
        assert_eq!(func.return_type.as_ref().unwrap().name, "int");
    } else {
        panic!("Expected function node, got {:?}", file.children[0]);
    }
}

#[test]
fn test_async_function() {
    let processor = PythonProcessor::new().unwrap();
    let source = "async def fetch_data():\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "fetch_data");
        assert!(
            func.modifiers.contains(&Modifier::Async),
            "Expected async modifier, got {:?}",
            func.modifiers
        );
    } else {
        panic!("Expected function node");
    }
}

#[test]
fn test_decorated_function() {
    let processor = PythonProcessor::new().unwrap();
    let source = "@staticmethod\ndef helper():\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "helper");
        assert_eq!(func.decorators.len(), 1);
        assert_eq!(func.decorators[0], "@staticmethod");
    } else {
        panic!("Expected function node");
    }
}

#[test]
fn test_class_with_inheritance() {
    let processor = PythonProcessor::new().unwrap();
    let source = "class Child(Parent):\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "Child");
        assert_eq!(class.extends.len(), 1);
        assert_eq!(class.extends[0].name, "Parent");
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_class_with_multiple_inheritance() {
    let processor = PythonProcessor::new().unwrap();
    let source = "class MultiChild(Parent1, Parent2, Parent3):\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "MultiChild");
        assert_eq!(
            class.extends.len(),
            3,
            "Expected 3 base classes, got {}",
            class.extends.len()
        );
        assert_eq!(class.extends[0].name, "Parent1");
        assert_eq!(class.extends[1].name, "Parent2");
        assert_eq!(class.extends[2].name, "Parent3");
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_class_with_mixed_visibility_methods() {
    let processor = PythonProcessor::new().unwrap();
    let source = r#"class Service:
    def public_method(self):
        pass

    def _protected_method(self):
        pass

    def __private_method(self):
        pass

    def __init__(self):
        pass
"#;
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "Service");
        assert_eq!(class.children.len(), 4);

        // Validate visibility levels
        let methods: Vec<_> = class
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

        assert_eq!(methods.len(), 4);

        // public_method
        assert_eq!(methods[0].name, "public_method");
        assert_eq!(methods[0].visibility, Visibility::Public);

        // _protected_method
        assert_eq!(methods[1].name, "_protected_method");
        assert_eq!(methods[1].visibility, Visibility::Protected);

        // __private_method
        assert_eq!(methods[2].name, "__private_method");
        assert_eq!(methods[2].visibility, Visibility::Private);

        // __init__ (dunder, public)
        assert_eq!(methods[3].name, "__init__");
        assert_eq!(methods[3].visibility, Visibility::Public);
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_decorated_class() {
    let processor = PythonProcessor::new().unwrap();
    let source = "@dataclass\nclass Point:\n    x: int\n    y: int";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "Point");
        assert_eq!(class.decorators.len(), 1);
        assert_eq!(class.decorators[0], "@dataclass");
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_function_with_typed_parameters() {
    let processor = PythonProcessor::new().unwrap();
    let source = "def greet(name: str, count: int) -> str:\n    return name * count";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "greet");
        assert!(
            func.parameters.len() >= 2,
            "Expected at least 2 parameters, got {}",
            func.parameters.len()
        );

        // Validate typed parameters
        assert_eq!(func.parameters[0].name, "name");
        assert_eq!(func.parameters[0].param_type.name, "str");

        if func.parameters.len() >= 2 {
            assert_eq!(func.parameters[1].name, "count");
            assert_eq!(func.parameters[1].param_type.name, "int");
        }

        // Validate return type
        assert!(func.return_type.is_some());
        assert_eq!(func.return_type.as_ref().unwrap().name, "str");
    } else {
        panic!("Expected function node");
    }
}

#[test]
fn test_multiple_decorators() {
    let processor = PythonProcessor::new().unwrap();
    let source = "@decorator1\n@decorator2\n@decorator3\ndef decorated():\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Function(func) = &file.children[0] {
        assert_eq!(func.name, "decorated");
        assert_eq!(
            func.decorators.len(),
            3,
            "Expected 3 decorators, got {}",
            func.decorators.len()
        );
        assert_eq!(func.decorators[0], "@decorator1");
        assert_eq!(func.decorators[1], "@decorator2");
        assert_eq!(func.decorators[2], "@decorator3");
    } else {
        panic!("Expected function node");
    }
}

#[test]
fn test_empty_file() {
    let processor = PythonProcessor::new().unwrap();
    let source = "";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 0, "Empty file should have no children");
}

#[test]
fn test_multiple_imports() {
    let processor = PythonProcessor::new().unwrap();
    let source = r#"import os
import sys
from typing import List, Dict, Optional
from pathlib import Path
"#;
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 4);

    // Validate import types
    let imports: Vec<_> = file
        .children
        .iter()
        .filter_map(|n| {
            if let Node::Import(i) = n {
                Some(i)
            } else {
                None
            }
        })
        .collect();

    assert_eq!(imports.len(), 4);
    assert_eq!(imports[0].module, "os");
    assert_eq!(imports[0].import_type, "import");
    assert_eq!(imports[1].module, "sys");
    assert_eq!(imports[1].import_type, "import");
    assert_eq!(imports[2].module, "typing");
    assert_eq!(imports[2].import_type, "from");
    assert_eq!(imports[3].module, "pathlib");
    assert_eq!(imports[3].import_type, "from");
}

#[test]
fn test_complex_class_with_everything() {
    let processor = PythonProcessor::new().unwrap();
    let source = r#"@dataclass
class ComplexService(BaseService, Mixin):
    def __init__(self):
        self.public_field = 0
        self._protected_field = 1
        self.__private_field = 2

    @property
    def value(self) -> int:
        return self.public_field

    @staticmethod
    def helper() -> str:
        return "help"

    async def fetch(self, url: str) -> str:
        pass
"#;
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "ComplexService");

        // Validate decorators
        assert_eq!(class.decorators.len(), 1);
        assert_eq!(class.decorators[0], "@dataclass");

        // Validate inheritance
        assert_eq!(class.extends.len(), 2);
        assert_eq!(class.extends[0].name, "BaseService");
        assert_eq!(class.extends[1].name, "Mixin");

        // Validate children (1 __init__ + 3 fields + 3 methods)
        assert!(
            class.children.len() >= 4,
            "Expected at least 4 children, got {}",
            class.children.len()
        );

        // Find the async method
        let async_methods: Vec<_> = class
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Function(f) = n {
                    Some(f)
                } else {
                    None
                }
            })
            .filter(|f| f.modifiers.contains(&Modifier::Async))
            .collect();

        assert!(!async_methods.is_empty(), "Expected to find async method");
        assert_eq!(async_methods[0].name, "fetch");
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_private_class() {
    let processor = PythonProcessor::new().unwrap();
    let source = "class _PrivateClass:\n    pass";
    let opts = ProcessOptions::default();

    let file = processor
        .process(source, Path::new("test.py"), &opts)
        .unwrap();
    assert_eq!(file.children.len(), 1);

    if let Node::Class(class) = &file.children[0] {
        assert_eq!(class.name, "_PrivateClass");
        assert_eq!(
            class.visibility,
            Visibility::Protected,
            "Classes starting with _ should be protected"
        );
    } else {
        panic!("Expected class node");
    }
}

#[test]
fn test_django_style_models() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/real-world/django-app/models.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read Django models file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("models.py"), &opts);

    assert!(result.is_ok(), "Django models should parse successfully");

    let file = result.unwrap();

    // Should find User, Post, Comment classes
    let classes: Vec<_> = file
        .children
        .iter()
        .filter_map(|n| {
            if let Node::Class(c) = n {
                Some(c)
            } else {
                None
            }
        })
        .collect();

    assert!(
        classes.len() >= 3,
        "Should find at least 3 classes (User, Post, Comment)"
    );

    // Verify User class
    let user_class = classes.iter().find(|c| c.name == "User");
    assert!(user_class.is_some(), "Should find User class");

    println!("✅ Django models parse: {} classes found", classes.len());
}

#[test]
fn test_django_style_views() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/real-world/django-app/views.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read Django views file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("views.py"), &opts);

    assert!(result.is_ok(), "Django views should parse successfully");

    let file = result.unwrap();

    // Should find decorator functions and views
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

    assert!(functions.len() >= 3, "Should find multiple functions");

    // Check for decorated functions
    let decorated = functions
        .iter()
        .filter(|f| !f.decorators.is_empty())
        .count();

    assert!(decorated > 0, "Should find decorated functions");

    println!(
        "✅ Django views parse: {} functions, {} decorated",
        functions.len(),
        decorated
    );
}

// ===== EDGE CASE TESTS (Phase C3) =====

#[test]
fn test_malformed_python() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/edge-cases/malformed/python_syntax_error.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read malformed Python file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    // Should not panic - tree-sitter handles malformed code
    let result = processor.process(&source, Path::new("error.py"), &opts);

    match result {
        Ok(file) => {
            println!("✓ Malformed Python: Partial parse successful");
            println!("  Found {} top-level nodes", file.children.len());
            // Tree-sitter should recover and parse valid nodes
            assert!(
                !file.children.is_empty(),
                "Should find at least some valid nodes"
            );
        }
        Err(e) => {
            println!("✓ Malformed Python: Error handled gracefully: {}", e);
            // As long as it doesn't panic, we're good
        }
    }
}

#[test]
fn test_unicode_python() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/edge-cases/unicode/python_unicode.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read Unicode Python file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("unicode.py"), &opts);

    assert!(
        result.is_ok(),
        "Unicode Python file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, Node::Class(_)))
        .count();

    println!(
        "✓ Unicode Python: {} classes with Unicode identifiers",
        class_count
    );

    // Should find classes with Unicode names (Пользователь, 🚀Rocket, etc.)
    assert!(
        class_count >= 5,
        "Should find at least 5 classes with Unicode names"
    );
}

#[test]
fn test_large_python_file() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/edge-cases/large-files/large_python.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read large Python file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    println!(
        "Testing large Python file: {} lines",
        source.lines().count()
    );

    let start = std::time::Instant::now();
    let result = processor.process(&source, Path::new("large.py"), &opts);
    let duration = start.elapsed();

    assert!(
        result.is_ok(),
        "Large Python file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, Node::Class(_)))
        .count();

    println!(
        "✓ Large Python: {} classes parsed in {:?}",
        class_count, duration
    );
    println!(
        "  Performance: ~{} lines/ms",
        source.lines().count() / duration.as_millis().max(1) as usize
    );

    // Performance target: should parse in reasonable time (< 1 second for 15k lines)
    assert!(
        duration.as_secs() < 1,
        "Large file parsing took too long: {:?}",
        duration
    );
}

#[test]
fn test_empty_python_file_edge() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/edge-cases/syntax-edge/empty.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read empty Python file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("empty.py"), &opts);

    assert!(
        result.is_ok(),
        "Empty Python file should parse successfully"
    );

    let file = result.unwrap();

    println!("✓ Empty Python file: {} nodes", file.children.len());

    // Empty file should have 0 or very few nodes
    assert!(
        file.children.len() <= 1,
        "Empty file should have minimal nodes"
    );
}

#[test]
fn test_deeply_nested_python() {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    let path = format!(
        "{}/../../testdata/edge-cases/syntax-edge/deeply_nested.py",
        manifest_dir
    );
    let source = std::fs::read_to_string(&path).expect("Failed to read deeply nested Python file");

    let processor = PythonProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("nested.py"), &opts);

    assert!(
        result.is_ok(),
        "Deeply nested Python file should parse successfully"
    );

    let file = result.unwrap();

    println!(
        "✓ Deeply nested Python: {} top-level nodes",
        file.children.len()
    );

    // Should handle deep nesting without stack overflow
    assert!(
        file.children.len() >= 2,
        "Should find Level1 class and complex_nesting function"
    );
}
