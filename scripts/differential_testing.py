#!/usr/bin/env python3
"""
Differential testing script for AI Distiller Python parser.
Compares parsing results against Python's ast module on large codebases.
"""

import ast
import json
import os
import re
import subprocess
import sys
import time
from collections import defaultdict
from pathlib import Path
from typing import Dict, List, Tuple, Optional
import argparse
import traceback

class DifferentialTester:
    def __init__(self, aid_binary: str = "./aid"):
        self.aid_binary = aid_binary
        self.results = {
            "total_files": 0,
            "parsed_successfully": 0,
            "parsing_failed": 0,
            "structural_matches": 0,
            "structural_mismatches": 0,
            "crashes": 0,
            "errors_by_type": defaultdict(int),
            "problem_files": [],
            "timing": {
                "total_time": 0,
                "python_ast_time": 0,
                "aid_time": 0
            }
        }
    
    def test_repository(self, repo_path: Path) -> Dict:
        """Test all Python files in a repository."""
        print(f"\n{'='*60}")
        print(f"Testing repository: {repo_path.name}")
        print(f"{'='*60}\n")
        
        python_files = list(repo_path.rglob("*.py"))
        print(f"Found {len(python_files)} Python files")
        
        for i, py_file in enumerate(python_files, 1):
            if i % 10 == 0:
                print(f"Progress: {i}/{len(python_files)} files processed...")
            
            self.test_file(py_file)
        
        return self.results
    
    def test_file(self, file_path: Path) -> bool:
        """Test a single Python file."""
        self.results["total_files"] += 1
        
        try:
            # Get Python AST
            start = time.time()
            python_ast = self.get_python_ast(file_path)
            self.results["timing"]["python_ast_time"] += time.time() - start
            
            if python_ast is None:
                self.results["parsing_failed"] += 1
                return False
            
            # Get AI Distiller output
            start = time.time()
            aid_output = self.get_aid_output(file_path)
            self.results["timing"]["aid_time"] += time.time() - start
            
            if aid_output is None:
                self.results["crashes"] += 1
                self.results["problem_files"].append({
                    "file": str(file_path),
                    "error": "AI Distiller crashed"
                })
                return False
            
            # Compare structures
            if self.compare_structures(python_ast, aid_output, file_path):
                self.results["structural_matches"] += 1
                self.results["parsed_successfully"] += 1
                return True
            else:
                self.results["structural_mismatches"] += 1
                self.results["parsed_successfully"] += 1
                return False
                
        except Exception as e:
            self.results["errors_by_type"][type(e).__name__] += 1
            self.results["problem_files"].append({
                "file": str(file_path),
                "error": str(e),
                "traceback": traceback.format_exc()
            })
            return False
    
    def get_python_ast(self, file_path: Path) -> Optional[Dict]:
        """Parse file with Python's ast module."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                source = f.read()
            
            tree = ast.parse(source, str(file_path))
            
            # Extract structure
            structure = {
                "functions": [],
                "classes": [],
                "async_functions": [],
                "decorators": set()
            }
            
            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    structure["functions"].append({
                        "name": node.name,
                        "args": len(node.args.args),
                        "lineno": node.lineno
                    })
                    for decorator in node.decorator_list:
                        if isinstance(decorator, ast.Name):
                            structure["decorators"].add(decorator.id)
                
                elif isinstance(node, ast.AsyncFunctionDef):
                    structure["async_functions"].append({
                        "name": node.name,
                        "args": len(node.args.args),
                        "lineno": node.lineno
                    })
                
                elif isinstance(node, ast.ClassDef):
                    methods = []
                    for item in node.body:
                        if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
                            methods.append(item.name)
                    
                    structure["classes"].append({
                        "name": node.name,
                        "methods": methods,
                        "bases": len(node.bases),
                        "lineno": node.lineno
                    })
            
            return structure
            
        except SyntaxError as e:
            # Python couldn't parse it either
            return None
        except Exception as e:
            print(f"Error parsing {file_path} with ast: {e}")
            return None
    
    def get_aid_output(self, file_path: Path) -> Optional[Dict]:
        """Parse file with AI Distiller."""
        try:
            # Run aid with markdown output (json-structured not implemented yet)
            cmd = [self.aid_binary, str(file_path), "--format", "md", "--stdout"]
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=5)
            
            if result.returncode != 0:
                return None
            
            # Parse Markdown output
            structure = {
                "functions": [],
                "classes": [],
                "async_functions": [],
                "decorators": set()
            }
            
            # Parse markdown format
            lines = result.stdout.strip().split('\n')
            current_class = None
            
            for line in lines:
                line = line.strip()
                
                # Skip empty lines and headers
                if not line or line.startswith('#') or line.startswith('='):
                    continue
                
                # Parse class: 🏛️ **Class** `ClassName`
                if '🏛️' in line and '**Class**' in line:
                    match = re.search(r'`([^`]+)`', line)
                    if match:
                        class_name = match.group(1)
                        current_class = {
                            "name": class_name,
                            "methods": [],
                            "bases": 0,  # Can't determine from markdown
                            "lineno": self._extract_line_number(line)
                        }
                        structure["classes"].append(current_class)
                
                # Parse function: 🔧 **Function** `function_name`
                elif '🔧' in line and '**Function**' in line:
                    match = re.search(r'`([^`]+)`', line)
                    if match:
                        func_name = match.group(1)
                        # Check if it's async
                        is_async = '_async_' in line or 'async' in line.lower()
                        
                        func_info = {
                            "name": func_name,
                            "args": 0,  # Can't determine from markdown easily
                            "lineno": self._extract_line_number(line)
                        }
                        
                        if current_class:
                            current_class["methods"].append(func_name)
                        elif is_async:
                            structure["async_functions"].append(func_info)
                        else:
                            structure["functions"].append(func_info)
            
            return structure
            
        except subprocess.TimeoutExpired:
            print(f"Timeout parsing {file_path} with aid")
            return None
        except Exception as e:
            print(f"Error parsing {file_path} with aid: {e}")
            return None
    
    def _extract_line_number(self, line: str) -> int:
        """Extract line number from markdown line like <sub>L123</sub>"""
        match = re.search(r'<sub>L(\d+)', line)
        if match:
            return int(match.group(1))
        return 0
    
    def compare_structures(self, python_ast: Dict, aid_output: Dict, file_path: Path) -> bool:
        """Compare structures from Python AST and AI Distiller."""
        mismatches = []
        
        # Compare function counts
        py_func_count = len(python_ast["functions"]) + len(python_ast["async_functions"])
        aid_func_count = len(aid_output["functions"]) + len(aid_output["async_functions"])
        
        if py_func_count != aid_func_count:
            mismatches.append(f"Function count: Python {py_func_count}, AID {aid_func_count}")
        
        # Compare class counts
        if len(python_ast["classes"]) != len(aid_output["classes"]):
            mismatches.append(f"Class count: Python {len(python_ast['classes'])}, AID {len(aid_output['classes'])}")
        
        # Compare function names
        py_func_names = {f["name"] for f in python_ast["functions"] + python_ast["async_functions"]}
        aid_func_names = {f["name"] for f in aid_output["functions"] + aid_output["async_functions"]}
        
        missing_in_aid = py_func_names - aid_func_names
        extra_in_aid = aid_func_names - py_func_names
        
        if missing_in_aid:
            mismatches.append(f"Missing functions: {missing_in_aid}")
        if extra_in_aid:
            mismatches.append(f"Extra functions: {extra_in_aid}")
        
        # Compare class names
        py_class_names = {c["name"] for c in python_ast["classes"]}
        aid_class_names = {c["name"] for c in aid_output["classes"]}
        
        missing_classes = py_class_names - aid_class_names
        extra_classes = aid_class_names - py_class_names
        
        if missing_classes:
            mismatches.append(f"Missing classes: {missing_classes}")
        if extra_classes:
            mismatches.append(f"Extra classes: {extra_classes}")
        
        if mismatches:
            self.results["problem_files"].append({
                "file": str(file_path),
                "mismatches": mismatches,
                "python_ast": python_ast,
                "aid_output": aid_output
            })
            return False
        
        return True
    
    def print_summary(self):
        """Print test results summary."""
        print(f"\n{'='*60}")
        print("DIFFERENTIAL TESTING SUMMARY")
        print(f"{'='*60}\n")
        
        print(f"Total files tested: {self.results['total_files']}")
        print(f"Successfully parsed: {self.results['parsed_successfully']}")
        print(f"Python parse failures: {self.results['parsing_failed']}")
        print(f"AI Distiller crashes: {self.results['crashes']}")
        print(f"Structural matches: {self.results['structural_matches']}")
        print(f"Structural mismatches: {self.results['structural_mismatches']}")
        
        if self.results['parsed_successfully'] > 0:
            accuracy = (self.results['structural_matches'] / self.results['parsed_successfully']) * 100
            print(f"\nStructural accuracy: {accuracy:.2f}%")
        
        print(f"\nTiming:")
        print(f"  Python AST time: {self.results['timing']['python_ast_time']:.2f}s")
        print(f"  AI Distiller time: {self.results['timing']['aid_time']:.2f}s")
        
        if self.results['errors_by_type']:
            print(f"\nErrors by type:")
            for error_type, count in self.results['errors_by_type'].items():
                print(f"  {error_type}: {count}")
        
        # Save detailed results
        results_file = Path("differential_testing_results.json")
        with open(results_file, 'w') as f:
            json.dump(self.results, f, indent=2, default=str)
        print(f"\nDetailed results saved to: {results_file}")


def clone_repository(repo_url: str, target_dir: Path) -> Path:
    """Clone a repository if it doesn't exist."""
    repo_name = repo_url.split('/')[-1].replace('.git', '')
    repo_path = target_dir / repo_name
    
    if repo_path.exists():
        print(f"Repository {repo_name} already exists at {repo_path}")
    else:
        print(f"Cloning {repo_url}...")
        subprocess.run(["git", "clone", repo_url, str(repo_path)], check=True)
    
    return repo_path


