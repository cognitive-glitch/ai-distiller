"""Test cases for complex type hints and annotations."""

from typing import (
    List, Dict, Tuple, Set, Optional, Union, Any,
    Callable, TypeVar, Generic, Protocol, Literal,
    TypedDict, Final, ClassVar, Annotated, TypeAlias,
    overload, cast, NoReturn, Never
)
from collections.abc import Sequence, Mapping, Iterator
from types import NotImplementedType
import sys

# Basic type hints
simple_var: int = 42
simple_list: list[int] = [1, 2, 3]
simple_dict: dict[str, int] = {"a": 1, "b": 2}

# Optional and Union
maybe_string: Optional[str] = None
string_or_int: Union[str, int] = "hello"
modern_union: str | int | None = 42  # Python 3.10+ syntax

# Complex nested types
nested_structure: dict[str, list[tuple[int, str]]] = {
    "items": [(1, "one"), (2, "two")]
}

# Callable types
simple_callback: Callable[[int, str], bool]
complex_callback: Callable[[int, ...], Optional[str]]  # Variable args
no_arg_callback: Callable[[], None]

# TypeVar and Generic
T = TypeVar('T')
K = TypeVar('K')
V = TypeVar('V')

class Container(Generic[T]):
    def __init__(self, value: T) -> None:
        self.value = value
    
    def get(self) -> T:
        return self.value

# Bounded TypeVar
Numeric = TypeVar('Numeric', int, float)

def add_numbers(a: Numeric, b: Numeric) -> Numeric:
    return a + b

# Protocol (structural subtyping)
class Drawable(Protocol):
    def draw(self) -> None: ...

class Resizable(Protocol):
    def resize(self, width: int, height: int) -> None: ...

# Intersection using Protocol
class Widget(Drawable, Resizable, Protocol):
    name: str

# TypedDict
class PersonDict(TypedDict):
    name: str
    age: int
    email: Optional[str]

# TypedDict with total=False
class PartialPersonDict(TypedDict, total=False):
    nickname: str
    phone: str

# Literal types
Mode = Literal["read", "write", "append"]
Status = Literal[200, 404, 500]

def open_file(path: str, mode: Mode) -> None:
    pass

# Type aliases
Vector = list[float]
Matrix = list[Vector]
JsonValue = Union[str, int, float, bool, None, Dict[str, 'JsonValue'], List['JsonValue']]

# Modern type alias syntax (Python 3.12+)
# type Point = tuple[float, float]  # New syntax

# Forward references
class Node:
    def __init__(self, value: int, next_node: Optional['Node'] = None):
        self.value = value
        self.next = next_node

# Annotated types
from typing import Annotated
PositiveInt = Annotated[int, "Must be positive"]
Email = Annotated[str, "Valid email format"]

# Final and ClassVar
class Constants:
    MAX_SIZE: Final[int] = 100
    instances: ClassVar[int] = 0
    
    def __init__(self):
        Constants.instances += 1

# Function overloading
@overload
def process(data: str) -> str: ...

@overload
def process(data: int) -> int: ...

@overload
def process(data: list[T]) -> list[T]: ...

def process(data: Union[str, int, list[Any]]) -> Union[str, int, list[Any]]:
    if isinstance(data, str):
        return data.upper()
    elif isinstance(data, int):
        return data * 2
    else:
        return data + data

# Complex function signature
def complex_function(
    required: str,
    /,  # Positional-only
    normal: int,
    *args: float,
    keyword_only: bool,
    another: Optional[Dict[str, List[Tuple[int, str]]]] = None,
    **kwargs: Any
) -> Union[Tuple[str, ...], NoReturn]:
    if not keyword_only:
        sys.exit(1)  # NoReturn
    return tuple(str(arg) for arg in args)

# Self type
from typing import Self  # Python 3.11+

class Chainable:
    def method(self) -> Self:
        return self

# ParamSpec and Concatenate (Python 3.10+)
from typing import ParamSpec, Concatenate

P = ParamSpec('P')
R = TypeVar('R')

def decorator(func: Callable[P, R]) -> Callable[Concatenate[int, P], R]:
    def wrapper(x: int, *args: P.args, **kwargs: P.kwargs) -> R:
        return func(*args, **kwargs)
    return wrapper

# TypeGuard (Python 3.10+)
from typing import TypeGuard

def is_str_list(val: list[object]) -> TypeGuard[list[str]]:
    return all(isinstance(x, str) for x in val)

# Never type (Python 3.11+)
def assert_never(value: Never) -> NoReturn:
    raise AssertionError(f"Unhandled value: {value}")

# Variadic generics using TypeVarTuple (Python 3.11+)
from typing import TypeVarTuple, Unpack

Shape = TypeVarTuple('Shape')

class Tensor(Generic[Unpack[Shape]]):
    pass

# Type hints in class definition
class AdvancedClass(Generic[T]):
    class_var: ClassVar[str] = "shared"
    instance_var: int
    _private_var: Optional[str]
    
    def __init__(self, value: T) -> None:
        self.value: Final[T] = value
        self._cache: dict[str, Any] = {}
    
    @property
    def cached_value(self) -> T:
        return self._cache.get("value", self.value)
    
    @cached_value.setter
    def cached_value(self, value: T) -> None:
        self._cache["value"] = value

# Complex real-world example
from typing import ContextManager
from abc import ABC, abstractmethod

class DataProcessor(ABC, Generic[T, V]):
    """Abstract data processor with complex type constraints."""
    
    @abstractmethod
    def process(
        self,
        data: Sequence[T],
        transform: Callable[[T], V],
        *,
        filter_func: Optional[Callable[[T], bool]] = None,
        error_handler: Callable[[T, Exception], Union[V, None]] = lambda x, e: None
    ) -> Iterator[V]:
        """Process data with transformation and optional filtering."""
        ...
    
    @abstractmethod
    def batch_process(
        self,
        data: Mapping[K, Sequence[T]],
        processors: dict[K, Callable[[T], V]]
    ) -> dict[K, list[V]]:
        """Batch process with different processors per key."""
        ...