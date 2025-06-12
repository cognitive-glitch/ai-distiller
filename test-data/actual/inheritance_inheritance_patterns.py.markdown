# input/inheritance_patterns.py

## Structure

📥 **Import** from `abc` import `ABC`, `abstractmethod` <sub>L3</sub>
📥 **Import** from `typing` import `Protocol`, `runtime_checkable`, `List`, `Optional` <sub>L4</sub>
📥 **Import** `json` <sub>L5</sub>
🏛️ **Class** `Serializable` (extends `Protocol`) <sub>L9-19</sub>
  🔧 **Function** `to_dict`(`self`) → `dict` <sub>L12-15</sub>
  🔧 **Function** `from_dict`(`self`, `data`: `dict`) → `None` <sub>L16-19</sub>
🏛️ **Class** `JSONMixin` <sub>L21-36</sub>
  🔧 **Function** `to_json`(`self`) → `str` <sub>L24-29</sub>
  🔧 **Function** `from_json`(`self`, `json_str`: `str`) → `None` <sub>L30-36</sub>
🏛️ **Class** `LoggerMixin` <sub>L37-43</sub>
  🔧 **Function** `log`(`self`, `message`: `str`, `level`: `str`) → `None` <sub>L40-43</sub>
🏛️ **Class** `Animal` (extends `ABC`) <sub>L45-65</sub>
  🔧 **Function** `__init__`(`self`, `name`: `str`, `species`: `str`) <sub>L48-51</sub>
  🔧 **Function** `make_sound`(`self`) → `str` <sub>L53-56</sub>
  🔧 **Function** `move`(`self`) → `str` <sub>L58-61</sub>
  🔧 **Function** `describe`(`self`) → `str` <sub>L62-65</sub>
🏛️ **Class** `Dog` (extends `Animal`, `JSONMixin`, `LoggerMixin`) <sub>L67-104</sub>
  🔧 **Function** `__init__`(`self`, `name`: `str`, `breed`: `str`) <sub>L70-74</sub>
  🔧 **Function** `make_sound`(`self`) → `str` <sub>L75-79</sub>
  🔧 **Function** `move`(`self`) → `str` <sub>L80-83</sub>
  🔧 **Function** `add_trick`(`self`, `trick`: `str`) → `None` <sub>L84-88</sub>
  🔧 **Function** `to_dict`(`self`) → `dict` <sub>L89-97</sub>
  🔧 **Function** `from_dict`(`self`, `data`: `dict`) → `None` <sub>L98-104</sub>
🏛️ **Class** `Storage` (extends `ABC`) <sub>L106-128</sub>
  🔧 **Function** `save`(`self`, `key`: `str`, `value`) → `None` <sub>L110-113</sub>
  🔧 **Function** `load`(`self`, `key`: `str`) → `Optional[Any]` <sub>L115-118</sub>
  🔧 **Function** `delete`(`self`, `key`: `str`) → `bool` <sub>L120-123</sub>
  🔧 **Function** `exists`(`self`, `key`: `str`) → `bool` <sub>L125-128</sub>
🏛️ **Class** `MemoryStorage` (extends `Storage`) <sub>L130-153</sub>
  🔧 **Function** `__init__`(`self`) <sub>L133-135</sub>
  🔧 **Function** `save`(`self`, `key`: `str`, `value`) → `None` <sub>L136-139</sub>
  🔧 **Function** `load`(`self`, `key`: `str`) → `Optional[Any]` <sub>L140-143</sub>
  🔧 **Function** `delete`(`self`, `key`: `str`) → `bool` <sub>L144-150</sub>
  🔧 **Function** `exists`(`self`, `key`: `str`) → `bool` <sub>L151-153</sub>
