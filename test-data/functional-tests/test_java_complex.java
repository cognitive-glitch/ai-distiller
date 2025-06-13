package com.example.test;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import static java.lang.System.out;

/**
 * Complex Java test file for AI Distiller functional testing
 */
@Entity
@Table(name = "users")
public class ComplexJavaClass extends BaseEntity implements Serializable, Comparable<ComplexJavaClass> {
    
    // Static fields
    public static final String CONSTANT = "TEST";
    private static final Logger logger = LoggerFactory.getLogger(ComplexJavaClass.class);
    
    // Instance fields
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String name;
    
    protected transient volatile Map<String, Object> cache = new ConcurrentHashMap<>();
    
    // Constructor
    public ComplexJavaClass(String name) {
        this.name = Objects.requireNonNull(name, "Name cannot be null");
        logger.info("Created instance with name: {}", name);
    }
    
    // Private constructor
    private ComplexJavaClass() {
        this.name = "default";
    }
    
    // Public methods
    @Override
    public int compareTo(ComplexJavaClass other) {
        return this.name.compareTo(other.name);
    }
    
    @Transactional
    public Optional<String> processData(List<String> data, boolean validate) throws ProcessingException {
        if (validate && !isValid(data)) {
            throw new ProcessingException("Invalid data provided");
        }
        
        return data.stream()
            .filter(Objects::nonNull)
            .map(String::toUpperCase)
            .findFirst();
    }
    
    // Protected method
    protected boolean isValid(List<String> data) {
        return data != null && !data.isEmpty();
    }
    
    // Private method
    private void logOperation(String operation) {
        logger.debug("Executing operation: {}", operation);
    }
    
    // Static method
    public static ComplexJavaClass createDefault() {
        return new ComplexJavaClass();
    }
    
    // Generic method
    public <T extends Comparable<T>> List<T> sortItems(Collection<T> items) {
        return items.stream()
            .sorted()
            .collect(Collectors.toList());
    }
    
    // Getters and setters
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
}

// Interface
interface DataProcessor<T> {
    void process(T data);
    default boolean canProcess(T data) { return data != null; }
}

// Enum
public enum Status {
    ACTIVE("active"),
    INACTIVE("inactive"),
    PENDING("pending");
    
    private final String value;
    
    Status(String value) {
        this.value = value;
    }
    
    public String getValue() { return value; }
}

// Record (Java 14+)
public record UserInfo(String name, int age, String email) {
    public UserInfo {
        Objects.requireNonNull(name, "Name cannot be null");
        if (age < 0) throw new IllegalArgumentException("Age cannot be negative");
    }
    
    public boolean isAdult() {
        return age >= 18;
    }
}

// Abstract class
abstract class BaseEntity {
    protected abstract void validate();
    
    public final void save() {
        validate();
        // Save logic here
    }
}