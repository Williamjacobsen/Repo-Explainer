package explainer

import (
	githubapi "github.com/Williamjacobsen/Repo-Explainer/backend/internal/github_api"
)

func ExplainFile(url string) string {
	content := githubapi.GetHtml(url)
	return Llm("Explain the purpose of the different functions, classes etc. Only explain purpose. Do it in 1-2 lines per function.\n" + content)
}
