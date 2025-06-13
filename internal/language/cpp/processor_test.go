package cpp

import (
	"context"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessor_Basic(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "simple_class",
			code: `
class MyClass {
public:
    int getValue() { return value; }
private:
    int value;
};`,
			expected: []string{
				"MyClass",
				"getValue",
				"value",
			},
		},
		{
			name: "namespace",
			code: `
namespace MyNamespace {
    void myFunction() {}
}`,
			expected: []string{
				"MyNamespace",
				"myFunction",
			},
		},
		{
			name: "struct_with_methods",
			code: `
struct Point {
    int x, y;
    int distance() { return x + y; }
};`,
			expected: []string{
				"Point",
				"x",
				"y", 
				"distance",
			},
		},
		{
			name: "virtual_functions",
			code: `
class Base {
public:
    virtual void doSomething() = 0;
    virtual void doAnother() {}
};`,
			expected: []string{
				"Base",
				"doSomething",
				"doAnother",
			},
		},
		{
			name: "template_function",
			code: `
template<typename T>
T max(T a, T b) {
    return a > b ? a : b;
}`,
			expected: []string{
				"max",
			},
		},
	}

	processor := NewProcessor()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.code)
			result, err := processor.Process(ctx, reader, "test.cpp")
			require.NoError(t, err)
			require.NotNil(t, result)

			// Convert result to string for easier checking
			content := convertToString(result)
			
			// Check that all expected elements are present
			for _, exp := range tt.expected {
				assert.Contains(t, content, exp, "Expected to find %s in output", exp)
			}
		})
	}
}

func TestProcessor_ComplexFeatures(t *testing.T) {
	code := `
#include <iostream>
#include <vector>

using namespace std;

// Template class
template<typename T>
class Container {
private:
    vector<T> items;
public:
    void add(T item) { items.push_back(item); }
    T get(int index) { return items[index]; }
};

// Abstract base class
class Shape {
public:
    virtual double area() = 0;
    virtual ~Shape() {}
};

// Derived class
class Circle : public Shape {
private:
    double radius;
public:
    Circle(double r) : radius(r) {}
    double area() override { return 3.14 * radius * radius; }
};

// Function with default parameter
void print(const string& msg = "Hello") {
    cout << msg << endl;
}

// Static member
class Counter {
    static int count;
public:
    Counter() { count++; }
    static int getCount() { return count; }
};

int Counter::count = 0;
`

	processor := NewProcessor()
	ctx := context.Background()
	
	reader := strings.NewReader(code)
	result, err := processor.Process(ctx, reader, "complex.cpp")
	require.NoError(t, err)
	require.NotNil(t, result)
	
	content := convertToString(result)
	
	// Check various C++ features
	assert.Contains(t, content, "Container")
	assert.Contains(t, content, "Shape")
	assert.Contains(t, content, "Circle")
	assert.Contains(t, content, "area")
	// assert.Contains(t, content, "override") // TODO: Add override modifier support
	assert.Contains(t, content, "static")
	assert.Contains(t, content, "count")
}

// Helper function to convert result to string representation
func convertToString(file *ir.DistilledFile) string {
	var sb strings.Builder
	
	// Recursively collect all names from the IR structure
	var collectNames func(node ir.DistilledNode)
	collectNames = func(node ir.DistilledNode) {
		switch n := node.(type) {
		case *ir.DistilledClass:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledStruct:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledFunction:
			sb.WriteString(n.Name + " ")
			for _, mod := range n.Modifiers {
				sb.WriteString(string(mod) + " ")
			}
		case *ir.DistilledField:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledPackage:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledEnum:
			sb.WriteString(n.Name + " ")
		}
		
		// Process children
		if node != nil {
			for _, child := range node.GetChildren() {
				collectNames(child)
			}
		}
	}
	
	// Process all top-level nodes
	for _, node := range file.Children {
		collectNames(node)
	}
	
	return sb.String()
}