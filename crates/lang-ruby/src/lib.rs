use distiller_core::{
    ProcessOptions,
    error::{DistilError, Result},
    ir::{self, *},
    processor::LanguageProcessor,
};
use parking_lot::Mutex;
use std::path::Path;
use std::sync::Arc;
use tree_sitter::{Node as TSNode, Parser};

pub struct RubyProcessor {
    parser: Arc<Mutex<Parser>>,
}

impl RubyProcessor {
    pub fn new() -> Result<Self> {
        let mut parser = Parser::new();
        parser
            .set_language(&tree_sitter_ruby::LANGUAGE.into())
            .map_err(|e| DistilError::parse_error("ruby", e.to_string()))?;

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

    fn parse_visibility(node: TSNode, source: &str) -> Visibility {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            if child.kind() == "visibility" {
                let text = Self::node_text(child, source);
                return match text.as_str() {
                    "private" => Visibility::Private,
                    "protected" => Visibility::Protected,
                    _ => Visibility::Public,
                };
            }
        }

        // Check for @private comment (RDoc convention)
        if let Some(parent) = node.parent() {
            let mut cursor = parent.walk();
            for sibling in parent.children(&mut cursor) {
                if sibling.kind() == "comment" {
                    let comment = Self::node_text(sibling, source);
                    if comment.contains("@private") {
                        return Visibility::Private;
                    }
                }
            }
        }

        Visibility::Public
    }

    fn parse_class(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        let mut name = String::new();
        let mut extends = Vec::new();
        let mut children = Vec::new();
        let visibility = Self::parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "constant" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "superclass" => {
                    let mut sc_cursor = child.walk();
                    for sc_child in child.children(&mut sc_cursor) {
                        if sc_child.kind() == "constant" {
                            extends.push(TypeRef::new(Self::node_text(sc_child, source)));
                        }
                    }
                }
                "body_statement" => {
                    self.parse_body(child, source, &mut children)?;
                }
                _ => {}
            }
        }

        Ok(Some(Class {
            name,
            visibility,
            extends,
            implements: vec![],
            type_params: vec![],
            decorators: vec![],
            modifiers: vec![],
            children,
            line_start,
            line_end,
        }))
    }

    fn parse_module(&self, node: TSNode, source: &str) -> Result<Option<Class>> {
        // Ruby modules are similar to classes
        let mut name = String::new();
        let mut children = Vec::new();
        let visibility = Self::parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "constant" => {
                    if name.is_empty() {
                        name = Self::node_text(child, source);
                    }
                }
                "body_statement" => {
                    self.parse_body(child, source, &mut children)?;
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
            decorators: vec!["module".to_string()],
            modifiers: vec![],
            children,
            line_start,
            line_end,
        }))
    }

    #[allow(clippy::unused_self)]
    fn parse_method(&self, node: TSNode, source: &str) -> Result<Option<Function>> {
        let mut name = String::new();
        let mut parameters = Vec::new();
        let visibility = Self::parse_visibility(node, source);

        let line_start = node.start_position().row + 1;
        let line_end = node.end_position().row + 1;

        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier" | "constant" => {
                    // For singleton methods like "def self.method_name",
                    // we want the last identifier (the method name, not "self")
                    let text = Self::node_text(child, source);
                    if !text.is_empty() {
                        name = text;
                    }
                }
                "method_parameters" => {
                    Self::parse_parameters(child, source, &mut parameters)?;
                }
                _ => {}
            }
        }

        Ok(Some(Function {
            name,
            visibility,
            parameters,
            return_type: None,
            decorators: vec![],
            modifiers: vec![],
            type_params: vec![],
            implementation: None,
            line_start,
            line_end,
        }))
    }

    fn parse_parameters(node: TSNode, source: &str, params: &mut Vec<Parameter>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "identifier"
                | "optional_parameter"
                | "splat_parameter"
                | "hash_splat_parameter"
                | "block_parameter"
                | "keyword_parameter" => {
                    let param_name = if child.kind() == "optional_parameter"
                        || child.kind() == "keyword_parameter"
                    {
                        let mut param_cursor = child.walk();
                        let mut found_name = String::new();
                        for param_child in child.children(&mut param_cursor) {
                            if param_child.kind() == "identifier" {
                                found_name = Self::node_text(param_child, source);
                                break;
                            }
                        }
                        found_name
                    } else {
                        Self::node_text(child, source)
                    };

                    if !param_name.is_empty() {
                        params.push(Parameter {
                            name: param_name,
                            param_type: TypeRef::new(""),
                            default_value: None,
                            is_variadic: child.kind() == "splat_parameter",
                            is_optional: child.kind() == "optional_parameter",
                            decorators: vec![],
                        });
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }

    fn parse_body(&self, node: TSNode, source: &str, children: &mut Vec<ir::Node>) -> Result<()> {
        let mut cursor = node.walk();
        for child in node.children(&mut cursor) {
            match child.kind() {
                "method" | "singleton_method" => {
                    if let Some(method) = self.parse_method(child, source)? {
                        children.push(ir::Node::Function(method));
                    }
                }
                "class" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        children.push(ir::Node::Class(class));
                    }
                }
                "module" => {
                    if let Some(module) = self.parse_module(child, source)? {
                        children.push(ir::Node::Class(module));
                    }
                }
                _ => {}
            }
        }
        Ok(())
    }
}

