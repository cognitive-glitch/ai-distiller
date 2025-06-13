"""
Complex Python test file for AI Distiller functional testing
Includes classes, inheritance, protocols, decorators, async/await, type hints
"""

from typing import (
    Any, Dict, List, Optional, Union, Generic, TypeVar, Protocol,
    Callable, Awaitable, AsyncGenerator, Iterator
)
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, auto
from contextlib import asynccontextmanager
import asyncio
import logging
from functools import wraps, cached_property

# Type variables
T = TypeVar('T')
U = TypeVar('U', bound='Comparable')

# Protocols
class Comparable(Protocol):
    """Protocol for comparable objects"""
    def __lt__(self, other: Any) -> bool: ...
    def __eq__(self, other: Any) -> bool: ...

class Repository(Protocol[T]):
    """Generic repository protocol"""
    async def save(self, entity: T) -> T: ...
    async def find_by_id(self, id: int) -> Optional[T]: ...
    async def delete(self, id: int) -> None: ...

# Enums
class Status(Enum):
    ACTIVE = auto()
    INACTIVE = auto()
    PENDING = auto()

class Priority(Enum):
    LOW = 1
    MEDIUM = 2
    HIGH = 3

# Data classes
@dataclass
class User:
    """User data model"""
    id: int
    name: str
    email: Optional[str] = None
    roles: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def __post_init__(self):
        if not self.name.strip():
            raise ValueError("Name cannot be empty")
    
    @property
    def display_name(self) -> str:
        return f"{self.name} ({self.email})" if self.email else self.name

@dataclass(frozen=True)
class UserInfo:
    """Immutable user info"""
    name: str
    age: int
    email: str
    
    def __post_init__(self):
        if self.age < 0:
            raise ValueError("Age cannot be negative")
    
    @property
    def is_adult(self) -> bool:
        return self.age >= 18

