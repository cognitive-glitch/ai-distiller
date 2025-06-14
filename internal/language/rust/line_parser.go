package rust

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// LineParser implements a simple line-based parser for Rust
type LineParser struct {
	source   []byte
	filename string
	lines    []string
	
	// Parser state
	currentLine      int
	insideComment    bool
	insideImpl       bool
	currentClass     *ir.DistilledClass
	currentIndent    int
}

var (
	// Regular expressions for Rust constructs
	useRe        = regexp.MustCompile(`^\s*(?:pub\s+)?use\s+(.+);`)
	modRe        = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)mod\s+(\w+)`)
	structRe     = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)struct\s+(\w+)(?:<[^>]+>)?`)
	enumRe       = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)enum\s+(\w+)(?:<[^>]+>)?`)
	traitRe      = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)trait\s+(\w+)(?:<[^>]+>)?`)
	// More robust regex that handles paths with :: and basic generics
	implRe       = regexp.MustCompile(`^\s*impl(?:<[^>]+>)?\s+(?:([\w\s:<>,]+?)\s+for\s+)?([\w\s:<>,]+)`)
	fnRe         = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)((?:async\s+)?(?:unsafe\s+)?(?:const\s+)?(?:extern\s+)?)?fn\s+(\w+)(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*->\s*([^{]+))?`)
	constRe      = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)const\s+(\w+):\s*([^=]+)(?:\s*=\s*(.+))?;`)
	staticRe     = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)static\s+(?:(mut)\s+)?(\w+):\s*([^=]+)(?:\s*=\s*(.+))?;`)
	typeRe       = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)type\s+(\w+)(?:<[^>]+>)?\s*=\s*(.+);`)
	fieldRe      = regexp.MustCompile(`^\s*((?:pub(?:\([^)]+\))?\s+)?)(\w+):\s*(.+),?$`)
	enumVariantRe = regexp.MustCompile(`^\s*(\w+)(?:\(([^)]+)\)|\{[^}]+\})?,?$`)
)

// NewLineParser creates a new line-based parser
func NewLineParser(source []byte, filename string) *LineParser {
	lines := strings.Split(string(source), "\n")
	return &LineParser{
		source:   source,
		filename: filename,
		lines:    lines,
	}
}

// Parse processes the source code and returns the IR
func (p *LineParser) Parse() *ir.DistilledFile {
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   len(p.lines),
			},
		},
		Path:     p.filename,
		Language: "rust",
		Version:  "2021",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}

	for p.currentLine < len(p.lines) {
		line := p.lines[p.currentLine]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			p.currentLine++
			continue
		}

		// Handle comments
		if strings.HasPrefix(trimmed, "//") {
			p.parseLineComment(file, nil)
			continue
		}
		if strings.HasPrefix(trimmed, "/*") {
			p.parseBlockComment(file, nil)
			continue
		}

		// Parse top-level constructs
		if matches := useRe.FindStringSubmatch(line); matches != nil {
			p.parseUse(file, matches)
		} else if matches := modRe.FindStringSubmatch(line); matches != nil {
			p.parseMod(file, matches)
		} else if matches := structRe.FindStringSubmatch(line); matches != nil {
			p.parseStruct(file, matches)
		} else if matches := enumRe.FindStringSubmatch(line); matches != nil {
			p.parseEnum(file, matches)
		} else if matches := traitRe.FindStringSubmatch(line); matches != nil {
			p.parseTrait(file, matches)
		} else if matches := implRe.FindStringSubmatch(line); matches != nil {
			p.parseImpl(file, matches)
		} else if matches := fnRe.FindStringSubmatch(line); matches != nil {
			// Only parse as top-level function if not inside another construct
			if p.currentLine == 0 || !p.isInsideBlock() {
				p.parseFunction(file, nil, matches)
			} else {
				p.currentLine++
			}
		} else if matches := constRe.FindStringSubmatch(line); matches != nil {
			p.parseConst(file, nil, matches)
		} else if matches := staticRe.FindStringSubmatch(line); matches != nil {
			p.parseStatic(file, nil, matches)
		} else if matches := typeRe.FindStringSubmatch(line); matches != nil {
			p.parseType(file, nil, matches)
		} else {
			p.currentLine++
		}
	}

	return file
}

