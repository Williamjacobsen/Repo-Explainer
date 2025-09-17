package parser

type HTMLNode struct {
	Tag      string
	attributes map[string]string
	Position int
}

type HTMLTag struct {
	Tag string
	IndexSuffix int
}