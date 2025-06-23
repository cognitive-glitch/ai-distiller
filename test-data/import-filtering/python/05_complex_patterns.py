# Test Pattern 5: Complex Import Patterns
# Tests __init__.py re-exports, multiline imports, relative imports, and edge cases

from __future__ import annotations
import asyncio
import concurrent.futures
from concurrent.futures import (
    ThreadPoolExecutor,
    ProcessPoolExecutor,
    as_completed,
    wait,
    FIRST_COMPLETED
)
from . import utils
from ..core import base_handler
from ...lib import (
    helper as lib_helper,
    validator,
    formatter
)

# Star import from __init__.py
from .submodule import *  # Assume this exports process_data, validate_input

# Import with trailing comma (valid Python)
from typing import (
    Any,
    Callable,
    TypeVar,
    Generic,
)

# Import used in decorator
import functools
import time

# Import used in exception handling
import pickle
import json

# Import for metaclass
import abc

# Side-effect import
import matplotlib.pyplot as plt  # This sets up matplotlib backend

def timer(func: Callable) -> Callable:
    """Timer decorator using functools and time"""
    @functools.wraps(func)
    def wrapper(*args: Any, **kwargs: Any) -> Any:
        start = time.time()
        result = func(*args, **kwargs)
        end = time.time()
        print(f"{func.__name__} took {end - start} seconds")
        return result
    return wrapper

class AbstractProcessor(metaclass=abc.ABCMeta):
    """Abstract base class using abc module"""
    
    @abc.abstractmethod
    def process(self, data: Any) -> Any:
        pass

@timer
async def async_processor():
    """Uses asyncio and concurrent.futures"""
    async with asyncio.TaskGroup() as tg:
        # Using ThreadPoolExecutor
        with ThreadPoolExecutor(max_workers=5) as executor:
            futures = []
            for i in range(10):
                future = executor.submit(process_data, i)  # From * import
                futures.append(future)
            
            # Using as_completed
            for future in as_completed(futures):
                result = future.result()
                
    # Using ProcessPoolExecutor and FIRST_COMPLETED
    with ProcessPoolExecutor() as executor:
        future1 = executor.submit(validate_input, "test")  # From * import
        future2 = executor.submit(lib_helper.help_function, "arg")
        
        done, pending = wait([future1, future2], return_when=FIRST_COMPLETED)

def serialize_data(obj: Any) -> bytes:
    """Try pickle first, fall back to json"""
    try:
        return pickle.dumps(obj)
    except pickle.PickleError:
        # Fall back to JSON
        return json.dumps(obj).encode()

# Using relative imports
base_handler.register(AbstractProcessor)
validator.validate_schema({})
formatter.format_output("")

# TypeVar and Generic are used
T = TypeVar('T')

class Container(Generic[T]):
    def __init__(self, value: T):
        self.value = value

# matplotlib.pyplot is a side-effect import (should be kept)
# utils is imported but not directly used (might be used by * imports)
# Any and Callable are used, TypeVar and Generic are used