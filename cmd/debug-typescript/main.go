//go:build cgo
// +build cgo

package main

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	typescript "tree-sitter-typescript"
)

func main() {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(typescript.Language()))

	// Test namespace declaration
	code := `
// Namespace declaration
namespace Analytics {
    export interface Event {
        name: string;
        data: Record<string, any>;
    }
    
    export function trackEvent(event: Event): void {
        console.log('Tracking:', event.name);
    }
}
`

	tree := parser.Parse(nil, []byte(code))
	if tree == nil {
		panic("Failed to parse")
	}
	defer tree.Close()

	// Print AST
	printAST(tree.RootNode(), []byte(code), 0)
}

func printAST(node *sitter.Node, source []byte, depth int) {
	indent := strings.Repeat("  ", depth)
	nodeType := node.Type()
	
	if !node.IsNamed() {
		return
	}
	
	text := ""
	if int(node.EndByte()) <= len(source) {
		text = string(source[node.StartByte():node.EndByte()])
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		text = strings.ReplaceAll(text, "\n", "\\n")
	}
	
	fmt.Printf("%s%s: %s\n", indent, nodeType, text)
	
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		printAST(child, source, depth+1)
	}
}