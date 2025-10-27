// Test Pattern 5: Complex Import Patterns
// Tests nested classes, generics, lambdas, and package-info imports

package com.example.importtest;

import com.example.project.models.User;
import com.example.project.models.User.Address;  // Nested class import
import com.example.project.models.User.Status;   // Nested enum import
import com.example.project.services.UserService;
import com.example.project.dao.UserDao;
import java.awt.List;  // Name clash with java.util.List
import java.util.function.Function;
import java.util.function.Predicate;
import java.util.function.Consumer;
import java.util.function.Supplier;
import java.util.stream.Stream;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import javax.inject.Inject;
import javax.inject.Singleton;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Not using: UserDao, java.awt.List, Consumer, Supplier, TimeUnit

@Singleton
public class ComplexImports {

    private static final Logger LOGGER = LoggerFactory.getLogger(ComplexImports.class);

    @Inject
    private UserService userService;

    public void demonstrateNestedClasses() {
        // Using User and nested classes
        User user = new User("John Doe");

        // Using nested Address class
        Address address = new Address("123 Main St", "City");
        user.setAddress(address);

        // Using nested Status enum
        user.setStatus(Status.ACTIVE);

        LOGGER.info("Created user: {}", user);
    }

    public void demonstrateFunctionalInterfaces() {
        // Using Function
        Function<String, Integer> stringLength = String::length;

        // Using Predicate
        Predicate<User> isActive = user -> user.getStatus() == Status.ACTIVE;

        // Using Stream with functional interfaces
        Stream.of("apple", "banana", "cherry")
            .map(stringLength)  // Using Function
            .filter(len -> len > 5)  // Inline predicate
            .forEach(System.out::println);
    }

    public CompletableFuture<User> demonstrateAsync() {
        // Using CompletableFuture
        return CompletableFuture.supplyAsync(() -> {
            LOGGER.debug("Fetching user asynchronously");
            return userService.findById(1L);
        }).thenApply(user -> {
            // Transform the user
            user.setStatus(Status.VERIFIED);
            return user;
        }).exceptionally(throwable -> {
            LOGGER.error("Error fetching user", throwable);
            return null;
        });
    }

    // Generic method using imported types
    public <T> Stream<T> filterAndLog(Stream<T> stream, Predicate<T> filter) {
        return stream
            .filter(filter)
            .peek(item -> LOGGER.trace("Processing item: {}", item));
    }
}

// Additional class in same file using imports
class ImportHelper {
    // This class uses User import
    public void process(User user) {
        System.out.println("Processing: " + user.getName());
    }
}