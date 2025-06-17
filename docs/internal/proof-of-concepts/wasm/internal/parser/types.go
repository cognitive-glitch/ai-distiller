package parser

// ParseTree represents the result of parsing
type ParseTree struct {
	RootType  string
	NodeCount int
	HasErrors bool
	// In a real implementation, this would contain the actual tree structure
}