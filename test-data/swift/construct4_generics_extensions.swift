// MARK: - Construct 4: Generics & Extensions

// Test 1: A generic data container struct.
struct PaginatedResponse<T: Codable> {
    let page: Int
    let totalPages: Int
    let items: [T]
}

// Test 2: A generic function with a protocol constraint.
func findFirst<C: Collection>(in collection: C, where predicate: (C.Element) -> Bool) -> C.Element? {
    for element in collection {
        if predicate(element) {
            return element
        }
    }
    return nil
}

// Test 3: An extension on an existing type (our UserProfile from Construct 2).
// This tests the ability to link code across different declarations.
extension UserProfile {
    // Add a convenience initializer.
    init(id: UUID, username: String) {
        self.id = id
        self.username = username
    }
    
    // Add a static factory method.
    static func createGuest() -> UserProfile {
        // We can safely unwrap here because "guest" is > 3 chars.
        return UserProfile(username: "guest")!
    }
}

// Test 4: A simple property wrapper.
@propertyWrapper
struct Trimmed {
    private var value: String = ""
    
    var wrappedValue: String {
        get { value }
        set { value = newValue.trimmingCharacters(in: .whitespacesAndNewlines) }
    }
    
    init(wrappedValue: String) {
        self.wrappedValue = wrappedValue
    }
}

struct RegistrationForm {
    @Trimmed var email: String
}