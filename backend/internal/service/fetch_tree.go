package fetch_tree

import "github.com/Williamjacobsen/Repo-Explainer/backend/internal/githubapi"

func FetchTree(url string) {
	githubapi.FetchDirectory(url)
}
