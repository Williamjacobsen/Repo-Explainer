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
				childCount++
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

// -------------------- data types --------------------

type Node struct {
	Tag      string
	StartPos int
	EndPos   int
}

type Tree struct {
	Node
	Children []*Tree

	Parent     *Tree
	childIndex int
}

// -------------------- config --------------------

var voidTags = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

var rawTextTags = map[string]bool{
	"script": true,
	"style":  true,
}

func ShouldBeNested(body string, tag string) bool {
	if len(tag) == 0 {
		return false
	}
	if tag[0] == '/' {
		tag = tag[1:]
	}

	if voidTags[tag] {
		return false
	}

	if rawTextTags[tag] {
		SkipRawText(body, tag)
		return false
	}

	return true
}

func SkipRawText(body, tag string) {
	close := "</" + tag + ">"
	idx := strings.Index(strings.ToLower(body[DiscoveredPointer:]), close)
	if idx >= 0 {
		DiscoveredPointer += idx + len(close)
	} else {
		DiscoveredPointer = len(body)
	}
}

var DiscoveredPointer int

// -------------------- parsing helpers --------------------

func GetNewTag(body string, pos int) (Node, error) {
	tag := ""
	isTag := false

	for i := DiscoveredPointer; i < len(body); i++ {
		switch body[i] {
		case '<':

			// HTML Comment
			if i+4 <= len(body) && strings.HasPrefix(body[i:], "<!--") {
				end := strings.Index(body[i+4:], "-->")
				i = i + 4 + end + 3 - 1 // Land on '>'
				DiscoveredPointer = i + 1
				continue
			}

			isTag = true
		case '>':
			tag = strings.TrimSpace(tag)
			if tag == "" {
				return Node{}, fmt.Errorf("tag is empty")
			}
			tag = strings.ToLower(strings.Split(tag, " ")[0])
			// TODO: also get attributes etc

			DiscoveredPointer = i + 1
			return Node{Tag: tag, StartPos: i + 1}, nil
		default:
			if isTag {
				if body[i] == '\n' {
					tag += " "
				} else {
					tag += string(body[i])
				}
			}
		}
	}

	return Node{}, fmt.Errorf("could not find the next tag")
}

func GetNextTag2(body string, pos int) (Node, error) {
	//if DiscoveredPointer <= pos {
	//	return GetNewTag(body, pos)
	//}
	return GetNewTag(body, pos)

	//return Node{}, fmt.Errorf("could not find the next tag")
}

func ParseXpath2(xpath string) ([]string, error) {
	var xpathNodes []string
	if xpath == "" {
		return []string{}, fmt.Errorf("xpath is nothing")
	}
	if xpath[0] == '/' {
		xpathNodes = strings.Split(xpath[1:], "/")
	} else {
		return []string{}, fmt.Errorf("xpath has a wrong format (doesn't start with '/')")
	}
	if len(xpathNodes) == 0 {
		return []string{}, fmt.Errorf("len(xpathNodes) is 0")
	}

	return xpathNodes, nil
}

// -------------------- tree construction --------------------

func GetRoot(body string, tree *Tree) (*Tree, error) {
	var node Node
	pos := 0
	for {
		var err error
		node, err = GetNextTag2(body, pos)
		if err != nil {
			return nil, fmt.Errorf("could not get root")
		}

		if node.Tag == "html" {
			break
		}

		pos = node.StartPos
	}

	root := &Tree{
		Node: Node{
			Tag:      node.Tag,
			StartPos: node.StartPos,
		},
		Children: []*Tree{},
	}

	return root, nil
}

func EnsureTreeExists(body string, tree *Tree) (*Tree, error) {
	if tree == nil {
		var err error
		tree, err = GetRoot(body, tree)
		if err != nil {
			return nil, fmt.Errorf("failed to get root: %w", err)
		}
	}

	return tree, nil
}

func AppendNextTag(body string, tree *Tree) (*Tree, error) {
	node, err := GetNextTag2(body, tree.Node.StartPos)
	if err != nil {
		return tree, fmt.Errorf("could not get next tag")
	}

	if node.Tag[0] == '/' {
		if tree.Parent != nil {
			return tree.Parent, nil
		}
		return tree, nil
	}

	child := &Tree{
		Node: Node{
			Tag:      node.Tag,
			StartPos: node.StartPos,
		},
		Children:   []*Tree{},
		Parent:     tree,
		childIndex: len(tree.Children),
	}

	tree.Children = append(tree.Children, child)

	if ShouldBeNested(body, node.Tag) {
		return child, nil
	}
	return tree, nil
}

// -------------------- API stubs using the tree --------------------

func ConstructTree(body string) (tree *Tree, err error) {
	node, err := GetNextTag2(body, 0)
	if err != nil {
		panic("Couldn't get first tag")
	}

	tree = &Tree{
		Node: Node{
			Tag:      node.Tag,
			StartPos: node.StartPos,
		},
	}

	for i := node.StartPos; i < len(body); i = node.StartPos {
		node, err = GetNextTag2(body, i)
		if err != nil {
			fmt.Println("Coundn't get next tag:", err)
			break
		}
		fmt.Println(node)

		// is it a closing tag?
		// 		is the tag the same as the current tree node?
		// 			move back to parent tree node
		// 		else
		// 			add it to the tree
		// else
		// 		add it to the tree and make it the current tree node

		if node.Tag[0] == '/' {
			

			
		} else {

		}


	}

	return tree, nil
}

func GetTagByXpath2(body string, xpath string, tree *Tree) (*Tree, error) {
	xpathNodes, err := ParseXpath2(xpath)
	if err != nil {
		panic(err)
	}

	fmt.Println(xpathNodes)

	if tree == nil {
		tree, err = EnsureTreeExists(body, tree)
		if err != nil {
			return &Tree{}, fmt.Errorf("could not get root tag <html>: %w", err)
		} else if tree == nil {
			return &Tree{}, fmt.Errorf("EnsureTreeExists returned nil tree")
		}
	}

	child, _ := AppendNextTag(body, tree)

	for i := 0; i < len(body); i++ {
		child, _ = AppendNextTag(body, child)
	}

	return tree, nil
}

func GetChildren2(body string, xpath string, tree *Tree) {
	//tree, _ = GetTagByXpath2(body, xpath, tree)
	tree, _ = ConstructTree(body)

	PrintTree(tree)
}

// -------------------- helper functions --------------------

func PrintLinesAboveAndBelow(body string, documentPosition int) {
	start := max(documentPosition-200, 0)
	end := min(documentPosition+200, len(body))
	fmt.Print("\nLines Above\n\n")
	fmt.Print(body[start:end])
	fmt.Print("\n\nLines Below Ended\n\n")
}

func PrintTree(tree *Tree) {
	printTreeRecursive(tree, 0)
}

func printTreeRecursive(tree *Tree, depth int) {
	if tree == nil {
		fmt.Println("<nil>")
		return
	}
	indent := strings.Repeat(" ", depth)
	fmt.Printf("%s<%s pos=%d>\n", indent, tree.Node.Tag, tree.Node.StartPos)
	for _, child := range tree.Children {
		printTreeRecursive(child, depth+1)
	}
}

/*
Notable metion:
<div>
	<!-- '"` --><!-- </textarea></xmp> --></option>
	</form>
	<form>
	</form>
</div>
The comment is an HTML injection defense trick
*/
