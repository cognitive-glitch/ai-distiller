package aiactions

import (
	"github.com/janreges/ai-distiller/internal/ai"
)

// Register registers all built-in actions to the provided registry
func Register(registry *ai.ActionRegistry) {
	// Flow actions (complex, multi-file output)
	registry.Register(&DeepAnalysisFlowAction{})
	registry.Register(&MultiFileDocsFlowAction{})

	// Template-based prompt actions (using external markdown templates)
	registry.Register(&TemplateRefactoringPromptAction{})
	registry.Register(&TemplateComplexCodebasePromptAction{})
	registry.Register(&TemplateSecurityPromptAction{})
	registry.Register(&TemplatePerformancePromptAction{})
	registry.Register(&TemplateBestPracticesPromptAction{})
	registry.Register(&TemplateBugHuntingPromptAction{})
	registry.Register(&TemplateSingleFileDocsPromptAction{})
	registry.Register(&TemplateDiagramsPromptAction{})

	// Original embedded prompt actions (kept as fallback)
	// registry.Register(&RefactoringPromptAction{})
	// registry.Register(&ComplexCodebasePromptAction{})
	// registry.Register(&SecurityPromptAction{})
	// registry.Register(&PerformancePromptAction{})
}
