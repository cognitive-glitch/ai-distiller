package typescript

import (
	"context"
	"regexp"
	"strings"

	"github.com/janreges/ai-distiller/internal/ir"
)

// EnhancedParser provides improved TypeScript parsing with multi-line support
type EnhancedParser struct {
	source   []byte
	filename string
	isTSX    bool
	lines    []string
}

// State for tracking context
type parserState struct {
	insideClass     bool
	insideInterface bool
	insideEnum      bool
	currentClass    *ir.DistilledClass
	currentInterface *ir.DistilledInterface
	currentEnum     *ir.DistilledEnum
	braceLevel      int
}

// Pre-compiled regexes
var (
	importRegex      = regexp.MustCompile(`^\s*import\s+(?:type\s+)?(?:\{[^}]+\}|[\w*]+)(?:\s+as\s+\w+)?\s+from\s+['"\` + "`" + `]([^'"\` + "`" + `]+)['"\` + "`" + `]`)
	typeImportRegex  = regexp.MustCompile(`^\s*import\s+type\s+`)
	interfaceRegex   = regexp.MustCompile(`^\s*(?:export\s+)?interface\s+(\w+)(?:<[^>]+>)?(?:\s+extends\s+([^{]+))?`)
	typeAliasRegex   = regexp.MustCompile(`^\s*(?:export\s+)?type\s+(\w+)(?:<[^>]+>)?\s*=\s*(.+?)(?:;|$)`)
	enumRegex        = regexp.MustCompile(`^\s*(?:export\s+)?enum\s+(\w+)`)
	classRegex       = regexp.MustCompile(`^\s*(?:export\s+)?(?:abstract\s+)?class\s+(\w+)(?:<[^>]+>)?(?:\s+extends\s+([^{]+?))?(?:\s+implements\s+([^{]+))?`)
	functionRegex    = regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+(\*?)(\w+)(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*:\s*([^{;]+))?`)
	methodRegex      = regexp.MustCompile(`^\s*(public|private|protected|static|readonly|async|get|set)?\s*(\*?)(\w+)\s*(?:<[^>]+>)?\s*\(([^)]*)\)(?:\s*:\s*([^{;]+))?`)
	propertyRegex    = regexp.MustCompile(`^\s*(public|private|protected|static|readonly)?\s*(\w+)(?:\?)?(?:\s*:\s*([^=;]+))?(?:\s*=\s*([^;]+))?`)
	constRegex       = regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)(?:\s*:\s*([^=;]+))?(?:\s*=\s*(.+?))?(?:;|$)`)
	enumMemberRegex  = regexp.MustCompile(`^\s*(\w+)(?:\s*=\s*([^,}]+))?`)
)

// NewEnhancedParser creates a new enhanced parser
func NewEnhancedParser() *EnhancedParser {
	return &EnhancedParser{}
}

// ProcessSource processes TypeScript source code
func (p *EnhancedParser) ProcessSource(ctx context.Context, source []byte, filename string, isTSX bool) (*ir.DistilledFile, error) {
	p.source = source
	p.filename = filename
	p.isTSX = isTSX
	
	// Split into lines for multi-line processing
	p.lines = strings.Split(string(source), "\n")
	
	file := &ir.DistilledFile{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: 1,
				EndLine:   len(p.lines),
			},
		},
		Path:     filename,
		Language: "typescript",
		Version:  "5.0",
		Children: []ir.DistilledNode{},
		Errors:   []ir.DistilledError{},
	}
	
	// Process the file with state tracking
	state := &parserState{}
	
	for i := 0; i < len(p.lines); i++ {
		line := p.lines[i]
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines and single-line comments
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		
		// Track brace level
		state.braceLevel += strings.Count(line, "{") - strings.Count(line, "}")
		
		// Handle state transitions
		if state.insideClass && state.braceLevel == 0 {
			state.insideClass = false
			state.currentClass = nil
		}
		if state.insideInterface && state.braceLevel == 0 {
			state.insideInterface = false
			state.currentInterface = nil
		}
		if state.insideEnum && state.braceLevel == 0 {
			state.insideEnum = false
			state.currentEnum = nil
		}
		
		// Process based on current state
		if state.insideClass {
			p.processClassMember(trimmed, i+1, state)
		} else if state.insideInterface {
			p.processInterfaceMember(trimmed, i+1, state)
		} else if state.insideEnum {
			p.processEnumMember(trimmed, i+1, state)
		} else {
			// Top-level declarations
			if importRegex.MatchString(trimmed) {
				p.processImport(trimmed, i+1, file)
			} else if interfaceRegex.MatchString(trimmed) {
				intf := p.processInterface(trimmed, i+1, file)
				state.insideInterface = true
				state.currentInterface = intf
				state.braceLevel = 1
			} else if typeAliasRegex.MatchString(trimmed) {
				p.processTypeAlias(trimmed, i+1, file)
			} else if enumRegex.MatchString(trimmed) {
				enum := p.processEnum(trimmed, i+1, file)
				state.insideEnum = true
				state.currentEnum = enum
				state.braceLevel = 1
			} else if classRegex.MatchString(trimmed) {
				class := p.processClass(trimmed, i+1, file)
				state.insideClass = true
				state.currentClass = class
				state.braceLevel = 1
			} else if functionRegex.MatchString(trimmed) {
				p.processFunction(trimmed, i+1, file)
			} else if constRegex.MatchString(trimmed) {
				p.processVariable(trimmed, i+1, file)
			}
		}
	}
	
	return file, nil
}

