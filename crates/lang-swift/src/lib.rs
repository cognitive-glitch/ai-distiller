use distiller_core::{
    error::{DistilError, Result},
    ir::{self, *},
    processor::LanguageProcessor,
    ProcessOptions,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;
use tree_sitter::{Node as TSNode, Parser};

pub struct SwiftProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl SwiftProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_swift::LANGUAGE.into())
            .map_err(|e| DistilError::parse_error("swift", e.to_string()))?;

        Ok(Self {
            parser: Arc::new(Mutex::new(parser)),
        })
    }

    fn node_text(&self, node: TSNode, source: &str) -> String {
        if node.start_byte() > node.end_byte() || node.end_byte() > source.len() {
            return String::new();
        }
        source[node.start_byte()..node.end_byte()].to_string()
    }

    fn parse_visibility(&self, node: TSNode, source: &str) -> Visibility {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "modifiers" {
                let text = self.node_text(child, source);
                // Check for various visibility modifiers
                if text.contains("private") {
                    return Visibility::Private;
                } else if text.contains("fileprivate") {
                    return Visibility::Private;
                } else if text.contains("public") || text.contains("open") {
                    return Visibility::Public;
                } else if text.contains("internal") {
                    return Visibility::Internal;
                }
            }
        }
        // Swift defaults to internal visibility
        Visibility::Internal
    }

    fn parse_type_parameters(&self, node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut type_params = Vec::new();
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameters" {
                let mut tp_cursor = child.walk();
                for tp_child in child.children(&mut tp_cursor) {
                    if tp_child.kind() == "type_parameter" {
                        let mut name = String::new();
                        let mut constraints = Vec::new();

                        let mut param_cursor = tp_child.walk();
                        for param_child in tp_child.children(&mut param_cursor) {
                            match param_child.kind() {
                                "type_identifier" => {
                                    if name.is_empty() {
                                        name = self.node_text(param_child, source);
                                    }
                                }
                                "type_constraint" => {
                                    let constraint_text = self.node_text(param_child, source);
                                    if !constraint_text.is_empty() {
                                        constraints.push(TypeRef::new(constraint_text));
                                    }
                                }
                                _ => {}
                            }
                        }

                        if !name.is_empty() {
                            type_params.push(TypeParam {
                                name,
                                constraints,
                                default: None,
                            });
                        }
                    }
                }
            }
        }
        type_params
    }

    fn parse_protocol(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut children = Vec::new();
        let mut extends = Vec::new();
        let visibility = self.parse_visibility(node, source);
        let type_params = self.parse_type_parameters(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "type_inheritance_clause" => {
                    self.parse_type_inheritance(child, source, &mut extends)?;
                }
                "protocol_body" => {
                    self.parse_protocol_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        Ok(Some(Class {
            name,
            visibility,
            extends,
            implements: vec![],
            type_params,
            decorators: vec!["protocol".to_string()],
            modifiers: vec![],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_protocol_body(&self, node: TSNode, source: &str, children: &mut Vec<ir::Node>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "protocol_function_declaration" | "protocol_property_declaration" => {
                    if let Some(method) = self.parse_protocol_requirement(child, source)? {
                        children.push(ir::Node::Function(method));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }

    fn parse_protocol_requirement(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let visibility = Visibility::Public; // Protocol requirements are always public

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "parameter_clause" => {
                    self.parse_parameters(child, source, &mut parameters)?;
                }
                "type_annotation" => {
                    return_type = self.parse_type_annotation(child, source);
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            Ok(Some(Function {
                name,
                visibility,
                parameters,
                return_type,
                decorators: vec![],
                modifiers: vec![],
                type_params: vec![],
                implementation: None,
                line_start,
                line_end,
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_struct(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut children = Vec::new();
        let visibility = self.parse_visibility(node, source);
        let type_params = self.parse_type_parameters(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "type_inheritance_clause" => {
                    self.parse_type_inheritance(child, source, &mut extends)?;
                }
                "struct_body" => {
                    self.parse_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        Ok(Some(Class {
            name,
            visibility,
            extends: vec![],
            implements: extends, // In Swift, structs implement protocols
            type_params,
            decorators: vec!["struct".to_string()],
            modifiers: vec![],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut children = Vec::new();
        let mut modifiers = vec![];
        let visibility = self.parse_visibility(node, source);
        let type_params = self.parse_type_parameters(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        // Check for 'open' modifier
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "modifiers" {
                let text = self.node_text(child, source);
                if text.contains("open") {
                    modifiers.push(Modifier::Final); // Use existing modifier, open is similar to non-final
                }
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
                "type_inheritance_clause" => {
                    self.parse_type_inheritance(child, source, &mut extends)?;
                }
                "class_body" => {
                    self.parse_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        // First item in extends is the superclass, rest are protocols
        let (extends, implements) = if !extends.is_empty() {
            (vec![extends[0].clone()], extends[1..].to_vec())
        } else {
            (vec![], vec![])
        };

        Ok(Some(Class {
            name,
            visibility,
            extends,
            implements,
            type_params,
            decorators: vec![],
            modifiers,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_enum(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut children = Vec::new();
        let visibility = self.parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "type_inheritance_clause" => {
                    self.parse_type_inheritance(child, source, &mut extends)?;
                }
                "enum_class_body" => {
                    self.parse_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        Ok(Some(Class {
            name,
            visibility,
            extends: vec![],
            implements: extends,
            type_params: vec![],
            decorators: vec!["enum".to_string()],
            modifiers: vec![],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_type_inheritance(&self, node: TSNode, source: &str, extends: &mut Vec<TypeRef>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_identifier" || child.kind() == "user_type" {
                let type_name = self.node_text(child, source);
                if !type_name.is_empty() && type_name != ":" {
                    extends.push(TypeRef::new(type_name));
                }
            }
        }
        Ok(())
    }

    fn parse_function(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let mut return_type = None;
        let visibility = self.parse_visibility(node, source);
        let type_params = self.parse_type_parameters(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "parameter_clause" => {
                    self.parse_parameters(child, source, &mut parameters)?;
                }
                "type_annotation" => {
                    return_type = self.parse_type_annotation(child, source);
                }
                _ => {}
            }
        }

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type,
            decorators: vec![],
            modifiers: vec![],
            type_params,
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_parameters(&self, node: TSNode, source: &str, params: &mut Vec<Parameter>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "parameter" {
                self.parse_single_parameter(child, source, params)?;
            }
        }
        Ok(())
    }

    fn parse_single_parameter(&self, node: TSNode, source: &str, params: &mut Vec<Parameter>) -> Result<()> {
        let mut name = String::new();
        let mut param_type = TypeRef::new("");
        let mut is_variadic = false;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" => {
                    // Swift can have external and internal parameter names
                    // We'll use the last identifier as the parameter name
                    name = self.node_text(child, source);
                }
                "type_annotation" => {
                    if let Some(type_ref) = self.parse_type_annotation(child, source) {
                        param_type = type_ref;
                    }
                }
                "variadic_parameter" => {
                    is_variadic = true;
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            params.push(Parameter {
                name,
                param_type,
                default_value: None,
                is_variadic,
                is_optional: false,
                decorators: vec![],
            });
        }

        Ok(())
    }

    fn parse_type_annotation(&self, node: TSNode, source: &str) -> Option<TypeRef> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if matches!(child.kind(), "type_identifier" | "user_type" | "optional_type") {
                let type_name = self.node_text(child, source);
                if !type_name.is_empty() && type_name != ":" {
                    return Some(TypeRef::new(type_name));
                }
            }
        }
        None
    }

    fn parse_property(&self, node: TSNode, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;
        let visibility = self.parse_visibility(node, source);
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "pattern_binding" => {
                    let mut binding_cursor = child.walk();
                    for binding_child in child.children(&mut binding_cursor) {
                        match binding_child.kind() {
                            "simple_identifier" => {
                                name = self.node_text(binding_child, source);
                            }
                            "type_annotation" => {
                                field_type = self.parse_type_annotation(binding_child, source);
                            }
                            _ => {}
                        }
                    }
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            Ok(Some(Field {
                name,
                visibility,
                field_type,
                default_value: None,
                modifiers: vec![],
                line,
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_body(&self, node: TSNode, source: &str, children: &mut Vec<ir::Node>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "function_declaration" => {
                    if let Some(func) = self.parse_function(child, source)? {
                        children.push(ir::Node::Function(func));
                    }
                }
                "property_declaration" => {
                    if let Some(field) = self.parse_property(child, source)? {
                        children.push(ir::Node::Field(field));
                    }
                }
                "class_declaration" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        children.push(ir::Node::Class(class));
                    }
                }
                "struct_declaration" => {
                    if let Some(struct_decl) = self.parse_struct(child, source)? {
                        children.push(ir::Node::Class(struct_decl));
                    }
                }
                "enum_declaration" => {
                    if let Some(enum_decl) = self.parse_enum(child, source)? {
                        children.push(ir::Node::Class(enum_decl));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }
}

impl LanguageProcessor for SwiftProcessor {
    fn language(&self) -> &'static str {
        "Swift"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["swift"]
    }

    fn can_process(&self, path: &Path) -> bool {
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| self.supported_extensions().contains(&ext))
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(path.display().to_string(), "Failed to parse Swift source")
        })?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: vec![],
        };

        let root = tree.root_node();
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            match child.kind() {
                "class_declaration" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        file.children.push(ir::Node::Class(class));
                    }
                }
                "struct_declaration" => {
                    if let Some(struct_decl) = self.parse_struct(child, source)? {
                        file.children.push(ir::Node::Class(struct_decl));
                    }
                }
                "enum_declaration" => {
                    if let Some(enum_decl) = self.parse_enum(child, source)? {
                        file.children.push(ir::Node::Class(enum_decl));
                    }
                }
                "protocol_declaration" => {
                    if let Some(protocol) = self.parse_protocol(child, source)? {
                        file.children.push(ir::Node::Class(protocol));
                    }
                }
                "function_declaration" => {
                    if let Some(func) = self.parse_function(child, source)? {
                        file.children.push(ir::Node::Function(func));
                    }
                }
                _ => {}
            }
        }

        Ok(file)
    }
}

impl Default for SwiftProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create Swift processor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = SwiftProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = SwiftProcessor::new().unwrap();
        assert!(processor.can_process(&PathBuf::from("test.swift")));
        assert!(!processor.can_process(&PathBuf::from("test.py")));
        assert!(!processor.can_process(&PathBuf::from("test.rs")));
    }

    #[test]
    fn test_enum_parsing() {
        let source = r#"
enum TemperatureScale: String {
    case celsius = "°C"
    case fahrenheit = "°F"
}
"#;
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(enum_decl) = &file.children[0] {
            assert_eq!(enum_decl.name, "TemperatureScale");
            assert!(enum_decl.decorators.contains(&"enum".to_string()));
            assert_eq!(enum_decl.implements.len(), 1);
            assert_eq!(enum_decl.implements[0].name, "String");
        } else {
            panic!("Expected an enum");
        }
    }

    #[test]
    fn test_struct_with_protocol() {
        let source = r#"
public struct Point: Describable, Equatable {
    public var x: Int
    public var y: Int

    public var magnitude: Double {
        sqrt(Double(x * x + y * y))
    }
}
"#;
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(struct_decl) = &file.children[0] {
            assert_eq!(struct_decl.name, "Point");
            assert!(struct_decl.decorators.contains(&"struct".to_string()));
            assert_eq!(struct_decl.visibility, Visibility::Public);
            assert_eq!(struct_decl.implements.len(), 2);
        } else {
            panic!("Expected a struct");
        }
    }

    #[test]
    fn test_class_with_inheritance() {
        let source = r#"
open class Rectangle: Describable {
    public private(set) var origin: Point

    public init(origin: Point, width: Int, height: Int) {
        self.origin = origin
    }

    open var description: String {
        "Rect@\(origin)"
    }
}
"#;
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Rectangle");
            assert_eq!(class.visibility, Visibility::Public);
            assert_eq!(class.implements.len(), 1);
            assert_eq!(class.implements[0].name, "Describable");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_protocol_definition() {
        let source = r#"
public protocol Describable {
    var description: String { get }
}
"#;
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(protocol) = &file.children[0] {
            assert_eq!(protocol.name, "Describable");
            assert!(protocol.decorators.contains(&"protocol".to_string()));
            assert_eq!(protocol.visibility, Visibility::Public);
        } else {
            panic!("Expected a protocol");
        }
    }

    #[test]
    fn test_generic_struct() {
        let source = r#"
public struct Stack<Element> {
    private var storage: [Element] = []

    public mutating func push(_ element: Element) {
        storage.append(element)
    }

    public func peek() -> Element? {
        storage.last
    }
}
"#;
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(struct_decl) = &file.children[0] {
            assert_eq!(struct_decl.name, "Stack");
            assert!(struct_decl.decorators.contains(&"struct".to_string()));
            assert_eq!(struct_decl.type_params.len(), 1);
            assert_eq!(struct_decl.type_params[0].name, "Element");
        } else {
            panic!("Expected a generic struct");
        }
    }
}

#[cfg(test)]
mod debug_tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn debug_enum_parsing() {
        let source = r#"
enum TemperatureScale: String {
    case celsius = "°C"
}
"#;
        let mut parser = Parser::new();
        parser.set_language(&tree_sitter_swift::LANGUAGE.into()).unwrap();
        let tree = parser.parse(source, None).unwrap();
        
        let root = tree.root_node();
        eprintln!("Root: {}", root.kind());
        
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            eprintln!("  Child: {}", child.kind());
            
            let mut child_cursor = child.walk();
            for grandchild in child.children(&mut child_cursor) {
                eprintln!("    Grandchild: {}", grandchild.kind());
            }
        }
        
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        
        if let Ok(file) = result {
            eprintln!("File children: {}", file.children.len());
            for (i, child) in file.children.iter().enumerate() {
                match child {
                    ir::Node::Class(c) => {
                        eprintln!("  [{}] Class: {}, decorators: {:?}", i, c.name, c.decorators);
                    }
                    _ => {
                        eprintln!("  [{}] Other node", i);
                    }
                }
            }
        }
    }
}

    #[test]
    fn debug_struct_parsing() {
        let source = r#"
public struct Point {
    public var x: Int
}
"#;
        let mut parser = Parser::new();
        parser.set_language(&tree_sitter_swift::LANGUAGE.into()).unwrap();
        let tree = parser.parse(source, None).unwrap();
        
        let root = tree.root_node();
        eprintln!("\n=== STRUCT DEBUG ===");
        eprintln!("Root: {}", root.kind());
        
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            eprintln!("  Child: {}", child.kind());
            
            let mut child_cursor = child.walk();
            for grandchild in child.children(&mut child_cursor) {
                eprintln!("    Grandchild: {}", grandchild.kind());
            }
        }
        
        let processor = SwiftProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.swift"), &opts);
        
        if let Ok(file) = result {
            eprintln!("File children: {}", file.children.len());
            for (i, child) in file.children.iter().enumerate() {
                match child {
                    ir::Node::Class(c) => {
                        eprintln!("  [{}] Class: {}, decorators: {:?}", i, c.name, c.decorators);
                    }
                    _ => {
                        eprintln!("  [{}] Other node", i);
                    }
                }
            }
        }
    }

    #[test]
    fn debug_class_with_protocol() {
        let source = r#"
open class Rectangle: Describable {
    public var x: Int
}
"#;
        let mut parser = Parser::new();
        parser.set_language(&tree_sitter_swift::LANGUAGE.into()).unwrap();
        let tree = parser.parse(source, None).unwrap();
        
        let root = tree.root_node();
        eprintln!("\n=== CLASS WITH PROTOCOL DEBUG ===");
        eprintln!("Root: {}", root.kind());
        
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            eprintln!("  Child: {}", child.kind());
            
            let mut child_cursor = child.walk();
            for grandchild in child.children(&mut child_cursor) {
                eprintln!("    Grandchild: {} | text: {:?}", grandchild.kind(), 
                    &source[grandchild.start_byte()..grandchild.end_byte().min(source.len())]);
            }
        }
    }
}
