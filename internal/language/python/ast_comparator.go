package python

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/janreges/ai-distiller/internal/ir"
)

// ASTComparator compares our parsed output with Python's ast module
type ASTComparator struct {
	pythonPath string
}

// NewASTComparator creates a new AST comparator
func NewASTComparator() *ASTComparator {
	return &ASTComparator{
		pythonPath: "python3", // Can be configured
	}
}

// PythonAST represents the structure returned by Python's ast module
type PythonAST struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name,omitempty"`
	Functions  []PythonFunction       `json:"functions"`
	Classes    []PythonClass          `json:"classes"`
	Imports    []PythonImport         `json:"imports"`
	Errors     []string               `json:"errors"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type PythonFunction struct {
	Name       string   `json:"name"`
	Args       []string `json:"args"`
	LineNumber int      `json:"lineno"`
	IsAsync    bool     `json:"is_async"`
	Decorators []string `json:"decorators"`
}

type PythonClass struct {
	Name       string           `json:"name"`
	Bases      []string         `json:"bases"`
	Methods    []PythonFunction `json:"methods"`
	LineNumber int              `json:"lineno"`
}

type PythonImport struct {
	Module string   `json:"module"`
	Names  []string `json:"names"`
	IsFrom bool     `json:"is_from"`
	Level  int      `json:"level"` // For relative imports
}

// CompareResult holds the comparison results
type CompareResult struct {
	Match           bool
	MissingInOurs   []string
	MissingInPython []string
	Differences     []string
	OurNodeCount    int
	PythonNodeCount int
}

// CompareFile compares a file parsed by our parser with Python's ast
func (c *ASTComparator) CompareFile(filename string, ourAST *ir.DistilledFile) (*CompareResult, error) {
	// Get Python AST
	pythonAST, err := c.getPythonAST(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get Python AST: %w", err)
	}

	// Compare structures
	result := &CompareResult{
		Match:           true,
		MissingInOurs:   []string{},
		MissingInPython: []string{},
		Differences:     []string{},
	}

	// Build maps for comparison
	ourFuncs := make(map[string]*ir.DistilledFunction)
	ourClasses := make(map[string]*ir.DistilledClass)

	for _, node := range ourAST.Children {
		switch n := node.(type) {
		case *ir.DistilledFunction:
			ourFuncs[n.Name] = n
			result.OurNodeCount++
		case *ir.DistilledClass:
			ourClasses[n.Name] = n
			result.OurNodeCount++
		}
	}

	// Compare functions
	for _, pyFunc := range pythonAST.Functions {
		result.PythonNodeCount++
		if ourFunc, exists := ourFuncs[pyFunc.Name]; exists {
			// Compare details
			if len(ourFunc.Parameters) != len(pyFunc.Args) {
				result.Differences = append(result.Differences,
					fmt.Sprintf("Function %s: parameter count mismatch (ours: %d, python: %d)",
						pyFunc.Name, len(ourFunc.Parameters), len(pyFunc.Args)))
			}

			// Check async modifier
			isOurAsync := false
			for _, mod := range ourFunc.Modifiers {
				if mod == ir.ModifierAsync {
					isOurAsync = true
					break
				}
			}
			if isOurAsync != pyFunc.IsAsync {
				result.Differences = append(result.Differences,
					fmt.Sprintf("Function %s: async mismatch", pyFunc.Name))
			}
		} else {
			result.MissingInOurs = append(result.MissingInOurs,
				fmt.Sprintf("function: %s", pyFunc.Name))
			result.Match = false
		}
	}

	// Compare classes
	for _, pyClass := range pythonAST.Classes {
		result.PythonNodeCount++
		if ourClass, exists := ourClasses[pyClass.Name]; exists {
			// Compare base classes
			if len(ourClass.Extends) != len(pyClass.Bases) {
				result.Differences = append(result.Differences,
					fmt.Sprintf("Class %s: base class count mismatch (ours: %d, python: %d)",
						pyClass.Name, len(ourClass.Extends), len(pyClass.Bases)))
			}

			// Count methods
			ourMethodCount := 0
			for _, child := range ourClass.Children {
				if _, ok := child.(*ir.DistilledFunction); ok {
					ourMethodCount++
				}
			}
			if ourMethodCount != len(pyClass.Methods) {
				result.Differences = append(result.Differences,
					fmt.Sprintf("Class %s: method count mismatch (ours: %d, python: %d)",
						pyClass.Name, ourMethodCount, len(pyClass.Methods)))
			}
		} else {
			result.MissingInOurs = append(result.MissingInOurs,
				fmt.Sprintf("class: %s", pyClass.Name))
			result.Match = false
		}
	}

	// Check for items we found that Python didn't
	for name := range ourFuncs {
		found := false
		for _, pyFunc := range pythonAST.Functions {
			if pyFunc.Name == name {
				found = true
				break
			}
		}
		if !found {
			result.MissingInPython = append(result.MissingInPython,
				fmt.Sprintf("function: %s", name))
			result.Match = false
		}
	}

	for name := range ourClasses {
		found := false
		for _, pyClass := range pythonAST.Classes {
			if pyClass.Name == name {
				found = true
				break
			}
		}
		if !found {
			result.MissingInPython = append(result.MissingInPython,
				fmt.Sprintf("class: %s", name))
			result.Match = false
		}
	}

	if len(result.Differences) > 0 {
		result.Match = false
	}

	return result, nil
}

// getPythonAST runs a Python script to parse the file using ast module
func (c *ASTComparator) getPythonAST(filename string) (*PythonAST, error) {
	// Python script to extract AST information
	pythonScript := `
