//! XML formatter for AI Distiller
//!
//! Structured XML output for legacy systems and XML-based tooling.
//! Provides proper XML escaping and semantic structure.

use distiller_core::ir::{
    Class, Comment, Directory, Enum, Field, File, Function, Import, Interface, Modifier, Node,
    Package, Parameter, RawContent, Struct, TypeAlias, TypeParam, TypeRef, Visibility,
};
use std::fmt::Write;

/// XML formatter options
#[derive(Debug, Clone)]
pub struct XmlFormatterOptions {
    /// Indent output for readability
    pub indent: bool,
    /// Number of spaces per indent level
    pub indent_size: usize,
}

impl Default for XmlFormatterOptions {
    fn default() -> Self {
        Self {
            indent: true,
            indent_size: 2,
        }
    }
}

/// XML formatter
pub struct XmlFormatter {
    options: XmlFormatterOptions,
}

impl XmlFormatter {
    /// Create a new XML formatter with default options
    #[must_use]
    pub fn new() -> Self {
        Self {
            options: XmlFormatterOptions::default(),
        }
    }

    /// Create a new XML formatter with custom options
    #[must_use]
    pub fn with_options(options: XmlFormatterOptions) -> Self {
        Self { options }
    }

    /// Format a single file as XML
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_file(&self, file: &File) -> Result<String, std::fmt::Error> {
        let mut output = String::new();
        writeln!(output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")?;
        self.format_file_element(&mut output, file, 0)?;
        Ok(output)
    }

    /// Format multiple files as XML
    ///
    /// # Errors
    ///
    /// Returns an error if formatting or serialization fails
    pub fn format_files(&self, files: &[File]) -> Result<String, std::fmt::Error> {
        let mut output = String::new();
        writeln!(output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")?;
        writeln!(output, "<files>")?;

        for file in files {
            self.format_file_element(&mut output, file, 1)?;
        }

        writeln!(output, "</files>")?;
        Ok(output)
    }

    /// Format a file element
    fn format_file_element(
        &self,
        output: &mut String,
        file: &File,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{}<file path=\"{}\">", ind, escape_xml(&file.path))?;

        for child in &file.children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{ind}</file>")?;
        Ok(())
    }

    /// Format a node
    fn format_node(
        &self,
        output: &mut String,
        node: &Node,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        match node {
            Node::File(file) => self.format_file_element(output, file, indent),
            Node::Directory(dir) => self.format_directory(output, dir, indent),
            Node::Package(pkg) => self.format_package(output, pkg, indent),
            Node::Import(import) => self.format_import(output, import, indent),
            Node::Class(class) => self.format_class(output, class, indent),
            Node::Interface(interface) => self.format_interface(output, interface, indent),
            Node::Struct(struct_node) => self.format_struct(output, struct_node, indent),
            Node::Enum(enum_node) => self.format_enum(output, enum_node, indent),
            Node::TypeAlias(type_alias) => self.format_type_alias(output, type_alias, indent),
            Node::Function(function) => self.format_function(output, function, indent),
            Node::Field(field) => self.format_field(output, field, indent),
            Node::Comment(comment) => self.format_comment(output, comment, indent),
            Node::RawContent(raw) => self.format_raw_content(output, raw, indent),
        }
    }

    /// Format a directory
    fn format_directory(
        &self,
        output: &mut String,
        dir: &Directory,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(
            output,
            "{}<directory path=\"{}\">",
            ind,
            escape_xml(&dir.path)
        )?;
        for child in &dir.children {
            self.format_node(output, child, indent + 1)?;
        }
        writeln!(output, "{ind}</directory>")?;
        Ok(())
    }

    /// Format a package
    fn format_package(
        &self,
        output: &mut String,
        pkg: &Package,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(
            output,
            "{}<package name=\"{}\">",
            ind,
            escape_xml(&pkg.name)
        )?;
        for child in &pkg.children {
            self.format_node(output, child, indent + 1)?;
        }
        writeln!(output, "{ind}</package>")?;
        Ok(())
    }

    /// Format an import
    fn format_import(
        &self,
        output: &mut String,
        import: &Import,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{ind}<import")?;
        write!(output, " type=\"{}\"", import.import_type)?;
        write!(output, " module=\"{}\"", escape_xml(&import.module))?;
        if let Some(line) = import.line {
            write!(output, " line=\"{line}\"")?;
        }

        if import.symbols.is_empty() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            for symbol in &import.symbols {
                let symbol_ind = self.indent(indent + 1);
                write!(
                    output,
                    "{}<symbol name=\"{}\"",
                    symbol_ind,
                    escape_xml(&symbol.name)
                )?;
                if let Some(ref alias) = symbol.alias {
                    write!(output, " alias=\"{}\"", escape_xml(alias))?;
                }
                writeln!(output, " />")?;
            }
            writeln!(output, "{ind}</import>")?;
        }
        Ok(())
    }

