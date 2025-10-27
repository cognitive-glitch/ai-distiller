#include <iostream>
#include <string>
#include <vector>
#include <memory>
#include <algorithm>
#include <functional>
#include <type_traits>
#include <future>
#include <thread>
#include <chrono>
#include <exception>
#include <utility>

/**
 * @brief Advanced C++ features demonstration including SFINAE, smart pointers,
 *        move semantics, and template specialization
 */

/**
 * @brief SFINAE trait to detect if a type has a specific method
 * @tparam T Type to check
 * @tparam Args Method argument types
 */
template<typename T, typename... Args>
struct has_process_method {
private:
    template<typename U>
    static auto test(int) -> decltype(std::declval<U>().process(std::declval<Args>()...), std::true_type{});

    template<typename>
    static std::false_type test(...);

public:
    static constexpr bool value = decltype(test<T>(0))::value;
};

/**
 * @brief Helper variable template for has_process_method
 */
template<typename T, typename... Args>
constexpr bool has_process_method_v = has_process_method<T, Args...>::value;

/**
 * @brief Generic resource manager with RAII and move semantics
 * @tparam Resource The type of resource to manage
 * @tparam Deleter Function object for resource cleanup
 */
template<typename Resource, typename Deleter = std::default_delete<Resource>>
class ResourceManager {
public:
    /**
     * @brief Constructor taking ownership of resource
     * @param resource Pointer to resource
     * @param deleter Custom deleter (optional)
     */
    explicit ResourceManager(Resource* resource = nullptr,
                           Deleter deleter = Deleter{})
        : resource_(resource), deleter_(std::move(deleter)) {}

    /**
     * @brief Move constructor
     * @param other Source object
     */
    ResourceManager(ResourceManager&& other) noexcept
        : resource_(std::exchange(other.resource_, nullptr)),
          deleter_(std::move(other.deleter_)) {}

    /**
     * @brief Move assignment operator
     * @param other Source object
     * @return Reference to this object
     */
    ResourceManager& operator=(ResourceManager&& other) noexcept {
        if (this != &other) {
            reset();
            resource_ = std::exchange(other.resource_, nullptr);
            deleter_ = std::move(other.deleter_);
        }
        return *this;
    }

    /**
     * @brief Destructor
     */
    ~ResourceManager() {
        reset();
    }

    /**
     * @brief Deleted copy constructor
     */
    ResourceManager(const ResourceManager&) = delete;

    /**
     * @brief Deleted copy assignment
     */
    ResourceManager& operator=(const ResourceManager&) = delete;

    /**
     * @brief Get raw pointer to resource
     * @return Raw pointer
     */
    Resource* get() const noexcept { return resource_; }

    /**
     * @brief Release ownership of resource
     * @return Raw pointer to released resource
     */
    Resource* release() noexcept {
        return std::exchange(resource_, nullptr);
    }

    /**
     * @brief Reset with new resource
     * @param resource New resource pointer
     */
    void reset(Resource* resource = nullptr) {
        if (resource_) {
            deleter_(resource_);
        }
        resource_ = resource;
    }

    /**
     * @brief Boolean conversion operator
     * @return true if managing a resource
     */
    explicit operator bool() const noexcept {
        return resource_ != nullptr;
    }

    /**
     * @brief Dereference operator
     * @return Reference to resource
     */
    Resource& operator*() const {
        return *resource_;
    }

    /**
     * @brief Arrow operator
     * @return Pointer to resource
     */
    Resource* operator->() const noexcept {
        return resource_;
    }

private:
    Resource* resource_;  ///< Managed resource
    Deleter deleter_;     ///< Custom deleter
};

/**
 * @brief Base processor interface
 */
class IProcessor {
public:
    /**
     * @brief Virtual destructor
     */
    virtual ~IProcessor() = default;

    /**
     * @brief Process data
     * @param data Input data
     * @return Processed result
     */
    virtual std::string process(const std::string& data) = 0;