// parseUse parses use statements
func (p *LineParser) parseUse(file *ir.DistilledFile, matches []string) {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		ImportType: "use",
		Module:     matches[1],
		Symbols:    []ir.ImportedSymbol{},
	}

	// Simple parsing of use statements
	usePath := matches[1]
	if strings.Contains(usePath, "{") && strings.Contains(usePath, "}") {
		// Handle use std::{io, fmt}; style
		if idx := strings.Index(usePath, "{"); idx > 0 {
			basePath := strings.TrimSpace(usePath[:idx])
			basePath = strings.TrimSuffix(basePath, "::")
			imp.Module = basePath
			
			// Extract symbols
			symbolsPart := usePath[idx+1 : strings.Index(usePath, "}")]
			symbols := strings.Split(symbolsPart, ",")
			for _, sym := range symbols {
				sym = strings.TrimSpace(sym)
				if sym != "" && sym != "self" {
					imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: sym})
				}
			}
		}
	} else if strings.Contains(usePath, " as ") {
		// Handle use foo as bar;
		parts := strings.Split(usePath, " as ")
		imp.Module = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  imp.Module,
				Alias: strings.TrimSpace(parts[1]),
			})
		}
	}

	file.Children = append(file.Children, imp)
	p.currentLine++
}

// parseMod parses module declarations
func (p *LineParser) parseMod(file *ir.DistilledFile, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]

	mod := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Modifiers:  []ir.Modifier{}, // Rust modules are represented as classes
		Children:   []ir.DistilledNode{},
	}

	p.currentLine++
	
	// Check if it's a module with body
	if p.currentLine < len(p.lines) {
		// Check if brace is on the same line
		currentLineText := p.lines[p.currentLine-1]
		if strings.HasSuffix(strings.TrimSpace(currentLineText), "{") {
			endLine := p.parseBlock(file, mod)
			mod.Location.EndLine = endLine
		} else {
			// Check next line
			nextLine := strings.TrimSpace(p.lines[p.currentLine])
			if nextLine == "{" {
				p.currentLine++
				endLine := p.parseBlock(file, mod)
				mod.Location.EndLine = endLine
				} else {
				mod.Location.EndLine = mod.Location.StartLine
			}
		}
	}

	file.Children = append(file.Children, mod)
}

// parseStruct parses struct declarations
func (p *LineParser) parseStruct(file *ir.DistilledFile, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]

	strct := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Modifiers:  []ir.Modifier{ir.ModifierStruct},
		Children:   []ir.DistilledNode{},
	}

	p.currentLine++
	
	// Check for struct body
	if p.currentLine < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.currentLine])
		if line == "{" || strings.HasSuffix(strings.TrimSpace(p.lines[p.currentLine-1]), "{") {
			// If brace is on same line or next line
			if line == "{" {
				p.currentLine++
			}
			p.parseStructFields(strct)
		} else if strings.HasPrefix(line, "(") {
			// Tuple struct
			p.parseTupleFields(strct, line)
		}
	}

	strct.Location.EndLine = p.currentLine + 1
	file.Children = append(file.Children, strct)
}

// parseEnum parses enum declarations
func (p *LineParser) parseEnum(file *ir.DistilledFile, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]

	enum := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}

	p.currentLine++
	
	// Parse enum variants
	if p.currentLine < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.currentLine])
		if line == "{" || strings.HasSuffix(strings.TrimSpace(p.lines[p.currentLine-1]), "{") {
			if line == "{" {
				p.currentLine++
			}
			p.parseEnumVariants(enum)
		}
	}

	enum.Location.EndLine = p.currentLine + 1
	file.Children = append(file.Children, enum)
}

// parseTrait parses trait declarations
func (p *LineParser) parseTrait(file *ir.DistilledFile, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]

	trait := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Modifiers:  []ir.Modifier{ir.ModifierAbstract},
		Children:   []ir.DistilledNode{},
	}

	p.currentLine++
	
	// Parse trait body
	if p.currentLine < len(p.lines) && strings.TrimSpace(p.lines[p.currentLine]) == "{" {
		p.currentLine++
		endLine := p.parseBlock(file, trait)
		trait.Location.EndLine = endLine
	}

	file.Children = append(file.Children, trait)
}

// parseImpl parses impl blocks
func (p *LineParser) parseImpl(file *ir.DistilledFile, matches []string) {
	var implName string
	trait := strings.TrimSpace(matches[1])
	implType := strings.TrimSpace(matches[2])
	
	if trait != "" {
		// impl Trait for Type
		implName = fmt.Sprintf("impl %s for %s", trait, implType)
	} else {
		// impl Type
		implName = fmt.Sprintf("impl %s", implType)
	}

	impl := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       implName,
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}

	p.currentLine++
	p.insideImpl = true
	p.currentClass = impl
	
	// Parse impl body
	if p.currentLine < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.currentLine])
		if line == "{" || strings.HasSuffix(strings.TrimSpace(p.lines[p.currentLine-1]), "{") {
			if line == "{" {
				p.currentLine++
			}
			endLine := p.parseBlock(file, impl)
			impl.Location.EndLine = endLine
		}
	}

	p.insideImpl = false
	p.currentClass = nil
	file.Children = append(file.Children, impl)
}

