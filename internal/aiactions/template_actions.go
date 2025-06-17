package aiactions

import (
	"github.com/janreges/ai-distiller/internal/ai"
)

// TemplateRefactoringPromptAction generates refactoring analysis prompts using templates
type TemplateRefactoringPromptAction struct{}

// Ensure TemplateRefactoringPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateRefactoringPromptAction)(nil)

func (a *TemplateRefactoringPromptAction) Name() string {
	return "prompt-for-refactoring-suggestion"
}

func (a *TemplateRefactoringPromptAction) Description() string {
	return "Generate comprehensive refactoring analysis prompt using template system"
}

func (a *TemplateRefactoringPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateRefactoringPromptAction) DefaultOutput() string {
	return "./.aid/REFACTORING-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateRefactoringPromptAction) Validate() error {
	return nil
}

func (a *TemplateRefactoringPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("refactoring", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateSecurityPromptAction generates security analysis prompts using templates
type TemplateSecurityPromptAction struct{}

// Ensure TemplateSecurityPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateSecurityPromptAction)(nil)

func (a *TemplateSecurityPromptAction) Name() string {
	return "prompt-for-security-analysis"
}

func (a *TemplateSecurityPromptAction) Description() string {
	return "Generate comprehensive security audit prompt using template system"
}

func (a *TemplateSecurityPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateSecurityPromptAction) DefaultOutput() string {
	return "./.aid/SECURITY-AUDIT.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateSecurityPromptAction) Validate() error {
	return nil
}

func (a *TemplateSecurityPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("security", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplatePerformancePromptAction generates performance analysis prompts using templates
type TemplatePerformancePromptAction struct{}

// Ensure TemplatePerformancePromptAction implements ContentAction
var _ ai.ContentAction = (*TemplatePerformancePromptAction)(nil)

func (a *TemplatePerformancePromptAction) Name() string {
	return "prompt-for-performance-analysis"
}

func (a *TemplatePerformancePromptAction) Description() string {
	return "Generate comprehensive performance optimization prompt using template system"
}

func (a *TemplatePerformancePromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplatePerformancePromptAction) DefaultOutput() string {
	return "./.aid/PERFORMANCE-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplatePerformancePromptAction) Validate() error {
	return nil
}

func (a *TemplatePerformancePromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("performance", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateComplexCodebasePromptAction generates comprehensive codebase analysis prompts using templates
type TemplateComplexCodebasePromptAction struct{}

// Ensure TemplateComplexCodebasePromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateComplexCodebasePromptAction)(nil)

func (a *TemplateComplexCodebasePromptAction) Name() string {
	return "prompt-for-complex-codebase-analysis"
}

func (a *TemplateComplexCodebasePromptAction) Description() string {
	return "Generate comprehensive codebase analysis prompt using template system"
}

func (a *TemplateComplexCodebasePromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateComplexCodebasePromptAction) DefaultOutput() string {
	return "./.aid/COMPLEX-CODEBASE-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateComplexCodebasePromptAction) Validate() error {
	return nil
}

func (a *TemplateComplexCodebasePromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("complex-codebase", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateBestPracticesPromptAction generates best practices analysis prompts using templates
type TemplateBestPracticesPromptAction struct{}

// Ensure TemplateBestPracticesPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateBestPracticesPromptAction)(nil)

func (a *TemplateBestPracticesPromptAction) Name() string {
	return "prompt-for-best-practices-analysis"
}

func (a *TemplateBestPracticesPromptAction) Description() string {
	return "Generate comprehensive best practices analysis prompt using template system"
}

func (a *TemplateBestPracticesPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateBestPracticesPromptAction) DefaultOutput() string {
	return "./.aid/BEST-PRACTICES-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateBestPracticesPromptAction) Validate() error {
	return nil
}

func (a *TemplateBestPracticesPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("best-practices", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateBugHuntingPromptAction generates bug hunting analysis prompts using templates
type TemplateBugHuntingPromptAction struct{}

// Ensure TemplateBugHuntingPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateBugHuntingPromptAction)(nil)

func (a *TemplateBugHuntingPromptAction) Name() string {
	return "prompt-for-bug-hunting"
}

func (a *TemplateBugHuntingPromptAction) Description() string {
	return "Generate comprehensive bug hunting analysis prompt using template system"
}

func (a *TemplateBugHuntingPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateBugHuntingPromptAction) DefaultOutput() string {
	return "./.aid/BUG-HUNTING-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateBugHuntingPromptAction) Validate() error {
	return nil
}

func (a *TemplateBugHuntingPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("bug-hunting", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateSingleFileDocsPromptAction generates single file documentation prompts using templates
type TemplateSingleFileDocsPromptAction struct{}

// Ensure TemplateSingleFileDocsPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateSingleFileDocsPromptAction)(nil)

func (a *TemplateSingleFileDocsPromptAction) Name() string {
	return "prompt-for-single-file-docs"
}

func (a *TemplateSingleFileDocsPromptAction) Description() string {
	return "Generate comprehensive single file documentation prompt using template system"
}

func (a *TemplateSingleFileDocsPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateSingleFileDocsPromptAction) DefaultOutput() string {
	return "./.aid/SINGLE-FILE-DOCS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateSingleFileDocsPromptAction) Validate() error {
	return nil
}

func (a *TemplateSingleFileDocsPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("single-file-docs", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}

// TemplateDiagramsPromptAction generates comprehensive Mermaid diagram prompts using templates
type TemplateDiagramsPromptAction struct{}

// Ensure TemplateDiagramsPromptAction implements ContentAction
var _ ai.ContentAction = (*TemplateDiagramsPromptAction)(nil)

func (a *TemplateDiagramsPromptAction) Name() string {
	return "prompt-for-diagrams"
}

func (a *TemplateDiagramsPromptAction) Description() string {
	return "Generate comprehensive Mermaid diagram analysis prompt for creating 10 most beneficial diagrams"
}

func (a *TemplateDiagramsPromptAction) Type() ai.ActionType {
	return ai.ActionTypePrompt
}

func (a *TemplateDiagramsPromptAction) DefaultOutput() string {
	return "./.aid/DIAGRAMS-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md"
}

func (a *TemplateDiagramsPromptAction) Validate() error {
	return nil
}

func (a *TemplateDiagramsPromptAction) GenerateContent(ctx *ai.ActionContext) (*ai.ContentResult, error) {
	data := CreateTemplateData(ctx.BaseName)
	prompt, err := LoadTemplate("diagrams", data)
	if err != nil {
		return nil, err
	}

	return &ai.ContentResult{
		ContentBefore: prompt,
		ContentAfter:  "",
	}, nil
}
