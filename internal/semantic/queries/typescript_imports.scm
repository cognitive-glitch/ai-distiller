;; TypeScript import statements query
;; Finds all import and export statements

;; ES6 imports: import { symbol } from "module"
(import_statement) @import.statement

;; Dynamic imports: import("module")
(call_expression
  function: (import) @dynamic_import) @dynamic_import.statement

;; Require statements: const module = require("module")
(variable_declarator
  value: (call_expression
    function: (identifier) @require.function
    (#eq? @require.function "require"))) @require.statement

;; Export statements
(export_statement) @export.statement