import ast
import json
import sys

def extract_ast(filename):
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            source = f.read()

        tree = ast.parse(source, filename)

        result = {
            "type": "module",
            "functions": [],
            "classes": [],
            "imports": [],
            "errors": []
        }

        for node in tree.body:
            if isinstance(node, ast.FunctionDef) or isinstance(node, ast.AsyncFunctionDef):
                func_info = {
                    "name": node.name,
                    "args": [arg.arg for arg in node.args.args],
                    "lineno": node.lineno,
                    "is_async": isinstance(node, ast.AsyncFunctionDef),
                    "decorators": [d.id if isinstance(d, ast.Name) else str(d) for d in node.decorator_list]
                }
                result["functions"].append(func_info)

            elif isinstance(node, ast.ClassDef):
                class_info = {
                    "name": node.name,
                    "bases": [],
                    "methods": [],
                    "lineno": node.lineno
                }

                # Extract base class names
                for base in node.bases:
                    if isinstance(base, ast.Name):
                        class_info["bases"].append(base.id)
                    elif isinstance(base, ast.Attribute):
                        class_info["bases"].append(f"{base.value.id if isinstance(base.value, ast.Name) else '?'}.{base.attr}")

                # Extract methods
                for item in node.body:
                    if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
                        method_info = {
                            "name": item.name,
                            "args": [arg.arg for arg in item.args.args],
                            "lineno": item.lineno,
                            "is_async": isinstance(item, ast.AsyncFunctionDef),
                            "decorators": [d.id if isinstance(d, ast.Name) else str(d) for d in item.decorator_list]
                        }
                        class_info["methods"].append(method_info)

                result["classes"].append(class_info)

            elif isinstance(node, ast.Import):
                for alias in node.names:
                    result["imports"].append({
                        "module": alias.name,
                        "names": [],
                        "is_from": False,
                        "level": 0
                    })

            elif isinstance(node, ast.ImportFrom):
                names = [alias.name for alias in node.names] if node.names else []
                result["imports"].append({
                    "module": node.module or "",
                    "names": names,
                    "is_from": True,
                    "level": node.level or 0
                })

        print(json.dumps(result, indent=2))

    except SyntaxError as e:
        result = {
            "type": "module",
            "functions": [],
            "classes": [],
            "imports": [],
            "errors": [f"SyntaxError: {e}"]
        }
        print(json.dumps(result, indent=2))
    except Exception as e:
        result = {
            "type": "module",
            "functions": [],
            "classes": [],
            "imports": [],
            "errors": [f"Error: {e}"]
        }
        print(json.dumps(result, indent=2))

if __name__ == "__main__":
    if len(sys.argv) > 1:
        extract_ast(sys.argv[1])
    else:
        print(json.dumps({"errors": ["No filename provided"]}, indent=2))
`

	// Run Python script
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	cmd := exec.Command(c.pythonPath, "-c", pythonScript, absPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run Python: %w, stderr: %s", err, stderr.String())
	}

	// Parse JSON output
	var result PythonAST
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Python output: %w, output: %s", err, out.String())
	}

	return &result, nil
}
