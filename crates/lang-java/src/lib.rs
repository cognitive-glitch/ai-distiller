use distiller_core::{
    ProcessOptions,
    error::{DistilError, Result},
    ir::{
        self, Class, Field, File, Function, Import, Modifier, Parameter, TypeParam, TypeRef,
        Visibility,
    },
    processor::LanguageProcessor,
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

    fn node_text(node: TSNode, source: &str) -> String {
        if node.start_byte() > node.end_byte() || node.end_byte() > source.len() {
            return String::new();
        }
        source[node.start_byte()..node.end_byte()].to_string()
    }

    fn parse_modifiers(node: TSNode, source: &str) -> (Visibility, Vec<Modifier>, Vec<String>) {
        let mut visibility = Visibility::Internal; // Java default is package-private
        let mut modifiers = Vec::new();
        let mut decorators = Vec::new();
        let mut has_visibility_keyword = false;
        let mut cursor = node.walk();

        // Find the modifiers child node
        for child in node.children(&mut cursor) {
            if child.kind() == "modifiers" {
                // Iterate through children of the modifiers node
                let mut mod_cursor = child.walk();
                for mod_child in child.children(&mut mod_cursor) {
                    match mod_child.kind() {
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
                        "marker_annotation" | "annotation" => {
                            decorators.push(Self::node_text(mod_child, source));
                        }
                        _ => {}
                    }
                }
                break;
            }
        }

        // If no visibility keyword, use Internal (package-private)
        if !has_visibility_keyword {
            visibility = Visibility::Internal;
        }

        (visibility, modifiers, decorators)
    }

    fn parse_type_parameters(node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter" {
                // Get first type_identifier child for name
                let mut type_cursor = child.walk();
                let mut name = String::new();
                for type_child in child.children(&mut type_cursor) {
                    if type_child.kind() == "type_identifier" {
                        name = Self::node_text(type_child, source);
                        break;
                    }
                }
                if !name.is_empty() {
                    let mut constraints = Vec::new();

                    if let Some(bound_node) = child.child_by_field_name("bound") {
                        constraints.push(TypeRef::new(Self::node_text(bound_node, source)));
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

    fn parse_interface_list(node: TSNode, source: &str) -> Vec<TypeRef> {
        let mut interfaces = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_list" {
                // type_list contains the actual type nodes
                let mut type_cursor = child.walk();
                for type_child in child.children(&mut type_cursor) {
                    if type_child.kind() == "type_identifier" || type_child.kind() == "generic_type"
                    {
                        interfaces.push(TypeRef::new(Self::node_text(type_child, source)));
                    }
                }
            } else if child.kind() == "type_identifier" || child.kind() == "generic_type" {
                // Direct type nodes (for extends_interfaces)
                interfaces.push(TypeRef::new(Self::node_text(child, source)));
            }
        }

        interfaces
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut implements = Vec::new();
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
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
                    name = Self::node_text(child, source);
                }
                "type_parameters" => {
                    type_params = Self::parse_type_parameters(child, source);
                }
                "superclass" => {
                    if let Some(type_node) = child.child_by_field_name("type") {
                        extends.push(TypeRef::new(Self::node_text(type_node, source)));
                    }
                }
                "super_interfaces" => {
                    implements = Self::parse_interface_list(child, source);
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
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
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
                    name = Self::node_text(child, source);
                }
                "type_parameters" => {
                    type_params = Self::parse_type_parameters(child, source);
                }
                "extends_interfaces" => {
                    extends = Self::parse_interface_list(child, source);
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
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
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
                    name = Self::node_text(child, source);
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

    fn parse_enum(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
        let mut children = Vec::new();
        let mut enum_constants = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {}
                "identifier" => {
                    name = Self::node_text(child, source);
                }
                "enum_body" => {
                    let mut body_cursor = child.walk();
                    for body_child in child.children(&mut body_cursor) {
                        match body_child.kind() {
                            "enum_constant" => {
                                let mut const_cursor = body_child.walk();
                                for const_child in body_child.children(&mut const_cursor) {
                                    if const_child.kind() == "identifier" {
                                        let const_name = Self::node_text(const_child, source);
                                        enum_constants.push(const_name);
                                        break;
                                    }
                                }
                            }
                            "enum_body_declarations" => {
                                let mut decl_cursor = body_child.walk();
                                for decl_child in body_child.children(&mut decl_cursor) {
                                    match decl_child.kind() {
                                        "field_declaration" => {
                                            let fields = Self::parse_field(decl_child, source)?;
                                            for field in fields {
                                                children.push(ir::Node::Field(field));
                                            }
                                        }
                                        "constructor_declaration" => {
                                            if let Some(constructor) =
                                                self.parse_constructor(decl_child, source)?
                                            {
                                                children.push(ir::Node::Function(constructor));
                                            }
                                        }
                                        "method_declaration" => {
                                            if let Some(method) =
                                                self.parse_method(decl_child, source)?
                                            {
                                                children.push(ir::Node::Function(method));
                                            }
                                        }
                                        _ => {}
                                    }
                                }
                            }
                            _ => {}
                        }
                    }
                }
                _ => {}
            }
        }

        for const_name in enum_constants {
            children.insert(
                0,
                ir::Node::Field(Field {
                    name: const_name,
                    visibility: Visibility::Public,
                    field_type: Some(TypeRef::new(name.clone())),
                    default_value: None,
                    modifiers: vec![Modifier::Static, Modifier::Final],
                    line: line_start,
                }),
            );
        }

        Ok(Some(Class {
            name,
            visibility,
            extends: vec![],
            implements: vec![],
            type_params: vec![],
            decorators: vec!["enum".to_string()],
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
                    let fields = Self::parse_field(child, source)?;
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
                "enum_declaration" => {
                    if let Some(nested_enum) = self.parse_enum(child, source)? {
                        children.push(ir::Node::Class(nested_enum));
                    }
                }
                _ => {}
            }
        }

        Ok(())
    }

    #[allow(clippy::unused_self)]
    fn parse_annotation_body(
        &self,
        node: TSNode,
        source: &str,
        children: &mut Vec<ir::Node>,
    ) -> Result<()> {
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "annotation_type_element_declaration"
                && let Some(method) = Self::parse_annotation_element(child, source)?
            {
                children.push(ir::Node::Function(method));
            }
        }

        Ok(())
    }

    fn parse_annotation_element(node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "type" | "integral_type" | "floating_point_type" | "boolean_type" => {
                    return_type = Some(TypeRef::new(Self::node_text(child, source)));
                }
                "identifier" => {
                    name = Self::node_text(child, source);
                }
                _ => {}
            }
        }

        if name.is_empty() {
            Ok(None)
        } else {
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
        }
    }

    fn parse_field(node: TSNode, source: &str) -> Result<Vec<Field>> {
        let mut fields = Vec::new();
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
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
                    field_type = Some(TypeRef::new(Self::node_text(child, source)));
                }
                "variable_declarator" => {
                    if let Some(name_node) = child.child_by_field_name("name") {
                        let name = Self::node_text(name_node, source);
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

    #[allow(clippy::unused_self)]
    fn parse_method(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let (visibility, modifiers, method_decorators) = Self::parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let mut decorators = method_decorators; // Start with annotations from modifiers
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifiers" => {
                    // Already parsed
                }
                "type_parameters" => {
                    type_params = Self::parse_type_parameters(child, source);
                }
                "type"
                | "void_type"
                | "integral_type"
                | "floating_point_type"
                | "boolean_type"
                | "generic_type"
                | "type_identifier"
                | "array_type" => {
                    return_type = Some(TypeRef::new(Self::node_text(child, source)));
                }
                "identifier" => {
                    name = Self::node_text(child, source);
                }
                "formal_parameters" => {
                    parameters = Self::parse_parameters(child, source)?;
                }
                "marker_annotation" | "annotation" => {
                    // This case probably never executes since annotations are in modifiers
                    decorators.push(Self::node_text(child, source));
                }
                _ => {}
            }
        }

        if name.is_empty() {
            Ok(None)
        } else {
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
        }
    }

    #[allow(clippy::unused_self)]
    fn parse_constructor(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let (visibility, modifiers, _) = Self::parse_modifiers(node, source);
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
                    name = Self::node_text(child, source);
                }
                "formal_parameters" => {
                    parameters = Self::parse_parameters(child, source)?;
                }
                _ => {}
            }
        }

        if name.is_empty() {
            Ok(None)
        } else {
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
        }
    }

    fn parse_parameters(node: TSNode, source: &str) -> Result<Vec<Parameter>> {
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
                            param_type = TypeRef::new(Self::node_text(param_child, source));
                        }
                        "identifier" => {
                            name = Self::node_text(param_child, source);
                        }
                        "variable_declarator" => {
                            // For spread_parameter, identifier is inside variable_declarator
                            if let Some(id_node) = param_child.child_by_field_name("name") {
                                name = Self::node_text(id_node, source);
                            } else {
                                // Fallback: find first identifier child
                                let mut var_cursor = param_child.walk();
                                for var_child in param_child.children(&mut var_cursor) {
                                    if var_child.kind() == "identifier" {
                                        name = Self::node_text(var_child, source);
                                        break;
                                    }
                                }
                            }
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
                        let module = Self::node_text(import_node, source);
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
                "enum_declaration" => {
                    if let Some(enum_decl) = self.parse_enum(child, source)? {
                        file.children.push(ir::Node::Class(enum_decl));
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

    // ===== Enhanced Test Coverage (11 new tests) =====

    #[test]
    fn test_empty_file() {
        let source = "";
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 0, "Empty file should have no children");
    }

    #[test]
    fn test_interface_with_default_methods() {
        let source = r#"
interface Logger {
    void log(String message);

    default void debug(String message) {
        log("DEBUG: " + message);
    }

    default void error(String message) {
        log("ERROR: " + message);
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(interface) = &file.children[0] {
            assert_eq!(interface.name, "Logger");
            assert!(interface.decorators.contains(&"interface".to_string()));

            let methods: Vec<_> = interface
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(
                methods.len(),
                3,
                "Expected 3 methods (1 abstract + 2 default)"
            );
            assert_eq!(methods[0].name, "log");
            assert_eq!(methods[1].name, "debug");
            assert_eq!(methods[2].name, "error");
        } else {
            panic!("Expected interface node");
        }
    }

    #[test]
    fn test_abstract_class() {
        let source = r#"
abstract class Shape {
    protected String color;

    public abstract double area();
    public abstract double perimeter();

    public void setColor(String color) {
        this.color = color;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Shape");
            assert!(
                class.modifiers.contains(&Modifier::Abstract),
                "Expected abstract modifier on class"
            );

            let abstract_methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .filter(|f| f.modifiers.contains(&Modifier::Abstract))
                .collect();

            assert_eq!(abstract_methods.len(), 2, "Expected 2 abstract methods");
            assert_eq!(abstract_methods[0].name, "area");
            assert_eq!(abstract_methods[1].name, "perimeter");
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_generic_class() {
        let source = r#"
public class Box<T> {
    private T content;

    public void set(T content) {
        this.content = content;
    }

    public T get() {
        return content;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Box");
            assert_eq!(class.type_params.len(), 1, "Expected one type parameter");
            assert_eq!(class.type_params[0].name, "T");

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

            assert_eq!(fields.len(), 1);
            assert_eq!(fields[0].name, "content");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(methods.len(), 2);
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_annotations() {
        let source = r#"
@Deprecated
@SuppressWarnings("unchecked")
public class LegacyService {
    @Override
    @Deprecated
    public String toString() {
        return "LegacyService";
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "LegacyService");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(methods.len(), 1);
            assert_eq!(methods[0].name, "toString");
            assert!(
                !methods[0].decorators.is_empty(),
                "Expected method to have annotations"
            );
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_inner_class() {
        let source = r#"
public class Outer {
    private int outerField;

    public class Inner {
        private int innerField;

        public void innerMethod() {
            outerField = 10;
        }
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(outer) = &file.children[0] {
            assert_eq!(outer.name, "Outer");

            let nested_classes: Vec<_> = outer
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Class(c) = n {
                        Some(c)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(nested_classes.len(), 1, "Expected one nested class");
            assert_eq!(nested_classes[0].name, "Inner");
            assert_eq!(nested_classes[0].visibility, Visibility::Public);
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_static_inner_class() {
        let source = r#"
public class Outer {
    private static int staticField;

    public static class StaticInner {
        private int value;

        public void method() {
            staticField = 42;
        }
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(outer) = &file.children[0] {
            assert_eq!(outer.name, "Outer");

            let nested_classes: Vec<_> = outer
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Class(c) = n {
                        Some(c)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(nested_classes.len(), 1, "Expected one static nested class");
            assert_eq!(nested_classes[0].name, "StaticInner");
            // Note: tree-sitter-java may not always detect 'static' on nested classes
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_enum_with_methods() {
        let source = r#"
public enum Status {
    ACTIVE, INACTIVE, PENDING;

    public boolean isActive() {
        return this == ACTIVE;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        // Note: tree-sitter-java parses enums as enum_declaration, not class_declaration
        // This test validates that we don't crash on enum parsing
        assert!(!file.children.is_empty(), "Expected enum to be parsed");
    }

    #[test]
    fn test_varargs_method() {
        let source = r#"
public class Util {
    public static int sum(int... numbers) {
        int total = 0;
        for (int n : numbers) {
            total += n;
        }
        return total;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Util");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(methods.len(), 1);
            assert_eq!(methods[0].name, "sum");
            assert_eq!(methods[0].parameters.len(), 1);
            assert!(
                methods[0].parameters[0].is_variadic,
                "Expected parameter to be variadic"
            );
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_synchronized_method() {
        let source = r#"
public class Counter {
    private int count = 0;

    public synchronized void increment() {
        count++;
    }

    public synchronized int getCount() {
        return count;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Counter");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(methods.len(), 2);
            assert_eq!(methods[0].name, "increment");
            assert_eq!(methods[1].name, "getCount");
            // Note: synchronized is not currently tracked in modifiers but parsed correctly
        } else {
            panic!("Expected class node");
        }
    }

    #[test]
    fn test_final_class() {
        let source = r#"
public final class Constants {
    public static final int MAX_VALUE = 100;
    public static final String APP_NAME = "MyApp";

    public final String getName() {
        return APP_NAME;
    }
}
"#;
        let processor = JavaProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.java"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Constants");
            assert!(
                class.modifiers.contains(&Modifier::Final),
                "Expected final modifier on class"
            );

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

            assert_eq!(fields.len(), 2);
            assert!(
                fields[0].modifiers.contains(&Modifier::Final),
                "Expected final modifier on field"
            );

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| {
                    if let ir::Node::Function(f) = n {
                        Some(f)
                    } else {
                        None
                    }
                })
                .collect();

            assert_eq!(methods.len(), 1);
            assert!(
                methods[0].modifiers.contains(&Modifier::Final),
                "Expected final modifier on method"
            );
        } else {
            panic!("Expected class node");
        }
    }
}
#[cfg(test)]
mod debug_tests {
    use super::*;
    use tree_sitter::Node as TSNode;

    fn print_tree(node: TSNode, source: &str, depth: usize) {
        let indent = "  ".repeat(depth);
        let kind = node.kind();

        let start = node.start_byte();
        let end = node.end_byte();
        let text = if end > start && end <= source.len() {
            &source[start..end]
        } else {
            ""
        };

        let text_preview = if text.len() > 60 {
            format!("{}...", &text[..60].replace('\n', "\\n"))
        } else {
            text.replace('\n', "\\n")
        };

        eprintln!("{}[{}] \"{}\"", indent, kind, text_preview);

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            print_tree(child, source, depth + 1);
        }
    }

    #[test]
    #[ignore]
    fn debug_enum_ast() {
        let source = r#"public enum Status {
    ACTIVE, INACTIVE, PENDING;

    private final String label;

    Status(String label) {
        this.label = label;
    }

    public String getLabel() {
        return label;
    }

    public boolean isActive() {
        return this == ACTIVE;
    }
}"#;

        let processor = JavaProcessor::new().unwrap();
        let mut parser = processor.parser.lock();
        let tree = parser.parse(source, None).unwrap();
        let root = tree.root_node();

        eprintln!("\n=== Java Enum AST Structure ===\n");
        print_tree(root, source, 0);

        panic!("Debug output - check stderr");
    }
    #[test]
    #[ignore]
    fn debug_varargs_ast() {
        let source = r#"public class Util {
    public static int sum(int... numbers) {
        int total = 0;
        for (int n : numbers) {
            total += n;
        }
        return total;
    }
}"#;

        let processor = JavaProcessor::new().unwrap();
        let mut parser = processor.parser.lock();
        let tree = parser.parse(source, None).unwrap();
        let root = tree.root_node();

        eprintln!("\n=== Java Varargs AST Structure ===\n");
        print_tree(root, source, 0);

        panic!("Debug output - check stderr");
    }

    #[test]
    #[ignore]
    fn debug_annotations_ast() {
        let source = r#"@Deprecated
@SuppressWarnings("unchecked")
public class LegacyService {
    @Override
    @Deprecated
    public String toString() {
        return "LegacyService";
    }
}"#;

        let processor = JavaProcessor::new().unwrap();
        let mut parser = processor.parser.lock();
        let tree = parser.parse(source, None).unwrap();
        let root = tree.root_node();

        eprintln!("\n=== Java Annotations AST Structure ===\n");
        print_tree(root, source, 0);

        panic!("Debug output - check stderr");
    }
}
