//
//  01_basic.swift
//  DistillerSample
//
//  Demonstrates: constants, variables, optionals, enums, simple functions.
//

import Foundation

/// Temperature‐related helper utilities.
enum TemperatureScale: String {
    case celsius = "°C"
    case fahrenheit = "°F"
}

/// An error that can occur while converting temperatures.
///
/// Even though errors are introduced later in depth,
/// having a minimal `Error` here provides optional handling practice.
enum TemperatureError: Error {
    case negativeAbsolute
}

/// Converts a temperature value from one scale to another.
///
/// - Parameters:
///   - value: The numeric temperature to convert.
///   - from:  Original scale.
///   - to:    Desired scale.
/// - Returns: The converted value, or `nil` if conversion is impossible (e.g. below absolute zero).
func convert(_ value: Double,
             from: TemperatureScale,
             to: TemperatureScale) -> Double? {

    let absoluteZeroCelsius = -273.15
    let absoluteZeroFahrenheit = -459.67

    switch (from, to) {
    case (.celsius, .fahrenheit):
        guard value >= absoluteZeroCelsius else { return nil }
        return (value * 9/5) + 32
    case (.fahrenheit, .celsius):
        guard value >= absoluteZeroFahrenheit else { return nil }
        return (value - 32) * 5/9
    default:
        // Same scale; no conversion needed.
        return value
    }
}

/// Entry-point for quick playground testing.
///
/// Not marked `public` to keep visibility minimal for a basic example.
fileprivate func demo() {
    let celsius: Double = 22
    if let f = convert(celsius, from: .celsius, to: .fahrenheit) {
        print("\(celsius)°C → \(f)°F")
    }

    // Optional binding with early exit.
    guard let impossible = convert(-500, from: .celsius, to: .fahrenheit) else {
        print("❌ Below absolute zero – conversion refused")
        return
    }
    print(impossible) // never reached
}

/// Private helper for validation
private func validateTemperature(_ value: Double, scale: TemperatureScale) -> Bool {
    switch scale {
    case .celsius:
        return value >= -273.15
    case .fahrenheit:
        return value >= -459.67
    }
}

/// Internal constant for configuration
internal let defaultScale: TemperatureScale = .celsius

demo()