    /// Generic helper for formatting container types (class, interface, struct, enum)
    #[allow(clippy::too_many_arguments)]
    fn format_container(
        &self,
        output: &mut String,
        tag: &str,
        name: &str,
        visibility: Visibility,
        line_start: usize,
        line_end: usize,
        modifiers: &[Modifier],
        decorators: &[String],
        type_params: &[TypeParam],
        extends: &[TypeRef],
        implements: &[TypeRef],
        children: &[Node],
        enum_type: Option<&TypeRef>,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);

        // Decorators
        for decorator in decorators {
            writeln!(
                output,
                "{}<decorator value=\"{}\" />",
                ind,
                escape_xml(decorator)
            )?;
        }

        // Opening tag with attributes
        write!(output, "{ind}<{tag}")?;
        write!(output, " name=\"{}\"", escape_xml(name))?;
        write!(output, " visibility=\"{}\"", visibility_str(visibility))?;
        write!(output, " line-start=\"{line_start}\"")?;
        write!(output, " line-end=\"{line_end}\"")?;
        if !modifiers.is_empty() {
            write!(
                output,
                " modifiers=\"{}\"",
                escape_xml(&modifiers_to_string(modifiers))
            )?;
        }

        // Check if self-closing
        if type_params.is_empty()
            && extends.is_empty()
            && implements.is_empty()
            && children.is_empty()
            && enum_type.is_none()
        {
            writeln!(output, " />")?;
            return Ok(());
        }

        writeln!(output, ">")?;

