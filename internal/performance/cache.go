package performance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// Cache provides intelligent caching for parsed files
type Cache struct {
	dir           string
	maxSize       int64
	maxAge        time.Duration
	enableMetrics bool
	metrics       *CacheMetrics
	mutex         sync.RWMutex
	index         map[string]*CacheEntry
}

// CacheEntry represents a cached file entry
type CacheEntry struct {
	Key          string    `json:"key"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	FileModTime  time.Time `json:"file_mod_time"`
	Language     string    `json:"language"`
	Version      string    `json:"version"`
	CachedAt     time.Time `json:"cached_at"`
	AccessCount  int64     `json:"access_count"`
	LastAccess   time.Time `json:"last_access"`
	Options      string    `json:"options"`
	ResultPath   string    `json:"result_path"`
	ResultSize   int64     `json:"result_size"`
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits           int64
	Misses         int64
	Evictions      int64
	TotalSize      int64
	EntryCount     int64
	AverageHitTime time.Duration
	mutex          sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache(cacheDir string) *Cache {
	cache := &Cache{
		dir:           cacheDir,
		maxSize:       1024 * 1024 * 1024, // 1GB default
		maxAge:        24 * time.Hour,      // 24 hours default
		enableMetrics: true,
		metrics:       &CacheMetrics{},
		index:         make(map[string]*CacheEntry),
	}

	// Create cache directory
	_ = os.MkdirAll(cacheDir, 0755)

	// Load existing cache index
	cache.loadIndex()

	return cache
}

// WithMaxSize sets the maximum cache size in bytes
func (c *Cache) WithMaxSize(size int64) *Cache {
	c.maxSize = size
	return c
}

// WithMaxAge sets the maximum age for cache entries
func (c *Cache) WithMaxAge(age time.Duration) *Cache {
	c.maxAge = age
	return c
}

// WithMetrics enables or disables cache metrics
func (c *Cache) WithMetrics(enabled bool) *Cache {
	c.enableMetrics = enabled
	return c
}

// Get retrieves a cached result if available and valid
func (c *Cache) Get(filePath string, opts processor.ProcessOptions) (*ir.DistilledFile, bool) {
	startTime := time.Now()

	key := c.generateKey(filePath, opts)

	c.mutex.RLock()
	entry, exists := c.index[key]
	c.mutex.RUnlock()

	if !exists {
		c.recordMiss()
		return nil, false
	}

	// Check if file has been modified
	info, err := os.Stat(filePath)
	if err != nil || info.ModTime().After(entry.FileModTime) {
		// File modified, invalidate cache
		_ = c.Remove(key)
		c.recordMiss()
		return nil, false
	}

	// Check if cache entry is too old
	if time.Since(entry.CachedAt) > c.maxAge {
		_ = c.Remove(key)
		c.recordMiss()
		return nil, false
	}

	// Load cached result
	result, err := c.loadCachedResult(entry)
	if err != nil {
		_ = c.Remove(key)
		c.recordMiss()
		return nil, false
	}

	// Update access statistics
	c.updateAccess(entry)

	// Record hit
	c.recordHit(time.Since(startTime))

	return result, true
}

// Put stores a result in the cache
func (c *Cache) Put(filePath string, opts processor.ProcessOptions, result *ir.DistilledFile) error {
	key := c.generateKey(filePath, opts)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Store result
	resultPath := filepath.Join(c.dir, key+".json")
	if err := c.storeCachedResult(result, resultPath); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Get result size
	resultInfo, err := os.Stat(resultPath)
	if err != nil {
		return fmt.Errorf("failed to stat result file: %w", err)
	}

	// Create cache entry
	entry := &CacheEntry{
		Key:         key,
		FilePath:    filePath,
		FileSize:    info.Size(),
		FileModTime: info.ModTime(),
		Language:    result.Language,
		Version:     result.Version,
		CachedAt:    time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Options:     c.serializeOptions(opts),
		ResultPath:  resultPath,
		ResultSize:  resultInfo.Size(),
	}

	// Add to index
	c.mutex.Lock()
	c.index[key] = entry
	c.mutex.Unlock()

	// Save index
	c.saveIndex()

	// Check size limits and evict if necessary
	c.evictIfNecessary()

	return nil
}

// Remove removes an entry from the cache
func (c *Cache) Remove(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.index[key]
	if !exists {
		return nil
	}

	// Remove result file
	if err := os.Remove(entry.ResultPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove from index
	delete(c.index, key)

	// Update metrics
	if c.enableMetrics {
		c.metrics.mutex.Lock()
		c.metrics.TotalSize -= entry.ResultSize
		c.metrics.EntryCount--
		c.metrics.mutex.Unlock()
	}

	return nil
}

// Clear removes all entries from the cache
func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove all result files
	for key, entry := range c.index {
		if err := os.Remove(entry.ResultPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		delete(c.index, key)
	}

	// Reset metrics
	if c.enableMetrics {
		c.metrics.mutex.Lock()
		c.metrics.TotalSize = 0
		c.metrics.EntryCount = 0
		c.metrics.mutex.Unlock()
	}

	// Save empty index
	c.saveIndex()

	return nil
}

// Stats returns cache statistics
func (c *Cache) Stats() *CacheMetrics {
	if !c.enableMetrics {
		return nil
	}

	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	// Return a copy
	return &CacheMetrics{
		Hits:           c.metrics.Hits,
		Misses:         c.metrics.Misses,
		Evictions:      c.metrics.Evictions,
		TotalSize:      c.metrics.TotalSize,
		EntryCount:     c.metrics.EntryCount,
		AverageHitTime: c.metrics.AverageHitTime,
	}
}

// generateKey creates a unique key for the file and options
func (c *Cache) generateKey(filePath string, opts processor.ProcessOptions) string {
	h := sha256.New()
	h.Write([]byte(filePath))
	h.Write([]byte(c.serializeOptions(opts)))
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars
}

// serializeOptions converts options to a string for caching
func (c *Cache) serializeOptions(opts processor.ProcessOptions) string {
	data, _ := json.Marshal(opts)
	return string(data)
}

// loadCachedResult loads a result from disk
func (c *Cache) loadCachedResult(entry *CacheEntry) (*ir.DistilledFile, error) {
	data, err := os.ReadFile(entry.ResultPath)
	if err != nil {
		return nil, err
	}

	var result ir.DistilledFile
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// storeCachedResult stores a result to disk
func (c *Cache) storeCachedResult(result *ir.DistilledFile, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// updateAccess updates access statistics for an entry
func (c *Cache) updateAccess(entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry.AccessCount++
	entry.LastAccess = time.Now()
}

// recordHit records a cache hit
func (c *Cache) recordHit(duration time.Duration) {
	if !c.enableMetrics {
		return
	}

	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.Hits++

	// Update average hit time
	if c.metrics.Hits == 1 {
		c.metrics.AverageHitTime = duration
	} else {
		total := time.Duration(c.metrics.Hits-1) * c.metrics.AverageHitTime
		c.metrics.AverageHitTime = (total + duration) / time.Duration(c.metrics.Hits)
	}
}

// recordMiss records a cache miss
func (c *Cache) recordMiss() {
	if !c.enableMetrics {
		return
	}

	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.Misses++
}

// evictIfNecessary removes old entries if cache is too large
func (c *Cache) evictIfNecessary() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Calculate current size
	var totalSize int64
	for _, entry := range c.index {
		totalSize += entry.ResultSize
	}

	if totalSize <= c.maxSize {
		return
	}

	// Sort entries by last access time (oldest first)
	type entryWithKey struct {
		key   string
		entry *CacheEntry
	}

	var entries []entryWithKey
	for key, entry := range c.index {
		entries = append(entries, entryWithKey{key, entry})
	}

	// Sort by last access time
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.LastAccess.After(entries[j].entry.LastAccess) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries until under size limit
	for _, entryWithKey := range entries {
		if totalSize <= c.maxSize {
			break
		}

		entry := entryWithKey.entry

		// Remove file
		os.Remove(entry.ResultPath)

		// Remove from index
		delete(c.index, entryWithKey.key)

		totalSize -= entry.ResultSize

		// Update metrics
		if c.enableMetrics {
			c.metrics.mutex.Lock()
			c.metrics.Evictions++
			c.metrics.TotalSize -= entry.ResultSize
			c.metrics.EntryCount--
			c.metrics.mutex.Unlock()
		}
	}

	// Save updated index
	c.saveIndex()
}

// loadIndex loads the cache index from disk
func (c *Cache) loadIndex() {
	indexPath := filepath.Join(c.dir, "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return // Index doesn't exist yet
	}

	var index map[string]*CacheEntry
	if err := json.Unmarshal(data, &index); err != nil {
		return // Corrupted index
	}

	c.mutex.Lock()
	c.index = index
	c.mutex.Unlock()

	// Update metrics
	if c.enableMetrics {
		var totalSize int64
		var entryCount int64

		for _, entry := range index {
			totalSize += entry.ResultSize
			entryCount++
		}

		c.metrics.mutex.Lock()
		c.metrics.TotalSize = totalSize
		c.metrics.EntryCount = entryCount
		c.metrics.mutex.Unlock()
	}
}

// saveIndex saves the cache index to disk
func (c *Cache) saveIndex() {
	indexPath := filepath.Join(c.dir, "index.json")

	c.mutex.RLock()
	data, err := json.MarshalIndent(c.index, "", "  ")
	c.mutex.RUnlock()

	if err != nil {
		return
	}

	_ = os.WriteFile(indexPath, data, 0644)
}

// GetHitRate returns the cache hit rate as a percentage
func (m *CacheMetrics) GetHitRate() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	total := m.Hits + m.Misses
	if total == 0 {
		return 0
	}
	return float64(m.Hits) / float64(total) * 100
}

// String returns formatted cache metrics
func (m *CacheMetrics) String() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return fmt.Sprintf(
		"Cache Metrics:\n"+
			"  Hit Rate: %.1f%% (%d hits, %d misses)\n"+
			"  Entries: %d\n"+
			"  Total Size: %.2f MB\n"+
			"  Evictions: %d\n"+
			"  Average Hit Time: %v",
		m.GetHitRate(),
		m.Hits,
		m.Misses,
		m.EntryCount,
		float64(m.TotalSize)/1024/1024,
		m.Evictions,
		m.AverageHitTime,
	)
}

// CachedProcessor combines normal processing with intelligent caching
type CachedProcessor struct {
	processor *processor.Processor
	cache     *Cache
}

// NewCachedProcessor creates a new processor with caching
func NewCachedProcessor(cacheDir string) *CachedProcessor {
	return &CachedProcessor{
		processor: processor.New(),
		cache:     NewCache(cacheDir),
	}
}

// ProcessFile processes a file with caching
func (p *CachedProcessor) ProcessFile(filePath string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	// Try cache first
	if result, hit := p.cache.Get(filePath, opts); hit {
		return result, nil
	}

	// Process file
	result, err := p.processor.ProcessFile(filePath, opts)
	if err != nil {
		return nil, err
	}

	// Cache result
	if err := p.cache.Put(filePath, opts, result); err != nil {
		// Log warning but don't fail the operation
		fmt.Fprintf(os.Stderr, "Warning: failed to cache result for %s: %v\n", filePath, err)
	}

	return result, nil
}

// GetCache returns the cache instance for direct access
func (p *CachedProcessor) GetCache() *Cache {
	return p.cache
}