"""Edge cases and unusual Python constructs."""

# Unicode identifiers
def 你好(名字: str) -> str:
    """Chinese function name."""
    return f"你好, {名字}!"

π = 3.14159
Δ = 0.001

# Very long lines
def function_with_very_long_signature(parameter_one: str, parameter_two: int, parameter_three: float, parameter_four: bool, parameter_five: list, parameter_six: dict, parameter_seven: tuple = (1, 2, 3)) -> Optional[Dict[str, Union[int, float, str, List[Any]]]]:
    """Function with extremely long signature that should be handled properly."""
    pass

# Nested functions and closures
def outer_function(x: int) -> Callable[[int], int]:
    """Creates a closure."""
    multiplier = 2
    
    def inner_function(y: int) -> int:
        """Inner function using closure."""
        nonlocal multiplier
        
        def deeply_nested(z: int) -> int:
            """Deeply nested function."""
            return x + y + z * multiplier
        
        return deeply_nested(10)
    
    return inner_function

# Class with many special methods
class MagicClass:
    """Class demonstrating many magic methods."""
    
    def __init__(self, value: Any):
        self._value = value
    
    def __str__(self) -> str:
        return f"MagicClass({self._value})"
    
    def __repr__(self) -> str:
        return f"MagicClass(value={self._value!r})"
    
    def __eq__(self, other: Any) -> bool:
        if isinstance(other, MagicClass):
            return self._value == other._value
        return False
    
    def __lt__(self, other: 'MagicClass') -> bool:
        return self._value < other._value
    
    def __len__(self) -> int:
        return len(str(self._value))
    
    def __getitem__(self, key: int) -> str:
        return str(self._value)[key]
    
    def __setitem__(self, key: int, value: str) -> None:
        raise NotImplementedError("Immutable")
    
    def __delitem__(self, key: int) -> None:
        raise NotImplementedError("Immutable")
    
    def __iter__(self) -> Iterator[str]:
        return iter(str(self._value))
    
    def __contains__(self, item: str) -> bool:
        return item in str(self._value)
    
    def __call__(self, *args, **kwargs) -> Any:
        return self._value(*args, **kwargs) if callable(self._value) else self._value
    
    def __enter__(self) -> 'MagicClass':
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb) -> bool:
        return False
    
    def __getattr__(self, name: str) -> Any:
        return getattr(self._value, name)
    
    def __setattr__(self, name: str, value: Any) -> None:
        if name.startswith('_'):
            super().__setattr__(name, value)
        else:
            setattr(self._value, name, value)
    
    def __delattr__(self, name: str) -> None:
        if not name.startswith('_'):
            delattr(self._value, name)

# Metaclass
class SingletonMeta(type):
    """Singleton metaclass."""
    _instances = {}
    
    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super().__call__(*args, **kwargs)
        return cls._instances[cls]

class Singleton(metaclass=SingletonMeta):
    """Singleton class using metaclass."""
    
    def __init__(self):
        self.value = 42

# Generator functions and expressions
def fibonacci(n: int) -> Generator[int, None, None]:
    """Generate Fibonacci sequence."""
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b

# List comprehensions with complex conditions
complex_list = [
    x * y 
    for x in range(10) 
    if x % 2 == 0 
    for y in range(10) 
    if y % 3 == 0
    if x * y > 10
]

# Dict and set comprehensions
word_lengths = {word: len(word) for word in "hello world".split()}
unique_chars = {char for word in "hello world".split() for char in word}

# Walrus operator (Python 3.8+)
if (n := len(complex_list)) > 5:
    print(f"List has {n} elements")

# Type annotations edge cases
ComplexType = Dict[str, List[Union[int, Tuple[str, Optional[float]]]]]
RecursiveType = List['RecursiveType']

# Global and nonlocal
global_var = 100

def modify_global():
    """Modifies global variable."""
    global global_var
    global_var += 1

# Empty class
class Empty:
    """Empty class with just docstring."""
    pass

# Class with slots
class Optimized:
    """Class using slots for optimization."""
    __slots__ = ['x', 'y', 'z']
    
    def __init__(self, x: int, y: int, z: int):
        self.x = x
        self.y = y
        self.z = z

# Async functions
async def async_function(url: str) -> str:
    """Async function example."""
    import asyncio
    await asyncio.sleep(0.1)
    return f"Fetched: {url}"

# Context manager
class MyContext:
    """Custom context manager."""
    
    async def __aenter__(self):
        """Async enter."""
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async exit."""
        pass