    /**
     * @brief Get processor name
     * @return Processor identifier
     */
    virtual std::string getName() const = 0;
};

/**
 * @brief Concrete text processor
 */
class TextProcessor : public IProcessor {
public:
    /**
     * @brief Constructor
     * @param name Processor name
     */
    explicit TextProcessor(const std::string& name) : name_(name) {}

    /**
     * @brief Process text data
     * @param data Input text
     * @return Processed text
     */
    std::string process(const std::string& data) override {
        return "[" + name_ + "] " + data;
    }

    /**
     * @brief Get processor name
     * @return Name string
     */
    std::string getName() const override {
        return name_;
    }

private:
    std::string name_;  ///< Processor name
};

/**
 * @brief Template specialization example for different numeric types
 * @tparam T Numeric type
 */
template<typename T>
class Calculator {
public:
    /**
     * @brief Add two values
     * @param a First value
     * @param b Second value
     * @return Sum
     */
    static T add(const T& a, const T& b) {
        return a + b;
    }

    /**
     * @brief Multiply two values
     * @param a First value
     * @param b Second value
     * @return Product
     */
    static T multiply(const T& a, const T& b) {
        return a * b;
    }

    /**
     * @brief Get type information
     * @return Type name string
     */
    static std::string getTypeName() {
        return "generic";
    }
};

/**
 * @brief Specialization for double
 */
template<>
class Calculator<double> {
public:
    /**
     * @brief Add two doubles with precision handling
     * @param a First value
     * @param b Second value
     * @return Sum
     */
    static double add(const double& a, const double& b) {
        return a + b;
    }

    /**
     * @brief Multiply two doubles
     * @param a First value
     * @param b Second value
     * @return Product
     */
    static double multiply(const double& a, const double& b) {
        return a * b;
    }

    /**
     * @brief Divide two doubles with zero check
     * @param a Dividend
     * @param b Divisor
     * @return Quotient
     * @throws std::invalid_argument if divisor is zero
     */
    static double divide(const double& a, const double& b) {
        if (std::abs(b) < 1e-10) {
            throw std::invalid_argument("Division by zero");
        }
        return a / b;
    }

    /**
     * @brief Get type information
     * @return Type name string
     */
    static std::string getTypeName() {
        return "double";
    }
};

/**
 * @brief Partial specialization for pointer types
 * @tparam T Pointed-to type
 */
template<typename T>
class Calculator<T*> {
public:
    /**
     * @brief Add offset to pointer
     * @param ptr Base pointer
     * @param offset Offset to add
     * @return New pointer
     */
    static T* add(T* ptr, std::ptrdiff_t offset) {
        return ptr + offset;
    }

    /**
     * @brief Get type information
     * @return Type name string
     */
    static std::string getTypeName() {
        return "pointer";
    }
};

/**
 * @brief Advanced processing pipeline with async operations
 */
class ProcessingPipeline {
public:
    /**
     * @brief Constructor
     */
    ProcessingPipeline() = default;

    /**
     * @brief Add processor to pipeline
     * @param processor Unique pointer to processor
     */
    void addProcessor(std::unique_ptr<IProcessor> processor) {
        processors_.push_back(std::move(processor));
    }

    /**
     * @brief Process data through pipeline synchronously
     * @param input Input data
     * @return Processed result
     */
    std::string process(const std::string& input) const {
        std::string result = input;
        for (const auto& processor : processors_) {
            result = processor->process(result);
        }
        return result;
    }

    /**
     * @brief Process data asynchronously
     * @param input Input data
     * @return Future containing the result
     */
    std::future<std::string> processAsync(const std::string& input) const {
        return std::async(std::launch::async, [this, input]() {
            return this->process(input);
        });
    }

