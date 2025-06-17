//go:build !cgo
// +build !cgo

package language

import (
	"context"
	"fmt"
	"io"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/janreges/ai-distiller/internal/processor"
)

// RegisterTreeSitterProcessors registers stub processors when CGO is disabled
func RegisterTreeSitterProcessors() {
	// Register stub processors that return errors
	stubLanguages := []string{
		"cpp", "c++",
		"csharp", "c#",
		"java",
		"javascript", "js",
		"kotlin",
		"php",
		"python", "py",
		"ruby", "rb",
		"rust", "rs",
		"swift",
		"typescript", "ts",
	}

	for _, lang := range stubLanguages {
		p := &stubProcessor{language: lang}
		processor.Register(p)
	}
}

type stubProcessor struct {
	language string
}

func (p *stubProcessor) Language() string {
	return p.language
}

func (p *stubProcessor) Version() string {
	return "0.0.0-nocgo"
}

func (p *stubProcessor) SupportedExtensions() []string {
	switch p.language {
	case "python", "py":
		return []string{".py", ".pyi"}
	case "javascript", "js":
		return []string{".js", ".jsx", ".mjs", ".cjs"}
	case "typescript", "ts":
		return []string{".ts", ".tsx"}
	case "go":
		return []string{".go"}
	case "java":
		return []string{".java"}
	case "cpp", "c++":
		return []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx", ".c", ".h"}
	case "csharp", "c#":
		return []string{".cs"}
	case "ruby", "rb":
		return []string{".rb"}
	case "php":
		return []string{".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".phps"}
	case "swift":
		return []string{".swift"}
	case "kotlin":
		return []string{".kt", ".kts"}
	case "rust", "rs":
		return []string{".rs"}
	default:
		return []string{}
	}
}

func (p *stubProcessor) CanProcess(filename string) bool {
	return false // Always return false when CGO is disabled
}

func (p *stubProcessor) Process(ctx context.Context, reader io.Reader, filename string) (*ir.DistilledFile, error) {
	return nil, fmt.Errorf("%s parser requires CGO to be enabled", p.language)
}

func (p *stubProcessor) ProcessWithOptions(ctx context.Context, reader io.Reader, filename string, opts processor.ProcessOptions) (*ir.DistilledFile, error) {
	return nil, fmt.Errorf("%s parser requires CGO to be enabled", p.language)
}