"""
A plugin manager that discovers, registers, and executes plugins.
This demonstrates composition, custom decorators, and advanced typing.
"""
from typing import Protocol, List, Dict, Callable, Any

class Plugin(Protocol):
    """A protocol defining the interface for a valid plugin."""
    name: str
    def execute(self, data: Dict[str, Any]) -> None:
        ...

_registry: Dict[str, Plugin] = {}

def register_plugin(name: str) -> Callable[[Callable[[], Plugin]], None]:
    """A decorator to register a plugin creation function."""
    def decorator(plugin_creator: Callable[[], Plugin]) -> None:
        print(f"Registering plugin: {name}")
        # We store the creator, not the instance, for lazy loading.
        _registry[name] = plugin_creator
    return decorator

class DataProcessingPlugin:
    name = "data_processor"
    def execute(self, data: Dict[str, Any]) -> None:
        print(f"Processing data: {data.keys()}")

@register_plugin(name=DataProcessingPlugin.name)
def create_data_plugin() -> Plugin:
    return DataProcessingPlugin()


class PluginManager:
    """Manages the lifecycle of plugins."""
    def __init__(self):
        # Manager holds instances, registry holds creators
        self._instances: Dict[str, Plugin] = {}

    def activate(self, name: str) -> None:
        if name in _registry and name not in self._instances:
            plugin_creator = _registry[name]
            self._instances[name] = plugin_creator()

    def run_all(self, data: Dict[str, Any]):
        if not self._instances:
            print("No active plugins to run.")
            return
        for name, instance in self._instances.items():
            print(f"--- Running plugin: {name} ---")
            instance.execute(data)