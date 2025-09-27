package githubapi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

const (
	URL                  = "https://github.com/Williamjacobsen/Analyse-Github-Repositories/tree/main"
	BRANCH               = "main"
	LOGGING              = true
	NUMBER_OF_WORKERS    = 30
)

func GetGithubRepo() {
	startNow := time.Now()
	fileUrls := discoverAllDirectoriesConcurrently()
	discoverAllFilesTime := time.Since(startNow)

	if LOGGING {
		for _, fileUrl := range fileUrls {
			fmt.Println(fileUrl)
		}
	}

	if LOGGING {
		fmt.Println("\nTime taken to discover all files:", discoverAllFilesTime)
	}
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
		path := url.PathEscape(item.Get("path").Str)
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

func discoverAllDirectoriesConcurrently() []string {
	if !strings.HasSuffix(URL, "/tree/main") {
		log.Fatal("URL must end with /tree/main")
	}

	repo := strings.TrimPrefix(URL, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/tree/main")

	rawFileUrl := "https://raw.githubusercontent.com/" + repo + "/refs/heads/" + BRANCH

	html := getHtml(URL)
	json := getJson(html, `{"props":{"initialPayload":`)
	items := gjson.Get(json, "props.initialPayload.tree.items")

	rootDirs, rootFiles := getDirectories(items, URL, rawFileUrl)

	toDiscoverCh := make(chan string, 500)
	toDiscoverChResultFiles := make(chan []string, 500)

	var workWG sync.WaitGroup
	var workersWG sync.WaitGroup

	workersWG.Add(NUMBER_OF_WORKERS)
	for w := 1; w <= NUMBER_OF_WORKERS; w++ {
		go worker(toDiscoverCh, toDiscoverChResultFiles, rawFileUrl, URL, &workWG, &workersWG)
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

	return fileUrls
}