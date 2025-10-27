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

pub struct JavaProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl JavaProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_java::LANGUAGE.into())
            .map_err(|e| DistilError::parse_error("java", e.to_string()))?;

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

    fn parse_modifiers(&self, node: TSNode, source: &str) -> (Visibility, Vec<Modifier>) {
        let mut visibility = Visibility::Internal; // Java default is package-private
        let mut modifiers = Vec::new();
        let mut has_visibility_keyword = false;
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            let text = self.node_text(child, source);
            match text.as_str() {
                "public" => {
                    visibility = Visibility::Public;
                    has_visibility_keyword = true;
                }
                "protected" => {
                    visibility = Visibility::Protected;
                    has_visibility_keyword = true;
                }
                "private" => {
                    visibility = Visibility::Private;
                    has_visibility_keyword = true;
                }
                "static" => modifiers.push(Modifier::Static),
                "final" => modifiers.push(Modifier::Final),
                "abstract" => modifiers.push(Modifier::Abstract),
                _ => {}
            }
        }

        // If no visibility keyword, use Internal (package-private)
        if !has_visibility_keyword {
            visibility = Visibility::Internal;
        }

        (visibility, modifiers)
    }

    fn parse_type_parameters(&self, node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter" {
                // Get first type_identifier child for name
                let mut type_cursor = child.walk();
                let mut name = String::new();
                for type_child in child.children(&mut type_cursor) {
                    if type_child.kind() == "type_identifier" {
                        name = self.node_text(type_child, source);
                        break;
                    }
                }
                if !name.is_empty() {
                    let mut constraints = Vec::new();

                    if let Some(bound_node) = child.child_by_field_name("bound") {
                        constraints.push(TypeRef::new(self.node_text(bound_node, source)));
                    }

                    params.push(TypeParam {
                        name,
                        constraints,
                        default: None,
                    });
                }
            }
        }

        params
    }

    fn parse_interface_list(&self, node: TSNode, source: &str) -> Vec<TypeRef> {
        let mut interfaces = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_list" {
                // type_list contains the actual type nodes
                let mut type_cursor = child.walk();
                for type_child in child.children(&mut type_cursor) {
                    if type_child.kind() == "type_identifier" || type_child.kind() == "generic_type"
                    {
                        interfaces.push(TypeRef::new(self.node_text(type_child, source)));
                    }
                }
            } else if child.kind() == "type_identifier" || child.kind() == "generic_type" {
                // Direct type nodes (for extends_interfaces)
                interfaces.push(TypeRef::new(self.node_text(child, source)));
            }
        }

        interfaces
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut implements = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "superclass" => {
                    if let Some(type_node) = child.child_by_field_name("type") {
                        extends.push(TypeRef::new(self.node_text(type_node, source)));
                    }
                }
                "super_interfaces" => {
                    implements = self.parse_interface_list(child, source);
                }
                "class_body" => {
                    self.parse_class_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

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

    fn parse_interface(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "extends_interfaces" => {
                    extends = self.parse_interface_list(child, source);
                }
                "interface_body" => {
                    self.parse_class_body(child, source, &mut children)?;
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
            decorators: vec!["interface".to_string()],
            modifiers,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_annotation(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "annotation_type_body" => {
                    self.parse_annotation_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        Ok(Some(Class {
            name,
            visibility,
            extends: vec![],
            implements: vec![],
            type_params: vec![],
            decorators: vec!["annotation".to_string()],
            modifiers,
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_class_body(
        &self,
        node: TSNode,
        source: &str,
        children: &mut Vec<ir::Node>,
    ) -> Result<()> {
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_declaration" | "constant_declaration" => {
                    let fields = self.parse_field(child, source)?;
                    for field in fields {
                        children.push(ir::Node::Field(field));
                    }
                }
                "method_declaration" => {
                    if let Some(method) = self.parse_method(child, source)? {
                        children.push(ir::Node::Function(method));
                    }
                }
                "constructor_declaration" => {
                    if let Some(constructor) = self.parse_constructor(child, source)? {
                        children.push(ir::Node::Function(constructor));
                    }
                }
                "class_declaration" => {
                    if let Some(nested_class) = self.parse_class(child, source)? {
                        children.push(ir::Node::Class(nested_class));
                    }
                }
                "interface_declaration" => {
                    if let Some(nested_interface) = self.parse_interface(child, source)? {
                        children.push(ir::Node::Class(nested_interface));
                    }
                }
                "annotation_type_declaration" => {
                    if let Some(nested_annotation) = self.parse_annotation(child, source)? {
                        children.push(ir::Node::Class(nested_annotation));
                    }
                }
                _ => {}
            }
        }

        Ok(())
    }

    fn parse_annotation_body(
        &self,
        node: TSNode,
        source: &str,
        children: &mut Vec<ir::Node>,
    ) -> Result<()> {
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "annotation_type_element_declaration" {
                if let Some(method) = self.parse_annotation_element(child, source)? {
                    children.push(ir::Node::Function(method));
                }
            }
        }

        Ok(())
    }

    fn parse_annotation_element(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type" | "integral_type" | "floating_point_type" | "boolean_type" => {
                    return_type = Some(TypeRef::new(self.node_text(child, source)));
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            Ok(Some(Function {
                name,
                visibility: Visibility::Public,
                parameters: vec![],
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

    fn parse_field(&self, node: TSNode, source: &str) -> Result<Vec<Field>> {
        let mut fields = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut field_type = None;
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "type"
                | "integral_type"
                | "floating_point_type"
                | "boolean_type"
                | "void_type"
                | "generic_type"
                | "type_identifier"
                | "array_type" => {
                    field_type = Some(TypeRef::new(self.node_text(child, source)));
                }
                "variable_declarator" => {
                    if let Some(name_node) = child.child_by_field_name("name") {
                        let name = self.node_text(name_node, source);
                        fields.push(Field {
                            name,
                            visibility,
                            field_type: field_type.clone(),
                            default_value: None,
                            modifiers: modifiers.clone(),
                            line,
                        });
                    }
                }
                _ => {}
            }
        }

        Ok(fields)
    }

    fn parse_method(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let mut decorators = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "type"
                | "void_type"
                | "integral_type"
                | "floating_point_type"
                | "boolean_type"
                | "generic_type"
                | "type_identifier"
                | "array_type" => {
                    return_type = Some(TypeRef::new(self.node_text(child, source)));
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                "marker_annotation" | "annotation" => {
                    decorators.push(self.node_text(child, source));
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
                decorators,
                modifiers,
                type_params,
                implementation: None,
                line_start,
                line_end,
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_constructor(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut parameters = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "identifier" => {
                    name = self.node_text(child, source);
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source)?;
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            Ok(Some(Function {
                name,
                visibility,
                parameters,
                return_type: None,
                decorators: vec!["constructor".to_string()],
                modifiers,
                type_params: vec![],
                implementation: None,
                line_start,
                line_end,
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_parameters(&self, node: TSNode, source: &str) -> Result<Vec<Parameter>> {
        let mut parameters = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "formal_parameter" || child.kind() == "spread_parameter" {
                let mut name = String::new();
                let mut param_type = TypeRef::new("");
                let mut is_variadic = false;

                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    match param_child.kind() {
                        "type"
                        | "integral_type"
                        | "floating_point_type"
                        | "boolean_type"
                        | "generic_type"
                        | "type_identifier"
                        | "array_type" => {
                            param_type = TypeRef::new(self.node_text(param_child, source));
                        }
                        "identifier" => {
                            name = self.node_text(param_child, source);
                        }
                        "..." => {
                            is_variadic = true;
                        }
                        _ => {}
                    }
                }

                if !name.is_empty() {
                    parameters.push(Parameter {
                        name,
                        param_type,
                        default_value: None,
                        is_variadic,
                        is_optional: false,
                        decorators: vec![],
                    });
                }
            }
        }

        Ok(parameters)
    }
}

impl LanguageProcessor for JavaProcessor {
    fn language(&self) -> &'static str {
        "Java"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["java"]
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(path.display().to_string(), "Failed to parse Java source")
        })?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: vec![],
        };

        let root = tree.root_node();
        let mut cursor = root.walk();

        for child in root.children(&mut cursor) {
            match child.kind() {
                "package_declaration" => {
                    // Skip package for now
                }
                "import_declaration" => {
                    if let Some(import_node) = child.child_by_field_name("name") {
                        let module = self.node_text(import_node, source);
                        file.children.push(ir::Node::Import(Import {
                            import_type: "import".to_string(),
                            module,
                            symbols: vec![],
                            is_type: false,
                            line: Some(child.start_position().row + 1),
                        }));
                    }
                }
                "class_declaration" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        file.children.push(ir::Node::Class(class));
                    }
                }
                "interface_declaration" => {
                    if let Some(interface) = self.parse_interface(child, source)? {
                        file.children.push(ir::Node::Class(interface));
                    }
                }
                "annotation_type_declaration" => {
                    if let Some(annotation) = self.parse_annotation(child, source)? {
                        file.children.push(ir::Node::Class(annotation));
                    }
                }
                _ => {}
            }
        }

        Ok(file)
    }
}

impl Default for JavaProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create Java processor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = JavaProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = JavaProcessor::new().unwrap();
        assert!(processor.can_process(&PathBuf::from("Test.java")));
        assert!(!processor.can_process(&PathBuf::from("test.py")));
    }

    #[test]
    fn test_basic_class_parsing() {
        let source = r#"
public class Basic {
    private static final String GREETING = "Hello";

    public static void main(String[] args) {
        System.out.println(GREETING);
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Basic.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Basic");
            assert_eq!(class.visibility, Visibility::Public);
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, ir::Node::Field(_)))
                    .count(),
                1
            );
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, ir::Node::Function(_)))
                    .count(),
                1
            );
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_interface_with_generics() {
        let source = r#"
interface DataStore<T> {
    void save(T item);
    T findById(String id);
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("DataStore.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(interface) = &file.children[0] {
            assert_eq!(interface.name, "DataStore");
            assert!(interface.decorators.contains(&"interface".to_string()));
            assert_eq!(interface.type_params.len(), 1);
            assert_eq!(interface.type_params[0].name, "T");
            assert_eq!(
                interface
                    .children
                    .iter()
                    .filter(|n| matches!(n, ir::Node::Function(_)))
                    .count(),
                2
            );
        } else {
            panic!("Expected an interface");
        }
    }

    #[test]
    fn test_class_with_inheritance() {
        let source = r#"
abstract class BaseStore<T> implements DataStore<T> {
    public void save(T item) {}
    protected abstract void log(T item);
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("BaseStore.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "BaseStore");
            assert!(class.modifiers.contains(&Modifier::Abstract));
            assert_eq!(class.implements.len(), 1);
            assert_eq!(class.implements[0].name, "DataStore<T>");
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, ir::Node::Function(_)))
                    .count(),
                2
            );
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_annotation_declaration() {
        let source = r#"
@interface Auditable {
    String value() default "DEFAULT";
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Auditable.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(annotation) = &file.children[0] {
            assert_eq!(annotation.name, "Auditable");
            assert!(annotation.decorators.contains(&"annotation".to_string()));
            assert_eq!(
                annotation
                    .children
                    .iter()
                    .filter(|n| matches!(n, ir::Node::Function(_)))
                    .count(),
                1
            );
        } else {
            panic!("Expected an annotation");
        }
    }

    #[test]
    fn test_visibility_modifiers() {
        let source = r#"
public class Visibility {
    public String publicField;
    protected String protectedField;
    private String privateField;
    String packagePrivateField;
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Visibility.java"), &opts)
            .unwrap();

        if let ir::Node::Class(class) = &file.children[0] {
            let fields: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Field(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(fields.len(), 4);
            assert_eq!(fields[0].visibility, Visibility::Public);
            assert_eq!(fields[1].visibility, Visibility::Protected);
            assert_eq!(fields[2].visibility, Visibility::Private);
            assert_eq!(fields[3].visibility, Visibility::Internal);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_constructor_parsing() {
        let source = r#"
public class SimpleOOP {
    public SimpleOOP(String id, String name) {}
    public SimpleOOP(String id) {}
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("SimpleOOP.java"), &opts)
            .unwrap();

        if let ir::Node::Class(class) = &file.children[0] {
            let constructors: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .filter(|f| f.decorators.contains(&"constructor".to_string()))
                .collect();

            assert_eq!(constructors.len(), 2);
            assert_eq!(constructors[0].parameters.len(), 2);
            assert_eq!(constructors[1].parameters.len(), 1);
        } else {
            panic!("Expected a class");
        }
    }
}