// processImport processes import statements
func (p *EnhancedParser) processImport(line string, lineNum int, file *ir.DistilledFile) {
	imp := &ir.DistilledImport{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		ImportType: "import",
		Symbols:    []ir.ImportedSymbol{},
	}
	
	// Check for type-only import
	if typeImportRegex.MatchString(line) {
		imp.IsType = true
	}
	
	// Extract module
	if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
		imp.Module = matches[1]
	}
	
	// Simple symbol extraction (not perfect but better)
	importPart := line
	if idx := strings.Index(line, " from "); idx > 0 {
		importPart = line[:idx]
	}
	importPart = strings.TrimPrefix(importPart, "import ")
	importPart = strings.TrimPrefix(importPart, "type ")
	
	if strings.Contains(importPart, "{") && strings.Contains(importPart, "}") {
		// Named imports
		start := strings.Index(importPart, "{")
		end := strings.Index(importPart, "}")
		if start >= 0 && end > start {
			names := importPart[start+1 : end]
			for _, name := range strings.Split(names, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					// Handle 'as' aliases
					parts := strings.Split(name, " as ")
					if len(parts) == 2 {
						imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
							Name:  strings.TrimSpace(parts[0]),
							Alias: strings.TrimSpace(parts[1]),
						})
					} else {
						imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: name})
					}
				}
			}
		}
	} else if strings.Contains(importPart, "*") {
		// Namespace import
		parts := strings.Split(importPart, " as ")
		if len(parts) == 2 {
			imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{
				Name:  "*",
				Alias: strings.TrimSpace(parts[1]),
			})
		}
	} else if importPart != "" {
		// Default import
		imp.Symbols = append(imp.Symbols, ir.ImportedSymbol{Name: strings.TrimSpace(importPart)})
	}
	
	file.Children = append(file.Children, imp)
}

// processInterface processes interface declarations
func (p *EnhancedParser) processInterface(line string, lineNum int, file *ir.DistilledFile) *ir.DistilledInterface {
	intf := &ir.DistilledInterface{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
		Children:   []ir.DistilledNode{},
	}
	
	if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
		intf.Name = matches[1]
		
		// Handle extends
		if len(matches) > 2 && matches[2] != "" {
			extends := strings.Split(matches[2], ",")
			for _, ext := range extends {
				ext = strings.TrimSpace(ext)
				if ext != "" {
					intf.Extends = append(intf.Extends, ir.TypeRef{Name: ext})
				}
			}
		}
	}
	
	file.Children = append(file.Children, intf)
	return intf
}

// processInterfaceMember processes interface members
func (p *EnhancedParser) processInterfaceMember(line string, lineNum int, state *parserState) {
	if state.currentInterface == nil || line == "}" || line == "" {
		return
	}
	
	// Try to match as property
	if strings.Contains(line, ":") && !strings.Contains(line, "(") {
		field := &ir.DistilledField{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine: lineNum,
					EndLine:   lineNum,
				},
			},
			Visibility: ir.VisibilityPublic,
		}
		
		// Simple property parsing
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			namePart := strings.TrimSpace(parts[0])
			typePart := strings.TrimSpace(parts[1])
			typePart = strings.TrimSuffix(typePart, ";")
			typePart = strings.TrimSuffix(typePart, ",")
			
			// Handle optional
			if strings.HasSuffix(namePart, "?") {
				namePart = strings.TrimSuffix(namePart, "?")
				// TODO: Mark as optional in IR
			}
			
			// Handle readonly
			if strings.HasPrefix(namePart, "readonly ") {
				namePart = strings.TrimPrefix(namePart, "readonly ")
				field.Modifiers = append(field.Modifiers, ir.ModifierReadonly)
			}
			
			field.Name = namePart
			field.Type = &ir.TypeRef{Name: typePart}
			
			state.currentInterface.Children = append(state.currentInterface.Children, field)
		}
	} else if strings.Contains(line, "(") {
		// Method signature
		fn := &ir.DistilledFunction{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine: lineNum,
					EndLine:   lineNum,
				},
			},
			Visibility: ir.VisibilityPublic,
			Parameters: []ir.Parameter{},
		}
		
		// Simple method parsing
		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 3 {
			fn.Name = matches[3]
			
			// Return type
			if len(matches) > 5 && matches[5] != "" {
				returnType := strings.TrimSpace(matches[5])
				returnType = strings.TrimSuffix(returnType, ";")
				fn.Returns = &ir.TypeRef{Name: returnType}
			}
		}
		
		state.currentInterface.Children = append(state.currentInterface.Children, fn)
	}
}

