package performance

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/janreges/ai-distiller/internal/processor"
)

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	testFiles         []string
	processor         *PerformanceProcessor
	results           map[string]*BenchmarkResult
	configurations    []*PerformanceConfig
}

// BenchmarkResult contains detailed benchmark results
type BenchmarkResult struct {
	Configuration     *PerformanceConfig
	Mode             PerformanceMode
	TestFile         string
	FileSize         int64
	FileSizeMB       float64
	ProcessingTime   time.Duration
	ThroughputMBps   float64
	MemoryUsageMB    int64
	CacheHitRate     float64
	ErrorCount       int
	Success          bool
	Error            error
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(cacheDir string) *BenchmarkSuite {
	return &BenchmarkSuite{
		processor:      NewPerformanceProcessor(cacheDir),
		results:        make(map[string]*BenchmarkResult),
		configurations: generateBenchmarkConfigurations(),
	}
}

// AddTestFile adds a test file to the benchmark suite
func (b *BenchmarkSuite) AddTestFile(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("test file does not exist: %s", filePath)
	}
	b.testFiles = append(b.testFiles, filePath)
	return nil
}

// AddTestDirectory adds all processable files from a directory
func (b *BenchmarkSuite) AddTestDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file can be processed
		proc := processor.New()
		if proc.CanProcess(path) {
			b.testFiles = append(b.testFiles, path)
		}

		return nil
	})
}

// RunAllBenchmarks runs comprehensive performance benchmarks
func (b *BenchmarkSuite) RunAllBenchmarks(ctx context.Context) error {
	if len(b.testFiles) == 0 {
		return fmt.Errorf("no test files added to benchmark suite")
	}

	fmt.Println("=== AI Distiller Performance Benchmark Suite ===")
	fmt.Printf("Test files: %d\n", len(b.testFiles))
	fmt.Printf("Configurations: %d\n", len(b.configurations))
	fmt.Printf("Total benchmarks: %d\n\n", len(b.testFiles)*len(b.configurations))

	// Run benchmarks for each configuration
	for i, config := range b.configurations {
		fmt.Printf("Running configuration %d/%d...\n", i+1, len(b.configurations))
		
		// Update processor configuration
		b.processor.WithConfig(config)
		
		// Test each file
		for _, testFile := range b.testFiles {
			result := b.runSingleBenchmark(ctx, testFile, config)
			key := b.generateResultKey(testFile, config, i)
			b.results[key] = result
			
			// Print progress
			if result.Success {
				fmt.Printf("  ✅ %s: %.2f MB/s\n", filepath.Base(testFile), result.ThroughputMBps)
			} else {
				fmt.Printf("  ❌ %s: %v\n", filepath.Base(testFile), result.Error)
			}
		}
		fmt.Println()
	}

	return nil
}

// RunModeComparison compares different performance modes on the same files
func (b *BenchmarkSuite) RunModeComparison(ctx context.Context) error {
	if len(b.testFiles) == 0 {
		return fmt.Errorf("no test files added to benchmark suite")
	}

	modes := []PerformanceMode{ModeStandard, ModeCached, ModeConcurrent, ModeStreaming}
	modeNames := []string{"Standard", "Cached", "Concurrent", "Streaming"}

	fmt.Println("=== Performance Mode Comparison ===")
	fmt.Printf("Test files: %d\n", len(b.testFiles))
	fmt.Printf("Modes: %v\n\n", modeNames)

	config := DefaultPerformanceConfig()
	b.processor.WithConfig(config)

	// Test each mode
	for i, mode := range modes {
		fmt.Printf("Testing %s mode...\n", modeNames[i])
		
		for _, testFile := range b.testFiles {
			result := b.runModeBenchmark(ctx, testFile, mode, config)
			key := fmt.Sprintf("mode_%s_%s", modeNames[i], filepath.Base(testFile))
			b.results[key] = result
			
			if result.Success {
				fmt.Printf("  ✅ %s: %.2f MB/s (%.2f MB, %v)\n", 
					filepath.Base(testFile), result.ThroughputMBps, result.FileSizeMB, result.ProcessingTime)
			} else {
				fmt.Printf("  ❌ %s: %v\n", filepath.Base(testFile), result.Error)
			}
		}
		fmt.Println()
	}

	return nil
}

