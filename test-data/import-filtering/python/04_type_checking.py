# Test Pattern 4: Type Checking Imports and __all__ exports
# Tests TYPE_CHECKING, type-only imports, and __all__ influence

from typing import List, Dict, Optional, TYPE_CHECKING, Union
from collections import Counter
import xml.etree.ElementTree as ET

if TYPE_CHECKING:
    from .models import User, Product  # Only for type checking
    from datetime import datetime  # Only for type checking

# Import used in docstring
import requests

# Import used in type comment
import pandas  # type: ignore

# Define exports
__all__ = ["process_items", "analyze_data", "XMLProcessor"]

def process_items(items: List[int]) -> Dict[str, Optional[str]]:
    """Process a list of items and return statistics.

    Args:
        items: List of integers to process

    Returns:
        Dictionary with total count and item frequency

    Note:
        Uses :class:`requests.Session` for API calls if needed.
    """
    counts = Counter(items)
    return {"total": str(len(items)), "counts": str(counts)}

def analyze_data(data: Union[List, "pandas.DataFrame"]) -> str:
    """Analyze data which can be a list or pandas DataFrame.

    Args:
        data: Either a list or a :class:`pandas.DataFrame`
    """
    if hasattr(data, 'shape'):  # It's a DataFrame
        return f"DataFrame with shape {data.shape}"
    return f"List with {len(data)} items"

class XMLProcessor:
    """Process XML data.

    This class uses :mod:`xml.etree.ElementTree` for parsing.
    """

    def parse(self, content: str) -> Optional["ET.Element"]:
        """Parse XML content."""
        # ET is not actually used in implementation
        return None

if TYPE_CHECKING:
    def type_check_only_function(user: User, when: datetime) -> Product:
        """This function only exists for type checking."""
        pass

# Counter is used, ET is referenced in docstring but not used in code
# requests is referenced in docstring
# pandas is referenced in type comment and docstring
# User, Product, datetime are only in TYPE_CHECKING block