# AI Distiller CLI Arguments - Comprehensive Test Task List

## Critical Issues Found

### 🚨 HIGH PRIORITY BUGS
1. **SEGMENTATION FAULT**: `aid --help-extended` crashes with segfault - CRITICAL stability issue
2. **Missing help topics**: Several help commands mentioned but need verification

## Test Setup Requirements

Create standardized test project structure:
```
test_project/
├── .git/                  # Git repo with 10+ commits
├── src/
│   ├── main.py           # Python: public/private functions, comments, docstrings, imports
│   ├── component.ts      # TypeScript: public/protected/internal, decorators
│   ├── utils.go          # Go: exported/unexported functions
│   └── Service.java      # Java: public/private/protected methods
├── docs/
│   └── README.md
├── config.json
└── .gitignore
```

## 1. Core I/O and Execution

### Path Arguments
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Directory processing | `aid ./test_project/src` | Processes all files in src/, creates .aid/ output | ⏳ |
| Single file | `aid ./test_project/src/main.py` | Processes only main.py, creates .aid.main.py.txt | ⏳ |
| Current directory | `aid .` | Processes current directory files | ⏳ |
| Absolute path | `aid /full/path/to/project` | Processes using absolute path | ⏳ |

### Output Control
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Custom output file | `aid ./test_project/src -o custom.txt` | All output in custom.txt, no .aid/ folder | ⏳ |
| Stdout only | `aid ./test_project/src/main.py --stdout` | Prints to console, verify file behavior | ⏳ |
| Stdout + file | `aid ./test_project/src -o out.txt --stdout` | Output to BOTH out.txt AND stdout | ⏳ |
| Default naming | `aid ./test_project/src` | Creates .aid.src.pub.txt (default options) | ⏳ |

### Format Options
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Text format (default) | `aid ./test_project/src --format text` | Compact text with `<file path="...">` tags | ⏳ |
| Markdown format | `aid ./test_project/src --format md` | Markdown with emojis, headers, structure | ⏳ |
| JSONL format | `aid ./test_project/src --format jsonl` | One JSON object per line | ⏳ |
| JSON structured | `aid ./test_project/src --format json-structured` | Single well-formed JSON object | ⏳ |
| XML format | `aid ./test_project/src --format xml` | Valid XML structure | ⏳ |

### Worker Control
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Auto workers (default) | `aid ./test_project/src -w 0` | Uses automatic worker count | ⏳ |
| Serial processing | `aid ./test_project/src -w 1` | Single-threaded processing | ⏳ |
| Multiple workers | `aid ./test_project/src -w 4` | Uses 4 worker threads | ⏳ |

### Version Info
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Version display | `aid --version` | Shows version string and exits | ⏳ |

## 2. AI Actions (10 total)

### Prompt Generation Actions
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Refactoring prompt | `aid ./test_project/src --ai-action prompt-for-refactoring-suggestion` | Creates REFACTORING-SUGGESTION.timestamp.md | ⏳ |
| Complex codebase | `aid ./test_project --ai-action prompt-for-complex-codebase-analysis` | Creates COMPLEX-CODEBASE-ANALYSIS.timestamp.md | ⏳ |
| Security analysis | `aid ./test_project/src --ai-action prompt-for-security-analysis` | Creates SECURITY-ANALYSIS.timestamp.md | ⏳ |
| Performance analysis | `aid ./test_project/src --ai-action prompt-for-performance-analysis` | Creates PERFORMANCE-ANALYSIS.timestamp.md | ⏳ |
| Best practices | `aid ./test_project/src --ai-action prompt-for-best-practices-analysis` | Creates BEST-PRACTICES-ANALYSIS.timestamp.md | ⏳ |
| Bug hunting | `aid ./test_project/src --ai-action prompt-for-bug-hunting` | Creates BUG-HUNTING-ANALYSIS.timestamp.md | ⏳ |
| Single file docs | `aid ./test_project/src/main.py --ai-action prompt-for-single-file-docs` | Creates SINGLE-FILE-DOCS.timestamp.md | ⏳ |
| Diagrams | `aid ./test_project/src --ai-action prompt-for-diagrams` | Creates DIAGRAMS-ANALYSIS.timestamp.md | ⏳ |

### Flow Generation Actions
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Deep analysis flow | `aid ./test_project/src --ai-action flow-for-deep-file-to-file-analysis` | Creates .aid/ directory with ANALYSIS-TASK-LIST.md | ⏳ |
| Multi-file docs flow | `aid ./test_project/src --ai-action flow-for-multi-file-docs` | Creates .aid/ directory with DOCS-TASK-LIST.md | ⏳ |

