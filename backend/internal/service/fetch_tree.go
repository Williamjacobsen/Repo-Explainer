package fetch_tree

import (
	"fmt"

	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/githubapi"
)

func FetchTree(url string) {
	//body := githubapi.FetchPage(url)
	//HTMLNodes := githubapi.ParseRootDirectory(body)

	//fmt.Println(HTMLNodes)

	fmt.Println(githubapi.ParseSubDirectory(githubapi.FetchPage("https://github.com/Williamjacobsen/ClosedAI/tree/main/System%20Design")))
}
