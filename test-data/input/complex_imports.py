"""Complex import patterns testing."""

# Standard library imports
import os
import sys
from pathlib import Path
from typing import List, Dict, Optional, Union, TypeVar, Generic
from collections import defaultdict, OrderedDict
from datetime import datetime, timedelta
from abc import ABC, abstractmethod

# Relative imports (in a package context)
# from . import sibling_module
# from ..parent import something
# from ...grandparent import other_thing

# Import with alias
import numpy as np
from pandas import DataFrame as DF

# Star imports (should be handled)
from math import *

# Multi-line imports
from some_long_module_name import (
    VeryLongClassName,
    AnotherLongClassName,
    YetAnotherOne,
    AndOneMore
)

# Type variable
T = TypeVar('T')

class Container(Generic[T]):
    """Generic container class."""
    
    def __init__(self):
        self._items: List[T] = []
    
    def add(self, item: T) -> None:
        """Add item to container."""
        self._items.append(item)