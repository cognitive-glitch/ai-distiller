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
#include <concepts>      // C++20
#include <coroutine>     // C++20
#include <ranges>        // C++20
#include <span>          // C++20

/**
 * @brief Very complex C++ features including C++20 concepts, coroutines,
 *        advanced template metaprogramming, and sophisticated type manipulation
 */

// ============================================================================
// C++20 CONCEPTS SECTION
// ============================================================================

/**
 * @brief Concept for arithmetic types
 * @tparam T Type to check
 */
template<typename T>
concept Arithmetic = std::is_arithmetic_v<T>;

/**
 * @brief Concept for types that can be incremented
 * @tparam T Type to check
 */
template<typename T>
concept Incrementable = requires(T t) {
    ++t;
    t++;
};

/**
 * @brief Concept for container-like types
 * @tparam T Type to check
 */
template<typename T>
concept Container = requires(T t) {
    t.begin();
    t.end();
    t.size();
    typename T::value_type;
};

/**
 * @brief Concept for callable types with specific signature
 * @tparam F Function type
 * @tparam Args Argument types
 */
template<typename F, typename... Args>
concept Callable = requires(F f, Args... args) {
    f(args...);
};

/**
 * @brief Advanced concept combining multiple requirements
 * @tparam T Type to check
 */
template<typename T>
concept AdvancedNumeric = Arithmetic<T> && Incrementable<T> &&
    requires(T a, T b) {
        { a + b } -> std::convertible_to<T>;
        { a - b } -> std::convertible_to<T>;
        { a * b } -> std::convertible_to<T>;
        { a / b } -> std::convertible_to<T>;
    };

// ============================================================================
// COROUTINES SECTION (C++20)
// ============================================================================

/**
 * @brief Simple coroutine return type
 */
struct SimpleTask {
    struct promise_type {
        /**
         * @brief Get return object
         * @return SimpleTask instance
         */
        SimpleTask get_return_object() {
            return SimpleTask{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        /**
         * @brief Initial suspend
         * @return Never suspend initially
         */
        std::suspend_never initial_suspend() { return {}; }

        /**
         * @brief Final suspend
         * @return Always suspend finally
         */
        std::suspend_always final_suspend() noexcept { return {}; }

        /**
         * @brief Handle return void
         */
        void return_void() {}

        /**
         * @brief Handle unhandled exception
         */
        void unhandled_exception() {}
    };

    /**
     * @brief Constructor
     * @param h Coroutine handle
     */
    explicit SimpleTask(std::coroutine_handle<promise_type> h) : handle_(h) {}

    /**
     * @brief Destructor
     */
    ~SimpleTask() {
        if (handle_) {
            handle_.destroy();
        }
    }

    /**
     * @brief Move constructor
     * @param other Other task
     */
    SimpleTask(SimpleTask&& other) noexcept : handle_(std::exchange(other.handle_, {})) {}

    /**
     * @brief Move assignment
     * @param other Other task
     * @return Reference to this
     */
    SimpleTask& operator=(SimpleTask&& other) noexcept {
        if (this != &other) {
            if (handle_) {
                handle_.destroy();
            }
            handle_ = std::exchange(other.handle_, {});
        }
        return *this;
    }

    /**
     * @brief Deleted copy operations
     */
    SimpleTask(const SimpleTask&) = delete;
    SimpleTask& operator=(const SimpleTask&) = delete;

private:
    std::coroutine_handle<promise_type> handle_;  ///< Coroutine handle
};

/**
 * @brief Generator coroutine for producing values
 * @tparam T Value type
 */
template<typename T>
struct Generator {
    struct promise_type {
        T current_value;  ///< Current generated value

