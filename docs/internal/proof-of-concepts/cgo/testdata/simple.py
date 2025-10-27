def hello():
    """Say hello to the world."""
    print("Hello, world!")


class Config:
    """Configuration class for the application."""

    def __init__(self, port=8080, host="localhost"):
        self.port = port
        self.host = host

    def get_url(self):
        """Return the full URL for the server."""
        return f"http://{self.host}:{self.port}"


if __name__ == "__main__":
    hello()
    config = Config()
    print(config.get_url())