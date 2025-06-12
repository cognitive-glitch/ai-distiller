package python

import (
	"fmt"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// ParseErrorKind represents types of parsing errors
type ParseErrorKind string

const (
	ErrorKindSyntax       ParseErrorKind = "syntax"
	ErrorKindIndentation  ParseErrorKind = "indentation"
	ErrorKindUnclosedExpr ParseErrorKind = "unclosed_expression"
	ErrorKindInvalidName  ParseErrorKind = "invalid_name"
	ErrorKindIncomplete   ParseErrorKind = "incomplete_definition"
)

// ParseError represents a parsing error with recovery information
type ParseError struct {
	Line     int
	Column   int
	Message  string
	Kind     ParseErrorKind
	Severity string // "error", "warning", "info"
}

// ErrorCollector collects errors during parsing
type ErrorCollector struct {
	errors []ParseError
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]ParseError, 0),
	}
}

// AddError adds an error to the collector
func (ec *ErrorCollector) AddError(line, column int, message string, kind ParseErrorKind) {
	ec.errors = append(ec.errors, ParseError{
		Line:     line,
		Column:   column,
		Message:  message,
		Kind:     kind,
		Severity: "error",
	})
}

// AddWarning adds a warning to the collector
func (ec *ErrorCollector) AddWarning(line, column int, message string, kind ParseErrorKind) {
	ec.errors = append(ec.errors, ParseError{
		Line:     line,
		Column:   column,
		Message:  message,
		Kind:     kind,
		Severity: "warning",
	})
}

// ToDistilledErrors converts collected errors to IR errors
func (ec *ErrorCollector) ToDistilledErrors() []ir.DistilledError {
	result := make([]ir.DistilledError, 0, len(ec.errors))
	for _, err := range ec.errors {
		result = append(result, ir.DistilledError{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine:   err.Line,
					StartColumn: err.Column,
					EndLine:     err.Line,
					EndColumn:   err.Column,
				},
			},
			Message:  err.Message,
			Severity: err.Severity,
			Code:     string(err.Kind),
		})
	}
	return result
}

// validatePythonName checks if a name is valid Python identifier
func validatePythonName(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}

	// Check Python keywords
	keywords := []string{
		"False", "None", "True", "and", "as", "assert", "async", "await",
		"break", "class", "continue", "def", "del", "elif", "else", "except",
		"finally", "for", "from", "global", "if", "import", "in", "is",
		"lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try",
		"while", "with", "yield",
	}
	
	for _, kw := range keywords {
		if name == kw {
			return fmt.Errorf("'%s' is a Python keyword", name)
		}
	}

	// Basic identifier validation (simplified)
	if !isValidIdentifierStart(rune(name[0])) {
		return fmt.Errorf("invalid identifier: must start with letter or underscore")
	}

	return nil
}

// isValidIdentifierStart checks if a rune can start a Python identifier
func isValidIdentifierStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r > 127
}

// findUnclosedParenthesis finds unclosed parentheses in a line
func findUnclosedParenthesis(line string) (int, bool) {
	depth := 0
	inString := false
	stringChar := rune(0)
	
	for i, ch := range line {
		if !inString {
			if ch == '"' || ch == '\'' {
				inString = true
				stringChar = ch
			} else if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				if depth < 0 {
					return i, true // Too many closing parens
				}
			}
		} else if ch == stringChar && (i == 0 || line[i-1] != '\\') {
			inString = false
		}
	}
	
	return -1, depth > 0
}

// detectIndentationError checks for indentation errors
func detectIndentationError(lines []string, lineNum int) *ParseError {
	if lineNum >= len(lines) {
		return nil
	}
	
	line := lines[lineNum]
	if strings.TrimSpace(line) == "" {
		return nil
	}
	
	// Count leading spaces/tabs
	spaces := 0
	tabs := 0
	for _, ch := range line {
		if ch == ' ' {
			spaces++
		} else if ch == '\t' {
			tabs++
		} else {
			break
		}
	}
	
	// Mixed indentation
	if spaces > 0 && tabs > 0 {
		return &ParseError{
			Line:     lineNum + 1,
			Column:   1,
			Message:  "mixed spaces and tabs in indentation",
			Kind:     ErrorKindIndentation,
			Severity: "error",
		}
	}
	
	// Check if indentation is multiple of 4 (common Python convention)
	if spaces > 0 && spaces%4 != 0 {
		return &ParseError{
			Line:     lineNum + 1,
			Column:   1,
			Message:  fmt.Sprintf("indentation is not a multiple of 4 spaces (%d spaces)", spaces),
			Kind:     ErrorKindIndentation,
			Severity: "warning",
		}
	}
	
	return nil
}

// tryRecoverFromError attempts to recover from a parsing error
func tryRecoverFromError(lines []string, startIdx int, errorKind ParseErrorKind) int {
	switch errorKind {
	case ErrorKindUnclosedExpr:
		// Look for the closing parenthesis on subsequent lines
		depth := 1
		for i := startIdx + 1; i < len(lines); i++ {
			line := lines[i]
			for _, ch := range line {
				if ch == '(' {
					depth++
				} else if ch == ')' {
					depth--
					if depth == 0 {
						return i + 1
					}
				}
			}
		}
		
	case ErrorKindIncomplete:
		// For incomplete definitions, skip to next non-indented line
		if startIdx >= len(lines) {
			return len(lines)
		}
		
		baseIndent := len(lines[startIdx]) - len(strings.TrimLeft(lines[startIdx], " \t"))
		for i := startIdx + 1; i < len(lines); i++ {
			line := lines[i]
			if strings.TrimSpace(line) == "" {
				continue
			}
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			if currentIndent <= baseIndent {
				return i
			}
		}
	}
	
	// Default: skip to next line
	return startIdx + 1
}