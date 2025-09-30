package main

import (
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/explainer"
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/formatter"
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/github_api"
)

func main() {
	fileUrls := githubapi.GetRepo("https://github.com/Williamjacobsen/Repo-Explainer/tree/main")
	
	safeCopy := append([]string(nil), fileUrls...)
	formatter.UrlsToAsciiTree(safeCopy)	

	explainer.Llm("Hello how are you?")

	explainer.ExplainFile(fileUrls[5])

	explainer.Llm("Explain the purpose of the different functions, classes etc. Only explain purpose. Do it in 1-2 lines per function.\n" + githubapi.GetHtml(fileUrls[5]))

}
