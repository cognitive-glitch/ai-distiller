"""Test cases for nested structures and indentation handling."""

import os
from typing import List, Optional

class TopLevelClass:
    """Top level class with nested structures."""
    
    class_variable: int = 42
    
    def __init__(self, name: str):
        self.name = name
        self._private_var = None
    
    class NestedClass:
        """Class nested inside another class."""
        
        def nested_method(self) -> str:
            return "nested"
        
        class DoublyNestedClass:
            """Two levels of nesting."""
            pass
    
    def outer_method(self) -> None:
        """Method with nested function."""
        
        def inner_function(x: int) -> int:
            """Function inside method."""
            return x * 2
        
        # Local class inside method
        class LocalClass:
            pass
        
        result = inner_function(5)
    
    @property
    def computed_property(self) -> str:
        """Property with complex logic."""
        return f"{self.name}_computed"
    
    @staticmethod
    def static_method() -> None:
        """Static method."""
        pass
    
    @classmethod
    def class_method(cls) -> "TopLevelClass":
        """Class method returning instance."""
        return cls("default")

def top_level_function():
    """Top level function after class."""
    
    # Nested function at module level
    def helper():
        pass
    
    return helper()

# Another top level class
class AnotherTopLevel:
    """Should be sibling of TopLevelClass."""
    
    def method(self):
        # Multi-line string that looks like code
        code_string = """
        def fake_function():
            # This is inside a string
            class FakeClass:
                pass
        """
        return code_string

# Edge case: function that looks like it's indented but isn't
def looks_indented():
        """Unusual but valid indentation in docstring.
        def this_is_not_a_function():
            pass
        """
        pass

# Decorators with nested structures
@decorator_factory(
    param=lambda x: x * 2
)
class DecoratedClass:
    pass