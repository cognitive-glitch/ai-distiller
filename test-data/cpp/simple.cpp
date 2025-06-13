#include <iostream>
#include <string>
#include "myheader.h"

using namespace std;

// Global constant
const int MAX_SIZE = 100;

// Simple function
int add(int a, int b) {
    return a + b;
}

// Function with default parameter
void printMessage(const string& msg = "Hello") {
    cout << msg << endl;
}

// Simple class
class Point {
private:
    int x, y;
    
public:
    // Constructor
    Point(int x = 0, int y = 0) : x(x), y(y) {}
    
    // Destructor
    ~Point() {}
    
    // Getter methods
    int getX() const { return x; }
    int getY() const { return y; }
    
    // Setter methods
    void setX(int newX) { x = newX; }
    void setY(int newY) { y = newY; }
    
    // Static method
    static Point origin() {
        return Point(0, 0);
    }
};

// Simple struct (public by default)
struct Rectangle {
    int width;
    int height;
    
    int area() const {
        return width * height;
    }
};

// Enum
enum Color {
    RED,
    GREEN,
    BLUE
};

// Namespace
namespace Utils {
    void log(const string& message) {
        cout << "[LOG] " << message << endl;
    }
}

int main() {
    Point p(10, 20);
    Rectangle rect = {5, 10};
    
    cout << "Point: (" << p.getX() << ", " << p.getY() << ")" << endl;
    cout << "Rectangle area: " << rect.area() << endl;
    
    Utils::log("Program completed");
    
    return 0;
}