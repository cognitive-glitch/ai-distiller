# test_02_intermediate.py
from typing import Optional

class Car:
    """
    Represents a car with basic attributes and methods.

    This class tests the parsing of a standard class structure, including
    the initializer, instance attributes, public/protected/private methods,
    and a property for controlled attribute access.
    """
    def __init__(self, make: str, model: str, year: int):
        self.make = make
        self.model = model
        self.year = year
        self._engine_status: str = "off"  # Protected attribute
        self.__mileage: int = 0  # Private attribute

    @property
    def mileage(self) -> int:
        """Read-only property to access the car's mileage."""
        return self.__mileage

    def start_engine(self) -> None:
        """Starts the car's engine."""
        if self._engine_status == "off":
            self._engine_status = "on"
            self.__log_activity("Engine started.")

    def drive(self, distance: int) -> None:
        """Drives the car for a certain distance."""
        if self._engine_status == "on" and distance > 0:
            self.__mileage += distance
            self.__log_activity(f"Drove {distance} miles.")

    def _get_diagnostic_code(self) -> Optional[str]:
        """A protected method to get a diagnostic code."""
        return "P0300" if self.__mileage > 100000 else None

    def __log_activity(self, message: str) -> None:
        """A private method for internal logging."""
        print(f"[LOG] {self.make} {self.model}: {message}")

if __name__ == "__main__":
    my_car = Car("Toyota", "Corolla", 2021)
    my_car.start_engine()
    my_car.drive(150)
    print(f"Car mileage: {my_car.mileage}")