        /**
         * @brief Get return object
         * @return Generator instance
         */
        Generator get_return_object() {
            return Generator{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        /**
         * @brief Initial suspend
         * @return Always suspend initially
         */
        std::suspend_always initial_suspend() { return {}; }

        /**
         * @brief Final suspend
         * @return Always suspend finally
         */
        std::suspend_always final_suspend() noexcept { return {}; }

        /**
         * @brief Handle return void
         */
        void return_void() {}

        /**
         * @brief Handle unhandled exception
         */
        void unhandled_exception() {}

        /**
         * @brief Handle co_yield
         * @param value Value to yield
         * @return Suspend always
         */
        std::suspend_always yield_value(T value) {
            current_value = value;
            return {};
        }
    };

    /**
     * @brief Constructor
     * @param h Coroutine handle
     */
    explicit Generator(std::coroutine_handle<promise_type> h) : handle_(h) {}

    /**
     * @brief Destructor
     */
    ~Generator() {
        if (handle_) {
            handle_.destroy();
        }
    }

    /**
     * @brief Move constructor
     * @param other Other generator
     */
    Generator(Generator&& other) noexcept : handle_(std::exchange(other.handle_, {})) {}

    /**
     * @brief Move assignment
     * @param other Other generator
     * @return Reference to this
     */
    Generator& operator=(Generator&& other) noexcept {
        if (this != &other) {
            if (handle_) {
                handle_.destroy();
            }
            handle_ = std::exchange(other.handle_, {});
        }
        return *this;
    }

    /**
     * @brief Deleted copy operations
     */
    Generator(const Generator&) = delete;
    Generator& operator=(const Generator&) = delete;

    /**
     * @brief Check if generator has more values
     * @return true if more values available
     */
    bool next() {
        if (handle_) {
            handle_.resume();
            return !handle_.done();
        }
        return false;
    }

    /**
     * @brief Get current value
     * @return Current value
     */
    T value() const {
        return handle_.promise().current_value;
    }

private:
    std::coroutine_handle<promise_type> handle_;  ///< Coroutine handle
};

// ============================================================================
// ADVANCED TEMPLATE METAPROGRAMMING
// ============================================================================

/**
 * @brief Compile-time computation of type traits
 * @tparam T Type to analyze
 */
template<typename T>
struct TypeAnalyzer {
    static constexpr bool is_pointer = std::is_pointer_v<T>;
    static constexpr bool is_reference = std::is_reference_v<T>;
    static constexpr bool is_const = std::is_const_v<std::remove_reference_t<T>>;
    static constexpr size_t size = sizeof(T);
    static constexpr size_t alignment = alignof(T);

    /**
     * @brief Get type category as string
     * @return Type category description
     */
    static constexpr const char* category() {
        if constexpr (std::is_integral_v<T>) {
            return "integral";
        } else if constexpr (std::is_floating_point_v<T>) {
            return "floating_point";
        } else if constexpr (std::is_pointer_v<T>) {
            return "pointer";
        } else if constexpr (std::is_class_v<T>) {
            return "class";
        } else {
            return "other";
        }
    }
};

/**
 * @brief Advanced SFINAE for method detection
 * @tparam T Type to check
 */
template<typename T>
class HasAdvancedMethods {
private:
    template<typename U>
    static auto test_serialize(int)
        -> decltype(std::declval<U>().serialize(), std::true_type{});
    template<typename>
    static std::false_type test_serialize(...);

    template<typename U>
    static auto test_deserialize(int)
        -> decltype(std::declval<U>().deserialize(std::string{}), std::true_type{});
    template<typename>
    static std::false_type test_deserialize(...);

