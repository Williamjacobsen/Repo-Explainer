package explainer

import (
	"fmt"

	githubapi "github.com/Williamjacobsen/Repo-Explainer/backend/internal/github_api"
)

func ExplainFile(url string) {
	fmt.Println(githubapi.GetHtml(url))
}
