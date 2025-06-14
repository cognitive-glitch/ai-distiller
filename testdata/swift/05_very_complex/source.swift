//
//  05_very_complex.swift
//  DistillerSample
//
//  Demonstrates: actors, advanced generics with type erasure, custom operators,
//  metaprogramming with Mirror, advanced memory management, concurrent programming,
//  protocol compositions, custom collection types.
//

import Foundation

// MARK: - Core Protocols with Associated Types

/// A protocol defining a generic event with a specific payload.
public protocol Event {
    associatedtype Payload
    var payload: Payload { get }
    static var name: String { get }
}

extension Event {
    public static var name: String { String(describing: Self.self) }

    /// Private helper for event validation
    private func isValidEvent() -> Bool {
        return true // Always valid for this example
    }
}

/// A protocol for a type that can handle a specific kind of event.
public protocol EventHandler: AnyObject {
    associatedtype HandledEvent: Event
    func handle(event: HandledEvent) async
}

// MARK: - Type Erasure for Heterogeneous Storage

/// A weak reference wrapper to prevent retain cycles in the subscriber list.
fileprivate struct Weak<T: AnyObject> {
    weak var value: T?
    init(_ value: T) { self.value = value }

    /// Private validation helper
    private var isAlive: Bool {
        return value != nil
    }
}

/// A type-erased wrapper for any `EventHandler`.
/// This allows storing handlers for different event types in the same collection.
private struct AnyEventHandler {
    private let _handle: (any Event) async -> Void
    private let canHandle: (any Event.Type) -> Bool
    let objectId: ObjectIdentifier
    
    init<H: EventHandler>(_ handler: H) {
        self._handle = { event in
            if let concreteEvent = event as? H.HandledEvent {
                await handler.handle(event: concreteEvent)
            }
        }
        self.canHandle = { eventType in
            eventType == H.HandledEvent.self
        }
        self.objectId = ObjectIdentifier(handler)
    }
    
    func handle(event: any Event) async {
        await _handle(event)
    }

    /// Private helper for handler validation
    private func canHandleEvent<E: Event>(_ eventType: E.Type) -> Bool {
        return canHandle(eventType)
    }
}

// MARK: - The Core Concurrent Component: The Actor

/// An actor that manages event subscriptions and dispatches events concurrently.
public actor EventBus {
    private var handlers: [String: [Weak<AnyObject>]] = [:]
    private var handlerMap: [ObjectIdentifier: AnyEventHandler] = [:]
    private var eventHistory: [String] = []

    public func subscribe<H: EventHandler>(_ handler: H) {
        let eventName = H.HandledEvent.name
        let weakHandler = Weak(handler)
        let anyHandler = AnyEventHandler(handler)
        
        handlers[eventName, default: []].append(weakHandler)
        handlerMap[anyHandler.objectId] = anyHandler
        
        // Clean up nil references
        handlers[eventName]?.removeAll { $0.value == nil }
    }
    
    public func post<E: Event>(_ event: E) async {
        let eventName = E.name
        eventHistory.append(eventName)
        
        guard let potentialHandlers = handlers[eventName] else { return }
        
        // Dispatch to all valid, non-nil handlers concurrently
        await withTaskGroup(of: Void.self) { group in
            for weakHandler in potentialHandlers {
                if let handlerId = weakHandler.value.map(ObjectIdentifier.init),
                   let anyHandler = handlerMap[handlerId] {
                    group.addTask {
                        await anyHandler.handle(event: event)
                    }
                }
            }
        }
    }

    /// Private method for handler cleanup
    private func cleanupExpiredHandlers() {
        for (eventName, handlerList) in handlers {
            handlers[eventName] = handlerList.filter { $0.value != nil }
        }
    }

    /// Internal method for bus statistics
    internal var handlerCount: Int {
        return handlerMap.count
    }
}

// MARK: - Custom Operators and Precedence

precedencegroup SubscriptionPrecedence {
    associativity: left
    higherThan: AssignmentPrecedence
}

infix operator ~>: SubscriptionPrecedence

/// Custom operator for subscribing an EventHandler to an EventBus.
public func ~> <H: EventHandler>(handler: H, bus: EventBus) {
    Task {
        await bus.subscribe(handler)
    }
}

// Custom operator for event composition
infix operator <>: AdditionPrecedence

public func <> <T, U>(lhs: T, rhs: U) -> (T, U) {
    return (lhs, rhs)
}

// MARK: - Metaprogramming with Mirror

/// Uses Mirror to generate a debug description of any event payload.
public func detailedDebugLog<T>(_ value: T) -> String {
    let mirror = Mirror(reflecting: value)
    var output = "[\(String(describing: T.self))] "
    
    for child in mirror.children {
        if let label = child.label {
            output += "\(label): \(child.value), "
        }
    }
    return String(output.dropLast(2)) // Remove trailing ", "
}

/// Advanced reflection utility
public struct ReflectionAnalyzer {
    /// Analyzes any object and returns its structure
    public static func analyze<T>(_ object: T) -> StructureInfo {
        let mirror = Mirror(reflecting: object)
        let properties = mirror.children.compactMap { child -> PropertyInfo? in
            guard let label = child.label else { return nil }
            return PropertyInfo(name: label, type: String(describing: type(of: child.value)))
        }
        
        return StructureInfo(
            typeName: String(describing: T.self),
            properties: properties,
            childCount: mirror.children.count
        )
    }

    /// Private helper for type analysis
    private static func analyzeType<T>(_ type: T.Type) -> String {
        return String(describing: type)
    }
}

public struct StructureInfo {
    public let typeName: String
    public let properties: [PropertyInfo]
    public let childCount: Int

    /// Private validation method
    private func isValid() -> Bool {
        return !typeName.isEmpty && childCount >= 0
    }
}

