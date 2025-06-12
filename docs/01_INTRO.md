# **Project Brief: AI Distiller**

## **1. Vision & Mission**

**Mission:** To develop a blazingly fast, cross-platform command-line utility, aid, written in Go. This tool will intelligently "distill" source code from any project into a compact, structured format, optimized for the context window of Large Language Models (LLMs).

**Core Problem:** Large and complex codebases overwhelm the context windows of modern AIs, preventing them from understanding the overall architecture, public APIs, and key data structures. **AI Distiller** solves this by creating a high-density "blueprint" of the code, enabling AI to reason about large projects effectively.

**Target Audience:** Developers, AI engineers, and technical writers who use AI assistants (like Claude, Zen (Gemini), ChatGPT) for code analysis, documentation generation, refactoring, and onboarding.

## **2. Core Principles & Mandates**

This project will be governed by the following non-negotiable principles:

* **English-Only Domain:** All code, comments, documentation, commit messages, and discussions will be in English.  
* **Single Native Binary:** The ultimate deliverable is a single, dependency-free executable for each target platform (Windows, macOS, Linux) and architecture (x86_64, ARM64). No runtime installation (Node, Python, etc.) shall be required for the end-user.  
* **Architecture First:** The initial architectural design is paramount. No new source files or major structural changes are permitted without following the collaboration protocol.  
* **Test-Driven Development (TDD):** Every feature must be accompanied by comprehensive unit and integration tests. The test suite must pass before any code is committed.  
* **User-Centric CLI Design:** The command-line interface must be intuitive, self-documenting, and follow standard POSIX conventions.  
* **Performance is a Feature:** The tool must be extremely fast, with minimal startup time and efficient file processing.

## **3. AI Collaboration & Governance Model**

This project will be developed by a team of two AI agents: **Claude** and **Zen** (Gemini).

* **Claude (Primary Developer):** Claude is responsible for writing the code, tests, and documentation based on the defined tasks.  
* **Zen (Architect & Reviewer):** Zen acts as the lead architect and reviewer.  
  * **Oponentura (Challenge & Review):** Claude **must** submit all non-trivial proposals (e.g., choice of a third-party library, complex algorithm design, API changes) to Zen for review and constructive criticism.  
  * **Architectural Lock:** Any deviation from the initially approved architecture (e.g., adding a new source file) **must** be formally proposed to Zen. The proposal must include the rationale and the original architectural plan. Zen's approval is required to proceed. Zen may suggest implementing the functionality within the existing structure.  
* **External Knowledge:** Both agents are encouraged to use Context7 (hypothetical MCP tool) to fetch the latest documentation for languages, frameworks, and libraries to ensure the parsing logic is up-to-date.  
* **Project Repository:** The single source of truth is the GitHub repository: https://github.com/janreges/ai-distiller

## **4. Phase 1: System Architecture & Foundation**

The first and most critical phase is to design a robust and scalable architecture.

### **Task 1.1: Detailed Architecture Document**

The AI team will produce a DESIGN_ARCHITECTURE.md file containing:

1. **Directory & File Structure:** A complete tree of all planned source code files and their specific responsibilities.  
2. **Data Flow Diagram:** A visual or text-based diagram showing how data flows from a user's CLI command, through the parsers, to the final output file or stdout.  
3. **Core Data Structures:** Definition of the key Go structs (e.g., CliOptions, FileInfo, ProcessingResult).  
4. **Language Processor Interface:** A formal Go interface that every language-specific parser must implement. This ensures consistency.  
5. **Error Handling Strategy:** A defined approach for propagating and reporting errors to the user via stderr.  
6. **Configuration Management:** How the tool will manage internal settings and language-specific rules.

### **Task 1.2: Proposed Directory Structure (Initial Draft for Discussion)**

/  
├── cmd/  
│   └── aid/  
│       └── main.go           # Entry point, CLI parsing, orchestrates the main logic  
├── internal/  
│   ├── distiller/  
│   │   ├── distiller.go      # Core distilling logic, file traversal, orchestration  
│   │   └── distiller_test.go  
│   ├── parser/  
│   │   ├── parser.go         # Dispatches to the correct language processor  
│   │   ├── parser_test.go  
│   │   └── interface.go      # Defines the LanguageProcessor interface  
│   ├── language/             # Each supported language gets its own file  
│   │   ├── go.go  
│   │   ├── go_test.go  
│   │   ├── python.go  
│   │   ├── python_test.go  
│   │   ├── javascript.go  
│   │   ├── javascript_test.go  
│   │   └── ...               # etc. for all languages  
│   └── config/  
│       └── config.go         # Handles language definitions, file extensions etc.  
├── pkg/  
│   └── utils/  
│       └── utils.go          # Helper functions (e.g., file I/O, string manipulation)  
├── testdata/                 # Sample source files for all languages for testing  
│   ├── python/  
│   │   ├── simple_class.py  
│   │   └── complex_module.py  
│   └── ...  
├── docs/  
│   ├── README.md             # Main project README  
│   └── integration_guide.md  # Guide for integrating `aid` as an MCP tool  
├── scripts/  
│   └── install.sh            # Universal installer script  
├── .github/  
│   └── workflows/  
│       └── release.yml       # GitHub Actions for CI/CD  
└── Makefile                  # Makefile for common dev tasks

## **5. Core Functionality & CLI Specification**

### **5.1. CLI Help Output (aid --help)**

The CLI must provide a clear, comprehensive help message.

AI Distiller (aid) - A smart source code summarizer for LLMs.

USAGE:  
  aid [path] [flags]

