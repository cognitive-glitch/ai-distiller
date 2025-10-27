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

pub struct PhpProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl PhpProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_php::LANGUAGE_PHP.into())
            .map_err(|e| DistilError::parse_error("php", e.to_string()))?;
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
        let visibility = Visibility::Public; // PHP classes are public
        let modifiers = Vec::new();
        let type_params = Vec::new();
        let decorators = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "name" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "base_clause" => {
                    extends = self.parse_base_clause(child, source);
                }
                "declaration_list" => {
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

    fn parse_trait(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let extends = Vec::new();
        let implements = Vec::new();
        let visibility = Visibility::Public;
        let modifiers = Vec::new();
        let type_params = Vec::new();
        let decorators = vec!["trait".to_string()];
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "name" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "declaration_list" => {
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

    fn parse_base_clause(&self, node: TSNode, source: &str) -> Vec<TypeRef> {
        let mut bases = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "name" {
                bases.push(TypeRef::new(self.node_text(child, source)));
            }
        }

        bases
    }

    fn parse_class_body(&self, node: TSNode, source: &str, children: &mut Vec<Node>) -> Result<()> {
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "method_declaration" => {
                    if let Some(method) = self.parse_method(child, source)? {
                        children.push(Node::Function(method));
                    }
                }
                "property_declaration" => {
                    if let Some(property) = self.parse_property(child, source)? {
                        children.push(Node::Field(property));
                    }
                }
                _ => {}
            }
        }

        Ok(())
    }

    fn parse_method(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let mut visibility = Visibility::Public;
        let modifiers = Vec::new();
        let type_params = Vec::new();
        let decorators = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "visibility_modifier" => {
                    visibility = self.parse_visibility(child, source);
                }
                "name" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source);
                }
                "primitive_type" | "named_type" | "optional_type" => {
                    if return_type.is_none() {
                        return_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
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
            parameters,
            return_type,
            type_params,
            decorators,
            line_start,
            line_end,
            implementation: None,
        }))
    }

    fn parse_visibility(&self, node: TSNode, source: &str) -> Visibility {
        let text = self.node_text(node, source);
        match text.as_str() {
            "public" => Visibility::Public,
            "protected" => Visibility::Protected,
            "private" => Visibility::Private,
            _ => Visibility::Public,
        }
    }

    fn parse_parameters(&self, node: TSNode, source: &str) -> Vec<Parameter> {
        let mut parameters = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "simple_parameter" || child.kind() == "variadic_parameter" {
                let mut param_type = TypeRef::new("mixed".to_string());
                let mut name = String::new();

                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    match param_child.kind() {
                        "primitive_type" | "named_type" | "optional_type" => {
                            param_type = TypeRef::new(self.node_text(param_child, source));
                        }
                        "variable_name" => {
                            name = self.node_text(param_child, source);
                        }
                        _ => {}
                    }
                }

                if !name.is_empty() {
                    parameters.push(Parameter {
                        name,
                        param_type,
                        default_value: None,
                        is_variadic: child.kind() == "variadic_parameter",
                        is_optional: false,
                        decorators: Vec::new(),
                    });
                }
            }
        }

        parameters
    }

    fn parse_property(&self, node: TSNode, source: &str) -> Result<Option<Field>> {
        let mut name = String::new();
        let mut field_type = None;
        let mut visibility = Visibility::Public;
        let modifiers = Vec::new();
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "visibility_modifier" => {
                    visibility = self.parse_visibility(child, source);
                }
                "primitive_type" | "named_type" | "optional_type" => {
                    if field_type.is_none() {
                        field_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
                }
                "property_element" => {
                    let mut elem_cursor = child.walk();
                    for elem_child in child.children(&mut elem_cursor) {
                        if elem_child.kind() == "variable_name" {
                            name = self.node_text(elem_child, source);
                        }
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
            visibility,
            field_type,
            default_value: None,
            modifiers,
            line,
        }))
    }

    fn parse_use(&self, node: TSNode, source: &str) -> Option<Import> {
        let mut module = String::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "namespace_use_clause" {
                module = self.node_text(child, source);
            }
        }

        if module.is_empty() {
            return None;
        }

        Some(Import {
            import_type: "use".to_string(),
            module,
            symbols: Vec::new(),
            is_type: false,
            line: Some(node.start_position().row + 1),
        })
    }

    fn process_node(&self, node: TSNode, source: &str, file: &mut File) -> Result<()> {
        match node.kind() {
            "class_declaration" => {
                if let Some(class) = self.parse_class(node, source)? {
                    file.children.push(Node::Class(class));
                }
            }
            "trait_declaration" => {
                if let Some(trait_class) = self.parse_trait(node, source)? {
                    file.children.push(Node::Class(trait_class));
                }
            }
            "namespace_use_declaration" => {
                if let Some(import) = self.parse_use(node, source) {
                    file.children.push(Node::Import(import));
                }
            }
            "function_definition" => {
                // Top-level functions
                if let Some(func) = self.parse_top_level_function(node, source)? {
                    file.children.push(Node::Function(func));
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

    fn parse_top_level_function(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let visibility = Visibility::Public;
        let modifiers = Vec::new();
        let type_params = Vec::new();
        let decorators = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "name" => {
                    if name.is_empty() {
                        name = self.node_text(child, source);
                    }
                }
                "formal_parameters" => {
                    parameters = self.parse_parameters(child, source);
                }
                "primitive_type" | "named_type" | "optional_type" => {
                    if return_type.is_none() {
                        return_type = Some(TypeRef::new(self.node_text(child, source)));
                    }
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
            parameters,
            return_type,
            type_params,
            decorators,
            line_start,
            line_end,
            implementation: None,
        }))
    }
}

impl LanguageProcessor for PhpProcessor {
    fn language(&self) -> &'static str {
        "php"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["php"]
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
            .ok_or_else(|| DistilError::parse_error("php", "Failed to parse source"))?;

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
        let processor = PhpProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = PhpProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("test.php")));
        assert!(!processor.can_process(Path::new("test.py")));
    }

    #[test]
    fn test_basic_class_parsing() {
        let source = r#"<?php
class User {
    public int $id;
    private string $email;

    public function __construct(int $id, string $name) {
        $this->id = $id;
    }

    public function getEmail(): string {
        return $this->email;
    }
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_user = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "User"
            } else {
                false
            }
        });
        assert!(has_user, "Expected User class");
    }

    #[test]
    fn test_trait_parsing() {
        let source = r#"<?php
trait Timestampable {
    protected ?DateTime $createdAt;

    public function getCreatedAt(): ?DateTime {
        return $this->createdAt;
    }
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Timestampable.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_trait = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Timestampable" && class.decorators.contains(&"trait".to_string())
            } else {
                false
            }
        });
        assert!(has_trait, "Expected Timestampable trait");
    }

    #[test]
    fn test_namespace_and_use() {
        let source = r#"<?php
namespace App\Basic;

use DateTime;
use InvalidArgumentException;

class User {
    public int $id;
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.php"), &opts)
            .unwrap();

        let import_count = file
            .children
            .iter()
            .filter(|child| matches!(child, Node::Import(_)))
            .count();

        assert!(import_count >= 2, "Expected at least 2 use statements");
    }

    #[test]
    fn test_typed_properties() {
        let source = r#"<?php
class User {
    public int $id;
    private string $name;
    protected ?DateTime $createdAt;
    public array $preferences = [];
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_typed_props = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "User"
                    && class.children.iter().any(|c| {
                        if let Node::Field(field) = c {
                            field.name == "$id" && field.field_type.is_some()
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_typed_props, "Expected typed properties");
    }

    #[test]
    fn test_visibility_modifiers() {
        let source = r#"<?php
class Test {
    public function publicMethod() {}
    protected function protectedMethod() {}
    private function privateMethod() {}
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Test.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_visibility = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "Test"
                    && class.children.iter().any(|c| {
                        if let Node::Function(func) = c {
                            matches!(
                                func.visibility,
                                Visibility::Public | Visibility::Protected | Visibility::Private
                            )
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_visibility, "Expected visibility modifiers");
    }

    #[test]
    fn test_return_types() {
        let source = r#"<?php
class User {
    public function getEmail(): string {}
    public function getId(): int {}
    public function getCreatedAt(): ?DateTime {}
    public function getPreferences(): array {}
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_return_types = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "User"
                    && class.children.iter().any(|c| {
                        if let Node::Function(func) = c {
                            func.name == "getEmail" && func.return_type.is_some()
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_return_types, "Expected return types on methods");
    }

    #[test]
    fn test_constructor() {
        let source = r#"<?php
class User {
    public function __construct(int $id, string $name) {
        $this->id = $id;
    }
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.php"), &opts)
            .unwrap();

        assert!(!file.children.is_empty());
        let has_constructor = file.children.iter().any(|child| {
            if let Node::Class(class) = child {
                class.name == "User"
                    && class.children.iter().any(|c| {
                        if let Node::Function(func) = c {
                            func.name == "__construct"
                        } else {
                            false
                        }
                    })
            } else {
                false
            }
        });
        assert!(has_constructor, "Expected __construct method");
    }

    #[test]
    fn test_top_level_function() {
        let source = r#"<?php
function validateEmail(string $email): bool {
    return filter_var($email, FILTER_VALIDATE_EMAIL) !== false;
}
"#;
        let processor = PhpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("functions.php"), &opts)
            .unwrap();

        let has_function = file.children.iter().any(|child| {
            if let Node::Function(func) = child {
                func.name == "validateEmail"
            } else {
                false
            }
        });
        assert!(has_function, "Expected validateEmail function");
    }
}
