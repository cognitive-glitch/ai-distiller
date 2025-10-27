//! XML formatter for AI Distiller
//!
//! Structured XML output for legacy systems and XML-based tooling.
//! Provides proper XML escaping and semantic structure.

use distiller_core::ir::*;
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
    pub fn new() -> Self {
        Self {
            options: XmlFormatterOptions::default(),
        }
    }

    /// Create a new XML formatter with custom options
    pub fn with_options(options: XmlFormatterOptions) -> Self {
        Self { options }
    }

    /// Format a single file as XML
    pub fn format_file(&self, file: &File) -> Result<String, std::fmt::Error> {
        let mut output = String::new();
        writeln!(output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")?;
        self.format_file_element(&mut output, file, 0)?;
        Ok(output)
    }

    /// Format multiple files as XML
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
    fn format_file_element(&self, output: &mut String, file: &File, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{}<file path=\"{}\">", ind, escape_xml(&file.path))?;

        for child in &file.children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{}</file>", ind)?;
        Ok(())
    }

    /// Format a node
    fn format_node(&self, output: &mut String, node: &Node, indent: usize) -> Result<(), std::fmt::Error> {
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
    fn format_directory(&self, output: &mut String, dir: &Directory, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{}<directory path=\"{}\">", ind, escape_xml(&dir.path))?;
        for child in &dir.children {
            self.format_node(output, child, indent + 1)?;
        }
        writeln!(output, "{}</directory>", ind)?;
        Ok(())
    }

    /// Format a package
    fn format_package(&self, output: &mut String, pkg: &Package, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{}<package name=\"{}\">", ind, escape_xml(&pkg.name))?;
        for child in &pkg.children {
            self.format_node(output, child, indent + 1)?;
        }
        writeln!(output, "{}</package>", ind)?;
        Ok(())
    }

    /// Format an import
    fn format_import(&self, output: &mut String, import: &Import, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<import", ind)?;
        write!(output, " type=\"{}\"", import.import_type)?;
        write!(output, " module=\"{}\"", escape_xml(&import.module))?;
        if let Some(line) = import.line {
            write!(output, " line=\"{}\"", line)?;
        }

        if import.symbols.is_empty() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            for symbol in &import.symbols {
                let symbol_ind = self.indent(indent + 1);
                write!(output, "{}<symbol name=\"{}\"", symbol_ind, escape_xml(&symbol.name))?;
                if let Some(ref alias) = symbol.alias {
                    write!(output, " alias=\"{}\"", escape_xml(alias))?;
                }
                writeln!(output, " />")?;
            }
            writeln!(output, "{}</import>", ind)?;
        }
        Ok(())
    }

    /// Format a class
    fn format_class(&self, output: &mut String, class: &Class, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        for decorator in &class.decorators {
            writeln!(output, "{}<decorator value=\"{}\" />", ind, escape_xml(decorator))?;
        }

        write!(output, "{}<class", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&class.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(class.visibility))?;
        write!(output, " line-start=\"{}\"", class.line_start)?;
        write!(output, " line-end=\"{}\"", class.line_end)?;
        if !class.modifiers.is_empty() {
            write!(output, " modifiers=\"{}\"", escape_xml(&modifiers_to_string(&class.modifiers)))?;
        }
        writeln!(output, ">")?;

        if !class.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-params>", type_params_ind)?;
            for param in &class.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{}</type-params>", type_params_ind)?;
        }

        if !class.extends.is_empty() {
            let extends_ind = self.indent(indent + 1);
            writeln!(output, "{}<extends>", extends_ind)?;
            for type_ref in &class.extends {
                self.format_type_ref(output, type_ref, indent + 2)?;
            }
            writeln!(output, "{}</extends>", extends_ind)?;
        }

        if !class.implements.is_empty() {
            let implements_ind = self.indent(indent + 1);
            writeln!(output, "{}<implements>", implements_ind)?;
            for type_ref in &class.implements {
                self.format_type_ref(output, type_ref, indent + 2)?;
            }
            writeln!(output, "{}</implements>", implements_ind)?;
        }

        for child in &class.children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{}</class>", ind)?;
        Ok(())
    }

    /// Format an interface
    fn format_interface(&self, output: &mut String, interface: &Interface, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<interface", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&interface.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(interface.visibility))?;
        write!(output, " line-start=\"{}\"", interface.line_start)?;
        write!(output, " line-end=\"{}\"", interface.line_end)?;
        writeln!(output, ">")?;

        if !interface.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-params>", type_params_ind)?;
            for param in &interface.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{}</type-params>", type_params_ind)?;
        }

        if !interface.extends.is_empty() {
            let extends_ind = self.indent(indent + 1);
            writeln!(output, "{}<extends>", extends_ind)?;
            for type_ref in &interface.extends {
                self.format_type_ref(output, type_ref, indent + 2)?;
            }
            writeln!(output, "{}</extends>", extends_ind)?;
        }

        for child in &interface.children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{}</interface>", ind)?;
        Ok(())
    }

    /// Format a struct
    fn format_struct(&self, output: &mut String, struct_node: &Struct, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<struct", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&struct_node.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(struct_node.visibility))?;
        write!(output, " line-start=\"{}\"", struct_node.line_start)?;
        write!(output, " line-end=\"{}\"", struct_node.line_end)?;
        writeln!(output, ">")?;

        if !struct_node.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-params>", type_params_ind)?;
            for param in &struct_node.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{}</type-params>", type_params_ind)?;
        }

        for child in &struct_node.children {
            self.format_node(output, child, indent + 1)?;
        }

        writeln!(output, "{}</struct>", ind)?;
        Ok(())
    }

    /// Format an enum
    fn format_enum(&self, output: &mut String, enum_node: &Enum, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<enum", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&enum_node.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(enum_node.visibility))?;
        write!(output, " line-start=\"{}\"", enum_node.line_start)?;
        write!(output, " line-end=\"{}\"", enum_node.line_end)?;

        if let Some(ref enum_type) = enum_node.enum_type {
            writeln!(output, ">")?;
            let type_ind = self.indent(indent + 1);
            writeln!(output, "{}<type>", type_ind)?;
            self.format_type_ref(output, enum_type, indent + 2)?;
            writeln!(output, "{}</type>", type_ind)?;
            for child in &enum_node.children {
                self.format_node(output, child, indent + 1)?;
            }
            writeln!(output, "{}</enum>", ind)?;
        } else {
            if enum_node.children.is_empty() {
                writeln!(output, " />")?;
            } else {
                writeln!(output, ">")?;
                for child in &enum_node.children {
                    self.format_node(output, child, indent + 1)?;
                }
                writeln!(output, "{}</enum>", ind)?;
            }
        }
        Ok(())
    }

    /// Format a type alias
    fn format_type_alias(&self, output: &mut String, type_alias: &TypeAlias, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<type-alias", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&type_alias.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(type_alias.visibility))?;
        write!(output, " line=\"{}\"", type_alias.line)?;
        writeln!(output, ">")?;

        if !type_alias.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-params>", type_params_ind)?;
            for param in &type_alias.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{}</type-params>", type_params_ind)?;
        }

        let alias_ind = self.indent(indent + 1);
        writeln!(output, "{}<alias-type>", alias_ind)?;
        self.format_type_ref(output, &type_alias.alias_type, indent + 2)?;
        writeln!(output, "{}</alias-type>", alias_ind)?;

        writeln!(output, "{}</type-alias>", ind)?;
        Ok(())
    }

    /// Format a function
    fn format_function(&self, output: &mut String, function: &Function, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        for decorator in &function.decorators {
            writeln!(output, "{}<decorator value=\"{}\" />", ind, escape_xml(decorator))?;
        }

        write!(output, "{}<function", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&function.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(function.visibility))?;
        write!(output, " line-start=\"{}\"", function.line_start)?;
        write!(output, " line-end=\"{}\"", function.line_end)?;
        if !function.modifiers.is_empty() {
            write!(output, " modifiers=\"{}\"", escape_xml(&modifiers_to_string(&function.modifiers)))?;
        }
        writeln!(output, ">")?;

        if !function.type_params.is_empty() {
            let type_params_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-params>", type_params_ind)?;
            for param in &function.type_params {
                self.format_type_param(output, param, indent + 2)?;
            }
            writeln!(output, "{}</type-params>", type_params_ind)?;
        }

        if !function.parameters.is_empty() {
            let params_ind = self.indent(indent + 1);
            writeln!(output, "{}<parameters>", params_ind)?;
            for param in &function.parameters {
                self.format_parameter(output, param, indent + 2)?;
            }
            writeln!(output, "{}</parameters>", params_ind)?;
        }

        if let Some(ref return_type) = function.return_type {
            let return_ind = self.indent(indent + 1);
            writeln!(output, "{}<return-type>", return_ind)?;
            self.format_type_ref(output, return_type, indent + 2)?;
            writeln!(output, "{}</return-type>", return_ind)?;
        }

        if let Some(ref impl_body) = function.implementation {
            let impl_ind = self.indent(indent + 1);
            writeln!(output, "{}<implementation>", impl_ind)?;
            writeln!(output, "{}", escape_xml(impl_body))?;
            writeln!(output, "{}</implementation>", impl_ind)?;
        }

        writeln!(output, "{}</function>", ind)?;
        Ok(())
    }

    /// Format a field
    fn format_field(&self, output: &mut String, field: &Field, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<field", ind)?;
        write!(output, " name=\"{}\"", escape_xml(&field.name))?;
        write!(output, " visibility=\"{}\"", visibility_str(field.visibility))?;
        write!(output, " line=\"{}\"", field.line)?;
        if !field.modifiers.is_empty() {
            write!(output, " modifiers=\"{}\"", escape_xml(&modifiers_to_string(&field.modifiers)))?;
        }

        if let Some(ref field_type) = field.field_type {
            writeln!(output, ">")?;
            let type_ind = self.indent(indent + 1);
            writeln!(output, "{}<type>", type_ind)?;
            self.format_type_ref(output, field_type, indent + 2)?;
            writeln!(output, "{}</type>", type_ind)?;
            if let Some(ref default_value) = field.default_value {
                writeln!(output, "{}<default-value>{}</default-value>", type_ind, escape_xml(default_value))?;
            }
            writeln!(output, "{}</field>", ind)?;
        } else {
            if let Some(ref default_value) = field.default_value {
                write!(output, " default=\"{}\"", escape_xml(default_value))?;
            }
            writeln!(output, " />")?;
        }
        Ok(())
    }

    /// Format a comment
    fn format_comment(&self, output: &mut String, comment: &Comment, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<comment", ind)?;
        write!(output, " format=\"{}\"", comment.format)?;
        write!(output, " line=\"{}\"", comment.line)?;
        writeln!(output, ">")?;
        writeln!(output, "{}", escape_xml(&comment.text))?;
        writeln!(output, "{}</comment>", ind)?;
        Ok(())
    }

    /// Format raw content
    fn format_raw_content(&self, output: &mut String, raw: &RawContent, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        writeln!(output, "{}<raw-content>", ind)?;
        writeln!(output, "{}", escape_xml(&raw.content))?;
        writeln!(output, "{}</raw-content>", ind)?;
        Ok(())
    }

    /// Format a type parameter
    fn format_type_param(&self, output: &mut String, param: &TypeParam, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<type-param name=\"{}\"", ind, escape_xml(&param.name))?;

        if param.constraints.is_empty() && param.default.is_none() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            if !param.constraints.is_empty() {
                let constraints_ind = self.indent(indent + 1);
                writeln!(output, "{}<constraints>", constraints_ind)?;
                for constraint in &param.constraints {
                    self.format_type_ref(output, constraint, indent + 2)?;
                }
                writeln!(output, "{}</constraints>", constraints_ind)?;
            }
            if let Some(ref default) = param.default {
                let default_ind = self.indent(indent + 1);
                writeln!(output, "{}<default>", default_ind)?;
                self.format_type_ref(output, default, indent + 2)?;
                writeln!(output, "{}</default>", default_ind)?;
            }
            writeln!(output, "{}</type-param>", ind)?;
        }
        Ok(())
    }

    /// Format a type reference
    fn format_type_ref(&self, output: &mut String, type_ref: &TypeRef, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<type name=\"{}\"", ind, escape_xml(&type_ref.name))?;

        if type_ref.type_args.is_empty() {
            writeln!(output, " />")?;
        } else {
            writeln!(output, ">")?;
            let args_ind = self.indent(indent + 1);
            writeln!(output, "{}<type-args>", args_ind)?;
            for arg in &type_ref.type_args {
                self.format_type_ref(output, arg, indent + 2)?;
            }
            writeln!(output, "{}</type-args>", args_ind)?;
            writeln!(output, "{}</type>", ind)?;
        }
        Ok(())
    }

    /// Format a parameter
    fn format_parameter(&self, output: &mut String, param: &Parameter, indent: usize) -> Result<(), std::fmt::Error> {
        let ind = self.indent(indent);
        write!(output, "{}<parameter name=\"{}\"", ind, escape_xml(&param.name))?;
        if param.is_variadic {
            write!(output, " variadic=\"true\"")?;
        }
        if param.is_optional {
            write!(output, " optional=\"true\"")?;
        }
        writeln!(output, ">")?;

        let type_ind = self.indent(indent + 1);
        writeln!(output, "{}<type>", type_ind)?;
        self.format_type_ref(output, &param.param_type, indent + 2)?;
        writeln!(output, "{}</type>", type_ind)?;

        if let Some(ref default_value) = param.default_value {
            writeln!(output, "{}<default-value>{}</default-value>", type_ind, escape_xml(default_value))?;
        }

        writeln!(output, "{}</parameter>", ind)?;
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
    modifiers.iter()
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
            children: vec![
                Node::Class(Class {
                    name: "Example".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: Vec::new(),
                    extends: Vec::new(),
                    implements: Vec::new(),
                    children: vec![
                        Node::Function(Function {
                            name: "__init__".to_string(),
                            visibility: Visibility::Public,
                            modifiers: Vec::new(),
                            decorators: Vec::new(),
                            type_params: Vec::new(),
                            parameters: vec![
                                Parameter {
                                    name: "self".to_string(),
                                    param_type: TypeRef::new("Self"),
                                    default_value: None,
                                    is_variadic: false,
                                    is_optional: false,
                                    decorators: Vec::new(),
                                },
                            ],
                            return_type: None,
                            implementation: None,
                            line_start: 2,
                            line_end: 3,
                        }),
                    ],
                    line_start: 1,
                    line_end: 3,
                }),
            ],
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
            children: vec![
                Node::Function(Function {
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
                }),
            ],
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
                children: vec![
                    Node::Function(Function {
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
                    }),
                ],
            },
            File {
                path: "file2.py".to_string(),
                children: vec![
                    Node::Function(Function {
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
                    }),
                ],
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
            children: vec![
                Node::Function(Function {
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
                }),
            ],
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
            children: vec![
                Node::Class(Class {
                    name: "Container".to_string(),
                    visibility: Visibility::Public,
                    modifiers: Vec::new(),
                    decorators: Vec::new(),
                    type_params: vec![
                        TypeParam {
                            name: "T".to_string(),
                            constraints: Vec::new(),
                            default: None,
                        },
                    ],
                    extends: Vec::new(),
                    implements: Vec::new(),
                    children: Vec::new(),
                    line_start: 1,
                    line_end: 3,
                }),
            ],
        };

        let formatter = XmlFormatter::new();
        let result = formatter.format_file(&file).unwrap();

        assert!(result.contains("<type-params>"));
        assert!(result.contains("<type-param name=\"T\""));
    }
}
