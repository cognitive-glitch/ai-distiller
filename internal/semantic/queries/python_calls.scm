;; Python function and method calls query
;; Finds all function calls, method calls, and constructor calls

;; Simple function calls: func(args)
(call
  function: (identifier) @call.function) @call.simple

;; Method calls: obj.method(args)
(call
  function: (attribute
    attribute: (identifier) @call.method)) @call.method

;; Constructor calls (class instantiation): Class(args)
(call
  function: (identifier) @constructor.class) @constructor.call

;; Built-in function calls
(call
  function: (identifier) @builtin.function
  (#match? @builtin.function "^(len|str|int|float|bool|list|dict|set|tuple|range|print|input|type)$")) @builtin.call