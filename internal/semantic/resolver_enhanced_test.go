package semantic

import (
	"context"
	"strings"
	"testing"
)

func TestEnhancedResolver_BuiltinResolution(t *testing.T) {
	// Create a resolver with TypeScript support
	resolver := NewResolver("/test/project")
	
	// Create test semantic graph
	semanticGraph := NewSemanticGraph("/test/project")
	
	// Add a TypeScript file with builtin function calls
	testCode := `
	// Test file with builtin calls
	console.log('Hello World');
	const arr = Array.from([1, 2, 3]);
	const now = Date.now();
	const parsed = parseInt('123');
	const jsonData = JSON.parse('{"key": "value"}');
	`
	
	// Create analyzer and parse the test code
	analyzer, err := NewTypeScriptAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create TypeScript analyzer: %v", err)
	}
	
	reader := strings.NewReader(testCode)
	analysis, err := analyzer.AnalyzeFile(context.Background(), reader, "/test/project/main.ts")
	if err != nil {
		t.Fatalf("Failed to analyze test file: %v", err)
	}
	
	// Add to semantic graph
	semanticGraph.AddSymbolTable(analysis.SymbolTable)
	for _, call := range analysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	
	// Run resolution
	ctx := context.Background()
	err = resolver.ResolveProject(ctx, semanticGraph)
	if err != nil {
		t.Fatalf("Failed to resolve project: %v", err)
	}
	
	// Check that builtin calls were resolved
	resolvedBuiltins := 0
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved && strings.HasPrefix(string(callSite.CalleeID), "<builtin>::") {
			resolvedBuiltins++
			t.Logf("Resolved builtin call: %s -> %s", callSite.CalleeName, callSite.CalleeID)
		}
	}
	
	if resolvedBuiltins == 0 {
		t.Error("Expected some builtin function calls to be resolved")
	}
	
	// Verify specific builtins
	expectedBuiltins := map[string]bool{
		"console": false,
		"Array":   false,
		"Date":    false,
		"parseInt": false,
		"JSON":    false,
	}
	
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved {
			for builtin := range expectedBuiltins {
				if strings.Contains(callSite.CalleeName, builtin) || strings.Contains(string(callSite.CalleeID), builtin) {
					expectedBuiltins[builtin] = true
				}
			}
		}
	}
	
	for builtin, found := range expectedBuiltins {
		if !found {
			t.Errorf("Expected builtin '%s' to be resolved", builtin)
		}
	}
}

