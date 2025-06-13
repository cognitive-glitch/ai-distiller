// Simple Kotlin example to test basic parsing
package com.example.simple

import kotlin.math.PI
import kotlin.math.pow

// A simple data class
data class Point(val x: Double, val y: Double) {
    // Calculate distance from origin
    fun distanceFromOrigin(): Double {
        return kotlin.math.sqrt(x.pow(2) + y.pow(2))
    }
}

// An enum with custom properties
enum class Color(val rgb: Int) {
    RED(0xFF0000),
    GREEN(0x00FF00),
    BLUE(0x0000FF);
    
    fun toHex(): String = "#${rgb.toString(16).padStart(6, '0')}"
}

// A regular class with primary constructor
class Circle(val center: Point, val radius: Double) {
    val area: Double
        get() = PI * radius * radius
    
    val circumference: Double
        get() = 2 * PI * radius
    
    fun contains(point: Point): Boolean {
        val distance = kotlin.math.sqrt(
            (point.x - center.x).pow(2) + (point.y - center.y).pow(2)
        )
        return distance <= radius
    }
}

// An object (singleton)
object MathConstants {
    const val E = 2.71828
    const val PHI = 1.61803
    
    fun factorial(n: Int): Long {
        return if (n <= 1) 1 else n * factorial(n - 1)
    }
}

// Extension function
fun String.isPalindrome(): Boolean {
    return this == this.reversed()
}

// Top-level function with default parameters
fun greet(name: String = "World", excited: Boolean = false): String {
    val punctuation = if (excited) "!" else "."
    return "Hello, $name$punctuation"
}

// Main function
fun main() {
    val p1 = Point(3.0, 4.0)
    println("Distance from origin: ${p1.distanceFromOrigin()}")
    
    val circle = Circle(Point(0.0, 0.0), 5.0)
    println("Circle area: ${circle.area}")
    println("Contains p1: ${circle.contains(p1)}")
    
    println("Red color hex: ${Color.RED.toHex()}")
    
    println(greet("Kotlin", true))
    println("racecar".isPalindrome())
}