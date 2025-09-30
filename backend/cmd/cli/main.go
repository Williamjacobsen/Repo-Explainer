package main

import (
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/explainer"
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/formatter"
	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/github_api"
)

func main() {
	fileUrls := githubapi.GetRepo("https://github.com/Williamjacobsen/Repo-Explainer/tree/main")
	formatter.UrlsToAsciiTree(fileUrls)	

	explainer.Llm("Hello how are you?")
}
