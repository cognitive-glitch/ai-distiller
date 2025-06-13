# test_01_basic.py
from typing import List

def process_data(items: List[str], repeat: int = 2) -> List[str]:
    """
    Processes a list of strings by repeating each item.

    This function serves as a basic test for function signature parsing,
    including type hints for arguments and return values, default
    argument values, and a standard docstring.

    Args:
        items: A list of strings to be processed.
        repeat: The number of times to repeat each item.

    Returns:
        A new list with each item repeated.
    """
    processed_items: List[str] = [item * repeat for item in items]
    return processed_items

if __name__ == "__main__":
    sample_data = ["a", "b", "c"]
    result = process_data(sample_data)
    print(f"Basic function result: {result}")