    /**
     * @brief Process multiple inputs in parallel
     * @param inputs Vector of input strings
     * @return Vector of futures
     */
    std::vector<std::future<std::string>> processMultiple(
        const std::vector<std::string>& inputs) const {
        std::vector<std::future<std::string>> futures;
        futures.reserve(inputs.size());

        for (const auto& input : inputs) {
            futures.push_back(processAsync(input));
        }

        return futures;
    }

    /**
     * @brief Get number of processors
     * @return Processor count
     */
    size_t getProcessorCount() const {
        return processors_.size();
    }

protected:
    /**
     * @brief Protected method for derived classes
     * @param index Processor index
     * @return Reference to processor at index
     */
    const IProcessor& getProcessor(size_t index) const {
        if (index >= processors_.size()) {
            throw std::out_of_range("Processor index out of range");
        }
        return *processors_[index];
    }

private:
    /// Vector of processor smart pointers
    std::vector<std::unique_ptr<IProcessor>> processors_;

    /**
     * @brief Private helper for internal operations
     * @param data Data to validate
     * @return true if data is valid
     */
    bool validateData(const std::string& data) const {
        return !data.empty();
    }
};

/**
 * @brief Template function with SFINAE
 * @tparam T Type to process
 * @param obj Object to process
 * @param data Data to pass to process method
 * @return Processed result if T has process method, empty string otherwise
 */
template<typename T>
auto safeProcess(T& obj, const std::string& data)
    -> std::enable_if_t<has_process_method_v<T, std::string>, std::string> {
    return obj.process(data);
}

/**
 * @brief SFINAE overload for types without process method
 * @tparam T Type to process
 * @param obj Object (unused)
 * @param data Data (unused)
 * @return Empty string
 */
template<typename T>
auto safeProcess(T& obj, const std::string& data)
    -> std::enable_if_t<!has_process_method_v<T, std::string>, std::string> {
    return "";
}

/**
 * @brief Factory function with perfect forwarding
 * @tparam T Type to create
 * @tparam Args Constructor argument types
 * @param args Constructor arguments
 * @return Unique pointer to created object
 */
template<typename T, typename... Args>
std::unique_ptr<T> makeUnique(Args&&... args) {
    return std::make_unique<T>(std::forward<Args>(args)...);
}

/**
 * @brief Exception class for processing errors
 */
class ProcessingException : public std::exception {
public:
    /**
     * @brief Constructor
     * @param message Error message
     */
    explicit ProcessingException(const std::string& message)
        : message_(message) {}

    /**
     * @brief Get error message
     * @return Error message C-string
     */
    const char* what() const noexcept override {
        return message_.c_str();
    }

private:
    std::string message_;  ///< Error message
};

/**
 * @brief Demonstration function for advanced features
 */
void demonstrateAdvancedFeatures() {
    // Resource management with custom deleter
    auto customDeleter = [](TextProcessor* p) {
        std::cout << "Custom deleting processor: " << p->getName() << std::endl;
        delete p;
    };

    ResourceManager<TextProcessor, decltype(customDeleter)> manager(
        new TextProcessor("Advanced"), customDeleter);

    // Processing pipeline
    ProcessingPipeline pipeline;
    pipeline.addProcessor(makeUnique<TextProcessor>("First"));
    pipeline.addProcessor(makeUnique<TextProcessor>("Second"));

    // Async processing
    auto future = pipeline.processAsync("Hello World");
    std::string result = future.get();

    // Template specialization
    auto intSum = Calculator<int>::add(5, 10);
    auto doubleSum = Calculator<double>::add(3.14, 2.86);

    // SFINAE demonstration
    TextProcessor processor("SFINAE Test");
    std::string sfResult = safeProcess(processor, "test data");

    std::cout << "Async result: " << result << std::endl;
    std::cout << "Int sum: " << intSum << std::endl;
    std::cout << "Double sum: " << doubleSum << std::endl;
    std::cout << "SFINAE result: " << sfResult << std::endl;
}