    template<typename U>
    static auto test_validate(int)
        -> decltype(std::declval<U>().validate(), std::true_type{});
    template<typename>
    static std::false_type test_validate(...);

public:
    static constexpr bool has_serialize = decltype(test_serialize<T>(0))::value;
    static constexpr bool has_deserialize = decltype(test_deserialize<T>(0))::value;
    static constexpr bool has_validate = decltype(test_validate<T>(0))::value;
    static constexpr bool is_serializable = has_serialize && has_deserialize;
};

/**
 * @brief Template specialization dispatcher
 * @tparam T Type to dispatch
 * @tparam Enable SFINAE parameter
 */
template<typename T, typename Enable = void>
struct Dispatcher {
    /**
     * @brief Default dispatch
     * @param value Value to process
     * @return Processed string
     */
    static std::string dispatch(const T& value) {
        return "unknown type";
    }
};

/**
 * @brief Specialization for arithmetic types
 * @tparam T Arithmetic type
 */
template<typename T>
struct Dispatcher<T, std::enable_if_t<std::is_arithmetic_v<T>>> {
    /**
     * @brief Dispatch for arithmetic types
     * @param value Arithmetic value
     * @return Processed string
     */
    static std::string dispatch(const T& value) {
        return "arithmetic: " + std::to_string(value);
    }
};

/**
 * @brief Specialization for container types
 * @tparam T Container type
 */
template<typename T>
struct Dispatcher<T, std::enable_if_t<Container<T>>> {
    /**
     * @brief Dispatch for container types
     * @param container Container value
     * @return Processed string
     */
    static std::string dispatch(const T& container) {
        return "container with " + std::to_string(container.size()) + " elements";
    }
};

// ============================================================================
// ADVANCED CLASSES WITH C++20 FEATURES
// ============================================================================

/**
 * @brief Advanced processor using concepts and coroutines
 * @tparam T Data type (must satisfy AdvancedNumeric concept)
 */
template<AdvancedNumeric T>
class AdvancedProcessor {
public:
    /**
     * @brief Constructor
     * @param name Processor name
     */
    explicit AdvancedProcessor(const std::string& name) : name_(name) {}

    /**
     * @brief Process data using concepts
     * @param data Input data
     * @return Processed result
     */
    T process(const T& data) requires Incrementable<T> {
        T result = data;
        ++result;
        return result * static_cast<T>(2);
    }

    /**
     * @brief Process container of data
     * @tparam Container Container type
     * @param container Input container
     * @return Processed results
     */
    template<Container Container>
    auto processContainer(const Container& container)
        -> std::vector<typename Container::value_type>
        requires std::same_as<typename Container::value_type, T> {

        std::vector<T> results;
        results.reserve(container.size());

        for (const auto& item : container) {
            results.push_back(process(item));
        }

        return results;
    }

    /**
     * @brief Generate sequence using coroutine
     * @param start Starting value
     * @param count Number of values to generate
     * @return Generator coroutine
     */
    Generator<T> generateSequence(T start, size_t count) {
        for (size_t i = 0; i < count; ++i) {
            co_yield start + static_cast<T>(i);
        }
    }

    /**
     * @brief Process data asynchronously
     * @param data Input data
     * @return SimpleTask coroutine
     */
    SimpleTask processAsync(const T& data) {
        // Simulate async work
        std::this_thread::sleep_for(std::chrono::milliseconds(10));
        auto result = process(data);
        std::cout << name_ << " processed " << data << " -> " << result << std::endl;
        co_return;
    }

protected:
    /**
     * @brief Protected method for derived classes
     * @param value Value to validate
     * @return true if valid
     */
    bool validateInput(const T& value) const requires std::totally_ordered<T> {
        return value >= static_cast<T>(0);
    }

private:
    std::string name_;  ///< Processor name

    /**
     * @brief Private helper using consteval
     * @param value Compile-time value
     * @return Processed compile-time result
     */
    consteval T processAtCompileTime(T value) const {
        return value * static_cast<T>(3);
    }
};

/**
 * @brief Sophisticated type-erased container
 */
class TypeErasedContainer {
public:
    /**
     * @brief Store any type that satisfies the concept
     * @tparam T Type to store
     * @param value Value to store
     */
    template<typename T>
    void store(T&& value) requires (!std::is_same_v<std::decay_t<T>, TypeErasedContainer>) {
        using DecayedT = std::decay_t<T>;

        data_.emplace_back(std::make_unique<Wrapper<DecayedT>>(std::forward<T>(value)));
    }

    /**
     * @brief Process all stored values
     * @tparam F Function type
     * @param f Processing function
     */
    template<Callable<const std::any&> F>
    void processAll(F&& f) const {
        for (const auto& wrapper : data_) {
            f(wrapper->getValue());
        }
    }

    /**
     * @brief Get stored value by type
     * @tparam T Type to retrieve
     * @return Optional value
     */
    template<typename T>
    std::optional<T> get() const {
        for (const auto& wrapper : data_) {
            if (auto* typed = dynamic_cast<const Wrapper<T>*>(wrapper.get())) {
                return std::any_cast<T>(typed->getValue());
            }
        }
        return std::nullopt;
    }

