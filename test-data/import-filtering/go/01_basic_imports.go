//go:build ignore

// Test Pattern 1: Basic Imports
// Tests standard package imports and usage detection

package main

import (
	"fmt"
	"os"
	"net/http"
	"time"
	"encoding/json"
	"strings"
	"io"
	"bytes"
)

// Not using: time, strings, io, bytes

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func main() {
	// Using fmt
	fmt.Println("Hello, Go!")
	
	// Using os
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting hostname: %v\n", err)
	}
	
	// Using net/http
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Using encoding/json
		resp := Response{
			Message: "Welcome to " + hostname,
			Status:  200,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	
	// Start server
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}