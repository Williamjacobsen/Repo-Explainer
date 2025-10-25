from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, HttpUrl
from typing import Dict, Optional
import uuid
from pathlib import Path
import asyncio
from concurrent.futures import ThreadPoolExecutor
import os
import subprocess
import shutil
import json
import requests

app = FastAPI(title="Repository Analyzer API", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

analysis_jobs: Dict[str, Dict] = {}

executor = ThreadPoolExecutor(max_workers=3)

class AnalysisRequest(BaseModel):
    repo_url: HttpUrl
    model: str = "llama3.2"

class AnalysisResponse(BaseModel):
    job_id: str
    status: str
    message: str

class JobStatus(BaseModel):
    job_id: str
    status: str
    progress: Optional[str] = None
    result: Optional[Dict] = None
    error: Optional[str] = None


class GitHubRepoFetcher:
    """Fetch all files from a GitHub repository by cloning it."""
    
    def __init__(self):
        unique_id = uuid.uuid4().hex
        self.temp_dir = f"./temp_repo_{unique_id}"
        self.repo_url = None
    
    def clone_repo(self, repo_url: str) -> bool:
        try:
            self.repo_url = repo_url
            
            if os.path.exists(self.temp_dir):
                shutil.rmtree(self.temp_dir)
            
            print(f"Cloning repository: {repo_url}")
            subprocess.run(
                ["git", "clone", repo_url, self.temp_dir],
                check=True,
                capture_output=True,
                text=True,
                timeout=300
            )
            print("Repository cloned successfully!")
            return True
        
        except subprocess.CalledProcessError as e:
            print(f"Error cloning repository: {e.stderr}")
            return False
        except subprocess.TimeoutExpired:
            print("Repository clone timed out")
            return False
        except Exception as e:
            print(f"Unexpected error: {e}")
            return False
    
    def get_repo_name(self) -> str:
        if self.repo_url:
            return self.repo_url.rstrip('/').split('/')[-1].replace('.git', '')
        return Path(self.temp_dir).name
    
    def get_temp_dir(self) -> str:
        return self.temp_dir
    
    def cleanup(self):
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            print(f"Cleaned up temporary directory: {self.temp_dir}")


class LocalLLMAnalyzer:
    """Analyze repository structure using a local LLM (Ollama)."""
    
    def __init__(self, model: str = "llama3.2", base_url: str = "http://localhost:11434"):
        self.model = model
        self.base_url = base_url
        self.api_url = f"{base_url}/api/generate"
    
    def call_llm(self, prompt: str) -> str:
        try:
            payload = {
                "model": self.model,
                "prompt": prompt,
                "stream": False
            }
            
            response = requests.post(self.api_url, json=payload, timeout=60)
            response.raise_for_status()
            
            return response.json()["response"]
        
        except Exception as e:
            print(f"Error calling LLM: {e}")
            return f"Error: Could not analyze - {str(e)}"
    
    def read_file_content(self, file_path: Path, max_lines: int = 500) -> str:
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
    
    def explain_folder(self, folder_explanations: list, folder_path: str) -> Dict:
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


def analyze_repository_sync(job_id: str, repo_url: str, model: str):
    """Synchronous function to analyze repository."""
    fetcher = GitHubRepoFetcher()
    analyzer = LocalLLMAnalyzer(model=model)
    
    try:
        analysis_jobs[job_id]["status"] = "cloning"
        analysis_jobs[job_id]["progress"] = "Cloning repository..."
        
        if not fetcher.clone_repo(repo_url):
            analysis_jobs[job_id]["status"] = "failed"
            analysis_jobs[job_id]["error"] = "Failed to clone repository"
            return
        
        repo_name = fetcher.get_repo_name()
        temp_dir = fetcher.get_temp_dir()
        
        analysis_jobs[job_id]["status"] = "analyzing"
        analysis_jobs[job_id]["progress"] = "Analyzing repository structure..."
        
        analysis = analyzer.analyze_directory(Path(temp_dir), Path(temp_dir))
        
        output_file = f"./analyses/{repo_name}_{job_id}_analysis.json"
        os.makedirs("./analyses", exist_ok=True)
        
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(analysis, f, indent=2, ensure_ascii=False)
        
        analysis_jobs[job_id]["status"] = "completed"
        analysis_jobs[job_id]["progress"] = "Analysis complete"
        analysis_jobs[job_id]["result"] = analysis
        
    except Exception as e:
        analysis_jobs[job_id]["status"] = "failed"
        analysis_jobs[job_id]["error"] = str(e)
        print(f"Error during analysis: {e}")
    
    finally:
        fetcher.cleanup()


async def analyze_repository_async(job_id: str, repo_url: str, model: str):
    """Async wrapper for repository analysis."""
    loop = asyncio.get_event_loop()
    await loop.run_in_executor(executor, analyze_repository_sync, job_id, repo_url, model)


@app.get("/")
async def root():
    return {
        "message": "Repository Analyzer API",
        "version": "1.0.0",
        "endpoints": {
            "POST /analyze": "Start repository analysis",
            "GET /status/{job_id}": "Get analysis status",
            "GET /health": "Health check"
        }
    }


@app.get("/health")
async def health_check():
    """Check if Ollama is accessible."""
    try:
        response = requests.get("http://localhost:11434/api/tags", timeout=5)
        ollama_status = "online" if response.status_code == 200 else "offline"
    except:
        ollama_status = "offline"
    
    return {
        "status": "online",
        "ollama": ollama_status
    }


@app.post("/analyze", response_model=AnalysisResponse)
async def start_analysis(
    request: AnalysisRequest,
    background_tasks: BackgroundTasks
):
    """Start a new repository analysis job."""
    
    try:
        response = requests.get("http://localhost:11434/api/tags", timeout=5)
        if response.status_code != 200:
            raise HTTPException(status_code=503, detail="Ollama service is not available")
    except:
        raise HTTPException(status_code=503, detail="Cannot connect to Ollama. Is it running?")
    
    job_id = str(uuid.uuid4())
    
    analysis_jobs[job_id] = {
        "job_id": job_id,
        "status": "queued",
        "repo_url": str(request.repo_url),
        "model": request.model,
        "progress": "Job queued",
        "result": None,
        "error": None
    }
    
    background_tasks.add_task(analyze_repository_async, job_id, str(request.repo_url), request.model)
    
    return AnalysisResponse(
        job_id=job_id,
        status="queued",
        message="Analysis job started"
    )


@app.get("/status/{job_id}", response_model=JobStatus)
async def get_job_status(job_id: str):
    """Get the status of an analysis job."""
    
    if job_id not in analysis_jobs:
        raise HTTPException(status_code=404, detail="Job not found")
    
    job = analysis_jobs[job_id]
    
    return JobStatus(
        job_id=job["job_id"],
        status=job["status"],
        progress=job.get("progress"),
        result=job.get("result"),
        error=job.get("error")
    )


@app.delete("/job/{job_id}")
async def delete_job(job_id: str):
    """Delete a job from memory."""
    
    if job_id not in analysis_jobs:
        raise HTTPException(status_code=404, detail="Job not found")
    
    del analysis_jobs[job_id]
    
    return {"message": "Job deleted successfully"}


@app.get("/jobs")
async def list_jobs():
    """List all jobs."""
    
    return {
        "total": len(analysis_jobs),
        "jobs": [
            {
                "job_id": job["job_id"],
                "status": job["status"],
                "repo_url": job["repo_url"]
            }
            for job in analysis_jobs.values()
        ]
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)