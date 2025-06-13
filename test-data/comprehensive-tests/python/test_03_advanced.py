# test_03_advanced.py
import time
from functools import wraps

def timing_decorator(func):
    """A custom decorator to measure function execution time."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.perf_counter()
        result = func(*args, **kwargs)
        end_time = time.perf_counter()
        print(f"Function '{func.__name__}' executed in {end_time - start_time:.4f}s")
        return result
    return wrapper

class Device:
    """A base class for electronic devices."""
    def __init__(self, device_id: str):
        self._device_id = device_id
        self.is_on = False

    def power_on(self):
        self.is_on = True
        return f"Device {self._device_id} is now ON."

class SmartPhone(Device):
    """
    A smartphone that inherits from Device.

    Tests inheritance, method overriding, super() calls, and various
    decorators (@staticmethod, @classmethod, custom).
    """
    _MAX_APPS = 100

    def __init__(self, device_id: str, os: str):
        super().__init__(device_id)
        self.os = os
        self.installed_apps = []

    @timing_decorator
    def install_app(self, app_name: str):
        """Installs an application on the phone."""
        if len(self.installed_apps) < self._MAX_APPS:
            self.installed_apps.append(app_name)
            time.sleep(0.1) # Simulate installation time

    @staticmethod
    def is_portable() -> bool:
        """A static method, not bound to instance or class."""
        return True

    @classmethod
    def create_default_phone(cls, device_id: str) -> 'SmartPhone':
        """A class method to create a phone with a default OS."""
        return cls(device_id, "Android")

    def power_on(self):
        """Overrides the parent method to add more functionality."""
        base_message = super().power_on()
        return f"{base_message} Booting {self.os}..."

if __name__ == "__main__":
    phone = SmartPhone.create_default_phone("SP-123")
    print(phone.power_on())
    phone.install_app("AI Distiller Companion")
    print(f"Is portable? {SmartPhone.is_portable()}")