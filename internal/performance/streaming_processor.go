package performance

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// StreamingProcessor handles very large files with memory-efficient streaming
type StreamingProcessor struct {
	chunkSize    int
	bufferSize   int
	maxMemoryMB  int64
	enableMetrics bool
}

// StreamingResult contains results from streaming processing
type StreamingResult struct {
	File         *ir.DistilledFile
	ChunksRead   int
	TotalLines   int
	ProcessTime  time.Duration
	PeakMemoryMB int64
	Error        error
}

// ChunkProcessor interface for processing file chunks
type ChunkProcessor interface {
	ProcessChunk(chunk []string, chunkIndex int) ([]*ir.DistilledNode, error)
	Finalize(allNodes []*ir.DistilledNode) (*ir.DistilledFile, error)
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor() *StreamingProcessor {
	return &StreamingProcessor{
		chunkSize:    1000,   // Lines per chunk
		bufferSize:   64 * 1024, // 64KB buffer
		maxMemoryMB:  512,    // 512MB memory limit
		enableMetrics: true,
	}
}

// WithChunkSize sets the number of lines per chunk
func (p *StreamingProcessor) WithChunkSize(size int) *StreamingProcessor {
	if size > 0 {
		p.chunkSize = size
	}
	return p
}

// WithBufferSize sets the I/O buffer size
func (p *StreamingProcessor) WithBufferSize(size int) *StreamingProcessor {
	if size > 0 {
		p.bufferSize = size
	}
	return p
}

// WithMemoryLimit sets the maximum memory usage in MB
func (p *StreamingProcessor) WithMemoryLimit(limitMB int64) *StreamingProcessor {
	if limitMB > 0 {
		p.maxMemoryMB = limitMB
	}
	return p
}

// ProcessLargeFile processes a large file using streaming approach
func (p *StreamingProcessor) ProcessLargeFile(
	ctx context.Context,
	filePath string,
	opts processor.ProcessOptions,
) (*StreamingResult, error) {
	startTime := time.Now()

	// Check file size (for potential future optimizations)
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create buffered reader
	reader := bufio.NewReaderSize(file, p.bufferSize)

	// Process file in chunks
	result, err := p.processStreamingFile(ctx, reader, filePath, opts)
	if err != nil {
		return nil, err
	}

	result.ProcessTime = time.Since(startTime)
	return result, nil
}

// ProcessReader processes an io.Reader using streaming approach
func (p *StreamingProcessor) ProcessReader(
	ctx context.Context,
	reader io.Reader,
	filename string,
	opts processor.ProcessOptions,
) (*StreamingResult, error) {
	startTime := time.Now()

	// Create buffered reader if not already buffered
	var bufferedReader *bufio.Reader
	if br, ok := reader.(*bufio.Reader); ok {
		bufferedReader = br
	} else {
		bufferedReader = bufio.NewReaderSize(reader, p.bufferSize)
	}

	// Process file in chunks
	result, err := p.processStreamingFile(ctx, bufferedReader, filename, opts)
	if err != nil {
		return nil, err
	}

	result.ProcessTime = time.Since(startTime)
	return result, nil
}

// processStreamingFile implements the core streaming logic
func (p *StreamingProcessor) processStreamingFile(
	ctx context.Context,
	reader *bufio.Reader,
	filename string,
	opts processor.ProcessOptions,
) (*StreamingResult, error) {
	result := &StreamingResult{
		ChunksRead: 0,
		TotalLines: 0,
	}

	// Get appropriate processor
	proc, ok := processor.GetByFilename(filename)
	if !ok {
		return nil, fmt.Errorf("no processor found for file: %s", filename)
	}

	// Create chunk processor based on language
	chunkProcessor, err := p.createChunkProcessor(proc, filename, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk processor: %w", err)
	}

	var allNodes []*ir.DistilledNode
	var currentChunk []string
	var lineCount int

	// Read file line by line
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}

		if line != "" {
			currentChunk = append(currentChunk, strings.TrimRight(line, "\n\r"))
			lineCount++
		}

		// Process chunk when it's full or we've reached EOF
		if len(currentChunk) >= p.chunkSize || (err == io.EOF && len(currentChunk) > 0) {
			nodes, chunkErr := chunkProcessor.ProcessChunk(currentChunk, result.ChunksRead)
			if chunkErr != nil {
				result.Error = chunkErr
				// Continue processing other chunks
			} else {
				allNodes = append(allNodes, nodes...)
			}

			result.ChunksRead++
			result.TotalLines += len(currentChunk)
			currentChunk = nil

			// Check memory usage
			if p.enableMetrics {
				currentMemMB := p.getCurrentMemoryMB()
				if currentMemMB > p.maxMemoryMB {
					return nil, fmt.Errorf("memory limit exceeded: %d MB > %d MB", currentMemMB, p.maxMemoryMB)
				}
				if currentMemMB > result.PeakMemoryMB {
					result.PeakMemoryMB = currentMemMB
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	// Finalize processing
	file, err := chunkProcessor.Finalize(allNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize processing: %w", err)
	}

	result.File = file
	return result, nil
}

// createChunkProcessor creates appropriate chunk processor based on language
func (p *StreamingProcessor) createChunkProcessor(
	proc processor.LanguageProcessor,
	filename string,
	opts processor.ProcessOptions,
) (ChunkProcessor, error) {
	language := proc.Language()

	switch language {
	case "python":
		return NewPythonChunkProcessor(filename, opts), nil
	case "javascript", "typescript":
		return NewJavaScriptChunkProcessor(filename, opts), nil
	case "java":
		return NewJavaChunkProcessor(filename, opts), nil
	case "go":
		return NewGoChunkProcessor(filename, opts), nil
	default:
		return NewGenericChunkProcessor(proc, filename, opts), nil
	}
}

// getCurrentMemoryMB returns current memory usage in MB
func (p *StreamingProcessor) getCurrentMemoryMB() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc / 1024 / 1024)
}

// Generic chunk processor for any language
type GenericChunkProcessor struct {
	processor processor.LanguageProcessor
	filename  string
	opts      processor.ProcessOptions
	nodes     []*ir.DistilledNode
}

func NewGenericChunkProcessor(
	proc processor.LanguageProcessor,
	filename string,
	opts processor.ProcessOptions,
) *GenericChunkProcessor {
	return &GenericChunkProcessor{
		processor: proc,
		filename:  filename,
		opts:      opts,
		nodes:     make([]*ir.DistilledNode, 0),
	}
}

func (p *GenericChunkProcessor) ProcessChunk(
	chunk []string,
	chunkIndex int,
) ([]*ir.DistilledNode, error) {
	// Combine chunk lines
	content := strings.Join(chunk, "\n")
	reader := strings.NewReader(content)

	// Process chunk
	ctx := context.Background()
	file, err := p.processor.Process(ctx, reader, p.filename)
	if err != nil {
		return nil, err
	}

	// Extract top-level nodes and convert to []*ir.DistilledNode
	var nodes []*ir.DistilledNode
	for _, child := range file.Children {
		nodes = append(nodes, &child)
	}
	return nodes, nil
}

func (p *GenericChunkProcessor) Finalize(
	allNodes []*ir.DistilledNode,
) (*ir.DistilledFile, error) {
	// Convert []*ir.DistilledNode back to []ir.DistilledNode
	var children []ir.DistilledNode
	for _, node := range allNodes {
		if node != nil {
			children = append(children, *node)
		}
	}

	// Create final distilled file
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   len(children),
			},
		},
		Path:     p.filename,
		Language: p.processor.Language(),
		Version:  p.processor.Version(),
		Children: children,
		Errors:   []ir.DistilledError{},
	}

	return file, nil
}

