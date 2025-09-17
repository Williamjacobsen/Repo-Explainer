package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetElementByXpath(body string, xpath string) HTMLNode {
	nodes := strings.Split(xpath[1:], "/")

	documentPosition := 0

	var parsedNode HTMLTag
	for _, node := range nodes {
		parsedNode = ParseXpathTag(node)

		_HTMLNode, success := GetNextTag(body, parsedNode.Tag, parsedNode.IndexSuffix, documentPosition)
		if !success {
			PrintLinesAboveAndBelow(body, documentPosition)
			panic(fmt.Sprintf("Tag <%s> not found in body", parsedNode.Tag))
		}

		documentPosition = _HTMLNode.Position
	}

	attributes := GetAttributes(body, documentPosition)

	return HTMLNode{Tag: parsedNode.Tag, Position: documentPosition + 1, attributes: attributes}
}

func GetNextTag(body string, nextTag string, nextTagIndex int, documentPosition int) (HTMLNode, bool) {
	tag := ""
	isTag := false

	for i := documentPosition; i < len(body); i++ {
		if body[i] == '<' {
			isTag = true
			continue

		} else if isTag && (body[i] == '>' || body[i] == ' ') {
			tag = strings.TrimSpace(tag)

			if body[i] == '/' { // will never run?
				tag = ""
				isTag = false
				continue
			}

			if nextTag == tag && nextTagIndex == 1 {
				return HTMLNode{Tag: tag, Position: i}, true
			} else if nextTag == tag && nextTagIndex > 1 {
				nextTagIndex--
			}

			tag = ""
			isTag = false
			continue
		}

		if isTag {
			tag += string(body[i])
		}
	}

	return HTMLNode{}, false
}

func ParseXpathTag(tag string) HTMLTag {
	indexStart := strings.Index(tag, "[")
	if indexStart == -1 {
		return HTMLTag{Tag: tag, IndexSuffix: 1}
	}
	indexSuffix := ""
	for i := indexStart + 1; i < len(tag); i++ {
		if tag[i] == ']' {
			break
		}
		indexSuffix += string(tag[i])
	}

	index, err := strconv.Atoi(indexSuffix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting indexSuffix '%s' to int: %v\n", indexSuffix, err)
		os.Exit(1)
		return HTMLTag{}
	}

	return HTMLTag{Tag: tag[:indexStart], IndexSuffix: index}
}

func GetAttributes(body string, documentPosition int) map[string]string {

	// TODO / BUG FIX: map[<tr class:DirectoryContent-module__Box_3--zI0N1]

	attributes := make(map[string]string)
	var attribute string
	for i := documentPosition + 1; i < len(body); i++ {
		if body[i] == '>' {
			break
		} else if body[i] == ' ' && body[i-1] == '"' {
			parts := strings.SplitN(attribute, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := strings.Trim(parts[1], `"`)
				attributes[key] = value
			}

			attribute = ""
		} else {
			attribute += string(body[i])
		}
	}

	if attribute != "" {
		parts := strings.SplitN(attribute, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := strings.Trim(parts[1], `"`)
			attributes[key] = value
		}
	}

	return attributes
}

// I should make a better way of traversing the DOM, like in GetChildren

func GetChildren(body string, xpath string) int {
	_HTMLNode := GetElementByXpath(body, xpath)

	childCount := 0

	currentPath := xpath

	isTag := false
	isOpeningTag := true
	tag := ""

	for i := _HTMLNode.Position; i < len(body); i++ {

		if body[i] == '<' {
			isTag = true

			if body[i+1] == '/' {
				isOpeningTag = false
			}
		}

		if isTag {
			if body[i] == '\n' {
				tag += " "
			} else {
				tag += string(body[i])
			}
		}

		if body[i] == '>' {
			if tag == "<br>" || tag[1] == '!' {
			} else if isOpeningTag {
				currentPath += "/" + GetNameFromTag(tag)
			} else {
				lastTag := ParseXpathTag(currentPath[strings.LastIndex(currentPath, "/")+1:]).Tag
				if lastTag != GetNameFromTag(tag) {
					PrintLinesAboveAndBelow(body, i)
					panic("lastTag: " + lastTag + " tag: " + tag + " currentPath: " + currentPath)
				}
				currentPath = currentPath[:strings.LastIndex(currentPath, "/")]
			}

			if tag[len(tag)-2] == '/' {
				currentPath = currentPath[:strings.LastIndex(currentPath, "/")]				
			}

			isTag = false
			isOpeningTag = true
			tag = ""

			if xpath[:strings.LastIndex(xpath, "/")] == currentPath {
				break
			} else if xpath == currentPath {
				childCount++;
			}

		}

	}

	return childCount
}

func GetNameFromTag(tag string) string {
	if tag[1] == '/' {
		return tag[2 : len(tag)-1]
	}

	return strings.Split(tag[1:len(tag)-1], " ")[0]
}

func IsClosingTag(body string, documentPosition int) bool {
	for i := documentPosition - 1; i > 0; i-- {
		switch body[i] {
		case '/':
			return true
		case '<':
			return false
		}
	}

	return false
}

func GetCurrentTag(body string, documentPosition int) string {
	var i int
	for i = documentPosition - 1; i > 0; i-- {
		if body[i] == '<' || body[i] == '/' {
			break
		}
	}

	return body[i+1 : documentPosition-1]
}

func PrintLinesAboveAndBelow(body string, documentPosition int) {
    start := max(documentPosition - 200, 0)
    end := min(documentPosition + 200, len(body))
    fmt.Print("\nLines Above\n\n")
    fmt.Print(body[start:end])
    fmt.Print("\n\nLines Below Ended\n\n")
}