### AI Action Output Control
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Custom AI output | `aid ./test_project/src --ai-action prompt-for-security-analysis --ai-output security.md` | Output goes to security.md | ⏳ |
| Template variables | `aid ./test_project/src --ai-action prompt-for-refactoring-suggestion --ai-output "refactor-%%YYYY-MM-DD%%.md"` | Uses date in filename | ⏳ |

## 3. Visibility Filtering (0/1 syntax)

### Individual Visibility Flags
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Public only (default) | `aid ./test_project/src/main.py` | Shows only public functions/classes | ⏳ |
| Disable public | `aid ./test_project/src/main.py --public 0` | Hides public members (output mostly empty) | ⏳ |
| Enable private | `aid ./test_project/src/main.py --private 1` | Shows private members (underscored in Python) | ⏳ |
| Enable protected | `aid ./test_project/src/component.ts --protected 1` | Shows protected members in TypeScript | ⏳ |
| Enable internal | `aid ./test_project/src/component.ts --internal 1` | Shows internal/package-private members | ⏳ |
| All visibility levels | `aid ./test_project/src --public 1 --private 1 --protected 1 --internal 1` | Shows all visibility levels | ⏳ |

## 4. Content Filtering (0/1 syntax)

### Content Include/Exclude
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Include comments | `aid ./test_project/src/main.py --comments 1` | Shows regular comments (# in Python) | ⏳ |
| Exclude docstrings | `aid ./test_project/src/main.py --docstrings 0` | Hides docstrings (default: 1) | ⏳ |
| Include implementation | `aid ./test_project/src/main.py --implementation 1` | Shows function/method bodies | ⏳ |
| Exclude imports | `aid ./test_project/src/main.py --imports 0` | Hides import statements (default: 1) | ⏳ |
| Exclude annotations | `aid ./test_project/src/component.ts --annotations 0` | Hides decorators/annotations (default: 1) | ⏳ |

## 5. Alternative Filtering

### Category-based Filtering
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Include only specific | `aid ./test_project/src --include-only=private,comments` | Shows ONLY private members and comments | ⏳ |
| Exclude specific | `aid ./test_project/src --exclude-items=implementation,comments` | Excludes implementations and comments | ⏳ |
| Override precedence | `aid ./test_project/src --implementation 1 --exclude-items=implementation` | exclude-items should win over individual flags | ⏳ |

### Valid Categories
Test all valid category names:
- `public`, `protected`, `internal`, `private`
- `comments`, `docstrings`, `implementation`, `imports`, `annotations`

## 6. File Selection Patterns

### Include/Exclude Patterns
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Include specific types | `aid ./test_project --include="*.py,*.ts"` | Processes only Python and TypeScript files | ⏳ |
| Exclude patterns | `aid ./test_project/src --exclude="*.go"` | Excludes Go files | ⏳ |
| Complex patterns | `aid ./test_project --include="*.py" --exclude="*main*"` | Python files except those with 'main' | ⏳ |
| Multiple includes | `aid ./test_project --include "*.py" --include "*.ts"` | Test multiple --include flags | ⏳ |
| Multiple excludes | `aid ./test_project --exclude "*test*" --exclude "*.json"` | Test multiple --exclude flags | ⏳ |

### Precedence Testing
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Conflicting patterns | `aid ./test_project --include="*.py" --exclude="*.py"` | Should exclude win? No files processed? | ⏳ |

## 7. Special Processing Modes

### Raw Mode
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Raw processing | `aid ./test_project/src/main.py --raw` | Outputs raw text, no parsing | ⏳ |
| Raw ignores filters | `aid ./test_project/src/main.py --raw --private 0` | --private 0 should be ignored | ⏳ |

### Language Override
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Force language | `echo "func main() {}" | aid --lang go --stdout` | Parses stdin as Go code | ⏳ |
| Auto detection | `echo "def main():" | aid --lang auto --stdout` | Auto-detects as Python | ⏳ |

### Experimental Features
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Tree-sitter parser | `aid ./test_project/src --tree-sitter` | Uses tree-sitter parser (experimental) | ⏳ |

### Recursion Control
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Non-recursive | `aid ./test_project -r 0` | Processes only root files, not subdirectories | ⏳ |
| Recursive (default) | `aid ./test_project -r 1` | Processes all subdirectories | ⏳ |

## 8. Git History Mode

### Git Mode Activation
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Git history default | `aid .git` | Shows last 200 commits (default limit) | ⏳ |
| Limited commits | `aid .git --git-limit=5` | Shows only last 5 commits | ⏳ |
| All commits | `aid .git --git-limit=0` | Shows entire commit history | ⏳ |
| With AI prompt | `aid .git --git-limit=10 --with-analysis-prompt` | Includes comprehensive AI analysis prompt | ⏳ |

## 9. Path Control

### Path Format Control
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Relative paths (default) | `aid ./test_project/src --file-path-type relative` | Shows relative paths in output | ⏳ |
| Absolute paths | `aid ./test_project/src --file-path-type absolute` | Shows absolute paths in output | ⏳ |
| Custom prefix | `aid ./test_project/src --relative-path-prefix "source/"` | Prefixes relative paths with "source/" | ⏳ |

## 10. Help & Diagnostics

### Help Commands
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Basic help | `aid --help` | Shows basic help (works) | ✅ |
| Extended help | `aid --help-extended` | 🚨 CRASHES - segmentation fault | ❌ |
| Actions help | `aid help actions` | Shows AI actions documentation (works) | ✅ |
| Filtering help | `aid help filtering` | Shows filtering documentation | ⏳ |
| Git help | `aid help git` | Shows git mode documentation | ⏳ |
| Cheat sheet | `aid --cheat` | Shows quick reference card | ⏳ |

### Diagnostic Options
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Verbose level 1 | `aid ./test_project/src -v` | Shows basic processing info | ⏳ |
| Verbose level 2 | `aid ./test_project/src -vv` | Shows detailed processing info | ⏳ |
| Verbose level 3 | `aid ./test_project/src -vvv` | Shows full trace with data dumps | ⏳ |
| Strict mode | `aid ./broken_project --strict` | Fails on first syntax error | ⏳ |

## 11. Critical Interaction Tests

### Filter Precedence
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| include-only wins | `aid ./test_project/src --public 1 --include-only=private` | Only private shown, --public ignored | ⏳ |
| exclude-items wins | `aid ./test_project/src --implementation 1 --exclude-items=implementation` | Implementation excluded despite --implementation 1 | ⏳ |

### AI Action + Format
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| AI action format | `aid ./test_project/src --ai-action prompt-for-security-analysis --format jsonl` | Behavior unclear - needs clarification | ⏳ |

### Raw Mode Override
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Raw overrides all | `aid ./test_project/src --raw --implementation 0 --lang python` | Raw text output, all flags ignored | ⏳ |

## 12. Edge Cases & Error Handling

### Invalid Arguments
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Invalid format | `aid ./test_project/src --format invalid` | Error message, non-zero exit | ⏳ |
| Invalid AI action | `aid ./test_project/src --ai-action nonexistent` | Error message, non-zero exit | ⏳ |
| Invalid language | `aid ./test_project/src --lang cobol` | Error message, non-zero exit | ⏳ |

### File System Edge Cases
| Test | Command | Expected Result | Status |
|------|---------|----------------|--------|
| Non-existent path | `aid ./nonexistent` | Error message, non-zero exit | ⏳ |
| Permission denied | `aid /root` | Error message about permissions | ⏳ |
| Empty directory | `aid ./empty_dir` | Handles gracefully | ⏳ |

## Test Execution Strategy

### Phase 1: Critical Functionality (Priority High)
1. Fix segfault in --help-extended
2. Test all Core I/O options
3. Test all AI actions
4. Verify visibility and content filtering

### Phase 2: Advanced Features (Priority Medium)
1. Test file selection patterns
2. Test special modes
3. Test git mode
4. Test path control

### Phase 3: Edge Cases (Priority Low)
1. Test all interactions
2. Test error handling
3. Performance testing with large codebases

## Expected Findings

### Likely Issues
1. **CRITICAL**: --help-extended segfault
2. Filter precedence unclear/inconsistent
3. AI action + format interaction undefined
4. Some help topics might not exist
5. Error messages might be inconsistent

### Success Criteria
- All flags work as documented
- No crashes or segfaults
- Consistent error handling
- Clear precedence rules
- All help topics accessible

## Notes for Implementation

- Use `--stdout` for most tests to see output immediately
- Create comprehensive test project with multiple languages
- Document any discrepancies between help text and behavior
- File bug reports for crashes and inconsistencies
- Update CLAUDE.md with corrected documentation

---

**Generated**: 2025-06-19
**Status**: Ready for systematic testing
**Priority**: Fix segfault first, then systematic testing by priority