// processTypeAlias processes type alias declarations
func (p *EnhancedParser) processTypeAlias(line string, lineNum int, file *ir.DistilledFile) {
	alias := &ir.DistilledTypeAlias{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
	}
	
	if matches := typeAliasRegex.FindStringSubmatch(line); len(matches) > 2 {
		alias.Name = matches[1]
		alias.Type = ir.TypeRef{Name: strings.TrimSpace(matches[2])}
	}
	
	file.Children = append(file.Children, alias)
}

// processEnum processes enum declarations
func (p *EnhancedParser) processEnum(line string, lineNum int, file *ir.DistilledFile) *ir.DistilledEnum {
	enum := &ir.DistilledEnum{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
		Children:   []ir.DistilledNode{},
	}
	
	if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
		enum.Name = matches[1]
	}
	
	file.Children = append(file.Children, enum)
	return enum
}

// processEnumMember processes enum members
func (p *EnhancedParser) processEnumMember(line string, lineNum int, state *parserState) {
	if state.currentEnum == nil || line == "}" || line == "" {
		return
	}
	
	if matches := enumMemberRegex.FindStringSubmatch(line); len(matches) > 1 {
		field := &ir.DistilledField{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine: lineNum,
					EndLine:   lineNum,
				},
			},
			Name:       matches[1],
			Visibility: ir.VisibilityPublic,
			Modifiers:  []ir.Modifier{ir.ModifierStatic, ir.ModifierReadonly},
		}
		
		if len(matches) > 2 && matches[2] != "" {
			field.DefaultValue = strings.TrimSpace(matches[2])
		}
		
		state.currentEnum.Children = append(state.currentEnum.Children, field)
	}
}

// processClass processes class declarations
func (p *EnhancedParser) processClass(line string, lineNum int, file *ir.DistilledFile) *ir.DistilledClass {
	class := &ir.DistilledClass{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
		Children:   []ir.DistilledNode{},
	}
	
	if strings.Contains(line, "abstract ") {
		class.Modifiers = append(class.Modifiers, ir.ModifierAbstract)
	}
	
	if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
		class.Name = matches[1]
		
		// Handle extends
		if len(matches) > 2 && matches[2] != "" {
			class.Extends = append(class.Extends, ir.TypeRef{Name: strings.TrimSpace(matches[2])})
		}
		
		// Handle implements
		if len(matches) > 3 && matches[3] != "" {
			implements := strings.Split(matches[3], ",")
			for _, impl := range implements {
				impl = strings.TrimSpace(impl)
				if impl != "" {
					class.Implements = append(class.Implements, ir.TypeRef{Name: impl})
				}
			}
		}
	}
	
	file.Children = append(file.Children, class)
	return class
}

