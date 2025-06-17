package language

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// LanguageDetector detects programming language from code content
type LanguageDetector struct {
	maxBytes int
}

// NewDetector creates a new language detector
func NewDetector() *LanguageDetector {
	return &LanguageDetector{
		maxBytes: 64 * 1024, // 64KB for detection
	}
}

// DetectFromReader detects language from an io.Reader
func (d *LanguageDetector) DetectFromReader(reader io.Reader) (string, error) {
	// Read up to maxBytes for detection
	buf := make([]byte, d.maxBytes)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Create a buffer that can be re-read
	content := buf[:n]

	// Try to detect language
	detected := d.detectLanguage(content)
	if detected == "" {
		return "", fmt.Errorf("could not detect language")
	}

	return detected, nil
}

// detectLanguage analyzes content and returns detected language
func (d *LanguageDetector) detectLanguage(content []byte) string {
	text := string(content)

	// 1. Check shebang first
	if lang := d.checkShebang(text); lang != "" {
		return lang
	}

	// 2. Check for high-confidence markers
	scores := make(map[string]int)

	// PHP - highest priority due to unique <?php tag
	if matched, _ := regexp.MatchString(`<\?php`, text); matched {
		return "php"
	}

	// Rust - unique syntax patterns
	if d.hasRustPatterns(text) {
		scores["rust"] += 100
	}

	// Go - package and := are very distinctive
	if d.hasGoPatterns(text) {
		scores["go"] += 100
	}

	// C++ - preprocessor directives
	if d.hasCppPatterns(text) {
		scores["c++"] += 100
	}

	// Swift - import UIKit/Foundation
	if d.hasSwiftPatterns(text) {
		scores["swift"] += 100
	}

	// C# - using System, namespace
	if d.hasCSharpPatterns(text) {
		scores["c#"] += 100
	}

	// Kotlin - fun and val keywords
	if d.hasKotlinPatterns(text) {
		scores["kotlin"] += 90
	}

	// Java - must check after Kotlin
	if d.hasJavaPatterns(text) {
		scores["java"] += 80
	}

	// Ruby - def/end blocks
	if d.hasRubyPatterns(text) {
		scores["ruby"] += 90
	}

	// Python - def with colons
	if d.hasPythonPatterns(text) {
		scores["python"] += 85
	}

	// TypeScript - must check before JavaScript
	if d.hasTypeScriptPatterns(text) {
		return "typescript"
	}

	// JavaScript - fallback for JS family
	if d.hasJavaScriptPatterns(text) {
		scores["javascript"] += 70
	}

	// Find language with highest score
	maxScore := 0
	detectedLang := ""
	for lang, score := range scores {
		if score > maxScore {
			maxScore = score
			detectedLang = lang
		}
	}

	// Require minimum score threshold
	if maxScore >= 70 {
		return detectedLang
	}

	return ""
}

// checkShebang checks for shebang line
func (d *LanguageDetector) checkShebang(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return ""
	}

	shebang := lines[0]
	if !strings.HasPrefix(shebang, "#!") {
		return ""
	}

	switch {
	case strings.Contains(shebang, "python"):
		return "python"
	case strings.Contains(shebang, "ruby"):
		return "ruby"
	case strings.Contains(shebang, "node"):
		return "javascript"
	case strings.Contains(shebang, "php"):
		return "php"
	case strings.Contains(shebang, "swift"):
		return "swift"
	}

	return ""
}

// Language-specific pattern detection functions

