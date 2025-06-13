// Basic C++ features test
#include <vector>
#include <memory>

// Forward declaration
class Engine;

// Template function
template<typename T>
T max(T a, T b) {
    return (a > b) ? a : b;
}

// Abstract base class with virtual functions
class Vehicle {
protected:
    std::string brand;
    int year;
    
public:
    Vehicle(const std::string& brand, int year) 
        : brand(brand), year(year) {}
    
    // Pure virtual function
    virtual void start() = 0;
    
    // Virtual function with implementation
    virtual void stop() {
        // Default implementation
    }
    
    // Const member function
    std::string getBrand() const { return brand; }
    
    // Friend function declaration
    friend void displayVehicle(const Vehicle& v);
};

// Derived class with inheritance
class Car : public Vehicle {
private:
    int numDoors;
    Engine* engine;
    
public:
    // Constructor with member initializer list
    Car(const std::string& brand, int year, int doors) 
        : Vehicle(brand, year), numDoors(doors), engine(nullptr) {}
    
    // Override virtual function
    void start() override {
        // Car-specific start implementation
    }
    
    // Destructor
    virtual ~Car() {
        delete engine;
    }
};

// Template class
template<typename T>
class Stack {
private:
    std::vector<T> elements;
    
public:
    void push(const T& elem) {
        elements.push_back(elem);
    }
    
    T pop() {
        T elem = elements.back();
        elements.pop_back();
        return elem;
    }
    
    bool empty() const {
        return elements.empty();
    }
};

// Union example
union Data {
    int i;
    float f;
    char str[20];
};

// Inline function
inline int square(int x) {
    return x * x;
}

// Constexpr function (C++11)
constexpr int factorial(int n) {
    return n <= 1 ? 1 : n * factorial(n - 1);
}

// Function with noexcept
void safeFunction() noexcept {
    // This function promises not to throw
}

// Operator overloading
class Complex {
    double real, imag;
    
public:
    Complex(double r = 0, double i = 0) : real(r), imag(i) {}
    
    // Operator+ overload
    Complex operator+(const Complex& other) const {
        return Complex(real + other.real, imag + other.imag);
    }
    
    // Friend operator<<
    friend std::ostream& operator<<(std::ostream& os, const Complex& c);
};

// Static member variable
class Counter {
private:
    static int count;
    
public:
    Counter() { count++; }
    static int getCount() { return count; }
};

// Initialize static member
int Counter::count = 0;

// Using auto and lambda (C++11)
void modernCpp() {
    auto vec = std::vector<int>{1, 2, 3, 4, 5};
    
    auto sum = [](int a, int b) -> int {
        return a + b;
    };
    
    auto result = sum(10, 20);
}