// processClassMember processes class members
func (p *EnhancedParser) processClassMember(line string, lineNum int, state *parserState) {
	if state.currentClass == nil || line == "}" || line == "" {
		return
	}
	
	// Skip constructor for now (too complex)
	if strings.Contains(line, "constructor(") {
		return
	}
	
	// Try to match as method
	if strings.Contains(line, "(") {
		fn := &ir.DistilledFunction{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine: lineNum,
					EndLine:   lineNum,
				},
			},
			Visibility: ir.VisibilityPublic,
			Modifiers:  []ir.Modifier{},
			Parameters: []ir.Parameter{},
		}
		
		if matches := methodRegex.FindStringSubmatch(line); len(matches) > 3 {
			// Handle modifiers
			if matches[1] != "" {
				switch matches[1] {
				case "private":
					fn.Visibility = ir.VisibilityPrivate
				case "protected":
					fn.Visibility = ir.VisibilityProtected
				case "static":
					fn.Modifiers = append(fn.Modifiers, ir.ModifierStatic)
				case "async":
					fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
				}
			}
			
			// Handle getter/setter
			if matches[1] == "get" || matches[1] == "set" {
				fn.Name = matches[1] + " " + matches[3]
			} else {
				fn.Name = matches[3]
			}
			
			// Return type
			if len(matches) > 5 && matches[5] != "" {
				returnType := strings.TrimSpace(matches[5])
				returnType = strings.TrimSuffix(returnType, " {")
				fn.Returns = &ir.TypeRef{Name: returnType}
			}
		}
		
		state.currentClass.Children = append(state.currentClass.Children, fn)
	} else {
		// Try as property
		field := &ir.DistilledField{
			BaseNode: ir.BaseNode{
				Location: ir.Location{
					StartLine: lineNum,
					EndLine:   lineNum,
				},
			},
			Visibility: ir.VisibilityPublic,
			Modifiers:  []ir.Modifier{},
		}
		
		if matches := propertyRegex.FindStringSubmatch(line); len(matches) > 2 {
			// Handle modifiers
			if matches[1] != "" {
				switch matches[1] {
				case "private":
					field.Visibility = ir.VisibilityPrivate
				case "protected":
					field.Visibility = ir.VisibilityProtected
				case "static":
					field.Modifiers = append(field.Modifiers, ir.ModifierStatic)
				case "readonly":
					field.Modifiers = append(field.Modifiers, ir.ModifierReadonly)
				}
			}
			
			field.Name = matches[2]
			
			// Type
			if len(matches) > 3 && matches[3] != "" {
				field.Type = &ir.TypeRef{Name: strings.TrimSpace(matches[3])}
			}
			
			// Default value
			if len(matches) > 4 && matches[4] != "" {
				field.DefaultValue = strings.TrimSpace(matches[4])
			}
		}
		
		if field.Name != "" {
			state.currentClass.Children = append(state.currentClass.Children, field)
		}
	}
}

// processFunction processes function declarations
func (p *EnhancedParser) processFunction(line string, lineNum int, file *ir.DistilledFile) {
	fn := &ir.DistilledFunction{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
		Parameters: []ir.Parameter{},
	}
	
	if matches := functionRegex.FindStringSubmatch(line); len(matches) > 2 {
		// Handle generator
		if matches[1] == "*" {
			fn.Name = "*" + matches[2]
		} else {
			fn.Name = matches[2]
		}
		
		// Handle async
		if strings.Contains(line, "async ") {
			fn.Modifiers = append(fn.Modifiers, ir.ModifierAsync)
		}
		
		// Simple parameter parsing (just count for now)
		if len(matches) > 3 && matches[3] != "" {
			params := strings.Split(matches[3], ",")
			for _, param := range params {
				param = strings.TrimSpace(param)
				if param != "" {
					p := ir.Parameter{}
					// Very simple parsing
					parts := strings.Split(param, ":")
					if len(parts) >= 1 {
						p.Name = strings.TrimSpace(parts[0])
						p.Name = strings.TrimSuffix(p.Name, "?")
					}
					if len(parts) >= 2 {
						p.Type = ir.TypeRef{Name: strings.TrimSpace(strings.Join(parts[1:], ":"))}
					}
					fn.Parameters = append(fn.Parameters, p)
				}
			}
		}
		
		// Return type
		if len(matches) > 4 && matches[4] != "" {
			fn.Returns = &ir.TypeRef{Name: strings.TrimSpace(matches[4])}
		}
	}
	
	file.Children = append(file.Children, fn)
}

// processVariable processes variable declarations
func (p *EnhancedParser) processVariable(line string, lineNum int, file *ir.DistilledFile) {
	field := &ir.DistilledField{
		BaseNode: ir.BaseNode{
			Location: ir.Location{
				StartLine: lineNum,
				EndLine:   lineNum,
			},
		},
		Visibility: ir.VisibilityPublic,
		Modifiers:  []ir.Modifier{},
	}
	
	if strings.HasPrefix(line, "const ") || strings.Contains(line, " const ") {
		field.Modifiers = append(field.Modifiers, ir.ModifierFinal)
	}
	
	if matches := constRegex.FindStringSubmatch(line); len(matches) > 1 {
		field.Name = matches[1]
		
		// Type
		if len(matches) > 2 && matches[2] != "" {
			field.Type = &ir.TypeRef{Name: strings.TrimSpace(matches[2])}
		}
		
		// Default value
		if len(matches) > 3 && matches[3] != "" {
			field.DefaultValue = strings.TrimSpace(matches[3])
		}
	}
	
	file.Children = append(file.Children, field)
}