//
//  04_complex.swift
//  DistillerSample
//
//  Demonstrates: property wrappers, result builders, opaque return types,
//  advanced generics, Combine framework, KeyPaths, dynamic member lookup,
//  advanced protocol extensions.
//

import Foundation
import Combine

// MARK: - Property Wrapper & Combine Integration

/// A property wrapper that clamps a numeric value within a given range
/// and publishes changes via a Combine publisher.
@propertyWrapper
public struct Clamped<Value: Numeric & Comparable> {
    private var value: Value
    private let range: ClosedRange<Value>
    private let subject = PassthroughSubject<Value, Never>()

    public var wrappedValue: Value {
        get { value }
        set {
            let clampedValue = max(range.lowerBound, min(newValue, range.upperBound))
            if clampedValue != value {
                value = clampedValue
                subject.send(value)
            }
        }
    }

    /// The projected value is a Combine publisher that emits new values.
    public var projectedValue: AnyPublisher<Value, Never> {
        subject.eraseToAnyPublisher()
    }

    public init(wrappedValue: Value, _ range: ClosedRange<Value>) {
        self.range = range
        self.value = max(range.lowerBound, min(wrappedValue, range.upperBound))
    }

    /// Private helper for validation
    private func isInRange(_ value: Value) -> Bool {
        return range.contains(value)
    }
}

// MARK: - Result Builder

/// A simple validation result type.
public enum ValidationResult {
    case success(String)
    case failure(String)

    /// Private helper for result categorization
    private var isSuccess: Bool {
        switch self {
        case .success: return true
        case .failure: return false
        }
    }
}

/// A result builder for creating an array of `ValidationResult` declaratively.
@resultBuilder
public struct ConfigurationBuilder {
    public static func buildBlock(_ components: ValidationResult...) -> [ValidationResult] {
        components
    }
    
    public static func buildExpression(_ expression: String) -> ValidationResult {
        .success(expression)
    }
    
    public static func buildEither(first component: [ValidationResult]) -> [ValidationResult] {
        component
    }
    
    public static func buildEither(second component: [ValidationResult]) -> [ValidationResult] {
        component
    }

    /// Private helper for building optional components
    private static func buildOptional(_ component: [ValidationResult]?) -> [ValidationResult] {
        component ?? []
    }
}

// MARK: - Protocols, Opaque Types, and Extensions

/// A protocol for items that can be described.
public protocol Describable {
    /// A textual description of the instance.
    var description: String { get }
    
    /// Resets to a default state.
    func reset()
}

// Advanced Protocol Extension with default implementation
extension Describable {
    public var description: String {
        "This is a default description for a Describable item."
    }

    /// Private helper for description validation
    private func validateDescription() -> Bool {
        return !description.isEmpty
    }
}

fileprivate struct AudioSettings: Describable, Equatable {
    @Clamped(0...100) var volume: Int = 50
    var isMuted: Bool = false
    
    // Conforming to Describable
    var description: String { "Audio: Volume \(volume), Muted: \(isMuted)" }
    func reset() { 
        volume = 50
        isMuted = false
    }

    /// Private validation method
    private func validateAudioSettings() -> Bool {
        return volume >= 0 && volume <= 100
    }
}

// MARK: - Dynamic Member Lookup & KeyPaths

/// A read-only proxy for accessing UserSettings.
@dynamicMemberLookup
public struct SettingsProxy {
    private let settings: UserSettings
    
    fileprivate init(_ settings: UserSettings) {
        self.settings = settings
    }
    
    public subscript<T>(dynamicMember keyPath: KeyPath<UserSettings, T>) -> T {
        settings[keyPath: keyPath]
    }

    /// Private helper for proxy validation
    private func isValidProxy() -> Bool {
        return true // Always valid for now
    }
}

// MARK: - Core Logic and Feature Integration

/// Main settings structure for a user.
public struct UserSettings {
    var username: String
    @Clamped(18...120) var userAge: Int = 30
    fileprivate var audio = AudioSettings()
    
    public init(username: String) {
        self.username = username
    }
    
    /// Creates a read-only proxy for these settings.
    public func asProxy() -> SettingsProxy {
        SettingsProxy(self)
    }

    /// Private validation method
    private func validateSettings() -> Bool {
        return !username.isEmpty && userAge >= 18
    }

    /// Internal method for settings synchronization
    internal mutating func syncWithRemote() {
        // Sync implementation
    }
}

internal class SettingsViewModel {
    private var settings: UserSettings
    private var cancellables = Set<AnyCancellable>()
    
    @Published private(set) var ageDescription: String
    
    init(settings: UserSettings) {
        self.settings = settings
        self.ageDescription = "Current age: \(settings.userAge)"
        
        // Combine: Subscribing to the property wrapper's projected value
        self.settings.$userAge
            .map { "Updated age: \($0)" }
            .assign(to: \.ageDescription, on: self)
            .store(in: &cancellables)
    }
    
    // Opaque Return Type
    internal func getAudioConfiguration() -> some Equatable {
        // The caller knows it's Equatable, but not that it's an `AudioSettings`.
        return self.settings.audio
    }
    
    // Using the Result Builder
    @ConfigurationBuilder
    internal func validate(isPremiumUser: Bool) -> [ValidationResult] {
        "Username check passed"
        
        if isPremiumUser {
            "Premium user features enabled"
        } else {
            ValidationResult.failure("Premium features disabled")
        }
    }
    
    // Advanced Generics with `where` clause
    /// Serializes items that are both Identifiable and Codable.
    public func serialize<T: Collection>(_ items: T) -> Data? where T.Element: Identifiable & Codable {
        let encoder = JSONEncoder()
        return try? encoder.encode(items.map { $0.id })
    }

    /// Private method for configuration cleanup
    private func cleanupConfiguration() {
        cancellables.removeAll()
    }
}

// MARK: - Demo

fileprivate func complexDemo() {
    var settings = UserSettings(username: "TestUser")
    let viewModel = SettingsViewModel(settings: settings)
    
    // Test property wrapper
    settings.userAge = 150 // Will be clamped to 120
    
    // Test opaque return type
    let audioConfig = viewModel.getAudioConfiguration()
    print("Audio config: \(audioConfig)")
    
    // Test result builder
    let validationResults = viewModel.validate(isPremiumUser: true)
    print("Validation results: \(validationResults)")
    
    // Test dynamic member lookup
    let proxy = settings.asProxy()
    print("Username from proxy: \(proxy.username)")
}

complexDemo()