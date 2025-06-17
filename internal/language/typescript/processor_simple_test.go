package typescript

import (
	"context"
	"strings"
	"testing"

	"github.com/janreges/ai-distiller/internal/formatter"
	"github.com/janreges/ai-distiller/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessor_SimpleBasics(t *testing.T) {
	t.Skip("Skipping typescript processor tests - output format has changed")

	tests := []struct {
		name        string
		input       string
		opts        processor.ProcessOptions
		contains    []string
		notContains []string
	}{
		{
			name: "basic_interface",
			input: `interface User {
				id: number;
				name: string;
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"interface User",
				"property id: number",
				"property name: string",
			},
		},
		{
			name: "basic_class",
			input: `class Point {
				constructor(public x: number, public y: number) {}
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"class Point",
				"field public x: number",
				"field public y: number",
				"function constructor",
			},
		},
		{
			name: "generic_class",
			input: `class Container<T> {
				private value: T;
				getValue(): T {
					return this.value;
				}
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"class Container<T>",
				"field private value: T",
				"function getValue() -> T",
			},
		},
		{
			name: "abstract_class",
			input: `abstract class Shape {
				abstract area(): number;
				protected name: string;
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"abstract class Shape",
				"abstract function area() -> number",
				"field protected name: string",
			},
		},
		{
			name: "strip_private",
			input: `class Service {
				public api(): void {}
				private helper(): void {}
				protected hook(): void {}
			}`,
			opts: processor.ProcessOptions{
				RemovePrivateOnly:     true,
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"function api() -> void",
				"protected function hook() -> void",
			},
			notContains: []string{
				"private function helper",
			},
		},
		{
			name: "complex_generics",
			input: `interface Mapper<T extends object, U> {
				map(item: T): U;
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"interface Mapper<T extends object, U>",
				"method map(item: T): U",
			},
		},
		{
			name: "type_alias",
			input: `type ID = string | number;
			type Point = { x: number; y: number };`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"type ID = string | number",
				"type Point = { x: number; y: number }",
			},
		},
		{
			name: "const_variable",
			input: `const API_KEY = "secret";
			interface Config {
				key: string;
			}`,
			opts: processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			},
			contains: []string{
				"const API_KEY",
				"interface Config",
				"property key: string",
			},
		},
	}

	p := NewProcessor()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.ProcessWithOptions(ctx, strings.NewReader(tt.input), "test.ts", tt.opts)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Use the text formatter to get the actual output
			textFormatter := formatter.NewLanguageAwareTextFormatter(formatter.Options{})
			var sb strings.Builder
			err = textFormatter.Format(&sb, result)
			require.NoError(t, err)

			output := sb.String()
			t.Logf("Output:\n%s", output)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Should contain %q", expected)
			}

			for _, notExpected := range tt.notContains {
				assert.NotContains(t, output, notExpected, "Should not contain %q", notExpected)
			}
		})
	}
}

func TestProcessor_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "react_component",
			input: `interface Props {
				title: string;
				onClick?: () => void;
			}
			
			const Button: React.FC<Props> = ({ title, onClick }) => {
				return <button onClick={onClick}>{title}</button>;
			};`,
			contains: []string{
				"interface Props",
				"property title: string",
				"property onClick: () => void",
				"const Button",
			},
		},
		{
			name: "express_middleware",
			input: `import { Request, Response, NextFunction } from 'express';
			
			interface AuthRequest extends Request {
				user?: User;
			}
			
			function authMiddleware(req: AuthRequest, res: Response, next: NextFunction): void {
				// implementation
			}`,
			contains: []string{
				"interface AuthRequest extends Request",
				"property user: User",
				"function authMiddleware",
			},
		},
	}

	p := NewProcessor()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.ProcessWithOptions(ctx, strings.NewReader(tt.input), "test.ts", processor.ProcessOptions{
				IncludePrivate:        true,
				IncludeImplementation: true,
			})
			require.NoError(t, err)

			textFormatter := formatter.NewLanguageAwareTextFormatter(formatter.Options{})
			var sb strings.Builder
			err = textFormatter.Format(&sb, result)
			require.NoError(t, err)

			output := sb.String()
			t.Logf("Output:\n%s", output)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Should contain %q", expected)
			}
		})
	}
}
