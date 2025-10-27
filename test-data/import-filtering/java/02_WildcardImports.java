// Test Pattern 2: Wildcard Imports and Specific Class Imports
// Tests import package.* alongside specific imports

package com.example.importtest;

import java.util.*;  // Will use ArrayList, HashSet, Collections, Stream
import java.io.IOException;
import java.io.BufferedReader;
import java.nio.file.*;  // Will use Files, Paths, Path
import java.net.URI;
import java.time.Duration;
import java.util.concurrent.*;  // Will use ExecutorService, Executors

// Not using: BufferedReader, URI, Duration

public class WildcardImports {

    public void demonstrateCollections() {
        // Using from java.util.*
        ArrayList<String> list = new ArrayList<>();
        HashSet<Integer> set = new HashSet<>();
        set.add(1);
        set.add(2);

        // Using Collections from java.util.*
        Collections.sort(list);

        // Using Stream API from java.util.*
        List<Integer> numbers = Arrays.asList(1, 2, 3, 4, 5);
        int sum = numbers.stream()
            .filter(n -> n > 2)
            .mapToInt(Integer::intValue)
            .sum();
        System.out.println("Sum: " + sum);
    }

    public void demonstrateNIO() throws IOException {
        // Using from java.nio.file.*
        Path path = Paths.get("test.txt");

        if (Files.exists(path)) {
            List<String> lines = Files.readAllLines(path);
            System.out.println("Lines: " + lines.size());
        }

        // Using Files for directory operations
        Files.list(Paths.get("."))
            .filter(Files::isRegularFile)
            .forEach(System.out::println);
    }

    public void demonstrateConcurrency() {
        // Using from java.util.concurrent.*
        ExecutorService executor = Executors.newFixedThreadPool(4);

        // Submit some tasks
        for (int i = 0; i < 10; i++) {
            final int taskId = i;
            executor.submit(() -> {
                System.out.println("Task " + taskId + " running");
            });
        }

        executor.shutdown();
    }
}