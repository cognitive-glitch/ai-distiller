package python

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

// Benchmark scenarios with different file sizes
var benchmarkSizes = []struct {
	name  string
	lines int
	gen   func(int) string
}{
	{"Small_50", 50, generateSmallFile},
	{"Medium_500", 500, generateMediumFile},
	{"Large_5000", 5000, generateLargeFile},
	{"VeryLarge_10000", 10000, generateLargeFile},
}

// BenchmarkProcessor measures parsing performance
func BenchmarkProcessor(b *testing.B) {
	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			// Generate test content
			content := size.gen(size.lines)
			contentBytes := []byte(content)

			// Create processor once
			p := NewProcessor()
			ctx := context.Background()

			// Reset timer to exclude setup
			b.ResetTimer()
			b.ReportAllocs()

			// Run benchmark
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(contentBytes)
				_, err := p.Process(ctx, reader, "bench_test.py")
				if err != nil {
					b.Fatal(err)
				}
			}

			// Report metrics
			b.SetBytes(int64(len(contentBytes)))
			b.ReportMetric(float64(size.lines)/b.Elapsed().Seconds(), "lines/sec")
		})
	}
}

// BenchmarkProcessorParallel measures concurrent parsing performance
func BenchmarkProcessorParallel(b *testing.B) {
	// Use medium-sized file for parallel benchmark
	content := generateMediumFile(500)
	contentBytes := []byte(content)

	// Create processor
	p := NewProcessor()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := bytes.NewReader(contentBytes)
			_, err := p.Process(ctx, reader, "bench_test.py")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSpecificFeatures benchmarks specific parsing features
func BenchmarkSpecificFeatures(b *testing.B) {
	scenarios := []struct {
		name    string
		content string
	}{
		{
			"SimpleClass",
			`class MyClass:
    def __init__(self):
        pass
    def method(self):
        pass
`,
		},
		{
			"ComplexImports",
			`from typing import (
    List, Dict, Optional, Union,
    TypeVar, Generic, Protocol,
    Callable, AsyncIterator, Tuple
)
import os
import sys
from collections import defaultdict as dd
`,
		},
		{
			"NestedStructures",
			`class Outer:
    class Inner:
        class Deepest:
            def method(self):
                def inner_func():
                    pass
                return inner_func
`,
		},
		{
			"TypeAnnotations",
			`def complex_function(
    arg1: Dict[str, Union[int, float]],
    arg2: List[Tuple[str, Optional[int]]],
    arg3: Callable[[str, int], bool]
) -> Optional[Dict[str, Any]]:
    pass
`,
		},
	}

	p := NewProcessor()
	ctx := context.Background()

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			contentBytes := []byte(scenario.content)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(contentBytes)
				_, err := p.Process(ctx, reader, "bench_test.py")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Generator functions for different file types

func generateSmallFile(lines int) string {
	var sb strings.Builder
	sb.WriteString("# Small Python file for benchmarking\n")
	sb.WriteString("import os\n")
	sb.WriteString("from typing import List, Dict\n\n")

	// Generate a few classes with methods
	classCount := lines / 10
	for i := 0; i < classCount; i++ {
		sb.WriteString(fmt.Sprintf("class Class%d:\n", i))
		sb.WriteString("    def __init__(self):\n")
		sb.WriteString("        self.value = 0\n")
		sb.WriteString("    def method(self, x: int) -> int:\n")
		sb.WriteString("        return x * 2\n\n")
	}

	return sb.String()
}

func generateMediumFile(lines int) string {
	var sb strings.Builder
	sb.WriteString("# Medium Python file for benchmarking\n")
	sb.WriteString("from typing import List, Dict, Optional, Union, Tuple\n")
	sb.WriteString("import sys\n")
	sb.WriteString("import json\n")
	sb.WriteString("from collections import defaultdict\n\n")

	// Mix of classes and functions
	itemCount := lines / 20
	for i := 0; i < itemCount; i++ {
		// Class with multiple methods
		sb.WriteString(fmt.Sprintf("class Service%d:\n", i))
		sb.WriteString("    \"\"\"Service class with documentation.\"\"\"\n")
		sb.WriteString("    \n")
		sb.WriteString("    def __init__(self, config: Dict[str, Any]):\n")
		sb.WriteString("        self.config = config\n")
		sb.WriteString("        self._cache = {}\n")
		sb.WriteString("    \n")
		sb.WriteString("    def process(self, data: List[str]) -> Optional[str]:\n")
		sb.WriteString("        \"\"\"Process data and return result.\"\"\"\n")
		sb.WriteString("        if not data:\n")
		sb.WriteString("            return None\n")
		sb.WriteString("        return data[0]\n")
		sb.WriteString("    \n")
		sb.WriteString("    @property\n")
		sb.WriteString("    def is_ready(self) -> bool:\n")
		sb.WriteString("        return True\n\n")

		// Standalone function
		sb.WriteString(fmt.Sprintf("def process_item_%d(item: Union[str, int]) -> str:\n", i))
		sb.WriteString("    \"\"\"Process a single item.\"\"\"\n")
		sb.WriteString("    return str(item)\n\n")
	}

	return sb.String()
}

func generateLargeFile(lines int) string {
	var sb strings.Builder
	sb.WriteString("# Large Python file for benchmarking\n")
	sb.WriteString("# This simulates a real-world Python module\n\n")

	// Complex imports
	sb.WriteString("from __future__ import annotations\n")
	sb.WriteString("from typing import (\n")
	sb.WriteString("    List, Dict, Optional, Union, Tuple,\n")
	sb.WriteString("    TypeVar, Generic, Protocol, Callable,\n")
	sb.WriteString("    AsyncIterator, ClassVar, Final\n")
	sb.WriteString(")\n")
	sb.WriteString("import asyncio\n")
	sb.WriteString("import json\n")
	sb.WriteString("import logging\n")
	sb.WriteString("from abc import ABC, abstractmethod\n")
	sb.WriteString("from dataclasses import dataclass, field\n\n")

	// Generate various Python constructs
	currentLine := 15
	classIndex := 0
	funcIndex := 0

	for currentLine < lines {
		choice := currentLine % 3

		switch choice {
		case 0:
			// Generate a class
			linesAdded := generateBenchmarkClass(&sb, classIndex)
			currentLine += linesAdded
			classIndex++

		case 1:
			// Generate a function
			linesAdded := generateBenchmarkFunction(&sb, funcIndex)
			currentLine += linesAdded
			funcIndex++

		case 2:
			// Generate a dataclass or enum
			linesAdded := generateBenchmarkDataStructure(&sb, classIndex)
			currentLine += linesAdded
			classIndex++
		}
	}

	return sb.String()
}

func generateBenchmarkClass(sb *strings.Builder, index int) int {
	lines := 0

	sb.WriteString(fmt.Sprintf("class Component%d(ABC):\n", index))
	lines++
	sb.WriteString("    \"\"\"Component with abstract methods.\"\"\"\n")
	lines++
	sb.WriteString("    \n")
	lines++
	sb.WriteString("    def __init__(self, name: str, config: Optional[Dict[str, Any]] = None):\n")
	lines++
	sb.WriteString("        self.name = name\n")
	lines++
	sb.WriteString("        self.config = config or {}\n")
	lines++
	sb.WriteString("        self._initialized = False\n")
	lines++
	sb.WriteString("    \n")
	lines++
	sb.WriteString("    @abstractmethod\n")
	lines++
	sb.WriteString("    def process(self, data: Any) -> Any:\n")
	lines++
	sb.WriteString("        \"\"\"Process data.\"\"\"\n")
	lines++
	sb.WriteString("        pass\n")
	lines++
	sb.WriteString("    \n")
	lines++
	sb.WriteString("    def initialize(self) -> None:\n")
	lines++
	sb.WriteString("        \"\"\"Initialize component.\"\"\"\n")
	lines++
	sb.WriteString("        self._initialized = True\n")
	lines++
	sb.WriteString("    \n")
	lines++
	sb.WriteString("    @property\n")
	lines++
	sb.WriteString("    def is_initialized(self) -> bool:\n")
	lines++
	sb.WriteString("        return self._initialized\n\n")
	lines += 2

	return lines
}

func generateBenchmarkFunction(sb *strings.Builder, index int) int {
	lines := 0

	sb.WriteString(fmt.Sprintf("async def async_operation_%d(\n", index))
	lines++
	sb.WriteString("    client: AsyncClient,\n")
	lines++
	sb.WriteString("    request: Request[T],\n")
	lines++
	sb.WriteString("    timeout: Optional[float] = None\n")
	lines++
	sb.WriteString(") -> Response[T]:\n")
	lines++
	sb.WriteString("    \"\"\"Perform async operation.\"\"\"\n")
	lines++
	sb.WriteString("    try:\n")
	lines++
	sb.WriteString("        result = await client.execute(request, timeout=timeout)\n")
	lines++
	sb.WriteString("        return Response(success=True, data=result)\n")
	lines++
	sb.WriteString("    except Exception as e:\n")
	lines++
	sb.WriteString("        logger.error(f\"Operation failed: {e}\")\n")
	lines++
	sb.WriteString("        return Response(success=False, error=str(e))\n\n")
	lines += 2

	return lines
}

func generateBenchmarkDataStructure(sb *strings.Builder, index int) int {
	lines := 0

	if index%2 == 0 {
		// Generate dataclass
		sb.WriteString(fmt.Sprintf("@dataclass\n"))
		lines++
		sb.WriteString(fmt.Sprintf("class DataModel%d:\n", index))
		lines++
		sb.WriteString("    \"\"\"Data model with type annotations.\"\"\"\n")
		lines++
		sb.WriteString("    id: int\n")
		lines++
		sb.WriteString("    name: str\n")
		lines++
		sb.WriteString("    tags: List[str] = field(default_factory=list)\n")
		lines++
		sb.WriteString("    metadata: Dict[str, Any] = field(default_factory=dict)\n")
		lines++
		sb.WriteString("    active: bool = True\n")
		lines++
		sb.WriteString("    \n")
		lines++
		sb.WriteString("    def validate(self) -> bool:\n")
		lines++
		sb.WriteString("        return bool(self.name)\n\n")
		lines += 2
	} else {
		// Generate enum
		sb.WriteString(fmt.Sprintf("class Status%d(Enum):\n", index))
		lines++
		sb.WriteString("    \"\"\"Status enumeration.\"\"\"\n")
		lines++
		sb.WriteString("    PENDING = \"pending\"\n")
		lines++
		sb.WriteString("    ACTIVE = \"active\"\n")
		lines++
		sb.WriteString("    COMPLETED = \"completed\"\n")
		lines++
		sb.WriteString("    FAILED = \"failed\"\n\n")
		lines += 2
	}

	return lines
}