def main():
    parser = argparse.ArgumentParser(description="Differential testing for AI Distiller Python parser")
    parser.add_argument("--aid-binary", default="./aid", help="Path to aid binary")
    parser.add_argument("--repo", help="Repository URL or path to test")
    parser.add_argument("--repos-dir", default="test_repos", help="Directory for test repositories")
    args = parser.parse_args()
    
    # Create repos directory
    repos_dir = Path(args.repos_dir)
    repos_dir.mkdir(exist_ok=True)
    
    # Build aid binary if needed
    if not Path(args.aid_binary).exists():
        print("Building aid binary...")
        subprocess.run(["make", "build"], check=True)
    
    tester = DifferentialTester(args.aid_binary)
    
    if args.repo:
        # Test single repository
        if args.repo.startswith("http"):
            repo_path = clone_repository(args.repo, repos_dir)
        else:
            repo_path = Path(args.repo)
        
        tester.test_repository(repo_path)
    else:
        # Test default set of repositories
        default_repos = [
            "https://github.com/psf/requests.git",
            "https://github.com/pallets/flask.git",
            "https://github.com/python-attrs/attrs.git",
        ]
        
        for repo_url in default_repos:
            repo_path = clone_repository(repo_url, repos_dir)
            tester.test_repository(repo_path)
    
    tester.print_summary()


if __name__ == "__main__":
    main()