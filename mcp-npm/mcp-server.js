#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');
const readline = require('readline');

// Get the path to the aid binary
const binaryName = os.platform() === 'win32' ? 'aid.exe' : 'aid';
const binaryPath = path.join(__dirname, 'bin', binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error(`Error: AI Distiller binary not found at ${binaryPath}`);
  console.error('Please reinstall the package: npm install -g @janreges/ai-distiller-mcp');
  process.exit(1);
}

// MCP Server implementation
class MCPServer {
  constructor() {
    this.capabilities = {
      name: "AI Distiller MCP",
      version: "1.0.0",
      methods: [
        "tools/list",
        "tools/call"
      ]
    };
    
    this.tools = [
      // Core Analysis Engine
      {
        name: "aid_analyze",
        description: "Core AI Distiller analysis engine with automatic pagination for large outputs. Use specialized tools (aid_hunt_bugs, aid_suggest_refactoring, etc.) when available. This tool directly maps to aid --ai-action for advanced or custom analysis flows. Responses are automatically paginated when exceeding ~20000 tokens.\n\nIMPORTANT: This tool generates analysis files on disk. The output includes file paths to the generated analysis. For best results, read these files directly instead of trying to process the entire analysis through MCP responses.",
        inputSchema: {
          type: "object",
          properties: {
            ai_action: { 
              type: "string",
              enum: [
                "flow-for-deep-file-to-file-analysis",
                "flow-for-multi-file-docs",
                "prompt-for-refactoring-suggestion",
                "prompt-for-complex-codebase-analysis",
                "prompt-for-security-analysis",
                "prompt-for-performance-analysis",
                "prompt-for-best-practices-analysis",
                "prompt-for-bug-hunting",
                "prompt-for-single-file-docs",
                "prompt-for-diagrams"
              ]
            },
            target_path: { type: "string" },
            user_query: { type: "string" },
            output_format: { type: "string", enum: ["md", "text", "json"] },
            include_private: { type: "boolean" },
            include_implementation: { type: "boolean" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["ai_action", "target_path"]
        }
      },
      
      // Specialized Tools
      {
        name: "aid_hunt_bugs",
        description: "Systematically scans code files to identify potential bugs, logical errors, race conditions, and quality issues. Use when you suspect hidden bugs or want a comprehensive code health check. Returns detailed bug analysis with explanations and fix suggestions.\n\nOUTPUT: Generates a detailed markdown file with bug analysis. The response includes the file path - read this file directly for the complete analysis rather than processing through MCP pagination.",
        inputSchema: {
          type: "object",
          properties: {
            target_path: { type: "string" },
            focus_area: { type: "string" },
            include_private: { type: "boolean" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["target_path"]
        }
      },
      {
        name: "aid_suggest_refactoring",
        description: "Analyzes code to identify and suggest specific refactoring opportunities with concrete examples. Use to improve code quality, readability, maintainability, or performance. Returns actionable refactoring suggestions with before/after code examples.\n\nOUTPUT: Generates a comprehensive refactoring analysis markdown file. The response includes the file path - read this file directly for detailed suggestions with code examples.",
        inputSchema: {
          type: "object",
          properties: {
            target_path: { type: "string" },
            refactoring_goal: { type: "string" },
            include_implementation: { type: "boolean" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["target_path", "refactoring_goal"]
        }
      },
      {
        name: "aid_generate_diagram",
        description: "Generates architectural diagrams from source code using Mermaid format. Creates 10 beneficial diagrams including flowcharts, sequence diagrams, class diagrams, and architecture overviews. Perfect for understanding complex systems and documenting architecture.\n\nOUTPUT: Generates a markdown file with multiple Mermaid diagrams. The response includes the file path - read this file to view and render all diagrams.",
        inputSchema: {
          type: "object",
          properties: {
            target_path: { type: "string" },
            diagram_focus: { type: "string" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["target_path"]
        }
      },
      {
        name: "aid_analyze_security",
        description: "Performs comprehensive security analysis with OWASP Top 10 focus. Identifies potential vulnerabilities, security anti-patterns, and weak points. Use for security audits, compliance checks, or before production deployment. Returns security findings with risk levels and remediation steps.\n\nOUTPUT: Generates a detailed security audit markdown file. The response includes the file path - read this file for complete vulnerability analysis and remediation recommendations.",
        inputSchema: {
          type: "object",
          properties: {
            target_path: { type: "string" },
            security_focus: { type: "string" },
            include_private: { type: "boolean" },
            include_implementation: { type: "boolean" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["target_path"]
        }
      },
      {
        name: "aid_generate_docs",
        description: "Generates comprehensive documentation for source code including API references, usage examples, and developer guides. Creates structured documentation workflows for single files or entire projects. Perfect for creating technical documentation from code.\n\nOUTPUT: Generates one or more markdown documentation files. The response includes file paths - read these files directly for the complete documentation.",
        inputSchema: {
          type: "object",
          properties: {
            target_path: { type: "string" },
            doc_type: { type: "string", enum: ["single-file-docs", "multi-file-docs", "api-reference"] },
            audience: { type: "string" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" }
          },
          required: ["target_path"]
        }
      },
      
      // Legacy Tools (Backwards Compatibility)
      {
        name: "distill_file",
        description: "Extracts essential code structure from a single file. Legacy tool - prefer aid_analyze or specialized tools for new workflows.\n\nNOTE: For files that might exceed token limits, the tool will warn you. Consider using more restrictive parameters (include_implementation=false, include_private=false) or using aid_analyze tools that save results to files.",
        inputSchema: {
          type: "object",
          properties: {
            file_path: { type: "string" },
            include_private: { type: "boolean" },
            include_implementation: { type: "boolean" },
            include_comments: { type: "boolean" },
            output_format: { type: "string", enum: ["text", "md", "json"] }
          },
          required: ["file_path"]
        }
      },
      {
        name: "distill_directory",
        description: "Extracts code structure from directories with automatic pagination for large results. Returns paginated responses when content exceeds ~20000 tokens. Use page_token to get subsequent pages.\n\nCACHING STRATEGY for large codebases:\n- First page: Call with no_cache=true to ensure fresh data and populate cache\n- Subsequent pages: Use cached data (default) for consistency\n- Cache TTL: 5 minutes\n- Alternative: For very large analyses, consider using aid_analyze which saves results to files that can be read directly",
        inputSchema: {
          type: "object",
          properties: {
            directory_path: { type: "string" },
            recursive: { type: "boolean" },
            include_private: { type: "boolean" },
            include_implementation: { type: "boolean" },
            include_patterns: { type: "string" },
            exclude_patterns: { type: "string" },
            output_format: { type: "string", enum: ["text", "md", "json"] },
            page_size: { type: "number" },
            page_token: { type: "string" },
            no_cache: { type: "boolean" }
          },
          required: ["directory_path"]
        }
      },
      {
        name: "list_files",
        description: "Lists project files with language detection and statistics.",
        inputSchema: {
          type: "object",
          properties: {
            path: { type: "string" },
            pattern: { type: "string" },
            recursive: { type: "boolean" }
          }
        }
      },
      
      // Meta Tools
      {
        name: "get_capabilities",
        description: "Returns comprehensive information about AI Distiller capabilities, supported languages, and available tools.",
        inputSchema: {
          type: "object",
          properties: {}
        }
      }
    ];
  }
  
  async handleRequest(request) {
    const { jsonrpc, method, params, id } = request;
    
    if (jsonrpc !== "2.0") {
      return this.error(id, -32600, "Invalid Request");
    }
    
    switch (method) {
      case "initialize":
        return {
          jsonrpc: "2.0",
          id,
          result: {
            protocolVersion: "0.1.0",
            capabilities: this.capabilities
          }
        };
        
      case "tools/list":
        return {
          jsonrpc: "2.0",
          id,
          result: {
            tools: this.tools
          }
        };
        
      case "tools/call":
        return await this.handleToolCall(params, id);
        
      default:
        return this.error(id, -32601, "Method not found");
    }
  }
  
  async handleToolCall(params, id) {
    const { name, arguments: args } = params;
    
    try {
      let aidArgs = [];
      
      switch (name) {
        // Core Analysis Engine
        case "aid_analyze":
          aidArgs = [
            args.target_path,
            `--ai-action=${args.ai_action}`
          ];
          if (args.output_format) aidArgs.push(`--format=${args.output_format}`);
          if (args.include_private) aidArgs.push('--private=1', '--protected=1', '--internal=1');
          if (args.include_implementation) aidArgs.push('--implementation=1');
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          if (args.user_query) aidArgs.push(`--ai-query=${args.user_query}`);
          break;
          
        // Specialized Tools
        case "aid_hunt_bugs":
          aidArgs = [
            args.target_path,
            '--ai-action=prompt-for-bug-hunting'
          ];
          if (args.include_private !== false) aidArgs.push('--private=1', '--protected=1', '--internal=1');
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          break;
          
        case "aid_suggest_refactoring":
          aidArgs = [
            args.target_path,
            '--ai-action=prompt-for-refactoring-suggestion'
          ];
          if (args.include_implementation !== false) aidArgs.push('--implementation=1');
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          break;
          
        case "aid_generate_diagram":
          aidArgs = [
            args.target_path,
            '--ai-action=prompt-for-diagrams'
          ];
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          break;
          
        case "aid_analyze_security":
          aidArgs = [
            args.target_path,
            '--ai-action=prompt-for-security-analysis'
          ];
          if (args.include_private !== false) aidArgs.push('--private=1', '--protected=1', '--internal=1');
          if (args.include_implementation !== false) aidArgs.push('--implementation=1');
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          break;
          
        case "aid_generate_docs":
          aidArgs = [
            args.target_path,
            '--ai-action=' + (args.doc_type === 'single-file-docs' ? 'prompt-for-single-file-docs' : 
                              args.doc_type === 'api-reference' ? 'flow-for-multi-file-docs' : 
                              'flow-for-multi-file-docs')
          ];
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          break;
          
        // Legacy Tools
        case "distill_file":
          aidArgs = [args.file_path, '--stdout'];
          if (args.output_format) aidArgs.push(`--format=${args.output_format}`);
          if (args.include_private) aidArgs.push('--private=1', '--protected=1', '--internal=1');
          if (args.include_implementation) aidArgs.push('--implementation=1');
          if (args.include_comments) aidArgs.push('--comments=1');
          break;
          
        case "distill_directory":
          aidArgs = [args.directory_path, '--stdout'];
          if (args.output_format) aidArgs.push(`--format=${args.output_format}`);
          if (args.recursive === false) aidArgs.push('--recursive=0');
          if (args.include_private) aidArgs.push('--private=1', '--protected=1', '--internal=1');
          if (args.include_implementation) aidArgs.push('--implementation=1');
          if (args.include_patterns) aidArgs.push(`--include=${args.include_patterns}`);
          if (args.exclude_patterns) aidArgs.push(`--exclude=${args.exclude_patterns}`);
          // TODO: Pagination support would require state management
          break;
          
        case "list_files":
          // For list_files, we'll use a simple directory listing approach
          // since the aid binary doesn't have a direct list_files command
          const listResult = await this.listFiles(args);
          return {
            jsonrpc: "2.0",
            id,
            result: {
              content: [{
                type: "text",
                text: JSON.stringify(listResult, null, 2)
              }]
            }
          };
          
        case "get_capabilities":
          const capabilities = {
            server_name: "AI Distiller MCP",
            server_version: "1.0.0",
            root_path: process.env.AID_ROOT || process.cwd(),
            cache_dir: process.env.AID_CACHE_DIR || path.join(require('os').homedir(), '.cache', 'aid'),
            tools: {
              specialized: [
                "aid_hunt_bugs",
                "aid_suggest_refactoring",
                "aid_generate_diagram", 
                "aid_analyze_security",
                "aid_generate_docs"
              ],
              core: ["aid_analyze"],
              legacy: ["distill_file", "distill_directory", "list_files"],
              meta: ["get_capabilities"]
            },
            ai_actions: [
              "flow-for-deep-file-to-file-analysis",
              "flow-for-multi-file-docs",
              "prompt-for-refactoring-suggestion",
              "prompt-for-complex-codebase-analysis",
              "prompt-for-security-analysis",
              "prompt-for-performance-analysis",
              "prompt-for-best-practices-analysis",
              "prompt-for-bug-hunting",
              "prompt-for-single-file-docs",
              "prompt-for-diagrams"
            ],
            supported_languages: [
              "python", "typescript", "javascript", "go", "java",
              "csharp", "rust", "ruby", "swift", "kotlin", "php", "cpp", "c"
            ],
            supported_formats: ["text", "md", "json", "xml", "jsonl"],
            features: [
              "ai_actions", "pattern_filtering", "specialized_analysis",
              "diagram_generation", "security_analysis", "bug_hunting",
              "refactoring_suggestions", "documentation_generation"
            ]
          };
          return {
            jsonrpc: "2.0",
            id,
            result: {
              content: [{
                type: "text",
                text: JSON.stringify(capabilities, null, 2)
              }]
            }
          };
          
        default:
          return this.error(id, -32602, `Unknown tool: ${name}`);
      }
      
      const result = await this.runAid(aidArgs);
      
      return {
        jsonrpc: "2.0",
        id,
        result: {
          content: [{
            type: "text",
            text: result
          }]
        }
      };
    } catch (error) {
      return this.error(id, -32603, error.message);
    }
  }
  
  async listFiles(args) {
    const fs = require('fs').promises;
    const path = require('path');
    
    const basePath = path.resolve(args.path || '.');
    const pattern = args.pattern;
    const recursive = args.recursive !== false;
    
    const results = [];
    
    async function scan(dir) {
      try {
        const items = await fs.readdir(dir, { withFileTypes: true });
        
        for (const item of items) {
          const fullPath = path.join(dir, item.name);
          const relativePath = path.relative(basePath, fullPath);
          
          if (item.isFile()) {
            if (!pattern || relativePath.match(new RegExp(pattern))) {
              const stats = await fs.stat(fullPath);
              results.push({
                path: relativePath,
                size: stats.size,
                modified: stats.mtime.toISOString()
              });
            }
          } else if (item.isDirectory() && recursive && !item.name.startsWith('.')) {
            await scan(fullPath);
          }
        }
      } catch (err) {
        // Skip directories we can't read
      }
    }
    
    await scan(basePath);
    
    return {
      path: basePath,
      file_count: results.length,
      files: results
    };
  }
  
  runAid(args) {
    return new Promise((resolve, reject) => {
      const child = spawn(binaryPath, args, {
        cwd: process.env.AID_ROOT || process.cwd(),
        env: process.env
      });
      
      let stdout = '';
      let stderr = '';
      
      child.stdout.on('data', (data) => {
        stdout += data.toString();
      });
      
      child.stderr.on('data', (data) => {
        stderr += data.toString();
      });
      
      child.on('close', (code) => {
        if (code !== 0) {
          reject(new Error(stderr || `Aid exited with code ${code}`));
        } else {
          resolve(stdout);
        }
      });
      
      child.on('error', (err) => {
        reject(err);
      });
    });
  }
  
  error(id, code, message) {
    return {
      jsonrpc: "2.0",
      id,
      error: {
        code,
        message
      }
    };
  }
}

// Start the server
const server = new MCPServer();
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', async (line) => {
  try {
    const request = JSON.parse(line);
    const response = await server.handleRequest(request);
    console.log(JSON.stringify(response));
  } catch (error) {
    console.error(JSON.stringify({
      jsonrpc: "2.0",
      error: {
        code: -32700,
        message: "Parse error"
      }
    }));
  }
});

// Log to stderr for debugging
console.error(`AI Distiller MCP Server started
Binary: ${binaryPath}
Working directory: ${process.env.AID_ROOT || process.cwd()}`);