package csharp

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
using System;

public class MyClass {
    private int value;

    public int GetValue() {
        return value;
    }
}`,
			expected: []string{
				"System",
				"MyClass",
				"value",
				"GetValue",
			},
		},
		{
			name: "namespace",
			code: `
namespace MyNamespace {
    public class Test {
        public void Method() {}
    }
}`,
			expected: []string{
				"MyNamespace",
				"Test",
				"Method",
			},
		},
		{
			name: "interface",
			code: `
public interface IService {
    void Execute();
    string Name { get; }
}`,
			expected: []string{
				"IService",
				"Execute",
				"Name",
			},
		},
		{
			name: "properties",
			code: `
public class Person {
    public string FirstName { get; set; }
    public string LastName { get; set; }
    public string FullName => $"{FirstName} {LastName}";
}`,
			expected: []string{
				"Person",
				"FirstName",
				"LastName",
				"FullName",
			},
		},
		{
			name: "enum",
			code: `
public enum Status {
    Active = 1,
    Inactive = 0
}`,
			expected: []string{
				"Status",
				"Active",
				"Inactive",
			},
		},
		{
			name: "struct",
			code: `
public struct Point {
    public int X { get; set; }
    public int Y { get; set; }

    public double Distance() {
        return Math.Sqrt(X * X + Y * Y);
    }
}`,
			expected: []string{
				"Point",
				"X",
				"Y",
				"Distance",
			},
		},
		{
			name: "generic_class",
			code: `
public class Container<T> {
    private T item;

    public void Set(T value) {
        item = value;
    }

    public T Get() {
        return item;
    }
}`,
			expected: []string{
				"Container",
				"item",
				"Set",
				"Get",
			},
		},
		{
			name: "async_method",
			code: `
public class Service {
    public async Task<string> GetDataAsync() {
        await Task.Delay(100);
        return "data";
    }
}`,
			expected: []string{
				"Service",
				"GetDataAsync",
				"async",
			},
		},
	}

	processor := NewProcessor()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.code)
			result, err := processor.Process(ctx, reader, "test.cs")
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

func TestProcessor_ModernFeatures(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "record",
			code: `
public record Person(string FirstName, string LastName) {
    public string FullName => $"{FirstName} {LastName}";
}`,
			expected: []string{
				"Person",
				"FirstName",
				"LastName",
				"FullName",
			},
		},
		{
			name: "file_scoped_namespace",
			code: `
namespace MyApp.Services;

public class UserService {
    public void ProcessUser() {}
}`,
			expected: []string{
				"MyApp.Services",
				"UserService",
				"ProcessUser",
			},
		},
		{
			name: "init_only_property",
			code: `
public class Config {
    public string Name { get; init; }
    public int Value { get; init; }
}`,
			expected: []string{
				"Config",
				"Name",
				"Value",
			},
		},
		{
			name: "pattern_matching",
			code: `
public class Calculator {
    public int Calculate(object value) {
        return value switch {
            int n => n * 2,
            string s => s.Length,
            _ => 0
        };
    }
}`,
			expected: []string{
				"Calculator",
				"Calculate",
			},
		},
		{
			name: "nullable_reference_types",
			code: `
public class NullableExample {
    public string? OptionalName { get; set; }
    public string RequiredName { get; set; } = "";
}`,
			expected: []string{
				"NullableExample",
				"OptionalName",
				"RequiredName",
			},
		},
	}

	processor := NewProcessor()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.code)
			result, err := processor.Process(ctx, reader, "test.cs")
			require.NoError(t, err)
			require.NotNil(t, result)

			content := convertToString(result)

			for _, exp := range tt.expected {
				assert.Contains(t, content, exp, "Expected to find %s in output", exp)
			}
		})
	}
}

func TestProcessor_ComplexFeatures(t *testing.T) {
	code := `
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace MyApp {
    // Interface with default implementation
    public interface ILogger {
        void Log(string message);
        void LogError(string message) => Log($"ERROR: {message}");
    }

    // Abstract base class
    public abstract class ServiceBase : ILogger {
        public abstract string Name { get; }
        public abstract Task ExecuteAsync();

        public virtual void Log(string message) {
            Console.WriteLine($"[{Name}]: {message}");
        }
    }

    // Sealed class with events
    public sealed class UserService : ServiceBase {
        private readonly List<User> users = new();

        public event EventHandler<UserEventArgs>? UserAdded;

        public override string Name => "UserService";

        public override async Task ExecuteAsync() {
            await Task.Delay(100);
            Log("Service executed");
        }

        public void AddUser(User user) {
            users.Add(user);
            UserAdded?.Invoke(this, new UserEventArgs(user));
        }
    }

    // Generic class with constraints
    public class Repository<T> where T : class, IEntity, new() {
        private readonly Dictionary<int, T> items = new();

        public void Add(T item) {
            items[item.Id] = item;
        }

        public T? Get(int id) {
            return items.TryGetValue(id, out var item) ? item : null;
        }
    }

    // Delegate
    public delegate void NotificationHandler(string message);

    // Enum with flags
    [Flags]
    public enum Permissions {
        None = 0,
        Read = 1,
        Write = 2,
        Delete = 4,
        All = Read | Write | Delete
    }

    // Extension methods
    public static class StringExtensions {
        public static bool IsNullOrEmpty(this string? value) {
            return string.IsNullOrEmpty(value);
        }
    }
}
`

	processor := NewProcessor()
	ctx := context.Background()

	reader := strings.NewReader(code)
	result, err := processor.Process(ctx, reader, "complex.cs")
	require.NoError(t, err)
	require.NotNil(t, result)

	content := convertToString(result)

	// Check various C# features
	assert.Contains(t, content, "MyApp")
	assert.Contains(t, content, "ILogger")
	assert.Contains(t, content, "ServiceBase")
	assert.Contains(t, content, "UserService")
	assert.Contains(t, content, "Repository")
	assert.Contains(t, content, "NotificationHandler")
	assert.Contains(t, content, "Permissions")
	assert.Contains(t, content, "StringExtensions")
	assert.Contains(t, content, "abstract")
	assert.Contains(t, content, "sealed")
	assert.Contains(t, content, "override")
	assert.Contains(t, content, "async")
}

func TestProcessor_NullableDebug(t *testing.T) {
	code := `public class NullableExample {
    public string? OptionalName { get; set; }
    public string RequiredName { get; set; } = "";
}`

	processor := NewProcessor()
	ctx := context.Background()

	reader := strings.NewReader(code)
	result, err := processor.Process(ctx, reader, "nullable.cs")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Print structure for debugging
	t.Logf("File has %d children", len(result.Children))
	for _, node := range result.Children {
		if class, ok := node.(*ir.DistilledClass); ok {
			t.Logf("Class: %s with %d children", class.Name, len(class.Children))
			for _, child := range class.Children {
				if field, ok := child.(*ir.DistilledField); ok {
					t.Logf("  Field: %s (Type: %v)", field.Name, field.Type)
				}
			}
		}
	}

	// Check that properties exist
	content := convertToString(result)
	assert.Contains(t, content, "OptionalName")
	assert.Contains(t, content, "RequiredName")
}

func TestProcessor_Visibility(t *testing.T) {
	code := `
