// Test Pattern 4: Templates, Type Traits, and SFINAE
// Tests complex template usage with includes

#include <iostream>
#include <type_traits>
#include <utility>
#include <tuple>
#include <functional>
#include <memory>
#include <limits>
#include <iterator>
#include <concepts>  // C++20
#include <ranges>    // C++20

// Not using: limits, ranges

using namespace std;

// Template using type_traits for SFINAE
template<typename T>
typename enable_if<is_integral<T>::value, T>::type
safeAdd(T a, T b) {
    // Check for overflow (not using numeric_limits here though)
    return a + b;
}

template<typename T>
typename enable_if<is_floating_point<T>::value, T>::type
safeAdd(T a, T b) {
    return a + b;
}

// Using C++20 concepts (if available)
#if __cplusplus >= 202002L
template<typename T>
concept Numeric = is_arithmetic_v<T>;

template<Numeric T>
T multiply(T a, T b) {
    return a * b;
}
#endif

// Template using tuple and utility
template<typename... Args>
class VariadicContainer {
private:
    tuple<Args...> data;

public:
    VariadicContainer(Args... args) : data(forward<Args>(args)...) {
        // Using utility's forward
    }

    template<size_t Index>
    auto get() -> decltype(std::get<Index>(data)) {
        return std::get<Index>(data);
    }

    static constexpr size_t size() {
        return tuple_size<decltype(data)>::value;
    }
};

// Using functional for callbacks
template<typename T>
class EventEmitter {
private:
    vector<function<void(T)>> callbacks;

public:
    void on(function<void(T)> callback) {
        callbacks.push_back(callback);
    }

    void emit(T value) {
        for (auto& callback : callbacks) {
            callback(value);
        }
    }
};

// Template with iterator traits
template<typename Iterator>
void processRange(Iterator first, Iterator last) {
    using ValueType = typename iterator_traits<Iterator>::value_type;

    cout << "Processing range of " << typeid(ValueType).name() << " values:" << endl;

    for (auto it = first; it != last; ++it) {
        cout << "  " << *it << endl;
    }
}

// Smart pointer factory using memory
template<typename T, typename... Args>
unique_ptr<T> makeUnique(Args&&... args) {
    return unique_ptr<T>(new T(forward<Args>(args)...));
}

// Trait detector using type_traits
template<typename T>
struct has_toString {
private:
    template<typename U>
    static auto test(int) -> decltype(declval<U>().toString(), true_type{});

    template<typename>
    static false_type test(...);

public:
    static constexpr bool value = decltype(test<T>(0))::value;
};

// Example class with toString
class Message {
    string content;
public:
    Message(const string& msg) : content(msg) {}
    string toString() const { return content; }
};

int main() {
    // Using SFINAE-enabled functions
    cout << "Safe add integers: " << safeAdd(10, 20) << endl;
    cout << "Safe add doubles: " << safeAdd(10.5, 20.3) << endl;

    #if __cplusplus >= 202002L
    // Using concepts
    cout << "Multiply with concept: " << multiply(5, 6) << endl;
    #endif

    // Using variadic template with tuple
    VariadicContainer<int, string, double> container(42, "hello", 3.14);
    cout << "First element: " << container.get<0>() << endl;
    cout << "Second element: " << container.get<1>() << endl;
    cout << "Container size: " << container.size() << endl;

    // Using event emitter with functional
    EventEmitter<string> emitter;
    emitter.on([](const string& msg) {
        cout << "Event received: " << msg << endl;
    });
    emitter.emit("Hello, Events!");

    // Using iterator traits
    vector<int> numbers = {1, 2, 3, 4, 5};
    processRange(numbers.begin(), numbers.end());

    // Using smart pointer factory
    auto msg = makeUnique<Message>("Smart pointer message");
    cout << "Message: " << msg->toString() << endl;

    // Using trait detector
    cout << "Message has toString: " << has_toString<Message>::value << endl;
    cout << "int has toString: " << has_toString<int>::value << endl;

    return 0;
}