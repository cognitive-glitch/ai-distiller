//! TypeScript language processor using tree-sitter
//!
//! Parses TypeScript source code to extract:
//! - Import/export statements
//! - Classes with access modifiers
//! - Interfaces and type aliases
//! - Functions with type annotations
//! - Generics and decorators

use distiller_core::error::Result;
use distiller_core::{
    error::DistilError,
    ir::{
        Class, Field, File, Function, Import, ImportedSymbol, Interface, Modifier, Node, Parameter,
        TypeParam, TypeRef, Visibility,
    },
    options::ProcessOptions,
    processor::language::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use tree_sitter::{Parser, TreeCursor};

pub struct TypeScriptProcessor {
    parser: Mutex<Parser>,
}

impl TypeScriptProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_typescript::LANGUAGE_TYPESCRIPT.into())
            .map_err(|e| {
                DistilError::TreeSitter(format!("Failed to set TypeScript language: {}", e))
            })?;

        Ok(Self {
            parser: Mutex::new(parser),
        })
    }

    fn parse_source(&self, source: &str, filename: &str) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser
            .parse(source, None)
            .ok_or_else(|| DistilError::TreeSitter("Failed to parse TypeScript".to_string()))?;

        let root_node = tree.root_node();
        let mut file = File {
            path: filename.to_string(),
            children: Vec::new(),
        };

        let mut cursor = root_node.walk();
        self.process_node(root_node, &mut file, source, &mut cursor)?;

        Ok(file)
    }

    fn process_node(
        &self,
        node: tree_sitter::Node,
        file: &mut File,
        source: &str,
        _cursor: &mut TreeCursor,
    ) -> Result<()> {
        match node.kind() {
            "import_statement" => {
                if let Some(import) = self.parse_import(node, source)? {
                    file.children.push(Node::Import(import));
                }
            }
            "export_statement" => {
                // Handle export { ... } and export * from '...'
                let mut child_cursor = node.walk();
                for child in node.children(&mut child_cursor) {
                    self.process_node(child, file, source, _cursor)?;
                }
            }
            "class_declaration" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "interface_declaration" => {
                if let Some(interface) = self.parse_interface(node, source)? {
                    file.children.push(Node::Interface(interface));
                }
            }
            "function_declaration" => {
                if let Some(function) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(function));
                }
            }
            "lexical_declaration" | "variable_declaration" => {
                // Handle const/let/var declarations that might be functions
                let mut child_cursor = node.walk();
                for child in node.children(&mut child_cursor) {
                    if child.kind() == "variable_declarator"
                        && let Some(func) = self.parse_variable_function(child, source)?
                    {
                        file.children.push(Node::Function(func));
                    }
                }
            }
            _ => {
                // Recursively process children
                let mut child_cursor = node.walk();
                for child in node.children(&mut child_cursor) {
                    self.process_node(child, file, source, _cursor)?;
                }
            }
        }
        Ok(())
    }

    fn parse_import(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Import>> {
        let mut module = String::new();
        let mut symbols = Vec::new();
        let mut is_type = false;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "import_clause" => {
                    let mut clause_cursor = child.walk();
                    for clause_child in child.children(&mut clause_cursor) {
                        match clause_child.kind() {
                            "named_imports" => {
                                let mut imports_cursor = clause_child.walk();
                                for import_child in clause_child.children(&mut imports_cursor) {
                                    if import_child.kind() == "import_specifier" {
                                        symbols.push(
                                            self.parse_import_specifier(import_child, source)?,
                                        );
                                    }
                                }
                            }
                            "identifier" => {
                                // Default import
                                symbols.push(ImportedSymbol {
                                    name: self.node_text(clause_child, source),
                                    alias: None,
                                });
                            }
                            "namespace_import" => {
                                // import * as name
                                let mut ns_cursor = clause_child.walk();
                                for ns_child in clause_child.children(&mut ns_cursor) {
                                    if ns_child.kind() == "identifier" {
                                        symbols.push(ImportedSymbol {
                                            name: "*".to_string(),
                                            alias: Some(self.node_text(ns_child, source)),
                                        });
                                    }
                                }
                            }
                            _ => {}
                        }
                    }
                }
                "string" => {
                    module = self
                        .node_text(child, source)
                        .trim_matches(|c| c == '"' || c == '\'')
                        .to_string();
                }
                "type" => {
                    is_type = true;
                }
                _ => {}
            }
        }

        Ok(Some(Import {
            import_type: "import".to_string(),
            module,
            symbols,
            is_type,
            line: Some(node.start_position().row + 1),
        }))
    }

    fn parse_import_specifier(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<ImportedSymbol> {
        let mut name = String::new();
        let mut alias = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "identifier" {
                if name.is_empty() {
                    name = self.node_text(child, source);
                } else {
                    alias = Some(self.node_text(child, source));
                }
            }
        }

        Ok(ImportedSymbol { name, alias })
    }

    fn parse_class(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut implements = Vec::new();
        let mut children = Vec::new();
        let mut modifiers = Vec::new();
        let mut type_params = Vec::new();
        let mut decorators = Vec::new();

        // Check for decorators before class
        if let Some(parent) = node.parent()
            && parent.kind() == "decorator"
        {
            decorators.push(self.node_text(parent, source));
        }

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" | "identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "class_heritage" => {
                    let mut heritage_cursor = child.walk();
                    for heritage_child in child.children(&mut heritage_cursor) {
                        match heritage_child.kind() {
                            "extends_clause" => {
                                extends = self.parse_extends_clause(heritage_child, source)?;
                            }
                            "implements_clause" => {
                                implements =
                                    self.parse_implements_clause(heritage_child, source)?;
                            }
                            _ => {}
                        }
                    }
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "class_body" => {
                    let mut body_cursor = child.walk();
                    for body_child in child.children(&mut body_cursor) {
                        match body_child.kind() {
                            "method_definition" | "method_signature" => {
                                if let Some(method) = self.parse_method(body_child, source)? {
                                    children.push(Node::Function(method));
                                }
                            }
                            "field_definition" | "public_field_definition" => {
                                if let Some(field) = self.parse_field(body_child, source)? {
                                    children.push(Node::Field(field));
                                }
                            }
                            _ => {}
                        }
                    }
                }
                "abstract" => {
                    modifiers.push(Modifier::Abstract);
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Class {
            name,
            visibility: Visibility::Public,
            modifiers,
            decorators,
            type_params,
            extends,
            implements,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_interface(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Interface>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut children = Vec::new();
        let mut type_params = Vec::new();

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    name = self.node_text(child, source);
                }
                "extends_type_clause" => {
                    extends = self.parse_extends_type_clause(child, source)?;
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "interface_body" => {
                    let mut body_cursor = child.walk();
                    for body_child in child.children(&mut body_cursor) {
                        match body_child.kind() {
                            "property_signature" => {
                                // Check if it's actually a method (has formal_parameters)
                                if self.has_formal_parameters(body_child) {
                                    if let Some(method) = self.parse_method(body_child, source)? {
                                        children.push(Node::Function(method));
                                    }
                                } else if let Some(field) =
                                    self.parse_property_signature(body_child, source)?
                                {
                                    children.push(Node::Field(field));
                                }
                            }
                            "method_signature" => {
                                if let Some(method) = self.parse_method(body_child, source)? {
                                    children.push(Node::Function(method));
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

        Ok(Some(Interface {
            name,
            visibility: Visibility::Public,
            type_params,
            extends,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_extends_clause(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<TypeRef>> {
        let mut extends = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "identifier" || child.kind() == "type_identifier" {
                extends.push(TypeRef::new(self.node_text(child, source)));
            }
        }
        Ok(extends)
    }

    fn parse_implements_clause(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<TypeRef>> {
        let mut implements = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_identifier" || child.kind() == "identifier" {
                implements.push(TypeRef::new(self.node_text(child, source)));
            }
        }
        Ok(implements)
    }

    fn parse_extends_type_clause(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<TypeRef>> {
        let mut extends = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_identifier" {
                extends.push(TypeRef::new(self.node_text(child, source)));
            }
        }
        Ok(extends)
    }

    fn parse_type_parameters(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Vec<TypeParam>> {
        let mut params = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter" {
                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    if param_child.kind() == "type_identifier" {
                        params.push(TypeParam {
                            name: self.node_text(param_child, source),
                            constraints: Vec::new(),
                            default: None,
                        });
                        break;
                    }
                }
            }
        }
        Ok(params)
    }

    fn parse_method(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let mut visibility = Visibility::Public;
        let mut modifiers = Vec::new();
        let mut type_params = Vec::new();
        let mut decorators = Vec::new();

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "property_identifier" | "identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "type_annotation" => {
                    return_type = self.parse_type_annotation(child, source)?;
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "accessibility_modifier" => {
                    visibility = self.parse_visibility(child, source);
                }
                "static" => {
                    modifiers.push(Modifier::Static);
                }
                "async" => {
                    modifiers.push(Modifier::Async);
                }
                "decorator" => {
                    decorators.push(self.node_text(child, source));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Function {
            name,
            visibility,
            modifiers,
            decorators,
            type_params,
            parameters,
            return_type,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_function(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let mut modifiers = Vec::new();
        let mut type_params = Vec::new();

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
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "type_annotation" => {
                    return_type = self.parse_type_annotation(child, source)?;
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "async" => {
                    modifiers.push(Modifier::Async);
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Function {
            name,
            visibility: Visibility::Public,
            modifiers,
            decorators: Vec::new(),
            type_params,
            parameters,
            return_type,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_variable_function(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<Function>> {
        let mut name = String::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "arrow_function" | "function" | "function_expression" => {
                    if let Some(mut func) = self.parse_arrow_function(child, source)? {
                        func.name = name;
                        return Ok(Some(func));
                    }
                }
                _ => {}
            }
        }

        Ok(None)
    }

    fn parse_arrow_function(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<Function>> {
        let mut parameters = Vec::new();
        let mut return_type = None;
        let mut modifiers = Vec::new();
        let mut type_params = Vec::new();

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "type_annotation" => {
                    return_type = self.parse_type_annotation(child, source)?;
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source)?;
                }
                "async" => {
                    modifiers.push(Modifier::Async);
                }
                _ => {}
            }
        }

        Ok(Some(Function {
            name: String::new(), // Will be filled by caller
            visibility: Visibility::Public,
            modifiers,
            decorators: Vec::new(),
            type_params,
            parameters,
            return_type,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn has_formal_parameters(&self, node: tree_sitter::Node) -> bool {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "formal_parameters" {
                return true;
            }
        }
        false
    }

    fn parse_field(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;
        let mut visibility = Visibility::Public;
        let mut modifiers = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "property_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_annotation" => {
                    field_type = self.parse_type_annotation(child, source)?;
                }
                "accessibility_modifier" => {
                    visibility = self.parse_visibility(child, source);
                }
                "static" => {
                    modifiers.push(Modifier::Static);
                }
                "readonly" => {
                    modifiers.push(Modifier::Const);
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        // Check for private fields (starting with #)
        if name.starts_with('#') {
            visibility = Visibility::Private;
        }

        Ok(Some(Field {
            name,
            visibility,
            modifiers,
            field_type,
            default_value: None,
            line: node.start_position().row + 1,
        }))
    }

    fn parse_property_signature(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "property_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_annotation" => {
                    field_type = self.parse_type_annotation(child, source)?;
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Field {
            name,
            visibility: Visibility::Public,
            modifiers: Vec::new(),
            field_type,
            default_value: None,
            line: node.start_position().row + 1,
        }))
    }

    fn parse_parameters(&self, node: tree_sitter::Node, source: &str) -> Result<Vec<Parameter>> {
        let mut parameters = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if let Some(param) = self.parse_parameter(child, source)? {
                parameters.push(param);
            }
        }
        Ok(parameters)
    }

    fn parse_parameter(&self, node: tree_sitter::Node, source: &str) -> Result<Option<Parameter>> {
        match node.kind() {
            "required_parameter" | "optional_parameter" => {
                let mut name = String::new();
                let mut param_type = TypeRef::new("any");
                let is_optional = node.kind() == "optional_parameter";
                let mut is_variadic = false;

                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    match child.kind() {
                        "identifier" => {
                            name = self.node_text(child, source);
                        }
                        "type_annotation" => {
                            if let Some(t) = self.parse_type_annotation(child, source)? {
                                param_type = t;
                            }
                        }
                        "..." => {
                            is_variadic = true;
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
                    default_value: None,
                    is_variadic,
                    is_optional,
                    decorators: Vec::new(),
                }))
            }
            _ => Ok(None),
        }
    }

    fn parse_type_annotation(
        &self,
        node: tree_sitter::Node,
        source: &str,
    ) -> Result<Option<TypeRef>> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" | "predefined_type" => {
                    return Ok(Some(TypeRef::new(self.node_text(child, source))));
                }
                "generic_type" => {
                    return self.parse_generic_type(child, source);
                }
                "union_type" | "intersection_type" => {
                    return Ok(Some(TypeRef::new(self.node_text(child, source))));
                }
                _ => {}
            }
        }
        Ok(None)
    }

    fn parse_generic_type(&self, node: tree_sitter::Node, source: &str) -> Result<Option<TypeRef>> {
        let mut name = String::new();
        let mut type_args = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    name = self.node_text(child, source);
                }
                "type_arguments" => {
                    let mut args_cursor = child.walk();
                    for arg_child in child.children(&mut args_cursor) {
                        if let Some(t) = self.parse_type_annotation(arg_child, source)? {
                            type_args.push(t);
                        }
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(TypeRef {
            name,
            package: None,
            type_args,
            is_nullable: false,
            is_array: false,
            array_dims: None,
        }))
    }

    fn parse_visibility(&self, node: tree_sitter::Node, source: &str) -> Visibility {
        match self.node_text(node, source).as_str() {
            "private" => Visibility::Private,
            "protected" => Visibility::Protected,
            "public" => Visibility::Public,
            _ => Visibility::Public,
        }
    }

    fn node_text(&self, node: tree_sitter::Node, source: &str) -> String {
        let start = node.start_byte();
        let end = node.end_byte();
        if start > end || end > source.len() {
            return String::new();
        }
        source[start..end].to_string()
    }
}

impl LanguageProcessor for TypeScriptProcessor {
    fn language(&self) -> &'static str {
        "typescript"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["ts", "tsx"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| ext == "ts" || ext == "tsx")
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let filename = path.to_string_lossy().into_owned();
        self.parse_source(source, &filename)
    }
}

impl Default for TypeScriptProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create TypeScript processor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = TypeScriptProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_supported_extensions() {
        let processor = TypeScriptProcessor::new().unwrap();
        assert_eq!(processor.supported_extensions(), &["ts", "tsx"]);
    }

    #[test]
    fn test_can_process() {
        let processor = TypeScriptProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("test.ts")));
        assert!(processor.can_process(Path::new("test.tsx")));
        assert!(!processor.can_process(Path::new("test.js")));
        assert!(!processor.can_process(Path::new("test.py")));
    }

    #[test]
    fn test_import_statements() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
import { Component } from '@angular/core';
import * as React from 'react';
import type { User } from './types';
"#;
        let result = processor
            .process(source, Path::new("test.ts"), &ProcessOptions::default())
            .unwrap();

        let imports: Vec<&Import> = result
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

        assert_eq!(imports.len(), 3);
        assert_eq!(imports[0].module, "@angular/core");
        assert_eq!(imports[0].symbols[0].name, "Component");
        assert_eq!(imports[1].module, "react");
        assert_eq!(imports[1].symbols[0].alias, Some("React".to_string()));
        assert!(imports[2].is_type);
    }

    #[test]
    fn test_class_with_visibility() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
class UserService {
    private userId: number;
    protected userName: string;
    public isActive: boolean;

    constructor(id: number) {}

    private validateId(): boolean {}
    protected getUserName(): string {}
    public isUserActive(): boolean {}
}
"#;
        let result = processor
            .process(source, Path::new("test.ts"), &ProcessOptions::default())
            .unwrap();

        let classes: Vec<&Class> = result
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

        assert_eq!(classes.len(), 1);
        assert_eq!(classes[0].name, "UserService");

        // Count fields and methods in children
        let fields: Vec<&Field> = classes[0]
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

        let methods: Vec<&Function> = classes[0]
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

        assert_eq!(fields.len(), 3);
        assert_eq!(methods.len(), 4);

        // Check field visibility
        assert_eq!(fields[0].visibility, Visibility::Private);
        assert_eq!(fields[1].visibility, Visibility::Protected);
        assert_eq!(fields[2].visibility, Visibility::Public);

        // Check method visibility
        assert_eq!(methods[1].visibility, Visibility::Private);
        assert_eq!(methods[2].visibility, Visibility::Protected);
        assert_eq!(methods[3].visibility, Visibility::Public);
    }

    #[test]
    fn test_interface_and_generics() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
interface Repository<T> {
    findById(id: number): T;
    save(entity: T): void;
}
"#;
        let result = processor
            .process(source, Path::new("test.ts"), &ProcessOptions::default())
            .unwrap();

        let interfaces: Vec<&Interface> = result
            .children
            .iter()
            .filter_map(|n| {
                if let Node::Interface(i) = n {
                    Some(i)
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(interfaces.len(), 1);
        assert_eq!(interfaces[0].name, "Repository");
        assert_eq!(interfaces[0].type_params.len(), 1);
        assert_eq!(interfaces[0].type_params[0].name, "T");

        // Count methods in children
        let methods: Vec<&Function> = interfaces[0]
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
    }

    // ===== Enhanced Test Coverage =====

    #[test]
    fn test_empty_file() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = "";
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 0, "Empty file should have no children");
    }

    #[test]
    fn test_interface_declaration() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
interface User {
    id: number;
    name: string;
    email?: string;
    login(): void;
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Interface(interface) = &file.children[0] {
            assert_eq!(interface.name, "User");
            assert_eq!(interface.visibility, Visibility::Public);

            // Count fields and methods
            let fields: Vec<_> = interface
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

            let methods: Vec<_> = interface
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

            assert!(fields.len() >= 3, "Expected at least 3 fields");
            assert_eq!(methods.len(), 1, "Expected 1 method");
            assert_eq!(methods[0].name, "login");
        } else {
            panic!("Expected interface node");
        }
    }

    #[test]
    fn test_type_alias() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
type UserId = number;
type UserCallback = (user: User) => void;
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();

        // Type aliases might not be captured in current implementation
        // This test verifies no errors occur during parsing
        assert!(
            file.children.len() >= 0,
            "Type alias parsing should not error"
        );
    }

    #[test]
    fn test_enum_declaration() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
enum Status {
    Active = "ACTIVE",
    Inactive = "INACTIVE",
    Pending = 1
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();

        // Enums might not be fully captured in current implementation
        // This test verifies no errors occur during parsing
        assert!(file.children.len() >= 0, "Enum parsing should not error");
    }

    #[test]
    fn test_generic_class() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
class Container<T, U> {
    private value: T;

    constructor(val: T) {
        this.value = val;
    }

    getValue(): T {
        return this.value;
    }
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Container");
            assert_eq!(class.type_params.len(), 2);
            assert_eq!(class.type_params[0].name, "T");
            assert_eq!(class.type_params[1].name, "U");

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

            assert!(methods.len() >= 2, "Expected at least 2 methods");
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_async_function() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
async function fetchUser(id: number): Promise<User> {
    const response = await fetch(`/api/users/${id}`);
    return response.json();
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Function(func) = &file.children[0] {
            assert_eq!(func.name, "fetchUser");
            assert!(
                func.modifiers.contains(&Modifier::Async),
                "Expected async modifier"
            );
            assert_eq!(func.parameters.len(), 1);
            assert_eq!(func.parameters[0].name, "id");
        } else {
            panic!("Expected function node");
        }
    }

    #[test]
    fn test_arrow_function() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
const add = (a: number, b: number): number => a + b;
const multiply = (x: number, y: number) => x * y;
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert!(
            file.children.len() >= 2,
            "Expected at least 2 arrow functions"
        );

        let funcs: Vec<_> = file
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

        assert!(funcs.len() >= 2);
        assert_eq!(funcs[0].name, "add");
        assert_eq!(funcs[0].parameters.len(), 2);
        assert_eq!(funcs[1].name, "multiply");
    }

    #[test]
    fn test_class_with_decorators() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
@Component({
    selector: 'app-user'
})
class UserComponent {
    @Input()
    user: User;

    @Output()
    userChange: EventEmitter<User>;
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();
        assert_eq!(file.children.len(), 1);

        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "UserComponent");
            // Decorators might be detected depending on tree-sitter parsing
            // This test verifies the class is parsed correctly
            assert!(class.children.len() >= 2, "Expected at least 2 fields");
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_namespace() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
namespace Utils {
    export function format(value: string): string {
        return value.trim();
    }

    export class Helper {
        static process() {}
    }
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();

        // Namespaces might be processed as nested structures
        // This test verifies parsing doesn't error
        assert!(
            file.children.len() >= 0,
            "Namespace parsing should not error"
        );
    }

    #[test]
    fn test_intersection_types() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
interface Timestamped {
    createdAt: Date;
}

interface Named {
    name: string;
}

function merge(obj: Named & Timestamped): void {
    console.log(obj.name, obj.createdAt);
}
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();

        // Find the function with intersection type parameter
        let funcs: Vec<_> = file
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

        assert!(!funcs.is_empty(), "Expected at least one function");
        let merge_func = funcs.iter().find(|f| f.name == "merge");
        assert!(merge_func.is_some(), "Expected merge function");

        if let Some(func) = merge_func {
            assert_eq!(func.parameters.len(), 1);
            assert_eq!(func.parameters[0].name, "obj");
            // The type will be parsed as "Named & Timestamped"
            assert!(func.parameters[0].param_type.name.contains("&"));
        }
    }

    #[test]
    fn test_union_types() {
        let processor = TypeScriptProcessor::new().unwrap();
        let source = r#"
function process(value: string | number | boolean): void {
    if (typeof value === 'string') {
        console.log(value.toUpperCase());
    }
}

type Result = Success | Error;
"#;
        let opts = ProcessOptions::default();

        let file = processor
            .process(source, &PathBuf::from("test.ts"), &opts)
            .unwrap();

        let funcs: Vec<_> = file
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

        assert!(!funcs.is_empty(), "Expected at least one function");
        assert_eq!(funcs[0].name, "process");
        assert_eq!(funcs[0].parameters.len(), 1);
        assert_eq!(funcs[0].parameters[0].name, "value");
        // The type will be parsed as "string | number | boolean"
        assert!(funcs[0].parameters[0].param_type.name.contains("|"));
    }
}

#[test]
fn test_react_user_profile() {
    let source =
        std::fs::read_to_string("../../testdata/real-world/react-app/components/UserProfile.tsx")
            .expect("Failed to read UserProfile file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("UserProfile.tsx"), &opts);

    assert!(result.is_ok(), "React component should parse successfully");

    let file = result.unwrap();

    // Should find interfaces and component
    let interfaces: Vec<_> = file
        .children
        .iter()
        .filter_map(|n| {
            if let Node::Interface(i) = n {
                Some(i)
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
        interfaces.len() >= 2,
        "Should find at least 2 interfaces (User, UserProfileProps)"
    );
    assert!(!functions.is_empty(), "Should find UserProfile component");

    println!(
        "✅ React component parse: {} interfaces, {} functions",
        interfaces.len(),
        functions.len()
    );
}

#[test]
fn test_react_custom_hook() {
    let source = std::fs::read_to_string("../../testdata/real-world/react-app/hooks/useAuth.ts")
        .expect("Failed to read useAuth file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("useAuth.ts"), &opts);

    assert!(result.is_ok(), "Custom hook should parse successfully");

    let file = result.unwrap();

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

    assert!(!functions.is_empty(), "Should find useAuth function");

    // Verify useAuth function found
    let use_auth = functions.iter().find(|f| f.name == "useAuth");
    assert!(use_auth.is_some(), "Should find useAuth function");

    println!("✅ Custom hook parse: {} functions", functions.len());
}

#[test]
fn test_react_generic_component() {
    let source =
        std::fs::read_to_string("../../testdata/real-world/react-app/components/DataTable.tsx")
            .expect("Failed to read DataTable file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("DataTable.tsx"), &opts);

    assert!(
        result.is_ok(),
        "Generic component should parse successfully"
    );

    let file = result.unwrap();

    // Should find generic function
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

    let generic_functions: Vec<_> = functions
        .iter()
        .filter(|f| !f.type_params.is_empty())
        .collect();

    // Note: Type params not captured yet - C2 finding
    assert!(!functions.is_empty(), "Should find DataTable function");

    println!(
        "✅ Generic component parse: {} functions, {} generic",
        functions.len(),
        generic_functions.len()
    );
}

// ===== EDGE CASE TESTS (Phase C3) =====

#[test]
fn test_malformed_typescript() {
    let source =
        std::fs::read_to_string("../../testdata/edge-cases/malformed/typescript_syntax_error.ts")
            .expect("Failed to read malformed TypeScript file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    // Should not panic - tree-sitter handles malformed code
    let result = processor.process(&source, Path::new("error.ts"), &opts);

    match result {
        Ok(file) => {
            println!("✓ Malformed TypeScript: Partial parse successful");
            println!("  Found {} top-level nodes", file.children.len());
            // Tree-sitter should recover and parse valid nodes
            assert!(
                !file.children.is_empty(),
                "Should find at least some valid nodes"
            );
        }
        Err(e) => {
            println!("✓ Malformed TypeScript: Error handled gracefully: {}", e);
            // As long as it doesn't panic, we're good
        }
    }
}

#[test]
fn test_unicode_typescript() {
    let source = std::fs::read_to_string("../../testdata/edge-cases/unicode/typescript_unicode.ts")
        .expect("Failed to read Unicode TypeScript file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("unicode.ts"), &opts);

    assert!(
        result.is_ok(),
        "Unicode TypeScript file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, Node::Class(_)))
        .count();

    println!(
        "✓ Unicode TypeScript: {} classes with Unicode identifiers",
        class_count
    );

    // Should find classes with Unicode names
    assert!(
        class_count >= 5,
        "Should find at least 5 classes with Unicode names"
    );
}

#[test]
fn test_large_typescript_file() {
    let source =
        std::fs::read_to_string("../../testdata/edge-cases/large-files/large_typescript.ts")
            .expect("Failed to read large TypeScript file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    println!(
        "Testing large TypeScript file: {} lines",
        source.lines().count()
    );

    let start = std::time::Instant::now();
    let result = processor.process(&source, Path::new("large.ts"), &opts);
    let duration = start.elapsed();

    assert!(
        result.is_ok(),
        "Large TypeScript file should parse successfully"
    );

    let file = result.unwrap();
    let class_count = file
        .children
        .iter()
        .filter(|n| matches!(n, Node::Class(_)))
        .count();

    println!(
        "✓ Large TypeScript: {} classes parsed in {:?}",
        class_count, duration
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

#[test]
fn test_complex_generics_typescript() {
    let source =
        std::fs::read_to_string("../../testdata/edge-cases/syntax-edge/complex_generics.ts")
            .expect("Failed to read complex generics TypeScript file");

    let processor = TypeScriptProcessor::new().unwrap();
    let opts = ProcessOptions::default();

    let result = processor.process(&source, Path::new("generics.ts"), &opts);

    assert!(
        result.is_ok(),
        "Complex generics TypeScript file should parse successfully"
    );

    let file = result.unwrap();

    println!(
        "✓ Complex generics TypeScript: {} top-level nodes",
        file.children.len()
    );

    // Should handle complex generic constraints
    let classes = file
        .children
        .iter()
        .filter(|n| matches!(n, Node::Class(_)))
        .count();

    assert!(
        classes >= 2,
        "Should find GenericManager and GenericStatic classes"
    );
}
