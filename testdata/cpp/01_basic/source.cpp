#include <iostream>
#include <string>
#include <vector>

/**
 * @brief A basic Point class demonstrating fundamental C++ concepts
 *
 * This class showcases basic C++ features like constructors, destructors,
 * public/private members, and basic inheritance.
 */
class Point {
public:
    /**
     * @brief Default constructor
     */
    Point() : x_(0), y_(0) {}

    /**
     * @brief Parameterized constructor
     * @param x The x coordinate
     * @param y The y coordinate
     */
    Point(double x, double y) : x_(x), y_(y) {}

    /**
     * @brief Copy constructor
     * @param other The point to copy from
     */
    Point(const Point& other) : x_(other.x_), y_(other.y_) {}

    /**
     * @brief Virtual destructor for proper inheritance
     */
    virtual ~Point() = default;

    /**
     * @brief Get the x coordinate
     * @return The x coordinate
     */
    double getX() const { return x_; }

    /**
     * @brief Get the y coordinate
     * @return The y coordinate
     */
    double getY() const { return y_; }

    /**
     * @brief Set the x coordinate
     * @param x The new x coordinate
     */
    void setX(double x) { x_ = x; }

    /**
     * @brief Set the y coordinate
     * @param y The new y coordinate
     */
    void setY(double y) { y_ = y; }

    /**
     * @brief Calculate distance from origin
     * @return Distance from origin
     */
    virtual double distanceFromOrigin() const {
        return std::sqrt(x_ * x_ + y_ * y_);
    }

protected:
    /**
     * @brief Validate coordinates (protected utility)
     * @return true if coordinates are valid
     */
    bool validateCoordinates() const {
        return !std::isnan(x_) && !std::isnan(y_);
    }

private:
    double x_;  ///< X coordinate
    double y_;  ///< Y coordinate

    /**
     * @brief Private helper for internal calculations
     */
    void internalCalculation() {
        // Some private computation
        double temp = x_ + y_;
        (void)temp; // Suppress unused variable warning
    }
};

/**
 * @brief A 3D point class extending Point
 */
class Point3D : public Point {
public:
    /**
     * @brief Constructor for 3D point
     * @param x X coordinate
     * @param y Y coordinate
     * @param z Z coordinate
     */
    Point3D(double x, double y, double z) : Point(x, y), z_(z) {}

    /**
     * @brief Get the z coordinate
     * @return The z coordinate
     */
    double getZ() const { return z_; }

    /**
     * @brief Set the z coordinate
     * @param z The new z coordinate
     */
    void setZ(double z) { z_ = z; }

    /**
     * @brief Override distance calculation for 3D
     * @return Distance from origin in 3D space
     */
    double distanceFromOrigin() const override {
        return std::sqrt(getX() * getX() + getY() * getY() + z_ * z_);
    }

private:
    double z_;  ///< Z coordinate
};

/**
 * @brief Basic template class for demonstration
 * @tparam T The type to store
 */
template<typename T>
class Container {
public:
    /**
     * @brief Constructor
     * @param value Initial value
     */
    explicit Container(const T& value) : value_(value) {}

    /**
     * @brief Get the stored value
     * @return Reference to the stored value
     */
    const T& getValue() const { return value_; }

    /**
     * @brief Set the stored value
     * @param value New value to store
     */
    void setValue(const T& value) { value_ = value; }

private:
    T value_;  ///< Stored value
};

/**
 * @brief Utility functions namespace
 */
namespace MathUtils {
    /**
     * @brief Calculate the maximum of two values
     * @tparam T Type of values to compare
     * @param a First value
     * @param b Second value
     * @return The maximum value
     */
    template<typename T>
    T max(const T& a, const T& b) {
        return (a > b) ? a : b;
    }

    /**
     * @brief Calculate the minimum of two values
     * @tparam T Type of values to compare
     * @param a First value
     * @param b Second value
     * @return The minimum value
     */
    template<typename T>
    T min(const T& a, const T& b) {
        return (a < b) ? a : b;
    }
}

/**
 * @brief Main function demonstrating basic usage
 * @return Exit code
 */
int main() {
    // Create some points
    Point p1(3.0, 4.0);
    Point3D p2(1.0, 2.0, 3.0);

    // Use containers
    Container<int> intContainer(42);
    Container<std::string> stringContainer("Hello");

    // Use utility functions
    int maxVal = MathUtils::max(10, 20);
    double minVal = MathUtils::min(1.5, 2.5);

    // Output results
    std::cout << "Point distance: " << p1.distanceFromOrigin() << std::endl;
    std::cout << "3D Point distance: " << p2.distanceFromOrigin() << std::endl;
    std::cout << "Container value: " << intContainer.getValue() << std::endl;
    std::cout << "Max value: " << maxVal << std::endl;

    return 0;
}