package main

import (
	fetch_tree "github.com/Williamjacobsen/Repo-Explainer/backend/internal/service"
)

func main() {
	fetch_tree.FetchTree("https://github.com/Williamjacobsen/ClosedAI/tree/main")
}