// runSingleBenchmark runs a benchmark for a single file and configuration
func (b *BenchmarkSuite) runSingleBenchmark(
	ctx context.Context,
	testFile string,
	config *PerformanceConfig,
) *BenchmarkResult {
	result := &BenchmarkResult{
		Configuration: config,
		TestFile:      testFile,
	}

	// Get file info
	info, err := os.Stat(testFile)
	if err != nil {
		result.Error = err
		return result
	}

	result.FileSize = info.Size()
	result.FileSizeMB = float64(info.Size()) / 1024 / 1024

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	runtime.GC() // Force garbage collection

	// Run benchmark
	startTime := time.Now()
	_, err = b.processor.ProcessFile(ctx, testFile, processor.ProcessOptions{})
	endTime := time.Now()

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	result.ProcessingTime = endTime.Sub(startTime)
	result.MemoryUsageMB = int64((memAfter.Alloc - memBefore.Alloc) / 1024 / 1024)
	
	if err != nil {
		result.Error = err
		result.ErrorCount = 1
		return result
	}

	// Calculate throughput
	if result.ProcessingTime > 0 {
		result.ThroughputMBps = result.FileSizeMB / result.ProcessingTime.Seconds()
	}

	// Get cache hit rate if applicable
	if config.CacheEnabled {
		if cacheStats := b.processor.cachedProcessor.GetCache().Stats(); cacheStats != nil {
			result.CacheHitRate = cacheStats.GetHitRate()
		}
	}

	result.Success = true
	return result
}

// runModeBenchmark runs a benchmark for a specific performance mode
func (b *BenchmarkSuite) runModeBenchmark(
	ctx context.Context,
	testFile string,
	mode PerformanceMode,
	config *PerformanceConfig,
) *BenchmarkResult {
	result := &BenchmarkResult{
		Configuration: config,
		Mode:         mode,
		TestFile:     testFile,
	}

	// Get file info
	info, err := os.Stat(testFile)
	if err != nil {
		result.Error = err
		return result
	}

	result.FileSize = info.Size()
	result.FileSizeMB = float64(info.Size()) / 1024 / 1024

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	runtime.GC()

	// Run benchmark with specific mode
	startTime := time.Now()
	_, err = b.processor.ProcessFileWithMode(ctx, testFile, processor.ProcessOptions{}, mode)
	endTime := time.Now()

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	result.ProcessingTime = endTime.Sub(startTime)
	result.MemoryUsageMB = int64((memAfter.Alloc - memBefore.Alloc) / 1024 / 1024)
	
	if err != nil {
		result.Error = err
		result.ErrorCount = 1
		return result
	}

	// Calculate throughput
	if result.ProcessingTime > 0 {
		result.ThroughputMBps = result.FileSizeMB / result.ProcessingTime.Seconds()
	}

	result.Success = true
	return result
}

// GetBestConfiguration returns the configuration with the best overall performance
func (b *BenchmarkSuite) GetBestConfiguration() (*PerformanceConfig, *BenchmarkResult, error) {
	if len(b.results) == 0 {
		return nil, nil, fmt.Errorf("no benchmark results available")
	}

	var bestConfig *PerformanceConfig
	var bestResult *BenchmarkResult
	bestScore := 0.0

	for _, result := range b.results {
		if !result.Success {
			continue
		}

		// Score based on throughput and memory efficiency
		// Higher throughput is better, lower memory usage is better
		score := result.ThroughputMBps / math.Max(float64(result.MemoryUsageMB), 1.0)
		
		if score > bestScore {
			bestScore = score
			bestConfig = result.Configuration
			bestResult = result
		}
	}

	if bestConfig == nil {
		return nil, nil, fmt.Errorf("no successful benchmark results found")
	}

	return bestConfig, bestResult, nil
}

