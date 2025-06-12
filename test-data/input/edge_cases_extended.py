"""Extended edge cases for Python parser testing."""

# Empty classes and functions
class EmptyClass: pass
class AnotherEmpty:
    """Just a docstring."""

def empty_func(): pass
def another_empty():
    """Just a docstring."""
    
# Unicode in identifiers (Python 3 allows this)
class ΜαθηματικάΣύμβολα:
    """Greek class name."""
    π = 3.14159
    
    def υπολογίστε(self, x: float) -> float:
        """Greek method name."""
        return x * self.π

def 计算(数字: int) -> int:
    """Chinese function name."""
    return 数字 * 2

# Type annotations edge cases
def forward_ref(x: "MyClass") -> "AnotherClass":
    """Forward reference in quotes."""
    pass

def union_types(x: int | str | None) -> list[str] | None:
    """Python 3.10+ union syntax."""
    pass

# Comments everywhere
class CommentedClass:  # inline comment on class
    # Comment before method
    def method(  # comment after method name
        self,  # comment on self
        arg: int  # comment on arg
    ) -> None:  # comment on return
        # Comment in body
        pass  # inline comment in body

# Walrus operator in function
def uses_walrus(data: list[int]) -> int:
    """Uses assignment expression."""
    if (n := len(data)) > 0:
        return n
    return 0

# Match statement (Python 3.10+)
def pattern_matching(value):
    """Should not confuse parser."""
    match value:
        case 0:
            return "zero"
        case [x, y]:
            return f"list with {x}, {y}"
        case {"key": value}:
            return f"dict with key={value}"
        case _:
            return "other"

# Structural pattern matching in class
class Point:
    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y
    
    def __match_args__(self):
        return ("x", "y")

# Type guards
def is_string_list(val: list[object]) -> TypeGuard[list[str]]:
    """Type guard function."""
    return all(isinstance(x, str) for x in val)

# Positional-only and keyword-only parameters
def complex_params(pos_only, /, standard, *, kw_only: int = 0) -> None:
    """Mix of parameter types."""
    pass

# Very long lines
very_long_variable_name_that_might_exceed_typical_line_length_limits_in_some_editors_or_tools = {"key": "This is a very long string value that continues for quite a while to test how the parser handles extremely long lines"}

# Ellipsis usage
def abstract_method(self) -> None: ...
class AbstractClass:
    field: int = ...  # Ellipsis in type stub

# Raw strings and f-strings
pattern = r"(\w+)\s*=\s*(['\"])(.+?)\2"
formatted = f"""Multi-line
f-string with {pattern!r} representation"""

# Lambda with type annotations (not standard but some tools support)
typed_lambda: Callable[[int], int] = lambda x: x * 2

# Nested comprehensions
matrix = [[i * j for j in range(5)] for i in range(5)]
flattened = [x for row in matrix for x in row if x % 2 == 0]

# Generator expressions and comprehensions
gen = (x * x for x in range(10) if x % 2 == 0)
dict_comp = {k: v for k, v in zip(range(5), range(5, 10))}
set_comp = {x % 3 for x in range(20)}

# Context managers
class CustomContext:
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

# Yield and yield from
def generator_func():
    yield 1
    yield from range(2, 5)
    return "done"

# Async generators and comprehensions
async def async_gen():
    for i in range(5):
        yield i

async def async_comprehension():
    return [i async for i in async_gen()]