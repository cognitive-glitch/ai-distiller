"""Example Python module to demonstrate new test format"""

class Calculator:
    """A simple calculator class"""

    def __init__(self):
        self._history = []

    def add(self, a: float, b: float) -> float:
        """Add two numbers"""
        result = a + b
        self._history.append(f"add({a}, {b}) = {result}")
        return result

    def _clear_history(self):
        """Private method to clear history"""
        self._history = []

    @property
    def history(self) -> list:
        """Get calculation history"""
        return self._history.copy()