// 01_basic.rs
// A test case for basic Rust constructs.

/// The main entry point of the application.
fn main() {
    println!("Starting basic application...");
    let app_name = settings::APP_NAME;
    let version = settings::get_version();
    println!("Running {} v{}", app_name, version);
}

/// A module to handle application settings.
/// This tests the parser's ability to handle inline module definitions.
pub mod settings {
    /// The public name of the application.
    pub const APP_NAME: &str = "AI Distiller";

    // A private constant, not visible outside this module.
    const MAJOR_VERSION: u8 = 1;
    const MINOR_VERSION: u8 = 0;

    /// Returns the full version string.
    /// This function is public and can be called from outside `settings`.
    pub fn get_version() -> String {
        // This function demonstrates internal logic and calls a private function.
        if is_stable() {
            format!("{}.{}", MAJOR_VERSION, MINOR_VERSION)
        } else {
            format!("{}.{}-beta", MAJOR_VERSION, MINOR_VERSION)
        }
    }

    // A private function, only callable within the `settings` module.
    // The parser should correctly identify its scope and visibility.
    fn is_stable() -> bool {
        // A simple check. In a real app, this might be a compile-time flag.
        true
    }

    /// Internal utility for debug builds
    pub(crate) fn debug_info() -> String {
        format!("Debug info: v{}.{}", MAJOR_VERSION, MINOR_VERSION)
    }

    /// Private helper for version validation
    fn validate_version() -> bool {
        MAJOR_VERSION > 0 && MINOR_VERSION >= 0
    }
}