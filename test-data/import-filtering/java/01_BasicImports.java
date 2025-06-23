// Test Pattern 1: Basic Class and Package Imports
// Tests standard imports for classes and their usage

package com.example.importtest;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedList;
import java.io.File;
import java.io.IOException;
import java.net.URL;
import java.net.URI;
import java.time.LocalDateTime;
import java.time.ZonedDateTime;

// Not using: LinkedList, URL, URI, ZonedDateTime

public class BasicImports {
    
    public void processData() throws IOException {
        // Using ArrayList
        ArrayList<String> list = new ArrayList<>();
        list.add("item1");
        list.add("item2");
        System.out.println(list);
        
        // Using HashMap
        HashMap<String, Integer> map = new HashMap<>();
        map.put("count", 42);
        
        // Using File and IOException
        File f = new File("test.txt");
        if (!f.exists()) {
            throw new IOException("File not found: " + f.getAbsolutePath());
        }
        
        // Using LocalDateTime
        LocalDateTime now = LocalDateTime.now();
        System.out.println("Current time: " + now);
    }
    
    public static void main(String[] args) {
        try {
            new BasicImports().processData();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}