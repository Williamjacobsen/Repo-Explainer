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

	return HTMLNode{Tag: "test", Position: 1}
}

func GetTag() {

}

func CountChildren() {

}

func GetHref() {

}
