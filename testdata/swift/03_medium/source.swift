//
//  03_medium.swift
//  DistillerSample
//
//  Demonstrates: generics, extensions, protocol with associatedtype,
//  constrained extensions, custom error types, higher-order functions.
//

import Foundation

// MARK: - Generic Stack

/// A simple LIFO stack.
public struct Stack<Element> {
    private var storage: [Element] = []

    public init() {}

    /// Pushes a new element.
    public mutating func push(_ element: Element) {
        storage.append(element)
    }

    /// Returns the top element without removing it.
    public func peek() -> Element? { storage.last }

    /// Removes and returns the top element.
    public mutating func pop() -> Element? { storage.popLast() }

    /// Private helper for validation
    private func isValid() -> Bool {
        return storage.count >= 0
    }

    /// Internal method for debugging
    internal var count: Int {
        return storage.count
    }
}

// Constrained extension adding `average()` only when `Element` is numeric.
public extension Stack where Element: BinaryInteger {
    /// Computes the arithmetic mean of all integers in the stack.
    /// - Throws: `MathError.emptyStack` if there are no elements.
    func average() throws -> Double {
        guard !storage.isEmpty else { throw MathError.emptyStack }
        let sum = storage.reduce(0, +)
        return Double(sum) / Double(storage.count)
    }

    /// Private helper for numeric validation
    private func validateNumericElements() -> Bool {
        return storage.allSatisfy { $0 >= 0 }
    }
}

/// Errors for math-centric operations.
public enum MathError: Error, CustomStringConvertible {
    case emptyStack
    case divideByZero
    case negativeValue

    public var description: String {
        switch self {
        case .emptyStack:   return "Stack contains no elements."
        case .divideByZero: return "Attempted division by zero."
        case .negativeValue: return "Negative values not allowed."
        }
    }

    /// Private helper for error categorization
    private var isRecoverable: Bool {
        switch self {
        case .emptyStack, .negativeValue: return true
        case .divideByZero: return false
        }
    }
}

// MARK: - Protocol & Extension with AssociatedType

/// Describes a cache that stores key/value pairs.
public protocol Cacheable {
    associatedtype Key: Hashable
    associatedtype Value

    mutating func insert(_ value: Value, for key: Key)
    func value(for key: Key) -> Value?
    func contains(key: Key) -> Bool
}

/// Dictionary already satisfies `Cacheable` when paired properly.
extension Dictionary: Cacheable {
    public mutating func insert(_ value: Value, for key: Key) {
        self[key] = value
    }

    public func contains(key: Key) -> Bool {
        return self[key] != nil
    }
}

/// Generic cache with expiration
public struct ExpiringCache<Key: Hashable, Value>: Cacheable {
    private struct CacheEntry {
        let value: Value
        let expiration: Date
    }

    private var storage: [Key: CacheEntry] = [:]
    private let timeToLive: TimeInterval

    public init(timeToLive: TimeInterval = 300) {
        self.timeToLive = timeToLive
    }

    public mutating func insert(_ value: Value, for key: Key) {
        let expiration = Date().addingTimeInterval(timeToLive)
        storage[key] = CacheEntry(value: value, expiration: expiration)
    }

    public func value(for key: Key) -> Value? {
        guard let entry = storage[key] else { return nil }
        return Date() < entry.expiration ? entry.value : nil
    }

    public func contains(key: Key) -> Bool {
        return value(for: key) != nil
    }

    /// Private cleanup method
    private mutating func removeExpiredEntries() {
        let now = Date()
        storage = storage.filter { $0.value.expiration > now }
    }

    /// Internal method for cache statistics
    internal var cacheSize: Int {
        return storage.count
    }
}

// MARK: - Demo

fileprivate func mediumDemo() {
    var intStack = Stack<Int>()
    (1...5).forEach(intStack.push)

    do {
        let avg = try intStack.average()
        print("Average:", avg)
    } catch {
        print("⚠️", error)
    }

    var cache: [String: URL] = [:]
    cache.insert(URL(string: "https://swift.org")!, for: "swift")
    print("Cache lookup:", cache.value(for: "swift") as Any)

    var expiringCache = ExpiringCache<String, String>(timeToLive: 60)
    expiringCache.insert("cached_value", for: "test_key")
    print("Expiring cache contains key:", expiringCache.contains(key: "test_key"))
}

mediumDemo()