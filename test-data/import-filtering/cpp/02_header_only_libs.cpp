// Test Pattern 2: Header-Only Libraries and Forward Declarations
// Tests includes for header-only libraries and forward declarations

#include <algorithm>
#include <numeric>
#include <functional>
#include <iterator>
#include <type_traits>
#include <utility>
#include <chrono>
#include <random>
#include <thread>
#include <mutex>

// Header-only libraries (simulated)
#include "json.hpp"  // Like nlohmann/json
#include "catch2/catch.hpp"  // Testing framework
#include "spdlog/spdlog.h"  // Logging library

// Forward declarations
class MyForwardDeclaredClass;
struct ForwardDeclaredStruct;
namespace MyNamespace {
    class AnotherClass;
}

// Not using: numeric, type_traits, utility, thread, mutex, catch2/catch.hpp, ForwardDeclaredStruct, MyNamespace::AnotherClass

using namespace std;
using namespace std::chrono;

template<typename T>
void processContainer(vector<T>& container) {
    // Using algorithm
    sort(container.begin(), container.end());
    
    // Using iterator
    copy(container.begin(), container.end(), 
         ostream_iterator<T>(cout, " "));
    cout << endl;
    
    // Using functional
    auto sum = accumulate(container.begin(), container.end(), T{}, plus<T>());
    cout << "Sum: " << sum << endl;
}

class TimedOperation {
private:
    steady_clock::time_point start;
    
public:
    TimedOperation() : start(steady_clock::now()) {
        // Using chrono
    }
    
    ~TimedOperation() {
        auto end = steady_clock::now();
        auto duration = duration_cast<milliseconds>(end - start);
        cout << "Operation took " << duration.count() << " ms" << endl;
    }
};

void demonstrateRandom() {
    // Using random
    random_device rd;
    mt19937 gen(rd());
    uniform_int_distribution<> dis(1, 100);
    
    vector<int> numbers;
    for (int i = 0; i < 10; ++i) {
        numbers.push_back(dis(gen));
    }
    
    processContainer(numbers);
}

void demonstrateJSON() {
    // Using json.hpp (simulated nlohmann/json usage)
    json j;
    j["name"] = "Test";
    j["value"] = 42;
    j["array"] = {1, 2, 3};
    
    cout << "JSON: " << j.dump() << endl;
}

void demonstrateLogging() {
    // Using spdlog
    spdlog::info("This is an info message");
    spdlog::warn("This is a warning");
}

// Using forward declared class
void useForwardDeclared(MyForwardDeclaredClass* obj) {
    // Can only use pointer/reference with forward declaration
    if (obj != nullptr) {
        cout << "Got forward declared object" << endl;
    }
}

int main() {
    {
        TimedOperation timer;
        demonstrateRandom();
    }
    
    demonstrateJSON();
    demonstrateLogging();
    
    return 0;
}