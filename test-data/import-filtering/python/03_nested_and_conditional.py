# Test Pattern 3: Nested/Sub-module Imports and Conditional Imports
# Tests imports from nested packages and conditional imports

import requests.exceptions
from urllib.parse import urlparse, quote
import platform
import sys

# Conditional imports based on Python version
if sys.version_info.major >= 3 and sys.version_info.minor >= 8:
    from typing import Literal
else:
    from typing_extensions import Literal

# Conditional import based on platform
if platform.system() == "Windows":
    import winreg
else:
    import pwd

# Import only used in function
def setup_logging():
    import logging
    logging.basicConfig(level=logging.INFO)
    logging.info("Logging setup completed")

def make_request(url: str):
    """Uses requests.exceptions and urlparse, but not quote"""
    try:
        # Simulating usage
        parsed = urlparse(url)
        if not parsed.scheme:
            raise requests.exceptions.InvalidURL("No scheme in URL")
    except requests.exceptions.RequestException as e:
        print(f"Request error: {e}")

    # Using Literal from conditional import
    mode: Literal["read", "write"] = "read"

    # Platform-specific code
    if platform.system() == "Windows":
        # Using winreg on Windows
        print("Windows system")
        # winreg usage would go here
    else:
        # Using pwd on Unix
        print("Unix system")
        # pwd usage would go here

# Call the function to ensure logging import is used
setup_logging()
make_request("https://example.com")

# quote from urllib.parse is not used