func TestEnhancedResolver_MemberAccessResolution(t *testing.T) {
	resolver := NewResolver("/test/project")
	semanticGraph := NewSemanticGraph("/test/project")
	
	// Create test files with member access patterns
	classCode := `
	export class UserService {
		private users: User[] = [];
		
		public findUser(id: number): User | null {
			return this.users.find(u => u.id === id) || null;
		}
		
		public static getInstance(): UserService {
			return new UserService();
		}
	}
	
	export interface User {
		id: number;
		name: string;
	}
	`
	
	mainCode := `
	import { UserService, User } from './user-service';
	
	const service = UserService.getInstance();
	const user = service.findUser(123);
	
	if (user) {
		console.log(user.name);
	}
	
	// Method chaining
	const result = service.findUser(456)?.name || 'Unknown';
	`
	
	// Analyze both files
	analyzer, err := NewTypeScriptAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create TypeScript analyzer: %v", err)
	}
	
	// Analyze UserService file
	serviceReader := strings.NewReader(classCode)
	serviceAnalysis, err := analyzer.AnalyzeFile(context.Background(), serviceReader, "/test/project/user-service.ts")
	if err != nil {
		t.Fatalf("Failed to analyze service file: %v", err)
	}
	
	// Analyze main file
	mainReader := strings.NewReader(mainCode)
	mainAnalysis, err := analyzer.AnalyzeFile(context.Background(), mainReader, "/test/project/main.ts")
	if err != nil {
		t.Fatalf("Failed to analyze main file: %v", err)
	}
	
	// Add dependency from main to service
	mainAnalysis.Dependencies = append(mainAnalysis.Dependencies, DependencyInfo{
		SourceFile:   "/test/project/main.ts",
		TargetModule: "./user-service",
		ImportType:   "import",
		IsRelative:   true,
	})
	
	// Add to semantic graph
	semanticGraph.AddSymbolTable(serviceAnalysis.SymbolTable)
	semanticGraph.AddSymbolTable(mainAnalysis.SymbolTable)
	
	for _, call := range serviceAnalysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	for _, call := range mainAnalysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	
	// Run resolution
	ctx := context.Background()
	err = resolver.ResolveProject(ctx, semanticGraph)
	if err != nil {
		t.Fatalf("Failed to resolve project: %v", err)
	}
	
	// Check member access resolution
	resolvedMemberCalls := 0
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved && strings.Contains(callSite.CalleeName, ".") {
			resolvedMemberCalls++
			t.Logf("Resolved member access: %s -> %s", callSite.CalleeName, callSite.CalleeID)
		}
	}
	
	if resolvedMemberCalls == 0 {
		t.Error("Expected some member access calls to be resolved")
	}
	
	// Verify specific method calls were resolved
	expectedMethods := []string{"getInstance", "findUser"}
	for _, expectedMethod := range expectedMethods {
		found := false
		for _, callSite := range semanticGraph.CallSites {
			if callSite.IsResolved && (callSite.CalleeName == expectedMethod || strings.Contains(string(callSite.CalleeID), expectedMethod)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected method '%s' to be resolved", expectedMethod)
		}
	}
}

func TestEnhancedResolver_TypeInference(t *testing.T) {
	resolver := NewResolver("/test/project")
	semanticGraph := NewSemanticGraph("/test/project")
	
	testCode := `
	interface Calculator {
		add(a: number, b: number): number;
		multiply(a: number, b: number): number;
	}
	
	class BasicCalculator implements Calculator {
		add(a: number, b: number): number {
			return a + b;
		}
		
		multiply(a: number, b: number): number {
			return a * b;
		}
	}
	
	function createCalculator(): Calculator {
		return new BasicCalculator();
	}
	
	// Type inference scenarios
	const calc = createCalculator();
	const sum = calc.add(5, 3);
	const product = calc.multiply(4, 7);
	
	// Generic function with inference
	function processNumber<T extends number>(value: T, processor: (val: T) => T): T {
		return processor(value);
	}
	
	const doubled = processNumber(5, (x) => x * 2);
	`
	
	analyzer, err := NewTypeScriptAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create TypeScript analyzer: %v", err)
	}
	
	reader := strings.NewReader(testCode)
	analysis, err := analyzer.AnalyzeFile(context.Background(), reader, "/test/project/calculator.ts")
	if err != nil {
		t.Fatalf("Failed to analyze test file: %v", err)
	}
	
	// Add to semantic graph
	semanticGraph.AddSymbolTable(analysis.SymbolTable)
	for _, call := range analysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	
	// Run resolution
	ctx := context.Background()
	err = resolver.ResolveProject(ctx, semanticGraph)
	if err != nil {
		t.Fatalf("Failed to resolve project: %v", err)
	}
	
	// Check that method calls on inferred types were resolved
	resolvedCalls := 0
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved {
			resolvedCalls++
			t.Logf("Resolved call: %s -> %s", callSite.CalleeName, callSite.CalleeID)
		}
	}
	
	if resolvedCalls == 0 {
		t.Error("Expected some calls to be resolved through type inference")
	}
	
	// Verify specific interface method calls were resolved
	interfaceMethods := []string{"add", "multiply"}
	for _, method := range interfaceMethods {
		found := false
		for _, callSite := range semanticGraph.CallSites {
			if callSite.IsResolved && callSite.CalleeName == method {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected interface method '%s' to be resolved", method)
		}
	}
}

