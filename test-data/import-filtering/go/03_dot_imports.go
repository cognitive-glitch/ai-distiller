// Test Pattern 3: Dot Imports and Grouped Imports
// Tests dot imports that bring package contents into current namespace

package main

import (
	. "fmt"
	"io/ioutil"
	"log"
	. "strings"
	"regexp"
	"unicode"
	"sort"
)

// Not using: io/ioutil, regexp, unicode, sort

func main() {
	// Using Println from fmt (dot import)
	Println("Hello from dot import!")
	
	// Using log normally
	log.Println("Regular import still works")
	
	// Using functions from strings (dot import)
	text := "Hello, World!"
	lower := ToLower(text)       // strings.ToLower
	upper := ToUpper(text)       // strings.ToUpper
	parts := Split(text, ", ")   // strings.Split
	joined := Join(parts, " - ") // strings.Join
	
	Printf("Original: %s\n", text)
	Printf("Lower: %s\n", lower)
	Printf("Upper: %s\n", upper)
	Printf("Joined: %s\n", joined)
	
	// Using HasPrefix from strings (dot import)
	if HasPrefix(text, "Hello") {
		Println("Text starts with Hello")
	}
	
	// More string functions from dot import
	trimmed := TrimSpace("  spaced  ")
	replaced := Replace(text, "World", "Go", 1)
	
	Println("Trimmed:", trimmed)
	Println("Replaced:", replaced)
}