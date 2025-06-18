package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// PaginationConfig holds pagination settings
type PaginationConfig struct {
	DefaultPageSize int
	MinPageSize     int
	MaxPageSize     int
}

// DefaultPaginationConfig returns default pagination settings
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPageSize: 20000,  // tokens
		MinPageSize:     1000,
		MaxPageSize:     20000,
	}
}

// PageToken represents the state for cursor-based pagination
type PageToken struct {
	PageNumber    int    `json:"page"`
	DirectoryHash string `json:"dir_hash"`
	Timestamp     int64  `json:"ts"`
	LastItemPath  string `json:"last_path"`
}

// Encode converts PageToken to opaque string
func (pt *PageToken) Encode() string {
	data, _ := json.Marshal(pt)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodePageToken decodes an opaque page token string
func DecodePageToken(token string) (*PageToken, error) {
	if token == "" {
		return nil, nil
	}
	
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid page token format: %w", err)
	}
	
	var pt PageToken
	if err := json.Unmarshal(data, &pt); err != nil {
		return nil, fmt.Errorf("invalid page token content: %w", err)
	}
	
	return &pt, nil
}

// EstimateTokens estimates the number of tokens in a text string
// Uses the rule: ~4 characters = 1 token, plus 10% margin for JSON overhead
func EstimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 4.0 * 1.1))
}

// EstimateJSONTokens estimates tokens for a JSON-serializable object
func EstimateJSONTokens(obj interface{}) int {
	data, err := json.Marshal(obj)
	if err != nil {
		// Fallback estimation based on string representation
		return EstimateTokens(fmt.Sprintf("%+v", obj))
	}
	return EstimateTokens(string(data))
}

// ContentUnit represents a single unit of content that can be paginated
type ContentUnit struct {
	Path         string      `json:"path"`
	Type         string      `json:"type"` // "file", "directory", "summary"
	Content      interface{} `json:"content"`
	EstimatedTokens int     `json:"estimated_tokens"`
}

// PaginatedResult holds the result of pagination
type PaginatedResult struct {
	Units           []ContentUnit `json:"units"`
	EstimatedTokens int          `json:"estimated_tokens"`
	NextPageToken   *string      `json:"next_page_token,omitempty"`
	TotalPages      int          `json:"total_pages,omitempty"`
	CurrentPage     int          `json:"current_page"`
	IsTruncated     bool         `json:"is_truncated,omitempty"`
}

// Paginator handles pagination logic
type Paginator struct {
	config   PaginationConfig
	units    []ContentUnit
	dirHash  string
}

// NewPaginator creates a new paginator
func NewPaginator(units []ContentUnit, dirHash string) *Paginator {
	return &Paginator{
		config:  DefaultPaginationConfig(),
		units:   units,
		dirHash: dirHash,
	}
}

// GetPage returns a specific page of results
func (p *Paginator) GetPage(pageToken *PageToken, requestedPageSize int) (*PaginatedResult, error) {
	// Validate page size
	pageSize := requestedPageSize
	if pageSize <= 0 {
		pageSize = p.config.DefaultPageSize
	} else if pageSize < p.config.MinPageSize {
		pageSize = p.config.MinPageSize
	} else if pageSize > p.config.MaxPageSize {
		pageSize = p.config.MaxPageSize
	}
	
	// Determine starting position
	startIdx := 0
	if pageToken != nil {
		// Validate token matches current data
		if pageToken.DirectoryHash != p.dirHash {
			return nil, fmt.Errorf("page token is stale: directory contents have changed")
		}
		
		// Find the last item from previous page
		for i, unit := range p.units {
			if unit.Path == pageToken.LastItemPath {
				startIdx = i + 1
				break
			}
		}
	}
	
	// Build page by accumulating units until we hit the token limit
	result := &PaginatedResult{
		Units:           []ContentUnit{},
		EstimatedTokens: 0,
		CurrentPage:     1,
	}
	
	if pageToken != nil {
		result.CurrentPage = pageToken.PageNumber + 1
	}
	
	tokenBudget := pageSize
	for i := startIdx; i < len(p.units); i++ {
		unit := p.units[i]
		
		// Check if adding this unit would exceed our budget
		if result.EstimatedTokens > 0 && result.EstimatedTokens+unit.EstimatedTokens > tokenBudget {
			// Create next page token
			nextToken := &PageToken{
				PageNumber:    result.CurrentPage,
				DirectoryHash: p.dirHash,
				Timestamp:     timeNow().Unix(),
				LastItemPath:  p.units[i-1].Path,
			}
			encoded := nextToken.Encode()
			result.NextPageToken = &encoded
			break
		}
		
		result.Units = append(result.Units, unit)
		result.EstimatedTokens += unit.EstimatedTokens
	}
	
	// Calculate total pages (approximate)
	avgTokensPerUnit := p.calculateAverageTokensPerUnit()
	if avgTokensPerUnit > 0 {
		result.TotalPages = int(math.Ceil(float64(len(p.units)) / float64(pageSize/avgTokensPerUnit)))
	}
	
	return result, nil
}

// calculateAverageTokensPerUnit calculates the average tokens per unit
func (p *Paginator) calculateAverageTokensPerUnit() int {
	if len(p.units) == 0 {
		return 0
	}
	
	totalTokens := 0
	for _, unit := range p.units {
		totalTokens += unit.EstimatedTokens
	}
	
	return totalTokens / len(p.units)
}

// CalculateDirectoryHash creates a hash of directory contents for cache invalidation
func CalculateDirectoryHash(paths []string) string {
	h := sha256.New()
	for _, path := range paths {
		h.Write([]byte(path))
		h.Write([]byte("\n"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// SplitIntoUnits splits content into pageable units
func SplitIntoUnits(content string, format string) []ContentUnit {
	if format != "text" {
		// For non-text formats, return as single unit
		return []ContentUnit{
			{
				Type:            "content",
				Content:         content,
				EstimatedTokens: EstimateTokens(content),
			},
		}
	}
	
	// For text format, split by file markers
	units := []ContentUnit{}
	files := strings.Split(content, "<file path=")
	
	for i, file := range files {
		if i == 0 && file == "" {
			continue // Skip empty first split
		}
		
		fileContent := file
		if i > 0 {
			fileContent = "<file path=" + file
		}
		
		// Extract file path
		pathEnd := strings.Index(fileContent, ">")
		if pathEnd == -1 {
			continue
		}
		
		path := ""
		if i > 0 {
			pathStart := strings.Index(fileContent, `"`) + 1
			pathEndQuote := strings.Index(fileContent[pathStart:], `"`)
			if pathEndQuote > 0 {
				path = fileContent[pathStart : pathStart+pathEndQuote]
			}
		}
		
		units = append(units, ContentUnit{
			Path:            path,
			Type:            "file",
			Content:         fileContent,
			EstimatedTokens: EstimateTokens(fileContent),
		})
	}
	
	return units
}