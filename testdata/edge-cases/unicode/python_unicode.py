"""
Edge case: Python with Unicode identifiers and special characters.
Tests parser's Unicode handling.
"""

# Unicode in class names (Python 3+ supports this)
class Пользователь:  # Russian: User
    """User class with Cyrillic name."""

    def __init__(self, имя: str):  # Russian: name
        self.имя = имя

    def получить_имя(self) -> str:  # Russian: get_name
        return self.имя


# Emoji in identifiers (valid in Python 3)
class 🚀Rocket:
    """Class with emoji in name."""

    def __init__(self):
        self.速度 = 0  # Chinese: speed

    def 加速(self):  # Chinese: accelerate
        self.速度 += 1


# Arabic identifiers
class مستخدم:  # Arabic: User
    def __init__(self, اسم: str):  # Arabic: name
        self.اسم = اسم


# Greek identifiers
class Χρήστης:  # Greek: User
    def __init__(self, όνομα: str):  # Greek: name
        self.όνομα = όνομα


# Unicode in strings and comments
def process_text(text: str) -> str:
    """Process text with various Unicode characters.

    Supports: 中文, العربية, Русский, Ελληνικά, עברית
    """
    # Emoji in comments: 🎉 🚀 ✨ 🔥
    return f"Processed: {text} ✓"


# Zero-width characters (invisible but valid)
class User​Manager:  # Contains zero-width space (U+200B)
    """Manager with zero-width space in name."""
    pass


# Right-to-left markers
def ‏מתודה‏():  # Hebrew with RTL markers
    """Function with RTL markers."""
    pass
