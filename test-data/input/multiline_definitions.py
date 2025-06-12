"""Test cases for multiline definitions and edge cases."""

# Standard multiline function definition
def my_complex_function(
    arg1: str,
    arg2: int = 123,
    arg3: Optional[List[str]] = None
) -> bool:
    """My docstring."""
    return True

# Multiline class definition
class MyDerivedClass(
    parent_module.ParentClass,
    another.MixIn
):
    pass

# Function with comments in definition
def function_with_comments(
    arg1: str,  # Inline comment
    
    # A whole line comment
    arg2: int
):
    pass

# Function with nested types in parameters
def function_with_nested_types(
    arg1: Dict[str, Union[int, float]]
) -> Callable[[int], str]:
    pass

# Class with complex inheritance
class ComplexClass(
    Base[T],
    Protocol,
    metaclass=ABCMeta
):
    """Complex class with generics and metaclass."""
    
    def method_one(self) -> None:
        pass

# Async function with decorators
@decorator_one
@decorator_two(param="value")
async def async_function(
    param: str
) -> AsyncIterator[int]:
    """Async function with multiple decorators."""
    yield 42

# Function with very long type annotations
def function_with_long_types(
    callback: Callable[[str, int, Optional[Dict[str, Any]]], Union[List[str], None]],
    data: Dict[str, Union[str, int, float, List[Union[str, int]]]]
) -> Tuple[Optional[str], List[Dict[str, Any]]]:
    """Function with complex nested type annotations."""
    pass