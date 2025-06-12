.PHONY: test test-unit test-integration test-aid test-all clean

# Run all tests
test-all: test-unit test-integration test-aid

# Run unit tests
test-unit:
	@echo "=== Running Unit Tests ==="
	go test -v distiller_test.go

# Run integration tests for ai-distiller CLI
test-integration: build
	@echo "=== Running AI-Distiller Integration Tests ==="
	go test -v integration_test.go

# Run integration tests for aid CLI
test-aid: build-aid
	@echo "=== Running AID Integration Tests ==="
	go test -v integration_cli_test.go

# Build the CLI for integration tests
build:
	@echo "=== Building AI Distiller ==="
	go build -o ../ai-distiller ../cmd/ai-distiller/main.go

# Build the aid CLI
build-aid:
	@echo "=== Building AID CLI ==="
	go build -o ../aid ../cmd/aid/main.go || echo "AID build skipped (may not be implemented yet)"

# Run format validation
validate:
	@echo "=== Running Format Validation ==="
	go run format_validator.go

# Run test runner for all scenarios
run-scenarios:
	@echo "=== Running Test Scenarios ==="
	go run test_runner.go

# Clean generated files
clean:
	@echo "=== Cleaning Test Outputs ==="
	rm -rf actual/*
	rm -f ../ai-distiller
	rm -f validation_report.md
	rm -f quality_analysis.md

# Quick test - just the basic functionality
test-quick:
	@echo "=== Quick Test ==="
	@echo "Testing basic class parsing..."
	@go test -v -run TestDistiller/full_output_basic_class distiller_test.go
	@echo ""
	@echo "Testing private member filtering..."
	@go test -v -run TestDistiller/no_private_basic_class distiller_test.go

# Benchmark
benchmark:
	@echo "=== Running Benchmarks ==="
	go test -bench=. -benchtime=10s integration_test.go