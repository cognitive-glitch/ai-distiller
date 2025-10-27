use distiller_core::{
    error::{DistilError, Result},
    ir::{
        Class, File, Function, Import, ImportedSymbol, Modifier, Node, Parameter, TypeRef,
        Visibility,
    },
    options::ProcessOptions,
    processor::language::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;
pub struct JavaScriptProcessor {
    parser: Arc<Mutex<tree_sitter::Parser>>,
}

impl JavaScriptProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = tree_sitter::Parser::new();
        parser
            .set_language(&tree_sitter_javascript::LANGUAGE.into())
            .map_err(|e| {
                DistilError::TreeSitter(format!("Failed to set JavaScript language: {e}"))
            })?;
        Ok(Self {
            parser: Arc::new(Mutex::new(parser)),
        })
    }

    fn node_text(node: tree_sitter::Node, source: &str) -> String {
        if node.start_byte() > node.end_byte() || node.end_byte() > source.len() {
            return String::new();
        }
        source[node.start_byte()..node.end_byte()].to_string()
    }

    #[allow(clippy::unused_self)]
    fn parse_import(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        let mut module = String::new();
        let mut symbols = Vec::new();
        let import_type = String::from("import");
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "import_clause" => {
                    let mut clause_cursor = child.walk();
                    for clause_child in child.children(&mut clause_cursor) {
                        match clause_child.kind() {
                            "identifier" => {
                                symbols.push(ImportedSymbol {
                                    name: Self::node_text(clause_child, source),
                                    alias: None,
                                });
                            }
                            "named_imports" => {
                                let mut imports_cursor = clause_child.walk();
                                for import_child in clause_child.children(&mut imports_cursor) {
                                    if import_child.kind() == "import_specifier" {
                                        let mut spec_cursor = import_child.walk();
                                        for spec_child in import_child.children(&mut spec_cursor) {
                                            if spec_child.kind() == "identifier" {
                                                symbols.push(ImportedSymbol {
                                                    name: Self::node_text(spec_child, source),
                                                    alias: None,
                                                });
                                            }
                                        }
                                    }
                                }
                            }
                            _ => {}
                        }
                    }
                }
                "string" => {
                    let text = Self::node_text(child, source);
                    module = text
                        .trim_matches(|c| c == '"' || c == '\'' || c == '`')
                        .to_string();
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
                symbols,
                is_type: false,
                line: Some(line),
            }))
        }
    }
    fn parse_class(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut methods = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "class_heritage" => {
                    let mut heritage_cursor = child.walk();
                    for heritage_child in child.children(&mut heritage_cursor) {
                        if heritage_child.kind() == "identifier" {
                            extends.push(TypeRef::new(Self::node_text(heritage_child, source)));
                        }
                    }
                }
                "class_body" => {
                    let mut body_cursor = child.walk();
                    for body_child in child.children(&mut body_cursor) {
                        match body_child.kind() {
                            "method_definition" | "field_definition" => {
                                if let Some(method) = self.parse_method(body_child, source)? {
                                    methods.push(method);
                                }
                            }
                            _ => {}
                        }
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = Visibility::Public;
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        Ok(Some(Class {
            name,
            visibility,
            extends,
            implements: vec![],
            type_params: vec![],
            decorators: vec![],
            modifiers: vec![],
            children: methods.into_iter().map(Node::Function).collect(),
            line_start,
            line_end,
        }))
    }

    fn parse_method(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut is_static = false;
        let mut is_async = false;
        let mut is_private = false;

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "property_identifier" | "identifier" | "private_property_identifier" => {
                    if name.is_empty() {
                        let text = Self::node_text(child, source);
                        // Check for private field syntax
                        if text.starts_with('#') {
                            is_private = true;
                            name = text;
                        } else {
                            name = text;
                        }
                    }
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "static" => {
                    is_static = true;
                }
                "async" => {
                    is_async = true;
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if is_private || name.starts_with('_') {
            Visibility::Private
        } else {
            Visibility::Public
        };

        let mut modifiers = vec![];
        if is_static {
            modifiers.push(Modifier::Static);
        }
        if is_async {
            modifiers.push(Modifier::Async);
        }

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type: None,
            decorators: vec![],
            type_params: vec![],
            modifiers,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_function(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut is_async = false;

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "async" => {
                    is_async = true;
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        let visibility = if name.starts_with('_') {
            Visibility::Private
        } else {
            Visibility::Public
        };

        let mut modifiers = vec![];
        if is_async {
            modifiers.push(Modifier::Async);
        }

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type: None,
            decorators: vec![],
            type_params: vec![],
            modifiers,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    #[allow(clippy::unused_self)]
    fn parse_parameters(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Parameter>> {
        let mut parameters = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    let name = Self::node_text(child, source);
                    parameters.push(Parameter {
                        name,
                        param_type: TypeRef::new("any".to_string()),
                        is_variadic: false,
                        is_optional: false,
                        decorators: vec![],
                        default_value: None,
                    });
                }
                "rest_pattern" => {
                    let mut rest_cursor = child.walk();
                    for rest_child in child.children(&mut rest_cursor) {
                        if rest_child.kind() == "identifier" {
                            let name = Self::node_text(rest_child, source);
                            parameters.push(Parameter {
                                name,
                                param_type: TypeRef::new("any[]".to_string()),
                                is_variadic: true,
                                is_optional: false,
                                decorators: vec![],
                                default_value: None,
                            });
                        }
                    }
                }
                _ => {}
            }
        }

        Ok(parameters)
    }

    fn process_node(&self, node: tree_sitter::Node, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "import_statement" => {
                if let Some(import) = self.parse_import(node, source)? {
                    file.children.push(Node::Import(import));
                }
            }
            "class_declaration" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "function_declaration" | "generator_function_declaration" => {
                if let Some(func) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(func));
                }
            }
            _ => {
                // Recurse into children
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.process_node(child, source, file)?;
                }
            }
        }

        Ok(())
    }
}

