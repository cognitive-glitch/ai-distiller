package performance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// ConcurrentProcessor provides high-performance concurrent file processing
type ConcurrentProcessor struct {
	maxWorkers    int
	bufferSize    int
	enableMetrics bool
	metrics       *ProcessingMetrics
}

// ProcessingMetrics tracks performance statistics
type ProcessingMetrics struct {
	FilesProcessed    int64
	TotalBytes        int64
	ProcessingTime    time.Duration
	AverageFileTime   time.Duration
	ConcurrencyLevel  int
	MemoryPeakMB      int64
	ErrorCount        int64
	CacheHits         int64
	CacheMisses       int64
	mutex             sync.RWMutex
}

// FileTask represents a file processing task
type FileTask struct {
	Path     string
	Info     os.FileInfo
	Options  processor.ProcessOptions
	Result   chan FileResult
}

// FileResult represents the result of processing a file
type FileResult struct {
	File  *ir.DistilledFile
	Error error
	Path  string
	Size  int64
	Time  time.Duration
}

// BatchResult represents results from processing multiple files
type BatchResult struct {
	Files   []*ir.DistilledFile
	Errors  []error
	Metrics *ProcessingMetrics
}

// NewConcurrentProcessor creates a new high-performance processor
func NewConcurrentProcessor() *ConcurrentProcessor {
	return &ConcurrentProcessor{
		maxWorkers:    runtime.NumCPU() * 2, // Optimal for I/O bound tasks
		bufferSize:    1024,                 // Buffer for file tasks
		enableMetrics: true,
		metrics:       &ProcessingMetrics{},
	}
}

// WithWorkers sets the number of concurrent workers
func (p *ConcurrentProcessor) WithWorkers(workers int) *ConcurrentProcessor {
	if workers > 0 {
		p.maxWorkers = workers
	}
	return p
}

// WithBufferSize sets the task buffer size
func (p *ConcurrentProcessor) WithBufferSize(size int) *ConcurrentProcessor {
	if size > 0 {
		p.bufferSize = size
	}
	return p
}

// WithMetrics enables or disables performance metrics
func (p *ConcurrentProcessor) WithMetrics(enabled bool) *ConcurrentProcessor {
	p.enableMetrics = enabled
	return p
}

// ProcessDirectory processes all files in a directory concurrently
func (p *ConcurrentProcessor) ProcessDirectory(
	ctx context.Context,
	dirPath string,
	opts processor.ProcessOptions,
) (*BatchResult, error) {
	startTime := time.Now()

	// Find all processable files
	files, err := p.findProcessableFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return &BatchResult{
			Files:   []*ir.DistilledFile{},
			Errors:  []error{},
			Metrics: p.metrics,
		}, nil
	}

	// Process files concurrently
	results, errors := p.processFilesConcurrently(ctx, files, opts)

	// Update metrics
	if p.enableMetrics {
		p.updateMetrics(len(files), time.Since(startTime), errors)
	}

	return &BatchResult{
		Files:   results,
		Errors:  errors,
		Metrics: p.metrics,
	}, nil
}

// ProcessFiles processes a list of files concurrently
func (p *ConcurrentProcessor) ProcessFiles(
	ctx context.Context,
	filePaths []string,
	opts processor.ProcessOptions,
) (*BatchResult, error) {
	startTime := time.Now()

	// Filter for processable files
	var files []string
	for _, path := range filePaths {
		if processor := processor.New(); processor.CanProcess(path) {
			files = append(files, path)
		}
	}

	if len(files) == 0 {
		return &BatchResult{
			Files:   []*ir.DistilledFile{},
			Errors:  []error{},
			Metrics: p.metrics,
		}, nil
	}

	// Process files concurrently
	results, errors := p.processFilesConcurrently(ctx, files, opts)

	// Update metrics
	if p.enableMetrics {
		p.updateMetrics(len(files), time.Since(startTime), errors)
	}

	return &BatchResult{
		Files:   results,
		Errors:  errors,
		Metrics: p.metrics,
	}, nil
}

