#!/usr/bin/env python3
"""
Generate large test files (10k+ lines) for performance testing.
"""

def generate_large_python(filename: str, num_classes: int = 500):
    """Generate large Python file with many classes."""
    with open(filename, 'w') as f:
        f.write('"""\n')
        f.write(f'Large Python file with {num_classes} classes.\n')
        f.write('Edge case: Performance testing with 10k+ lines.\n')
        f.write('"""\n\n')
        f.write('from typing import List, Optional, Dict, Any\n')
        f.write('from datetime import datetime\n\n')

        for i in range(num_classes):
            f.write(f'class TestClass{i}:\n')
            f.write(f'    """Test class number {i}."""\n\n')
            f.write(f'    def __init__(self, id: int, name: str):\n')
            f.write(f'        self.id = id\n')
            f.write(f'        self.name = name\n')
            f.write(f'        self._private_field = None\n\n')

            f.write(f'    def get_id(self) -> int:\n')
            f.write(f'        """Get the ID."""\n')
            f.write(f'        return self.id\n\n')

            f.write(f'    def set_name(self, name: str) -> None:\n')
            f.write(f'        """Set the name."""\n')
            f.write(f'        self.name = name\n\n')

            f.write(f'    @property\n')
            f.write(f'    def display_name(self) -> str:\n')
            f.write(f'        """Get display name."""\n')
            f.write(f'        return f"{{self.name}} ({{self.id}})""\n\n')

            f.write(f'    def _private_method(self) -> None:\n')
            f.write(f'        """Private method."""\n')
            f.write(f'        self._private_field = "updated"\n\n')

            f.write(f'    @classmethod\n')
            f.write(f'    def create(cls, name: str) -> "TestClass{i}":\n')
            f.write(f'        """Factory method."""\n')
            f.write(f'        return cls(id={i}, name=name)\n\n')

        f.write(f'\n# Total classes: {num_classes}\n')
        f.write(f'# Estimated lines: ~{num_classes * 20}\n')


def generate_large_typescript(filename: str, num_classes: int = 500):
    """Generate large TypeScript file with many classes."""
    with open(filename, 'w') as f:
        f.write('/**\n')
        f.write(f' * Large TypeScript file with {num_classes} classes.\n')
        f.write(' * Edge case: Performance testing with 10k+ lines.\n')
        f.write(' */\n\n')

        for i in range(num_classes):
            f.write(f'interface TestInterface{i} {{\n')
            f.write(f'    id: number;\n')
            f.write(f'    name: string;\n')
            f.write(f'}}\n\n')

            f.write(f'class TestClass{i} implements TestInterface{i} {{\n')
            f.write(f'    private _privateField: string | null = null;\n\n')

            f.write(f'    constructor(\n')
            f.write(f'        public id: number,\n')
            f.write(f'        public name: string\n')
            f.write(f'    ) {{}}\n\n')

            f.write(f'    public getId(): number {{\n')
            f.write(f'        return this.id;\n')
            f.write(f'    }}\n\n')

            f.write(f'    public setName(name: string): void {{\n')
            f.write(f'        this.name = name;\n')
            f.write(f'    }}\n\n')

            f.write(f'    public get displayName(): string {{\n')
            f.write(f'        return `${{this.name}} (${{this.id}})`;\n')
            f.write(f'    }}\n\n')

            f.write(f'    private privateMethod(): void {{\n')
            f.write(f'        this._privateField = "updated";\n')
            f.write(f'    }}\n\n')

            f.write(f'    public static create(name: string): TestClass{i} {{\n')
            f.write(f'        return new TestClass{i}({i}, name);\n')
            f.write(f'    }}\n')
            f.write(f'}}\n\n')

        f.write(f'\n// Total classes: {num_classes}\n')
        f.write(f'// Estimated lines: ~{num_classes * 20}\n')


def generate_large_go(filename: str, num_structs: int = 500):
    """Generate large Go file with many structs."""
    with open(filename, 'w') as f:
        f.write('/**\n')
        f.write(f' * Large Go file with {num_structs} structs.\n')
        f.write(' * Edge case: Performance testing with 10k+ lines.\n')
        f.write(' */\n')
        f.write('package main\n\n')

        for i in range(num_structs):
            f.write(f'// TestStruct{i} is test struct number {i}.\n')
            f.write(f'type TestStruct{i} struct {{\n')
            f.write(f'\tID          int\n')
            f.write(f'\tName        string\n')
            f.write(f'\tprivateField string\n')
            f.write(f'}}\n\n')

            f.write(f'// NewTestStruct{i} creates a new instance.\n')
            f.write(f'func NewTestStruct{i}(id int, name string) *TestStruct{i} {{\n')
            f.write(f'\treturn &TestStruct{i}{{\n')
            f.write(f'\t\tID:   id,\n')
            f.write(f'\t\tName: name,\n')
            f.write(f'\t}}\n')
            f.write(f'}}\n\n')

            f.write(f'// GetID returns the ID.\n')
            f.write(f'func (t *TestStruct{i}) GetID() int {{\n')
            f.write(f'\treturn t.ID\n')
            f.write(f'}}\n\n')

            f.write(f'// SetName sets the name.\n')
            f.write(f'func (t *TestStruct{i}) SetName(name string) {{\n')
            f.write(f'\tt.Name = name\n')
            f.write(f'}}\n\n')

            f.write(f'// DisplayName returns formatted name.\n')
            f.write(f'func (t *TestStruct{i}) DisplayName() string {{\n')
            f.write(f'\treturn fmt.Sprintf("%s (%d)", t.Name, t.ID)\n')
            f.write(f'}}\n\n')

            f.write(f'func (t *TestStruct{i}) privateMethod() {{\n')
            f.write(f'\tt.privateField = "updated"\n')
            f.write(f'}}\n\n')

        f.write(f'\n// Total structs: {num_structs}\n')
        f.write(f'// Estimated lines: ~{num_structs * 20}\n')


if __name__ == '__main__':
    import os

    # Generate large files
    print("Generating large test files...")

    generate_large_python('large_python.py', num_classes=500)
    print(f"  ✓ large_python.py (~10k lines)")

    generate_large_typescript('large_typescript.ts', num_classes=500)
    print(f"  ✓ large_typescript.ts (~10k lines)")

    generate_large_go('large_go.go', num_structs=500)
    print(f"  ✓ large_go.go (~10k lines)")

    print("\nDone! Files created in testdata/edge-cases/large-files/")