        // Type parameters
        if !type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{type_params_ind}<type-params>")?;
            for param in type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{type_params_ind}</type-params>")?;
        }

        // Enum type
        if let Some(et) = enum_type {
            let type_ind = self.indent(indent + 1);
            writeln!(output, "{type_ind}<type>")?;
            self.format_type_ref(output, et, indent + 2)?;
            writeln!(output, "{type_ind}</type>")?;
        }

        // Extends
        if !extends.is_empty() {
            let extends_ind = self.indent(indent + 1);
            writeln!(output, "{extends_ind}<extends>")?;
            for type_ref in extends {
                self.format_type_ref(output, type_ref, indent + 2)?;
            }
            writeln!(output, "{extends_ind}</extends>")?;
        }

        // Implements
        if !implements.is_empty() {
            let implements_ind = self.indent(indent + 1);
            writeln!(output, "{implements_ind}<implements>")?;
            for type_ref in implements {
                self.format_type_ref(output, type_ref, indent + 2)?;
            }
            writeln!(output, "{implements_ind}</implements>")?;
        }

        // Children
        for child in children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{ind}</{tag}>")?;
        Ok(())
    }

    /// Format a class
    fn format_class(
        &self,
        output: &mut String,
        class: &Class,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        self.format_container(
            output,
            "class",
            &class.name,
            class.visibility,
            class.line_start,
            class.line_end,
            &class.modifiers,
            &class.decorators,
            &class.type_params,
            &class.extends,
            &class.implements,
            &class.children,
            None,
            indent,
        )
    }

    /// Format an interface
    fn format_interface(
        &self,
        output: &mut String,
        interface: &Interface,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        self.format_container(
            output,
            "interface",
            &interface.name,
            interface.visibility,
            interface.line_start,
            interface.line_end,
            &[], // interfaces don't have modifiers
            &[], // interfaces don't have decorators
            &interface.type_params,
            &interface.extends,
            &[], // interfaces don't implement
            &interface.children,
            None,
            indent,
        )
    }

    /// Format a struct
    fn format_struct(
        &self,
        output: &mut String,
        struct_node: &Struct,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        self.format_container(
            output,
            "struct",
            &struct_node.name,
            struct_node.visibility,
            struct_node.line_start,
            struct_node.line_end,
            &[], // structs don't have modifiers in IR
            &[], // structs don't have decorators
            &struct_node.type_params,
            &[], // structs don't extend
            &[], // structs don't implement
            &struct_node.children,
            None,
            indent,
        )
    }

    /// Format an enum
    fn format_enum(
        &self,
        output: &mut String,
        enum_node: &Enum,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        self.format_container(
            output,
            "enum",
            &enum_node.name,
            enum_node.visibility,
            enum_node.line_start,
            enum_node.line_end,
            &[], // enums don't have modifiers
            &[], // enums don't have decorators
            &[], // enums don't have type params
            &[], // enums don't extend
            &[], // enums don't implement
            &enum_node.children,
            enum_node.enum_type.as_ref(),
            indent,
        )
    }

    /// Format a type alias
    fn format_type_alias(
        &self,
        output: &mut String,
        type_alias: &TypeAlias,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{ind}<type-alias")?;
        write!(output, " name=\"{}\"", escape_xml(&type_alias.name))?;
        write!(
            output,
            " visibility=\"{}\"",
            visibility_str(type_alias.visibility)
        )?;
        write!(output, " line=\"{}\"", type_alias.line)?;
        writeln!(output, ">")?;

        if !type_alias.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{type_params_ind}<type-params>")?;
            for param in &type_alias.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{type_params_ind}</type-params>")?;
        }

        let alias_ind = self.indent(indent + 1);
        writeln!(output, "{alias_ind}<alias-type>")?;
        self.format_type_ref(output, &type_alias.alias_type, indent + 2)?;
        writeln!(output, "{alias_ind}</alias-type>")?;

        writeln!(output, "{ind}</type-alias>")?;
        Ok(())
    }

    /// Format a function
    fn format_function(
        &self,
        output: &mut String,
        function: &Function,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        for decorator in &function.decorators {
            writeln!(
                output,
                "{}<decorator value=\"{}\" />",
                ind,
                escape_xml(decorator)
            )?;
        }

        write!(output, "{ind}<function")?;
        write!(output, " name=\"{}\"", escape_xml(&function.name))?;
        write!(
            output,
            " visibility=\"{}\"",
            visibility_str(function.visibility)
        )?;
        write!(output, " line-start=\"{}\"", function.line_start)?;
        write!(output, " line-end=\"{}\"", function.line_end)?;
        if !function.modifiers.is_empty() {
            write!(
                output,
                " modifiers=\"{}\"",
                escape_xml(&modifiers_to_string(&function.modifiers))
            )?;
        }
        writeln!(output, ">")?;

        if !function.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{type_params_ind}<type-params>")?;
            for param in &function.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{type_params_ind}</type-params>")?;
        }

        if !function.parameters.is_empty() {
            let params_ind = self.indent(indent + 1);
            writeln!(output, "{params_ind}<parameters>")?;
            for param in &function.parameters {
                self.format_parameter(output, param, indent + 2)?;
            }
            writeln!(output, "{params_ind}</parameters>")?;
        }

        if let Some(ref return_type) = function.return_type {
            let return_ind = self.indent(indent + 1);
            writeln!(output, "{return_ind}<return-type>")?;
            self.format_type_ref(output, return_type, indent + 2)?;
            writeln!(output, "{return_ind}</return-type>")?;
        }

        if let Some(ref impl_body) = function.implementation {
            let impl_ind = self.indent(indent + 1);
            writeln!(output, "{impl_ind}<implementation>")?;
            writeln!(output, "{}", escape_xml(impl_body))?;
            writeln!(output, "{impl_ind}</implementation>")?;
        }

        writeln!(output, "{ind}</function>")?;
        Ok(())
    }

    /// Format a field
    fn format_field(
        &self,
        output: &mut String,
        field: &Field,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{ind}<field")?;
        write!(output, " name=\"{}\"", escape_xml(&field.name))?;
        write!(
            output,
            " visibility=\"{}\"",
            visibility_str(field.visibility)
        )?;
        write!(output, " line=\"{}\"", field.line)?;
        if !field.modifiers.is_empty() {
            write!(
                output,
                " modifiers=\"{}\"",
                escape_xml(&modifiers_to_string(&field.modifiers))
            )?;
        }

        if let Some(ref field_type) = field.field_type {
            writeln!(output, ">")?;
            let type_ind = self.indent(indent + 1);
            writeln!(output, "{type_ind}<type>")?;
            self.format_type_ref(output, field_type, indent + 2)?;
            writeln!(output, "{type_ind}</type>")?;
            if let Some(ref default_value) = field.default_value {
                writeln!(
                    output,
                    "{}<default-value>{}</default-value>",
                    type_ind,
                    escape_xml(default_value)
                )?;
            }
            writeln!(output, "{ind}</field>")?;
        } else {
            if let Some(ref default_value) = field.default_value {
                write!(output, " default=\"{}\"", escape_xml(default_value))?;
            }
            writeln!(output, " />")?;
        }
        Ok(())
    }

    /// Format a comment
    fn format_comment(
        &self,
        output: &mut String,
        comment: &Comment,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{ind}<comment")?;
        write!(output, " format=\"{}\"", comment.format)?;
        write!(output, " line=\"{}\"", comment.line)?;
        writeln!(output, ">")?;
        writeln!(output, "{}", escape_xml(&comment.text))?;
        writeln!(output, "{ind}</comment>")?;
        Ok(())
    }

    /// Format raw content
    fn format_raw_content(
        &self,
        output: &mut String,
        raw: &RawContent,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{ind}<raw-content>")?;
        writeln!(output, "{}", escape_xml(&raw.content))?;
        writeln!(output, "{ind}</raw-content>")?;
        Ok(())
    }

    /// Format a type parameter
    fn format_type_param(
        &self,
        output: &mut String,
        param: &TypeParam,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(
            output,
            "{}<type-param name=\"{}\"",
            ind,
            escape_xml(&param.name)
        )?;

        if param.constraints.is_empty() && param.default.is_none() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            if !param.constraints.is_empty() {
                let constraints_ind = self.indent(indent + 1);
                writeln!(output, "{constraints_ind}<constraints>")?;
                for constraint in &param.constraints {
                    self.format_type_ref(output, constraint, indent + 2)?;
                }
                writeln!(output, "{constraints_ind}</constraints>")?;
            }
            if let Some(ref default) = param.default {
                let default_ind = self.indent(indent + 1);
                writeln!(output, "{default_ind}<default>")?;
                self.format_type_ref(output, default, indent + 2)?;
                writeln!(output, "{default_ind}</default>")?;
            }
            writeln!(output, "{ind}</type-param>")?;
        }
        Ok(())
    }

    /// Format a type reference
    fn format_type_ref(
        &self,
        output: &mut String,
        type_ref: &TypeRef,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(
            output,
            "{}<type name=\"{}\"",
            ind,
            escape_xml(&type_ref.name)
        )?;

        if type_ref.type_args.is_empty() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            let args_ind = self.indent(indent + 1);
            writeln!(output, "{args_ind}<type-args>")?;
            for arg in &type_ref.type_args {
                self.format_type_ref(output, arg, indent + 2)?;
            }
            writeln!(output, "{args_ind}</type-args>")?;
            writeln!(output, "{ind}</type>")?;
        }
        Ok(())
    }

    /// Format a parameter
    fn format_parameter(
        &self,
        output: &mut String,
        param: &Parameter,
        indent: usize,
    ) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(
            output,
            "{}<parameter name=\"{}\"",
            ind,
            escape_xml(&param.name)
        )?;
        if param.is_variadic {
            write!(output, " variadic=\"true\"")?;
        }
        if param.is_optional {
            write!(output, " optional=\"true\"")?;
        }
        writeln!(output, ">")?;

        let type_ind = self.indent(indent + 1);
        writeln!(output, "{type_ind}<type>")?;
        self.format_type_ref(output, &param.param_type, indent + 2)?;
        writeln!(output, "{type_ind}</type>")?;

        if let Some(ref default_value) = param.default_value {
            writeln!(
                output,
                "{}<default-value>{}</default-value>",
                type_ind,
                escape_xml(default_value)
            )?;
        }

        writeln!(output, "{ind}</parameter>")?;
        Ok(())
    }

    /// Get indentation string
    fn indent(&self, level: usize) -> String {
        if self.options.indent {
            " ".repeat(level * self.options.indent_size)
        } else {
            String::new()
        }
    }
}

