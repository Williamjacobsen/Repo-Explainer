import requests
import time
import json

API_BASE_URL = "http://localhost:8000"

def test_health():
    """Test the health endpoint."""
    print("Testing health endpoint...")
    response = requests.get(f"{API_BASE_URL}/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}\n")

def start_analysis(repo_url: str, model: str = "llama3.2"):
    """Start a new analysis job."""
    print(f"Starting analysis for: {repo_url}")
    
    response = requests.post(
        f"{API_BASE_URL}/analyze",
        json={
            "repo_url": repo_url,
            "model": model
        }
    )
    
    if response.status_code == 200:
        data = response.json()
        print(f"Job started successfully!")
        print(f"Job ID: {data['job_id']}\n")
        return data['job_id']
    else:
        print(f"Error: {response.status_code}")
        print(f"Response: {response.text}")
        return None

def check_status(job_id: str):
    """Check the status of a job."""
    response = requests.get(f"{API_BASE_URL}/status/{job_id}")
    
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error checking status: {response.status_code}")
        return None

def poll_until_complete(job_id: str, interval: int = 5):
    """Poll job status until completion."""
    print(f"Polling job {job_id}...")
    
    while True:
        status = check_status(job_id)
        
        if not status:
            break
        
        print(f"Status: {status['status']} - {status.get('progress', 'No progress info')}")
        
        if status['status'] in ['completed', 'failed']:
            print("\n" + "="*60)
            if status['status'] == 'completed':
                print("ANALYSIS COMPLETED!")
                print("="*60 + "\n")
                print(json.dumps(status['result'], indent=2))
            else:
                print("ANALYSIS FAILED!")
                print("="*60 + "\n")
                print(f"Error: {status.get('error', 'Unknown error')}")
            break
        
        time.sleep(interval)

def list_jobs():
    """List all jobs."""
    print("Listing all jobs...")
    response = requests.get(f"{API_BASE_URL}/jobs")
    
    if response.status_code == 200:
        data = response.json()
        print(f"Total jobs: {data['total']}\n")
        for job in data['jobs']:
            print(f"Job ID: {job['job_id']}")
            print(f"  Status: {job['status']}")
            print(f"  Repo: {job['repo_url']}\n")
    else:
        print(f"Error: {response.status_code}")

def main():
    test_health()
    
    repo_url = "https://github.com/Williamjacobsen/AAU-Grouping-System.git"
    job_id = start_analysis(repo_url)
    
    if job_id:
        poll_until_complete(job_id)
        
        print("\n")
        list_jobs()

if __name__ == "__main__":
    main()