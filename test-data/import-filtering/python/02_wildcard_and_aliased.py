# Test Pattern 2: Wildcard and Aliased Imports
# Tests from module import * and import module as alias

from math import *
import json as jsn
from datetime import date as dt_date
import random
from typing import List as ListType

def calculate_and_log():
    """Uses sqrt from math.*, jsn alias, and dt_date alias"""
    result = sqrt(16)  # Using sqrt from wildcard import
    print(f"Square root: {result}")
    
    # Using sin from wildcard import
    angle = sin(pi / 2)
    
    data = {"key": "value"}
    print(jsn.dumps(data))  # Using aliased json
    
    today = dt_date.today()  # Using aliased date
    print(f"Today's date: {today}")
    
    # Using type alias
    numbers: ListType[int] = [1, 2, 3]
    
    # random is not used anywhere