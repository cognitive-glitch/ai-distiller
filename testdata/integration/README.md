# Integration Test Scenarios

This directory contains multi-file, multi-language test scenarios for comprehensive integration testing.

## Test Scenarios

### 1. Mixed Language Project (`mixed/`)
Tests multi-language processing in a single directory:
- Python + TypeScript + Go files
- Tests language detection
- Tests parallel processing
- Tests output ordering

### 2. Option Combinations (`options/`)
Tests different filtering options:
- Default (public only)
- With implementations
- With private/protected
- Various combinations

### 3. Error Handling (`errors/`)
Tests error scenarios:
- Malformed code
- Unsupported files
- Permission issues
- Edge cases

### 4. Real-World Patterns (`real-world/`)
Tests common project structures:
- MVC pattern
- Microservices
- Monorepo
- Library structure

### 5. Performance (`performance/`)
Tests with large codebases:
- Many files (100+)
- Large files (10k+ lines)
- Deep nesting
- Complex dependencies
