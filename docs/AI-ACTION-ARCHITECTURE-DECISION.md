# AI Action Architecture Decision

## Executive Summary

After consulting with Gemini Pro and o3, I've analyzed two different architectural approaches for the AI action system. This document synthesizes their insights and proposes the final architecture.

## Analysis of Proposed Architectures

### Gemini Pro's Approach

**Strengths:**
- Clean separation of concerns with AIAction interface
- Simple distinction between PromptAction and FlowAction
- Clear ActionRegistry pattern
- Minimal complexity for simple use cases

**Limitations:**
- Rigid I/O contract (only prepend/append content)
- Binary classification might be too restrictive
- Limited support for multi-file output
- No support for actions that need different distiller options

### o3's Approach

**Strengths:**
- Pipeline-based architecture with lifecycle hooks
- Flexible Artifact abstraction for multi-file support
- Capability-driven design (no rigid subclasses)
- Support for recursive distiller runs
- Better error isolation and security considerations

**Limitations:**
- Higher complexity for simple use cases
- Requires additional dependencies (pluggy)
- May be over-engineered for current needs

## Final Architecture Decision

Based on the analysis and considering AI Distiller is written in Go (not Python), I propose a hybrid approach that takes the best from both suggestions while fitting the Go ecosystem:

### Core Design Principles

1. **Start Simple, Allow Growth** - Begin with Gemini's cleaner approach but design for o3's flexibility
2. **Go-idiomatic** - Use interfaces and composition over complex inheritance
3. **No External Dependencies** - Avoid plugin systems initially
4. **Progressive Enhancement** - Simple actions stay simple, complex ones have full power

### Architecture Components

```go
// internal/ai/action.go

type ActionType string

const (
    ActionTypePrompt ActionType = "prompt"
    ActionTypeFlow   ActionType = "flow"
)

// Core interface that all AI actions must implement
type AIAction interface {
    // Basic metadata
    Name() string
    Description() string
    Type() ActionType
    
    // Default output path template
    DefaultOutput() string
    
    // Validation
    Validate() error
}

// Extended interface for actions that generate content
type ContentAction interface {
    AIAction
    // Generate content to wrap around distilled output
    GenerateContent(ctx *ActionContext) (*ContentResult, error)
}

// Extended interface for complex workflow actions
type FlowAction interface {
    AIAction
    // Execute complex workflow, can create multiple files
    ExecuteFlow(ctx *ActionContext) (*FlowResult, error)
}

// Context provided to all actions
type ActionContext struct {
    DistilledContent string
    ProjectPath      string
    BaseName         string
    Timestamp        time.Time
    Config           *ActionConfig
    // Future: Add more context as needed
}

// Configuration for action execution
type ActionConfig struct {
    OutputPath string
    // Future: Add more config options
}

// Result for content-generating actions
type ContentResult struct {
    ContentBefore string
    ContentAfter  string
}

// Result for flow actions
type FlowResult struct {
    Files map[string]string // path -> content
    Messages []string
}
```

### Action Registry

```go
// internal/ai/registry.go

type ActionRegistry struct {
    actions map[string]AIAction
}

func NewActionRegistry() *ActionRegistry {
    return &ActionRegistry{
        actions: make(map[string]AIAction),
    }
}

func (r *ActionRegistry) Register(action AIAction) error {
    if _, exists := r.actions[action.Name()]; exists {
        return fmt.Errorf("action %s already registered", action.Name())
    }
    r.actions[action.Name()] = action
    return nil
}

func (r *ActionRegistry) Get(name string) (AIAction, error) {
    action, ok := r.actions[name]
    if !ok {
        return nil, fmt.Errorf("action %s not found", name)
    }
    return action, nil
}

func (r *ActionRegistry) List() []AIAction {
    var list []AIAction
    for _, action := range r.actions {
        list = append(list, action)
    }
    return list
}
```

### Implementation Examples

```go
// internal/ai/actions/refactoring_prompt.go

type RefactoringPromptAction struct{}

func (a *RefactoringPromptAction) Name() string {
    return "prompt-for-refactoring-suggestion"
}

func (a *RefactoringPromptAction) Description() string {
    return "Generate AI prompt for refactoring analysis"
}

func (a *RefactoringPromptAction) Type() ActionType {
    return ActionTypePrompt
}

func (a *RefactoringPromptAction) DefaultOutput() string {
    return "./.aid/REFACTORING-SUGGESTION.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *RefactoringPromptAction) Validate() error {
    return nil
}

func (a *RefactoringPromptAction) GenerateContent(ctx *ActionContext) (*ContentResult, error) {
    prompt := generateRefactoringPrompt() // Complex prompt generation
    return &ContentResult{
        ContentBefore: prompt,
        ContentAfter: "",
    }, nil
}
```

### CLI Integration

```go
// internal/cli/root.go modifications

var (
    aiAction string
    aiOutput string
)

func init() {
    rootCmd.Flags().StringVar(&aiAction, "ai-action", "", "AI action to perform on distilled output")
    rootCmd.Flags().StringVar(&aiOutput, "ai-output", "", "Output path for AI action (default: action-specific)")
}
```

## Implementation Plan

1. **Phase 1: Core Infrastructure**
   - Create `internal/ai/` package structure
   - Implement base interfaces and registry
   - Add output path template handling

2. **Phase 2: Refactor Existing**
   - Convert `--ai-analysis-task-list` to `flow-for-deep-file-to-file-analysis`
   - Maintain backward compatibility with deprecation notice

3. **Phase 3: Implement Actions**
   - Start with simple PromptActions
   - Implement complex FlowActions
   - Add comprehensive tests

4. **Phase 4: Future Extensions**
   - Add pipeline support if needed
   - Consider plugin system for external actions
   - Add capability negotiation

## Key Decisions

1. **No Python-style plugins initially** - Keep it simple with compiled-in actions
2. **Two-tier interface system** - ContentAction for simple, FlowAction for complex
3. **Explicit type declaration** - Help users understand action capabilities
4. **Progressive disclosure** - Simple actions don't see complex APIs

## Migration Path

The existing `--ai-analysis-task-list` will:
1. Continue working with deprecation warning
2. Be reimplemented as `--ai-action=flow-for-deep-file-to-file-analysis`
3. Be removed in a future major version

## Security Considerations

- All file writes are sandboxed to project directory
- Template variables are sanitized
- No arbitrary code execution in templates
- Future: Consider running actions in isolated processes

## Conclusion

This architecture balances simplicity with extensibility, fits Go idioms, and provides a clear path for future enhancements without over-engineering the initial implementation.