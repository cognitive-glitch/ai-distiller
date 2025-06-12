"""Test cases for various import styles including multiline."""

# Standard imports
import os
import sys
import numpy as np

# Multiple imports on one line
import json, csv, xml.etree.ElementTree as ET

# From imports with aliases
from collections import defaultdict as dd, Counter as Cnt
from pathlib import Path

# Multiline import with parentheses
from typing import (
    List,
    Dict,
    Optional,
    Union,
    TypeVar,
    Generic,
    Protocol,
    Callable,
    # Comment inside import
    AsyncIterator,
    Tuple,
)

# Multiline import with backslash (edge case)
from package.subpackage.module import FirstClass, \
    SecondClass, ThirdClass

# Relative imports
from . import local_module
from .. import parent_module
from ...sibling import specific_function
from .submodule import (
    ClassA,
    ClassB as B,
    function_one,
    CONSTANT_VALUE
)

# Star imports
from math import *
from statistics import (
    mean,
    median,
    *  # Not common but valid syntax
)

# Import with very long module name
from very.long.deeply.nested.package.structure.with.many.levels import (
    SomeVeryLongClassNameThatMightBreakFormatting as ShortName,
    AnotherLongClassName,
    YetAnotherOne
)

# Try to confuse parser with similar keywords
import important_module
from implementation import something
class ImportantClass:
    """This is not an import."""
    pass