impl Default for XmlFormatter {
    fn default() -> Self {
        Self::new()
    }
}

/// Escape XML special characters
fn escape_xml(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&apos;")
}

/// Convert visibility to string
fn visibility_str(vis: Visibility) -> &'static str {
    match vis {
        Visibility::Public => "public",
        Visibility::Private => "private",
        Visibility::Protected => "protected",
        Visibility::Internal => "internal",
    }
}

/// Convert modifiers to comma-separated string
fn modifiers_to_string(modifiers: &[Modifier]) -> String {
    modifiers
        .iter()
        .map(|m| match m {
            Modifier::Static => "static",
            Modifier::Abstract => "abstract",
            Modifier::Final => "final",
            Modifier::Async => "async",
            Modifier::Virtual => "virtual",
            Modifier::Override => "override",
            Modifier::Const => "const",
            Modifier::Readonly => "readonly",
            Modifier::Mutable => "mutable",
            Modifier::Event => "event",
            Modifier::Data => "data",
            Modifier::Sealed => "sealed",
            Modifier::Inline => "inline",
        })
        .collect::<Vec<_>>()
        .join(",")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_xml_format_simple() {
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

        let formatter = XmlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.starts_with("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"));
        assert!(result.contains("<file path=\"test.py\">"));
        assert!(result.contains("<class name=\"Example\""));
        assert!(result.contains("<function name=\"__init__\""));
        assert!(result.contains("</file>"));
        assert!(result.contains("</class>"));
        assert!(result.contains("</function>"));
    }

    #[test]
    fn test_xml_escaping() {
        let file = File {
            path: "file<>&\".py".to_string(),
            children: vec![Node::Function(Function {
                name: "func<>&\"".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                parameters: Vec::new(),
                return_type: None,
                implementation: None,
                line_start: 1,
                line_end: 2,
            })],
        };

        let formatter = XmlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("&lt;"));
        assert!(result.contains("&gt;"));
        assert!(result.contains("&amp;"));
        assert!(result.contains("&quot;"));
        assert!(!result.contains("path=\"file<"));
        assert!(!result.contains("name=\"func<"));
    }

    #[test]
    fn test_xml_visibility() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![
                Node::Field(Field {
                    name: "_private".to_string(),
                    visibility: Visibility::Private,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("str")),
                    default_value: None,
                    line: 1,
                }),
                Node::Field(Field {
                    name: "public".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    field_type: Some(TypeRef::new("int")),
                    default_value: None,
                    line: 2,
                }),
            ],
        };

        let formatter = XmlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("visibility=\"private\""));
        assert!(result.contains("visibility=\"public\""));
    }

    #[test]
    fn test_xml_multiple_files() {
        let files = vec![
            File {
                path: "file1.py".to_string(),
                children: vec![Node::Function(Function {
                    name: "func1".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: Vec::new(),
                    return_type: None,
                    implementation: None,
                    line_start: 1,
                    line_end: 2,
                })],
            },
            File {
                path: "file2.py".to_string(),
                children: vec![Node::Function(Function {
                    name: "func2".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    parameters: Vec::new(),
                    return_type: None,
                    implementation: None,
                    line_start: 1,
                    line_end: 2,
                })],
            },
        ];

        let formatter = XmlFormatter::new();
        let result = formatter.format_files(&files).unwrap();

        assert!(result.contains("<files>"));
        assert!(result.contains("</files>"));
        assert!(result.contains("<file path=\"file1.py\">"));
        assert!(result.contains("<file path=\"file2.py\">"));
        assert!(result.contains("<function name=\"func1\""));
        assert!(result.contains("<function name=\"func2\""));
    }

    #[test]
    fn test_xml_compact() {
        let file = File {
            path: "test.py".to_string(),
            children: vec![Node::Function(Function {
                name: "hello".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: Vec::new(),
                parameters: Vec::new(),
                return_type: None,
                implementation: None,
                line_start: 1,
                line_end: 2,
            })],
        };

        let options = XmlFormatterOptions {
            indent: false,
            indent_size: 0,
        };
        let formatter = XmlFormatter::with_options(options);
        let result = formatter.format_file(&file).unwrap();

        assert!(!result.contains("  "));
    }

    #[test]
    fn test_xml_type_params() {
        let file = File {
            path: "test.ts".to_string(),
            children: vec![Node::Class(Class {
                name: "Container".to_string(),
                visibility: Visibility::Public,
                modifiers: Vec::new(),
                decorators: Vec::new(),
                type_params: vec![TypeParam {
                    name: "T".to_string(),
                    constraints: Vec::new(),
                    default: None,
                }],
                extends: Vec::new(),
                implements: Vec::new(),
                children: Vec::new(),
                line_start: 1,
                line_end: 3,
            })],
        };

        let formatter = XmlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("<type-params>"));
        assert!(result.contains("<type-param name=\"T\""));
    }
}
