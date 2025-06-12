"""Test cases for advanced f-string features."""

# Basic f-strings
name = "Alice"
age = 30
basic = f"Hello, {name}! You are {age} years old."

# Debug format (= specifier)
x = 42
y = 3.14
debug_format = f"{x=} and {y=}"
expression_debug = f"{x + y=}"

# Complex expressions
import math
complex_expr = f"Square root of 2 is {math.sqrt(2):.4f}"
ternary_in_fstring = f"Status: {'active' if age >= 18 else 'minor'}"

# Nested f-strings
inner = "world"
nested = f"Hello {f'beautiful {inner}'}"
double_nested = f"Outer {f'Middle {f"Inner {x}"}'}'"

# Format specifiers
number = 1234.5678
formatted_number = f"{number:,.2f}"  # 1,234.57
padded = f"{name:>10}"  # Right-aligned
centered = f"{name:^10}"  # Centered
binary = f"{x:b}"  # Binary representation
hex_format = f"{x:#x}"  # Hexadecimal with prefix

# Multiple line f-strings
multiline = f"""
Name: {name}
Age: {age}
Status: {'Active' if age >= 18 else 'Inactive'}
"""

# Dictionary and attribute access
data = {"key": "value", "count": 42}
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

point = Point(3, 4)
dict_access = f"Key value: {data['key']}, Count: {data['count']}"
attr_access = f"Point: ({point.x}, {point.y})"

# Function calls in f-strings
def greet(name):
    return f"Hello, {name}!"

func_in_fstring = f"Greeting: {greet('Bob')}"

# List comprehension in f-string
numbers = [1, 2, 3, 4, 5]
comp_in_fstring = f"Squares: {[n**2 for n in numbers]}"

# Complex format specifications
value = 42
complex_format = f"{value:0>5d}"  # 00042 (zero-padded to 5 digits)
percentage = f"{0.1234:.1%}"  # 12.3%

# Unicode in f-strings
emoji = "🐍"
unicode_fstring = f"Python {emoji} is awesome!"

# Escaping braces
escaped = f"{{This is in braces}}"
mixed_escape = f"{{x}} equals {x}"

# ERROR CASES

# Unclosed brace
# bad_unclosed = f"This is bad {x"  # SyntaxError

# Unmatched braces
# bad_unmatched = f"This {x}} is bad"  # ValueError

# Empty expression
# bad_empty = f"Empty {}"  # SyntaxError

# Backslash in expression
# bad_backslash = f"{x\n}"  # SyntaxError

# Comments in expressions (not allowed)
# bad_comment = f"{x # comment}"  # SyntaxError

# Complex valid example with everything
def format_report(title, data_points, precision=2):
    return f"""
{'=' * 50}
{title:^50}
{'=' * 50}

Summary:
{f"Total points: {len(data_points)}":<30} {f"Precision: {precision}":>20}

Data Analysis:
{'-' * 50}
{chr(10).join(f"{i+1:3d}. Value: {val:8.{precision}f} | Debug: {val=}" 
              for i, val in enumerate(data_points[:5]))}

Statistics:
Min: {min(data_points):,.{precision}f}
Max: {max(data_points):,.{precision}f}
Avg: {sum(data_points)/len(data_points):,.{precision}f}

{f"Report generated with {len(data_points)} data points":#^50}
"""

# Lambda in f-string (unusual but valid)
lambda_fstring = f"Result: {(lambda x: x * 2)(5)}"

# Walrus operator in f-string (Python 3.8+)
walrus_fstring = f"Length is {(n := len(numbers))} for {n} numbers"