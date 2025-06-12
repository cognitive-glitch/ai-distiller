"""Test cases for async/await syntax and context sensitivity."""

import asyncio
from typing import AsyncIterator, AsyncContextManager

# Basic async function
async def basic_async():
    return "Hello, async!"

# Async function with await
async def fetch_data(url: str):
    # Simulating async operation
    await asyncio.sleep(1)
    return f"Data from {url}"

# Multiple awaits
async def process_multiple():
    result1 = await fetch_data("api/users")
    result2 = await fetch_data("api/posts")
    return result1, result2

# Async for loop
async def async_generator():
    for i in range(5):
        await asyncio.sleep(0.1)
        yield i

async def use_async_for():
    results = []
    async for value in async_generator():
        results.append(value * 2)
    return results

# Async with statement
class AsyncResource:
    async def __aenter__(self):
        await asyncio.sleep(0.1)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await asyncio.sleep(0.1)

async def use_async_with():
    async with AsyncResource() as resource:
        return "Resource acquired"

# Async comprehensions
async def async_comprehension():
    # Async list comprehension
    results = [x async for x in async_generator()]
    
    # Async set comprehension
    unique = {x async for x in async_generator() if x % 2 == 0}
    
    # Async dict comprehension
    mapping = {x: x**2 async for x in async_generator()}
    
    return results, unique, mapping

# Nested async contexts
async def nested_async():
    async with AsyncResource() as resource1:
        async with AsyncResource() as resource2:
            async for value in async_generator():
                if value > 2:
                    break
            return value

# Async class methods
class AsyncClass:
    async def async_method(self):
        return await fetch_data("method/data")
    
    @classmethod
    async def async_classmethod(cls):
        return "Async class method"
    
    @staticmethod
    async def async_staticmethod():
        return "Async static method"

# Async properties (using descriptor protocol)
class AsyncProperty:
    def __init__(self, func):
        self.func = func
    
    def __get__(self, obj, type=None):
        return self.func(obj)

class WithAsyncProperty:
    @AsyncProperty
    async def data(self):
        return await fetch_data("property/data")

# ERROR CASES - These should be syntax errors

# ERROR: await outside async function
def sync_with_await():
    # await fetch_data("invalid")  # SyntaxError: 'await' outside async function
    pass

# ERROR: async for outside async function
def sync_with_async_for():
    # async for x in async_generator():  # SyntaxError
    #     pass
    pass

# ERROR: async with outside async function
def sync_with_async_with():
    # async with AsyncResource():  # SyntaxError
    #     pass
    pass

# ERROR: yield in async function without async generator
# async def bad_async_yield():
#     yield 1  # This is actually valid - creates async generator

# Valid: async generator
async def valid_async_generator():
    for i in range(3):
        await asyncio.sleep(0.1)
        yield i

# Complex example: all async features combined
async def complex_async_example():
    """Demonstrates various async features."""
    results = []
    
    # Async with and async for
    async with AsyncResource() as resource:
        async for item in async_generator():
            # Await in expression
            processed = await fetch_data(f"item/{item}")
            results.append(processed)
            
            # Conditional await
            if item > 2:
                extra = await fetch_data("extra")
                results.append(extra)
    
    # Async comprehension with await
    filtered = [
        await fetch_data(f"final/{x}")
        async for x in async_generator()
        if x % 2 == 0
    ]
    
    return results + filtered

# Async context manager as class
class AsyncContextManager:
    async def __aenter__(self):
        self.resource = await fetch_data("context/enter")
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await fetch_data("context/exit")
        if exc_type:
            print(f"Exception: {exc_val}")
        return False  # Don't suppress exceptions

# Top-level await (valid in notebooks/REPL, but not in regular scripts)
# This would be a syntax error in a regular Python file:
# result = await fetch_data("top-level")  # SyntaxError in script

# Async lambda (not directly supported)
# async_lambda = async lambda x: await fetch_data(x)  # SyntaxError

# But you can work around it:
async def make_async_func(x):
    return await fetch_data(x)

# Type hints with async
async def typed_async(
    data: list[str],
    processor: AsyncIterator[int]
) -> AsyncIterator[str]:
    async for item in processor:
        if item < len(data):
            yield data[item]