public struct PropertyInfo {
    public let name: String
    public let type: String

    /// Private helper for property validation
    private func validateProperty() -> Bool {
        return !name.isEmpty && !type.isEmpty
    }
}

// MARK: - Custom Collection Type

/// A basic thread-safe queue implementing Sequence.
public final class ThreadSafeQueue<T>: Sequence {
    private var elements: [T] = []
    private let lock = NSLock()

    public func enqueue(_ element: T) {
        lock.withLock {
            elements.append(element)
        }
    }

    public func dequeue() -> T? {
        lock.withLock {
            guard !elements.isEmpty else { return nil }
            return elements.removeFirst()
        }
    }
    
    public func makeIterator() -> IndexingIterator<[T]> {
        // Returns an iterator over a snapshot of the current state
        lock.withLock {
            return self.elements.makeIterator()
        }
    }

    /// Private helper for queue validation
    private func validateQueue() -> Bool {
        return lock.withLock { elements.count >= 0 }
    }

    /// Internal method for queue statistics
    internal var count: Int {
        return lock.withLock { elements.count }
    }
}

// Advanced collection with custom indexing
public struct CircularBuffer<Element>: Collection {
    private var storage: [Element?]
    private var head = 0
    private var tail = 0
    private let capacity: Int
    
    public var startIndex: Int { 0 }
    public var endIndex: Int { count }
    
    public init(capacity: Int) {
        self.capacity = capacity
        self.storage = Array(repeating: nil, count: capacity)
    }
    
    public func index(after i: Int) -> Int {
        return i + 1
    }
    
    public subscript(position: Int) -> Element {
        get {
            precondition(position < count, "Index out of range")
            let actualIndex = (head + position) % capacity
            return storage[actualIndex]!
        }
    }
    
    public var count: Int {
        return tail >= head ? tail - head : capacity - head + tail
    }

    /// Private helper for buffer validation
    private func isValidBuffer() -> Bool {
        return capacity > 0 && storage.count == capacity
    }

    /// Internal method for buffer manipulation
    internal mutating func append(_ element: Element) {
        storage[tail] = element
        tail = (tail + 1) % capacity
        if tail == head {
            head = (head + 1) % capacity
        }
    }
}

// MARK: - Advanced Memory Management and Concurrent Types

/// A service that demonstrates complex memory management patterns.
public final class ConcurrentEventLogger {
    // Strong reference to monitor
    public let monitor: ActivityMonitor
    
    // Weak reference to prevent cycles
    public weak var delegate: EventLoggerDelegate?
    
    // Unowned reference (careful usage required)
    private unowned let eventBus: EventBus
    
    private let queue = ThreadSafeQueue<String>()
    
    public init(monitor: ActivityMonitor, eventBus: EventBus) {
        self.monitor = monitor
        self.eventBus = eventBus
    }
    
    public func logEvent<E: Event>(_ event: E) async {
        let logEntry = detailedDebugLog(event.payload)
        queue.enqueue(logEntry)
        
        await monitor.record(logEntry)
        delegate?.didLogEvent(logEntry)
    }

    /// Private cleanup method
    private func cleanup() {
        // Cleanup resources
    }

    deinit {
        cleanup()
    }
}

/// Protocol for event logger delegation
public protocol EventLoggerDelegate: AnyObject {
    func didLogEvent(_ entry: String)
}

public actor ActivityMonitor {
    private(set) var logs = ThreadSafeQueue<String>()
    private var analysisCache: [String: StructureInfo] = [:]
    
    public func record(_ entry: String) {
        logs.enqueue(entry)
    }
    
    public func analyze<T>(_ object: T) -> StructureInfo {
        let key = String(describing: T.self)
        if let cached = analysisCache[key] {
            return cached
        }
        
        let analysis = ReflectionAnalyzer.analyze(object)
        analysisCache[key] = analysis
        return analysis
    }

    /// Private method for cache management
    private func clearCache() {
        analysisCache.removeAll()
    }

    /// Internal method for monitoring statistics
    internal var totalLogs: Int {
        return logs.count
    }
}

// MARK: - Example Usage

// Define some concrete events
struct UserLoggedInEvent: Event { 
    struct Payload { 
        let userId: UUID
        let name: String
        let timestamp: Date
    }
    let payload: Payload 
}

struct DataDownloadedEvent: Event { 
    let payload: Data 
}

// Define some concrete handlers
class UserActivityLogger: EventHandler {
    typealias HandledEvent = UserLoggedInEvent
    
    // unowned demo: this logger has a monitor reference
    unowned let monitor: ActivityMonitor
    
    init(monitor: ActivityMonitor) { 
        self.monitor = monitor 
    }
    
    func handle(event: UserLoggedInEvent) async {
        let log = detailedDebugLog(event.payload)
        print("UserActivityLogger handled: \(log)")
        await monitor.record(log)
    }

    /// Private helper for activity validation
    private func validateActivity() -> Bool {
        return true
    }
}

// MARK: - Demo

fileprivate func veryComplexDemo() async {
    let monitor = ActivityMonitor()
    let eventBus = EventBus()
    let logger = UserActivityLogger(monitor: monitor)
    
    // Use custom operator for subscription
    logger ~> eventBus
    
    // Create and post an event
    let event = UserLoggedInEvent(payload: .init(
        userId: UUID(),
        name: "TestUser",
        timestamp: Date()
    ))
    
    await eventBus.post(event)
    
    // Test reflection
    let analysis = await monitor.analyze(event.payload)
    print("Event analysis: \(analysis)")
    
    // Test custom collection
    let buffer = CircularBuffer<String>(capacity: 5)
    print("Buffer created with capacity: \(buffer.count)")
}

Task {
    await veryComplexDemo()
}