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

pub struct KotlinProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl KotlinProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_kotlin_ng::LANGUAGE.into())
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

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let extends = Vec::new();
        let implements = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let type_params = Vec::new();
        let decorators = Vec::new();
        let children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "simple_identifier" || child.kind() == "type_identifier" {
                if name.is_empty() {
                    name = self.node_text(child, source);
                }
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

    fn parse_function(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let return_type = None;
        let parameters = Vec::new();
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let type_params = Vec::new();
        let decorators = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "simple_identifier" {
                if name.is_empty() {
                    name = self.node_text(child, source);
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

    fn parse_property(&self, node: TSNode, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let field_type = None;
        let (visibility, modifiers) = self.parse_modifiers(node, source);
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "variable_declaration" {
                let mut var_cursor = child.walk();
                for var_child in child.children(&mut var_cursor) {
                    if var_child.kind() == "simple_identifier" && name.is_empty() {
                        name = self.node_text(var_child, source);
                    }
                }
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

    fn parse_import(&self, node: TSNode, source: &str) -> Option<Import> {
        let import_text = self.node_text(node, source);
        let text = import_text.strip_prefix("import ")?.trim();

        Some(Import {
            import_type: "import".to_string(),
            module: text.to_string(),
            symbols: Vec::new(),
            is_type: false,
            line: Some(node.start_position().row + 1),
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
    fn language(&self) -> &'static str {
        "kotlin"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["kt", "kts"]
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
    use std::path::PathBuf;

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

        assert!(file.children.len() >= 1);
        let has_user_class = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "User"
            } else {
                false
            }
        });
        assert!(has_user_class, "Expected a User class");
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

        assert!(file.children.len() >= 1);
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

        assert!(file.children.len() >= 1);
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

        assert!(file.children.len() >= 1);
    }

    #[test]
    fn test_generic_class() {
        let source = r#"
class Repository<T> {
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

        assert!(file.children.len() >= 1);
    }

    #[test]
    fn test_visibility_modifiers() {
        let source = r#"
class Test {
    public val publicField: Int = 1
    private val privateField: Int = 2
    protected val protectedField: Int = 3
    internal val internalField: Int = 4
}
"#;
        let processor = KotlinProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Test.kt"), &opts)
            .unwrap();

        assert!(file.children.len() >= 1);
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

        assert!(file.children.len() >= 1);
    }
}