// Language-specific chunk processors would be implemented here
// For now, we'll use the generic processor for all languages

type PythonChunkProcessor struct {
	*GenericChunkProcessor
}

func NewPythonChunkProcessor(filename string, opts processor.ProcessOptions) *PythonChunkProcessor {
	// Get Python processor
	proc, _ := processor.Get("python")
	return &PythonChunkProcessor{
		GenericChunkProcessor: NewGenericChunkProcessor(proc, filename, opts),
	}
}

type JavaScriptChunkProcessor struct {
	*GenericChunkProcessor
}

func NewJavaScriptChunkProcessor(filename string, opts processor.ProcessOptions) *JavaScriptChunkProcessor {
	// Get JavaScript processor
	proc, _ := processor.Get("javascript")
	return &JavaScriptChunkProcessor{
		GenericChunkProcessor: NewGenericChunkProcessor(proc, filename, opts),
	}
}

type JavaChunkProcessor struct {
	*GenericChunkProcessor
}

func NewJavaChunkProcessor(filename string, opts processor.ProcessOptions) *JavaChunkProcessor {
	// Get Java processor
	proc, _ := processor.Get("java")
	return &JavaChunkProcessor{
		GenericChunkProcessor: NewGenericChunkProcessor(proc, filename, opts),
	}
}

