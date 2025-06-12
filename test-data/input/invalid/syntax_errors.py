"""Test file with various syntax errors to test error recovery."""

# Valid import
import os

# Invalid import (missing module)
import

# Valid function
def valid_function():
    return True

# Function with unclosed parenthesis
def broken_function(x, y:
    return x + y

# Valid class
class ValidClass:
    def method(self):
        pass

# Class without colon
class MissingColon
    def method(self):
        pass

# Function with keyword name
def class():  # 'class' is a keyword
    pass

# Another valid function to ensure recovery
def another_valid():
    return "Parser should recover"

# Mixed indentation
def mixed_indent():
	    value = 1  # Tab then spaces
    return value

# Invalid class name starting with number
class 123Invalid:
    pass

# Valid import after errors
from sys import argv

# Final valid class
class FinalClass:
    """Should be parsed despite previous errors."""
    
    def __init__(self):
        self.value = 42