// GenerateReport generates a comprehensive benchmark report
func (b *BenchmarkSuite) GenerateReport() string {
	if len(b.results) == 0 {
		return "No benchmark results available"
	}

	report := "=== AI Distiller Performance Benchmark Report ===\n\n"

	// Summary statistics
	successful := 0
	totalThroughput := 0.0
	var throughputs []float64

	for _, result := range b.results {
		if result.Success {
			successful++
			totalThroughput += result.ThroughputMBps
			throughputs = append(throughputs, result.ThroughputMBps)
		}
	}

	sort.Float64s(throughputs)

	report += fmt.Sprintf("Total benchmarks: %d\n", len(b.results))
	report += fmt.Sprintf("Successful: %d\n", successful)
	report += fmt.Sprintf("Failed: %d\n", len(b.results)-successful)
	
	if successful > 0 {
		avgThroughput := totalThroughput / float64(successful)
		medianThroughput := throughputs[len(throughputs)/2]
		maxThroughput := throughputs[len(throughputs)-1]
		
		report += fmt.Sprintf("\nThroughput Statistics:\n")
		report += fmt.Sprintf("  Average: %.2f MB/s\n", avgThroughput)
		report += fmt.Sprintf("  Median: %.2f MB/s\n", medianThroughput)
		report += fmt.Sprintf("  Maximum: %.2f MB/s\n", maxThroughput)
	}

	// Best configuration
	if bestConfig, bestResult, err := b.GetBestConfiguration(); err == nil {
		report += fmt.Sprintf("\nBest Configuration:\n")
		report += fmt.Sprintf("  File: %s\n", filepath.Base(bestResult.TestFile))
		report += fmt.Sprintf("  Throughput: %.2f MB/s\n", bestResult.ThroughputMBps)
		report += fmt.Sprintf("  Memory: %d MB\n", bestResult.MemoryUsageMB)
		report += fmt.Sprintf("  Workers: %d\n", bestConfig.MaxWorkers)
		report += fmt.Sprintf("  Chunk Size: %d\n", bestConfig.ChunkSize)
		report += fmt.Sprintf("  Cache Enabled: %v\n", bestConfig.CacheEnabled)
	}

	// Top 5 results
	var sortedResults []*BenchmarkResult
	for _, result := range b.results {
		if result.Success {
			sortedResults = append(sortedResults, result)
		}
	}

	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].ThroughputMBps > sortedResults[j].ThroughputMBps
	})

	report += fmt.Sprintf("\nTop 5 Results:\n")
	for i, result := range sortedResults {
		if i >= 5 {
			break
		}
		report += fmt.Sprintf("  %d. %s: %.2f MB/s (%.2f MB, %v)\n",
			i+1, filepath.Base(result.TestFile), result.ThroughputMBps,
			result.FileSizeMB, result.ProcessingTime)
	}

	return report
}

// generateResultKey creates a unique key for benchmark results
func (b *BenchmarkSuite) generateResultKey(testFile string, config *PerformanceConfig, configIndex int) string {
	return fmt.Sprintf("config_%d_%s", configIndex, filepath.Base(testFile))
}

// generateBenchmarkConfigurations creates various configurations to test
func generateBenchmarkConfigurations() []*PerformanceConfig {
	base := DefaultPerformanceConfig()
	
	configs := []*PerformanceConfig{
		// Default configuration
		base,
		
		// High concurrency
		{
			MaxWorkers:             16,
			BufferSize:             2048,
			ChunkSize:              base.ChunkSize,
			StreamBufferSize:       base.StreamBufferSize,
			MaxMemoryMB:            base.MaxMemoryMB,
			CacheEnabled:           true,
			CacheMaxSize:           base.CacheMaxSize,
			CacheMaxAge:            base.CacheMaxAge,
			LargeFileThresholdMB:   base.LargeFileThresholdMB,
			ManyFilesThreshold:     base.ManyFilesThreshold,
			EnableAutoOptimization: true,
		},
		
		// Low memory
		{
			MaxWorkers:             4,
			BufferSize:             512,
			ChunkSize:              500,
			StreamBufferSize:       32 * 1024,
			MaxMemoryMB:            256,
			CacheEnabled:           false,
			CacheMaxSize:           512 * 1024 * 1024,
			CacheMaxAge:            base.CacheMaxAge,
			LargeFileThresholdMB:   25,
			ManyFilesThreshold:     base.ManyFilesThreshold,
			EnableAutoOptimization: true,
		},
		
		// Large files optimized
		{
			MaxWorkers:             base.MaxWorkers,
			BufferSize:             base.BufferSize,
			ChunkSize:              2000,
			StreamBufferSize:       128 * 1024,
			MaxMemoryMB:            1024,
			CacheEnabled:           true,
			CacheMaxSize:           base.CacheMaxSize,
			CacheMaxAge:            base.CacheMaxAge,
			LargeFileThresholdMB:   10,
			ManyFilesThreshold:     base.ManyFilesThreshold,
			EnableAutoOptimization: true,
		},
		
		// Cache disabled
		{
			MaxWorkers:             base.MaxWorkers,
			BufferSize:             base.BufferSize,
			ChunkSize:              base.ChunkSize,
			StreamBufferSize:       base.StreamBufferSize,
			MaxMemoryMB:            base.MaxMemoryMB,
			CacheEnabled:           false,
			CacheMaxSize:           0,
			CacheMaxAge:            0,
			LargeFileThresholdMB:   base.LargeFileThresholdMB,
			ManyFilesThreshold:     base.ManyFilesThreshold,
			EnableAutoOptimization: false,
		},
	}
	
	return configs
}