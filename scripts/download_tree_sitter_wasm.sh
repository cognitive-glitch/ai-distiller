#!/bin/bash

# Download or build tree-sitter WASM modules

WASM_DIR="internal/parser/wasm"
mkdir -p "$WASM_DIR"

echo "Attempting to download tree-sitter Python WASM module..."

# Try different sources
URLS=(
    "https://github.com/tree-sitter/tree-sitter-python/releases/download/v0.20.4/tree-sitter-python.wasm"
    "https://unpkg.com/tree-sitter-python@0.20.1/tree-sitter-python.wasm"
    "https://cdn.jsdelivr.net/npm/tree-sitter-python@0.20.1/tree-sitter-python.wasm"
)

for url in "${URLS[@]}"; do
    echo "Trying: $url"
    curl -L -f -o "$WASM_DIR/tree-sitter-python.wasm" "$url" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        # Check if it's actually a WASM file
        if file "$WASM_DIR/tree-sitter-python.wasm" | grep -q "WebAssembly"; then
            echo "Successfully downloaded tree-sitter-python.wasm"
            ls -lh "$WASM_DIR/tree-sitter-python.wasm"
            exit 0
        else
            echo "Downloaded file is not a valid WASM module"
            rm -f "$WASM_DIR/tree-sitter-python.wasm"
        fi
    fi
done

echo "Direct download failed. Attempting to build from source..."

# Check if we have necessary tools
if ! command -v npm &> /dev/null; then
    echo "npm is required to build tree-sitter WASM. Please install Node.js/npm."
    exit 1
fi

if ! command -v emcc &> /dev/null; then
    echo "Emscripten (emcc) is required to build WASM. Trying with tree-sitter CLI instead..."
    
    # Install tree-sitter CLI via npm
    echo "Installing tree-sitter CLI..."
    npm install -g tree-sitter-cli
    
    # Clone and build
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    echo "Cloning tree-sitter-python..."
    git clone --depth 1 https://github.com/tree-sitter/tree-sitter-python.git
    cd tree-sitter-python
    
    echo "Building WASM module..."
    tree-sitter build-wasm
    
    if [ -f "tree-sitter-python.wasm" ]; then
        cp tree-sitter-python.wasm "$OLDPWD/$WASM_DIR/"
        cd "$OLDPWD"
        rm -rf "$TEMP_DIR"
        echo "Successfully built tree-sitter-python.wasm"
        ls -lh "$WASM_DIR/tree-sitter-python.wasm"
        exit 0
    else
        echo "Failed to build WASM module"
        cd "$OLDPWD"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
fi

echo "Unable to obtain tree-sitter-python.wasm"
exit 1