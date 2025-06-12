package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonParser(t *testing.T) {
	// Create parser
	p, err := NewPythonParser()
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
	defer tree.Delete()

	root := tree.RootNode()
	if root == nil {
		t.Fatal("Root node is nil")
	}

	if root.HasError() {
		t.Error("Parse tree contains errors")
	}

	// Check we have some nodes
	if root.ChildCount() == 0 {
		t.Error("Root has no children")
	}
}

func BenchmarkParse(b *testing.B) {
	// Read test file
	source, err := os.ReadFile(filepath.Join("..", "..", "testdata", "simple.py"))
	if err != nil {
		b.Fatalf("Failed to read test file: %v", err)
	}

	// Create parser once
	p, err := NewPythonParser()
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
		tree.Delete()
	}

	b.SetBytes(int64(len(source)))
}

func BenchmarkStartup(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		p, err := NewPythonParser()
		if err != nil {
			b.Fatalf("Failed to create parser: %v", err)
		}
		p.Close()
	}
}