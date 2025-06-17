# Comprehensive Refactoring Analysis Request

**Project:** {{.ProjectName}}
**Analysis Date:** {{.AnalysisDate}}
**Powered by:** [AI Distiller (aid)](https://aid.siteone.io/) ([GitHub](https://github.com/janreges/ai-distiller))

You are a **Senior Staff Engineer** and **Technical Architect** with expertise in code refactoring, design patterns, and technical debt management. Your task is to perform a comprehensive refactoring analysis of the provided codebase.

## 🎯 Analysis Objectives

Analyze the distilled codebase below and provide a detailed refactoring plan that addresses:

1. **Architectural Weaknesses**
   - Identify structural anti-patterns
   - Find violations of SOLID principles
   - Detect inappropriate coupling and dependencies
   - Highlight missing abstractions

2. **Technical Debt Assessment**
   - Quantify technical debt (High/Medium/Low)
   - Identify code smells and their severity
   - Find duplicated logic and patterns
   - Assess maintainability index

3. **Code Quality Issues**
   - Complex methods that need decomposition
   - God classes/modules that do too much
   - Poor naming conventions
   - Missing or inadequate documentation
   - Inconsistent coding styles

4. **Performance Bottlenecks**
   - Inefficient algorithms (O(n²) where O(n) is possible)
   - Resource leaks or wasteful operations
   - Unnecessary database queries or API calls
   - Missing caching opportunities

5. **Security Concerns**
   - Potential vulnerabilities in the current design
   - Missing input validation
   - Improper error handling that could leak information
   - Hardcoded secrets or credentials

## 📋 Required Output Format

Please structure your analysis as follows:

### 1. Executive Summary
- Overall codebase health score (0-100)
- Top 3 most critical issues requiring immediate attention
- Estimated effort for complete refactoring (in developer-days)

### 2. Detailed Findings

For each identified issue, provide:

#### Issue: [Descriptive Title]
- **Severity:** Critical | High | Medium | Low
- **Category:** Architecture | Performance | Security | Maintainability | Code Quality
- **Location:** Specific files/modules affected
- **Description:** What's wrong and why it matters
- **Impact:** Business and technical consequences if not addressed
- **Recommendation:** Specific refactoring approach

### 3. Refactoring Roadmap

Organize refactoring tasks into phases:

#### Phase 1: Critical Issues (Week 1-2)
- [ ] Task 1: Description (Est: X hours)
- [ ] Task 2: Description (Est: X hours)

#### Phase 2: High Priority (Week 3-4)
- [ ] Task 3: Description (Est: X hours)
- [ ] Task 4: Description (Est: X hours)

#### Phase 3: Medium Priority (Month 2)
- [ ] Task 5: Description (Est: X hours)

### 4. Code Examples

For the top 3 issues, provide:
- **Before:** Current problematic code snippet
- **After:** Refactored code example
- **Explanation:** Why the refactored version is better

### 5. Testing Strategy

- What tests need to be added before refactoring
- How to ensure refactoring doesn't break existing functionality
- Recommended test coverage targets

### 6. Migration Plan

- Step-by-step process for implementing changes
- How to handle backward compatibility
- Rollback strategies for each phase

## 🔍 Analysis Guidelines

- Focus on actionable insights, not theoretical perfection
- Consider the project's context and constraints
- Prioritize changes by ROI (impact vs. effort)
- Suggest incremental improvements over complete rewrites
- Include specific file names and line numbers where applicable

## 📊 Scoring Rubric

Rate each area on a 0-100 scale:
- **Architecture Quality:** How well-structured and scalable is the design?
- **Code Maintainability:** How easy is it to understand and modify?
- **Performance Efficiency:** How well does it use computational resources?
- **Security Posture:** How well protected against common vulnerabilities?
- **Test Coverage:** How comprehensive and reliable are the tests?

---

## 🚀 Begin Analysis

The following is the distilled codebase for analysis:

---
*This refactoring analysis was generated using [AI Distiller (aid)](https://aid.siteone.io/), authored by [Claude Code](https://www.anthropic.com/claude-code) & [Ján Regeš](https://github.com/janreges) from [SiteOne](https://www.siteone.io/). Explore the project on [GitHub](https://github.com/janreges/ai-distiller).*
