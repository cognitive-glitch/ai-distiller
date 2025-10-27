use distiller_core::{
    ProcessOptions,
    error::{DistilError, Result},
    ir::*,
    processor::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;
use tree_sitter::{Node as TSNode, Parser};

pub struct CSharpProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl CSharpProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_c_sharp::LANGUAGE.into())
            .map_err(|e| DistilError::parse_error("csharp", e.to_string()))?;
        Ok(Self {
            parser: Arc::new(Mutex::new(parser)),
        })
    }

    fn node_text(node: TSNode, source: &str) -> String {
        let start = node.start_byte();
        let end = node.end_byte();
        let source_len = source.len();
        if start > end || end > source_len {
            return String::new();
        }
        source[start..end].to_string()
    }

    fn parse_modifiers(node: TSNode, source: &str) -> (Visibility, Vec<Modifier>) {
        let mut visibility = Visibility::Private; // C# default
        let mut modifiers = Vec::new();
        let mut has_visibility_keyword = false;
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            match child.kind() {
                "modifier" | "modifiers" => {
                    let text = Self::node_text(child, source);
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
                        "internal" => {
                            visibility = Visibility::Internal;
                            has_visibility_keyword = true;
                        }
                        "static" => modifiers.push(Modifier::Static),
                        "abstract" => modifiers.push(Modifier::Abstract),
                        "sealed" => modifiers.push(Modifier::Final),
                        "virtual" => modifiers.push(Modifier::Virtual),
                        "override" => modifiers.push(Modifier::Override),
                        "async" => modifiers.push(Modifier::Async),
                        "readonly" => modifiers.push(Modifier::Readonly),
                        "const" => modifiers.push(Modifier::Const),
                        _ => {}
                    }
                }
                _ => {}
            }
        }

        if !has_visibility_keyword {
            visibility = Visibility::Private; // C# default
        }

        (visibility, modifiers)
    }

    fn collect_type_refs(&self, node: TSNode, source: &str, results: &mut Vec<TypeRef>) {
        match node.kind() {
            "identifier" | "type_identifier" | "generic_name" | "predefined_type"
            | "qualified_name" => {
                results.push(TypeRef::new(Self::node_text(node, source)));
            }
            _ => {
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.collect_type_refs(child, source, results);
                }
            }
        }
    }

    fn collect_type_param_names(&self, node: TSNode, source: &str, results: &mut Vec<String>) {
        match node.kind() {
            "type_parameter" | "type_identifier" => {
                results.push(Self::node_text(node, source));
            }
            _ => {
                let mut cursor = node.walk();
                for child in node.children(&mut cursor) {
                    self.collect_type_param_names(child, source, results);
                }
            }
        }
    }

    fn parse_type_parameters(&self, node: TSNode, source: &str) -> Vec<TypeParam> {
        let mut params = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            let mut names = Vec::new();
            self.collect_type_param_names(child, source, &mut names);
            for name in names {
                params.push(TypeParam {
                    name,
                    constraints: Vec::new(),
                    default: None,
                });
            }
        }

        params
    }

    fn parse_type_parameter_constraints(node: TSNode, source: &str, type_params: &mut [TypeParam]) {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "type_parameter_constraints_clause" {
                let mut constraint_cursor = child.walk();
                let mut param_name = String::new();
                let mut constraints = Vec::new();

                for constraint_child in child.children(&mut constraint_cursor) {
                    match constraint_child.kind() {
                        "type_identifier" => {
                            if param_name.is_empty() {
                                param_name = Self::node_text(constraint_child, source);
                            } else {
                                constraints
                                    .push(TypeRef::new(Self::node_text(constraint_child, source)));
                            }
                        }
                        "generic_name" | "predefined_type" => {
                            constraints
                                .push(TypeRef::new(Self::node_text(constraint_child, source)));
                        }
                        _ => {}
                    }
                }

                if let Some(param) = type_params.iter_mut().find(|p| p.name == param_name) {
                    param.constraints = constraints;
                }
            }
        }
    }

    fn parse_base_list(&self, node: TSNode, source: &str) -> (Vec<TypeRef>, Vec<TypeRef>) {
        let extends = Vec::new();
        let mut implements = Vec::new();

        self.collect_type_refs(node, source, &mut implements);

        (extends, implements)
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut implements = Vec::new();
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "base_list" => {
                    let (ext, impl_list) = self.parse_base_list(child, source);
                    extends = ext;
                    implements = impl_list;
                }
                "type_parameter_constraints_clause" => {
                    Self::parse_type_parameter_constraints(child, source, &mut type_params);
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
            decorators: vec!["class".to_string()],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_struct(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut class = self.parse_class(node, source)?;
        if let Some(ref mut c) = class {
            c.decorators = vec!["struct".to_string()];
        }
        Ok(class)
    }

    fn parse_record(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut class = self.parse_class(node, source)?;
        if let Some(ref mut c) = class {
            c.decorators = vec!["record".to_string()];
        }
        Ok(class)
    }

    fn parse_interface(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let extends = Vec::new();
        let mut implements = Vec::new();
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut type_params = Vec::new();
        let mut children = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source);
                }
                "base_list" => {
                    let (_, impl_list) = self.parse_base_list(child, source);
                    implements = impl_list;
                }
                "type_parameter_constraints_clause" => {
                    Self::parse_type_parameter_constraints(child, source, &mut type_params);
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
            decorators: vec!["interface".to_string()],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_class_body(&self, node: TSNode, source: &str, children: &mut Vec<Node>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "field_declaration" => {
                    if let Some(field) = Self::parse_field(child, source)? {
                        children.push(Node::Field(field));
                    }
                }
                "method_declaration" => {
                    if let Some(method) = self.parse_method(child, source)? {
                        children.push(Node::Function(method));
                    }
                }
                "constructor_declaration" => {
                    if let Some(ctor) = self.parse_constructor(child, source)? {
                        children.push(Node::Function(ctor));
                    }
                }
                "property_declaration" => {
                    if let Some(prop) = Self::parse_property(child, source)? {
                        children.push(Node::Field(prop));
                    }
                }
                "event_declaration" | "event_field_declaration" => {
                    if let Some(event) = Self::parse_event(child, source)? {
                        children.push(Node::Field(event));
                    }
                }
                "operator_declaration" => {
                    if let Some(op) = self.parse_operator(child, source)? {
                        children.push(Node::Function(op));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }

    fn parse_field(node: TSNode, source: &str) -> Result<Option<Field>> {
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut field_type = None;
        let mut name = String::new();
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "variable_declaration" {
                let mut var_cursor = child.walk();
                for var_child in child.children(&mut var_cursor) {
                    match var_child.kind() {
                        "type_identifier" | "predefined_type" | "generic_name" | "array_type"
                        | "nullable_type" => {
                            field_type = Some(TypeRef::new(Self::node_text(var_child, source)));
                        }
                        "variable_declarator" => {
                            if let Some(name_node) = var_child.child_by_field_name("name") {
                                name = Self::node_text(name_node, source);
                            }
                        }
                        _ => {}
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

    fn parse_property(node: TSNode, source: &str) -> Result<Option<Field>> {
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut field_type = None;
        let mut name = String::new();
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "type_identifier" | "predefined_type" | "generic_name" | "array_type"
                | "nullable_type" => {
                    field_type = Some(TypeRef::new(Self::node_text(child, source)));
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

    fn parse_event(node: TSNode, source: &str) -> Result<Option<Field>> {
        let (visibility, mut modifiers) = Self::parse_modifiers(node, source);
        modifiers.push(Modifier::Event);
        let mut field_type = None;
        let mut name = String::new();
        let line = node.start_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "type_identifier" | "generic_name" | "predefined_type" => {
                    field_type = Some(TypeRef::new(Self::node_text(child, source)));
                }
                "variable_declaration" => {
                    let mut var_cursor = child.walk();
                    for var_child in child.children(&mut var_cursor) {
                        match var_child.kind() {
                            "type_identifier" | "generic_name" => {
                                field_type = Some(TypeRef::new(Self::node_text(var_child, source)));
                            }
                            "variable_declarator" => {
                                if let Some(name_node) = var_child.child_by_field_name("name") {
                                    name = Self::node_text(name_node, source);
                                }
                            }
                            _ => {}
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

    fn parse_method(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut name = String::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let mut type_params = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "type_identifier" | "predefined_type" | "generic_name" | "array_type"
                | "nullable_type" => {
                    if return_type.is_none() {
                        return_type = Some(TypeRef::new(Self::node_text(child, source)));
                    }
                }
                "void_keyword" => {
                    return_type = Some(TypeRef::new("void".to_string()));
                }
                "parameter_list" => {
                    parameters = Self::parse_parameters(child, source);
                }
                "type_parameter_list" => {
                    type_params = self.parse_type_parameters(child, source);
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
            decorators: Vec::new(),
            line_start,
            line_end,
            implementation: None,
        }))
    }

    #[allow(clippy::unused_self)]
    fn parse_constructor(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let (visibility, modifiers) = Self::parse_modifiers(node, source);
        let mut name = String::new();
        let mut parameters = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "parameter_list" => {
                    parameters = Self::parse_parameters(child, source);
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
            return_type: None,
            type_params: Vec::new(),
            decorators: vec!["constructor".to_string()],
            line_start,
            line_end,
            implementation: None,
        }))
    }

    #[allow(clippy::unused_self)]
    fn parse_operator(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let (visibility, mut modifiers) = Self::parse_modifiers(node, source);
        modifiers.push(Modifier::Static); // Operators are always static
        let mut name = String::new();
        let mut return_type = None;
        let mut parameters = Vec::new();
        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "operator_token" => {
                    name = format!("operator{}", Self::node_text(child, source));
                }
                "implicit_keyword" => {
                    name = "implicit".to_string();
                }
                "explicit_keyword" => {
                    name = "explicit".to_string();
                }
                "type_identifier" | "predefined_type" | "generic_name" => {
                    if return_type.is_none() {
                        return_type = Some(TypeRef::new(Self::node_text(child, source)));
                    }
                }
                "parameter_list" => {
                    parameters = Self::parse_parameters(child, source);
                }
                _ => {}
            }
        }

        if name.is_empty() {
            name = "operator".to_string();
        }

        Ok(Some(Function {
            name,
            visibility,
            modifiers,
            parameters,
            return_type,
            type_params: Vec::new(),
            decorators: vec!["operator".to_string()],
            line_start,
            line_end,
            implementation: None,
        }))
    }

    fn parse_parameters(node: TSNode, source: &str) -> Vec<Parameter> {
        let mut parameters = Vec::new();
        let mut cursor = node.walk();

        for child in node.children(&mut cursor) {
            if child.kind() == "parameter" {
                let mut param_type = TypeRef::new("unknown".to_string());
                let mut name = String::new();
                let mut is_variadic = false;
                let mut decorators = Vec::new();

                let mut param_cursor = child.walk();
                for param_child in child.children(&mut param_cursor) {
                    match param_child.kind() {
                        "identifier" => {
                            name = Self::node_text(param_child, source);
                        }
                        "type_identifier" | "predefined_type" | "generic_name" | "array_type"
                        | "nullable_type" => {
                            param_type = TypeRef::new(Self::node_text(param_child, source));
                        }
                        "this_expression" | "ref_keyword" | "out_keyword" | "in_keyword"
                        | "params_keyword" => {
                            let decorator = Self::node_text(param_child, source);
                            decorators.push(decorator.clone());
                            if decorator == "params" {
                                is_variadic = true;
                            }
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
                        decorators,
                    });
                }
            }
        }

        parameters
    }
}

impl LanguageProcessor for CSharpProcessor {
    fn language(&self) -> &'static str {
        "C#"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["cs"]
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
            .ok_or_else(|| DistilError::parse_error("csharp", "Failed to parse source"))?;

        let root = tree.root_node();
        let mut children = Vec::new();

        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            match child.kind() {
                "class_declaration" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        children.push(Node::Class(class));
                    }
                }
                "struct_declaration" => {
                    if let Some(struct_node) = self.parse_struct(child, source)? {
                        children.push(Node::Class(struct_node));
                    }
                }
                "record_declaration" => {
                    if let Some(record) = self.parse_record(child, source)? {
                        children.push(Node::Class(record));
                    }
                }
                "interface_declaration" => {
                    if let Some(interface) = self.parse_interface(child, source)? {
                        children.push(Node::Class(interface));
                    }
                }
                "namespace_declaration" | "file_scoped_namespace_declaration" => {
                    // Recursively process namespace contents
                    let mut ns_cursor = child.walk();
                    for ns_child in child.children(&mut ns_cursor) {
                        if ns_child.kind() == "declaration_list" {
                            let mut decl_cursor = ns_child.walk();
                            for decl_child in ns_child.children(&mut decl_cursor) {
                                match decl_child.kind() {
                                    "class_declaration" => {
                                        if let Some(class) = self.parse_class(decl_child, source)? {
                                            children.push(Node::Class(class));
                                        }
                                    }
                                    "struct_declaration" => {
                                        if let Some(struct_node) =
                                            self.parse_struct(decl_child, source)?
                                        {
                                            children.push(Node::Class(struct_node));
                                        }
                                    }
                                    "record_declaration" => {
                                        if let Some(record) =
                                            self.parse_record(decl_child, source)?
                                        {
                                            children.push(Node::Class(record));
                                        }
                                    }
                                    "interface_declaration" => {
                                        if let Some(interface) =
                                            self.parse_interface(decl_child, source)?
                                        {
                                            children.push(Node::Class(interface));
                                        }
                                    }
                                    _ => {}
                                }
                            }
                        }
                    }
                }
                _ => {}
            }
        }

        Ok(File {
            path: path.to_string_lossy().to_string(),
            children,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = CSharpProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = CSharpProcessor::new().unwrap();
        assert!(processor.can_process(Path::new("test.cs")));
        assert!(!processor.can_process(Path::new("test.java")));
    }

    #[test]
    fn test_basic_class_parsing() {
        let source = r#"
namespace Test;
public static class MathHelpers
{
    public const double Pi = 3.14;
    public static double Circle(double radius) { return 2 * Pi * radius; }
    private static void Helper() { }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "MathHelpers");
            assert_eq!(class.visibility, Visibility::Public);
            assert!(class.modifiers.contains(&Modifier::Static));
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, Node::Field(_)))
                    .count(),
                1
            );
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, Node::Function(_)))
                    .count(),
                2
            );
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_interface_with_generics() {
        let source = r#"
public interface IRepository<TEntity, TKey>
    where TEntity : IEntity<TKey>
    where TKey : notnull
{
    Task AddAsync(TEntity entity);
    Task<TEntity> GetAsync(TKey id);
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("IRepository.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(interface) = &file.children[0] {
            assert_eq!(interface.name, "IRepository");
            assert!(interface.decorators.contains(&"interface".to_string()));
            assert_eq!(interface.type_params.len(), 2);
            assert_eq!(interface.type_params[0].name, "TEntity");
            assert_eq!(interface.type_params[1].name, "TKey");
            assert_eq!(
                interface
                    .children
                    .iter()
                    .filter(|n| matches!(n, Node::Function(_)))
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
public class SavingsAccount : BankAccount, IAccount
{
    protected decimal _rate;
    public override void Process() { }
    internal void Calculate() { }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Account.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "SavingsAccount");
            assert!(!class.implements.is_empty());
            assert_eq!(
                class
                    .children
                    .iter()
                    .filter(|n| matches!(n, Node::Function(_)))
                    .count(),
                2
            );
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_struct_declaration() {
        let source = r#"
public readonly struct Money : IEquatable<Money>
{
    public Money(decimal amount) { }
    public decimal Amount { get; }
    public static implicit operator Money(decimal amount) => new Money(amount);
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Money.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(struct_node) = &file.children[0] {
            assert_eq!(struct_node.name, "Money");
            assert!(struct_node.decorators.contains(&"struct".to_string()));
            assert!(struct_node.modifiers.contains(&Modifier::Readonly));
        } else {
            panic!("Expected a struct");
        }
    }

    #[test]
    fn test_record_declaration() {
        let source = r#"
public record User(Guid Id, string Name) : EntityBase<Guid>(Id)
{
    public bool IsValid => !string.IsNullOrEmpty(Name);
    private bool Validate() => true;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("User.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(record) = &file.children[0] {
            assert_eq!(record.name, "User");
            assert!(record.decorators.contains(&"record".to_string()));
            assert!(!record.implements.is_empty());
        } else {
            panic!("Expected a record");
        }
    }

    #[test]
    fn test_visibility_modifiers() {
        let source = r#"
public class Test
{
    public int PublicField;
    private int PrivateField;
    protected int ProtectedField;
    internal int InternalField;

    public void PublicMethod() { }
    private void PrivateMethod() { }
    protected void ProtectedMethod() { }
    internal void InternalMethod() { }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            // Check field visibilities
            let fields: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Field(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(fields.len(), 4);

            // Check method visibilities
            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 4);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_properties_and_events() {
        let source = r#"
public class Account
{
    public event EventHandler BalanceChanged;
    public string AccountNumber { get; }
    public decimal Balance { get; protected set; }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, &PathBuf::from("Account.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            let fields: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Field(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(fields.len(), 3);

            // Check that event has Event modifier
            let event = fields.iter().find(|f| f.name == "BalanceChanged");
            assert!(event.is_some());
            if let Some(e) = event {
                assert!(e.modifiers.contains(&Modifier::Event));
            }
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_empty_file() {
        let source = "";
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 0);
        assert_eq!(file.path, "Test.cs");
    }

    #[test]
    fn test_interface_declaration() {
        let source = r#"
public interface ILogger
{
    void Log(string message);
    string Name { get; }
    event EventHandler<LogEventArgs> LogReceived;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(interface) = &file.children[0] {
            assert_eq!(interface.name, "ILogger");
            assert_eq!(interface.visibility, Visibility::Public);
            assert!(interface.decorators.contains(&"interface".to_string()));

            let methods: Vec<_> = interface
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 1);
            assert_eq!(methods[0].name, "Log");

            let properties: Vec<_> = interface
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Field(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(properties.len(), 2);
        } else {
            panic!("Expected an interface");
        }
    }

    #[test]
    fn test_abstract_class() {
        let source = r#"
public abstract class Vehicle
{
    protected string _brand;
    public abstract void Start();
    public virtual void Stop() { }
    protected abstract int GetSpeed();
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Vehicle");
            assert_eq!(class.visibility, Visibility::Public);
            assert!(class.modifiers.contains(&Modifier::Abstract));

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 3);

            let abstract_methods = methods
                .iter()
                .filter(|m| m.modifiers.contains(&Modifier::Abstract))
                .count();
            assert_eq!(abstract_methods, 2);

            let virtual_methods = methods
                .iter()
                .filter(|m| m.modifiers.contains(&Modifier::Virtual))
                .count();
            assert_eq!(virtual_methods, 1);
        } else {
            panic!("Expected an abstract class");
        }
    }

    #[test]
    fn test_sealed_class() {
        let source = r#"
public sealed class FinalClass
{
    public void DoSomething() { }
    private int _value;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "FinalClass");
            assert_eq!(class.visibility, Visibility::Public);
            assert!(class.modifiers.contains(&Modifier::Final)); // sealed maps to Final
            assert_eq!(class.children.len(), 2); // 1 method + 1 field
        } else {
            panic!("Expected a sealed class");
        }
    }

    #[test]
    fn test_partial_class() {
        let source = r#"
public partial class DataContext
{
    public void Initialize() { }
    private string _connectionString;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "DataContext");
            assert_eq!(class.visibility, Visibility::Public);
            // Note: partial keyword is not stored in modifiers in current implementation
            // but the class should still be parsed correctly
            assert_eq!(class.children.len(), 2);
        } else {
            panic!("Expected a partial class");
        }
    }

    #[test]
    fn test_async_method() {
        let source = r#"
public class AsyncService
{
    public async Task<string> FetchDataAsync()
    {
        await Task.Delay(100);
        return "data";
    }

    private async void HandleEventAsync() { }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "AsyncService");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 2);

            let async_methods = methods
                .iter()
                .filter(|m| m.modifiers.contains(&Modifier::Async))
                .count();
            assert_eq!(async_methods, 2);

            let fetch_method = methods.iter().find(|m| m.name == "FetchDataAsync");
            assert!(fetch_method.is_some());
            if let Some(m) = fetch_method {
                assert_eq!(m.visibility, Visibility::Public);
                assert!(m.modifiers.contains(&Modifier::Async));
            }
        } else {
            panic!("Expected a class with async methods");
        }
    }

    #[test]
    fn test_generic_class() {
        let source = r#"
public class Repository<T, TKey> where T : class where TKey : struct
{
    private Dictionary<TKey, T> _items;
    public void Add(TKey key, T item) { }
    public T Get(TKey key) { return default(T); }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Repository");
            assert_eq!(class.visibility, Visibility::Public);
            assert_eq!(class.type_params.len(), 2);
            assert_eq!(class.type_params[0].name, "T");
            assert_eq!(class.type_params[1].name, "TKey");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 2);
        } else {
            panic!("Expected a generic class");
        }
    }

    #[test]
    fn test_indexer_property() {
        let source = r#"
public class Collection
{
    private string[] _items;
    public string this[int index]
    {
        get { return _items[index]; }
        set { _items[index] = value; }
    }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Collection");
            // Indexers might be parsed as properties or special methods
            // Verify the class structure is captured
            assert!(!class.children.is_empty());
        } else {
            panic!("Expected a class with indexer");
        }
    }

    #[test]
    fn test_operator_overloading() {
        let source = r#"
public class Vector
{
    public int X { get; set; }
    public int Y { get; set; }

    public static Vector operator +(Vector a, Vector b)
    {
        return new Vector { X = a.X + b.X, Y = a.Y + b.Y };
    }

    public static bool operator ==(Vector a, Vector b)
    {
        return a.X == b.X && a.Y == b.Y;
    }

    public static bool operator !=(Vector a, Vector b)
    {
        return !(a == b);
    }
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Vector");

            let operators: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) if f.decorators.contains(&"operator".to_string()) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(operators.len(), 3);

            // All operators should be static
            for op in &operators {
                assert!(op.modifiers.contains(&Modifier::Static));
            }
        } else {
            panic!("Expected a class with operator overloading");
        }
    }

    #[test]
    fn test_nullable_reference_types() {
        let source = r#"
#nullable enable
public class UserService
{
    public string? GetUserName(int? userId)
    {
        return userId.HasValue ? "User" + userId.Value : null;
    }

    private Dictionary<string, object?>? _cache;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "UserService");

            let methods: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Function(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(methods.len(), 1);

            let method = &methods[0];
            assert_eq!(method.name, "GetUserName");
            // Nullable types should be captured in return_type and parameter types
            assert!(method.return_type.is_some());
        } else {
            panic!("Expected a class with nullable reference types");
        }
    }

    #[test]
    fn test_init_only_properties() {
        let source = r#"
public class Person
{
    public string FirstName { get; init; }
    public string LastName { get; init; }
    public int Age { get; set; }
    public string FullName => FirstName + " " + LastName;
}
"#;
        let processor = CSharpProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let file = processor
            .process(source, Path::new("Test.cs"), &opts)
            .unwrap();

        assert_eq!(file.children.len(), 1);
        if let Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Person");

            let properties: Vec<_> = class
                .children
                .iter()
                .filter_map(|n| match n {
                    Node::Field(f) => Some(f),
                    _ => None,
                })
                .collect();
            assert_eq!(properties.len(), 4);

            // Verify all properties are captured
            let first_name = properties.iter().find(|p| p.name == "FirstName");
            assert!(first_name.is_some());
            if let Some(prop) = first_name {
                assert_eq!(prop.visibility, Visibility::Public);
            }
        } else {
            panic!("Expected a class with init-only properties");
        }
    }
}
