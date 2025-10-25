import os
import subprocess
import shutil
from typing import Dict, List
from pathlib import Path
import uuid

class GitHubRepoFetcher:
    """Fetch all files from a GitHub repository by cloning it."""
    
    def __init__(self):
        """
        Args:
            temp_dir: Directory where the repository will be cloned temporarily
        """
        unique_id = uuid.uuid4().hex
        self.temp_dir = f"./temp_repo_{unique_id}"
    
    def clone_repo(self, repo_url: str) -> bool:
        """
        Args:
            repo_url: GitHub repository URL (https://github.com/owner/repo.git)
            
        Returns:
            True if successful, False otherwise
        """
        try:
            self.repo_url = repo_url

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
    
    def get_all_files(self, root_path: str = None) -> List[str]:
        """
        Args:
            root_path: Root path to start scanning (defaults to temp_dir)
            
        Returns:
            Files lists containing relative paths
        """
        if root_path is None:
            root_path = self.temp_dir
        
        files = [] 
        
        root_path_obj = Path(root_path)
        
        for item in root_path_obj.rglob("*"):
            if ".git" in item.parts:
                continue
            
            rel_path = item.relative_to(root_path_obj)
            rel_path_str = str(rel_path)
            
            if item.is_file():
                files.append(rel_path_str)
                print(f"File: {rel_path_str}")

        return files
    
    def get_repo_name(self) -> str:
        """Extract repository name from the cloned directory or URL."""
        if hasattr(self, 'repo_url'):
            return self.repo_url.rstrip('/').split('/')[-1].replace('.git', '')
        return Path(self.temp_dir).name

    def get_temp_dir(self) -> str:
        """Get the temporary directory path."""
        return self.temp_dir
    
    def cleanup(self):
        """Remove the temporary cloned repository."""
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            print(f"Cleaned up temporary directory: {self.temp_dir}")
