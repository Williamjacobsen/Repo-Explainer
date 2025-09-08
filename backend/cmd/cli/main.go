package main

import "github.com/Williamjacobsen/Repo-Explainer/backend/internal/githubapi"

func main() {
	githubapi.FetchTree("https://github.com/Williamjacobsen/Repo-Explainer/tree/main")
}
