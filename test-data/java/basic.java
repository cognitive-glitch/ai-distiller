package com.example.demo;

import java.util.List;
import java.util.ArrayList;
import java.util.stream.Collectors;
import static java.lang.Math.max;
import static java.lang.System.*;

/**
 * Example Java class demonstrating various language features
 */
@Deprecated
@SuppressWarnings("unchecked")
public class UserService {
    // Constants
    public static final String VERSION = "1.0.0";
    private static final int MAX_USERS = 1000;
    
    // Fields
    private final UserRepository repository;
    protected List<User> cache;
    volatile int accessCount;
    
    // Constructor
    public UserService(UserRepository repository) {
        this.repository = repository;
        this.cache = new ArrayList<>();
        this.accessCount = 0;
    }
    
    // Public methods
    public User findUser(Long id) {
        accessCount++;
        return repository.findById(id)
            .orElseThrow(() -> new UserNotFoundException("User not found: " + id));
    }
    
    @Override
    public String toString() {
        return "UserService{version=" + VERSION + ", users=" + cache.size() + "}";
    }
    
    // Protected method
    protected void clearCache() {
        cache.clear();
    }
    
    // Private method
    private void validateUser(User user) {
        if (user == null || user.getName() == null) {
            throw new IllegalArgumentException("Invalid user");
        }
    }
    
    // Static nested class
    public static class Statistics {
        private int totalUsers;
        private int activeUsers;
        
        public Statistics(int totalUsers, int activeUsers) {
            this.totalUsers = totalUsers;
            this.activeUsers = activeUsers;
        }
        
        public double getActivePercentage() {
            return (double) activeUsers / totalUsers * 100;
        }
    }
    
    // Inner class (non-static)
    private class CacheManager {
        void evictOldEntries() {
            cache.removeIf(user -> user.getLastAccess().isBefore(java.time.LocalDateTime.now().minusDays(7)));
        }
    }
}

// Interface
interface UserRepository {
    java.util.Optional<User> findById(Long id);
    List<User> findAll();
    void save(User user);
    void delete(Long id);
}

// Enum
public enum UserRole {
    ADMIN("Administrator"),
    USER("Regular User"),
    GUEST("Guest");
    
    private final String displayName;
    
    UserRole(String displayName) {
        this.displayName = displayName;
    }
    
    public String getDisplayName() {
        return displayName;
    }
}

// Record (Java 14+)
public record User(Long id, String name, String email, UserRole role) {
    // Compact constructor with validation
    public User {
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("Name cannot be blank");
        }
    }
    
    // Additional method
    public boolean isAdmin() {
        return role == UserRole.ADMIN;
    }
}

// Sealed class (Java 15+)
public sealed abstract class Payment permits CreditCardPayment, PayPalPayment, BankTransferPayment {
    protected final double amount;
    
    protected Payment(double amount) {
        this.amount = amount;
    }
    
    public abstract void process();
}

final class CreditCardPayment extends Payment {
    private final String cardNumber;
    
    CreditCardPayment(double amount, String cardNumber) {
        super(amount);
        this.cardNumber = cardNumber;
    }
    
    @Override
    public void process() {
        out.println("Processing credit card payment of " + amount);
    }
}

final class PayPalPayment extends Payment {
    private final String email;
    
    PayPalPayment(double amount, String email) {
        super(amount);
        this.email = email;
    }
    
    @Override
    public void process() {
        out.println("Processing PayPal payment of " + amount);
    }
}

non-sealed class BankTransferPayment extends Payment {
    private final String accountNumber;
    
    BankTransferPayment(double amount, String accountNumber) {
        super(amount);
        this.accountNumber = accountNumber;
    }
    
    @Override
    public void process() {
        out.println("Processing bank transfer of " + amount);
    }
}

// Generic class
class Container<T extends Comparable<T>> {
    private T value;
    
    public void setValue(T value) {
        this.value = value;
    }
    
    public T getValue() {
        return value;
    }
    
    public <U> U transform(java.util.function.Function<T, U> transformer) {
        return transformer.apply(value);
    }
}

// Exception class
class UserNotFoundException extends RuntimeException {
    public UserNotFoundException(String message) {
        super(message);
    }
    
    public UserNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }
}

// Functional interface
@FunctionalInterface
interface UserValidator {
    boolean validate(User user);
    
    // Default method
    default UserValidator and(UserValidator other) {
        return user -> this.validate(user) && other.validate(user);
    }
    
    // Static method
    static UserValidator emailValidator() {
        return user -> user.email() != null && user.email().contains("@");
    }
}

// Abstract class
abstract class BaseService {
    protected final String serviceName;
    
    protected BaseService(String serviceName) {
        this.serviceName = serviceName;
    }
    
    // Abstract method
    public abstract void initialize();
    
    // Concrete method
    public void shutdown() {
        out.println("Shutting down " + serviceName);
    }
}

// Class with varargs and annotations
class AnnotationExample {
    @SafeVarargs
    public final void processItems(List<String>... lists) {
        for (List<String> list : lists) {
            list.forEach(out::println);
        }
    }
    
    @SuppressWarnings({"unchecked", "rawtypes"})
    public void legacyMethod(List items) {
        items.add("legacy");
    }
}