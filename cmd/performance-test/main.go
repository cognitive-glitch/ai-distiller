package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/janreges/ai-distiller/internal/performance"
)

func main() {
	// Command line flags
	var (
		testDir     = flag.String("dir", "./test-data", "Directory containing test files")
		cacheDir    = flag.String("cache", "./cache", "Cache directory")
		outputFile  = flag.String("output", "", "Output file for benchmark report (optional)")
		runMode     = flag.String("mode", "all", "Benchmark mode: all, comparison, config")
		timeoutSec  = flag.Int("timeout", 300, "Timeout in seconds")
		verbose     = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	fmt.Println("=== AI Distiller Performance Testing Tool ===")
	fmt.Printf("Test directory: %s\n", *testDir)
	fmt.Printf("Cache directory: %s\n", *cacheDir)
	fmt.Printf("Mode: %s\n", *runMode)
	fmt.Printf("Timeout: %d seconds\n", *timeoutSec)
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSec)*time.Second)
	defer cancel()

	// Create cache directory
	if err := os.MkdirAll(*cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cache directory: %v\n", err)
		os.Exit(1)
	}

	// Create benchmark suite
	suite := performance.NewBenchmarkSuite(*cacheDir)

	// Add test files
	if err := addTestFiles(suite, *testDir, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add test files: %v\n", err)
		os.Exit(1)
	}

	// Run benchmarks based on mode
	switch *runMode {
	case "all":
		err := runAllBenchmarks(ctx, suite)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
			os.Exit(1)
		}

	case "comparison":
		err := runModeComparison(ctx, suite)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Mode comparison failed: %v\n", err)
			os.Exit(1)
		}

	case "config":
		err := runConfigOptimization(ctx, suite)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config optimization failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *runMode)
		os.Exit(1)
	}

	// Generate and display report
	report := suite.GenerateReport()
	fmt.Println(report)

	// Save report to file if specified
	if *outputFile != "" {
		if err := saveReport(report, *outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save report: %v\n", err)
		} else {
			fmt.Printf("\nReport saved to: %s\n", *outputFile)
		}
	}

	// Show recommendations
	showRecommendations(suite)
}

// addTestFiles discovers and adds test files to the benchmark suite
func addTestFiles(suite *performance.BenchmarkSuite, testDir string, verbose bool) error {
	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			"./test-data/functional-tests",
			"../test-data/functional-tests",
			"../../test-data/functional-tests",
		}
		
		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				testDir = alt
				found = true
				break
			}
		}
		
		if !found {
			return fmt.Errorf("test directory not found: %s", testDir)
		}
	}

	if verbose {
		fmt.Printf("Discovering test files in: %s\n", testDir)
	}

	// Add all files from test directory
	err := suite.AddTestDirectory(testDir)
	if err != nil {
		return fmt.Errorf("failed to add test directory: %w", err)
	}

	// Also add specific test files if they exist
	specificFiles := []string{
		"test_java_complex.java",
		"test_typescript_complex.ts", 
		"test_python_complex.py",
		"test_javascript_complex.js",
		"test_csharp_complex.cs",
		"test_rust_complex.rs",
	}

	for _, filename := range specificFiles {
		fullPath := filepath.Join(testDir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			_ = suite.AddTestFile(fullPath)
			if verbose {
				fmt.Printf("  Added: %s\n", filename)
			}
		}
	}

	if verbose {
		fmt.Println()
	}

	return nil
}

// runAllBenchmarks runs comprehensive benchmarks with all configurations
func runAllBenchmarks(ctx context.Context, suite *performance.BenchmarkSuite) error {
	fmt.Println("Running comprehensive performance benchmarks...")
	fmt.Println("This will test multiple configurations and may take several minutes.")
	fmt.Println()

	return suite.RunAllBenchmarks(ctx)
}

// runModeComparison compares different performance modes
func runModeComparison(ctx context.Context, suite *performance.BenchmarkSuite) error {
	fmt.Println("Running performance mode comparison...")
	fmt.Println("This will compare Standard, Cached, Concurrent, and Streaming modes.")
	fmt.Println()

	return suite.RunModeComparison(ctx)
}

