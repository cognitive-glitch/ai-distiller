;; TypeScript symbol declarations query
;; Finds all function, class, interface, and variable declarations

;; Function declarations
(function_declaration
  name: (identifier) @function.name) @function.definition

;; Arrow function expressions assigned to variables
(variable_declarator
  name: (identifier) @arrow_function.name
  value: (arrow_function)) @arrow_function.definition

;; Class declarations
(class_declaration
  name: (type_identifier) @class.name) @class.definition

;; Interface declarations
(interface_declaration
  name: (type_identifier) @interface.name) @interface.definition

;; Type alias declarations
(type_alias_declaration
  name: (type_identifier) @type.name) @type.definition

;; Enum declarations
(enum_declaration
  name: (identifier) @enum.name) @enum.definition

;; Variable declarations
(variable_declarator
  name: (identifier) @variable.name) @variable.declaration

;; Method definitions in classes
(method_definition
  name: (property_identifier) @method.name) @method.definition

;; Property definitions in classes
(public_field_definition
  name: (property_identifier) @property.name) @property.definition

;; Abstract method definitions
(abstract_method_signature
  name: (property_identifier) @abstract_method.name) @abstract_method.definition

;; Namespace declarations
(module_declaration
  name: (identifier) @namespace.name) @namespace.definition

;; Export declarations
(export_statement) @export.statement