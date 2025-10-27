#include <iostream>
#include <string>
#include <vector>
#include <memory>
#include <algorithm>
#include <functional>
#include <type_traits>
#include <utility>
#include <tuple>
#include <variant>
#include <optional>
#include <any>
#include <thread>
#include <future>
#include <chrono>

/**
 * @brief Complex C++ features including template metaprogramming, CRTP,
 *        variadic templates, perfect forwarding, and C++17 features
 */

/**
 * @brief Compile-time computation using constexpr
 * @param n Input number
 * @return Factorial of n
 */
constexpr long long factorial(int n) {
    return (n <= 1) ? 1 : n * factorial(n - 1);
}

/**
 * @brief Template metaprogramming: compile-time type list
 * @tparam Types List of types
 */
template<typename... Types>
struct TypeList {
    static constexpr size_t size = sizeof...(Types);
};

/**
 * @brief Get type at index from TypeList
 * @tparam Index Index to get
 * @tparam T First type
 * @tparam Rest Remaining types
 */
template<size_t Index, typename T, typename... Rest>
struct TypeAt {
    using type = typename TypeAt<Index - 1, Rest...>::type;
};

/**
 * @brief Specialization for index 0
 * @tparam T First type
 * @tparam Rest Remaining types
 */
template<typename T, typename... Rest>
struct TypeAt<0, T, Rest...> {
    using type = T;
};

/**
 * @brief Helper alias for TypeAt
 */
template<size_t Index, typename... Types>
using TypeAt_t = typename TypeAt<Index, Types...>::type;

/**
 * @brief SFINAE helper for checking if type is callable
 * @tparam F Function type
 * @tparam Args Argument types
 */
template<typename F, typename... Args>
struct is_callable {
private:
    template<typename U>
    static auto test(int) -> decltype(std::declval<U>()(std::declval<Args>()...), std::true_type{});

    template<typename>
    static std::false_type test(...);

public:
    static constexpr bool value = decltype(test<F>(0))::value;
};

/**
 * @brief Helper variable template for is_callable
 */
template<typename F, typename... Args>
constexpr bool is_callable_v = is_callable<F, Args...>::value;

/**
 * @brief CRTP base class for adding comparison operators
 * @tparam Derived The derived class type
 */
template<typename Derived>
class Comparable {
public:
    /**
     * @brief Inequality operator
     * @param other Other object to compare
     * @return true if not equal
     */
    bool operator!=(const Derived& other) const {
        return !static_cast<const Derived*>(this)->operator==(other);
    }

    /**
     * @brief Greater than operator
     * @param other Other object to compare
     * @return true if this > other
     */
    bool operator>(const Derived& other) const {
        return other < static_cast<const Derived&>(*this);
    }

    /**
     * @brief Less than or equal operator
     * @param other Other object to compare
     * @return true if this <= other
     */
    bool operator<=(const Derived& other) const {
        return !(static_cast<const Derived&>(*this) > other);
    }

    /**
     * @brief Greater than or equal operator
     * @param other Other object to compare
     * @return true if this >= other
     */
    bool operator>=(const Derived& other) const {
        return !(static_cast<const Derived&>(*this) < other);
    }

protected:
    /**
     * @brief Protected destructor to prevent deletion through base pointer
     */
    ~Comparable() = default;
};

/**
 * @brief Example class using CRTP
 */
class Point : public Comparable<Point> {
public:
    /**
     * @brief Constructor
     * @param x X coordinate
     * @param y Y coordinate
     */
    Point(double x, double y) : x_(x), y_(y) {}

    /**
     * @brief Equality operator
     * @param other Other point
     * @return true if points are equal
     */
    bool operator==(const Point& other) const {
        return std::abs(x_ - other.x_) < 1e-9 && std::abs(y_ - other.y_) < 1e-9;
    }

    /**
     * @brief Less than operator
     * @param other Other point
     * @return true if this point is "less" than other
     */
    bool operator<(const Point& other) const {
        return (x_ < other.x_) || (x_ == other.x_ && y_ < other.y_);
    }

    /**
     * @brief Get distance from origin
     * @return Distance value
     */
    double distance() const {
        return std::sqrt(x_ * x_ + y_ * y_);
    }

private:
    double x_;  ///< X coordinate
    double y_;  ///< Y coordinate
};

/**
 * @brief Variadic template for tuple-like operations
 * @tparam Args Types of arguments
 */
template<typename... Args>
class VariadicProcessor {
public:
    /**
     * @brief Constructor
     * @param args Values to store
     */
    explicit VariadicProcessor(Args... args) : data_(std::forward<Args>(args)...) {}

    /**
     * @brief Get element at index
     * @tparam Index Index to get
     * @return Reference to element
     */
    template<size_t Index>
    auto get() -> decltype(std::get<Index>(data_)) {
        return std::get<Index>(data_);
    }

    /**
     * @brief Apply function to all elements
     * @tparam F Function type
     * @param f Function to apply
     */
    template<typename F>
    void forEach(F&& f) {
        forEachImpl(std::forward<F>(f), std::index_sequence_for<Args...>{});
    }

