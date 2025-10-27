"""
Notification services for sending alerts to users.
Demonstrates inheritance and use of properties.
"""
from abc import ABC, abstractmethod

class BaseNotifier(ABC):
    """Abstract base for all notifiers."""
    @abstractmethod
    def send(self, message: str) -> None:
        raise NotImplementedError

class EmailNotifier(BaseNotifier):
    """Sends notifications via email."""

    def __init__(self, smtp_host: str, port: int, from_address: str):
        self._smtp_host = smtp_host
        self._port = port
        self.from_address = from_address
        self._connection = None # Represents a mock connection object

    @property
    def connection_info(self) -> str:
        """Returns a string with the current SMTP connection details."""
        return f"{self._smtp_host}:{self._port}"

    def _connect(self):
        """Internal method to establish a connection."""
        print(f"Connecting to {self.connection_info}...")
        self._connection = "CONNECTED" # Simulate connection

    def send(self, message: str) -> None:
        """Connects and sends an email."""
        if not self._connection:
            self._connect()
        print(f"Sending email from {self.from_address}: '{message}'")