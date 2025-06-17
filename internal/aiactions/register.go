package aiactions

import (
	"github.com/janreges/ai-distiller/internal/ai"
)

// Register registers all built-in actions to the provided registry
func Register(registry *ai.ActionRegistry) {
	// Flow actions (complex, multi-file output)
	registry.Register(&DeepAnalysisFlowAction{})
	
	// Prompt actions (simple, content generation)
	registry.Register(&RefactoringPromptAction{})
	registry.Register(&ComplexCodebasePromptAction{})
	registry.Register(&SecurityPromptAction{})
	registry.Register(&PerformancePromptAction{})
	
	// TODO: Register other actions as they are implemented:
	// - prompt-for-best-practices-analysis
	// - prompt-for-bug-hunting
	// - prompt-for-single-file-docs
	// - flow-for-multi-file-docs
}

