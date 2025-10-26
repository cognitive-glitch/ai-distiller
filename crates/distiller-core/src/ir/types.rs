//! Type system for IR nodes

use serde::{Deserialize, Serialize};

/// Visibility level of a code element
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Visibility {
    /// Public - accessible from anywhere
    Public,
    /// Protected - accessible from subclasses
    Protected,
    /// Internal/Package-private - accessible within package/module
    Internal,
    /// Private - accessible only within the same class/file
    Private,
}

impl Default for Visibility {
    fn default() -> Self {
        Self::Public
    }
}

/// Modifier for functions, classes, fields
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Modifier {
    Static,
    Abstract,
    Final,
    Async,
    Virtual,
    Override,
    Const,
    Readonly,
    Mutable,
    Event,
    Data,
    Sealed,
    Inline,
}

/// Type reference
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct TypeRef {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub package: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub type_args: Vec<TypeRef>,
    #[serde(skip_serializing_if = "std::ops::Not::not", default)]
    pub is_nullable: bool,
    #[serde(skip_serializing_if = "std::ops::Not::not", default)]
    pub is_array: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub array_dims: Option<usize>,
}

impl TypeRef {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            package: None,
            type_args: Vec::new(),
            is_nullable: false,
            is_array: false,
            array_dims: None,
        }
    }
}

/// Type parameter (generic)
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct TypeParam {
    pub name: String,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub constraints: Vec<TypeRef>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub default: Option<TypeRef>,
}

/// Function/method parameter
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Parameter {
    pub name: String,
    pub param_type: TypeRef,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub default_value: Option<String>,
    #[serde(skip_serializing_if = "std::ops::Not::not", default)]
    pub is_variadic: bool,
    #[serde(skip_serializing_if = "std::ops::Not::not", default)]
    pub is_optional: bool,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub decorators: Vec<String>,
}

/// Import symbol
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ImportedSymbol {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub alias: Option<String>,
}
