import os
import json
from pathlib import Path
from typing import Dict, List, Optional
import requests

class LocalLLMAnalyzer:
    """Analyze repository structure using a local LLM (Ollama)."""
    
    def __init__(self, model: str = "llama3.2", base_url: str = "http://localhost:11434"):
        """
        Args:
            model: Ollama model name (e.g., 'llama3.2', 'codellama', 'mistral')
            base_url: Ollama API base URL
        """
        self.model = model
        self.base_url = base_url
        self.api_url = f"{base_url}/api/generate"
    
    def call_llm(self, prompt: str) -> str:
        """
        Call the local LLM with a prompt.
        
        Args:
            prompt: The prompt to send to the LLM
            
        Returns:
            The LLM's response
        """
        try:
            payload = {
                "model": self.model,
                "prompt": prompt,
                "stream": False
            }
            
            response = requests.post(self.api_url, json=payload)
            response.raise_for_status()
            
            return response.json()["response"]
        
        except Exception as e:
            print(f"Error calling LLM: {e}")
            return f"Error: Could not analyze - {str(e)}"
    
    def read_file_content(self, file_path: Path, max_lines: int = 500) -> str:
        """
        Read file content with size limit.
        
        Args:
            file_path: Path to the file
            max_lines: Maximum number of lines to read
            
        Returns:
            File content as string
        """
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                lines = f.readlines()[:max_lines]
                content = ''.join(lines)
                
                if len(lines) == max_lines:
                    content += f"\n... (file truncated, showing first {max_lines} lines)"
                
                return content
        except Exception as e:
            return f"Error reading file: {str(e)}"
    
    def explain_file(self, file_path: Path, repo_root: Path) -> Dict:
        """
        Generate explanation for a single file.
        
        Args:
            file_path: Path to the file
            repo_root: Root path of the repository
            
        Returns:
            Dictionary with file explanation
        """
        rel_path = file_path.relative_to(repo_root)
        print(f"Analyzing file: {rel_path}")
        
        content = self.read_file_content(file_path)
        
        prompt = f"""Analyze this file and provide a concise explanation (2-3 sentences) of its purpose and functionality.

File: {rel_path}

Content:
{content}

Provide only the explanation, no additional commentary."""
        
        explanation = self.call_llm(prompt)
        
        return {
            "path": str(rel_path),
            "type": "file",
            "explanation": explanation.strip()
        }
    
    def explain_folder(self, folder_explanations: List[Dict], folder_path: str) -> Dict:
        """
        Generate explanation for a folder based on its contents.
        
        Args:
            folder_explanations: List of explanations for files/subfolders in this folder
            folder_path: Relative path to the folder
            
        Returns:
            Dictionary with folder explanation
        """
        print(f"Summarizing folder: {folder_path}")
        
        context = "\n\n".join([
            f"{item['path']}: {item['explanation']}"
            for item in folder_explanations
        ])
        
        prompt = f"""Based on the following file and subfolder explanations, provide a concise summary (2-3 sentences) of the purpose of the folder "{folder_path}".

Contents:
{context}

Provide only the folder summary, no additional commentary."""
        
        explanation = self.call_llm(prompt)
        
        return {
            "path": folder_path,
            "type": "folder",
            "explanation": explanation.strip(),
            "contents": folder_explanations
        }
    
    def analyze_directory(self, dir_path: Path, repo_root: Path) -> Dict:
        """
        Recursively analyze a directory structure.
        
        Args:
            dir_path: Path to analyze
            repo_root: Root path of the repository
            
        Returns:
            Dictionary with hierarchical explanations
        """
        items = []
        
        try:
            for item in sorted(dir_path.iterdir()):
                if item.name.startswith('.'):
                    continue
                
                if item.is_file():
                    file_explanation = self.explain_file(item, repo_root)
                    items.append(file_explanation)
                
                elif item.is_dir():
                    subdir_explanation = self.analyze_directory(item, repo_root)
                    items.append(subdir_explanation)
        
        except PermissionError:
            print(f"Permission denied: {dir_path}")
        
        rel_path = dir_path.relative_to(repo_root)
        folder_path = str(rel_path) if str(rel_path) != '.' else 'root'
        
        return self.explain_folder(items, folder_path)
    
    def save_analysis(self, analysis: Dict, output_file: str = "repo_analysis.json"):
        """
        Save analysis to a JSON file.
        
        Args:
            analysis: Analysis dictionary
            output_file: Output filename
        """
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(analysis, f, indent=2, ensure_ascii=False)
        
        print(f"\nAnalysis saved to {output_file}")
    
    def print_analysis(self, analysis: Dict, indent: int = 0):
        """
        Pretty print the analysis.
        
        Args:
            analysis: Analysis dictionary
            indent: Current indentation level
        """
        prefix = "  " * indent
        
        if analysis['type'] == 'file':
            print(f"{prefix}   {analysis['path']}")
            print(f"{prefix}   {analysis['explanation']}\n")
        else:
            print(f"{prefix}   {analysis['path']}")
            print(f"{prefix}   {analysis['explanation']}\n")
            
            for item in analysis.get('contents', []):
                self.print_analysis(item, indent + 1)

def analyse_repo(repo_path):
    analyzer = LocalLLMAnalyzer(model="llama3.2")
    
    print("Starting repository analysis...\n")
    print("=" * 60)
    
    analysis = analyzer.analyze_directory(Path(repo_path), Path(repo_path))
    
    print("\n" + "=" * 60)
    print("ANALYSIS COMPLETE")
    print("=" * 60 + "\n")
    
    analyzer.print_analysis(analysis)
    
    analyzer.save_analysis(analysis, "repo_analysis.json")

if __name__ == "__main__":
    analyse_repo("./temp_repo_057df994620947249c596e80048aff83")