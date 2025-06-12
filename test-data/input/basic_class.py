"""Basic class with various method types."""

class Person:
    """A person class with different visibility methods."""
    
    def __init__(self, name: str, age: int):
        """Initialize a person."""
        self.name = name
        self.age = age
        self._id = None  # Private attribute
    
    def get_info(self) -> str:
        """Get person information."""
        return f"{self.name} is {self.age} years old"
    
    def _calculate_id(self) -> int:
        """Private method to calculate ID."""
        return hash(self.name) % 1000
    
    @property
    def id(self) -> int:
        """Get the person's ID."""
        if self._id is None:
            self._id = self._calculate_id()
        return self._id
    
    @staticmethod
    def is_adult(age: int) -> bool:
        """Check if age represents an adult."""
        return age >= 18
    
    @classmethod
    def from_string(cls, data: str) -> 'Person':
        """Create person from string."""
        name, age = data.split(',')
        return cls(name, int(age))