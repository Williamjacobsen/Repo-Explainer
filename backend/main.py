from pathlib import Path
from repo_fetcher import GitHubRepoFetcher
from llm_analyzer import LocalLLMAnalyzer

def analyze_github_repo(repo_url: str, model: str = "llama3.2"):
    """
    Clone a GitHub repository and analyze it with a local LLM.
    
    Args:
        repo_url: GitHub repository URL
        model: Ollama model to use for analysis
    """
    fetcher = GitHubRepoFetcher()
    analyzer = LocalLLMAnalyzer(model=model)
    
    try:
        if not fetcher.clone_repo(repo_url):
            print("Failed to clone repository")
            return
        
        repo_name = fetcher.get_repo_name()
        temp_dir = fetcher.get_temp_dir()
        
        print("Starting repository analysis...\n")
        print("=" * 60)
        
        analysis = analyzer.analyze_directory(Path(temp_dir), Path(temp_dir))
        
        print("\n" + "=" * 60)
        print("ANALYSIS COMPLETE")
        print("=" * 60 + "\n")
        
        analyzer.print_analysis(analysis)
        
        output_file = f"{repo_name}_analysis.json"
        analyzer.save_analysis(analysis, output_file)
        
    finally:
        fetcher.cleanup()

def main():
    repo_url = "https://github.com/Williamjacobsen/AAU-Grouping-System.git"
    
    model = "llama3.2"
    
    print(f"Analyzing repository: {repo_url}")
    print(f"Using model: {model}\n")
    
    analyze_github_repo(repo_url, model)

if __name__ == "__main__":
    main()