// parseFunction parses function declarations
func (p *LineParser) parseFunction(file *ir.DistilledFile, parent ir.DistilledNode, matches []string) {
	visibility := p.parseVisibility(matches[1])
	modifiers := p.parseFunctionModifiers(matches[2])
	name := matches[3]
	params := matches[4]
	returnType := strings.TrimSpace(matches[5])

	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Modifiers:  modifiers,
		Parameters: p.parseParameters(params),
	}

	if returnType != "" {
		fn.Returns = &ir.TypeRef{Name: returnType}
	}

	p.currentLine++
	
	// Parse function body
	if p.currentLine < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.currentLine])
		prevLine := strings.TrimSpace(p.lines[p.currentLine-1])
		
		if strings.HasSuffix(prevLine, "{") {
			// Opening brace is on the same line as function declaration
			startLine := p.currentLine - 1
			p.currentLine-- // Go back to the line with the brace
			p.skipBlock()
			p.currentLine++ // Move past the closing brace
			fn.Location.EndLine = p.currentLine
			
			// Extract implementation
			if p.currentLine > startLine {
				var implLines []string
				for i := startLine; i <= p.currentLine && i < len(p.lines); i++ {
					implLines = append(implLines, p.lines[i])
				}
				fn.Implementation = strings.Join(implLines, "\n")
			}
		} else if line == "{" {
			// Opening brace is on the next line
			startLine := p.currentLine
			p.skipBlock()
			p.currentLine++ // Move past the closing brace
			fn.Location.EndLine = p.currentLine
			
			// Extract implementation
			if p.currentLine > startLine {
				var implLines []string
				for i := startLine; i <= p.currentLine && i < len(p.lines); i++ {
					implLines = append(implLines, p.lines[i])
				}
				fn.Implementation = strings.Join(implLines, "\n")
			}
		} else {
			// No body (e.g., trait method declaration)
			fn.Location.EndLine = p.currentLine
		}
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, fn)
		}
	} else {
		file.Children = append(file.Children, fn)
	}
}

// parseConst parses const declarations
func (p *LineParser) parseConst(file *ir.DistilledFile, parent ir.DistilledNode, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]
	typeStr := strings.TrimSpace(matches[3])
	value := strings.TrimSpace(matches[4])

	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		Name:         name,
		Visibility:   visibility,
		Modifiers:    []ir.Modifier{ir.ModifierFinal},
		Type:         &ir.TypeRef{Name: typeStr},
		DefaultValue: value,
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, field)
		}
	} else {
		file.Children = append(file.Children, field)
	}
	
	p.currentLine++
}

// parseStatic parses static declarations
func (p *LineParser) parseStatic(file *ir.DistilledFile, parent ir.DistilledNode, matches []string) {
	visibility := p.parseVisibility(matches[1])
	// isMut := matches[2] == "mut" // TODO: handle mutable statics differently if needed
	name := matches[3]
	typeStr := strings.TrimSpace(matches[4])
	value := strings.TrimSpace(matches[5])

	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		Name:         name,
		Visibility:   visibility,
		Modifiers:    []ir.Modifier{ir.ModifierStatic},
		Type:         &ir.TypeRef{Name: typeStr},
		DefaultValue: value,
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, field)
		}
	} else {
		file.Children = append(file.Children, field)
	}
	
	p.currentLine++
}

// parseType parses type alias declarations
func (p *LineParser) parseType(file *ir.DistilledFile, parent ir.DistilledNode, matches []string) {
	visibility := p.parseVisibility(matches[1])
	name := matches[2]
	typeStr := strings.TrimSpace(matches[3])

	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		Name:       name,
		Visibility: visibility,
		Type:       &ir.TypeRef{Name: typeStr},
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, field)
		}
	} else {
		file.Children = append(file.Children, field)
	}
	
	p.currentLine++
}

