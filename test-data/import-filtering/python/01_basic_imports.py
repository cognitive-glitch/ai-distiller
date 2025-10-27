# Test Pattern 1: Basic Imports with Clear Usage
# Tests simple direct imports of modules and specific objects from modules

import os
from collections import deque
import sys
import math
from pathlib import Path

def process_data():
    """Function that uses os and deque, but not sys or math"""
    print(os.path.join("a", "b"))
    q = deque([1, 2, 3])
    print(q)

    # Path is used here
    p = Path("/tmp")
    print(p.exists())

# No usage of sys or math in this file