// findProcessableFiles finds all files that can be processed
func (p *ConcurrentProcessor) findProcessableFiles(dirPath string) ([]string, error) {
	var files []string
	proc := processor.New()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip files that are too large (> 100MB)
		if info.Size() > 100*1024*1024 {
			return nil
		}

		// Check if we can process this file
		if proc.CanProcess(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// processFilesConcurrently processes files using worker pool pattern
func (p *ConcurrentProcessor) processFilesConcurrently(
	ctx context.Context,
	files []string,
	opts processor.ProcessOptions,
) ([]*ir.DistilledFile, []error) {
	// Create channels
	tasks := make(chan string, p.bufferSize)
	results := make(chan FileResult, len(files))

	// Start workers
	var wg sync.WaitGroup
	workers := p.maxWorkers
	if len(files) < workers {
		workers = len(files)
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go p.worker(ctx, tasks, results, opts, &wg)
	}

	// Send tasks
	go func() {
		defer close(tasks)
		for _, file := range files {
			select {
			case tasks <- file:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var processedFiles []*ir.DistilledFile
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("failed to process %s: %w", result.Path, result.Error))
		} else {
			processedFiles = append(processedFiles, result.File)
		}

		// Update per-file metrics
		if p.enableMetrics {
			p.updateFileMetrics(result)
		}
	}

	return processedFiles, errors
}

// worker processes files from the task channel
func (p *ConcurrentProcessor) worker(
	ctx context.Context,
	tasks <-chan string,
	results chan<- FileResult,
	opts processor.ProcessOptions,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	proc := processor.New()

	for {
		select {
		case filePath, ok := <-tasks:
			if !ok {
				return
			}

			// Process single file
			result := p.processFile(ctx, proc, filePath, opts)

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// processFile processes a single file with timing
func (p *ConcurrentProcessor) processFile(
	ctx context.Context,
	proc *processor.Processor,
	filePath string,
	opts processor.ProcessOptions,
) FileResult {
	startTime := time.Now()

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return FileResult{
			Error: err,
			Path:  filePath,
			Time:  time.Since(startTime),
		}
	}

	// Process file
	file, err := proc.ProcessFile(filePath, opts)
	processingTime := time.Since(startTime)

	return FileResult{
		File:  file,
		Error: err,
		Path:  filePath,
		Size:  info.Size(),
		Time:  processingTime,
	}
}

// updateMetrics updates processing metrics
func (p *ConcurrentProcessor) updateMetrics(
	fileCount int,
	totalTime time.Duration,
	errors []error,
) {
	p.metrics.mutex.Lock()
	defer p.metrics.mutex.Unlock()

	p.metrics.FilesProcessed += int64(fileCount)
	p.metrics.ProcessingTime += totalTime
	p.metrics.ErrorCount += int64(len(errors))
	p.metrics.ConcurrencyLevel = p.maxWorkers

	if p.metrics.FilesProcessed > 0 {
		p.metrics.AverageFileTime = time.Duration(
			int64(p.metrics.ProcessingTime) / p.metrics.FilesProcessed,
		)
	}

	// Update memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memMB := int64(m.Alloc / 1024 / 1024)
	if memMB > p.metrics.MemoryPeakMB {
		p.metrics.MemoryPeakMB = memMB
	}
}

// updateFileMetrics updates per-file metrics
func (p *ConcurrentProcessor) updateFileMetrics(result FileResult) {
	p.metrics.mutex.Lock()
	defer p.metrics.mutex.Unlock()

	p.metrics.TotalBytes += result.Size
}

// GetMetrics returns current processing metrics
func (p *ConcurrentProcessor) GetMetrics() *ProcessingMetrics {
	p.metrics.mutex.RLock()
	defer p.metrics.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &ProcessingMetrics{
		FilesProcessed:   p.metrics.FilesProcessed,
		TotalBytes:       p.metrics.TotalBytes,
		ProcessingTime:   p.metrics.ProcessingTime,
		AverageFileTime:  p.metrics.AverageFileTime,
		ConcurrencyLevel: p.metrics.ConcurrencyLevel,
		MemoryPeakMB:     p.metrics.MemoryPeakMB,
		ErrorCount:       p.metrics.ErrorCount,
		CacheHits:        p.metrics.CacheHits,
		CacheMisses:      p.metrics.CacheMisses,
	}
}

// ResetMetrics resets all performance metrics
func (p *ConcurrentProcessor) ResetMetrics() {
	p.metrics.mutex.Lock()
	defer p.metrics.mutex.Unlock()

	p.metrics = &ProcessingMetrics{}
}

// String returns formatted metrics as string
func (m *ProcessingMetrics) String() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return fmt.Sprintf(
		"Performance Metrics:\n"+
			"  Files Processed: %d\n"+
			"  Total Size: %.2f MB\n"+
			"  Processing Time: %v\n"+
			"  Average File Time: %v\n"+
			"  Concurrency Level: %d workers\n"+
			"  Memory Peak: %d MB\n"+
			"  Errors: %d\n"+
			"  Cache Hit Rate: %.1f%%",
		m.FilesProcessed,
		float64(m.TotalBytes)/1024/1024,
		m.ProcessingTime,
		m.AverageFileTime,
		m.ConcurrencyLevel,
		m.MemoryPeakMB,
		m.ErrorCount,
		getCacheHitRate(m),
	)
}

func getCacheHitRate(m *ProcessingMetrics) float64 {
	total := m.CacheHits + m.CacheMisses
	if total == 0 {
		return 0
	}
	return float64(m.CacheHits) / float64(total) * 100
}