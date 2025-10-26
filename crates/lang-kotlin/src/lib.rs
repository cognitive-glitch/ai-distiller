use distiller_core::{
    error::{DistilError, Result},
    ir::*,
    processor::LanguageProcessor,
    ProcessOptions,
};
use parking_lot::Mutex;
use std::path::PathBuf;
use std::sync::Arc;
use tree_sitter::{Node as TSNode, Parser};

pub struct KotlinProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl KotlinProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_kotlin::language())
            .map_err(|e| DistilError::parse_error("kotlin", e.to_string()))?;
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

    fn parse_modifiers(&self, node: TSNode, source: &str) -> (Visibility, Vec<Modifier>) {
        let mut visibility = Visibility::Public; // Kotlin default
        let mut modifiers = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            let text = self.node_text(child, source);
            match text.as_str() {
                "public" => visibility = Visibility::Public,
                "private" => visibility = Visibility::Private,
                "protected" => visibility = Visibility::Protected,
                "internal" => visibility = Visibility::Internal,
                "abstract" => modifiers.push(Modifier::Abstract),
                "open" => modifiers.push(Modifier::Virtual),
                "final" => modifiers.push(Modifier::Final),
                "override" => modifiers.push(Modifier::Override),
                "suspend" => modifiers.push(Modifier::Async),
                "inline" => modifiers.push(Modifier::Inline),
                "data" => modifiers.push(Modifier::Data),
                "sealed" => modifiers.push(Modifier::Sealed),
                "const" => modifiers.push(Modifier::Const),
                _ => {}
            }
        }

        (visibility, modifiers)
    }

    fn parse_type_parameters(&self, node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter" {
                let mut name = String::new();
                let mut constraints = Vec::new();

                let mut tp_cursor = child.walk();
                for tp_child in child.children(&mut tp_cursor) {
                    match tp_child.kind() {
                        "simple_identifier" | "type_identifier" => {
                            if name.is_empty() {
                                name = self.node_text(tp_child, source);
                            }
                        }
                        "type_constraint" | "user_type" => {
                            constraints.push(self.node_text(tp_child, source));
                        }
                        _ => {}
                    }
                }

                if !name.is_empty() {
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

    fn parse_type_constraints(&self, node: TSNode, source: &str, type_params: &mut [TypeParam]) {
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "type_constraint" {
                let mut param_name = String::new();
                let mut constraint = String::new();

                let mut tc_cursor = child.walk();
                for tc_child in child.children(&mut tc_cursor) {
                    match tc_child.kind() {
                        "simple_identifier" if param_name.is_empty() => {
                            param_name = self.node_text(tc_child, source);
                        }
                        "user_type" | "type_identifier" => {
                            constraint = self.node_text(tc_child, source);
                        }
                        _ => {}
                    }
                }

                if !param_name.is_empty() && !constraint.is_empty() {
                    if let Some(param) = type_params.iter_mut().find(|p| p.name == param_name) {
                        param.constraints.push(constraint);
                    }
                }
            }
        }
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut type_params = Vec::new();
        let extends = Vec::new();
        let mut implements = Vec::new();
        let mut decorators = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);

        // Determine class type
        let class_kind = node.kind();
        match class_kind {
            "class_declaration" => {}
            "object_declaration" => decorators.push("object".to_string()),
            _ => {}
        }

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" | "type_identifier" if name.is_empty() => {
                    name = self.node_text(child, source);
                }
                "type_parameters" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "type_constraints" => {
                    self.parse_type_constraints(child, source, &mut type_params);
                }
                "delegation_specifier" => {
                    let delegation_text = self.node_text(child, source);
                    if !delegation_text.is_empty() {
                        // Simple approach: everything after : goes to implements
                        implements.push(TypeRef::new(delegation_text));
                    }
                }
                "primary_constructor" => {
                    // Process constructor parameters (for data classes)
                    if let Some(func) = self.parse_constructor(child, source, &name)? {
                        return Ok(Some(Class {
                            name,
                            type_params,
                            extends,
                            implements,
                            decorators,
                            children: vec![Node::Function(func)],
                            visibility,
                            modifiers,
                                                        line_end: node.end_position().row + 1,
                        }));
                    }
                }
                "class_body" => {
                    let children = self.parse_class_body(child, source)?;
                    return Ok(Some(Class {
                        name,
                        type_params,
                        extends,
                        implements,
                        decorators,
                        children,
                        visibility,
                        modifiers,
                                                line_end: node.end_position().row + 1,
                    }));
                }
                _ => {}
            }
        }

        if !name.is_empty() {
            Ok(Some(Class {
                name,
                type_params,
                extends,
                implements,
                decorators,
                children: Vec::new(),
                visibility,
                modifiers,
                                line_end: node.end_position().row + 1,
            }))
        } else {
            Ok(None)
        }
    }

    fn parse_class_body(&self, node: TSNode, source: &str) -> Result<Vec<Node>> {
        let mut children = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "function_declaration" => {
                    if let Some(func) = self.parse_function(child, source)? {
                        children.push(Node::Function(func));
                    }
                }
                "property_declaration" => {
                    if let Some(field) = self.parse_property(child, source)? {
                        children.push(Node::Field(field));
                    }
                }
                "class_declaration" | "object_declaration" => {
                    if let Some(inner_class) = self.parse_class(child, source)? {
                        children.push(Node::Class(inner_class));
                    }
                }
                "companion_object" => {
                    if let Some(companion) = self.parse_companion_object(child, source)? {
                        children.push(Node::Class(companion));
                    }
                }
                _ => {}
            }
        }

        Ok(children)
    }

    fn parse_constructor(&self, node: TSNode, source: &str, class_name: &str) -> Result<Option<Function>> {
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let parameters = self.parse_parameters(node, source);

        Ok(Some(Function {
            name: class_name.to_string(),
            parameters,
            return_type: None,
            decorators: vec!["constructor".to_string()],
            children: Vec::new(),
            visibility,
            modifiers,
                        line_end: node.end_position().row + 1,
        }))
    }

    fn parse_function(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let parameters;
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let mut decorators = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" if name.is_empty() => {
                    name = self.node_text(child, source);
                }
                "user_type" | "type_identifier" => {
                    if return_type.is_none() {
                        return_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                _ => {}
            }
        }

        parameters = self.parse_parameters(node, source);

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Function {
            name,
            parameters,
            return_type,
            decorators,
            children: Vec::new(),
            visibility,
            modifiers,
                        line_end: node.end_position().row + 1,
        }))
    }

    fn parse_parameters(&self, node: TSNode, source: &str) -> Vec<Parameter> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "function_value_parameters" || child.kind() == "class_parameter" {
                let mut param_cursor = child.walk();
                for param_node in child.children(&mut param_cursor) {
                    if param_node.kind() == "parameter" || param_node.kind() == "class_parameter" {
                        if let Some(param) = self.parse_single_parameter(param_node, source) {
                            params.push(param);
                        }
                    }
                }
            }
        }

        params
    }

    fn parse_single_parameter(&self, node: TSNode, source: &str) -> Option<Parameter> {
        let mut name = String::new();
        let mut param_type = TypeRef::new("unknown".to_string());
        let mut decorators = Vec::new();

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "simple_identifier" if name.is_empty() => {
                    name = self.node_text(child, source);
                }
                "user_type" | "type_identifier" => {
                    param_type = TypeRef::new(self.node_text(child, source));
                }
                "vararg" => decorators.push("vararg".to_string()),
                _ => {}
            }
        }

        if !name.is_empty() {
            Some(Parameter {
                name,
                param_type,
                default_value: None,
                is_variadic: decorators.contains(&"vararg".to_string()),
                is_optional: false,
                decorators,
            })
        } else {
            None
        }
    }

    fn parse_property(&self, node: TSNode, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;
        let (visibility, modifiers) = self.parse_modifiers(node, source);

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "variable_declaration" => {
                    let mut var_cursor = child.walk();
                    for var_child in child.children(&mut var_cursor) {
                        if var_child.kind() == "simple_identifier" && name.is_empty() {
                            name = self.node_text(var_child, source);
                        }
                    }
                }
                "user_type" | "type_identifier" => {
                    if field_type.is_none() {
                        field_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                _ => {}
            }
        }

        if name.is_empty() {
            return Ok(None);
        }

        Ok(Some(Field {
            name,
            field_type,
            decorators: Vec::new(),
            visibility,
            modifiers,
                        line_end: node.end_position().row + 1,
        }))
    }

    fn parse_companion_object(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut children = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "class_body" {
                children = self.parse_class_body(child, source)?;
            }
        }

        Ok(Some(Class {
            name: "companion object".to_string(),
            type_params: Vec::new(),
            extends: Vec::new(),
            implements: Vec::new(),
            decorators: vec!["companion".to_string()],
            children,
            visibility: Visibility::Public,
            modifiers: vec![Modifier::Static],
                        line_end: node.end_position().row + 1,
        }))
    }

    fn parse_import(&self, node: TSNode, source: &str) -> Option<Import> {
        let import_text = self.node_text(node, source);
        let text = import_text.strip_prefix("import ")?.trim();

        Some(Import {
            module: text.to_string(),
            symbols: Vec::new(),
            import_type: "import".to_string(),
                        line_end: node.end_position().row + 1,
        })
    }

    fn process_node(&self, node: TSNode, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "import_header" => {
                if let Some(import) = self.parse_import(node, source) {
                    file.children.push(Node::Import(import));
                }
            }
            "class_declaration" | "object_declaration" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "function_declaration" => {
                if let Some(func) = self.parse_function(node, source)? {
                    file.children.push(Node::Function(func));
                }
            }
            "property_declaration" => {
                if let Some(field) = self.parse_property(node, source)? {
                    file.children.push(Node::Field(field));
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
}

impl LanguageProcessor for KotlinProcessor {
    fn language(&self) -> &str {
        "kotlin"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["kt", "kts"]
    }

    fn can_process(&self, filename: &str) -> bool {
        filename.ends_with(".kt") || filename.ends_with(".kts")
    }

    fn process(
        &self,
        source: &str,
        path: &Path,
        _opts: &ProcessOptions,
    ) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser
            .parse(source, None)
            .ok_or_else(|| DistilError::parse_error("kotlin", "Failed to parse source"))?;

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

    #[test]
    fn test_processor_creation() {
        let processor = KotlinProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = KotlinProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("Test.kt")));
        assert!(processor.can_process(Path::new("script.kts")));
        assert!(!processor.can_process(Path::new("Test.java")));
    }

    #[test]
    fn test_data_class_parsing() {
        let source = r#"
data class User(
    val id: Long,
    val name: String,
    val email: String?
)
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "User");
            assert!(class.modifiers.contains(&Modifier::Data));
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_sealed_class_parsing() {
        let source = r#"
sealed class UserState {
    data class Active(val lastLogin: Long) : UserState()
    object Banned : UserState()
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("UserState.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "UserState");
            assert!(class.modifiers.contains(&Modifier::Sealed));
            assert!(class.children.len() >= 1);
        } else {
            panic!("Expected a sealed class");
        }
    }

    #[test]
    fn test_extension_function() {
        let source = r#"
fun String.isValidEmail(): Boolean {
    return contains("@") && contains(".")
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Extensions.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Function(func) = &file.children[0] {
            assert!(func.name.contains("isValidEmail"));
        } else {
            panic!("Expected a function");
        }
    }

    #[test]
    fn test_companion_object() {
        let source = r#"
class User {
    companion object {
        fun createUser(name: String): User {
            return User()
        }
    }
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "User");
            let has_companion = class.children.iter().any(|child| {
                if let Node::Class(c) = child {
                    c.decorators.contains(&"companion".to_string())
                } else {
                    false
                }
            });
            assert!(has_companion);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_generic_class() {
        let source = r#"
class Repository<T : Entity> where T : Comparable<T> {
    fun save(entity: T): T {
        return entity
    }
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Repository.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Repository");
            assert!(class.type_params.len() >= 1);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_visibility_modifiers() {
        let source = r#"
class Test {
    public val publicField: Int = 1
    private val privateField: Int = 2
    protected val protectedField: Int = 3
    internal val internalField: Int = 4

    public fun publicMethod() {}
    private fun privateMethod() {}
    protected fun protectedMethod() {}
    internal fun internalMethod() {}
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Test.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Test");
            assert!(class.children.len() >= 4);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_suspend_function() {
        let source = r#"
suspend fun fetchUser(id: Long): User? {
    return null
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Api.kt"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Function(func) = &file.children[0] {
            assert_eq!(func.name, "fetchUser");
            assert!(func.modifiers.contains(&Modifier::Async));
        } else {
            panic!("Expected a function");
        }
    }
}