impl LanguageProcessor for RubyProcessor {
    fn language(&self) -> &'static str {
        "Ruby"
    }

    fn supported_extensions(&self) -> &'static [&'static str] {
        &["rb", "rake", "gemspec"]
    }

    fn can_process(&self, path: &Path) -> bool {
        // Check for special files first
        if let Some(filename) = path.file_name().and_then(|n| n.to_str())
            && (filename == "Rakefile" || filename == "Gemfile")
        {
            return true;
        }

        // Then check extension
        path.extension()
            .and_then(|ext| ext.to_str())
            .map(|ext| self.supported_extensions().contains(&ext))
            .unwrap_or(false)
    }

    fn process(&self, source: &str, path: &Path, _opts: &ProcessOptions) -> Result<File> {
        let mut parser = self.parser.lock();
        let tree = parser.parse(source, None).ok_or_else(|| {
            DistilError::parse_error(path.display().to_string(), "Failed to parse Ruby source")
        })?;

        let mut file = File {
            path: path.to_string_lossy().to_string(),
            children: vec![],
        };

        let root = tree.root_node();
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            match child.kind() {
                "class" => {
                    if let Some(class) = self.parse_class(child, source)? {
                        file.children.push(ir::Node::Class(class));
                    }
                }
                "module" => {
                    if let Some(module) = self.parse_module(child, source)? {
                        file.children.push(ir::Node::Class(module));
                    }
                }
                "method" | "singleton_method" => {
                    if let Some(method) = self.parse_method(child, source)? {
                        file.children.push(ir::Node::Function(method));
                    }
                }
                _ => {}
            }
        }

        Ok(file)
    }
}

impl Default for RubyProcessor {
    fn default() -> Self {
        Self::new().expect("Failed to create Ruby processor")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_processor_creation() {
        let processor = RubyProcessor::new();
        assert!(processor.is_ok());
    }

    #[test]
    fn test_file_extension_detection() {
        let processor = RubyProcessor::new().unwrap();
        assert!(processor.can_process(&PathBuf::from("test.rb")));
        assert!(processor.can_process(&PathBuf::from("Rakefile")));
        assert!(processor.can_process(&PathBuf::from("Gemfile")));
        assert!(!processor.can_process(&PathBuf::from("test.py")));
    }

    #[test]
    fn test_class_with_methods() {
        let source = r#"
class User
  def initialize(name)
    @name = name
  end

  def greet
    puts "Hello, #{@name}!"
  end

  private

  def secret
    "secret"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "User");
            assert_eq!(class.children.len(), 3);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_module_definition() {
        let source = r#"
module Utilities
  def self.helper
    "help"
  end

  def instance_method
    "instance"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(module) = &file.children[0] {
            assert_eq!(module.name, "Utilities");
            assert!(module.decorators.contains(&"module".to_string()));
            assert_eq!(module.children.len(), 2);
        } else {
            panic!("Expected a module");
        }
    }

    #[test]
    fn test_method_parameters() {
        let source = r#"
def complex_method(required, optional = nil, *args, **kwargs, &block)
  # method body
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Function(func) = &file.children[0] {
            assert_eq!(func.name, "complex_method");
            assert!(func.parameters.len() >= 2); // At least required and optional
        } else {
            panic!("Expected a function");
        }
    }

    #[test]
    fn test_inheritance() {
        let source = r#"
class Admin < User
  def admin_method
    "admin"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, &PathBuf::from("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Admin");
            assert_eq!(class.extends.len(), 1);
            assert_eq!(class.extends[0].name, "User");
        } else {
            panic!("Expected a class");
        }
    }

    // ===== Enhanced Test Coverage (11 new tests) =====

