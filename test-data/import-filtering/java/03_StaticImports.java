// Test Pattern 3: Static Imports
// Tests import static for methods and fields

package com.example.importtest;

import static java.lang.Math.*;  // Will use PI, sin, cos, sqrt
import static java.lang.System.out;
import static java.lang.System.err;
import static java.util.Collections.emptyList;
import static java.util.Collections.singleton;
import static java.lang.Integer.parseInt;
import static java.lang.String.format;
import static java.util.Arrays.asList;
import static java.util.stream.Collectors.*;  // Will use toList, joining

// Not using: err, emptyList, singleton, parseInt

public class StaticImports {

    public void calculateCircle(double radius) {
        // Using PI from Math.*
        double circumference = 2 * PI * radius;
        double area = PI * pow(radius, 2);

        // Using out from System.out
        out.println("Radius: " + radius);
        out.println("Circumference: " + circumference);
        out.println("Area: " + area);
    }

    public void trigonometry() {
        // Using sin, cos, sqrt from Math.*
        double angle = PI / 4;  // 45 degrees
        double sinValue = sin(angle);
        double cosValue = cos(angle);
        double tanValue = sinValue / cosValue;

        // Using format from String.format
        String result = format("sin(45°) = %.2f, cos(45°) = %.2f", sinValue, cosValue);
        out.println(result);

        // Using sqrt from Math.*
        double hypotenuse = sqrt(sinValue * sinValue + cosValue * cosValue);
        out.println("Hypotenuse: " + hypotenuse);
    }

    public void collectionsDemo() {
        // Using asList from Arrays.asList
        var names = asList("Alice", "Bob", "Charlie");

        // Using stream collectors from Collectors.*
        String joined = names.stream()
            .map(String::toUpperCase)
            .collect(joining(", "));  // Using joining from Collectors

        out.println("Names: " + joined);

        // Using toList from Collectors
        var lengths = names.stream()
            .map(String::length)
            .collect(toList());

        out.println("Lengths: " + lengths);
    }

    public static void main(String[] args) {
        StaticImports demo = new StaticImports();
        demo.calculateCircle(5.0);
        demo.trigonometry();
        demo.collectionsDemo();
    }
}