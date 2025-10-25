import os
import subprocess
import shutil
from typing import Dict, List
from pathlib import Path

class GitHubRepoFetcher:
    """Fetch all files and folders from a GitHub repository by cloning it."""
    
    def __init__(self, temp_dir: str = "./temp_repo"):
        """
        Initialize the fetcher.
        
        Args:
            temp_dir: Directory where the repository will be cloned temporarily
        """
        self.temp_dir = temp_dir
    
    def clone_repo(self, repo_url: str) -> bool:
        """
        Clone a GitHub repository.
        
        Args:
            repo_url: GitHub repository URL (https://github.com/owner/repo.git)
            
        Returns:
            True if successful, False otherwise
        """
        try:
            if os.path.exists(self.temp_dir):
                shutil.rmtree(self.temp_dir)
            
            print(f"Cloning repository: {repo_url}")
            subprocess.run(
                ["git", "clone", repo_url, self.temp_dir],
                check=True,
                capture_output=True,
                text=True
            )
            print("Repository cloned successfully!\n")
            return True
        
        except subprocess.CalledProcessError as e:
            print(f"Error cloning repository: {e.stderr}")
            return False
        except Exception as e:
            print(f"Unexpected error: {e}")
            return False
    
    def get_all_files_and_folders(self, root_path: str = None) -> Dict[str, List[str]]:
        """
        Recursively get all files and folders from the cloned repository.
        
        Args:
            root_path: Root path to start scanning (defaults to temp_dir)
            
        Returns:
            Dictionary with 'files' and 'folders' lists containing relative paths
        """
        if root_path is None:
            root_path = self.temp_dir
        
        result = {
            "files": [],
            "folders": []
        }
        
        root_path_obj = Path(root_path)
        
        for item in root_path_obj.rglob("*"):
            if ".git" in item.parts:
                continue
            
            rel_path = item.relative_to(root_path_obj)
            rel_path_str = str(rel_path)
            
            if item.is_file():
                result['files'].append(rel_path_str)
                print(f"File: {rel_path_str}")
            elif item.is_dir():
                result['folders'].append(rel_path_str)
                print(f"Folder: {rel_path_str}")
        
        return result
    
    def save_to_file(self, data: Dict[str, List[str]], output_file: str = "repo_structure.txt"):
        """
        Save the repository structure to a text file.
        
        Args:
            data: Dictionary containing files and folders lists
            output_file: Output filename
        """
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write("=" * 50 + "\n")
            f.write("FOLDERS\n")
            f.write("=" * 50 + "\n")
            for folder in sorted(data['folders']):
                f.write(f"{folder}\n")
            
            f.write("\n" + "=" * 50 + "\n")
            f.write("FILES\n")
            f.write("=" * 50 + "\n")
            for file in sorted(data['files']):
                f.write(f"{file}\n")
            
            f.write(f"\n\nTotal Folders: {len(data['folders'])}\n")
            f.write(f"Total Files: {len(data['files'])}\n")
        
        print(f"\nStructure saved to {output_file}")
    
    def cleanup(self):
        """Remove the temporary cloned repository."""
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            print(f"Cleaned up temporary directory: {self.temp_dir}")

def main():
    repo_url = "https://github.com/Williamjacobsen/AAU-Grouping-System.git"  
    
    fetcher = GitHubRepoFetcher(temp_dir="./temp_repo")
    
    try:
        if fetcher.clone_repo(repo_url):
            print("Scanning repository structure...\n")
            repo_structure = fetcher.get_all_files_and_folders()
            
            print(f"\n{'=' * 50}")
            print(f"Total folders: {len(repo_structure['folders'])}")
            print(f"Total files: {len(repo_structure['files'])}")
            print(f"{'=' * 50}")
            
            fetcher.save_to_file(repo_structure, "repo_structure.txt")
    
    finally:
        fetcher.cleanup()

if __name__ == "__main__":
    main()