// Level 5: Advanced.java
package com.aidi.test.advanced;

import java.io.Serializable;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import java.nio.file.Files;
import java.nio.file.Path;

@Target(ElementType.PARAMETER)
@Retention(RetentionPolicy.RUNTIME)
@interface NotNull {}

public class Advanced<T extends Number & Runnable & Serializable> {

    // 1. Static initializer block
    static {
        System.out.println("Advanced class is being initialized.");
    }

    // 2. A field with a complex generic type
    private T complexField;

    // 3. Method with generic return type, bounded wildcard, and annotated parameter
    public <U extends T> U process(@NotNull U input) throws Exception {
        input.run(); // Call method from Runnable interface
        return input;
    }

    // 4. Anonymous inner class
    public Runnable createRunner(int value) {
        // 5. Local class inside a method
        class LocalRunner implements Runnable {
            @Override
            public void run() {
                System.out.println("Local runner executing with value: " + value);
            }
        }

        // Return an anonymous inner class instance
        return new Runnable() {
            private static final int MAX_ITERATIONS = 5;
            @Override
            public void run() {
                for (int i = 0; i < MAX_ITERATIONS; i++) {
                    System.out.println("Anonymous runner iteration " + i);
                }
            }
        };
    }

    // 6. Exception handling with multi-catch and try-with-resources
    public void readFile(String filePath) {
        try (var reader = Files.newBufferedReader(Path.of(filePath))) {
            // ...
        } catch (java.io.IOException | SecurityException e) {
            // Multi-catch block
            throw new RuntimeException("Failed to read file", e);
        }
    }
}