// Test Pattern 5: Complex Scenarios - Namespace, Nested, Local Includes
// Tests namespace usage, nested includes, and includes in local scopes

#include <iostream>
#include <vector>
#include <string>
#include <memory>
#include <map>

// Namespace aliasing
namespace fs = std::filesystem;
namespace chr = std::chrono;

// Using declarations
using std::cout;
using std::endl;
using std::vector;
using std::string;

// Nested namespace (C++17)
namespace Company::Project::Utils {
    void helperFunction() {
        cout << "Helper function called" << endl;
    }
}

// Include in namespace (unusual but valid)
namespace ThirdParty {
    #include <algorithm>  // Scoped to this namespace
    #include <numeric>
    
    template<typename T>
    T sum(const vector<T>& vec) {
        // Using numeric from this namespace's include
        return std::accumulate(vec.begin(), vec.end(), T{});
    }
}

// Not using: map, and some includes might be scoped

// Function with local include (very unusual but technically valid)
void processWithLocalInclude() {
    #include <functional>  // Local scope include
    
    vector<int> numbers = {1, 2, 3, 4, 5};
    
    // Using functional from local include
    std::function<int(int)> doubler = [](int x) { return x * 2; };
    
    for (auto& n : numbers) {
        n = doubler(n);
    }
    
    cout << "Doubled numbers: ";
    for (const auto& n : numbers) {
        cout << n << " ";
    }
    cout << endl;
}

// Class with includes in private section (also unusual)
class DataManager {
private:
    #include <queue>
    #include <stack>
    
    std::queue<string> messageQueue;
    std::stack<int> undoStack;
    
public:
    void addMessage(const string& msg) {
        messageQueue.push(msg);
    }
    
    string getNextMessage() {
        if (messageQueue.empty()) return "";
        string msg = messageQueue.front();
        messageQueue.pop();
        return msg;
    }
    
    void pushUndo(int value) {
        undoStack.push(value);
    }
    
    int popUndo() {
        if (undoStack.empty()) return -1;
        int value = undoStack.top();
        undoStack.pop();
        return value;
    }
};

// Include guards and multiple inclusion test
#ifndef MYHEADER_H
#define MYHEADER_H
    #include <set>  // Only included if MYHEADER_H not defined
#endif

// Pragma once alternative
#pragma once
#include <list>  // Should be included only once despite pragma

// Using preprocessor to conditionally use includes
#define USE_CUSTOM_ALLOCATOR 0

#if USE_CUSTOM_ALLOCATOR
    #include <memory_resource>
    using Allocator = std::pmr::polymorphic_allocator<std::byte>;
#else
    using Allocator = std::allocator<std::byte>;
#endif

// Anonymous namespace with includes
namespace {
    #include <regex>
    
    bool isValidEmail(const string& email) {
        // Using regex from anonymous namespace include
        std::regex pattern(R"([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})");
        return std::regex_match(email, pattern);
    }
}

int main() {
    // Using namespace alias
    Company::Project::Utils::helperFunction();
    
    // Using ThirdParty namespace functions
    vector<double> values = {1.1, 2.2, 3.3, 4.4};
    cout << "Sum using ThirdParty: " << ThirdParty::sum(values) << endl;
    
    // Call function with local include
    processWithLocalInclude();
    
    // Use class with private includes
    DataManager manager;
    manager.addMessage("First");
    manager.addMessage("Second");
    cout << "Message: " << manager.getNextMessage() << endl;
    
    manager.pushUndo(100);
    manager.pushUndo(200);
    cout << "Undo value: " << manager.popUndo() << endl;
    
    // Using set from conditional include
    #ifndef MYHEADER_H
        std::set<int> uniqueNumbers = {3, 1, 4, 1, 5, 9};
        cout << "Unique numbers: ";
        for (const auto& n : uniqueNumbers) {
            cout << n << " ";
        }
        cout << endl;
    #endif
    
    // Using list
    std::list<string> names = {"Alice", "Bob", "Charlie"};
    cout << "Names in list: ";
    for (const auto& name : names) {
        cout << name << " ";
    }
    cout << endl;
    
    // Using anonymous namespace function
    string email = "user@example.com";
    cout << "Is '" << email << "' valid? " << (isValidEmail(email) ? "Yes" : "No") << endl;
    
    // Using smart pointers from top-level include
    auto ptr = std::make_shared<string>("Shared string");
    cout << "Shared pointer content: " << *ptr << endl;
    
    return 0;
}