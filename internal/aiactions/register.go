package aiactions

import (
	"github.com/janreges/ai-distiller/internal/ai"
)

// Register registers all built-in actions to the provided registry
func Register(registry *ai.ActionRegistry) {
	// Flow actions (complex, multi-file output)
	_ = registry.Register(&DeepAnalysisFlowAction{})
	_ = registry.Register(&MultiFileDocsFlowAction{})

	// Template-based prompt actions (using external markdown templates)
	_ = registry.Register(&TemplateRefactoringPromptAction{})
	_ = registry.Register(&TemplateComplexCodebasePromptAction{})
	_ = registry.Register(&TemplateSecurityPromptAction{})
	_ = registry.Register(&TemplatePerformancePromptAction{})
	_ = registry.Register(&TemplateBestPracticesPromptAction{})
	_ = registry.Register(&TemplateBugHuntingPromptAction{})
	_ = registry.Register(&TemplateSingleFileDocsPromptAction{})
	_ = registry.Register(&TemplateDiagramsPromptAction{})

	// Original embedded prompt actions (kept as fallback)
	// registry.Register(&RefactoringPromptAction{})
	// registry.Register(&ComplexCodebasePromptAction{})
	// registry.Register(&SecurityPromptAction{})
	// registry.Register(&PerformancePromptAction{})
}