ARGUMENTS:  
  [path]   Path to the source directory or file. (default: current directory)

FLAGS:  
  -o, --output <file>         Path to the output file. (default: .<dir_name>.[options].aid.txt)  
      --stdout                Print output to stdout in addition to writing to a file.  
      --strip <items>         Comma-separated list of items to strip: 'comments', 'imports', 'whitespaces', 'newlines', 'implementation', 'non-public'  
      --include <glob>        Glob pattern for files to include (e.g., "*.go,*.js"). (default: all supported)  
      --exclude <glob>        Glob pattern for files to exclude (e.g., "*_test.go,node_modules/*").  
  -r, --recursive             Process directories recursively. (default: true)  
      --absolute-paths        Use absolute paths in the output instead of relative ones.  
  -v, --verbose               Enable verbose logging (-v, -vv, -vvv for more detail).  
  -h, --help                  Display this help message.

EXAMPLES:  
  # Distill the current directory, removing implementation and comments  
  aid --strip implementation,comments

  # Distill a specific Python project into a named file  
  aid ./my-python-app -o context.txt --include "*.py"

  # Distill only the public interface of a Typescript file to stdout  
  aid src/api.ts --strip implementation,non-public --stdout

### **5.2. Output Format**

The output must be a well-structured text file, using a simple XML-like format to encapsulate file contents.

<aid-distiller-output version="1.0" source="/path/to/project">  
  <file path="src/main.go">  
    <!-- ... distilled content of main.go ... -->  
  </file>  
  <file path="internal/parser/parser.go">  
    <!-- ... distilled content of parser.go ... -->  
  </file>  
</aid-distiller-output>

### **5.3. Default Behaviors**

* **No Path Argument:** Use the current working directory (.).  
* **No Output Argument:** Generate filename automatically in the format .DIRECTORY_NAME.[flags].aid.txt. Example: aid --strip comments,implementation in folder MyProject creates .MyProject.ncom.nimpl.aid.txt. The flag abbreviations must be deterministic and documented.  
* **Processing:** Default to recursive.

## **6. Language Support & Testing**

### **6.1. Supported Languages & Frameworks**

The tool must support parsing for the following (detection via file extension):  
Go, Python, PHP, Rust, JavaScript, TypeScript, Java, Kotlin, Swift, Objective-C, C#, C++, C, CSS, .NET (via C#), Ruby, Markdown, and framework-specific constructs for React (.jsx, .tsx), Svelte (.svelte), Vue (.vue), Angular (.ts).

### **6.2. Testing Strategy**

* The testdata directory will contain at least two source files for each supported language:  
  1. **Simple Case:** A file with a single class or a few functions.  
  2. **Complex Case:** A file with multiple classes, nested functions, complex signatures, and framework-specific syntax (e.g., a React component with hooks).  
* For each test file, unit tests must assert the correct output for every relevant combination of --strip options.

## **7. Tooling & Automation**

* **Makefile:** Provide targets for make build, make test, make lint, make run, make clean.  
* **GitHub Actions (release.yml):**  
  1. Trigger on new Git tags (e.g., v1.0.0).  
  2. Set up Go environment.  
  3. Run linter and tests.  
  4. Build binaries for windows/amd64, windows/arm64, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64.  
  5. Create a GitHub Release and upload all binaries and the install.sh script.  
* **Installer Script (install.sh):** A shell script that detects the user's OS and architecture, downloads the correct binary from the latest GitHub release, and places it in /usr/local/bin (or equivalent).

## **8. Documentation**

* **README.md:** Must be comprehensive, including a project summary, key features, installation instructions (for all platforms), detailed usage examples, and a contribution guide.  
* **./docs/integration_guide.md:** A specific guide explaining how another AI or tool could use aid as a function/tool by executing the CLI command and consuming its structured output.

## **9. Initial Task List (High-Level)**

1. [ ] **Phase 1:** Finalize and commit DESIGN_ARCHITECTURE.md after Claude/Zen consultation.  
2. [ ] **Phase 1:** Set up the Go project structure and Makefile.  
3. [ ] **Phase 1:** Implement the main.go with CLI flag parsing (using cobra or flag package) and help output.  
4. [ ] **Phase 1:** Create the core distiller logic for file traversal (recursive/non-recursive, include/exclude).  
5. [ ] **Phase 1:** Define the LanguageProcessor interface in internal/parser/interface.go.  
6. [ ] **Phase 2:** Implement the output generation logic (file or stdout) with the XML-like structure.  
7. [ ] **Phase 2:** Implement dynamic output filename generation.  
8. [ ] **Phase 2:** Implement verbosity levels using a logging library (e.g., slog).  
9. [ ] **Phase 3:** Implement the language processor for **Python** (python.go).  
10. [ ] **Phase 3:** Create comprehensive unit tests for the Python processor using testdata.  
11. [ ] **Phase 3:** Implement the language processor for **JavaScript/TypeScript**.  
12. [ ] **Phase 3:** Create comprehensive unit tests for the TS processor.  
13. [ ] **Phase 3:** Implement the language processor for **Go**.  
14. [ ] **Phase 3:** Create comprehensive unit tests for the Go processor.  
15. [ ] **Phase 3:** ... (continue for all other languages) ...  
16. [ ] **Phase 4:** Create the release.yml GitHub Actions workflow.  
17. [ ] **Phase 4:** Write and test the install.sh universal installer.  
18. [ ] **Phase 5:** Write the main README.md.  
19. [ ] **Phase 5:** Write the integration_guide.md.  
20. [ ] **Phase 5:** Final review of all code and documentation before the first release.