// parseBlock parses a block and returns the end line
// TODO: This uses simple brace counting which can be fooled by braces in strings/comments
func (p *LineParser) parseBlock(file *ir.DistilledFile, parent ir.DistilledNode) int {
	braceCount := 1
	

	for p.currentLine < len(p.lines) && braceCount > 0 {
		line := p.lines[p.currentLine]
		trimmed := strings.TrimSpace(line)
		
		// Debug all lines
		// fmt.Printf("parseBlock ALL: line %d, trimmed: %q, braceCount: %d\n", p.currentLine, trimmed, braceCount)

		// Parse nested constructs BEFORE counting braces
		if parent != nil && braceCount == 1 && trimmed != "}" {
			if matches := fnRe.FindStringSubmatch(trimmed); matches != nil {
				p.parseFunction(file, parent, matches)
				continue
			} else if matches := constRe.FindStringSubmatch(trimmed); matches != nil {
				p.parseConst(file, parent, matches)
				continue
			} else if matches := staticRe.FindStringSubmatch(trimmed); matches != nil {
				p.parseStatic(file, parent, matches)
				continue
			} else if matches := typeRe.FindStringSubmatch(trimmed); matches != nil {
				p.parseType(file, parent, matches)
				continue
			}
		}

		// Handle comments
		if strings.HasPrefix(trimmed, "//") {
			p.parseLineComment(file, parent)
			continue
		}
		if strings.HasPrefix(trimmed, "/*") {
			p.parseBlockComment(file, parent)
			continue
		}

		// Count braces AFTER trying to parse constructs
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if braceCount == 0 {
			break
		}

		p.currentLine++
	}

	if braceCount == 0 {
		p.currentLine++
	}

	return p.currentLine
}

// parseStructFields parses fields in a struct
func (p *LineParser) parseStructFields(parent *ir.DistilledClass) {
	for p.currentLine < len(p.lines) {
		line := p.lines[p.currentLine]
		trimmed := strings.TrimSpace(line)

		if trimmed == "}" {
			p.currentLine++
			break
		}

		if matches := fieldRe.FindStringSubmatch(line); matches != nil {
			visibility := p.parseVisibility(matches[1])
			name := matches[2]
			typeStr := strings.TrimSpace(strings.TrimSuffix(matches[3], ","))

			field := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: ir.Location{
						StartLine: p.currentLine + 1,
						EndLine:   p.currentLine + 1,
					},
				},
				Name:       name,
				Visibility: visibility,
				Type:       &ir.TypeRef{Name: typeStr},
			}

			parent.Children = append(parent.Children, field)
		}

		p.currentLine++
	}
}

// parseTupleFields parses tuple struct fields
func (p *LineParser) parseTupleFields(parent *ir.DistilledClass, line string) {
	// Extract types from parentheses
	if start := strings.Index(line, "("); start >= 0 {
		if end := strings.Index(line, ")"); end > start {
			fieldsStr := line[start+1 : end]
			fields := strings.Split(fieldsStr, ",")
			
			for i, fieldType := range fields {
				fieldType = strings.TrimSpace(fieldType)
				if fieldType != "" {
					field := &ir.DistilledField{
						BaseNode: ir.BaseNode{
							Location: ir.Location{
								StartLine: p.currentLine + 1,
								EndLine:   p.currentLine + 1,
							},
						},
						Name:       fmt.Sprintf("%d", i),
						Visibility: ir.VisibilityPrivate,
						Type:       &ir.TypeRef{Name: fieldType},
					}
					parent.Children = append(parent.Children, field)
				}
			}
		}
	}
	p.currentLine++
}

// parseEnumVariants parses enum variants
func (p *LineParser) parseEnumVariants(parent *ir.DistilledClass) {
	for p.currentLine < len(p.lines) {
		line := p.lines[p.currentLine]
		trimmed := strings.TrimSpace(line)

		if trimmed == "}" {
			p.currentLine++
			break
		}

		if matches := enumVariantRe.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			variant := &ir.DistilledField{
				BaseNode: ir.BaseNode{
					Location: ir.Location{
						StartLine: p.currentLine + 1,
						EndLine:   p.currentLine + 1,
					},
				},
				Name:       name,
				Visibility: ir.VisibilityPublic,
			}

			// Handle tuple or struct variants
			if matches[2] != "" {
				variant.Type = &ir.TypeRef{Name: fmt.Sprintf("(%s)", matches[2])}
			}

			parent.Children = append(parent.Children, variant)
		}

		p.currentLine++
	}
}