# Decorators
def log_calls(func: Callable) -> Callable:
    """Decorator to log function calls"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        logging.info(f"Calling {func.__name__} with args: {args}, kwargs: {kwargs}")
        return func(*args, **kwargs)
    return wrapper

def async_retry(max_attempts: int = 3):
    """Decorator for async function retry logic"""
    def decorator(func: Callable[..., Awaitable[T]]) -> Callable[..., Awaitable[T]]:
        @wraps(func)
        async def wrapper(*args, **kwargs) -> T:
            for attempt in range(max_attempts):
                try:
                    return await func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_attempts - 1:
                        raise
                    await asyncio.sleep(2 ** attempt)
            raise RuntimeError("Should not reach here")
        return wrapper
    return decorator

# Abstract base classes
class BaseService(ABC, Generic[T]):
    """Abstract base service class"""
    
    def __init__(self, logger: logging.Logger):
        self._logger = logger
        self._cache: Dict[int, T] = {}
    
    @abstractmethod
    async def validate(self, entity: T) -> bool:
        """Validate entity"""
        pass
    
    @abstractmethod
    async def process(self, entity: T) -> T:
        """Process entity"""
        pass
    
    @cached_property
    def cache_size(self) -> int:
        return len(self._cache)
    
    def _log_operation(self, operation: str) -> None:
        """Private logging method"""
        self._logger.debug(f"Executing operation: {operation}")

class UserService(BaseService[User]):
    """Concrete user service implementation"""
    
    def __init__(self, logger: logging.Logger, email_service: 'EmailService'):
        super().__init__(logger)
        self._email_service = email_service
        self.__private_data: Dict[str, Any] = {}
    
    async def validate(self, user: User) -> bool:
        """Validate user data"""
        return len(user.name.strip()) > 0 and user.id > 0
    
    @async_retry(max_attempts=3)
    async def process(self, user: User) -> User:
        """Process user with retry logic"""
        if not await self.validate(user):
            raise ValueError("User validation failed")
        
        # Simulate async processing
        await asyncio.sleep(0.1)
        self._cache[user.id] = user
        
        if user.email:
            await self._email_service.send_welcome(user.email)
        
        return user
    
    @log_calls
    def get_user_by_id(self, user_id: int) -> Optional[User]:
        """Get user from cache"""
        return self._cache.get(user_id)
    
    async def get_users_by_status(self, status: Status) -> List[User]:
        """Get users filtered by status"""
        # Complex filtering logic
        return [
            user for user in self._cache.values()
            if user.metadata.get('status') == status
        ]
    
    @staticmethod
    def create_default() -> 'UserService':
        """Create default service instance"""
        logger = logging.getLogger(__name__)
        email_service = MockEmailService()
        return UserService(logger, email_service)
    
    @classmethod
    def from_config(cls, config: Dict[str, Any]) -> 'UserService':
        """Create service from configuration"""
        logger = logging.getLogger(config.get('logger_name', __name__))
        email_service = MockEmailService()
        return cls(logger, email_service)
    
    def __private_method(self) -> None:
        """Private method (name mangling)"""
        pass

# Protocols for dependency injection
class EmailService(Protocol):
    """Email service protocol"""
    async def send_welcome(self, email: str) -> None: ...
    async def send_notification(self, email: str, message: str) -> None: ...

class CacheService(Protocol[T]):
    """Generic cache service protocol"""
    async def get(self, key: str) -> Optional[T]: ...
    async def set(self, key: str, value: T, ttl: Optional[int] = None) -> None: ...
    async def delete(self, key: str) -> bool: ...

# Concrete implementations
class MockEmailService:
    """Mock email service for testing"""
    
    async def send_welcome(self, email: str) -> None:
        """Send welcome email"""
        print(f"Sending welcome email to {email}")
        await asyncio.sleep(0.05)  # Simulate network delay
    
    async def send_notification(self, email: str, message: str) -> None:
        """Send notification email"""
        print(f"Sending notification to {email}: {message}")
        await asyncio.sleep(0.05)

class InMemoryCache(Generic[T]):
    """In-memory cache implementation"""
    
    def __init__(self):
        self._data: Dict[str, T] = {}
        self._ttl: Dict[str, float] = {}
    
    async def get(self, key: str) -> Optional[T]:
        """Get value from cache"""
        if key in self._data:
            return self._data[key]
        return None
    
    async def set(self, key: str, value: T, ttl: Optional[int] = None) -> None:
        """Set value in cache"""
        self._data[key] = value
        if ttl:
            import time
            self._ttl[key] = time.time() + ttl
    
    async def delete(self, key: str) -> bool:
        """Delete value from cache"""
        if key in self._data:
            del self._data[key]
            self._ttl.pop(key, None)
            return True
        return False

# Async context manager
class DatabaseTransaction:
    """Async database transaction context manager"""
    
    def __init__(self, connection):
        self._connection = connection
        self._transaction = None
    
    async def __aenter__(self):
        self._transaction = await self._connection.begin()
        return self._transaction
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if exc_type is None:
            await self._transaction.commit()
        else:
            await self._transaction.rollback()

# Complex async generator
async def process_users_batch(
    users: List[User],
    batch_size: int = 10
) -> AsyncGenerator[List[User], None]:
    """Process users in batches"""
    for i in range(0, len(users), batch_size):
        batch = users[i:i + batch_size]
        # Simulate async processing
        await asyncio.sleep(0.1)
        yield batch

# Function with complex type annotations
def transform_data(
    data: Dict[str, Union[str, int, List[str]]],
    transformers: Dict[str, Callable[[Any], Any]]
) -> Dict[str, Any]:
    """Transform data using provided transformers"""
    result = {}
    for key, value in data.items():
        if key in transformers:
            result[key] = transformers[key](value)
        else:
            result[key] = value
    return result

# Generator function
def fibonacci(n: int) -> Iterator[int]:
    """Generate fibonacci sequence"""
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b

# Property descriptors
class ValidatedProperty:
    """Descriptor for validated properties"""
    
    def __init__(self, validator: Callable[[Any], bool]):
        self.validator = validator
        self.name = None
    
    def __set_name__(self, owner, name):
        self.name = f"_{name}"
    
    def __get__(self, instance, owner):
        if instance is None:
            return self
        return getattr(instance, self.name, None)
    
    def __set__(self, instance, value):
        if not self.validator(value):
            raise ValueError(f"Invalid value for {self.name}")
        setattr(instance, self.name, value)

class ValidationExample:
    """Example class using property descriptor"""
    
    name = ValidatedProperty(lambda x: isinstance(x, str) and len(x) > 0)
    age = ValidatedProperty(lambda x: isinstance(x, int) and x >= 0)
    
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age

# Metaclass example
class SingletonMeta(type):
    """Singleton metaclass"""
    _instances: Dict[type, Any] = {}
    
    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super().__call__(*args, **kwargs)
        return cls._instances[cls]

class ConfigManager(metaclass=SingletonMeta):
    """Configuration manager using singleton pattern"""
    
    def __init__(self):
        self._config: Dict[str, Any] = {}
    
    def get(self, key: str, default: Any = None) -> Any:
        return self._config.get(key, default)
    
    def set(self, key: str, value: Any) -> None:
        self._config[key] = value

# Main execution
if __name__ == "__main__":
    async def main():
        """Main async function"""
        # Create services
        user_service = UserService.create_default()
        
        # Create test user
        user = User(id=1, name="John Doe", email="john@example.com")
        
        # Process user
        processed_user = await user_service.process(user)
        print(f"Processed user: {processed_user}")
        
        # Test batch processing
        users = [User(id=i, name=f"User {i}") for i in range(1, 6)]
        async for batch in process_users_batch(users, batch_size=2):
            print(f"Processing batch: {[u.name for u in batch]}")
    
    asyncio.run(main())