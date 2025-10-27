//! Text formatter for AI Distiller
//!
//! Ultra-compact plaintext format optimized for maximum AI context efficiency.
//! This is the most compact format with minimal syntax overhead, designed for
//! optimal AI consumption.

use distiller_core::ir::{
    Class, Comment, Enum, Field, File, Function, Import, Interface, Node, Package, Parameter,
    RawContent, Struct, TypeAlias, TypeParam, TypeRef, Visibility,
};
use std::fmt::Write as FmtWrite;

#[cfg(test)]
use distiller_core::ir::ImportedSymbol;

/// Text formatter options
#[derive(Debug, Clone, Default)]
pub struct TextFormatterOptions {
    /// Include implementation bodies
    pub include_implementation: bool,
}

/// Text formatter
pub struct TextFormatter {
    options: TextFormatterOptions,
}

impl TextFormatter {
    /// Create a new text formatter with default options
    #[must_use]
    pub fn new() -> Self {
        Self {
            options: TextFormatterOptions::default(),
        }
    }

    /// Create a new text formatter with custom options
    #[must_use]
    pub fn with_options(options: TextFormatterOptions) -> Self {
        Self { options }
    }

    /// Format a single file
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_file(&self, file: &File) -> Result<String, std::fmt::Error> {
        let mut output = String::new();
        writeln!(output, "<file path=\"{}\">", file.path)?;

        for child in &file.children {
            self.format_node(&mut output, child, 0)?;
        }

        writeln!(output, "</file>")?;
        Ok(output)
    }

    /// Format multiple files
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_files(&self, files: &[File]) -> Result<String, std::fmt::Error> {
        let mut output = String::new();

        for file in files {
            output.push_str(&self.format_file(file)?);
            output.push('\n'); // Blank line between files
        }

        Ok(output)
    }

    /// Format a node with indentation
    fn format_node(
        &self,
        output: &mut String,
        node: &Node,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        match node {
            Node::Import(import) => self.format_import(output, import, indent)?,
            Node::Class(class) => self.format_class(output, class, indent)?,
            Node::Interface(interface) => self.format_interface(output, interface, indent)?,
            Node::Struct(struct_node) => self.format_struct(output, struct_node, indent)?,
            Node::Enum(enum_node) => self.format_enum(output, enum_node, indent)?,
            Node::TypeAlias(alias) => self.format_type_alias(output, alias, indent)?,
            Node::Function(func) => self.format_function(output, func, indent)?,
            Node::Field(field) => self.format_field(output, field, indent)?,
            Node::Comment(comment) => self.format_comment(output, comment, indent)?,
            Node::Package(package) => self.format_package(output, package, indent)?,
            Node::RawContent(raw) => Self::format_raw(output, raw, indent)?,
            Node::File(_) | Node::Directory(_) => {
                // Files and directories are handled separately
            }
        }
        Ok(())
    }

    /// Format an import statement
    #[allow(clippy::unused_self)]
    fn format_import(
        &self,
        output: &mut String,
        import: &Import,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);

        if import.import_type == "from" {
            // Python-style: from module import symbol1, symbol2
            let symbols = import
                .symbols
                .iter()
                .map(|s| {
                    if let Some(ref alias) = s.alias {
                        format!("{} as {}", s.name, alias)
                    } else {
                        s.name.clone()
                    }
                })
                .collect::<Vec<_>>()
                .join(", ");
            writeln!(output, "{}from {} import {}", ind, import.module, symbols)?;
        } else {
            // Standard import
            if import.symbols.is_empty() {
                writeln!(output, "{}import {}", ind, import.module)?;
            } else {
                let symbols = import
                    .symbols
                    .iter()
                    .map(|s| {
                        if let Some(ref alias) = s.alias {
                            format!("{} as {}", s.name, alias)
                        } else {
                            s.name.clone()
                        }
                    })
                    .collect::<Vec<_>>()
                    .join(", ");
                writeln!(output, "{}import {} ({})", ind, import.module, symbols)?;
            }
        }

