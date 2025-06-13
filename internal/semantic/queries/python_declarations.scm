;; Python symbol declarations query
;; Finds all function, class, method, and variable declarations

;; Function definitions
(function_definition
  name: (identifier) @function.name) @function.definition

;; Class definitions  
(class_definition
  name: (identifier) @class.name) @class.definition

;; Variable assignments (simple cases)
(assignment
  left: (identifier) @variable.name) @variable.assignment

;; Constants (ALL_CAPS variable assignments)
(assignment
  left: (identifier) @constant.name
  (#match? @constant.name "^[A-Z][A-Z0-9_]*$")) @constant.assignment

;; Decorated definitions (for properties, static methods, etc.)
(decorated_definition
  definition: (function_definition
    name: (identifier) @decorated.name)) @decorated.definition