func TestEnhancedResolver_CrossFileResolution(t *testing.T) {
	resolver := NewResolver("/test/project")
	semanticGraph := NewSemanticGraph("/test/project")
	
	// Utility module
	utilsCode := `
	export function validateEmail(email: string): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
	}
	
	export function formatDate(date: Date): string {
		return date.toLocaleDateString();
	}
	
	export const CONFIG = {
		apiUrl: 'https://api.example.com',
		timeout: 5000
	};
	`
	
	// Main module that uses utils
	mainCode := `
	import { validateEmail, formatDate, CONFIG } from './utils';
	import * as DateUtils from './date-utils';
	
	class UserManager {
		processUser(email: string, birthDate: Date): boolean {
			if (!validateEmail(email)) {
				return false;
			}
			
			const formattedDate = formatDate(birthDate);
			console.log('User birthday:', formattedDate);
			
			// Call function from namespace import
			const age = DateUtils.calculateAge(birthDate);
			
			return true;
		}
	}
	
	// Static access to imported constant
	fetch(CONFIG.apiUrl);
	`
	
	// Date utilities module
	dateUtilsCode := `
	export function calculateAge(birthDate: Date): number {
		const today = new Date();
		const age = today.getFullYear() - birthDate.getFullYear();
		return age;
	}
	
	export function isLeapYear(year: number): boolean {
		return (year % 4 === 0 && year % 100 !== 0) || (year % 400 === 0);
	}
	`
	
	analyzer, err := NewTypeScriptAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create TypeScript analyzer: %v", err)
	}
	
	// Analyze all files
	utilsReader := strings.NewReader(utilsCode)
	utilsAnalysis, err := analyzer.AnalyzeFile(context.Background(), utilsReader, "/test/project/utils.ts")
	if err != nil {
		t.Fatalf("Failed to analyze utils file: %v", err)
	}
	
	dateUtilsReader := strings.NewReader(dateUtilsCode)
	dateUtilsAnalysis, err := analyzer.AnalyzeFile(context.Background(), dateUtilsReader, "/test/project/date-utils.ts")
	if err != nil {
		t.Fatalf("Failed to analyze date-utils file: %v", err)
	}
	
	mainReader := strings.NewReader(mainCode)
	mainAnalysis, err := analyzer.AnalyzeFile(context.Background(), mainReader, "/test/project/main.ts")
	if err != nil {
		t.Fatalf("Failed to analyze main file: %v", err)
	}
	
	// Set up dependencies
	mainAnalysis.Dependencies = []DependencyInfo{
		{
			SourceFile:   "/test/project/main.ts",
			TargetModule: "./utils",
			ImportType:   "import",
			IsRelative:   true,
		},
		{
			SourceFile:   "/test/project/main.ts",
			TargetModule: "./date-utils",
			ImportType:   "import",
			IsRelative:   true,
		},
	}
	
	// Add to semantic graph
	semanticGraph.AddSymbolTable(utilsAnalysis.SymbolTable)
	semanticGraph.AddSymbolTable(dateUtilsAnalysis.SymbolTable)
	semanticGraph.AddSymbolTable(mainAnalysis.SymbolTable)
	
	for _, call := range utilsAnalysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	for _, call := range dateUtilsAnalysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	for _, call := range mainAnalysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	
	// Run resolution
	ctx := context.Background()
	err = resolver.ResolveProject(ctx, semanticGraph)
	if err != nil {
		t.Fatalf("Failed to resolve project: %v", err)
	}
	
	// Check cross-file resolution
	crossFileResolvedCalls := 0
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved {
			// Check if resolved call is to a different file
			callSiteFile := callSite.Location.FilePath
			if symbol := findSymbolByID(semanticGraph, callSite.CalleeID); symbol != nil {
				symbolFile := symbol.Location.FilePath
				if callSiteFile != symbolFile {
					crossFileResolvedCalls++
					t.Logf("Cross-file resolution: %s in %s -> %s in %s", 
						callSite.CalleeName, callSiteFile, symbol.Name, symbolFile)
				}
			}
		}
	}
	
	if crossFileResolvedCalls == 0 {
		t.Error("Expected some cross-file function calls to be resolved")
	}
	
	// Verify specific cross-file calls
	expectedCrossCalls := []string{"validateEmail", "formatDate", "calculateAge"}
	for _, expectedCall := range expectedCrossCalls {
		found := false
		for _, callSite := range semanticGraph.CallSites {
			if callSite.IsResolved && callSite.CalleeName == expectedCall {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected cross-file call '%s' to be resolved", expectedCall)
		}
	}
}