// parseLineComment parses line comments
func (p *LineParser) parseLineComment(file *ir.DistilledFile, parent ir.DistilledNode) {
	line := p.lines[p.currentLine]
	text := strings.TrimPrefix(strings.TrimSpace(line), "//")
	
	format := "line"
	if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "!") {
		format = "doc"
		text = strings.TrimPrefix(text, "/")
		text = strings.TrimPrefix(text, "!")
	}

	comment := &ir.DistilledComment{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: p.currentLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		Text:   strings.TrimSpace(text),
		Format: format,
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, comment)
		}
	} else {
		file.Children = append(file.Children, comment)
	}

	p.currentLine++
}

// parseBlockComment parses block comments
func (p *LineParser) parseBlockComment(file *ir.DistilledFile, parent ir.DistilledNode) {
	startLine := p.currentLine
	var lines []string
	
	for p.currentLine < len(p.lines) {
		line := p.lines[p.currentLine]
		lines = append(lines, line)
		
		if strings.Contains(line, "*/") {
			break
		}
		p.currentLine++
	}

	text := strings.Join(lines, "\n")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	
	format := "block"
	if strings.HasPrefix(text, "*") || strings.HasPrefix(text, "!") {
		format = "doc"
		text = strings.TrimPrefix(text, "*")
		text = strings.TrimPrefix(text, "!")
	}

	comment := &ir.DistilledComment{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: startLine + 1,
				EndLine:   p.currentLine + 1,
			},
		},
		Text:   strings.TrimSpace(text),
		Format: format,
	}

	if parent != nil {
		if class, ok := parent.(*ir.DistilledClass); ok {
			class.Children = append(class.Children, comment)
		}
	} else {
		file.Children = append(file.Children, comment)
	}

	p.currentLine++
}

// Helper methods

func (p *LineParser) parseVisibility(vis string) ir.Visibility {
	vis = strings.TrimSpace(vis)
	
	if vis == "" {
		return ir.VisibilityPrivate
	}
	
	if strings.HasPrefix(vis, "pub(crate)") {
		return ir.VisibilityInternal
	}
	
	if strings.HasPrefix(vis, "pub(super)") || strings.HasPrefix(vis, "pub(in") {
		return ir.VisibilityInternal
	}
	
	if strings.HasPrefix(vis, "pub") {
		return ir.VisibilityPublic
	}
	
	return ir.VisibilityPrivate
}

func (p *LineParser) parseFunctionModifiers(mods string) []ir.Modifier {
	var modifiers []ir.Modifier
	
	if strings.Contains(mods, "async") {
		modifiers = append(modifiers, ir.ModifierAsync)
	}
	if strings.Contains(mods, "const") {
		modifiers = append(modifiers, ir.ModifierFinal)
	}
	
	return modifiers
}

func (p *LineParser) parseParameters(params string) []ir.Parameter {
	var parameters []ir.Parameter
	
	if params == "" {
		return parameters
	}
	
	// Simple parameter parsing
	parts := strings.Split(params, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Handle self parameters
		if strings.HasPrefix(part, "self") || strings.HasPrefix(part, "&self") || strings.HasPrefix(part, "&mut self") {
			parameters = append(parameters, ir.Parameter{
				Name: part,
			})
			continue
		}
		
		// Handle regular parameters
		if idx := strings.Index(part, ":"); idx > 0 {
			name := strings.TrimSpace(part[:idx])
			typeStr := strings.TrimSpace(part[idx+1:])
			
			// Handle mut parameters
			if strings.HasPrefix(name, "mut ") {
				name = strings.TrimPrefix(name, "mut ")
				name = "mut " + name
			}
			
			parameters = append(parameters, ir.Parameter{
				Name: name,
				Type: ir.TypeRef{Name: typeStr},
			})
		}
	}
	
	return parameters
}

// skipBlock skips over a code block by counting braces
// TODO: This block skipping logic is based on brace counting and is not robust.
// It can be fooled by braces inside strings or comments. This is a known
// limitation of the line-based parser and a key reason to move to the
// tree-sitter implementation.
func (p *LineParser) skipBlock() {
	braceCount := 1
	p.currentLine++
	
	for p.currentLine < len(p.lines) && braceCount > 0 {
		line := p.lines[p.currentLine]
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		if braceCount > 0 {
			p.currentLine++
		}
	}
}

func (p *LineParser) getIndent(line string) int {
	indent := 0
	for _, ch := range line {
		if ch == ' ' {
			indent++
		} else if ch == '\t' {
			indent += 4
		} else {
			break
		}
	}
	return indent
}

// isInsideBlock checks if we're currently inside a block by checking indentation
func (p *LineParser) isInsideBlock() bool {
	if p.currentLine >= len(p.lines) {
		return false
	}
	// Check if current line is indented
	line := p.lines[p.currentLine]
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}