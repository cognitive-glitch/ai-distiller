"""Test cases for error recovery - parser should continue after errors."""

# Valid import
import os

# Missing colon on class definition
class MissingColon
    pass  # Parser should recover and continue

# Valid function
def valid_function():
    """This should be parsed correctly."""
    return True

# Missing closing parenthesis
def broken_function(arg1, arg2
    # Parser should handle this gracefully
    pass

# Valid class after error
class ValidClass:
    """This class should be parsed despite previous errors."""
    
    def method_one(self):
        return "works"

# Syntax error in import
from invalid import (
    something,
    # Missing closing parenthesis

# Parser should recover for this valid import
import sys

# Invalid indentation
def another_function():
  return 1  # 2 spaces instead of 4
 return 2   # Wrong indentation - parser should handle

# Valid function after indentation error
def final_function():
    """Should be parsed correctly."""
    return True

# Incomplete type annotation
def function_with_broken_type(arg: List[str) -> None:
    pass

# Valid class at the end
class FinalClass:
    """Should still be parsed."""
    pass