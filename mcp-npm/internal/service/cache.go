package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTLSeconds int
	BaseDir    string
}

// DefaultCacheConfig returns default cache settings
func DefaultCacheConfig(cacheDir string) CacheConfig {
	return CacheConfig{
		TTLSeconds: 300, // 5 minutes
		BaseDir:    filepath.Join(cacheDir, "mcp"),
	}
}

// CacheEntry represents a cached response
type CacheEntry struct {
	CreatedAt     time.Time     `json:"created_at"`
	TTLSeconds    int           `json:"ttl_seconds"`
	DirectoryHash string        `json:"directory_hash"`
	Units         []ContentUnit `json:"units"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// IsValid checks if cache entry is still valid
func (ce *CacheEntry) IsValid() bool {
	elapsed := time.Since(ce.CreatedAt)
	return elapsed.Seconds() < float64(ce.TTLSeconds)
}

// ResponseCache manages MCP response caching
type ResponseCache struct {
	config CacheConfig
}

// NewResponseCache creates a new cache instance
func NewResponseCache(cacheDir string) *ResponseCache {
	config := DefaultCacheConfig(cacheDir)
	
	// Ensure cache directory exists
	os.MkdirAll(config.BaseDir, 0755)
	
	return &ResponseCache{
		config: config,
	}
}

// GenerateCacheKey creates a cache key from tool name and parameters
func (rc *ResponseCache) GenerateCacheKey(toolName string, params map[string]interface{}) string {
	// Create deterministic key by hashing tool name and sorted parameters
	h := sha256.New()
	h.Write([]byte(toolName))
	
	// Sort parameter keys for consistency
	keys := make([]string, 0, len(params))
	for k := range params {
		// Skip pagination parameters
		if k == "page_token" || k == "page_size" || k == "no_cache" || k == "cursor" {
			continue
		}
		keys = append(keys, k)
	}
	
	// Hash each parameter
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(fmt.Sprintf("%v", params[k])))
	}
	
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// Get retrieves a cached entry if valid
func (rc *ResponseCache) Get(cacheKey string) (*CacheEntry, error) {
	cachePath := filepath.Join(rc.config.BaseDir, cacheKey+".json")
	
	// Check if file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, nil // Cache miss
	}
	
	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}
	
	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}
	
	// Check validity
	if !entry.IsValid() {
		// Remove expired cache
		os.Remove(cachePath)
		return nil, nil // Cache expired
	}
	
	return &entry, nil
}

// Put stores a new cache entry
func (rc *ResponseCache) Put(cacheKey string, units []ContentUnit, dirHash string, metadata map[string]interface{}) error {
	entry := CacheEntry{
		CreatedAt:     time.Now(),
		TTLSeconds:    rc.config.TTLSeconds,
		DirectoryHash: dirHash,
		Units:         units,
		Metadata:      metadata,
	}
	
	// Serialize to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize cache: %w", err)
	}
	
	// Write to file
	cachePath := filepath.Join(rc.config.BaseDir, cacheKey+".json")
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}
	
	return nil
}

// Clear removes a specific cache entry
func (rc *ResponseCache) Clear(cacheKey string) error {
	cachePath := filepath.Join(rc.config.BaseDir, cacheKey+".json")
	return os.Remove(cachePath)
}

// ClearExpired removes all expired cache entries
func (rc *ResponseCache) ClearExpired() error {
	entries, err := os.ReadDir(rc.config.BaseDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			cachePath := filepath.Join(rc.config.BaseDir, entry.Name())
			
			// Try to read and check validity
			data, err := os.ReadFile(cachePath)
			if err != nil {
				continue
			}
			
			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				// Invalid cache file, remove it
				os.Remove(cachePath)
				continue
			}
			
			if !cacheEntry.IsValid() {
				os.Remove(cachePath)
			}
		}
	}
	
	return nil
}

// timeNow is a wrapper for time.Now() to allow mocking in tests
var timeNow = time.Now