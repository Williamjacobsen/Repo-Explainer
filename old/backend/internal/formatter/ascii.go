package formatter

import (
	"fmt"
	"strings"
)

type node struct {
	name     string
	children []*node
	isFile   bool
}

func UrlsToAsciiTree(urls []string) {
	fmt.Println("\nGithub Repository Structure:")
	for i, s := range urls {
		nodes := strings.Split(s, "/")
		nodes = nodes[8:]
		urls[i] = strings.Join(nodes, "/")
	}

	tree := constructTree(urls)
	printTree(tree)
}

func constructTree(urls []string) *node {
	root := newNode("ROOT")

	for _, url := range urls {
		urlNodes := strings.Split(url, "/")
		curr := root

		for _, urlNode := range urlNodes {
			curr = findOrCreateChild(curr, urlNode)
		}
		curr.isFile = true
	}

	return root
}

func findOrCreateChild(parent *node, name string) *node {
	for _, child := range parent.children {
		if child.name == name {
			return child
		}
	}
	child := newNode(name)
	parent.children = append(parent.children, child)
	return child
}

func newNode(name string) *node {
	return &node{name: name}
}

func printTree(tree *node) {
	printTreeHelper(tree, 0)
}

func printTreeHelper(tree *node, depth int) {
	if tree.isFile {
		fmt.Println(strings.Repeat(" ", depth*2) + tree.name)
	} else {
		fmt.Println(strings.Repeat(" ", depth*2) + tree.name + "/")
	}

	for _, child := range tree.children {
		printTreeHelper(child, depth+1)
	}
}
