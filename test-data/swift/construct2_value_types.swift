// MARK: - Construct 2: Value Types & State

// Test 1: Enum with associated values. A core Swift feature.
enum NetworkResponse {
    case success(data: Data)
    case failure(error: Error)
}

// Test 2: A struct with mixed properties and a mutating method.
struct UserProfile {
    // Stored properties
    let id: UUID
    private(set) var username: String // Writable internally, readable publicly.
    
    // Computed property
    var initials: String {
        username.prefix(2).uppercased()
    }
    
    // Failable initializer
    init?(username: String) {
        if username.count < 3 {
            return nil // Test failable initialization
        }
        self.id = UUID()
        self.username = username
    }
    
    // Mutating method required for value types
    mutating func updateUsername(_ newName: String) {
        guard !newName.isEmpty else { return }
        self.username = newName
    }
}