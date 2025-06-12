package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWASMPythonParser(t *testing.T) {
	// Create parser
	p, err := NewWASMPythonParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Test simple Python code
	source := []byte(`
def hello():
    print("Hello, world!")

hello()
`)

	tree, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if tree.RootType != "module" {
		t.Errorf("Expected root type 'module', got %s", tree.RootType)
	}

	if tree.NodeCount == 0 {
		t.Error("Expected some nodes in the tree")
	}

	if tree.HasErrors {
		t.Error("Simple valid Python should not have errors")
	}
}

func TestWASMPythonParserWithErrors(t *testing.T) {
	p, err := NewWASMPythonParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Test Python with syntax error
	source := []byte(`
def broken(:
    print("This has a syntax error"
`)

	tree, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if !tree.HasErrors {
		t.Error("Expected errors in malformed Python")
	}
}

func BenchmarkWASMParse(b *testing.B) {
	// Read test file
	source, err := os.ReadFile(filepath.Join("..", "..", "testdata", "simple.py"))
	if err != nil {
		// Create testdata if it doesn't exist
		source = []byte(`
def hello():
    """Say hello to the world."""
    print("Hello, world!")

class Config:
    """Configuration class."""
    
    def __init__(self, port=8080, host="localhost"):
        self.port = port
        self.host = host
    
    def get_url(self):
        return f"http://{self.host}:{self.port}"

if __name__ == "__main__":
    hello()
    config = Config()
    print(config.get_url())
`)
	}

	// Create parser once
	p, err := NewWASMPythonParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tree, err := p.Parse(source)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
		_ = tree
	}

	b.SetBytes(int64(len(source)))
}

func BenchmarkWASMStartup(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		p, err := NewWASMPythonParser()
		if err != nil {
			b.Fatalf("Failed to create parser: %v", err)
		}
		p.Close()
	}
}