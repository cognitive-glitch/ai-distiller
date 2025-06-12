"""Test cases for pattern matching syntax errors and edge cases."""

# ERROR: case outside match
def invalid_case():
    case x:  # SyntaxError: case statement outside match
        return x

# ERROR: match without case
def match_without_case(value):
    match value:
        pass  # SyntaxError: expected case clause

# ERROR: duplicate capture variable names
def duplicate_captures(pair):
    match pair:
        case [x, x]:  # SyntaxError: multiple assignments to name 'x'
            return x

# ERROR: multiple starred names in sequence pattern
def multiple_stars(seq):
    match seq:
        case [*first, *second]:  # SyntaxError: multiple starred names in sequence pattern
            return first, second

# ERROR: invalid pattern syntax
def invalid_patterns(value):
    match value:
        case x + y:  # SyntaxError: invalid pattern (can't use operators)
            return x + y
        case f(x):  # SyntaxError: invalid pattern (can't call functions)
            return x

# ERROR: invalid use of underscore
def invalid_underscore(pair):
    match pair:
        case [_, _] if _ > 0:  # SyntaxError: can't use _ in guard
            return True

# ERROR: invalid class pattern
def invalid_class_pattern(obj):
    match obj:
        case Point[x, y]:  # SyntaxError: invalid pattern (should use parentheses)
            return x, y

# Valid but complex: nested match statements
def nested_match(outer, inner):
    match outer:
        case "process":
            match inner:
                case {"type": "data", "value": v}:
                    return f"Processing data: {v}"
                case {"type": "error", "message": m}:
                    return f"Processing error: {m}"
                case _:
                    return "Unknown inner"
        case "skip":
            return "Skipping"
        case _:
            return "Unknown outer"

# Edge case: match with only wildcard
def match_only_wildcard(value):
    match value:
        case _:
            return "matches anything"

# Edge case: empty guard
def empty_guard(value):
    match value:
        case x if :  # SyntaxError: invalid syntax (empty condition)
            return x

# Edge case: pattern with trailing comma
def trailing_comma_pattern(value):
    match value:
        case [x, y,]:  # Valid: trailing comma is allowed
            return x, y

# Complex valid pattern
def complex_valid_pattern(data):
    match data:
        case {"type": "user", "data": {"name": str(name), "age": int(age)}} if age >= 0:
            return f"Valid user: {name}, age {age}"
        case {"type": "user", "data": {"name": name}} if isinstance(name, str):
            return f"User without age: {name}"
        case _:
            return "Invalid user data"