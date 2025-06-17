# C++ Language Support

AI Distiller provides comprehensive support for C++ codebases using the tree-sitter parser, with support for modern C++ features including templates, concepts (C++20), and advanced type system constructs.

## Overview

C++ support in AI Distiller captures the essential structure of C++ code including templates, multiple inheritance, operator overloading, and modern language features. The distilled output preserves C++'s powerful type system while optimizing for AI consumption.

## Supported C++ Constructs

### Core Language Features

| Construct | Support Level | Notes |
|-----------|--------------|-------|
| **Classes** | ✅ Full | Including inheritance, nested classes |
| **Structs** | ✅ Full | Treated similar to classes |
| **Templates** | ✅ Full | Class and function templates |
| **Functions** | ✅ Full | Free functions, methods, operators |
| **Namespaces** | ✅ Full | Including nested namespaces |
| **Enums** | ✅ Full | enum and enum class |
| **Typedefs/Using** | ✅ Full | Type aliases |
| **Unions** | ⚠️ Partial | Basic support |
| **Operators** | ✅ Full | Overloaded operators |
| **Friend declarations** | ⚠️ Partial | Shown as comments |

### Advanced Features

| Feature | Support Level | Notes |
|---------|--------------|-------|
| **Virtual functions** | ✅ Full | Including pure virtual (= 0) |
| **Constructors/Destructors** | ✅ Full | All forms including move/copy |
| **Templates** | ✅ Full | With constraints |
| **Return types** | ✅ Full | Including trailing return types partially |
| **Parameter types** | ✅ Full | With defaults |
| **Const correctness** | ✅ Full | const methods and parameters |
| **Reference types** | ✅ Full | & and && (rvalue references) |
| **C++20 Concepts** | ⚠️ Partial | Shown as special comments |
| **Attributes** | ⚠️ Partial | [[nodiscard]], [[deprecated]] |

## Recent Fixes (2025-06-15)

1. **Missing return types** (✅ Fixed)
   - **Issue**: Function return types were not captured correctly
   - **Fix**: Improved return type extraction by collecting type tokens before function declarator
   - **Impact**: Return types now properly displayed for all functions

2. **Implementation always shown** (✅ Fixed)
   - **Issue**: Function implementations were displayed even with `--implementation=0`
   - **Fix**: Removed implementation display from C++ formatter
   - **Impact**: Proper respect for stripping options

3. **Parameter types** (✅ Fixed)
   - **Issue**: Function parameter types were sometimes missing
   - **Fix**: Enhanced parameter extraction in tree-sitter processor
   - **Impact**: Full parameter signatures preserved

## Key Features

### 1. **Template Support**

```cpp
// Input
template<typename T, typename Allocator = std::allocator<T>>
class Vector {
public:
    using value_type = T;
    using size_type = std::size_t;
    
    Vector() noexcept(noexcept(Allocator())) = default;
    explicit Vector(size_type count, const T& value = T());
    
    template<typename InputIt>
    Vector(InputIt first, InputIt last);
    
    void push_back(const T& value);
    void push_back(T&& value);
    
    template<typename... Args>
    void emplace_back(Args&&... args);
    
private:
    T* data_;
    size_type size_;
    size_type capacity_;
};
```

```
// Output (default)
template<typename T, typename Allocator = std::allocator<T>>
class Vector {
    using value_type = T;
    using size_type = std::size_t;
    
    Vector() noexcept(noexcept(Allocator())) = default;
    explicit Vector(size_type count, const T& value = T());
    
    template<typename InputIt>
    Vector(InputIt first, InputIt last);
    
    void push_back(const T& value);
    void push_back(T&& value);
    
    template<typename... Args>
    void emplace_back(Args&&... args);
};
```

### 2. **Modern C++ Features**

```cpp
// Input
class Widget {
public:
    Widget() = default;
    Widget(const Widget&) = delete;
    Widget(Widget&&) noexcept = default;
    Widget& operator=(const Widget&) = delete;
    Widget& operator=(Widget&&) noexcept = default;
    ~Widget() = default;
    
    [[nodiscard]] bool process() const;
    void update() && = delete;  // Only lvalues
    
    explicit operator bool() const { return valid_; }
    
private:
    bool valid_ = false;
};
```

```
// Output
class Widget {
    Widget() = default;
    Widget(const Widget&) = delete;
    Widget(Widget&&) noexcept = default;
    Widget& operator=(const Widget&) = delete;
    Widget& operator=(Widget&&) noexcept = default;
    ~Widget() = default;
    
    bool process() const;
    void update() && = delete;
    
    explicit operator bool() const;
};
```

### 3. **C++20 Concepts (Limited Support)**

```cpp
// Input
template<typename T>
concept Arithmetic = std::is_arithmetic_v<T>;

template<typename T>
    requires Arithmetic<T>
T add(T a, T b) {
    return a + b;
}
```

