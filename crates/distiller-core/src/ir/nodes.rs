//! IR node types

use super::types::{ImportedSymbol, Modifier, Parameter, TypeParam, TypeRef, Visibility};
use serde::{Deserialize, Serialize};

/// Root IR node - can be any type
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum Node {
    File(File),
    Directory(Directory),
    Package(Package),
    Import(Import),
    Class(Class),
    Interface(Interface),
    Struct(Struct),
    Enum(Enum),
    TypeAlias(TypeAlias),
    Function(Function),
    Field(Field),
    Comment(Comment),
    RawContent(RawContent),
}

/// File node
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct File {
    pub path: String,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
}

/// Directory node
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Directory {
    pub path: String,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
}

/// Package/module declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub name: String,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
}

/// Import statement
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Import {
    pub import_type: String, // "import", "from", "require", etc.
    pub module: String,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub symbols: Vec<ImportedSymbol>,
    #[serde(skip_serializing_if = "std::ops::Not::not", default)]
    pub is_type: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub line: Option<usize>,
}

/// Class declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Class {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub modifiers: Vec<Modifier>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub decorators: Vec<String>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_params: Vec<TypeParam>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub extends: Vec<TypeRef>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub implements: Vec<TypeRef>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
    pub line_start: usize,
    pub line_end: usize,
}

/// Interface declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Interface {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_params: Vec<TypeParam>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub extends: Vec<TypeRef>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
    pub line_start: usize,
    pub line_end: usize,
}

/// Struct declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Struct {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_params: Vec<TypeParam>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
    pub line_start: usize,
    pub line_end: usize,
}

/// Enum declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Enum {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub enum_type: Option<TypeRef>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub children: Vec<Node>,
    pub line_start: usize,
    pub line_end: usize,
}

/// Type alias
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TypeAlias {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_params: Vec<TypeParam>,
    pub alias_type: TypeRef,
    pub line: usize,
}

/// Function/method declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Function {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub modifiers: Vec<Modifier>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub decorators: Vec<String>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_params: Vec<TypeParam>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub parameters: Vec<Parameter>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub return_type: Option<TypeRef>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub implementation: Option<String>,
    pub line_start: usize,
    pub line_end: usize,
}

/// Field/property declaration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Field {
    pub name: String,
    pub visibility: Visibility,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub modifiers: Vec<Modifier>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub field_type: Option<TypeRef>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub default_value: Option<String>,
    pub line: usize,
}

/// Comment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Comment {
    pub text: String,
    pub format: String, // "line", "block", "doc"
    pub line: usize,
}

/// Raw content (unparsed)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RawContent {
    pub content: String,
}