        Ok(())
    }

    /// Format a class
    fn format_class(
        &self,
        output: &mut String,
        class: &Class,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(class.visibility);

        // Write decorators
        for decorator in &class.decorators {
            writeln!(output, "{ind}@{decorator}")?;
        }

        // Write class declaration
        write!(output, "{}{}class {}", ind, vis_symbol, class.name)?;

        // Type parameters
        if !class.type_params.is_empty() {
            write!(output, "<{}>", self.format_type_params(&class.type_params))?;
        }

        // Inheritance
        let mut inheritance = Vec::new();
        if !class.extends.is_empty() {
            inheritance.extend(class.extends.iter().map(|t| self.format_type_ref(t)));
        }
        if !class.implements.is_empty() {
            inheritance.extend(class.implements.iter().map(|t| self.format_type_ref(t)));
        }
        if !inheritance.is_empty() {
            write!(output, "({})", inheritance.join(", "))?;
        }

        writeln!(output, ":")?;

        // Children
        for child in &class.children {
            self.format_node(output, child, indent + 1)?;
        }

        Ok(())
    }

    /// Format an interface
    fn format_interface(
        &self,
        output: &mut String,
        interface: &Interface,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(interface.visibility);

        write!(output, "{}{}interface {}", ind, vis_symbol, interface.name)?;

        if !interface.type_params.is_empty() {
            write!(
                output,
                "<{}>",
                self.format_type_params(&interface.type_params)
            )?;
        }

        if !interface.extends.is_empty() {
            let extends = interface
                .extends
                .iter()
                .map(|t| self.format_type_ref(t))
                .collect::<Vec<_>>()
                .join(", ");
            write!(output, "({extends})")?;
        }

        writeln!(output, ":")?;

        for child in &interface.children {
            self.format_node(output, child, indent + 1)?;
        }

        Ok(())
    }

    /// Format a struct
    fn format_struct(
        &self,
        output: &mut String,
        struct_node: &Struct,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(struct_node.visibility);

        write!(output, "{}{}struct {}", ind, vis_symbol, struct_node.name)?;

        if !struct_node.type_params.is_empty() {
            write!(
                output,
                "<{}>",
                self.format_type_params(&struct_node.type_params)
            )?;
        }

        writeln!(output, ":")?;

        for child in &struct_node.children {
            self.format_node(output, child, indent + 1)?;
        }

        Ok(())
    }

    /// Format an enum
    fn format_enum(
        &self,
        output: &mut String,
        enum_node: &Enum,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(enum_node.visibility);

        write!(output, "{}{}enum {}", ind, vis_symbol, enum_node.name)?;

        if let Some(ref enum_type) = enum_node.enum_type {
            write!(output, ": {}", self.format_type_ref(enum_type))?;
        }

        writeln!(output, ":")?;

        for child in &enum_node.children {
            self.format_node(output, child, indent + 1)?;
        }

        Ok(())
    }

    /// Format a type alias
    fn format_type_alias(
        &self,
        output: &mut String,
        alias: &TypeAlias,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(alias.visibility);

        write!(output, "{}{}type {}", ind, vis_symbol, alias.name)?;

        if !alias.type_params.is_empty() {
            write!(output, "<{}>", self.format_type_params(&alias.type_params))?;
        }

        writeln!(output, " = {}", self.format_type_ref(&alias.alias_type))?;

        Ok(())
    }

    /// Format a function
    fn format_function(
        &self,
        output: &mut String,
        func: &Function,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(func.visibility);

        // Write decorators
        for decorator in &func.decorators {
            writeln!(output, "{ind}@{decorator}")?;
        }

        // Modifiers
        let mut modifiers = func
            .modifiers
            .iter()
            .map(|m| format!("{m:?}").to_lowercase())
            .collect::<Vec<_>>();
        let modifiers_str = if modifiers.is_empty() {
            String::new()
        } else {
            modifiers.push(String::new());
            modifiers.join(" ")
        };

        write!(
            output,
            "{}{}{}def {}",
            ind, vis_symbol, modifiers_str, func.name
        )?;

        // Type parameters
        if !func.type_params.is_empty() {
            write!(output, "<{}>", self.format_type_params(&func.type_params))?;
        }

        // Parameters
        write!(output, "(")?;
        let params = func
            .parameters
            .iter()
            .map(|p| self.format_parameter(p))
            .collect::<Vec<_>>()
            .join(", ");
        write!(output, "{params}")?;
        write!(output, ")")?;

        // Return type
        if let Some(ref ret_type) = func.return_type {
            write!(output, " -> {}", self.format_type_ref(ret_type))?;
        }

        if self.options.include_implementation {
            if let Some(ref implementation) = func.implementation {
                writeln!(output, ":")?;
                // Indent implementation
                for line in implementation.lines() {
                    writeln!(output, "{ind}    {line}")?;
                }
            } else {
                writeln!(output)?;
            }
        } else {
            writeln!(output)?;
        }

        Ok(())
    }

    /// Format a field
    fn format_field(
        &self,
        output: &mut String,
        field: &Field,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        let vis_symbol = Self::visibility_symbol(field.visibility);

        // Modifiers
        let mut modifiers = field
            .modifiers
            .iter()
            .map(|m| format!("{m:?}").to_lowercase())
            .collect::<Vec<_>>();
        let modifiers_str = if modifiers.is_empty() {
            String::new()
        } else {
            modifiers.push(String::new());
            modifiers.join(" ")
        };

        write!(
            output,
            "{}{}{}{}",
            ind, vis_symbol, modifiers_str, field.name
        )?;

        if let Some(ref field_type) = field.field_type {
            write!(output, ": {}", self.format_type_ref(field_type))?;
        }

        if let Some(ref default_value) = field.default_value {
            write!(output, " = {default_value}")?;
        }

        writeln!(output)?;

        Ok(())
    }

    /// Format a comment
    #[allow(clippy::unused_self)]
    fn format_comment(
        &self,
        output: &mut String,
        comment: &Comment,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);

        for line in comment.text.lines() {
            writeln!(output, "{ind}# {line}")?;
        }

        Ok(())
    }

    /// Format a package
    fn format_package(
        &self,
        output: &mut String,
        package: &Package,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = Self::indent(indent);
        writeln!(output, "{}package {}", ind, package.name)?;

        for child in &package.children {
            self.format_node(output, child, indent)?;
        }

        Ok(())
    }

    /// Format raw content
    fn format_raw(
        output: &mut String,
        raw: &RawContent,
        _indent: usize,
    ) -> Result<(), std::fmt::Error> {
        writeln!(output, "{}", raw.content)?;
        Ok(())
    }

    /// Format a type reference
    #[allow(clippy::only_used_in_recursion)]
    fn format_type_ref(&self, type_ref: &TypeRef) -> String {
        let mut result = type_ref.name.clone();

        if !type_ref.type_args.is_empty() {
            let args = type_ref
                .type_args
                .iter()
                .map(|t| self.format_type_ref(t))
                .collect::<Vec<_>>()
                .join(", ");
            write!(result, "<{args}>").unwrap();
        }

        if type_ref.is_array {
            result.push_str("[]");
        }

        if type_ref.is_nullable {
            result.push('?');
        }

        result
    }

    /// Format type parameters
    fn format_type_params(&self, type_params: &[TypeParam]) -> String {
        type_params
            .iter()
            .map(|tp| {
                let mut result = tp.name.clone();
                if !tp.constraints.is_empty() {
                    let constraints = tp
                        .constraints
                        .iter()
                        .map(|t| self.format_type_ref(t))
                        .collect::<Vec<_>>()
                        .join(" + ");
                    write!(result, ": {constraints}").unwrap();
                }
                result
            })
            .collect::<Vec<_>>()
            .join(", ")
    }

    /// Format a parameter
    fn format_parameter(&self, param: &Parameter) -> String {
        let mut result = param.name.clone();
        write!(result, ": {}", self.format_type_ref(&param.param_type)).unwrap();

        if let Some(ref default) = param.default_value {
            write!(result, " = {default}").unwrap();
        }

        if param.is_optional {
            result.push('?');
        }

        if param.is_variadic {
            result = format!("...{result}");
        }

        result
    }

    /// Get visibility symbol
    fn visibility_symbol(visibility: Visibility) -> &'static str {
        match visibility {
            Visibility::Public => "",
            Visibility::Private => "-",
            Visibility::Protected => "*",
            Visibility::Internal => "~",
        }
    }

    /// Generate indentation string
    fn indent(level: usize) -> String {
        "    ".repeat(level)
    }
}

