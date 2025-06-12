"""Complex Python module with various constructs."""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import Dict, List, Optional, Union

# Enum definition
class Status(Enum):
    PENDING = "pending"
    ACTIVE = "active"
    COMPLETED = "completed"

# Dataclass
@dataclass
class Task:
    id: int
    title: str
    status: Status = Status.PENDING
    
    def complete(self):
        self.status = Status.COMPLETED

# Abstract base class
class BaseProcessor(ABC):
    """Abstract processor base class."""
    
    @abstractmethod
    def process(self, data: Dict) -> Dict:
        """Process data."""
        pass
    
    def validate(self, data: Dict) -> bool:
        """Validate data."""
        return "id" in data and "value" in data

# Concrete implementation
class DataProcessor(BaseProcessor):
    """Concrete data processor."""
    
    def __init__(self, config: Optional[Dict] = None):
        self.config = config or {}
        self._cache: Dict[str, Union[str, int, float]] = {}
    
    def process(self, data: Dict) -> Dict:
        """Process data with caching."""
        key = data.get("id", "unknown")
        
        if key in self._cache:
            return {"result": self._cache[key], "cached": True}
        
        result = self._transform(data)
        self._cache[key] = result
        return {"result": result, "cached": False}
    
    def _transform(self, data: Dict) -> Union[str, int, float]:
        """Internal transformation method."""
        value = data.get("value", 0)
        if isinstance(value, str):
            return value.upper()
        elif isinstance(value, (int, float)):
            return value * 2
        return str(value)

# Nested classes
class Container:
    """Container with nested classes."""
    
    class InnerClass:
        """Inner class example."""
        
        def __init__(self, name: str):
            self.name = name
        
        class DeepNested:
            """Deeply nested class."""
            pass
    
    def create_inner(self, name: str) -> InnerClass:
        return self.InnerClass(name)

# Global function with decorators
def cached_function(key: str) -> str:
    """Example of a decorated function."""
    return f"cached_{key}"

# Type alias
ProcessorType = Union[BaseProcessor, DataProcessor]

# Module-level variable
DEFAULT_CONFIG = {
    "timeout": 30,
    "retries": 3
}