package explainer

import (
	"fmt"
	"regexp"
	"strings"
)

type Pattern struct {
	Kind  string
	Regex string
}

type LanguagePatterns struct {
	Patterns []Pattern
}

var Registry = map[string]LanguagePatterns{
	".go": {
		Patterns: []Pattern{
			{"function", `func\s+([A-Za-z_]\w*)\s*\(`},
		},
	},
	".py": {
		Patterns: []Pattern{
			{"function", `def\s+([A-Za-z_]\w*)\s*\(`},
		},
	},
	".js": {
		Patterns: []Pattern{
			{"function", `function\s+([A-Za-z_]\w*)\s*\(`},
			{"arrowFunc", `([A-Za-z_]\w*)\s*=\s*\([^)]*\)\s*=>`},
		},
	},
}

func GetMatches(fileExtension string, code string) []string {
	languagePatterns, ok := Registry[fileExtension]
	if !ok {
		return nil
	}

	var results []string
	for _, pattern := range languagePatterns.Patterns {
		re := regexp.MustCompile(pattern.Regex)
		for _, match := range re.FindAllStringSubmatch(code, -1) { // seconds param is a limit to how many matches
			if len(match) > 1 {
				results = append(results, fmt.Sprintf("%s: %s", pattern.Kind, match[1]))
			} else {
				results = append(results, match[0])
			}
		}
	}
	return results
}

func Example() {
	code := `
package main

import "fmt"

func Hello(name string) {
	fmt.Println("Hello", name)
}
`
	matches := GetMatches(".go", code)
	fmt.Println(strings.Join(matches, "\n"))
}
