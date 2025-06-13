;; TypeScript function and method calls query
;; Finds all function calls, method calls, and constructor calls

;; Function calls: func(args)
(call_expression
  function: (identifier) @call.function) @call.simple

;; Method calls: obj.method(args)
(call_expression
  function: (member_expression
    property: (property_identifier) @call.method)) @call.method

;; Constructor calls: new Class(args)
(new_expression
  constructor: (identifier) @constructor.class) @constructor.call

;; Generic function calls: func<T>(args)
(call_expression
  function: (identifier) @generic_call.function
  type_arguments: (type_arguments)) @generic_call.simple

;; Chained method calls: obj.method().method2(args)
(call_expression
  function: (member_expression
    object: (call_expression)
    property: (property_identifier) @chained_call.method)) @chained_call.method

;; Optional chaining calls: obj?.method(args)
(call_expression
  function: (member_expression
    object: (_)
    property: (property_identifier) @optional_call.method
    optional_chain: "?.")) @optional_call.method