package main

import (
	"fmt"
	"time"
)

// PUBLIC_CONSTANT is a public constant
const PUBLIC_CONSTANT = "hello"

// privateConstant is a private constant
const privateConstant = 42

// User represents a simple user model
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	age  int    // private field
}

// GetName returns the user's name (public method)
func (u *User) GetName() string {
	return u.Name
}

// setAge sets the user's age (private method)
func (u *User) setAge(newAge int) {
	u.age = newAge
}

// PublicFunction is a standalone public function
func PublicFunction() {
	fmt.Println("Hello from public function")
}

// privateFunction is a standalone private function
func privateFunction() {
	fmt.Println("Hello from private function")
}