package githubapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Williamjacobsen/Repo-Explainer/backend/internal/parser"
)

func FetchPage(git_url string) string {

	resp, err := http.Get(git_url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Github request failed %s", resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}

func ParseRootDirectory(body string) []parser.HTMLNode {
	childCount := parser.GetChildren(body, "/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody")

	HTMLNodes := []parser.HTMLNode{}
	for i := 2; i < childCount; i++ {
		_HTMLNode := parser.GetElementByXpath(body, fmt.Sprintf("/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody/tr[%d]/td[2]/div/div/div/div/a", i))
		HTMLNodes = append(HTMLNodes, _HTMLNode)
	}

	return HTMLNodes
}
