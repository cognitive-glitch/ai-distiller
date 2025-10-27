// 04_complex.rs
// A test for macros, async, unsafe FFI, and advanced ownership.
use std::sync::{Arc, Mutex};
use std::ffi::{c_char, CStr};
use std::future::Future;
use std::pin::Pin;

// FFI: Declaring an external function from a C library.
// The parser must handle `extern "C"` blocks.
extern "C" {
    fn validate_syntax_natively(input: *const c_char) -> i32;

    /// Additional FFI function for complex validation
    fn complex_validation(
        input: *const c_char,
        length: usize,
        callback: extern "C" fn(i32)
    ) -> i32;
}

/// A simple declarative macro for creating a new, validated config.
/// The parser must handle the unique syntax of `macro_rules!`.
macro_rules! new_config {
    ($name:expr, $version:expr) => {{
        let config = Config {
            name: $name.to_string(),
            version: $version,
            is_validated: false,
        };
        // In a real scenario, more complex logic would be here.
        println!("Created config via macro: {}", config.name);
        config
    }};

    // Multiple macro patterns
    ($name:expr) => {
        new_config!($name, 1)
    };
}

/// Macro for generating validation functions
macro_rules! generate_validator {
    ($fn_name:ident, $error_msg:expr) => {
        pub fn $fn_name(input: &str) -> Result<(), &'static str> {
            if input.is_empty() {
                Err($error_msg)
            } else {
                Ok(())
            }
        }
    };
}

// Using the macro to generate functions
generate_validator!(validate_name, "Name cannot be empty");
generate_validator!(validate_version, "Version cannot be empty");

#[derive(Clone)]
pub struct Config {
    name: String,
    version: u32,
    is_validated: bool,
}

/// A service that uses a shared, mutable cache.
/// This tests `Arc<Mutex<T>>`, a very common concurrent pattern.
pub struct AnalysisService {
    cache: Arc<Mutex<Vec<String>>>,
    async_processor: Option<Pin<Box<dyn Future<Output = String> + Send>>>,
}

impl AnalysisService {
    /// Create new analysis service
    pub fn new() -> Self {
        Self {
            cache: Arc::new(Mutex::new(Vec::new())),
            async_processor: None,
        }
    }

    /// Asynchronously validates a piece of code using the native FFI function.
    /// This tests `async fn` syntax and `unsafe` blocks.
    pub async fn validate_code(&self, code: &str) -> Result<bool, &'static str> {
        println!("Starting async validation...");
        tokio::time::sleep(std::time::Duration::from_millis(10)).await;

        let c_str = std::ffi::CString::new(code).map_err(|_| "Invalid CString")?;
        let result_code: i32;

        // The `unsafe` block is a critical syntactic construct to parse.
        unsafe {
            result_code = validate_syntax_natively(c_str.as_ptr());
        }

        let is_valid = result_code == 0;
        if is_valid {
            // Accessing shared state requires locking the mutex.
            let mut cache_guard = self.cache.lock().unwrap();
            cache_guard.push(format!("Validated: {}", self.name_from_code(code)));
        }
        Ok(is_valid)
    }

    // A private helper method.
    fn name_from_code(&self, code: &str) -> String {
        code.lines().next().unwrap_or("unknown").to_string()
    }

    /// Internal unsafe method for advanced operations
    pub(crate) unsafe fn direct_memory_access(&self, ptr: *mut u8, len: usize) -> Option<String> {
        if ptr.is_null() || len == 0 {
            return None;
        }

        let slice = std::slice::from_raw_parts(ptr, len);
        String::from_utf8(slice.to_vec()).ok()
    }

    /// Private async method
    async fn process_cache(&self) -> usize {
        let guard = self.cache.lock().unwrap();
        guard.len()
    }
}

/// Advanced trait with async methods
pub trait AsyncProcessor {
    type Item;
    type Error;

    async fn process_async(&self, item: Self::Item) -> Result<String, Self::Error>;

    /// Default async implementation
    async fn batch_process(&self, items: Vec<Self::Item>) -> Vec<Result<String, Self::Error>> {
        let mut results = Vec::new();
        for item in items {
            results.push(self.process_async(item).await);
        }
        results
    }
}

/// Implementation for the analysis service
impl AsyncProcessor for AnalysisService {
    type Item = String;
    type Error = &'static str;

    async fn process_async(&self, item: Self::Item) -> Result<String, Self::Error> {
        self.validate_code(&item).await?;
        Ok(format!("Processed: {}", item))
    }
}

/// Union type for advanced FFI
#[repr(C)]
pub union FFIData {
    integer: i64,
    floating: f64,
    bytes: [u8; 8],
}

impl FFIData {
    /// Safe constructor
    pub fn new_integer(value: i64) -> Self {
        Self { integer: value }
    }

    /// Unsafe getter
    pub unsafe fn get_integer(&self) -> i64 {
        self.integer
    }

    /// Private unsafe method
    unsafe fn get_bytes(&self) -> &[u8; 8] {
        &self.bytes
    }
}

// Note: This requires a tokio runtime to execute.
// e.g., `#[tokio::main]`
fn main() {
    let _config = new_config!("My Project", 1);
    let _simple_config = new_config!("Simple");

    // Test generated validators
    if let Err(e) = validate_name("") {
        println!("Validation error: {}", e);
    }

    println!("Complex constructs defined.");
}