    /**
     * @brief Get size of stored tuple
     * @return Number of elements
     */
    static constexpr size_t size() {
        return sizeof...(Args);
    }

private:
    std::tuple<Args...> data_;  ///< Stored data

    /**
     * @brief Implementation of forEach
     * @tparam F Function type
     * @tparam Indices Index sequence
     * @param f Function to apply
     * @param indices Index sequence
     */
    template<typename F, size_t... Indices>
    void forEachImpl(F&& f, std::index_sequence<Indices...>) {
        (f(std::get<Indices>(data_)), ...);  // C++17 fold expression
    }
};

/**
 * @brief Advanced SFINAE for perfect forwarding
 * @tparam Container Container type
 * @tparam T Element type
 */
template<typename Container, typename T>
auto insert_if_possible(Container& container, T&& value)
    -> decltype(container.insert(std::forward<T>(value)), void()) {
    container.insert(std::forward<T>(value));
}

/**
 * @brief Fallback for containers without insert
 * @tparam Container Container type
 * @tparam T Element type
 */
template<typename Container, typename T>
auto insert_if_possible(Container& container, T&& value)
    -> decltype(container.push_back(std::forward<T>(value)), void()) {
    container.push_back(std::forward<T>(value));
}

/**
 * @brief Template metaprogramming: compile-time string
 * @tparam N String length
 */
template<size_t N>
struct CompileTimeString {
    /**
     * @brief Constructor
     * @param str String literal
     */
    constexpr CompileTimeString(const char (&str)[N]) {
        std::copy_n(str, N, data);
    }

    char data[N];  ///< String data
};

/**
 * @brief Template deduction guide for CompileTimeString
 */
template<size_t N>
CompileTimeString(const char (&)[N]) -> CompileTimeString<N>;

/**
 * @brief Factory with variadic templates and perfect forwarding
 * @tparam T Type to create
 */
template<typename T>
class Factory {
public:
    /**
     * @brief Create instance with perfect forwarding
     * @tparam Args Constructor argument types
     * @param args Constructor arguments
     * @return Unique pointer to created instance
     */
    template<typename... Args>
    static std::unique_ptr<T> create(Args&&... args) {
        return std::make_unique<T>(std::forward<Args>(args)...);
    }

    /**
     * @brief Create instance with initialization from tuple
     * @tparam Tuple Tuple type
     * @param tuple Tuple of constructor arguments
     * @return Unique pointer to created instance
     */
    template<typename Tuple>
    static std::unique_ptr<T> createFromTuple(Tuple&& tuple) {
        return createFromTupleImpl(std::forward<Tuple>(tuple),
                                  std::make_index_sequence<std::tuple_size_v<std::decay_t<Tuple>>>{});
    }

private:
    /**
     * @brief Implementation of createFromTuple
     * @tparam Tuple Tuple type
     * @tparam Indices Index sequence
     * @param tuple Tuple of arguments
     * @param indices Index sequence
     * @return Unique pointer to created instance
     */
    template<typename Tuple, size_t... Indices>
    static std::unique_ptr<T> createFromTupleImpl(Tuple&& tuple, std::index_sequence<Indices...>) {
        return std::make_unique<T>(std::get<Indices>(std::forward<Tuple>(tuple))...);
    }
};

/**
 * @brief Visitor pattern implementation using std::variant
 */
using DataVariant = std::variant<int, double, std::string>;

/**
 * @brief Generic visitor for DataVariant
 */
struct DataVisitor {
    /**
     * @brief Visit int
     * @param value Integer value
     * @return Processed string
     */
    std::string operator()(int value) const {
        return "Integer: " + std::to_string(value);
    }

    /**
     * @brief Visit double
     * @param value Double value
     * @return Processed string
     */
    std::string operator()(double value) const {
        return "Double: " + std::to_string(value);
    }

    /**
     * @brief Visit string
     * @param value String value
     * @return Processed string
     */
    std::string operator()(const std::string& value) const {
        return "String: " + value;
    }
};

/**
 * @brief Advanced template class with multiple template parameters
 * @tparam T Primary type
 * @tparam Allocator Allocator type
 * @tparam Compare Comparison function type
 */
template<typename T,
         typename Allocator = std::allocator<T>,
         typename Compare = std::less<T>>
class AdvancedContainer {
public:
    using value_type = T;
    using allocator_type = Allocator;
    using compare_type = Compare;

    /**
     * @brief Constructor
     * @param comp Comparison function
     * @param alloc Allocator
     */
    explicit AdvancedContainer(const Compare& comp = Compare{},
                              const Allocator& alloc = Allocator{})
        : compare_(comp), allocator_(alloc) {}

    /**
     * @brief Insert element
     * @param value Value to insert
     */
    void insert(const T& value) {
        auto it = std::lower_bound(data_.begin(), data_.end(), value, compare_);
        data_.insert(it, value);
    }

