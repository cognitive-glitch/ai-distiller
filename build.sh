#!/bin/bash

# AI Distiller Build Script

set -e

echo "Building AI Distiller variants..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build flags
LDFLAGS='-s -w'

echo -e "${YELLOW}Building full version...${NC}"
go build -ldflags="$LDFLAGS" -trimpath -o aid cmd/aid/main.go
FULL_SIZE=$(ls -lh aid | awk '{print $5}')
echo -e "${GREEN}✓ Full version built: aid ($FULL_SIZE)${NC}"

echo -e "${YELLOW}Building lite version (without Kotlin, C#, C++, Java, Swift)...${NC}"
go build -tags lite -ldflags="$LDFLAGS" -trimpath -o aid-lite cmd/aid/main.go
LITE_SIZE=$(ls -lh aid-lite | awk '{print $5}')
echo -e "${GREEN}✓ Lite version built: aid-lite ($LITE_SIZE)${NC}"

echo -e "\n${YELLOW}Binary size comparison:${NC}"
echo "Full version: $FULL_SIZE"
echo "Lite version: $LITE_SIZE"

echo -e "\n${YELLOW}Languages supported:${NC}"
echo "Full version: Python, TypeScript, JavaScript, Go, Ruby, Swift, Rust, Java, C#, Kotlin, C++, PHP"
echo "Lite version: Python, TypeScript, JavaScript, Go, Ruby, Rust, PHP"

echo -e "\n${YELLOW}Recommendations:${NC}"
echo "1. Use 'aid-lite' for most common languages (saves ~17MB)"
echo "2. Use 'aid' only if you need Kotlin, C#, C++, Java, or Swift support"
echo "3. For further compression, consider using UPX: upx --best --lzma aid-lite"