type GoChunkProcessor struct {
	*GenericChunkProcessor
}

func NewGoChunkProcessor(filename string, opts processor.ProcessOptions) *GoChunkProcessor {
	// Get Go processor
	proc, _ := processor.Get("golang")
	return &GoChunkProcessor{
		GenericChunkProcessor: NewGenericChunkProcessor(proc, filename, opts),
	}
}

// StreamingBenchmark provides benchmarking capabilities
type StreamingBenchmark struct {
	ChunkSizes     []int
	BufferSizes    []int
	MemoryLimits   []int64
	TestFiles      []string
	Results        map[string]*StreamingBenchmarkResult
}

type StreamingBenchmarkResult struct {
	ChunkSize      int
	BufferSize     int
	MemoryLimitMB  int64
	ProcessingTime time.Duration
	PeakMemoryMB   int64
	ThroughputMBps float64
	Success        bool
	Error          error
}

// RunBenchmark runs comprehensive streaming performance benchmarks
func (b *StreamingBenchmark) RunBenchmark(ctx context.Context) error {
	b.Results = make(map[string]*StreamingBenchmarkResult)

	for _, testFile := range b.TestFiles {
		for _, chunkSize := range b.ChunkSizes {
			for _, bufferSize := range b.BufferSizes {
				for _, memoryLimit := range b.MemoryLimits {
					key := fmt.Sprintf("%s_chunk%d_buffer%d_mem%d",
						testFile, chunkSize, bufferSize, memoryLimit)

					result := b.runSingleBenchmark(ctx, testFile, chunkSize, bufferSize, memoryLimit)
					b.Results[key] = result
				}
			}
		}
	}

	return nil
}

func (b *StreamingBenchmark) runSingleBenchmark(
	ctx context.Context,
	testFile string,
	chunkSize int,
	bufferSize int,
	memoryLimit int64,
) *StreamingBenchmarkResult {
	result := &StreamingBenchmarkResult{
		ChunkSize:     chunkSize,
		BufferSize:    bufferSize,
		MemoryLimitMB: memoryLimit,
	}

	// Get file size
	info, err := os.Stat(testFile)
	if err != nil {
		result.Error = err
		return result
	}

	// Create streaming processor
	streamingProcessor := NewStreamingProcessor().
		WithChunkSize(chunkSize).
		WithBufferSize(bufferSize).
		WithMemoryLimit(memoryLimit)

	// Process file
	startTime := time.Now()
	streamResult, err := streamingProcessor.ProcessLargeFile(ctx, testFile, processor.DefaultProcessOptions())
	endTime := time.Now()

	if err != nil {
		result.Error = err
		return result
	}

	result.ProcessingTime = endTime.Sub(startTime)
	result.PeakMemoryMB = streamResult.PeakMemoryMB
	result.ThroughputMBps = float64(info.Size()) / 1024 / 1024 / result.ProcessingTime.Seconds()
	result.Success = true

	return result
}

// GetBestConfiguration returns the optimal configuration based on benchmark results
func (b *StreamingBenchmark) GetBestConfiguration() (int, int, int64, error) {
	if len(b.Results) == 0 {
		return 0, 0, 0, fmt.Errorf("no benchmark results available")
	}

	var bestResult *StreamingBenchmarkResult
	var bestKey string
	bestScore := 0.0

	for key, result := range b.Results {
		if !result.Success {
			continue
		}

		// Score based on throughput and memory efficiency
		score := result.ThroughputMBps / float64(result.PeakMemoryMB)
		if score > bestScore {
			bestScore = score
			bestResult = result
			bestKey = key
		}
	}

	if bestResult == nil {
		return 0, 0, 0, fmt.Errorf("no successful benchmark results found")
	}

	fmt.Printf("Best configuration: %s (score: %.4f)\n", bestKey, bestScore)
	return bestResult.ChunkSize, bestResult.BufferSize, bestResult.MemoryLimitMB, nil
}