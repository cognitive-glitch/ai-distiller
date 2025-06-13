// Level 1: Basic.java
package com.aidi.test.basic;

/**
 * A basic class to test fundamental parsing of methods,
 * variables, and control flow.
 */
public class Basic {

    private static final String GREETING_PREFIX = "Hello, ";

    /**
     * Standard Java entry point.
     * @param args Command line arguments (unused).
     */
    public static void main(String[] args) {
        String world = "World";
        int repetitions = 3;
        for (int i = 0; i < repetitions; i++) {
            String message = createGreeting(world, i + 1);
            System.out.println(message);
        }
    }

    /**
     * A private helper method to test visibility and method invocation.
     */
    private static String createGreeting(String name, int count) {
        // Tests basic string concatenation and expressions.
        if (name == null || name.isEmpty()) {
            return GREETING_PREFIX + "Anonymous!";
        }
        return GREETING_PREFIX + name + "! (call #" + count + ")";
    }
}