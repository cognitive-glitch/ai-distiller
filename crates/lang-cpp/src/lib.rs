use distiller_core::{
    error::{DistilError, Result},
    ir::*,
    processor::LanguageProcessor,
    ProcessOptions,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;
use tree_sitter::{Node as TSNode, Parser};

pub struct CppProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl CppProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_cpp::LANGUAGE.into())
            .map_err(|e| DistilError::parse_error("cpp", e.to_string()))?;
        Ok(Self {
            parser: Arc::new(Mutex::new(parser)),
        })
    }

    fn node_text(&self, node: TSNode, source: &str) -> String {
        let start = node.start_byte();
        let end = node.end_byte();
        let source_len = source.len();
        if start > end || end > source_len {
            return String::new();
        }
        source[start..end].to_string()
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let implements = Vec::new();
        let visibility = Visibility::Public; // C++ classes are public by default
        let modifiers = Vec::new();
        let mut type_params = Vec::new();
        let decorators = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        // Check if this is a template
        let parent = node.parent();
        if let Some(p) = parent {
            if p.kind() == "template_declaration" {
                type_params = self.parse_template_parameters(p, source);
            }
        }

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "base_class_clause" => {
                    extends = self.parse_base_classes(child, source);
                }
                "field_declaration_list" => {
                    self.parse_class_body(child, source, &mut children)?;
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
            modifiers,
            extends,
            implements,
            type_params,
            decorators,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_template_parameters(&self, node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "template_parameter_list" {
                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    if param_child.kind() == "type_parameter_declaration" {
                        let mut name = String::new();
                        let mut param_text_cursor = param_child.walk();
                        for text_child in param_child.children(&mut param_text_cursor) {
                            if text_child.kind() == "type_identifier" {
                                name = self.node_text(text_child, source);
                            }
                        }
                        if !name.is_empty() {
                            params.push(TypeParam {
                                name,
                                constraints: Vec::new(),
                                default: None,
                            });
                        }
                    }
                }
            }
        }

        params
    }

    fn parse_base_classes(&self, node: TSNode, source: &str) -> Vec<TypeRef> {
        let mut bases = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_identifier" {
                bases.push(TypeRef::new(self.node_text(child, source)));
            }
        }

        bases
    }

    fn parse_class_body(&self, node: TSNode, source: &str, children: &mut Vec<Node>) -> Result<()> {
        let mut current_visibility = Visibility::Private; // C++ default is private
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "access_specifier" => {
                    current_visibility = self.parse_access_specifier(child, source);
                }
                "function_definition" => {
                    if let Some(mut func) = self.parse_function(child, source)? {
                        func.visibility = current_visibility;
                        children.push(Node::Function(func));
                    }
                }
                "field_declaration" => {
                    if let Some(mut field) = self.parse_field(child, source)? {
                        field.visibility = current_visibility;
                        children.push(Node::Field(field));
                    }
                }
                "class_specifier" => {
                    // Nested class
                    if let Some(class) = self.parse_class(child, source)? {
                        children.push(Node::Class(class));
                    }
                }
                _ => {}
            }
        }

        Ok(())
    }

    fn parse_access_specifier(&self, node: TSNode, source: &str) -> Visibility {
        let text = self.node_text(node, source);
        match text.as_str() {
            "public" => Visibility::Public,
            "protected" => Visibility::Protected,
            "private" => Visibility::Private,
            _ => Visibility::Private,
        }
    }

    fn parse_function(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let visibility = Visibility::Public; // Will be overridden by caller
        let mut modifiers = Vec::new();
        let type_params = Vec::new();
        let decorators = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        // Parse return type (first child might be type)
        let mut cursor = node.walk();
        let mut first = true;
        for child in node.children(&mut cursor) {
            match child.kind() {
                "primitive_type" | "type_identifier" if first && return_type.is_none() => {
                    return_type = Some(TypeRef::new(self.node_text(child, source)));
                    first = false;
                }
                "function_declarator" => {
                    first = false;
                    name = self.parse_function_declarator(
                        child,
                        source,
                        &mut parameters,
                        &mut modifiers,
                    );
                }
                "type_qualifier" if self.node_text(child, source) == "virtual" => {
                    modifiers.push(Modifier::Virtual);
                }
                _ => {
                    first = false;
                }
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Function {
            name,
            visibility,
            modifiers,
            parameters,
            return_type,
            type_params,
            decorators,
            line_start,
            line_end,
            implementation: None,
        }))
    }

    fn parse_function_declarator(
        &self,
        node: TSNode,
        source: &str,
        parameters: &mut Vec<Parameter>,
        modifiers: &mut Vec<Modifier>,
    ) -> String {
        let mut name = String::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_identifier" | "identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "destructor_name" => {
                    name = self.node_text(child, source);
                }
                "parameter_list" => {
                    *parameters = self.parse_parameters(child, source);
                }
                "type_qualifier" => {
                    let text = self.node_text(child, source);
                    if text == "const" {
                        modifiers.push(Modifier::Const);
                    }
                }
                "virtual_specifier" => {
                    let text = self.node_text(child, source);
                    if text == "override" {
                        modifiers.push(Modifier::Override);
                    } else if text == "final" {
                        modifiers.push(Modifier::Final);
                    }
                }
                _ => {}
            }
        }

        name
    }

    fn parse_parameters(&self, node: TSNode, source: &str) -> Vec<Parameter> {
        let mut parameters = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "parameter_declaration"
                || child.kind() == "optional_parameter_declaration"
            {
                let mut param_type = TypeRef::new("unknown".to_string());
                let mut name = String::new();

                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    match param_child.kind() {
                        "primitive_type" | "type_identifier" => {
                            param_type = TypeRef::new(self.node_text(param_child, source));
                        }
                        "identifier" => {
                            name = self.node_text(param_child, source);
                        }
                        _ => {}
                    }
                }

                if name.is_empty() {
                    // C++ allows unnamed parameters
                    name = format!("param_{}", parameters.len());
                }

                parameters.push(Parameter {
                    name,
                    param_type,
                    default_value: None,
                    is_variadic: false,
                    is_optional: false,
                    decorators: Vec::new(),
                });
            }
        }

        parameters
    }

    fn parse_field(&self, node: TSNode, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;
        let visibility = Visibility::Private; // Will be overridden by caller
        let modifiers = Vec::new();
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "primitive_type" | "type_identifier" => {
                    if field_type.is_none() {
                        field_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                "field_identifier" => {
                    name = self.node_text(child, source);
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
            field_type,
            default_value: None,
            modifiers,
            line,
        }))
    }

    fn parse_namespace(&self, node: TSNode, source: &str, file: &mut File) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "declaration_list" {
                self.process_node(child, source, file)?;
            }
        }
        Ok(())
    }

    fn process_node(&self, node: TSNode, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "class_specifier" | "struct_specifier" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "template_declaration" => {
                // Check if it's a template class or function
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    if child.kind() == "class_specifier" || child.kind() == "struct_specifier" {
                        if let Some(class) = self.parse_class(child, source)? {
                            file.children.push(Node::Class(class));
                        }
                    } else if child.kind() == "function_definition" {
                        if let Some(func) = self.parse_function(child, source)? {
                            file.children.push(Node::Function(func));
                        }
                    }
                }
            }
            "function_definition" => {
                if let Some(func) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(func));
                }
            }
            "namespace_definition" => {
                self.parse_namespace(node, source, file)?;
            }
            "preproc_include" => {
                // Parse includes as imports
                let text = self.node_text(node, source);
                if let Some(import) = self.parse_include(text) {
                    file.children.push(Node::Import(import));
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

    fn parse_include(&self, text: String) -> Option<Import> {
        // Extract module from #include <module> or #include "module"
        let text = text.trim();
        if !text.starts_with("#include") {
            return None;
        }

        let module = if let Some(start) = text.find('<') {
            if let Some(end) = text.find('>') {
                text[start + 1..end].to_string()
            } else {
                return None;
            }
        } else if let Some(start) = text.find('"') {
            if let Some(end) = text.rfind('"') {
                if end > start {
                    text[start + 1..end].to_string()
                } else {
                    return None;
                }
            } else {
                return None;
            }
        } else {
            return None;
        };

        Some(Import {
            import_type: "include".to_string(),
            module,
            symbols: Vec::new(),
            is_type: false,
            line: None,
        })
    }
}

impl LanguageProcessor for CppProcessor {
    fn language(&self) -> &'static str {
        "cpp"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["cpp", "cc", "cxx", "hpp", "h", "hxx"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| self.supported_extensions().contains(&ext))
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser
            .parse(source, None)
            .ok_or_else(|| DistilError::parse_error("cpp", "Failed to parse source"))?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: Vec::new(),
        };

        self.process_node(tree.root_node(), source, &mut file)?;

        Ok(file)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = CppProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = CppProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("test.cpp")));
        assert!(processor.can_process(Path::new("test.hpp")));
        assert!(processor.can_process(Path::new("test.h")));
        assert!(!processor.can_process(Path::new("test.java")));
    }

    #[test]
    fn test_basic_class_parsing() {
        let source = r#"
class Point {
public:
    Point() : x_(0), y_(0) {}
    double getX() const { return x_; }
private:
    double x_;
    double y_;
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Point.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_point = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Point"
            } else {
                false
            }
        });
        assert!(has_point, "Expected Point class");
    }

    #[test]
    fn test_template_class_parsing() {
        let source = r#"
template<typename T>
class Container {
public:
    explicit Container(const T& value) : value_(value) {}
    const T& getValue() const { return value_; }
private:
    T value_;
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Container.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_template = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Container" && !class.type_params.is_empty()
            } else {
                false
            }
        });
        assert!(has_template, "Expected Container template class");
    }

    #[test]
    fn test_inheritance() {
        let source = r#"
class Point3D : public Point {
public:
    Point3D(double x, double y, double z) : Point(x, y), z_(z) {}
    double distanceFromOrigin() const override;
private:
    double z_;
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Point3D.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_inheritance = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Point3D" && !class.extends.is_empty()
            } else {
                false
            }
        });
        assert!(has_inheritance, "Expected Point3D with inheritance");
    }

    #[test]
    fn test_namespace_parsing() {
        let source = r#"
namespace MathUtils {
    template<typename T>
    T max(const T& a, const T& b) {
        return (a > b) ? a : b;
    }
}
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("MathUtils.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_function = file.children.iter().any(|child| {
            if let Node::Function(func) = child {
                func.name == "max"
            } else {
                false
            }
        });
        assert!(has_function, "Expected max function in namespace");
    }

    #[test]
    fn test_include_parsing() {
        let source = r#"
#include <iostream>
#include <string>
#include "myheader.h"
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("test.cpp"), &opts)
            .unwrap();

        let import_count = file
            .children
            .iter()
            .filter(|child| matches!(child, Node::Import(_)))
            .count();

        assert!(import_count >= 2, "Expected at least 2 includes");
    }

    #[test]
    fn test_virtual_functions() {
        let source = r#"
class Base {
public:
    virtual void process() = 0;
    virtual ~Base() = default;
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Base.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_base = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Base"
            } else {
                false
            }
        });
        assert!(has_base, "Expected Base class");
    }

    #[test]
    fn test_const_methods() {
        let source = r#"
class Point {
public:
    double getX() const { return x_; }
    void setX(double x) { x_ = x; }
private:
    double x_;
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Point.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_const_method = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Point"
                    && class.children.iter().any(|c| {
                        if let Node::Function(func) = c {
                            func.name == "getX" && func.modifiers.contains(&Modifier::Const)
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_const_method, "Expected const method getX");
    }

    #[test]
    fn test_override_modifier() {
        let source = r#"
class Derived : public Base {
public:
    void process() override { }
    void compute() final { }
};
"#;
        let processor = CppProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Derived.cpp"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_override = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Derived"
                    && class.children.iter().any(|c| {
                        if let Node::Function(func) = c {
                            func.name == "process" && func.modifiers.contains(&Modifier::Override)
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_override, "Expected override modifier on process method");
    }
}
