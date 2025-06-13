// Basic Swift example for testing

import Foundation
import SwiftUI
import Combine

// MARK: - Models

/// A person with a name and age
public struct Person: Codable, Identifiable {
    let id = UUID()
    public var name: String
    private var age: Int
    internal var email: String?
    
    public init(name: String, age: Int) {
        self.name = name
        self.age = age
    }
    
    public func description() -> String {
        return "\(name) is \(age) years old"
    }
    
    private func isAdult() -> Bool {
        return age >= 18
    }
}

// MARK: - Protocols

protocol Observable {
    associatedtype Output
    func observe() -> Output
}

protocol DataSource {
    func fetchData() async throws -> [Person]
}

// MARK: - Classes

open class ViewModel: ObservableObject {
    @Published public var people: [Person] = []
    @Published private(set) var isLoading = false
    
    private let dataSource: DataSource
    
    public init(dataSource: DataSource) {
        self.dataSource = dataSource
    }
    
    @MainActor
    public func loadData() async {
        isLoading = true
        defer { isLoading = false }
        
        do {
            people = try await dataSource.fetchData()
        } catch {
            print("Failed to load data: \(error)")
        }
    }
}

// MARK: - Extensions

extension Person: Equatable {
    public static func == (lhs: Person, rhs: Person) -> Bool {
        return lhs.id == rhs.id
    }
}

extension Array where Element: Identifiable {
    func findById(_ id: Element.ID) -> Element? {
        return first { $0.id == id }
    }
}

// MARK: - Enums

public enum Status {
    case pending
    case active(since: Date)
    case inactive(reason: String)
    
    var description: String {
        switch self {
        case .pending:
            return "Pending"
        case .active(let date):
            return "Active since \(date)"
        case .inactive(let reason):
            return "Inactive: \(reason)"
        }
    }
}

// MARK: - Generic Functions

func swap<T>(_ a: inout T, _ b: inout T) {
    let temp = a
    a = b
    b = temp
}

func filter<T>(_ items: [T], where predicate: (T) -> Bool) -> [T] {
    var result: [T] = []
    for item in items {
        if predicate(item) {
            result.append(item)
        }
    }
    return result
}

// MARK: - Async Functions

func fetchUser(id: String) async throws -> Person? {
    // Simulated network call
    try await Task.sleep(nanoseconds: 100_000_000)
    return Person(name: "John", age: 30)
}

// MARK: - Property Wrappers

@propertyWrapper
struct Capitalized {
    private var value: String = ""
    
    var wrappedValue: String {
        get { value }
        set { value = newValue.capitalized }
    }
}

// MARK: - Type Aliases

typealias UserID = String
typealias CompletionHandler = (Result<[Person], Error>) -> Void

// MARK: - Global Constants

let maxRetries = 3
public let apiVersion = "1.0.0"

// MARK: - Actors

actor DataCache {
    private var cache: [String: Person] = [:]
    
    func get(id: String) -> Person? {
        return cache[id]
    }
    
    func set(id: String, person: Person) {
        cache[id] = person
    }
}