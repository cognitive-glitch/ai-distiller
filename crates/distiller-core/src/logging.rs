//! Unified logging configuration for CLI and MCP server

use env_logger::Builder;
use std::io::Write;

/// Initialize logging with the given verbosity level
///
/// # Arguments
/// * `verbosity` - Verbosity level (0=error, 1=warn, 2=info, 3=debug, 4+=trace)
pub fn init_logging(verbosity: u8) {
    let level = match verbosity {
        0 => "error",
        1 => "warn",
        2 => "info",
        3 => "debug",
        _ => "trace",
    };

    Builder::from_env(env_logger::Env::default().default_filter_or(level))
        .format(|buf, record| writeln!(buf, "[{}] {}", record.level(), record.args()))
        .init();
}

/// Initialize logging from environment variable or default
pub fn init_logging_from_env(default: &str) {
    Builder::from_env(env_logger::Env::default().default_filter_or(default))
        .format(|buf, record| writeln!(buf, "[{}] {}", record.level(), record.args()))
        .init();
}
