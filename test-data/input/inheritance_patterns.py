"""Test inheritance, abstract classes, mixins, and protocols."""

from abc import ABC, abstractmethod
from typing import Protocol, runtime_checkable, List, Optional
import json

# Protocol (structural subtyping)
@runtime_checkable
class Serializable(Protocol):
    """Protocol for serializable objects."""
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        ...
    
    def from_dict(self, data: dict) -> None:
        """Load from dictionary."""
        ...

# Mixin classes
class JSONMixin:
    """Mixin for JSON serialization."""
    
    def to_json(self) -> str:
        """Serialize to JSON."""
        if hasattr(self, 'to_dict'):
            return json.dumps(self.to_dict())
        raise NotImplementedError("Object must implement to_dict")
    
    def from_json(self, json_str: str) -> None:
        """Deserialize from JSON."""
        if hasattr(self, 'from_dict'):
            self.from_dict(json.loads(json_str))
        else:
            raise NotImplementedError("Object must implement from_dict")

class LoggerMixin:
    """Mixin for logging capabilities."""
    
    def log(self, message: str, level: str = "INFO") -> None:
        """Log a message."""
        print(f"[{level}] {self.__class__.__name__}: {message}")

# Abstract base class
class Animal(ABC):
    """Abstract animal class."""
    
    def __init__(self, name: str, species: str):
        self.name = name
        self.species = species
    
    @abstractmethod
    def make_sound(self) -> str:
        """Make the animal's sound."""
        pass
    
    @abstractmethod
    def move(self) -> str:
        """Describe how the animal moves."""
        pass
    
    def describe(self) -> str:
        """Describe the animal."""
        return f"{self.name} is a {self.species}"

# Multiple inheritance
class Dog(Animal, JSONMixin, LoggerMixin):
    """Concrete dog class with multiple inheritance."""
    
    def __init__(self, name: str, breed: str):
        super().__init__(name, "Canis familiaris")
        self.breed = breed
        self._tricks: List[str] = []
    
    def make_sound(self) -> str:
        """Dogs bark."""
        self.log("Making sound")
        return "Woof!"
    
    def move(self) -> str:
        """Dogs run."""
        return "Running on four legs"
    
    def add_trick(self, trick: str) -> None:
        """Teach the dog a new trick."""
        self._tricks.append(trick)
        self.log(f"Learned new trick: {trick}")
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            'name': self.name,
            'species': self.species,
            'breed': self.breed,
            'tricks': self._tricks
        }
    
    def from_dict(self, data: dict) -> None:
        """Load from dictionary."""
        self.name = data['name']
        self.species = data['species']
        self.breed = data['breed']
        self._tricks = data.get('tricks', [])

# Interface-like abstract class
class Storage(ABC):
    """Abstract storage interface."""
    
    @abstractmethod
    def save(self, key: str, value: Any) -> None:
        """Save a value."""
        pass
    
    @abstractmethod
    def load(self, key: str) -> Optional[Any]:
        """Load a value."""
        pass
    
    @abstractmethod
    def delete(self, key: str) -> bool:
        """Delete a value."""
        pass
    
    @abstractmethod
    def exists(self, key: str) -> bool:
        """Check if key exists."""
        pass

# Concrete implementation
class MemoryStorage(Storage):
    """In-memory storage implementation."""
    
    def __init__(self):
        self._data: Dict[str, Any] = {}
    
    def save(self, key: str, value: Any) -> None:
        """Save to memory."""
        self._data[key] = value
    
    def load(self, key: str) -> Optional[Any]:
        """Load from memory."""
        return self._data.get(key)
    
    def delete(self, key: str) -> bool:
        """Delete from memory."""
        if key in self._data:
            del self._data[key]
            return True
        return False
    
    def exists(self, key: str) -> bool:
        """Check existence in memory."""
        return key in self._data