package githubapi

import (
	"fmt"
	"io"
	"net/http"
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

	fmt.Println(string(body))
	return string(body)
}
