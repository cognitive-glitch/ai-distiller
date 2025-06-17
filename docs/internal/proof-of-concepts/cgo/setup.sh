#!/bin/bash
set -e

echo "Setting up CGo PoC..."

# Create grammars directory
mkdir -p grammars

# Clone tree-sitter-python if not exists
if [ ! -d "grammars/tree-sitter-python" ]; then
    echo "Cloning tree-sitter-python..."
    git clone https://github.com/tree-sitter/tree-sitter-python grammars/tree-sitter-python
fi

# Build tree-sitter-python
cd grammars/tree-sitter-python/src
if [ ! -f "parser.c" ]; then
    echo "Error: parser.c not found. Tree-sitter-python structure may have changed."
    exit 1
fi

# Compile parser.c and scanner.c into object files
echo "Compiling tree-sitter-python..."
gcc -c -fPIC parser.c -o parser.o -I.
if [ -f "scanner.c" ]; then
    gcc -c -fPIC scanner.c -o scanner.o -I.
fi

# Create static library
ar rcs libtree-sitter-python.a parser.o scanner.o 2>/dev/null || ar rcs libtree-sitter-python.a parser.o

echo "Setup complete!"
echo "You can now run: make build"