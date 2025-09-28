package githubapi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	 urlpkg "net/url"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

const (
	BRANCH               = "main"
	LOGGING              = true
	NUMBER_OF_WORKERS    = 30
)

func GetRepo(url string) []string {
	startNow := time.Now()
	fileUrls := discoverAllDirectoriesConcurrently(url)
	discoverAllFilesTime := time.Since(startNow)

	if LOGGING {
		for _, fileUrl := range fileUrls {
			fmt.Println(fileUrl)
		}
	}

	if LOGGING {
		fmt.Println("\nTime taken to discover all files:", discoverAllFilesTime)
	}

	return fileUrls
}

func getHtml(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func getJson(html string, startOfJson string) string {
	startIndex := strings.Index(html, startOfJson)
	endIndex := -1

	currentlyOpenObjects := 0
	for i := startIndex; i < len(html); i++ {
		switch html[i] {
		case '{':
			currentlyOpenObjects++
		case '}':
			currentlyOpenObjects--
		}

		if currentlyOpenObjects == 0 {
			endIndex = i + 1
			break
		}
	}

	json := html[startIndex:endIndex]
	return json
}

func getDirectories(items gjson.Result, baseUrl string, rawFileUrl string) ([]string, []string) {
	var fileUrls []string
	var directoryUrls []string

	items.ForEach(func(_, item gjson.Result) bool {
		path := urlpkg.PathEscape(item.Get("path").Str)
		contentType := item.Get("contentType").Str

		switch contentType {
		case "file":
			fileUrls = append(fileUrls, rawFileUrl+"/"+path)
		case "directory":
			directoryUrls = append(directoryUrls, baseUrl+"/"+path)
		}

		return true
	})

	return directoryUrls, fileUrls
}

func getDirectoriesWrapper(url string, rawFileUrl string, baseUrl string) ([]string, []string) {
	html := getHtml(url)

	json := getJson(html, `{"payload":{`)

	items := gjson.Get(json, "payload.tree.items")

	return getDirectories(items, baseUrl, rawFileUrl)
}

func worker(
	toDiscoverCh chan string,
	toDiscoverChResultFiles chan<- []string,
	rawFileUrl string,
	baseUrl string,
	workWG *sync.WaitGroup,
	workersWG *sync.WaitGroup,
) {
	defer workersWG.Done()

	for urlToDiscover := range toDiscoverCh {
		if LOGGING {
			fmt.Println("Visiting:", urlToDiscover)
		}

		subDirs, fileUrls := getDirectoriesWrapper(urlToDiscover, rawFileUrl, baseUrl)

		for _, subDir := range subDirs {
			workWG.Add(1)
			toDiscoverCh <- subDir
		}

		if len(fileUrls) > 0 {
			toDiscoverChResultFiles <- fileUrls
		}

		workWG.Done()
	}
}

func discoverAllDirectoriesConcurrently(url string) []string {
	if !strings.HasSuffix(url, "/tree/main") {
		log.Fatal("URL must end with /tree/main")
	}

	repo := strings.TrimPrefix(url, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/tree/main")

	rawFileUrl := "https://raw.githubusercontent.com/" + repo + "/refs/heads/" + BRANCH

	html := getHtml(url)
	json := getJson(html, `{"props":{"initialPayload":`)
	items := gjson.Get(json, "props.initialPayload.tree.items")

	rootDirs, rootFiles := getDirectories(items, url, rawFileUrl)

	toDiscoverCh := make(chan string, 500)
	toDiscoverChResultFiles := make(chan []string, 500)

	var workWG sync.WaitGroup
	var workersWG sync.WaitGroup

	workersWG.Add(NUMBER_OF_WORKERS)
	for w := 1; w <= NUMBER_OF_WORKERS; w++ {
		go worker(toDiscoverCh, toDiscoverChResultFiles, rawFileUrl, url, &workWG, &workersWG)
	}

	go func() {
		workersWG.Wait()
		close(toDiscoverChResultFiles)
	}()

	for _, dir := range rootDirs {
		workWG.Add(1)
		toDiscoverCh <- dir
	}

	go func() {
		workWG.Wait()
		close(toDiscoverCh)
	}()

	fileUrls := append([]string(nil), rootFiles...)
	for results := range toDiscoverChResultFiles {
		fileUrls = append(fileUrls, results...)
	}

	for i := range fileUrls {
		if decoded, err := urlpkg.PathUnescape(fileUrls[i]); err == nil {
			fileUrls[i] = decoded
		}
	}	

	return fileUrls
}