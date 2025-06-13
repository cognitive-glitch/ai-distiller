// MARK: - Construct 3: Protocols & Inheritance

// Test 1: A protocol with properties and methods.
protocol Loggable {
    var logPrefix: String { get }
    func log(message: String)
}

// Test 2: Protocol extension with a default implementation.
extension Loggable {
    func log(message: String) {
        print("\(logPrefix): \(message)")
    }
}

// Test 3: A base class providing some functionality.
open class MediaAsset {
    private let uniqueID = UUID()
    
    open func getMetadata() -> [String: Any] {
        return ["id": uniqueID.uuidString]
    }
}

// Test 4: A subclass that inherits AND conforms to a protocol.
final class VideoAsset: MediaAsset, Loggable {
    let title: String
    var logPrefix: String { "Video(\(title))" }
    
    init(title: String) {
        self.title = title
        super.init()
    }
    
    // Test 5: Overriding a method from a superclass.
    override func getMetadata() -> [String: Any] {
        var metadata = super.getMetadata()
        metadata["title"] = self.title
        return metadata
    }
}