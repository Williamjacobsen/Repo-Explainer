package formatter

import (
	"fmt"
	"strings"
)

func UrlsToAsciiTree(urls []string) {
	for _, s := range urls {
		path := strings.Split(s, "/")
		path = path[8:]
		fmt.Println(strings.Join(path, "/"))
	}
}