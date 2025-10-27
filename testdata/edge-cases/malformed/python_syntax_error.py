"""
Edge case: Python file with syntax errors.
Parser should handle gracefully without crashing.
"""

# Valid class
class ValidClass:
    def valid_method(self):
        return 42

# Syntax error: Missing closing parenthesis
def broken_function(x, y
    return x + y

# Syntax error: Unclosed string
message = "This string never ends

# Valid function after errors
def another_valid_function():
    return "still parsing"

# Syntax error: Invalid indentation
class AnotherClass:
def misindented_method():
    pass

# Syntax error: Missing colon
class BrokenClass
    def method(self):
        pass
