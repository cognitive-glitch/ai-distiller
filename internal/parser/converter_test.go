package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConverter(t *testing.T) {
	source := []byte("def hello():\n    print('world')")
	conv := NewConverter("python", source)
	
	assert.Equal(t, "python", conv.language)
	assert.Equal(t, source, conv.source)
	assert.Empty(t, conv.errors)
}

func TestConverterHelpers(t *testing.T) {
	t.Run("ByteToLineCol", func(t *testing.T) {
		source := []byte("line1\nline2\nline3")
		conv := NewConverter("test", source)
		
		tests := []struct {
			offset uint32
			line   int
			col    int
		}{
			{0, 1, 1},     // Start of file
			{4, 1, 5},     // Before first newline
			{5, 1, 6},     // At first newline
			{6, 2, 1},     // Start of second line
			{11, 2, 6},    // At second newline
			{12, 3, 1},    // Start of third line
			{16, 3, 5},    // End of file
		}
		
		for _, tt := range tests {
			line, col := conv.byteToLineCol(tt.offset)
			assert.Equal(t, tt.line, line, "offset %d", tt.offset)
			assert.Equal(t, tt.col, col, "offset %d", tt.offset)
		}
	})
	
	t.Run("CountLines", func(t *testing.T) {
		tests := []struct {
			source string
			lines  int
		}{
			{"", 0},
			{"single line", 1},
			{"line1\nline2", 2},
			{"line1\nline2\n", 3},
			{"line1\nline2\nline3", 3},
			{"\n\n\n", 4},
		}
		
		conv := NewConverter("test", nil)
		for _, tt := range tests {
			lines := conv.countLines([]byte(tt.source))
			assert.Equal(t, tt.lines, lines, "source: %q", tt.source)
		}
	})
	
	t.Run("LastLineLength", func(t *testing.T) {
		tests := []struct {
			source string
			length int
		}{
			{"", 0},
			{"hello", 5},
			{"line1\nline2", 5},
			{"line1\n", 0},
			{"line1\nab", 2},
		}
		
		conv := NewConverter("test", nil)
		for _, tt := range tests {
			length := conv.lastLineLength([]byte(tt.source))
			assert.Equal(t, tt.length, length, "source: %q", tt.source)
		}
	})
}

// MockNode for testing without WASM
type MockNode struct {
	typ        string
	startByte  uint32
	endByte    uint32
	children   []*MockNode
	isNamed    bool
	isError    bool
}

func (m *MockNode) Type() (string, error) {
	return m.typ, nil
}

func (m *MockNode) ChildCount() (uint32, error) {
	return uint32(len(m.children)), nil
}

func (m *MockNode) Child(index uint32) (*Node, error) {
	if index >= uint32(len(m.children)) {
		return nil, nil
	}
	// Return a mock node - in real implementation would convert
	return &Node{}, nil
}

func (m *MockNode) IsNamed() (bool, error) {
	return m.isNamed, nil
}

func (m *MockNode) IsError() (bool, error) {
	return m.isError, nil
}

func (m *MockNode) StartByte() (uint32, error) {
	return m.startByte, nil
}

func (m *MockNode) EndByte() (uint32, error) {
	return m.endByte, nil
}

func TestGetNodeLocation(t *testing.T) {
	source := []byte("line1\nline2\nline3")
	conv := NewConverter("test", source)
	
	// Create a mock node wrapper for testing
	mockNode := &MockNode{
		startByte: 6,  // Start of "line2"
		endByte:   11, // End of "line2"
	}
	
	// We need to wrap this in a real Node for the test
	// For now, we'll test the byteToLineCol directly
	startLine, startCol := conv.byteToLineCol(mockNode.startByte)
	endLine, endCol := conv.byteToLineCol(mockNode.endByte)
	
	assert.Equal(t, 2, startLine)
	assert.Equal(t, 1, startCol)
	assert.Equal(t, 2, endLine)
	assert.Equal(t, 6, endCol)
}