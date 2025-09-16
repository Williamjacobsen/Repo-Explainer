package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
Get children from:
/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody

Go through child index 1 to length-2

Get href from:
/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody/tr[2]/td[2]/div/div/div/div/a
*/

func ParseDirectory(body string) []string {
	//node := GetElementByXpath(body, "/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody/tr[2]/td[2]/div/div/div/div/a")
	//fmt.Printf("Tag: %s, Position: %d\n", node.Tag, node.Position)

	GetChildren(body, "/html/body/div[1]/div[4]/div/main/turbo-frame/div/div/div/div/div[1]/react-partial/div/div/div[3]/div[1]/table/tbody")

	return []string{}
}

func GetElementByXpath(body string, xpath string) HTMLNode {
	nodes := strings.Split(xpath[1:], "/")

	fmt.Printf("Nodes: %s\n", nodes)

	documentPosition := 0

	for _, node := range nodes {
		parsedNode := ParseXpathTag(node)

		fmt.Printf("ParsedNode: %s, %d\n", parsedNode.Tag, parsedNode.IndexSuffix)

		fmt.Printf("GetNextTag(%s, %s, %d)\n", "body", parsedNode.Tag, documentPosition)

		_HTMLNode, success := GetNextTag(body, parsedNode.Tag, parsedNode.IndexSuffix, documentPosition)
		if !success {
			fmt.Printf("Tag <%s> not found in body\n", parsedNode.Tag)
			return HTMLNode{}
		}

		documentPosition = _HTMLNode.Position

		fmt.Print("\n")
	}

	attributes := GetAttributes(body, documentPosition)
	fmt.Println(attributes)
	fmt.Printf("Label: %s\n", attributes["aria-label"])

	// Current documentPosition is after the tag is closed
	return HTMLNode{Tag: "test", Position: documentPosition + 1}
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

			if body[i] == '/' {
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

func GetChildren(body string, xpath string) {

	// TODO: Make it work for div[3] etc

	_HTMLNode := GetElementByXpath(body, xpath)

	fmt.Println(_HTMLNode.Position)

	tagName := GetCurrentTag(body, _HTMLNode.Position)
	fmt.Println(tagName)

	isClosingTag := IsClosingTag(body, _HTMLNode.Position)
	fmt.Println(isClosingTag)

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
			fmt.Println(tag)

			if tag == "<br>" || tag[1] == '!' {
				fmt.Println("Skipping: " + tag)
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

			fmt.Println(currentPath)
			isTag = false
			isOpeningTag = true
			tag = ""
		}

	}

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
	fmt.Print("\nLines Above\n\n")
	for i := documentPosition-200; i < documentPosition+200; i++ {
		fmt.Print(string(body[i]))
	}
	fmt.Print("\n\nLines Below Ended\n")
	fmt.Println()
}
