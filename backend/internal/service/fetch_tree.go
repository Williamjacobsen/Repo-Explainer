package fetch_tree

import (
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/githubapi"
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/parser"
)

func FetchTree(url string) {
	body := githubapi.FetchPage(url)
	parser.ParseRootDirectory(body)
}
