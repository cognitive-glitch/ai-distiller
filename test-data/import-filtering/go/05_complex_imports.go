//go:build ignore

// Test Pattern 5: Complex Import Patterns
// Tests struct tags, CGO, and indirect usage

package main

/*
#include <stdlib.h>
#include <string.h>

char* get_version() {
    return "1.0.0";
}
*/
import "C"
import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"time"
	"unsafe"
)

// Not using: compress/gzip, errors, path/filepath

// Struct with tags that reference imported packages
type User struct {
	ID        int       `json:"id" db:"user_id"`
	Name      string    `json:"name" db:"user_name"`
	Email     string    `json:"email" db:"user_email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Using imports in struct tags and methods
type Config struct {
	AppName    string        `json:"app_name"`
	Version    string        `json:"version"`
	Timeout    time.Duration `json:"timeout"`
	BufferSize int           `json:"buffer_size"`
}

func main() {
	// Using C import (CGO)
	version := C.GoString(C.get_version())
	fmt.Printf("Version from C: %s\n", version)

	// Using unsafe (often used with CGO)
	ptr := unsafe.Pointer(&version)
	fmt.Printf("Pointer address: %v\n", ptr)

	// Using json with struct tags
	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Using bytes.Buffer
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(user); err != nil {
		log.Printf("Encoding error: %v", err)
	}

	fmt.Println("JSON output:")
	fmt.Println(buf.String())

	// Using reflect to inspect struct tags
	t := reflect.TypeOf(user)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		dbTag := field.Tag.Get("db")
		fmt.Printf("Field %s: json=%q db=%q\n", field.Name, jsonTag, dbTag)
	}

	// Using time in struct and directly
	fmt.Printf("Created at: %s\n", user.CreatedAt.Format(time.RFC3339))

	// The json import is used because of struct tags
	// The time import is used in struct fields and directly
	// The C import must be kept for CGO
}