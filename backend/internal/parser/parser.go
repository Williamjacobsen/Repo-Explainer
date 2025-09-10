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
	node := GetElementByXpath(body, "/html/body/div[1]/div[4]")
	
	fmt.Printf("Tag: %s, Position: %d\n", node.Tag, node.Position)

	return []string{}
}

func GetElementByXpath(body string, xpath string) HTMLNode {
	nodes := strings.Split(xpath[1:], "/")

	fmt.Printf("Nodes: %s\n", nodes)

	var _HTMLNode HTMLNode

	for _, node := range nodes {
		parsedNode := ParseTag(node)

		fmt.Printf("ParsedNode: %s, %d\n", parsedNode.Tag, parsedNode.IndexSuffix)

		_HTMLNode, success := GetNextTag(body, parsedNode.Tag, _HTMLNode.Position)
		if !success {
			fmt.Printf("Tag <%s> not found in body\n", parsedNode.Tag)
			return HTMLNode{}	
		}

		fmt.Printf("%s, %d\n\n", _HTMLNode.Tag, _HTMLNode.Position)
	}
	
	return HTMLNode{Tag: "test", Position: 1}
}

func GetNextTag(body string, nextTag string, documentPosition int) (HTMLNode, bool) {
	tag := ""
	isTag := false
	for i := documentPosition; i < len(body); i++ {
		if body[i] == '<' {
			isTag = true
			continue
		} else if isTag && (body[i] == '>' || body[i] == ' ') {
			tag = strings.TrimSpace(tag)

			if nextTag == tag {
				return HTMLNode{Tag: tag, Position: i}, true
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

func ParseTag(tag string) HTMLTag {
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

func CountChildren() {

}

func GetHref() {

}
