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

;; Object instantiations (obj = ClassName())
(assignment
  left: (identifier) @instantiation.variable
  right: (call
    function: (identifier) @instantiation.class)) @instantiation.assignment

;; Constants (ALL_CAPS variable assignments)
(assignment
  left: (identifier) @constant.name
  (#match? @constant.name "^[A-Z][A-Z0-9_]*$")) @constant.assignment

;; Methods within classes (functions inside class bodies)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @method.name) @method.definition))

;; Decorated methods within classes
(class_definition
  body: (block
    (decorated_definition
      definition: (function_definition
        name: (identifier) @decorated_method.name)) @decorated_method.definition))

;; Decorated definitions (for properties, static methods, etc.)
(decorated_definition
  definition: (function_definition
    name: (identifier) @decorated.name)) @decorated.definition