    /**
     * @brief Get size
     * @return Number of stored items
     */
    size_t size() const { return data_.size(); }

private:
    /**
     * @brief Base wrapper interface
     */
    struct WrapperBase {
        /**
         * @brief Virtual destructor
         */
        virtual ~WrapperBase() = default;

        /**
         * @brief Get stored value as any
         * @return std::any containing the value
         */
        virtual std::any getValue() const = 0;
    };

    /**
     * @brief Concrete wrapper for specific type
     * @tparam T Wrapped type
     */
    template<typename T>
    struct Wrapper : WrapperBase {
        /**
         * @brief Constructor
         * @param value Value to wrap
         */
        explicit Wrapper(T value) : value_(std::move(value)) {}

        /**
         * @brief Get value implementation
         * @return std::any containing the value
         */
        std::any getValue() const override {
            return value_;
        }

        T value_;  ///< Stored value
    };

    /// Vector of type-erased wrappers
    std::vector<std::unique_ptr<WrapperBase>> data_;
};

/**
 * @brief Module-like namespace (simulating C++20 modules)
 */
namespace AdvancedFeatures {
    /**
     * @brief Consteval function for compile-time string processing
     * @param str Input string
     * @return Processed string length
     */
    consteval size_t processString(const char* str) {
        size_t len = 0;
        while (str[len] != '\0') {
            ++len;
        }
        return len;
    }

    /**
     * @brief Constexpr function with complex logic
     * @tparam T Numeric type
     * @param value Input value
     * @return Processed result
     */
    template<AdvancedNumeric T>
    constexpr T complexCalculation(T value) {
        T result = value;
        for (int i = 0; i < 5; ++i) {
            result = result * static_cast<T>(2) + static_cast<T>(1);
        }
        return result;
    }

    /**
     * @brief Use ranges and views (C++20)
     * @tparam Range Range type
     * @param range Input range
     * @return Transformed results
     */
    template<std::ranges::range Range>
    auto processRange(Range&& range) {
        namespace views = std::views;

        return range | views::filter([](const auto& x) { return x > 0; })
                    | views::transform([](const auto& x) { return x * 2; })
                    | views::take(10);
    }
}

/**
 * @brief Demonstration of very complex features
 */
void demonstrateVeryComplexFeatures() {
    // Concepts demonstration
    AdvancedProcessor<int> intProcessor("IntProcessor");
    auto result = intProcessor.process(42);

    // Coroutine demonstration
    auto generator = intProcessor.generateSequence(1, 5);
    std::vector<int> generated;
    while (generator.next()) {
        generated.push_back(generator.value());
    }

    // Type-erased container
    TypeErasedContainer container;
    container.store(123);
    container.store(3.14);
    container.store(std::string("Hello"));

    // Process with lambda
    container.processAll([](const std::any& value) {
        // Type inspection would go here
        std::cout << "Processing stored value" << std::endl;
    });

    // Consteval demonstration
    constexpr auto strLen = AdvancedFeatures::processString("Hello, World!");
    constexpr auto calcResult = AdvancedFeatures::complexCalculation(5);

    // Ranges demonstration (C++20)
    std::vector<int> numbers = {-2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12};
    auto processed = AdvancedFeatures::processRange(numbers);

    std::cout << "Int processor result: " << result << std::endl;
    std::cout << "Generated values: " << generated.size() << std::endl;
    std::cout << "Container size: " << container.size() << std::endl;
    std::cout << "String length (consteval): " << strLen << std::endl;
    std::cout << "Calculation result (constexpr): " << calcResult << std::endl;

    // Type analysis demonstration
    using IntAnalysis = TypeAnalyzer<int>;
    std::cout << "Int category: " << IntAnalysis::category() << std::endl;
    std::cout << "Int size: " << IntAnalysis::size << std::endl;

    // Advanced method detection
    constexpr bool hasSerialize = HasAdvancedMethods<std::string>::has_serialize;
    std::cout << "String has serialize: " << (hasSerialize ? "true" : "false") << std::endl;
}