// runConfigOptimization finds the optimal configuration
func runConfigOptimization(ctx context.Context, suite *performance.BenchmarkSuite) error {
	fmt.Println("Running configuration optimization...")
	fmt.Println("This will find the best configuration for your system.")
	fmt.Println()

	// Run all benchmarks first
	err := suite.RunAllBenchmarks(ctx)
	if err != nil {
		return err
	}

	// Find best configuration
	bestConfig, bestResult, err := suite.GetBestConfiguration()
	if err != nil {
		return fmt.Errorf("failed to find best configuration: %w", err)
	}

	fmt.Println("\n=== Optimal Configuration Found ===")
	fmt.Printf("Best performance on: %s\n", filepath.Base(bestResult.TestFile))
	fmt.Printf("Throughput: %.2f MB/s\n", bestResult.ThroughputMBps)
	fmt.Printf("Memory usage: %d MB\n", bestResult.MemoryUsageMB)
	fmt.Printf("Processing time: %v\n", bestResult.ProcessingTime)
	fmt.Println("\nConfiguration details:")
	fmt.Printf("  Max Workers: %d\n", bestConfig.MaxWorkers)
	fmt.Printf("  Buffer Size: %d\n", bestConfig.BufferSize)
	fmt.Printf("  Chunk Size: %d\n", bestConfig.ChunkSize)
	fmt.Printf("  Stream Buffer: %d KB\n", bestConfig.StreamBufferSize/1024)
	fmt.Printf("  Max Memory: %d MB\n", bestConfig.MaxMemoryMB)
	fmt.Printf("  Cache Enabled: %v\n", bestConfig.CacheEnabled)
	if bestConfig.CacheEnabled {
		fmt.Printf("  Cache Max Size: %d MB\n", bestConfig.CacheMaxSize/1024/1024)
		fmt.Printf("  Cache Max Age: %v\n", bestConfig.CacheMaxAge)
	}
	fmt.Printf("  Large File Threshold: %d MB\n", bestConfig.LargeFileThresholdMB)
	fmt.Printf("  Many Files Threshold: %d\n", bestConfig.ManyFilesThreshold)

	return nil
}

// saveReport saves the benchmark report to a file
func saveReport(report, filename string) error {
	// Add timestamp to report
	timestampedReport := fmt.Sprintf("Generated: %s\n\n%s", time.Now().Format(time.RFC3339), report)
	
	return os.WriteFile(filename, []byte(timestampedReport), 0644)
}

// showRecommendations provides performance recommendations
func showRecommendations(suite *performance.BenchmarkSuite) {
	fmt.Println("\n=== Performance Recommendations ===")

	bestConfig, bestResult, err := suite.GetBestConfiguration()
	if err != nil {
		fmt.Println("Unable to generate recommendations: no successful benchmarks")
		return
	}

	fmt.Println("\nBased on your system and test files, we recommend:")
	fmt.Println()

	// Worker count recommendation
	if bestConfig.MaxWorkers <= 4 {
		fmt.Println("🔧 WORKERS: Use 4-8 workers for optimal balance of concurrency and resource usage")
	} else if bestConfig.MaxWorkers >= 16 {
		fmt.Println("🔧 WORKERS: Your system benefits from high concurrency (16+ workers)")
	} else {
		fmt.Printf("🔧 WORKERS: Use %d workers as determined by benchmarks\n", bestConfig.MaxWorkers)
	}

	// Memory recommendation
	if bestResult.MemoryUsageMB > 500 {
		fmt.Println("💾 MEMORY: Consider reducing chunk size or enabling streaming for large files")
	} else if bestResult.MemoryUsageMB < 100 {
		fmt.Println("💾 MEMORY: You can increase chunk sizes for better performance")
	}

	// Cache recommendation
	if bestConfig.CacheEnabled && bestResult.CacheHitRate > 50 {
		fmt.Printf("🚀 CACHE: Cache is effective (%.1f%% hit rate) - keep it enabled\n", bestResult.CacheHitRate)
	} else if !bestConfig.CacheEnabled {
		fmt.Println("🚀 CACHE: Consider enabling cache for repeated processing of the same files")
	}

	// Throughput recommendation
	if bestResult.ThroughputMBps > 100 {
		fmt.Printf("⚡ PERFORMANCE: Excellent throughput (%.1f MB/s) - current settings are optimal\n", bestResult.ThroughputMBps)
	} else if bestResult.ThroughputMBps < 10 {
		fmt.Printf("⚠️  PERFORMANCE: Low throughput (%.1f MB/s) - consider optimizing file sizes or system resources\n", bestResult.ThroughputMBps)
	} else {
		fmt.Printf("✅ PERFORMANCE: Good throughput (%.1f MB/s)\n", bestResult.ThroughputMBps)
	}

	// File size recommendations
	if bestConfig.LargeFileThresholdMB < 50 {
		fmt.Println("📁 FILES: Use streaming mode for files larger than 50MB")
	} else {
		fmt.Printf("📁 FILES: Current large file threshold (%dMB) works well for your use case\n", bestConfig.LargeFileThresholdMB)
	}

	fmt.Println("\nTo apply these settings, update your PerformanceConfig accordingly.")
}