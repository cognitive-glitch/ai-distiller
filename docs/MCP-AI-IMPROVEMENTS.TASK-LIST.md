# MCP AI Improvements - Task List

## Overview

This document tracks the implementation of a comprehensive AI action system for the AI Distiller (`aid`) CLI tool. The goal is to replace multiple AI-related CLI parameters with a unified `--ai-action=ENUM` system that provides various AI-powered analysis and documentation generation capabilities.

## Core Concept

Instead of having multiple CLI flags like `--ai-analysis-task-list`, we will implement:
- `--ai-action=<action-name>` - Select which AI action to perform
- `--ai-output=<path>` - Optional output path (each action has sensible defaults)

## Task List

### Phase 1: Research & Design
- [ ] Map all current AI-related CLI parameters in aid
- [ ] Consult with Gemini Pro and o3 for architecture design
- [ ] Design the AI Action interface/contract
- [ ] Plan the refactoring approach

### Phase 2: Core Implementation
- [ ] Implement base AI Action system (`--ai-action` parameter)
- [ ] Create AI Action interface/base class
- [ ] Implement action registry and discovery
- [ ] Add `--ai-output` parameter with template support
- [ ] Refactor existing `--ai-analysis-task-list` to new system

### Phase 3: AI Actions Implementation

#### Analysis Actions
- [ ] `prompt-for-refactoring-suggestion` - Generate refactoring analysis prompt
  - Output: `./.aid/REFACTORING-SUGGESTION.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Find architectural weaknesses, technical debt, and provide actionable steps

- [ ] `prompt-for-complex-codebase-analysis` - Comprehensive codebase analysis
  - Output: `./.aid/COMPLEX-CODEBASE-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Security, performance, best practices, quality scoring
  - Architecture diagram (ASCII/Mermaid)

- [ ] `prompt-for-security-analysis` - Security-focused analysis
  - Output: `./.aid/SECURITY-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Maximum emphasis on all security aspects

- [ ] `prompt-for-performance-analysis` - Performance optimization analysis
  - Output: `./.aid/PERFORMANCE-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Focus on performance techniques and language-specific optimizations

- [ ] `prompt-for-best-practices-analysis` - Best practices review
  - Output: `./.aid/BEST-PRACTICES-ANALYSIS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Error handling, logging, debugging support

- [ ] `prompt-for-bug-hunting` - Bug detection prompt
  - Output: `./.aid/BUG-HUNTING.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Specialized prompt for finding various types of bugs

#### Documentation Actions
- [ ] `prompt-for-single-file-docs` - Single-file documentation prompt
  - Output: `./.aid/SINGLE-FILE-DOCS.%YYYY-MM-DD.HH-MM-SS%.%folder-basename%.md`
  - Comprehensive single-file documentation

- [ ] `flow-for-multi-file-docs` - Multi-file documentation workflow
  - Output: `./.aid/README.md` + `./docs/` directory
  - Creates structured documentation with task list

#### Workflow Actions
- [ ] `flow-for-deep-file-to-file-analysis` - File-by-file analysis workflow
  - Current `--ai-analysis-task-list` functionality
  - Creates task list for systematic analysis

### Phase 4: Integration & Testing
- [ ] Update `aid --help` documentation
- [ ] Update CLAUDE.md with new AI actions
- [ ] Write comprehensive tests for AI action system
- [ ] Test all actions with real codebases
- [ ] Ensure backward compatibility where needed

### Phase 5: Documentation
- [ ] Create user documentation for each AI action
- [ ] Add examples to README
- [ ] Document the AI Action development guide
- [ ] Update all relevant documentation

## Technical Details

### AI Action Interface
Each AI action should implement:
- `Name() string` - Action identifier
- `Description() string` - Help text
- `DefaultOutput() string` - Default output path template
- `Execute(distilledContent string, outputPath string) error` - Main execution

### Output Path Templates
Support for variables in output paths:
- `%YYYY-MM-DD%` - Current date
- `%HH-MM-SS%` - Current time
- `%folder-basename%` - Base name of analyzed directory
- `%action%` - Action name

### Prompt Engineering
Each prompt should be:
- Well-tested and refined
- In English for maximum AI compatibility
- Structured for optimal AI understanding
- Include clear instructions and expected output format