func (d *LanguageDetector) hasRustPatterns(text string) bool {
	rustPatterns := []string{
		`\bfn\s+\w+\s*\(`,       // fn function_name(
		`\blet\s+mut\s+`,        // let mut
		`\bimpl\s+\w+\s+for\s+`, // impl Trait for Type
		`\btrait\s+\w+\s*\{`,    // trait Name {
		`\buse\s+std::`,         // use std::
		`::<\w+>`,               // ::<Type> turbofish
	}

	score := 0
	for _, pattern := range rustPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasGoPatterns(text string) bool {
	goPatterns := []string{
		`^package\s+\w+`,                       // package main
		`^import\s+\(`,                         // import (
		`\bfunc\s+(\(\w+\s+\*?\w+\)\s+)?\w+\(`, // func or method
		`:=`,                                   // := short assignment
		`\bdefer\s+`,                           // defer keyword
		`\bchan\s+`,                            // chan keyword
		`\bgoroutine\s+`,                       // go keyword for goroutines
	}

	score := 0
	for _, pattern := range goPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasCppPatterns(text string) bool {
	cppPatterns := []string{
		`#include\s*<\w+>`,           // #include <iostream>
		`#pragma\s+once`,             // #pragma once
		`\bstd::\w+`,                 // std::cout
		`\btemplate\s*<`,             // template<
		`\bclass\s+\w+\s*:\s*public`, // class inheritance
		`\bnamespace\s+\w+\s*\{`,     // namespace
	}

	score := 0
	for _, pattern := range cppPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasSwiftPatterns(text string) bool {
	swiftPatterns := []string{
		`\bimport\s+(UIKit|Foundation|SwiftUI)`, // iOS imports
		`\bvar\s+\w+\s*:\s*\w+`,                 // var name: Type
		`\blet\s+\w+\s*:\s*\w+`,                 // let name: Type
		`\bfunc\s+\w+\([^)]*\)\s*->`,            // func name() -> Type
		`\bprotocol\s+\w+`,                      // protocol Name
		`\?\.`,                                  // optional chaining
		`!\s*$`,                                 // forced unwrap at end of line
	}

	score := 0
	for _, pattern := range swiftPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasCSharpPatterns(text string) bool {
	csharpPatterns := []string{
		`\busing\s+System`,                // using System;
		`\bnamespace\s+\w+`,               // namespace Name
		`\[\w+\]`,                         // [Attribute]
		`\bpublic\s+class\s+\w+`,          // public class Name
		`\bget;\s*set;`,                   // get; set; properties
		`\basync\s+Task`,                  // async Task
		`\bpublic\s+static\s+void\s+Main`, // Main method
	}

	score := 0
	for _, pattern := range csharpPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasKotlinPatterns(text string) bool {
	kotlinPatterns := []string{
		`\bpackage\s+com\.`,   // package com.example
		`\bimport\s+kotlin\.`, // import kotlin.
		`\bfun\s+\w+\(`,       // fun functionName(
		`\bval\s+\w+`,         // val name
		`\bvar\s+\w+`,         // var name
		`\?\.`,                // safe call operator
		`!!`,                  // not-null assertion
		`\bdata\s+class`,      // data class
	}

	score := 0
	for _, pattern := range kotlinPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasJavaPatterns(text string) bool {
	javaPatterns := []string{
		`\bimport\s+java\.`,               // import java.
		`\bpublic\s+static\s+void\s+main`, // main method
		`\bSystem\.out\.println`,          // System.out.println
		`\bpublic\s+class\s+\w+`,          // public class
		`\bimplements\s+\w+`,              // implements Interface
		`\bextends\s+\w+`,                 // extends Class
		`@Override`,                       // @Override annotation
	}

	score := 0
	for _, pattern := range javaPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasRubyPatterns(text string) bool {
	rubyPatterns := []string{
		`\brequire\s+['"]`,  // require 'gem'
		`\bdef\s+\w+`,       // def method_name
		`\bend$`,            // end at line end
		`:\w+`,              // :symbol
		`\bputs\s+`,         // puts
		`\|\|=`,             // ||= operator
		`@\w+`,              // @instance_variable
		`\bclass\s+\w+\s*<`, // class inheritance
	}

	score := 0
	for _, pattern := range rubyPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasPythonPatterns(text string) bool {
	pythonPatterns := []string{
		`\bdef\s+\w+\([^)]*\):`, // def function():
		`\bimport\s+\w+`,        // import module
		`\bfrom\s+\w+\s+import`, // from module import
		`\bclass\s+\w+.*:`,      // class Name:
		`\bself\b`,              // self keyword
		`__init__`,              // __init__ method
		`""".*"""`,              // docstring
		`\bif\s+__name__\s*==\s*['"]__main__['"]`, // if __name__ == "__main__"
	}

	score := 0
	for _, pattern := range pythonPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}

func (d *LanguageDetector) hasTypeScriptPatterns(text string) bool {
	// TypeScript-specific patterns (not in JavaScript)
	tsPatterns := []string{
		`:\s*(number|string|boolean|void|any|unknown)`, // type annotations
		`\binterface\s+\w+`,                            // interface declaration
		`\btype\s+\w+\s*=`,                             // type alias
		`\bimplements\s+\w+`,                           // implements interface
		`\bprivate\s+\w+:`,                             // private property
		`\bas\s+(unknown|any)`,                         // type assertion
		`<\w+>`,                                        // generic type
		`\benum\s+\w+`,                                 // enum declaration
	}

	for _, pattern := range tsPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	return false
}

func (d *LanguageDetector) hasJavaScriptPatterns(text string) bool {
	jsPatterns := []string{
		`\bconst\s+\w+\s*=`,        // const variable
		`\blet\s+\w+\s*=`,          // let variable
		`\bfunction\s+\w+\(`,       // function declaration
		`\brequire\(['"]`,          // require('module')
		`\bconsole\.log`,           // console.log
		`=>`,                       // arrow function
		`\basync\s+function`,       // async function
		`\bawait\s+`,               // await keyword
		`\bexport\s+(default\s+)?`, // export
	}

	score := 0
	for _, pattern := range jsPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			score++
		}
	}

	return score >= 2
}
