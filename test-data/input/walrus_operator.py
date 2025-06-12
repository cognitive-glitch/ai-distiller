"""Test cases for assignment expressions (walrus operator := ) - Python 3.8+"""

# Valid uses of walrus operator

# In if statements
def walrus_in_if(data):
    if (n := len(data)) > 10:
        return f"Large dataset with {n} items"
    return f"Small dataset with {n} items"  # n is still in scope

# In while loops
def walrus_in_while(lines):
    results = []
    while (line := lines.readline()) != "":
        results.append(line.strip())
    return results

# In list comprehensions
def walrus_in_comprehension(values):
    return [y for x in values if (y := x * 2) > 10]

# In dict comprehensions
def walrus_dict_comp(items):
    return {name: normalized for item in items 
            if (name := item.get("name")) and (normalized := name.lower())}

# Multiple walrus operators
def multiple_walrus(a, b):
    if (x := a * 2) > 10 and (y := b * 3) < 20:
        return x + y
    return 0

# Walrus in function arguments
def walrus_in_call(data):
    return process_data(cleaned := data.strip(), len(cleaned))

def process_data(data, length):
    return f"Processing {length} chars"

# Walrus in lambda (valid but unusual)
compute = lambda x: (y := x + 1, y * 2)[-1]

# Walrus in conditional expression
def walrus_in_ternary(value):
    return "positive" if (n := value) > 0 else f"non-positive: {n}"

# Nested walrus operators
def nested_walrus(data):
    if (outer := len(data)) > 0:
        if (inner := data[0]) and (processed := inner.strip()):
            return f"First item '{processed}' in {outer} items"
    return "Empty"

# Walrus with unpacking
def walrus_unpacking(pairs):
    results = []
    for pair in pairs:
        if (coords := pair) and len(coords) == 2:
            x, y = coords
            results.append((x, y, x + y))
    return results

# ERROR CASES - These should be syntax errors

# ERROR: Walrus as statement (not allowed)
def walrus_as_statement():
    x := 5  # SyntaxError: invalid syntax

# ERROR: Walrus at module level (in this context)
# y := 10  # SyntaxError: invalid syntax (would need to be in a valid expression context)

# ERROR: Walrus in function definition default
# def bad_default(x := 10):  # SyntaxError: invalid syntax
#     return x

# ERROR: Walrus without parentheses in certain contexts
def walrus_without_parens(data):
    # This is actually valid in Python 3.8+
    if n := len(data):
        return n
    
    # But this would be an error:
    # return n := len(data)  # SyntaxError: invalid syntax

# Complex valid example
def analyze_data(records):
    """Complex but valid use of multiple walrus operators."""
    valid_records = [
        processed
        for record in records
        if (name := record.get("name"))
        and (age := record.get("age"))
        and age >= 18
        and (processed := {
            "name": name.upper(),
            "age": age,
            "category": "adult" if age >= 21 else "young adult"
        })
    ]
    
    if (count := len(valid_records)) > 0:
        return {
            "count": count,
            "records": valid_records,
            "average_age": sum(r["age"] for r in valid_records) / count
        }
    return {"count": 0, "records": [], "average_age": 0}