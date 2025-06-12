"""Test decorators, type hints, and metadata."""

from functools import lru_cache, wraps
from typing import Callable, Any, TypeVar, ParamSpec
from dataclasses import dataclass, field
from enum import Enum, auto

P = ParamSpec('P')
R = TypeVar('R')

def timer(func: Callable[P, R]) -> Callable[P, R]:
    """Time function execution."""
    @wraps(func)
    def wrapper(*args: P.args, **kwargs: P.kwargs) -> R:
        import time
        start = time.time()
        result = func(*args, **kwargs)
        print(f"{func.__name__} took {time.time() - start:.2f}s")
        return result
    return wrapper

def deprecated(message: str) -> Callable[[Callable[P, R]], Callable[P, R]]:
    """Mark function as deprecated."""
    def decorator(func: Callable[P, R]) -> Callable[P, R]:
        @wraps(func)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> R:
            import warnings
            warnings.warn(f"{func.__name__} is deprecated: {message}", 
                         DeprecationWarning, stacklevel=2)
            return func(*args, **kwargs)
        return wrapper
    return decorator

class Status(Enum):
    """Status enumeration."""
    PENDING = auto()
    PROCESSING = auto()
    COMPLETED = auto()
    FAILED = auto()

@dataclass
class Task:
    """Task with metadata."""
    id: int
    name: str
    status: Status = Status.PENDING
    tags: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        """Validate task after initialization."""
        if not self.name:
            raise ValueError("Task name cannot be empty")

class Processor:
    """Processor with decorated methods."""
    
    @lru_cache(maxsize=128)
    def expensive_operation(self, x: int) -> int:
        """Cached expensive operation."""
        return x ** 2
    
    @timer
    @deprecated("Use new_process instead")
    def process(self, data: Any) -> Any:
        """Process data (deprecated)."""
        return self._internal_process(data)
    
    def _internal_process(self, data: Any) -> Any:
        """Internal processing logic."""
        return data
    
    @staticmethod
    @timer
    def validate(data: Dict[str, Any]) -> bool:
        """Validate data structure."""
        return all(k in data for k in ['id', 'value'])
    
    @property
    def cache_info(self) -> Dict[str, Any]:
        """Get cache information."""
        return self.expensive_operation.cache_info()._asdict()