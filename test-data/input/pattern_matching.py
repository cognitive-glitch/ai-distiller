"""Test cases for Python 3.10+ pattern matching (match/case statements)."""

# Basic match statement
def basic_match(value):
    match value:
        case 0:
            return "zero"
        case 1:
            return "one"
        case _:
            return "other"

# Sequence patterns
def sequence_patterns(seq):
    match seq:
        case []:
            return "empty"
        case [x]:
            return f"single: {x}"
        case [x, y]:
            return f"pair: {x}, {y}"
        case [x, y, *rest]:
            return f"first two: {x}, {y}, rest: {rest}"
        case _:
            return "other sequence"

# Mapping patterns
def mapping_patterns(data):
    match data:
        case {"status": 200, "data": content}:
            return f"success: {content}"
        case {"status": 404}:
            return "not found"
        case {"status": code, "error": msg}:
            return f"error {code}: {msg}"
        case _:
            return "unknown response"

# Class patterns
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

def class_patterns(obj):
    match obj:
        case Point(x=0, y=0):
            return "origin"
        case Point(x=0, y=y):
            return f"on y-axis at {y}"
        case Point(x=x, y=0):
            return f"on x-axis at {x}"
        case Point(x=x, y=y):
            return f"point at ({x}, {y})"
        case _:
            return "not a point"

# Guards in patterns
def pattern_with_guards(value):
    match value:
        case x if x < 0:
            return "negative"
        case x if x == 0:
            return "zero"
        case x if x > 0 and x < 10:
            return "small positive"
        case x if x >= 10:
            return "large positive"
        case _:
            return "unexpected"

# Complex nested patterns
def nested_patterns(data):
    match data:
        case {"users": [{"name": name, "age": age} | {"username": name, "age": age}]} if age >= 18:
            return f"Adult user: {name}"
        case {"users": [{"name": name, "age": age}]} if age < 18:
            return f"Minor user: {name}"
        case {"users": [*users]} if len(users) > 1:
            return f"Multiple users: {len(users)}"
        case _:
            return "Invalid data"

# Or patterns
def or_patterns(value):
    match value:
        case "yes" | "y" | "true" | "1":
            return True
        case "no" | "n" | "false" | "0":
            return False
        case _:
            return None

# As patterns (capture)
def as_patterns(data):
    match data:
        case [x, y] as coords:
            return f"Coordinates {coords}: x={x}, y={y}"
        case {"name": name, **rest} as person:
            return f"Person {name} with data: {rest}"
        case _:
            return "Unknown"

# Wildcard patterns
def wildcard_patterns(value):
    match value:
        case [_, _, third, *_]:
            return f"Third element is {third}"
        case {"key": _, "value": val}:
            return f"Value is {val}, key ignored"
        case _:
            return "No match"

# Literal patterns
def literal_patterns(value):
    match value:
        case True:
            return "boolean true"
        case False:
            return "boolean false"
        case None:
            return "none value"
        case 42:
            return "the answer"
        case "hello":
            return "greeting"
        case 3.14:
            return "pi approximation"
        case _:
            return "other literal"