public class VisibilityTest {
    public int PublicField;
    private int privateField;
    protected int protectedField;
    internal int internalField;

    public void PublicMethod() {}
    private void PrivateMethod() {}
    protected void ProtectedMethod() {}
    internal void InternalMethod() {}
}`

	processor := NewProcessor()
	ctx := context.Background()

	reader := strings.NewReader(code)
	result, err := processor.Process(ctx, reader, "visibility.cs")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check visibility assignments
	var checkVisibility func(ir.DistilledNode)
	checkVisibility = func(node ir.DistilledNode) {
		switch n := node.(type) {
		case *ir.DistilledClass:
			assert.Equal(t, ir.VisibilityPublic, n.Visibility)
			for _, child := range n.Children {
				checkVisibility(child)
			}
		case *ir.DistilledField:
			switch n.Name {
			case "PublicField":
				assert.Equal(t, ir.VisibilityPublic, n.Visibility)
			case "privateField":
				assert.Equal(t, ir.VisibilityPrivate, n.Visibility)
			case "protectedField":
				assert.Equal(t, ir.VisibilityProtected, n.Visibility)
			case "internalField":
				assert.Equal(t, ir.VisibilityInternal, n.Visibility)
			}
		case *ir.DistilledFunction:
			switch n.Name {
			case "PublicMethod":
				assert.Equal(t, ir.VisibilityPublic, n.Visibility)
			case "PrivateMethod":
				assert.Equal(t, ir.VisibilityPrivate, n.Visibility)
			case "ProtectedMethod":
				assert.Equal(t, ir.VisibilityProtected, n.Visibility)
			case "InternalMethod":
				assert.Equal(t, ir.VisibilityInternal, n.Visibility)
			}
		}
	}

	for _, node := range result.Children {
		checkVisibility(node)
	}
}

// Helper function to convert result to string representation
func convertToString(file *ir.DistilledFile) string {
	var sb strings.Builder

	// Recursively collect all names and modifiers from the IR structure
	var collectInfo func(node ir.DistilledNode)
	collectInfo = func(node ir.DistilledNode) {
		switch n := node.(type) {
		case *ir.DistilledImport:
			if n.Module != "" {
				sb.WriteString(n.Module + " ")
			}
			for _, sym := range n.Symbols {
				sb.WriteString(sym.Name + " ")
			}
		case *ir.DistilledClass:
			sb.WriteString(n.Name + " ")
			for _, mod := range n.Modifiers {
				sb.WriteString(string(mod) + " ")
			}
		case *ir.DistilledStruct:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledInterface:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledFunction:
			sb.WriteString(n.Name + " ")
			for _, mod := range n.Modifiers {
				sb.WriteString(string(mod) + " ")
			}
		case *ir.DistilledField:
			sb.WriteString(n.Name + " ")
			for _, mod := range n.Modifiers {
				sb.WriteString(string(mod) + " ")
			}
		case *ir.DistilledPackage:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledEnum:
			sb.WriteString(n.Name + " ")
		case *ir.DistilledTypeAlias:
			sb.WriteString(n.Name + " ")
		}

		// Process children
		if node != nil {
			for _, child := range node.GetChildren() {
				collectInfo(child)
			}
		}
	}

	// Process all top-level nodes
	for _, node := range file.Children {
		collectInfo(node)
	}

	return sb.String()
}