    /**
     * @brief Insert element with perfect forwarding
     * @tparam Args Argument types
     * @param args Arguments to forward to constructor
     */
    template<typename... Args>
    void emplace(Args&&... args) {
        T value(std::forward<Args>(args)...);
        insert(value);
    }

    /**
     * @brief Find element
     * @param value Value to find
     * @return Iterator to found element or end()
     */
    auto find(const T& value) const {
        auto it = std::lower_bound(data_.begin(), data_.end(), value, compare_);
        return (it != data_.end() && !compare_(value, *it)) ? it : data_.end();
    }

    /**
     * @brief Get size
     * @return Number of elements
     */
    size_t size() const { return data_.size(); }

    /**
     * @brief Begin iterator
     * @return Iterator to beginning
     */
    auto begin() const { return data_.begin(); }

    /**
     * @brief End iterator
     * @return Iterator to end
     */
    auto end() const { return data_.end(); }

protected:
    /**
     * @brief Protected method for derived classes
     * @return Reference to comparison function
     */
    const Compare& getCompare() const { return compare_; }

private:
    std::vector<T> data_;     ///< Stored data
    Compare compare_;         ///< Comparison function
    Allocator allocator_;     ///< Allocator instance

    /**
     * @brief Private helper method
     * @param value Value to validate
     * @return true if value is valid
     */
    bool validateValue(const T& value) const {
        return true;  // Placeholder validation
    }
};

/**
 * @brief Template specialization for pointer types
 * @tparam T Pointed-to type
 * @tparam Allocator Allocator type
 * @tparam Compare Comparison type
 */
template<typename T, typename Allocator, typename Compare>
class AdvancedContainer<T*, Allocator, Compare> {
public:
    using value_type = T*;

    /**
     * @brief Constructor for pointer specialization
     */
    AdvancedContainer() = default;

    /**
     * @brief Insert pointer
     * @param ptr Pointer to insert
     */
    void insert(T* ptr) {
        if (ptr) {
            pointers_.push_back(ptr);
        }
    }

    /**
     * @brief Get size
     * @return Number of pointers
     */
    size_t size() const { return pointers_.size(); }

private:
    std::vector<T*> pointers_;  ///< Stored pointers
};

/**
 * @brief Complex example with async operations and futures
 */
class AsyncProcessor {
public:
    /**
     * @brief Process data asynchronously with timeout
     * @tparam T Data type
     * @param data Data to process
     * @param timeoutMs Timeout in milliseconds
     * @return Optional result
     */
    template<typename T>
    std::optional<std::string> processWithTimeout(const T& data, int timeoutMs) {
        auto future = std::async(std::launch::async, [data]() {
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            return processData(data);
        });

        if (future.wait_for(std::chrono::milliseconds(timeoutMs)) == std::future_status::ready) {
            return future.get();
        }

        return std::nullopt;
    }

    /**
     * @brief Process multiple items in parallel
     * @tparam Container Container type
     * @param items Items to process
     * @return Vector of results
     */
    template<typename Container>
    std::vector<std::string> processParallel(const Container& items) {
        std::vector<std::future<std::string>> futures;

        for (const auto& item : items) {
            futures.push_back(std::async(std::launch::async, [item]() {
                return processData(item);
            }));
        }

        std::vector<std::string> results;
        for (auto& future : futures) {
            results.push_back(future.get());
        }

        return results;
    }

private:
    /**
     * @brief Process single data item
     * @tparam T Data type
     * @param data Data to process
     * @return Processed result
     */
    template<typename T>
    static std::string processData(const T& data) {
        return "Processed: " + std::to_string(data);
    }

    /**
     * @brief Specialization for string
     * @param data String data
     * @return Processed result
     */
    static std::string processData(const std::string& data) {
        return "Processed: " + data;
    }
};

/**
 * @brief Demonstration of complex features
 */
void demonstrateComplexFeatures() {
    // Compile-time computation
    constexpr auto fact5 = factorial(5);

    // CRTP demonstration
    Point p1(1.0, 2.0);
    Point p2(3.0, 4.0);
    bool isLess = p1 < p2;

    // Variadic template
    VariadicProcessor<int, double, std::string> processor(42, 3.14, "Hello");

    // Variant visitor
    std::vector<DataVariant> variants = {42, 3.14, std::string("test")};
    for (const auto& var : variants) {
        std::string result = std::visit(DataVisitor{}, var);
        std::cout << result << std::endl;
    }

    // Advanced container
    AdvancedContainer<int> container;
    container.insert(3);
    container.insert(1);
    container.insert(4);

    // Async processing
    AsyncProcessor asyncProc;
    auto result = asyncProc.processWithTimeout(42, 200);

    std::cout << "Factorial(5): " << fact5 << std::endl;
    std::cout << "Point comparison: " << (isLess ? "true" : "false") << std::endl;
    std::cout << "Container size: " << container.size() << std::endl;
    std::cout << "Async result: " << (result ? *result : "timeout") << std::endl;
}