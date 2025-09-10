package parser

import (
	"fmt"
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
	node := GetElementByXpath(body, "/html/body/div")
	
	fmt.Printf("Tag: %s, Position: %d\n", node.Tag, node.Position)

	return []string{}
}

func GetElementByXpath(body string, xpath string) HTMLNode {
	nodes := strings.Split(xpath[1:], "/")
	
	fmt.Printf("Nodes: %s\n", nodes)
	fmt.Printf("First node: %s\n", nodes[0])

	_HTMLNode, success := GetNextTag(body, nodes[0], 0)
	if !success {
		fmt.Printf("Tag <%s> not found in body\n", nodes[0])
    	return HTMLNode{}	
	}

	fmt.Printf("%s, %d\n", _HTMLNode.Tag, _HTMLNode.Position)

	return HTMLNode{Tag: "test", Position: 1}
}

func GetNextTag(body string, nextTag string, currentPosition int) (HTMLNode, bool) {
	tag := ""
	isTag := false
	for i := 0; i < len(body); i++ {
		if body[i] == '<' {
			isTag = true
			continue
		} else if isTag && (body[i] == '>' || body[i] == ' ') {
			fmt.Printf("%s\n", tag)
			if nextTag == strings.TrimSpace(tag) {
				return HTMLNode{Tag: strings.TrimSpace(tag), Position: i}, true
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

func CountChildren() {

}

func GetHref() {

}