impl Default for TextFormatter {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_simple_class() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Class(Class {
                name: "Example".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                extends: Vec::new(),
                implements: Vec::new(),
                children: vec![Node::Function(Function {
                    name: "__init__".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: vec![Parameter {
                        name: "self".to_string(),
                        param_type: TypeRef::new("Self"),
                        default_value: None,
                        is_variadic: false,
                        is_optional: false,
                        decorators: Vec::new(),
                    }],
                    return_type: None,
                    implementation: None,
                    line_start: 2,
                    line_end: 3,
                })],
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = TextFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("<file path=\"test.py\">"));
        assert!(result.contains("class Example:"));
        assert!(result.contains("def __init__(self: Self)"));
        assert!(result.contains("</file>"));
    }

    #[test]
    fn test_visibility_symbols() {
        let _formatter = TextFormatter::new();
        assert_eq!(TextFormatter::visibility_symbol(Visibility::Public), "");
        assert_eq!(TextFormatter::visibility_symbol(Visibility::Private), "-");
        assert_eq!(TextFormatter::visibility_symbol(Visibility::Protected), "*");
        assert_eq!(TextFormatter::visibility_symbol(Visibility::Internal), "~");
    }

    #[test]
    fn test_import() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Import(Import {
                import_type: "from".to_string(),
                module: "typing".to_string(),
                symbols: vec![
                    ImportedSymbol {
                        name: "List".to_string(),
                        alias: None,
                    },
                    ImportedSymbol {
                        name: "Optional".to_string(),
                        alias: None,
                    },
                ],
                is_type: true,
                line: Some(1),
            })],
        };

        let formatter = TextFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("from typing import List, Optional"));
    }

    #[test]
    fn test_private_field() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Class(Class {
                name: "Example".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                extends: Vec::new(),
                implements: Vec::new(),
                children: vec![Node::Field(Field {
                    name: "_private_field".to_string(),
                    visibility: Visibility::Private,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("str")),
                    default_value: None,
                    line: 2,
                })],
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = TextFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("-_private_field: str"));
    }
}