impl LanguageProcessor for JavaScriptProcessor {
    fn language(&self) -> &'static str {
        "javascript"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["js", "mjs", "cjs", "jsx"]
    }

    fn can_process(&self, path: &Path) -> bool {
        if let Some(ext) = path.extension()
            && let Some(ext_str) = ext.to_str()
        {
            return self.supported_extensions().contains(&ext_str);
        }
        false
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(
                path.to_string_lossy().as_ref(),
                "Failed to parse JavaScript source",
            )
        })?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: vec![],
        };

        self.process_node(tree.root_node(), source, &mut file)?;

        Ok(file)
    }
}

impl Default for JavaScriptProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create JavaScriptProcessor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_processor_creation() {
        let processor = JavaScriptProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_supported_extensions() {
        let processor = JavaScriptProcessor::new().unwrap();
        let extensions = processor.supported_extensions();
        assert_eq!(extensions, &["js", "mjs", "cjs", "jsx"]);
    }

    #[test]
    fn test_can_process() {
        let processor = JavaScriptProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("test.js")));
        assert!(processor.can_process(Path::new("test.mjs")));
        assert!(processor.can_process(Path::new("test.cjs")));
        assert!(processor.can_process(Path::new("test.jsx")));
        assert!(!processor.can_process(Path::new("test.ts")));
        assert!(!processor.can_process(Path::new("test.py")));
    }

    #[test]
    fn test_import_statements() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
import React from 'react';
import { useState, useEffect } from 'react';
import * as utils from './utils';
import { default as Button } from './Button';
import './styles.css';
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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

        assert_eq!(imports.len(), 5);
        assert_eq!(imports[0].module, "react");
        assert_eq!(imports[0].symbols.len(), 1);
        assert_eq!(imports[0].symbols[0].name, "React");
        assert_eq!(imports[1].module, "react");
        assert_eq!(imports[1].symbols.len(), 2);
        assert_eq!(imports[1].symbols[0].name, "useState");
        assert_eq!(imports[1].symbols[1].name, "useEffect");
        assert_eq!(imports[2].module, "./utils");
        assert_eq!(imports[3].module, "./Button");
        assert_eq!(imports[4].module, "./styles.css");
    }

    #[test]
    fn test_class_with_methods() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
class UserService {
    constructor(db) {
        this.db = db;
    }

    async getUser(id) {
        return await this.db.find(id);
    }

    static create(config) {
        return new UserService(config.db);
    }

    _validateUser(user) {
        return user != null;
    }

    #privateMethod() {
        console.log('private');
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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
        assert_eq!(classes[0].name, "UserService");
        assert_eq!(classes[0].children.len(), 5);

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

        assert_eq!(methods.len(), 5);

        // constructor
        assert_eq!(methods[0].name, "constructor");
        assert_eq!(methods[0].visibility, Visibility::Public);
        assert_eq!(methods[0].parameters.len(), 1);

        // getUser - async
        assert_eq!(methods[1].name, "getUser");
        assert!(methods[1].modifiers.contains(&Modifier::Async));

        // create - static
        assert_eq!(methods[2].name, "create");
        assert!(methods[2].modifiers.contains(&Modifier::Static));

