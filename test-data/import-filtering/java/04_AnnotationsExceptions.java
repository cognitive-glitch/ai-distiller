// Test Pattern 4: Imports for Annotations, Exceptions, and JavaDoc
// Tests imports used in annotations, throws clauses, and documentation

package com.example.importtest;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import java.lang.annotation.ElementType;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.sql.SQLException;
import java.util.concurrent.TimeoutException;
import java.util.concurrent.ExecutionException;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Min;
import javax.validation.constraints.Email;
import com.example.CustomException;
import java.util.logging.Logger;
import java.util.Optional;

// Not using: ExecutionException, Email, Logger, Optional

/**
 * Demonstrates annotation and exception imports.
 * 
 * @see java.util.Optional
 * @see Logger for logging functionality
 */
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.TYPE, ElementType.METHOD})
@interface MyAnnotation {
    String value() default "";
}

public class AnnotationsExceptions {
    
    @MyAnnotation("class-level")
    static class DataProcessor {
        
        @NotNull
        private String name;
        
        @Min(0)
        private int count;
        
        /**
         * Process a file and handle various exceptions.
         * 
         * @param filename the file to process
         * @throws FileNotFoundException if file doesn't exist
         * @throws IOException for general I/O errors
         * @throws SQLException if database error occurs
         * @throws TimeoutException if operation times out
         * @see {@link FileNotFoundException}
         * @see {@link Optional} for null-safe returns
         */
        @MyAnnotation("method-level")
        public void processFile(@NotNull String filename) 
                throws FileNotFoundException, IOException, SQLException, TimeoutException {
            
            // Simulate file check
            if (!filename.exists()) {
                throw new FileNotFoundException("File not found: " + filename);
            }
            
            // Simulate I/O operation
            if (filename.contains("corrupt")) {
                throw new IOException("Corrupted file");
            }
            
            // Simulate database operation
            if (filename.contains("db")) {
                throw new SQLException("Database connection failed");
            }
            
            // Simulate timeout
            if (filename.length() > 100) {
                throw new TimeoutException("Operation timed out");
            }
            
            // Using custom exception
            if (filename.isEmpty()) {
                throw new CustomException("Invalid filename");
            }
        }
        
        /**
         * Validates email using annotation.
         * See {@link javax.validation.constraints.Email} for details.
         */
        public boolean validateEmail(String email) {
            // Email annotation is referenced in JavaDoc but not used in code
            return email.contains("@");
        }
    }
}