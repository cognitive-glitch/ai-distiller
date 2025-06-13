// Basic Rust example for testing

use std::collections::HashMap;
use std::io::{self, Read};

/// A simple struct demonstrating Rust features
#[derive(Debug, Clone)]
pub struct Person {
    pub name: String,
    age: u32,
    email: Option<String>,
}

impl Person {
    /// Creates a new Person
    pub fn new(name: String, age: u32) -> Self {
        Person {
            name,
            age,
            email: None,
        }
    }

    /// Gets the person's age
    pub fn get_age(&self) -> u32 {
        self.age
    }

    /// Sets the email address
    pub fn set_email(&mut self, email: String) {
        self.email = Some(email);
    }

    // Private helper method
    fn validate_email(email: &str) -> bool {
        email.contains('@')
    }
}

/// Trait for displaying items
pub trait Display {
    fn display(&self) -> String;
}

impl Display for Person {
    fn display(&self) -> String {
        format!("{} ({} years old)", self.name, self.age)
    }
}

/// Example enum
#[derive(Debug)]
pub enum Status {
    Active,
    Inactive,
    Pending(String),
    Error { code: u32, message: String },
}

/// Generic function
pub fn process_items<T: Display>(items: Vec<T>) -> Vec<String> {
    items.iter().map(|item| item.display()).collect()
}

/// Async function example
pub async fn fetch_data(url: &str) -> Result<String, io::Error> {
    // Simulated async operation
    Ok(format!("Data from {}", url))
}

// Module example
pub mod utils {
    /// Helper function in module
    pub fn format_name(first: &str, last: &str) -> String {
        format!("{} {}", first, last)
    }
}

// Type alias
type UserId = u64;

// Constants
const MAX_RETRIES: u32 = 3;
pub const API_VERSION: &str = "1.0";

// Static variable
static mut COUNTER: u32 = 0;

// Unit tests
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_person_creation() {
        let person = Person::new("Alice".to_string(), 30);
        assert_eq!(person.get_age(), 30);
    }
}