```
// Output
// C++20 Concept: template<typename T>
// concept Arithmetic = std::is_arithmetic_v<T>;

template<typename T>
T add(T a, T b);
```

## Visibility Rules

C++ visibility in AI Distiller:
- **public**: Accessible from anywhere
- **protected**: Accessible in class and derived classes
- **private**: Accessible only within the class
- **(default)**: private in class, public in struct

## Known Limitations

1. **Trailing return types**: `auto func() -> decltype(...)` not fully supported
2. **C++20 Concepts**: Shown as comments rather than parsed
3. **Template specializations**: Partial specializations may not be fully captured
4. **Preprocessor directives**: Macros and conditional compilation ignored
5. **Complex template metaprogramming**: SFINAE patterns simplified

## Best Practices

### 1. **Clear Template Constraints**

Use readable template constraints:

```cpp
// Good - Clear intent
template<typename T>
    requires std::is_integral_v<T>
class Counter {
    T count = 0;
};

// Also good - Traditional style
template<typename T, 
         typename = std::enable_if_t<std::is_integral_v<T>>>
class Counter {
    T count = 0;
};
```

### 2. **Explicit Access Specifiers**

Group members by access level:

```cpp
class Service {
public:
    // Public interface
    void start();
    void stop();
    
protected:
    // Extension points
    virtual void onStart();
    virtual void onStop();
    
private:
    // Implementation details
    void cleanup();
    bool running_ = false;
};
```

### 3. **RAII and Smart Pointers**

Use modern memory management:

```cpp
class ResourceManager {
public:
    using ResourcePtr = std::unique_ptr<Resource>;
    
    ResourcePtr acquire(const std::string& name);
    void release(ResourcePtr resource);
    
private:
    std::unordered_map<std::string, ResourcePtr> resources_;
};
```

## Output Examples

<details>
<summary>Complex Template Class</summary>

Input:
```cpp
template<typename Key, typename Value, 
         typename Hash = std::hash<Key>,
         typename KeyEqual = std::equal_to<Key>,
         typename Allocator = std::allocator<std::pair<const Key, Value>>>
class HashMap {
public:
    using key_type = Key;
    using mapped_type = Value;
    using value_type = std::pair<const Key, Value>;
    using iterator = /* implementation-defined */;
    
    HashMap() = default;
    explicit HashMap(std::size_t bucket_count);
    
    template<typename InputIt>
    HashMap(InputIt first, InputIt last);
    
    iterator find(const Key& key);
    const_iterator find(const Key& key) const;
    
    template<typename K>
    iterator find(const K& key);
    
    std::pair<iterator, bool> insert(const value_type& value);
    
    template<typename... Args>
    std::pair<iterator, bool> emplace(Args&&... args);
    
    Value& operator[](const Key& key);
    Value& operator[](Key&& key);
    
private:
    struct Node {
        value_type data;
        Node* next;
    };
    
    std::vector<Node*> buckets_;
    std::size_t size_ = 0;
    Hash hasher_;
    KeyEqual key_equal_;
};
```

Output (default):
```
template<typename Key, typename Value, typename Hash = std::hash<Key>, typename KeyEqual = std::equal_to<Key>, typename Allocator = std::allocator<std::pair<const Key, Value>>>
class HashMap {
    using key_type = Key;
    using mapped_type = Value;
    using value_type = std::pair<const Key, Value>;
    using iterator = /* implementation-defined */;
    
    HashMap() = default;
    explicit HashMap(std::size_t bucket_count);
    
    template<typename InputIt>
    HashMap(InputIt first, InputIt last);
    
    iterator find(const Key& key);
    const_iterator find(const Key& key) const;
    
    template<typename K>
    iterator find(const K& key);
    
    std::pair<iterator, bool> insert(const value_type& value);
    
    template<typename... Args>
    std::pair<iterator, bool> emplace(Args&&... args);
    
    Value& operator[](const Key& key);
    Value& operator[](Key&& key);
};
```

</details>

## Integration Tips

### For API Documentation
```bash
# Extract public API only
aid include/ --private=0 --protected=0 --internal=0 --implementation=0

# Include protected members for library users
aid include/ --private=0 --internal=0 --implementation=0
```

### For Code Review
```bash
# Full structure without implementations
aid src/ --implementation=0 --format text
```

### For Architecture Analysis
```bash
# High-level overview
aid . --private=0 --protected=0 --internal=0 --implementation=0 --comments=0
```

## Future Improvements

- Full C++20 concepts support
- Better trailing return type handling
- Template specialization tracking
- Constexpr/consteval support
- Module support (C++20)
- Coroutine support

## Contributing

C++ support is actively maintained. Key areas for contribution:
- C++20/23 features
- Template metaprogramming patterns
- Complex inheritance scenarios
- Preprocessor handling

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development setup.