    #[test]
    fn test_empty_file() {
        let source = "";
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 0, "Empty file should have no children");
    }

    #[test]
    fn test_class_methods() {
        let source = r#"
class Calculator
  def self.add(a, b)
    a + b
  end

  def self.multiply(x, y)
    x * y
  end

  def instance_method
    "instance"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Calculator");
            assert_eq!(class.children.len(), 3);

            // Check that class methods are captured
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

            assert_eq!(methods.len(), 3);
            assert_eq!(methods[0].name, "add");
            assert_eq!(methods[1].name, "multiply");
            assert_eq!(methods[2].name, "instance_method");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_attr_accessor() {
        let source = r#"
class Person
  attr_accessor :name, :age
  attr_reader :id
  attr_writer :email

  def initialize(id)
    @id = id
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Person");
            // At minimum, should have initialize method
            assert!(!class.children.is_empty());
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_block_syntax() {
        let source = r#"
def with_block(items, &block)
  items.each(&block)
end

def yield_example
  yield if block_given?
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 2);

        if let ir::Node::Function(func) = &file.children[0] {
            assert_eq!(func.name, "with_block");
            // Should capture block parameter
            assert!(!func.parameters.is_empty());
        } else {
            panic!("Expected a function");
        }
    }

    #[test]
    fn test_singleton_methods() {
        let source = r#"
class MyClass
  class << self
    def singleton_one
      "singleton"
    end

    def singleton_two(arg)
      arg
    end
  end

  def self.class_method
    "class method"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "MyClass");
            // At minimum should capture the self.class_method
            assert!(!class.children.is_empty());
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_module_mixin() {
        let source = r#"
module Loggable
  def log(message)
    puts message
  end
end

class Service
  include Loggable

  def perform
    log("performing")
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 2);

        // Check module
        if let ir::Node::Class(module) = &file.children[0] {
            assert_eq!(module.name, "Loggable");
            assert!(module.decorators.contains(&"module".to_string()));
        } else {
            panic!("Expected a module");
        }

        // Check class
        if let ir::Node::Class(class) = &file.children[1] {
            assert_eq!(class.name, "Service");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_visibility_keywords() {
        let source = r#"
class VisibilityExample
  def public_method
    "public"
  end

  private

  def private_method
    "private"
  end

  protected

  def protected_method
    "protected"
  end

  public

  def another_public
    "public again"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "VisibilityExample");
            assert_eq!(class.children.len(), 4);

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

            assert_eq!(methods.len(), 4);
            assert_eq!(methods[0].name, "public_method");
            assert_eq!(methods[1].name, "private_method");
            assert_eq!(methods[2].name, "protected_method");
            assert_eq!(methods[3].name, "another_public");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_method_aliasing() {
        let source = r#"
class Aliased
  def original_method
    "original"
  end

  alias_method :new_name, :original_method
  alias another_alias original_method

  def another_method
    "another"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Aliased");
            // At minimum should capture the actual method definitions
            assert!(class.children.len() >= 2);
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_nested_classes() {
        let source = r#"
class Outer
  class Inner
    def inner_method
      "inner"
    end
  end

  class AnotherInner
    def another_inner_method
      "another"
    end
  end

  def outer_method
    "outer"
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(outer) = &file.children[0] {
            assert_eq!(outer.name, "Outer");
            // Should have 2 nested classes + 1 method
            assert_eq!(outer.children.len(), 3);

            // Find nested classes
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

            assert_eq!(nested_classes.len(), 2);
            assert_eq!(nested_classes[0].name, "Inner");
            assert_eq!(nested_classes[1].name, "AnotherInner");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_multiple_inheritance() {
        let source = r#"
module Logging
  def log(msg)
    puts msg
  end
end

module Validation
  def validate
    true
  end
end

module Serialization
  def to_json
    "{}"
  end
end

class MultiMixin
  include Logging
  include Validation
  include Serialization

  def process
    validate && log("processing")
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        // 3 modules + 1 class
        assert_eq!(file.children.len(), 4);

        // Verify modules
        let modules: Vec<_> = file
            .children
            .iter()
            .filter_map(|n| {
                if let ir::Node::Class(c) = n {
                    if c.decorators.contains(&"module".to_string()) {
                        Some(c)
                    } else {
                        None
                    }
                } else {
                    None
                }
            })
            .collect();

        assert_eq!(modules.len(), 3);
        assert_eq!(modules[0].name, "Logging");
        assert_eq!(modules[1].name, "Validation");
        assert_eq!(modules[2].name, "Serialization");

        // Verify class
        if let ir::Node::Class(class) = &file.children[3] {
            assert_eq!(class.name, "MultiMixin");
        } else {
            panic!("Expected a class");
        }
    }

    #[test]
    fn test_multiple_methods_with_parameters() {
        let source = r#"
class Calculator
  def add(a, b)
    a + b
  end

  def subtract(x, y)
    x - y
  end

  def multiply(*numbers)
    numbers.reduce(1, :*)
  end

  def divide(numerator, denominator = 1)
    numerator / denominator
  end
end
"#;
        let processor = RubyProcessor::new().unwrap();
        let opts = ProcessOptions::default();
        let result = processor.process(source, Path::new("test.rb"), &opts);
        assert!(result.is_ok());

        let file = result.unwrap();
        assert_eq!(file.children.len(), 1);

        if let ir::Node::Class(class) = &file.children[0] {
            assert_eq!(class.name, "Calculator");
            assert_eq!(class.children.len(), 4);

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

            assert_eq!(methods.len(), 4);
            assert_eq!(methods[0].name, "add");
            assert_eq!(methods[0].parameters.len(), 2);
            assert_eq!(methods[1].name, "subtract");
            assert_eq!(methods[1].parameters.len(), 2);
            assert_eq!(methods[2].name, "multiply");
            assert!(!methods[2].parameters.is_empty()); // variadic parameter
            assert_eq!(methods[3].name, "divide");
            assert!(!methods[3].parameters.is_empty()); // required + optional
        } else {
            panic!("Expected a class");
        }
    }
}
