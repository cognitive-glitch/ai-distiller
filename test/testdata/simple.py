"""Simple Python module for testing."""

import os
from typing import List, Optional

class Calculator:
    """A simple calculator class."""
    
    def __init__(self, precision: int = 2):
        self.precision = precision
        self._history: List[str] = []
    
    def add(self, a: float, b: float) -> float:
        """Add two numbers."""
        result = a + b
        self._history.append(f"{a} + {b} = {result}")
        return round(result, self.precision)
    
    def multiply(self, a: float, b: float) -> float:
        """Multiply two numbers."""
        result = a * b
        self._history.append(f"{a} * {b} = {result}")
        return round(result, self.precision)
    
    def get_history(self) -> List[str]:
        """Get calculation history."""
        return self._history.copy()

def main():
    """Main entry point."""
    calc = Calculator()
    print(calc.add(1.5, 2.5))
    print(calc.multiply(3, 4))

if __name__ == "__main__":
    main()