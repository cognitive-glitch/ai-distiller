# input/basic_class.py

## Structure

🏛️ **Class** `Person` <sub>L3-36</sub>
  🔧 **Function** `__init__`(`self`, `name`: `str`, `age`: `int`) <sub>L6-11</sub>
    ```
            """Initialize a person."""
            self.name = name
            self.age = age
            self._id = None  # Private attribute
        
    ```
  🔧 **Function** `get_info`(`self`) → `str` <sub>L12-15</sub>
    ```
            """Get person information."""
            return f"{self.name} is {self.age} years old"
        
    ```
  🔧 **Function** `_calculate_id` _private_(`self`) → `int` <sub>L16-19</sub>
    ```
            """Private method to calculate ID."""
            return hash(self.name) % 1000
        
    ```
  🔧 **Function** `id`(`self`) → `int` <sub>L21-26</sub>
    ```
            """Get the person's ID."""
            if self._id is None:
                self._id = self._calculate_id()
            return self._id
        
    ```
  🔧 **Function** `is_adult`(`age`: `int`) → `bool` <sub>L28-31</sub>
    ```
            """Check if age represents an adult."""
            return age >= 18
        
    ```
  🔧 **Function** `from_string`(`cls`, `data`: `str`) → `'Person'` <sub>L33-36</sub>
    ```
            """Create person from string."""
            name, age = data.split(',')
            return cls(name, int(age))
    ```