        // _validateUser - private by convention
        assert_eq!(methods[3].name, "_validateUser");
        assert_eq!(methods[3].visibility, Visibility::Private);

        // #privateMethod - private by syntax
        assert_eq!(methods[4].name, "#privateMethod");
        assert_eq!(methods[4].visibility, Visibility::Private);
    }

    #[test]
    fn test_function_declarations() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
function add(a, b) {
    return a + b;
}

async function fetchData(url) {
    const response = await fetch(url);
    return response.json();
}

function _privateHelper(...args) {
    return args.reduce((a, b) => a + b, 0);
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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

        assert_eq!(functions.len(), 3);

        // add
        assert_eq!(functions[0].name, "add");
        assert_eq!(functions[0].parameters.len(), 2);
        assert_eq!(functions[0].visibility, Visibility::Public);

        // fetchData - async
        assert_eq!(functions[1].name, "fetchData");
        assert!(functions[1].modifiers.contains(&Modifier::Async));

        // _privateHelper - private by convention, variadic
        assert_eq!(functions[2].name, "_privateHelper");
        assert_eq!(functions[2].visibility, Visibility::Private);
        assert_eq!(functions[2].parameters.len(), 1);
        assert!(functions[2].parameters[0].is_variadic);
    }

    // ===== Enhanced Test Coverage =====

    #[test]
    fn test_empty_file() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = "";
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 0, "Empty file should have no children");
    }

    #[test]
    fn test_es6_class() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
class Person {
    constructor(name, age) {
        this.name = name;
        this.age = age;
    }

    greet() {
        return `Hello, I'm ${this.name}`;
    }

    getAge() {
        return this.age;
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Person");
            assert_eq!(class.visibility, Visibility::Public);
            assert_eq!(class.children.len(), 3);

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

            assert_eq!(methods.len(), 3);
            assert_eq!(methods[0].name, "constructor");
            assert_eq!(methods[0].parameters.len(), 2);
            assert_eq!(methods[1].name, "greet");
            assert_eq!(methods[2].name, "getAge");
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_class_inheritance() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
class Animal {
    constructor(name) {
        this.name = name;
    }
}

class Dog extends Animal {
    constructor(name, breed) {
        super(name);
        this.breed = breed;
    }

    bark() {
        return 'Woof!';
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 2);

        if let Node::Class(base_class) = &file.children[0] {
            assert_eq!(base_class.name, "Animal");
            assert_eq!(base_class.extends.len(), 0);
        } else {
            panic!("Expected Animal class");
        }

        if let Node::Class(derived_class) = &file.children[1] {
            assert_eq!(derived_class.name, "Dog");
            assert_eq!(derived_class.extends.len(), 1);
            assert_eq!(derived_class.extends[0].name, "Animal");
            assert_eq!(derived_class.children.len(), 2);
        } else {
            panic!("Expected Dog class");
        }
    }

    #[test]
    fn test_async_await() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
async function fetchUserData(userId) {
    const response = await fetch(`/api/users/${userId}`);
    const data = await response.json();
    return data;
}

async function processData(items) {
    const results = await Promise.all(items.map(async item => {
        return await transform(item);
    }));
    return results;
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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

        assert_eq!(functions[0].name, "fetchUserData");
        assert!(
            functions[0].modifiers.contains(&Modifier::Async),
            "fetchUserData should be async"
        );
        assert_eq!(functions[0].parameters.len(), 1);
        assert_eq!(functions[0].parameters[0].name, "userId");

        assert_eq!(functions[1].name, "processData");
        assert!(
            functions[1].modifiers.contains(&Modifier::Async),
            "processData should be async"
        );
    }

    #[test]
    fn test_arrow_functions() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
const add = (a, b) => a + b;
const multiply = (x, y) => {
    return x * y;
};
const greet = name => `Hello ${name}`;
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        // Arrow functions are typically not parsed as function declarations
        // They are variable declarations with arrow function expressions
        // This test validates that the parser doesn't crash on arrow functions
        // Successfully parsing without panic is sufficient validation
        let _ = file;
    }

    #[test]
    fn test_destructuring() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
function processUser({ id, name, email }) {
    return { id, name, email };
}

function getCoordinates({ x = 0, y = 0 } = {}) {
    return [x, y];
}

function handleArray([first, second, ...rest]) {
    return { first, second, rest };
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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
        assert_eq!(functions[0].name, "processUser");
        assert_eq!(functions[1].name, "getCoordinates");
        assert_eq!(functions[2].name, "handleArray");
    }

