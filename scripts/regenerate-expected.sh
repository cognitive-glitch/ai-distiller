#!/bin/bash

# Regenerate expected test files for a specific language
# This script should be run from the project root
#
# Usage: ./scripts/regenerate-expected.sh <language>
# Example: ./scripts/regenerate-expected.sh java

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if language argument is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Language argument is required${NC}"
    echo ""
    echo "Usage: $0 <language>"
    echo ""
    echo "Supported languages:"
    echo "  python, typescript, go, javascript, php, ruby"
    echo "  swift, rust, java, csharp, kotlin, cpp"
    echo ""
    echo "Special option:"
    echo "  really-all  - Regenerate ALL languages (use with caution)"
    exit 1
fi

LANGUAGE_ARG="$1"

# Ensure we're in the project root
if [ ! -f "go.mod" ] || [ ! -d "testdata" ]; then
    echo "Error: This script must be run from the project root directory"
    exit 1
fi

# Build the aid binary first
echo -e "${YELLOW}Building aid binary...${NC}"
go build -o aid cmd/aid/main.go

# Function to get file extension for a language
get_extension() {
    case $1 in
        python) echo "py" ;;
        typescript) echo "ts" ;;
        javascript) echo "js" ;;
        go) echo "go" ;;
        php) echo "php" ;;
        ruby) echo "rb" ;;
        swift) echo "swift" ;;
        rust) echo "rs" ;;
        java) echo "java" ;;
        csharp) echo "cs" ;;
        kotlin) echo "kt" ;;
        cpp) echo "cpp" ;;
        *) echo "$1" ;;
    esac
}

# Function to regenerate expected files for a language
regenerate_language() {
    local lang=$1
    local ext=$(get_extension "$lang")
    echo -e "${GREEN}Regenerating expected files for ${lang}...${NC}"

    # Check if language directory exists
    if [ ! -d "testdata/${lang}" ]; then
        echo "  Skipping ${lang} - directory not found"
        return
    fi

    # Process each test case
    for testdir in testdata/${lang}/*/; do
        if [ -d "$testdir" ]; then
            testname=$(basename "$testdir")
            echo "  Processing ${testname}..."

            # Create expected directory if it doesn't exist
            mkdir -p "${testdir}expected"

            # Find the source file
            sourcefile=""
            if [ -f "${testdir}source.${ext}" ]; then
                sourcefile="${testdir}source.${ext}"
            elif [ -f "${testdir}test.${ext}" ]; then
                sourcefile="${testdir}test.${ext}"
            elif [ -f "${testdir}main.${ext}" ]; then
                sourcefile="${testdir}main.${ext}"
            else
                # Try to find any source file with the right extension
                sourcefile=$(find "${testdir}" -maxdepth 1 -name "*.${ext}" | head -1)
            fi

            if [ -z "$sourcefile" ]; then
                echo "    Warning: No source file found in ${testdir}"
                continue
            fi

            # Generate default expected (all defaults)
            ./aid "$sourcefile" --stdout --format text > "${testdir}expected/default.txt"

            # Generate with implementation
            ./aid "$sourcefile" --stdout --format text --implementation=1 > "${testdir}expected/implementation=1.txt"

            # Generate without private members
            ./aid "$sourcefile" --stdout --format text --private=0 > "${testdir}expected/private=0.txt"

            # Generate without protected members
            ./aid "$sourcefile" --stdout --format text --protected=0 > "${testdir}expected/protected=0.txt"

            # Generate with only public members
            ./aid "$sourcefile" --stdout --format text --private=0 --protected=0 --internal=0 > "${testdir}expected/private=0,protected=0,internal=0.txt"

            # Generate with all visibility but no implementation
            ./aid "$sourcefile" --stdout --format text --private=1 --protected=1 --internal=1 --implementation=0 > "${testdir}expected/private=1,protected=1,internal=1,implementation=0.txt"

            # Generate with comments
            ./aid "$sourcefile" --stdout --format text --comments=1 > "${testdir}expected/comments=1.txt"

            # Generate without imports
            ./aid "$sourcefile" --stdout --format text --imports=0 > "${testdir}expected/imports=0.txt"

            # Remove old expected_*.txt files if they exist
            rm -f "${testdir}"expected_*.txt
        fi
    done
}

# List of supported languages
languages=(
    "python"
    "typescript"
    "go"
    "javascript"
    "php"
    "ruby"
    "swift"
    "rust"
    "java"
    "csharp"
    "kotlin"
    "cpp"
)

# Validate language argument
if [ "$LANGUAGE_ARG" = "really-all" ]; then
    echo -e "${YELLOW}WARNING: Regenerating ALL languages - this may overwrite many files!${NC}"
    echo "Press Ctrl+C within 3 seconds to cancel..."
    sleep 3

    # Process all languages
    for lang in "${languages[@]}"; do
        regenerate_language "$lang"
    done
    echo -e "${GREEN}All expected files regenerated successfully!${NC}"

elif [[ " ${languages[@]} " =~ " ${LANGUAGE_ARG} " ]]; then
    # Process single language
    regenerate_language "$LANGUAGE_ARG"
    echo -e "${GREEN}Expected files for ${LANGUAGE_ARG} regenerated successfully!${NC}"

else
    echo -e "${RED}Error: Unsupported language '${LANGUAGE_ARG}'${NC}"
    echo ""
    echo "Supported languages:"
    for lang in "${languages[@]}"; do
        echo "  $lang"
    done
    echo ""
    echo "Special option:"
    echo "  really-all  - Regenerate ALL languages (use with caution)"
    exit 1
fi