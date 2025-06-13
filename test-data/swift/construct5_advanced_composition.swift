// MARK: - Construct 5: Advanced Composition & Concurrency

import Foundation

// A protocol that our service's model objects must conform to.
// The `Sendable` constraint is critical for safe use across concurrency domains.
protocol IdentifiableModel: Codable, Sendable {
    associatedtype ID: Codable & Hashable
    var id: ID { get }
}

struct Product: IdentifiableModel {
    let id: Int
    let name: String
    let price: Double
}

// A result builder for constructing request headers.
@resultBuilder
struct HeaderBuilder {
    static func buildBlock(_ components: (String, String)...) -> [String: String] {
        return Dictionary(uniqueKeysWithValues: components)
    }
}

// The main, complex component. An actor for thread-safe state management.
actor APIService<Model: IdentifiableModel> {
    private let baseURL: URL
    private var cache: [Model.ID: Model] = [:]
    
    // Custom initializer using the result builder
    init(baseURL: URL, @HeaderBuilder headers: () -> [String: String]) {
        self.baseURL = baseURL
        // In a real scenario, we'd use these headers.
        print("Service configured with headers: \(headers())")
    }
    
    // An async method that can throw, returning an opaque type.
    // Opaque types (`some`) are a key abstraction feature.
    func fetch(id: Model.ID) async throws -> some IdentifiableModel {
        if let cached = cache[id] {
            return cached
        }
        
        let url = baseURL.appendingPathComponent("\(id)")
        let (data, _) = try await URLSession.shared.data(from: url)
        
        // This assumes Model is Decodable, which is guaranteed by the constraint.
        let model = try JSONDecoder().decode(Model.self, from: data)
        cache[id] = model // Safe mutation within the actor
        return model
    }
    
    // A non-async method to demonstrate actor isolation.
    // This method must be `nonisolated` if it doesn't touch actor state.
    nonisolated func getEndpoint(for id: Model.ID) -> URL {
        return baseURL.appendingPathComponent("\(id)")
    }
}

// Conditional conformance: Add functionality only if the Model meets extra criteria.
extension APIService where Model == Product {
    // A specific method only available for the Product service.
    func findMostExpensive(in ids: [Int]) async throws -> Product? {
        var mostExpensive: Product?
        for id in ids {
            if let product = try await fetch(id: id) as? Product {
                if mostExpensive == nil || product.price > mostExpensive!.price {
                    mostExpensive = product
                }
            }
        }
        return mostExpensive
    }
}