// Helper function to find symbol by ID in semantic graph
func findSymbolByID(graph *SemanticGraph, symbolID SymbolID) *Symbol {
	for _, symbolTable := range graph.FileSymbolTables {
		for _, symbol := range symbolTable.Symbols {
			if symbol.ID == symbolID {
				return symbol
			}
		}
	}
	return nil
}

func TestEnhancedResolver_ComplexInheritance(t *testing.T) {
	resolver := NewResolver("/test/project")
	semanticGraph := NewSemanticGraph("/test/project")
	
	testCode := `
	// Base interface
	interface Drawable {
		draw(): void;
	}
	
	// Extended interface
	interface Shape extends Drawable {
		area(): number;
		perimeter(): number;
	}
	
	// Base class
	abstract class GeometricShape implements Shape {
		protected color: string;
		
		constructor(color: string) {
			this.color = color;
		}
		
		abstract area(): number;
		abstract perimeter(): number;
		
		draw(): void {
			console.log('Drawing', this.color, 'shape');
		}
	}
	
	// Concrete implementation
	class Rectangle extends GeometricShape {
		constructor(
			private width: number,
			private height: number,
			color: string
		) {
			super(color);
		}
		
		area(): number {
			return this.width * this.height;
		}
		
		perimeter(): number {
			return 2 * (this.width + this.height);
		}
		
		// Additional method
		getDimensions(): { width: number; height: number } {
			return { width: this.width, height: this.height };
		}
	}
	
	// Usage with polymorphism
	function processShape(shape: Shape): void {
		shape.draw();  // Should resolve to base class method
		const area = shape.area();  // Should resolve to interface method
		console.log('Area:', area);
	}
	
	const rect = new Rectangle(10, 5, 'blue');
	processShape(rect);
	
	// Direct method calls
	rect.getDimensions();
	rect.perimeter();
	`
	
	analyzer, err := NewTypeScriptAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create TypeScript analyzer: %v", err)
	}
	
	reader := strings.NewReader(testCode)
	analysis, err := analyzer.AnalyzeFile(context.Background(), reader, "/test/project/shapes.ts")
	if err != nil {
		t.Fatalf("Failed to analyze test file: %v", err)
	}
	
	// Add to semantic graph
	semanticGraph.AddSymbolTable(analysis.SymbolTable)
	for _, call := range analysis.CallSites {
		semanticGraph.AddCallSite(call)
	}
	
	// Run resolution
	ctx := context.Background()
	err = resolver.ResolveProject(ctx, semanticGraph)
	if err != nil {
		t.Fatalf("Failed to resolve project: %v", err)
	}
	
	// Check inheritance-based resolution
	resolvedInheritanceCalls := 0
	for _, callSite := range semanticGraph.CallSites {
		if callSite.IsResolved {
			resolvedInheritanceCalls++
			t.Logf("Resolved inheritance call: %s -> %s", callSite.CalleeName, callSite.CalleeID)
		}
	}
	
	if resolvedInheritanceCalls == 0 {
		t.Error("Expected some inheritance-based method calls to be resolved")
	}
	
	// Verify polymorphic method calls
	polymorphicMethods := []string{"draw", "area", "perimeter", "getDimensions"}
	for _, method := range polymorphicMethods {
		found := false
		for _, callSite := range semanticGraph.CallSites {
			if callSite.IsResolved && callSite.CalleeName == method {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected polymorphic method '%s' to be resolved", method)
		}
	}
}