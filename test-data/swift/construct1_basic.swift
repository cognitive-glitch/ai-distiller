// MARK: - Construct 1: Basic Fundamentals

/// A public function to test basic arithmetic and parameter handling.
/// The distiller should identify parameters, return type, and the function's purpose from the docstring.
public func addTwoIntegers(a: Int, b: Int) -> Int {
    return a + b
}

/// An internal function demonstrating optional types, basic string interpolation,
/// and control flow with `guard let`. This is more idiomatic Swift than `if let` for early exits.
internal func createGreeting(for name: String?) -> String {
    // Test 1: Optional binding with `guard`. A key Swift pattern.
    guard let providedName = name, !providedName.isEmpty else {
        return "Hello, anonymous!"
    }
    
    // Test 2: Basic string interpolation.
    return "Hello, \(providedName)!"
}

// Test 3: Type inference for constants and variables.
let inferredConstant = "This is a String"
var inferredVariable = 1_000 // Tests numeric literals with separators.

// Test 4: File-private scope. How does the distiller differentiate this from `private`?
fileprivate func utilityHelper() {
    // This function is only visible within this source file.
    print("Performing a file-private utility task.")
}