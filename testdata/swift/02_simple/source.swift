//
//  02_simple.swift
//  DistillerSample
//
//  Demonstrates: value vs. reference semantics, protocol conformance,
//  property observers, access control.
//

import Foundation

// MARK: - Protocols

/// Something that can provide a textual description of itself.
public protocol Describable {
    var description: String { get }
}

/// Two-dimensional, integer-backed positioning.
public struct Point: Describable, Equatable, Hashable {
    public var x: Int
    public var y: Int

    /// Euclidean distance from origin (0,0).
    public var magnitude: Double {
        sqrt(Double(x * x + y * y))
    }

    public var description: String {
        "(\(x), \(y))"
    }

    /// Private helper for coordinate validation
    private func isValid() -> Bool {
        return x >= 0 && y >= 0
    }

    /// Internal method for coordinate transformation
    internal mutating func transform(by offset: Point) {
        x += offset.x
        y += offset.y
    }
}

/// Models an axis-aligned rectangle.
open class Rectangle: Describable {
    // Stored properties
    public private(set) var origin: Point
    public var size: (width: Int, height: Int) {
        didSet { areaCache = nil }     // Invalidate cache when size changes
    }

    // Lazy-calculated property cached for performance
    private var areaCache: Int?

    /// Designated initializer.
    public init(origin: Point = .init(x: 0, y: 0), width: Int, height: Int) {
        self.origin = origin
        self.size = (width, height)
    }

    /// Area is recomputed only when cache invalidated.
    public var area: Int {
        if let cached = areaCache { return cached }
        let computed = size.width * size.height
        areaCache = computed
        return computed
    }

    open var description: String {
        "Rect@\(origin) \(size.width)x\(size.height)"
    }

    /// Private method for bounds checking
    private func contains(point: Point) -> Bool {
        let endX = origin.x + size.width
        let endY = origin.y + size.height
        return point.x >= origin.x && point.x <= endX &&
               point.y >= origin.y && point.y <= endY
    }

    /// Protected method for subclasses
    internal func updateOrigin(to newOrigin: Point) {
        origin = newOrigin
    }
}

/// Specialized rectangle for UI elements
public class UIRectangle: Rectangle {
    private var isVisible: Bool = true

    /// Override with additional behavior
    public override var description: String {
        let base = super.description
        return "\(base) [visible: \(isVisible)]"
    }

    /// Public method specific to UI rectangles
    public func show() {
        isVisible = true
    }

    /// Public method to hide rectangle
    public func hide() {
        isVisible = false
    }

    /// Private UI-specific validation
    private func validateForDisplay() -> Bool {
        return size.width > 0 && size.height > 0 && isVisible
    }
}

// MARK: - Demo

internal func descriptorDemo() {
    let p = Point(x: 3, y: 4)
    let r = Rectangle(width: 10, height: 20)

    let describables: [Describable] = [p, r]
    describables.forEach { print($0.description) }
}

descriptorDemo()