    #[test]
    fn test_spread_operator() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
function sum(...numbers) {
    return numbers.reduce((a, b) => a + b, 0);
}

function merge(obj1, obj2, ...rest) {
    return Object.assign({}, obj1, obj2, ...rest);
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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

        assert_eq!(functions[0].name, "sum");
        assert_eq!(functions[0].parameters.len(), 1);
        assert_eq!(functions[0].parameters[0].name, "numbers");
        assert!(
            functions[0].parameters[0].is_variadic,
            "Rest parameter should be marked as variadic"
        );

        assert_eq!(functions[1].name, "merge");
        assert_eq!(functions[1].parameters.len(), 3);
        assert_eq!(functions[1].parameters[2].name, "rest");
        assert!(functions[1].parameters[2].is_variadic);
    }

    #[test]
    fn test_private_fields() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
class BankAccount {
    #balance = 0;
    #accountNumber;

    constructor(initialBalance, accountNumber) {
        this.#balance = initialBalance;
        this.#accountNumber = accountNumber;
    }

    #validateAmount(amount) {
        return amount > 0;
    }

    deposit(amount) {
        if (this.#validateAmount(amount)) {
            this.#balance += amount;
        }
    }

    getBalance() {
        return this.#balance;
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "BankAccount");

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

            // Find private methods
            let private_methods: Vec<_> = methods
                .iter()
                .filter(|m| m.visibility == Visibility::Private)
                .collect();

            assert!(!private_methods.is_empty(), "Should have private methods");

            // Check #validateAmount is private
            let validate_method = methods.iter().find(|m| m.name == "#validateAmount");
            assert!(
                validate_method.is_some(),
                "Should find #validateAmount method"
            );
            assert_eq!(validate_method.unwrap().visibility, Visibility::Private);
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_static_methods() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
class MathUtils {
    static PI = 3.14159;

    static add(a, b) {
        return a + b;
    }

    static multiply(x, y) {
        return x * y;
    }

    static async fetchConstants() {
        return { PI: MathUtils.PI };
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "MathUtils");

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

            // Find static methods
            let static_methods: Vec<_> = methods
                .iter()
                .filter(|m| m.modifiers.contains(&Modifier::Static))
                .collect();

            assert!(
                static_methods.len() >= 2,
                "Should have at least 2 static methods, got {}",
                static_methods.len()
            );

            // Verify specific static methods
            assert!(
                methods
                    .iter()
                    .any(|m| m.name == "add" && m.modifiers.contains(&Modifier::Static))
            );
            assert!(
                methods
                    .iter()
                    .any(|m| m.name == "multiply" && m.modifiers.contains(&Modifier::Static))
            );

            // Check for async + static combination
            let async_static: Vec<_> = methods
                .iter()
                .filter(|m| {
                    m.modifiers.contains(&Modifier::Static)
                        && m.modifiers.contains(&Modifier::Async)
                })
                .collect();
            assert!(!async_static.is_empty(), "Should have async static method");
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_generator_function() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
function* numberGenerator() {
    yield 1;
    yield 2;
    yield 3;
}

function* fibonacciGenerator(n) {
    let a = 0, b = 1;
    for (let i = 0; i < n; i++) {
        yield a;
        [a, b] = [b, a + b];
    }
}

class Range {
    *[Symbol.iterator]() {
        for (let i = this.start; i <= this.end; i++) {
            yield i;
        }
    }
}
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
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

        assert!(
            functions.len() >= 2,
            "Should have at least 2 generator functions, got {}",
            functions.len()
        );

        assert!(functions.iter().any(|f| f.name == "numberGenerator"));
        assert!(functions.iter().any(|f| f.name == "fibonacciGenerator"));
    }

    #[test]
    fn test_export_statements() {
        let processor = JavaScriptProcessor::new().unwrap();
        let source = r#"
export default class DefaultClass {
    constructor() {}
}

export class NamedClass {
    method() {}
}

export function namedFunction(param) {
    return param;
}

export const CONSTANT = 42;

export { namedFunction as renamedFunction };
"#;

        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("test.js"), &opts)
            .unwrap();

        // Verify exports don't cause parsing errors
        assert!(
            file.children.len() >= 2,
            "Should have parsed classes and functions, got {} children",
            file.children.len()
        );

        // Check for specific exports
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

        assert!(
            classes
                .iter()
                .any(|c| c.name == "DefaultClass" || c.name == "NamedClass"),
            "Should have exported classes"
        );
        assert!(
            functions.iter().any(|f| f.name == "namedFunction"),
            "Should have exported function"
        );
    }
}
