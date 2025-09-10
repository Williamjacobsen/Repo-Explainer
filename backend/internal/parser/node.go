package parser

type HTMLNode struct {
	Tag      string
	Position int
}

type HTMLTag struct {
	Tag string
	IndexSuffix int
}